package telegraph

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

// request seen by the fake API server
type seenRequest struct {
	method      string // API method, e.g. "getPage"
	escapedPath string // raw URL path, e.g. "/getPage/a%20b"
	form        url.Values
}

// fakeAPI starts a server that records requests and answers with respond(method, form).
func fakeAPI(t *testing.T, token string, respond func(method string, form url.Values) (status int, body string)) (*Client, *[]seenRequest) {
	t.Helper()
	var mu sync.Mutex
	var seen []seenRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		method, _, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
		mu.Lock()
		seen = append(seen, seenRequest{method: method, escapedPath: r.URL.EscapedPath(), form: r.PostForm})
		mu.Unlock()
		status, body := respond(method, r.PostForm)
		w.WriteHeader(status)
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	return NewClient(token, WithBaseURL(srv.URL), WithHTTPClient(srv.Client())), &seen
}

func ok(result string) (int, string) {
	return http.StatusOK, `{"ok":true,"result":` + result + `}`
}

const pageWithContent = `{"path":"Sample-12-15","url":"https://telegra.ph/Sample-12-15","title":"Sample","description":"","views":3,
	"content":[{"tag":"p","children":["Hello, ",{"tag":"a","attrs":{"href":"https://example.com"},"children":["world"]},"!"]},{"tag":"br"}]}`

func TestPageContentDecodesToNodeElements(t *testing.T) {
	c, _ := fakeAPI(t, "", func(string, url.Values) (int, string) { return ok(pageWithContent) })

	page, err := c.GetPage(context.Background(), "Sample-12-15", true)
	if err != nil {
		t.Fatal(err)
	}
	p, isElement := page.Content[0].(NodeElement)
	if !isElement {
		t.Fatalf("content[0] is %T, want NodeElement", page.Content[0])
	}
	if s, _ := p.Children[0].(string); s != "Hello, " {
		t.Errorf("first child = %#v, want text node", p.Children[0])
	}
	a, isElement := p.Children[1].(NodeElement)
	if !isElement || a.Tag != "a" || a.Attrs["href"] != "https://example.com" {
		t.Errorf("second child = %#v, want <a href>", p.Children[1])
	}
}

// Content fetched with GetPage must survive being sent back with EditPage.
func TestEditPageSendsFetchedContentUnchanged(t *testing.T) {
	c, seen := fakeAPI(t, "token", func(string, url.Values) (int, string) { return ok(pageWithContent) })
	ctx := context.Background()

	page, err := c.GetPage(ctx, "Sample-12-15", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.EditPage(ctx, EditPageParams{Path: page.Path, Title: page.Title, Content: page.Content}); err != nil {
		t.Fatal(err)
	}

	var original struct{ Content json.RawMessage }
	if err := json.Unmarshal([]byte(pageWithContent), &original); err != nil {
		t.Fatal(err)
	}
	assertSameJSON(t, (*seen)[1].form.Get("content"), string(original.Content))
}

func TestCreatePageRejectsUnsupportedNodeBeforeSending(t *testing.T) {
	c, seen := fakeAPI(t, "token", func(string, url.Values) (int, string) { return ok(`{}`) })

	_, err := c.CreatePage(context.Background(), CreatePageParams{Title: "t", Content: Nodes{"text", 42}})
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("err = %v, want ErrInvalidParameter", err)
	}
	if len(*seen) != 0 {
		t.Errorf("sent %d requests, want none", len(*seen))
	}
}

func TestCreatePageAcceptsElementPointers(t *testing.T) {
	c, seen := fakeAPI(t, "token", func(string, url.Values) (int, string) { return ok(`{}`) })

	content := Nodes{&NodeElement{Tag: "p", Children: Nodes{"x"}}}
	if _, err := c.CreatePage(context.Background(), CreatePageParams{Title: "t", Content: content, ReturnContent: true}); err != nil {
		t.Fatal(err)
	}
	form := (*seen)[0].form
	assertSameJSON(t, form.Get("content"), `[{"tag":"p","children":["x"]}]`)
	if form.Get("return_content") != "true" || form.Get("access_token") != "token" {
		t.Errorf("form = %v", form)
	}
}

