package telegraph

// http://telegra.ph/api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const (
	API_BASE_URL = "https://api.telegra.ph"
)

var IsVerbose bool = false

////////////////
// API resonse

type ApiResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type ApiResponseAccount struct {
	ApiResponse
	Result Account `json:"result,omitempty"`
}

type ApiResponsePage struct {
	ApiResponse
	Result Page `json:"result,omitempty"`
}

type ApiResponsePageList struct {
	ApiResponse
	Result PageList `json:"result,omitempty"`
}

type ApiResponsePageViews struct {
	ApiResponse
	Result PageViews `json:"result,omitempty"`
}

////////////////
// types

// http://telegra.ph/api#Account
type Account struct {
	ShortName  string `json:"short_name"`
	AuthorName string `json:"author_name"`
	AuthorUrl  string `json:"author_url"`

	AccessToken string `json:"access_token,omitempty"`
	AuthUrl     string `json:"auth_url,omitempty"`
	PageCount   int    `json:"page_count,omitempty"`
}

// http://telegra.ph/api#Node
type Node interface{} // XXX - can be a string, or NodeElement

// http://telegra.ph/api#NodeElement
type NodeElement struct {
	Tag      string            `json:"tag"`
	Attrs    map[string]string `json:"attrs,omitempty"`
	Children []Node            `json:"children,omitempty"`
}

// http://telegra.ph/api#Page
type Page struct {
	Path        string `json:"path"`
	Url         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorName  string `json:"author_name,omitempty"`
	AuthorUrl   string `json:"author_url,omitempty"`
	ImageUrl    string `json:"image_url,omitempty"`
	Content     []Node `json:"content,omitempty"`
	Views       int    `json:"views"`
	CanEdit     bool   `json:"can_edit,omitempty"`
}

// http://telegra.ph/api#PageList
type PageList struct {
	TotalCount int    `json:"total_count"`
	Pages      []Page `json:"pages"`
}

// http://telegra.ph/api#PageViews
type PageViews struct {
	Views int `json:"views"`
}

////////////////
// methods

// Create a new Telegraph account.
//
// shortName: 1-32 characters
// authorName: 0-128 characters (optional)
// authorUrl:  0-512 characters (optional)
//
// http://telegra.ph/api#createAccount
func CreateAccount(shortName, authorName, authorUrl string) (acc Account, err error) {
	url := fmt.Sprintf("%s/%s", API_BASE_URL, "createAccount")

	// params
	params := map[string]interface{}{
		"short_name": shortName,
	}
	if len(authorName) > 0 { // optional
		params["author_name"] = authorName
	}
	if len(authorUrl) > 0 { // optional
		params["author_url"] = authorUrl
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res ApiResponseAccount
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			} else {
				err = fmt.Errorf(res.Error)

				l("API error from %s: %s\n", url, err)
			}
		}
	}

	return Account{}, err
}

// Update information about a Telegraph account.
//
// accessToken: Access token of the Telegraph account
// shortName: 1-32 characters
// authorName: 0-128 characters (optional)
// authorUrl:  0-512 characters (optional)
//
// http://telegra.ph/api#editAccountInfo
func EditAccountInfo(accessToken, shortName, authorName, authorUrl string) (acc Account, err error) {
	url := fmt.Sprintf("%s/%s", API_BASE_URL, "editAccountInfo")

	// params
	params := map[string]interface{}{
		"access_token": accessToken,
		"short_name":   shortName,
	}
	if len(authorName) > 0 { // optional
		params["author_name"] = authorName
	}
	if len(authorUrl) > 0 { // optional
		params["author_url"] = authorUrl
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res ApiResponseAccount
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			} else {
				err = fmt.Errorf(res.Error)

				l("API error from %s: %s\n", url, err)
			}
		}
	}

	return Account{}, err
}

// Get information about a Telegraph account.
//
// accessToken: Access token of the Telegraph account
// fields: Available fields: "short_name", "author_name", "author_url", "auth_url", and "page_count" (default = ["short_name", "author_name", "author_url"])
//
// http://telegra.ph/api#getAccountInfo
func GetAccountInfo(accessToken string, fields []string) (acc Account, err error) {
	url := fmt.Sprintf("%s/%s", API_BASE_URL, "getAccountInfo")

	// params
	params := map[string]interface{}{
		"access_token": accessToken,
	}
	if len(fields) > 0 { // optional
		params["fields"] = fields
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res ApiResponseAccount
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			} else {
				err = fmt.Errorf(res.Error)

				l("API error from %s: %s\n", url, err)
			}
		}
	}

	return Account{}, err
}

// Revoke access token and generate a new one.
//
// accessToken: Access token of the Telegraph account
//
// http://telegra.ph/api#revokeAccessToken
func RevokeAccessToken(accessToken string) (acc Account, err error) {
	url := fmt.Sprintf("%s/%s", API_BASE_URL, "revokeAccessToken")

	// params
	params := map[string]interface{}{
		"access_token": accessToken,
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res ApiResponseAccount
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			} else {
				err = fmt.Errorf(res.Error)

				l("API error from %s: %s\n", url, err)
			}
		}
	}

	return Account{}, err
}

