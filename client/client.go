package client

import (
	"github.com/meinside/telegraph-go"
)

type Client struct {
	telegraph.Account
}

// Create a new Telegraph client.
func Create(shortName, authorName, authorUrl string) (*Client, error) {
	if client, err := telegraph.CreateAccount(shortName, authorName, authorUrl); err == nil {
		return &Client{client}, nil
	} else {
		return nil, err
	}
}

// Load a Telegraph client with access token.
func Load(accessToken string) (*Client, error) {
	if client, err := telegraph.GetAccountInfo(accessToken, []string{"short_name", "author_name", "author_url"}); err == nil {
		client.AccessToken = accessToken
		return &Client{client}, nil
	} else {
		return nil, err
	}
}

// Edit account info.
func (c *Client) EditAccountInfo(shortName, authorName, authorUrl string) (acc telegraph.Account, err error) {
	return telegraph.EditAccountInfo(c.AccessToken, shortName, authorName, authorUrl)
}

// Get account info.
func (c *Client) GetAccountInfo(fields []string) (acc telegraph.Account, err error) {
	return telegraph.GetAccountInfo(c.AccessToken, fields)
}

// Revoke access token.
func (c *Client) RevokeAccessToken() (acc telegraph.Account, err error) {
	return telegraph.RevokeAccessToken(c.AccessToken)
}

// Create a new page.
func (c *Client) CreatePage(title, authorName, authorUrl string, content []telegraph.Node, returnContent bool) (page telegraph.Page, err error) {
	return telegraph.CreatePage(c.AccessToken, title, authorName, authorUrl, content, returnContent)
}

// Edit a page.
func (c *Client) EditPage(path, title string, content []telegraph.Node, authorName, authorUrl string, returnContent bool) (page telegraph.Page, err error) {
	return telegraph.EditPage(c.AccessToken, path, title, content, authorName, authorUrl, returnContent)
}

// Get a page.
func (c *Client) GetPage(path string, returnContent bool) (page telegraph.Page, err error) {
	return telegraph.GetPage(path, returnContent)
}

// Get page list.
func (c *Client) GetPageList(offset, limit int) (list telegraph.PageList, err error) {
	return telegraph.GetPageList(c.AccessToken, offset, limit)
}

// Get views of a page.
func (c *Client) GetViews(path string, year, month, day, hour int) (views telegraph.PageViews, err error) {
	return telegraph.GetViews(path, year, month, day, hour)
}