func TestPageParamValidation(t *testing.T) {
	c, seen := fakeAPI(t, "token", func(string, url.Values) (int, string) { return ok(`{}`) })
	ctx := context.Background()
	content := Nodes{"x"}

	cases := map[string]func() error{
		"empty title": func() error {
			_, err := c.CreatePage(ctx, CreatePageParams{Content: content})
			return err
		},
		"title too long": func() error {
			_, err := c.CreatePage(ctx, CreatePageParams{Title: strings.Repeat("가", 257), Content: content})
			return err
		},
		"no content": func() error {
			_, err := c.CreatePage(ctx, CreatePageParams{Title: "t"})
			return err
		},
		"edit without path": func() error {
			_, err := c.EditPage(ctx, EditPageParams{Title: "t", Content: content})
			return err
		},
		"short name too long": func() error {
			_, err := c.CreateAccount(ctx, CreateAccountParams{ShortName: strings.Repeat("a", 33)})
			return err
		},
	}
	for name, run := range cases {
		if err := run(); !errors.Is(err, ErrInvalidParameter) {
			t.Errorf("%s: err = %v, want ErrInvalidParameter", name, err)
		}
	}
	// 256 multi-byte characters are within the limit
	if _, err := c.CreatePage(ctx, CreatePageParams{Title: strings.Repeat("가", 256), Content: content}); err != nil {
		t.Errorf("256-character title: %v", err)
	}
	if len(*seen) != 1 {
		t.Errorf("sent %d requests, want 1", len(*seen))
	}
}

func TestMethodsNeedingTokenFailWithoutIt(t *testing.T) {
	c, seen := fakeAPI(t, "", func(string, url.Values) (int, string) { return ok(`{}`) })
	ctx := context.Background()

	_, err := c.GetPageList(ctx, 0, 0)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Errorf("GetPageList err = %v, want ErrInvalidParameter", err)
	}
	if len(*seen) != 0 {
		t.Errorf("sent %d requests, want none", len(*seen))
	}
}

func TestRevokeAccessTokenUpdatesClient(t *testing.T) {
	c, seen := fakeAPI(t, "old", func(method string, _ url.Values) (int, string) {
		if method == "revokeAccessToken" {
			return ok(`{"short_name":"s","author_name":"","author_url":"","access_token":"new","auth_url":"https://edit.telegra.ph/auth/x"}`)
		}
		return ok(`{"short_name":"s","author_name":"","author_url":""}`)
	})
	ctx := context.Background()

	if _, err := c.RevokeAccessToken(ctx); err != nil {
		t.Fatal(err)
	}
	if got := c.AccessToken(); got != "new" {
		t.Fatalf("AccessToken() = %q, want new", got)
	}
	if _, err := c.GetAccountInfo(ctx); err != nil {
		t.Fatal(err)
	}
	if got := (*seen)[1].form.Get("access_token"); got != "new" {
		t.Errorf("next request used token %q, want new", got)
	}
}

func TestEditAccountInfoSendsOnlyGivenFields(t *testing.T) {
	c, seen := fakeAPI(t, "token", func(string, url.Values) (int, string) {
		return ok(`{"short_name":"s","author_name":"a","author_url":""}`)
	})

	if _, err := c.EditAccountInfo(context.Background(), EditAccountInfoParams{AuthorName: Ptr("a"), AuthorURL: Ptr("")}); err != nil {
		t.Fatal(err)
	}
	form := (*seen)[0].form
	if _, sent := form["short_name"]; sent {
		t.Errorf("short_name sent although not given: %v", form)
	}
	if got, sent := form["author_url"]; !sent || got[0] != "" {
		t.Errorf("author_url = %v, want sent as empty to clear it", got)
	}
	if form.Get("author_name") != "a" {
		t.Errorf("author_name = %q", form.Get("author_name"))
	}
}