// Create a new Telegraph page.
//
// accessToken: Access token of the Telegraph account
// title: 1-256 characters
// authorName: 0-128 characters (optional)
// authorUrl:  0-512 characters (optional)
// content: Array of Node
// returnContent: return created Page object or not
//
// http://telegra.ph/api#createPage
func CreatePage(accessToken, title, authorName, authorUrl string, content []Node, returnContent bool) (page Page, err error) {
	url := fmt.Sprintf("%s/%s", API_BASE_URL, "createPage")

	// params
	params := map[string]interface{}{
		"access_token": accessToken,
		"title":        title,
		"content":      castNodes(content),
	}
	if len(authorName) > 0 { // optional
		params["author_name"] = authorName
	}
	if len(authorUrl) > 0 { // optional
		params["author_url"] = authorUrl
	}
	if returnContent { // optional
		params["return_content"] = returnContent
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res ApiResponsePage
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			} else {
				err = fmt.Errorf(res.Error)

				l("API error from %s: %s\n", url, err)
			}
		}
	}

	return Page{}, err
}

// Edit an existing Telegraph page.
//
// accessToken: Access token of the Telegraph account
// path: Path to the page
// title: 1-256 characters
// content: Array of Node
// authorName: 0-128 characters (optional)
// authorUrl:  0-512 characters (optional)
// returnContent: return edited Page object or not
//
// http://telegra.ph/api#editPage
func EditPage(accessToken, path, title string, content []Node, authorName, authorUrl string, returnContent bool) (page Page, err error) {
	url := fmt.Sprintf("%s/%s/%s", API_BASE_URL, "editPage", path)

	// params
	params := map[string]interface{}{
		"access_token": accessToken,
		"title":        title,
		"content":      castNodes(content),
	}
	if len(authorName) > 0 { // optional
		params["author_name"] = authorName
	}
	if len(authorUrl) > 0 { // optional
		params["author_url"] = authorUrl
	}
	if returnContent { // optional
		params["return_content"] = returnContent
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res ApiResponsePage
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			} else {
				err = fmt.Errorf(res.Error)

				l("API error from %s: %s\n", url, err)
			}
		}
	}

	return Page{}, err
}

// Get a Telegraph page.
//
// path: Path to the Telegraph page
// returnContent: return the Page object or not
//
// http://telegra.ph/api#getPage
func GetPage(path string, returnContent bool) (page Page, err error) {
	url := fmt.Sprintf("%s/%s/%s", API_BASE_URL, "getPage", path)

	var bytes []byte
	if bytes, err = httpPost(url, map[string]interface{}{
		"return_content": returnContent,
	}); err == nil {
		var res ApiResponsePage
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			} else {
				err = fmt.Errorf(res.Error)

				l("API error from %s: %s\n", url, err)
			}
		}
	}

	return Page{}, err
}

// Get a list of pages belonging to a Telegraph account.
//
// accessToken: Access token of the Telegraph account
// offset: Sequential number of the first page (default = 0)
// limit: Number of pages to be returned (0-200, default = 50)
//
// http://telegra.ph/api#getPageList
func GetPageList(accessToken string, offset, limit int) (list PageList, err error) {
	url := fmt.Sprintf("%s/%s", API_BASE_URL, "getPageList")

	// params
	params := map[string]interface{}{
		"access_token": accessToken,
	}
	if offset > 0 { // optional
		params["offset"] = offset
	}
	if limit >= 0 { // optional
		params["limit"] = limit
	}

	var bytes []byte
	if bytes, err = httpPost(url, params); err == nil {
		var res ApiResponsePageList
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			} else {
				err = fmt.Errorf(res.Error)

				l("API error from %s: %s\n", url, err)
			}
		}
	}

	return PageList{}, err
}

// Get the number of views for a Telegraph page.
//
// path: Path to the Telegraph page
// year: 2000-2100 (required when 'month' is passed)
// month: 1-12 (required when 'day' is passed)
// day: 1-31 (required when 'hour' is passed)
// hour: 0-24 (pass -1 if none)
//
// http://telegra.ph/api#getViews
func GetViews(path string, year, month, day, hour int) (views PageViews, err error) {
	url := fmt.Sprintf("%s/%s/%s", API_BASE_URL, "getViews", path)

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
		var res ApiResponsePageViews
		if err = json.Unmarshal(bytes, &res); err == nil {
			if res.Ok {
				return res.Result, nil
			} else {
				err = fmt.Errorf(res.Error)

				l("API error from %s: %s\n", url, err)
			}
		}
	}

	return PageViews{}, err
}

// create a new node with given string
func NewNodeWithString(str string) Node {
	return Node(str)
}

// create a new node with given element
func NewNodeWithElement(tag string, attrs map[string]string, children []Node) Node {
	return Node(NodeElement{
		Tag:      tag,
		Attrs:    attrs,
		Children: children,
	})

}

// http post (www-form urlencoded)
func httpPost(apiUrl string, params map[string]interface{}) (jsonBytes []byte, err error) {
	v("sending post request to url: %s, params: %#v", apiUrl, params)

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
	if req, err = http.NewRequest("POST", apiUrl, bytes.NewBufferString(encoded)); err == nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(encoded)))

		var res *http.Response
		client := &http.Client{}
		if res, err = client.Do(req); err == nil {
			defer res.Body.Close()

			if jsonBytes, err = ioutil.ReadAll(res.Body); err == nil {
				return jsonBytes, nil
			} else {
				l("resonse read error: %s", err.Error())
			}
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
	if IsVerbose {
		l(format, v...)
	}
}
