package telegraph

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"unicode/utf8"
)

// CreateAccountParams are the parameters of CreateAccount.
type CreateAccountParams struct {
	ShortName  string // required, 1-32 characters
	AuthorName string // optional, 0-128 characters
	AuthorURL  string // optional, 0-512 characters
}

// CreateAccount creates a new Telegraph account. The returned Account carries its access token.
// It does not change the client's own token; use Create for a client bound to the new account.
//
// https://telegra.ph/api#createAccount
func (c *Client) CreateAccount(ctx context.Context, p CreateAccountParams) (Account, error) {
	if err := firstErr(
		checkLen("short_name", p.ShortName, 1, 32),
		checkLen("author_name", p.AuthorName, 0, 128),
		checkLen("author_url", p.AuthorURL, 0, 512),
	); err != nil {
		return Account{}, err
	}

	v := url.Values{}
	v.Set("short_name", p.ShortName)
	setIfNotEmpty(v, "author_name", p.AuthorName)
	setIfNotEmpty(v, "author_url", p.AuthorURL)

	return call[Account](ctx, c, "createAccount", "", v)
}

// EditAccountInfoParams are the parameters of EditAccountInfo.
// Only non-nil fields are sent; use Ptr("") to clear AuthorName or AuthorURL.
type EditAccountInfoParams struct {
	ShortName  *string // 1-32 characters
	AuthorName *string // 0-128 characters
	AuthorURL  *string // 0-512 characters
}

// EditAccountInfo updates the account's information.
//
// https://telegra.ph/api#editAccountInfo
func (c *Client) EditAccountInfo(ctx context.Context, p EditAccountInfoParams) (Account, error) {
	v, err := c.tokenValues()
	if err != nil {
		return Account{}, err
	}
	if p.ShortName != nil {
		if err := checkLen("short_name", *p.ShortName, 1, 32); err != nil {
			return Account{}, err
		}
		v.Set("short_name", *p.ShortName)
	}
	if p.AuthorName != nil {
		if err := checkLen("author_name", *p.AuthorName, 0, 128); err != nil {
			return Account{}, err
		}
		v.Set("author_name", *p.AuthorName)
	}
	if p.AuthorURL != nil {
		if err := checkLen("author_url", *p.AuthorURL, 0, 512); err != nil {
			return Account{}, err
		}
		v.Set("author_url", *p.AuthorURL)
	}

	return call[Account](ctx, c, "editAccountInfo", "", v)
}

// GetAccountInfo fetches the account's information.
// Without fields the API returns short_name, author_name and author_url.
//
// https://telegra.ph/api#getAccountInfo
func (c *Client) GetAccountInfo(ctx context.Context, fields ...AccountField) (Account, error) {
	v, err := c.tokenValues()
	if err != nil {
		return Account{}, err
	}
	if len(fields) > 0 {
		js, err := json.Marshal(fields)
		if err != nil {
			return Account{}, invalidParamf("fields: %v", err)
		}
		v.Set("fields", string(js))
	}

	return call[Account](ctx, c, "getAccountInfo", "", v)
}

// RevokeAccessToken revokes the access token and generates a new one,
// which the client uses from then on. The returned Account carries the new
// access_token and auth_url.
//
// https://telegra.ph/api#revokeAccessToken
func (c *Client) RevokeAccessToken(ctx context.Context) (Account, error) {
	v, err := c.tokenValues()
	if err != nil {
		return Account{}, err
	}

	account, err := call[Account](ctx, c, "revokeAccessToken", "", v)
	if err == nil && account.AccessToken != "" {
		c.setAccessToken(account.AccessToken)
	}
	return account, err
}

// CreatePageParams are the parameters of CreatePage.
type CreatePageParams struct {
	Title         string // required, 1-256 characters
	AuthorName    string // optional, 0-128 characters
	AuthorURL     string // optional, 0-512 characters
	Content       Nodes  // required, up to 64 KB
	ReturnContent bool   // return Content in the resulting Page
}

// CreatePage creates a new Telegraph page.
//
// https://telegra.ph/api#createPage
func (c *Client) CreatePage(ctx context.Context, p CreatePageParams) (Page, error) {
	v, err := c.tokenValues()
	if err != nil {
		return Page{}, err
	}
	if err := c.setPageValues(v, p.Title, p.AuthorName, p.AuthorURL, p.Content, p.ReturnContent); err != nil {
		return Page{}, err
	}

	return call[Page](ctx, c, "createPage", "", v)
}

// CreatePageWithHTML creates a new page whose content is converted from htmlContent
// with NewNodesWithHTML. Any p.Content is replaced.
func (c *Client) CreatePageWithHTML(ctx context.Context, p CreatePageParams, htmlContent string) (Page, error) {
	nodes, err := NewNodesWithHTML(htmlContent)
	if err != nil {
		return Page{}, err
	}
	p.Content = nodes
	return c.CreatePage(ctx, p)
}

// EditPageParams are the parameters of EditPage.
type EditPageParams struct {
	Path          string // required, path to the page
	Title         string // required, 1-256 characters
	AuthorName    string // optional, 0-128 characters
	AuthorURL     string // optional, 0-512 characters
	Content       Nodes  // required, up to 64 KB
	ReturnContent bool   // return Content in the resulting Page
}

// EditPage edits an existing Telegraph page.
//
// https://telegra.ph/api#editPage
func (c *Client) EditPage(ctx context.Context, p EditPageParams) (Page, error) {
	v, err := c.tokenValues()
	if err != nil {
		return Page{}, err
	}
	if p.Path == "" {
		return Page{}, invalidParamf("path is required")
	}
	if err := c.setPageValues(v, p.Title, p.AuthorName, p.AuthorURL, p.Content, p.ReturnContent); err != nil {
		return Page{}, err
	}

	return call[Page](ctx, c, "editPage", p.Path, v)
}

