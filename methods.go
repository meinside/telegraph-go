package telegraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

////////////////
// methods

// CreateAccount creates a new Telegraph account.
//
// shortName: 1-32 characters
// authorName: 0-128 characters (optional)
// authorURL:  0-512 characters (optional)
//
// http://telegra.ph/api#createAccount
func (c *Client) CreateAccount(shortName, authorName, authorURL string) (acc Account, err error) {
	url := fmt.Sprintf("%s/%s", APIBaseURL, "createAccount")

	// params
	params := map[string]interface{}{
		"short_name": shortName,
	}
	if len(authorName) > 0 { // optional
		params["author_name"] = authorName
	}
	if len(authorURL) > 0 { // optional
		params["author_url"] = authorURL
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res APIResponseAccount
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			}

			err = fmt.Errorf(res.Error)

			l("API error from %s: %s\n", url, err)
		}
	}

	return Account{}, err
}

// EditAccountInfo updates information about a Telegraph account.
//
// shortName: 1-32 characters
// authorName: 0-128 characters (optional)
// authorURL:  0-512 characters (optional)
//
// http://telegra.ph/api#editAccountInfo
func (c *Client) EditAccountInfo(shortName, authorName, authorURL string) (acc Account, err error) {
	url := fmt.Sprintf("%s/%s", APIBaseURL, "editAccountInfo")

	// params
	params := map[string]interface{}{
		"access_token": c.AccessToken,
		"short_name":   shortName,
	}
	if len(authorName) > 0 { // optional
		params["author_name"] = authorName
	}
	if len(authorURL) > 0 { // optional
		params["author_url"] = authorURL
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res APIResponseAccount
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			}

			err = fmt.Errorf(res.Error)

			l("API error from %s: %s\n", url, err)
		}
	}

	return Account{}, err
}

// GetAccountInfo fetches information about a Telegraph account.
//
// fields: Available fields: "short_name", "author_name", "author_url", "auth_url", and "page_count"
// (default = ["short_name", "author_name", "author_url"])
//
// http://telegra.ph/api#getAccountInfo
func (c *Client) GetAccountInfo(fields []string) (acc Account, err error) {
	url := fmt.Sprintf("%s/%s", APIBaseURL, "getAccountInfo")

	// params
	params := map[string]interface{}{
		"access_token": c.AccessToken,
	}
	if len(fields) > 0 { // optional
		params["fields"] = fields
	} else {
		params["fields"] = []string{"short_name", "author_name", "author_url"} // default
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res APIResponseAccount
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			}

			err = fmt.Errorf(res.Error)

			l("API error from %s: %s\n", url, err)
		}
	}

	return Account{}, err
}

// RevokeAccessToken revokes access token and generate a new one.
//
// http://telegra.ph/api#revokeAccessToken
func (c *Client) RevokeAccessToken() (acc Account, err error) {
	url := fmt.Sprintf("%s/%s", APIBaseURL, "revokeAccessToken")

	// params
	params := map[string]interface{}{
		"access_token": c.AccessToken,
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res APIResponseAccount
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			}

			err = fmt.Errorf(res.Error)

			l("API error from %s: %s\n", url, err)
		}
	}

	return Account{}, err
}

// CreatePage creates a new Telegraph page.
//
// title: 1-256 characters
// authorName: 0-128 characters (optional)
// authorURL:  0-512 characters (optional)
// content: Array of Node
// returnContent: return created Page object or not
//
// http://telegra.ph/api#createPage
func (c *Client) CreatePage(title, authorName, authorURL string, content []Node, returnContent bool) (page Page, err error) {
	url := fmt.Sprintf("%s/%s", APIBaseURL, "createPage")

	// params
	params := map[string]interface{}{
		"access_token": c.AccessToken,
		"title":        title,
		"content":      castNodes(content),
	}
	if len(authorName) > 0 { // optional
		params["author_name"] = authorName
	}
	if len(authorURL) > 0 { // optional
		params["author_url"] = authorURL
	}
	if returnContent { // optional
		params["return_content"] = returnContent
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res APIResponsePage
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			}

			err = fmt.Errorf(res.Error)

			l("API error from %s: %s\n", url, err)
		}
	}

	return Page{}, err
}

// CreatePageWithHTML creates a new page with HTML.
func (c *Client) CreatePageWithHTML(title, authorName, authorURL, htmlContent string, returnContent bool) (page Page, err error) {
	nodes, err := NewNodesWithHTML(htmlContent)

	if err == nil {
		return c.CreatePage(title, authorName, authorURL, nodes, returnContent)
	}

	return Page{}, err
}

// EditPage edits an existing Telegraph page.
//
// path: Path to the page
// title: 1-256 characters
// content: Array of Node
// authorName: 0-128 characters (optional)
// authorURL:  0-512 characters (optional)
// returnContent: return edited Page object or not
//
// http://telegra.ph/api#editPage
func (c *Client) EditPage(path, title string, content []Node, authorName, authorURL string, returnContent bool) (page Page, err error) {
	url := fmt.Sprintf("%s/%s/%s", APIBaseURL, "editPage", path)

	// params
	params := map[string]interface{}{
		"access_token": c.AccessToken,
		"title":        title,
		"content":      castNodes(content),
	}
	if len(authorName) > 0 { // optional
		params["author_name"] = authorName
	}
	if len(authorURL) > 0 { // optional
		params["author_url"] = authorURL
	}
	if returnContent { // optional
		params["return_content"] = returnContent
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res APIResponsePage
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			}

			err = fmt.Errorf(res.Error)

			l("API error from %s: %s\n", url, err)
		}
	}

	return Page{}, err
}