func TestGetAccountInfoFields(t *testing.T) {
	c, seen := fakeAPI(t, "token", func(string, url.Values) (int, string) { return ok(`{"page_count":3}`) })
	ctx := context.Background()

	if _, err := c.GetAccountInfo(ctx); err != nil {
		t.Fatal(err)
	}
	acc, err := c.GetAccountInfo(ctx, FieldShortName, FieldPageCount)
	if err != nil {
		t.Fatal(err)
	}
	if _, sent := (*seen)[0].form["fields"]; sent {
		t.Errorf("fields sent without being asked: %v", (*seen)[0].form)
	}
	assertSameJSON(t, (*seen)[1].form.Get("fields"), `["short_name","page_count"]`)
	if acc.PageCount != 3 {
		t.Errorf("PageCount = %d", acc.PageCount)
	}
}

// getPage ignores access_token, so the client must not hand it out there.
func TestGetPageDoesNotSendToken(t *testing.T) {
	c, seen := fakeAPI(t, "token", func(string, url.Values) (int, string) { return ok(`{"path":"p"}`) })

	if _, err := c.GetPage(context.Background(), "p", false); err != nil {
		t.Fatal(err)
	}
	form := (*seen)[0].form
	if form.Has("access_token") {
		t.Errorf("access_token sent to getPage: %v", form)
	}
	if form.Has("return_content") {
		t.Errorf("return_content sent although false: %v", form)
	}
}

func TestPathIsEscaped(t *testing.T) {
	c, seen := fakeAPI(t, "", func(string, url.Values) (int, string) { return ok(`{}`) })

	if _, err := c.GetPage(context.Background(), "a b?c#d/e", false); err != nil {
		t.Fatal(err)
	}
	if got, want := (*seen)[0].escapedPath, "/getPage/a%20b%3Fc%23d%2Fe"; got != want {
		t.Errorf("path = %s, want %s", got, want)
	}
}

func TestGetPageListLimit(t *testing.T) {
	c, seen := fakeAPI(t, "token", func(string, url.Values) (int, string) { return ok(`{"total_count":0,"pages":[]}`) })
	ctx := context.Background()

	if _, err := c.GetPageList(ctx, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetPageList(ctx, 10, 200); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][2]int{{-1, 0}, {0, -1}, {0, 201}} {
		if _, err := c.GetPageList(ctx, bad[0], bad[1]); !errors.Is(err, ErrInvalidParameter) {
			t.Errorf("GetPageList%v err = %v, want ErrInvalidParameter", bad, err)
		}
	}

	if form := (*seen)[0].form; form.Has("limit") || form.Has("offset") {
		t.Errorf("defaults sent explicitly: %v", form)
	}
	if form := (*seen)[1].form; form.Get("offset") != "10" || form.Get("limit") != "200" {
		t.Errorf("form = %v", form)
	}
	if len(*seen) != 2 {
		t.Errorf("sent %d requests, want 2", len(*seen))
	}
}

func TestGetViewsParams(t *testing.T) {
	c, seen := fakeAPI(t, "", func(string, url.Values) (int, string) { return ok(`{"views":5}`) })
	ctx := context.Background()

	valid := []struct {
		params GetViewsParams
		want   url.Values
	}{
		{GetViewsParams{}, url.Values{}},
		{GetViewsParams{Year: 2016}, url.Values{"year": {"2016"}}},
		{GetViewsParams{Year: 2016, Month: 12, Day: 15, Hour: Ptr(0)},
			url.Values{"year": {"2016"}, "month": {"12"}, "day": {"15"}, "hour": {"0"}}},
	}
	for i, tc := range valid {
		views, err := c.GetViews(ctx, "p", tc.params)
		if err != nil {
			t.Fatalf("%+v: %v", tc.params, err)
		}
		if views.Views != 5 {
			t.Errorf("views = %d", views.Views)
		}
		if got := (*seen)[i].form.Encode(); got != tc.want.Encode() {
			t.Errorf("%+v sent %q, want %q", tc.params, got, tc.want.Encode())
		}
	}

	invalid := []GetViewsParams{
		{Month: 1},                           // month without year
		{Year: 2016, Day: 1},                 // day without month
		{Year: 2016, Month: 1, Hour: Ptr(3)}, // hour without day
		{Year: 1999},                         // out of range
		{Year: 2016, Month: 13},              // out of range
		{Year: 2016, Month: 1, Day: 1, Hour: Ptr(25)}, // out of range
	}
	for _, p := range invalid {
		if _, err := c.GetViews(ctx, "p", p); !errors.Is(err, ErrInvalidParameter) {
			t.Errorf("%+v: err = %v, want ErrInvalidParameter", p, err)
		}
	}
	if len(*seen) != len(valid) {
		t.Errorf("sent %d requests, want %d", len(*seen), len(valid))
	}
}

