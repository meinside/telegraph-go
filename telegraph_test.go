package telegraph

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

const htmlForTest = `This page is for <b>testing</b>.
THIS IS THE FIRST LINE.
<p>
	<img src="https://telegra.ph/images/logo.png" alt="Telegraph logo">
</p>
<font color="#FF0000">THIS IS THE RED LINE.</font> <i>(coloring not supported)</i>`

// TestLiveAPI runs every method against the real Telegraph API.
// It creates a throwaway account and pages on telegra.ph; `go test -short` skips it.
// Set VERBOSE=true to log requests.
func TestLiveAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("calls the real Telegraph API")
	}

	var opts []Option
	if os.Getenv("VERBOSE") == "true" {
		opts = append(opts, WithLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	created, err := Create(ctx, CreateAccountParams{ShortName: "telegraph-test", AuthorName: "Telegraph Test"}, opts...)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// a client loaded from a saved token must work the same as the created one
	loaded, err := Load(ctx, created.AccessToken(), opts...)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	for name, client := range map[string]*Client{"created": created, "loaded": loaded} {
		t.Run(name, func(t *testing.T) { exerciseClient(ctx, t, client) })
	}

	old := loaded.AccessToken()
	acc, err := loaded.RevokeAccessToken(ctx)
	if err != nil {
		t.Fatalf("RevokeAccessToken: %v", err)
	}
	if acc.AccessToken == "" || loaded.AccessToken() != acc.AccessToken || acc.AccessToken == old {
		t.Fatalf("token not rotated: old %q, returned %q, client %q", old, acc.AccessToken, loaded.AccessToken())
	}
	if _, err := loaded.GetAccountInfo(ctx); err != nil {
		t.Errorf("GetAccountInfo with revoked-then-new token: %v", err)
	}
}

func exerciseClient(ctx context.Context, t *testing.T, c *Client) {
	acc, err := c.GetAccountInfo(ctx, FieldShortName, FieldPageCount)
	if err != nil {
		t.Fatalf("GetAccountInfo: %v", err)
	}
	t.Logf("account %q has %d pages", acc.ShortName, acc.PageCount)

	if _, err := c.EditAccountInfo(ctx, EditAccountInfoParams{AuthorName: Ptr("Telegraph API Test")}); err != nil {
		t.Errorf("EditAccountInfo: %v", err)
	}

	page, err := c.CreatePageWithHTML(ctx, CreatePageParams{Title: "Test page", AuthorName: "Telegraph Test", ReturnContent: true}, htmlForTest)
	if err != nil {
		t.Fatalf("CreatePageWithHTML: %v", err)
	}
	if !page.CanEdit {
		t.Error("CreatePage: CanEdit = false for the owner")
	}

	fetched, err := c.GetPage(ctx, page.Path, true)
	if err != nil {
		t.Fatalf("GetPage: %v", err)
	}

	// send the fetched content back unchanged
	if _, err := c.EditPage(ctx, EditPageParams{
		Path: page.Path, Title: "Test page (edited)", Content: fetched.Content, AuthorURL: "https://telegra.ph/", ReturnContent: true,
	}); err != nil {
		t.Errorf("EditPage: %v", err)
	}
	if _, err := c.EditPageWithHTML(ctx, EditPageParams{Path: page.Path, Title: "Test page (edited with HTML)"}, htmlForTest); err != nil {
		t.Errorf("EditPageWithHTML: %v", err)
	}

	list, err := c.GetPageList(ctx, 0, 0)
	if err != nil {
		t.Fatalf("GetPageList: %v", err)
	}
	if len(list.Pages) == 0 {
		t.Fatal("GetPageList returned no pages")
	}
	now := time.Now().UTC()
	if _, err := c.GetViews(ctx, list.Pages[0].Path, GetViewsParams{}); err != nil {
		t.Errorf("GetViews (total): %v", err)
	}
	if _, err := c.GetViews(ctx, list.Pages[0].Path, GetViewsParams{
		Year: now.Year(), Month: int(now.Month()), Day: now.Day(), Hour: Ptr(now.Hour()),
	}); err != nil {
		t.Errorf("GetViews (hour): %v", err)
	}
}
