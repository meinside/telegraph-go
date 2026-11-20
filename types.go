package telegraph

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Account represents a Telegraph account.
//
// https://telegra.ph/api#Account
type Account struct {
	ShortName  string `json:"short_name"`
	AuthorName string `json:"author_name"`
	AuthorURL  string `json:"author_url"`

	// AccessToken is only returned by createAccount and revokeAccessToken.
	AccessToken string `json:"access_token,omitempty"`
	// AuthURL is valid for one use and for 5 minutes only.
	AuthURL   string `json:"auth_url,omitempty"`
	PageCount int    `json:"page_count,omitempty"`
}

// AccountField names a field that GetAccountInfo can return.
type AccountField string

// Fields accepted by GetAccountInfo.
const (
	FieldShortName  AccountField = "short_name"
	FieldAuthorName AccountField = "author_name"
	FieldAuthorURL  AccountField = "author_url"
	FieldAuthURL    AccountField = "auth_url"
	FieldPageCount  AccountField = "page_count"
)

// Node is a DOM node: a string (text node) or a NodeElement.
// A *NodeElement is also accepted when sending content.
//
// https://telegra.ph/api#Node
type Node any

// Nodes is a list of Node.
//
// Decoding turns text nodes into string and element nodes into NodeElement,
// so content fetched from the API can be type-switched and sent back as is.
// Encoding fails on any value that is not a string, NodeElement or *NodeElement.
type Nodes []Node

// NodeElement represents a DOM element node.
//
// https://telegra.ph/api#NodeElement
type NodeElement struct {
	Tag      string            `json:"tag"`
	Attrs    map[string]string `json:"attrs,omitempty"`
	Children Nodes             `json:"children,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (ns Nodes) MarshalJSON() ([]byte, error) {
	out := make([]any, len(ns))
	for i, n := range ns {
		switch n := n.(type) {
		case string, NodeElement:
			out[i] = n
		case *NodeElement:
			if n == nil {
				return nil, fmt.Errorf("telegraph: node #%d is a nil *NodeElement", i)
			}
			out[i] = n
		default:
			return nil, fmt.Errorf("telegraph: node #%d has unsupported type %T (want string or NodeElement)", i, n)
		}
	}
	return json.Marshal(out)
}

// UnmarshalJSON implements json.Unmarshaler.
func (ns *Nodes) UnmarshalJSON(b []byte) error {
	var raws []json.RawMessage
	if err := json.Unmarshal(b, &raws); err != nil {
		return err
	}
	if raws == nil { // null
		*ns = nil
		return nil
	}

	out := make(Nodes, 0, len(raws))
	for i, raw := range raws {
		raw = bytes.TrimSpace(raw)
		if len(raw) == 0 {
			return fmt.Errorf("telegraph: node #%d is empty", i)
		}
		switch raw[0] {
		case '"':
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return err
			}
			out = append(out, s)
		case '{':
			var e NodeElement
			if err := json.Unmarshal(raw, &e); err != nil {
				return err
			}
			out = append(out, e)
		default:
			return fmt.Errorf("telegraph: node #%d is neither a string nor an object: %s", i, raw)
		}
	}
	*ns = out
	return nil
}

// Page represents a page on Telegraph.
//
// https://telegra.ph/api#Page
type Page struct {
	Path        string `json:"path"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorName  string `json:"author_name,omitempty"`
	AuthorURL   string `json:"author_url,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Content     Nodes  `json:"content,omitempty"`
	Views       int    `json:"views"`
	// CanEdit is only returned by createPage, editPage and getPageList, not by getPage.
	CanEdit bool `json:"can_edit,omitempty"`
}

// PageList represents a list of pages belonging to an account, most recently created first.
//
// https://telegra.ph/api#PageList
type PageList struct {
	TotalCount int    `json:"total_count"`
	Pages      []Page `json:"pages"`
}

// PageViews represents the number of page views for a page.
//
// https://telegra.ph/api#PageViews
type PageViews struct {
	Views int `json:"views"`
}

// Ptr returns a pointer to v, for optional parameters such as
// EditAccountInfoParams.AuthorName or GetViewsParams.Hour.
func Ptr[T any](v T) *T {
	return &v
}
