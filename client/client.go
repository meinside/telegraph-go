package client

import (
	"github.com/meinside/telegraph-go"
)

// Client struct
type Client struct {
	telegraph.Account
}

// Create creates a new Telegraph client.
func Create(shortName, authorName, authorURL string) (*Client, error) {
	client, err := telegraph.CreateAccount(shortName, authorName, authorURL)

	if err == nil {
		return &Client{client}, nil
	}

	return nil, err
}

// Load a Telegraph client with access token.
func Load(accessToken string) (*Client, error) {
	client, err := telegraph.GetAccountInfo(accessToken, []string{"short_name", "author_name", "author_url"})

	if err == nil {
		client.AccessToken = accessToken
		return &Client{client}, nil
	}

	return nil, err
}

// EditAccountInfo edits account info.
func (c *Client) EditAccountInfo(shortName, authorName, authorURL string) (acc telegraph.Account, err error) {
	return telegraph.EditAccountInfo(c.AccessToken, shortName, authorName, authorURL)
}

// GetAccountInfo fetches account info.
func (c *Client) GetAccountInfo(fields []string) (acc telegraph.Account, err error) {
	return telegraph.GetAccountInfo(c.AccessToken, fields)
}

// RevokeAccessToken revokes access token.
func (c *Client) RevokeAccessToken() (acc telegraph.Account, err error) {
	return telegraph.RevokeAccessToken(c.AccessToken)
}

// CreatePage creates a new page.
func (c *Client) CreatePage(title, authorName, authorURL string, content []telegraph.Node, returnContent bool) (page telegraph.Page, err error) {
	return telegraph.CreatePage(c.AccessToken, title, authorName, authorURL, content, returnContent)
}

// CreatePageWithHTML creates a new page with HTML.
func (c *Client) CreatePageWithHTML(title, authorName, authorURL, htmlContent string, returnContent bool) (page telegraph.Page, err error) {
	nodes, err := telegraph.NewNodesWithHTML(htmlContent)

	if err == nil {
		return telegraph.CreatePage(c.AccessToken, title, authorName, authorURL, nodes, returnContent)
	}

	return telegraph.Page{}, err
}

// EditPage edits a page.
func (c *Client) EditPage(path, title string, content []telegraph.Node, authorName, authorURL string, returnContent bool) (page telegraph.Page, err error) {
	return telegraph.EditPage(c.AccessToken, path, title, content, authorName, authorURL, returnContent)
}

// GetPage fetches a page.
func (c *Client) GetPage(path string, returnContent bool) (page telegraph.Page, err error) {
	return telegraph.GetPage(path, returnContent)
}

// GetPageList fetches page list.
func (c *Client) GetPageList(offset, limit int) (list telegraph.PageList, err error) {
	return telegraph.GetPageList(c.AccessToken, offset, limit)
}

// GetViews fetches views of a page.
func (c *Client) GetViews(path string, year, month, day, hour int) (views telegraph.PageViews, err error) {
	return telegraph.GetViews(path, year, month, day, hour)
}