func TestAPIError(t *testing.T) {
	c, _ := fakeAPI(t, "", func(string, url.Values) (int, string) {
		return http.StatusOK, `{"ok":false,"error":"FLOOD_WAIT_7"}`
	})

	_, err := c.GetPage(context.Background(), "p", false)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v (%T), want *APIError", err, err)
	}
	if apiErr.Method != "getPage" || apiErr.Code != "FLOOD_WAIT_7" {
		t.Errorf("APIError = %+v", apiErr)
	}
	if wait, isFlood := apiErr.FloodWait(); !isFlood || wait != 7*time.Second {
		t.Errorf("FloodWait() = %v, %v", wait, isFlood)
	}
	if _, isFlood := (&APIError{Code: "PAGE_NOT_FOUND"}).FloodWait(); isFlood {
		t.Error("PAGE_NOT_FOUND reported as flood wait")
	}
}

func TestHTTPError(t *testing.T) {
	c, _ := fakeAPI(t, "", func(string, url.Values) (int, string) {
		return http.StatusBadGateway, "<html>bad gateway</html>"
	})

	_, err := c.GetPage(context.Background(), "p", false)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("err = %v (%T), want *HTTPError", err, err)
	}
	if httpErr.StatusCode != http.StatusBadGateway || !strings.Contains(httpErr.Body, "bad gateway") {
		t.Errorf("HTTPError = %+v", httpErr)
	}
}

func TestContextCancellation(t *testing.T) {
	c, _ := fakeAPI(t, "", func(string, url.Values) (int, string) { return ok(`{}`) })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := c.GetPage(ctx, "p", false); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestLoggerRedactsAccessToken(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"ok":true,"result":{}}`)
	}))
	defer srv.Close()
	c := NewClient("secret-token", WithBaseURL(srv.URL), WithLogger(logger))

	if _, err := c.GetAccountInfo(context.Background()); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "secret-token") {
		t.Errorf("token leaked into log: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "getAccountInfo") {
		t.Errorf("request not logged: %s", buf.String())
	}
}

func TestCreateAndLoad(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		switch {
		case strings.HasPrefix(r.URL.Path, "/createAccount"):
			_, _ = fmt.Fprint(w, `{"ok":true,"result":{"short_name":"s","author_name":"","author_url":"","access_token":"created"}}`)
		case r.PostForm.Get("access_token") == "good":
			_, _ = fmt.Fprint(w, `{"ok":true,"result":{"short_name":"s"}}`)
		default:
			_, _ = fmt.Fprint(w, `{"ok":false,"error":"ACCESS_TOKEN_INVALID"}`)
		}
	}))
	defer srv.Close()
	ctx := context.Background()
	opt := WithBaseURL(srv.URL)

	c, err := Create(ctx, CreateAccountParams{ShortName: "s"}, opt)
	if err != nil || c.AccessToken() != "created" {
		t.Fatalf("Create = %v, %v", c, err)
	}
	if c, err := Load(ctx, "good", opt); err != nil || c == nil {
		t.Errorf("Load(good) = %v, %v", c, err)
	}
	if c, err := Load(ctx, "bad", opt); err == nil || c != nil {
		t.Errorf("Load(bad) = %v, %v; want nil client and error", c, err)
	}
}

func assertSameJSON(t *testing.T, got, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal([]byte(got), &g); err != nil {
		t.Fatalf("got invalid JSON %q: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("want invalid JSON %q: %v", want, err)
	}
	gb, _ := json.Marshal(g)
	wb, _ := json.Marshal(w)
	if !bytes.Equal(gb, wb) {
		t.Errorf("JSON mismatch\n got: %s\nwant: %s", gb, wb)
	}
}