// EditPageWithHTML edits an existing page with content converted from htmlContent
// with NewNodesWithHTML. Any p.Content is replaced.
func (c *Client) EditPageWithHTML(ctx context.Context, p EditPageParams, htmlContent string) (Page, error) {
	nodes, err := NewNodesWithHTML(htmlContent)
	if err != nil {
		return Page{}, err
	}
	p.Content = nodes
	return c.EditPage(ctx, p)
}

// GetPage fetches a Telegraph page by its path (e.g. "Sample-Page-12-15").
// No access token is sent: the API ignores it here and never returns Page.CanEdit
// (checked against the live API); CreatePage, EditPage and GetPageList do return it.
//
// https://telegra.ph/api#getPage
func (c *Client) GetPage(ctx context.Context, path string, returnContent bool) (Page, error) {
	if path == "" {
		return Page{}, invalidParamf("path is required")
	}

	v := url.Values{}
	if returnContent {
		v.Set("return_content", "true")
	}

	return call[Page](ctx, c, "getPage", path, v)
}

// GetPageList fetches pages of the account, most recently created first.
// offset is the sequential number of the first page; limit is 1-200,
// or 0 for the API's default of 50.
//
// https://telegra.ph/api#getPageList
func (c *Client) GetPageList(ctx context.Context, offset, limit int) (PageList, error) {
	v, err := c.tokenValues()
	if err != nil {
		return PageList{}, err
	}
	if offset < 0 {
		return PageList{}, invalidParamf("offset must not be negative: %d", offset)
	}
	if limit < 0 || limit > 200 {
		return PageList{}, invalidParamf("limit must be 0-200: %d", limit)
	}
	if offset > 0 {
		v.Set("offset", strconv.Itoa(offset))
	}
	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}

	return call[PageList](ctx, c, "getPageList", "", v)
}

// GetViewsParams narrows GetViews to a period. The zero value asks for total views.
// Each finer unit requires the coarser one: Month needs Year, Day needs Month, Hour needs Day.
type GetViewsParams struct {
	Year  int  // 2000-2100, 0 = not set
	Month int  // 1-12, 0 = not set
	Day   int  // 1-31, 0 = not set
	Hour  *int // 0-24, nil = not set (0 is a valid hour, hence the pointer)
}

// GetViews fetches the number of views for a Telegraph page.
//
// https://telegra.ph/api#getViews
func (c *Client) GetViews(ctx context.Context, path string, p GetViewsParams) (PageViews, error) {
	if path == "" {
		return PageViews{}, invalidParamf("path is required")
	}
	if err := p.validate(); err != nil {
		return PageViews{}, err
	}

	v := url.Values{}
	if p.Year != 0 {
		v.Set("year", strconv.Itoa(p.Year))
	}
	if p.Month != 0 {
		v.Set("month", strconv.Itoa(p.Month))
	}
	if p.Day != 0 {
		v.Set("day", strconv.Itoa(p.Day))
	}
	if p.Hour != nil {
		v.Set("hour", strconv.Itoa(*p.Hour))
	}

	return call[PageViews](ctx, c, "getViews", path, v)
}

func (p GetViewsParams) validate() error {
	switch {
	case p.Year != 0 && (p.Year < 2000 || p.Year > 2100):
		return invalidParamf("year must be 2000-2100: %d", p.Year)
	case p.Month != 0 && (p.Month < 1 || p.Month > 12):
		return invalidParamf("month must be 1-12: %d", p.Month)
	case p.Day != 0 && (p.Day < 1 || p.Day > 31):
		return invalidParamf("day must be 1-31: %d", p.Day)
	case p.Hour != nil && (*p.Hour < 0 || *p.Hour > 24):
		return invalidParamf("hour must be 0-24: %d", *p.Hour)
	case p.Month != 0 && p.Year == 0:
		return invalidParamf("month requires year")
	case p.Day != 0 && p.Month == 0:
		return invalidParamf("day requires month")
	case p.Hour != nil && p.Day == 0:
		return invalidParamf("hour requires day")
	}
	return nil
}

// tokenValues returns form values holding the access token, or an error when there is none.
func (c *Client) tokenValues() (url.Values, error) {
	token := c.AccessToken()
	if token == "" {
		return nil, invalidParamf("access token is required")
	}
	return url.Values{"access_token": {token}}, nil
}

func (c *Client) setPageValues(v url.Values, title, authorName, authorURL string, content Nodes, returnContent bool) error {
	if err := firstErr(
		checkLen("title", title, 1, 256),
		checkLen("author_name", authorName, 0, 128),
		checkLen("author_url", authorURL, 0, 512),
	); err != nil {
		return err
	}
	if len(content) == 0 {
		return invalidParamf("content is required")
	}
	js, err := json.Marshal(content)
	if err != nil {
		return invalidParamf("content: %v", err)
	}

	v.Set("title", title)
	v.Set("content", string(js))
	setIfNotEmpty(v, "author_name", authorName)
	setIfNotEmpty(v, "author_url", authorURL)
	if returnContent {
		v.Set("return_content", "true")
	}
	return nil
}

// checkLen checks the length of s in characters, as the API documents its limits.
func checkLen(name, s string, minLen, maxLen int) error {
	if n := utf8.RuneCountInString(s); n < minLen || n > maxLen {
		return invalidParamf("%s must be %d-%d characters, got %d", name, minLen, maxLen, n)
	}
	return nil
}

func setIfNotEmpty(v url.Values, key, value string) {
	if value != "" {
		v.Set(key, value)
	}
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