// GetPage fetches a Telegraph page.
//
// path: Path to the Telegraph page
// returnContent: return the Page object or not
//
// http://telegra.ph/api#getPage
func (c *Client) GetPage(path string, returnContent bool) (page Page, err error) {
	url := fmt.Sprintf("%s/%s/%s", APIBaseURL, "getPage", path)

	var bytes []byte
	if bytes, err = httpPost(url, map[string]interface{}{
		"return_content": returnContent,
	}); err == nil {
		var res APIResponsePage
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			}

			err = fmt.Errorf(res.Error)

			l("API error from %s: %s\n", url, err)
		}
	}

	return Page{}, err
}

// GetPageList fetches a list of pages belonging to a Telegraph account.
//
// offset: Sequential number of the first page (default = 0)
// limit: Number of pages to be returned (0-200, default = 50)
//
// http://telegra.ph/api#getPageList
func (c *Client) GetPageList(offset, limit int) (list PageList, err error) {
	url := fmt.Sprintf("%s/%s", APIBaseURL, "getPageList")

	// params
	params := map[string]interface{}{
		"access_token": c.AccessToken,
	}
	if offset > 0 { // optional
		params["offset"] = offset
	}
	if limit >= 0 { // optional
		params["limit"] = limit
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res APIResponsePageList
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			}

			err = fmt.Errorf(res.Error)

			l("API error from %s: %s\n", url, err)
		}
	}

	return PageList{}, err
}

// GetViews fetches the number of views for a Telegraph page.
//
// path: Path to the Telegraph page
// year: 2000-2100 (required when 'month' is passed)
// month: 1-12 (required when 'day' is passed)
// day: 1-31 (required when 'hour' is passed)
// hour: 0-24 (pass -1 if none)
//
// http://telegra.ph/api#getViews
func (c *Client) GetViews(path string, year, month, day, hour int) (views PageViews, err error) {
	url := fmt.Sprintf("%s/%s/%s", APIBaseURL, "getViews", path)

	// params
	params := map[string]interface{}{}
	if year > 0 { // optional
		params["year"] = year
	}
	if month > 0 { // optional
		params["month"] = month
	}
	if day > 0 { // optional
		params["day"] = day
	}
	if hour >= 0 { // optional
		params["hour"] = hour
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res APIResponsePageViews
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			}

			err = fmt.Errorf(res.Error)

			l("API error from %s: %s\n", url, err)
		}
	}

	return PageViews{}, err
}

// NewNodeWithString creates a new node with given string.
func NewNodeWithString(str string) Node {
	return Node(str)
}

// NewNodeWithElement creates a new node with given element.
func NewNodeWithElement(tag string, attrs map[string]string, children []Node) Node {
	return Node(NodeElement{
		Tag:      tag,
		Attrs:    attrs,
		Children: children,
	})

}

// NewNodesWithHTML creates new nodes with given HTML string.
func NewNodesWithHTML(html string) ([]Node, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))

	if err == nil {
		return traverseNodes(doc.Find("body").Contents()), nil
	}

	return nil, err
}

// traverse DOM for creating new nodes
func traverseNodes(selections *goquery.Selection) []Node {
	nodes := []Node{}

	var tag string
	var attrs map[string]string
	var element NodeElement

	selections.Each(func(_ int, child *goquery.Selection) {
		for _, node := range child.Nodes {
			switch node.Type {
			case html.TextNode:
				nodes = append(nodes, node.Data) // append text
			case html.ElementNode:
				// attributes
				attrs = map[string]string{}
				for _, attr := range node.Attr {
					attrs[attr.Key] = attr.Val
				}
				// new node element
				if len(node.Namespace) > 0 {
					tag = fmt.Sprintf("%s.%s", node.Namespace, node.Data)
				} else {
					tag = node.Data
				}
				element = NodeElement{
					Tag:      tag,
					Attrs:    attrs,
					Children: traverseNodes(child.Contents()),
				}

				nodes = append(nodes, element) // append element
			default:
				continue // skip other things
			}
		}
	})

	return nodes
}

// send HTTP POST request (www-form urlencoded)
func httpPost(apiURL string, params map[string]interface{}) (jsonBytes []byte, err error) {
	v("sending post request to url: %s, params: %#v", apiURL, params)

	var js []byte
	paramValues := url.Values{}
	for key, value := range params {
		switch value.(type) {
		case string:
			paramValues[key] = []string{value.(string)}
		default:
			if js, err = json.Marshal(value); err == nil {
				paramValues[key] = []string{string(js)}
			} else {
				l("param marshalling error for: %s (%s)", key, err)

				return []byte{}, err
			}
		}
	}
	encoded := paramValues.Encode()

	var req *http.Request
	if req, err = http.NewRequest("POST", apiURL, bytes.NewBufferString(encoded)); err == nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(encoded)))

		var res *http.Response
		client := &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 300 * time.Second,
				}).Dial,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}

		res, err = client.Do(req)

		if res != nil {
			defer res.Body.Close()
		}

		if err == nil {
			if jsonBytes, err = ioutil.ReadAll(res.Body); err == nil {
				return jsonBytes, nil
			}

			l("resonse read error: %s", err.Error())
		} else {
			l("request error: %s", err.Error())
		}
	} else {
		l("building request error: %s", err.Error())
	}

	return []byte{}, err
}

// cast nodes for marshalling
func castNodes(nodes []Node) []interface{} {
	castNodes := []interface{}{}

	for _, node := range nodes {
		switch node.(type) {
		case NodeElement:
			castNodes = append(castNodes, node)
		default:
			if cast, ok := node.(string); ok {
				castNodes = append(castNodes, cast)
			} else {
				l("param casting error: %#+v", node)
			}
		}
	}

	return castNodes
}

// for logging
func l(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// for logging verbosely
func v(format string, v ...interface{}) {
	if Verbose {
		l(format, v...)
	}
}
