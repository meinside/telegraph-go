package telegraph

// Various types

////////////////
// API resonse

// APIResponse struct
type APIResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// APIResponseAccount struct
type APIResponseAccount struct {
	APIResponse
	Result Account `json:"result,omitempty"`
}

// APIResponsePage struct
type APIResponsePage struct {
	APIResponse
	Result Page `json:"result,omitempty"`
}

// APIResponsePageList struct
type APIResponsePageList struct {
	APIResponse
	Result PageList `json:"result,omitempty"`
}

// APIResponsePageViews struct
type APIResponsePageViews struct {
	APIResponse
	Result PageViews `json:"result,omitempty"`
}

////////////////
// types

// Account type
//
// http://telegra.ph/api#Account
type Account struct {
	ShortName  string `json:"short_name"`
	AuthorName string `json:"author_name"`
	AuthorURL  string `json:"author_url"`

	AccessToken string `json:"access_token,omitempty"`
	AuthURL     string `json:"auth_url,omitempty"`
	PageCount   int    `json:"page_count,omitempty"`
}

// Node type
//
// http://telegra.ph/api#Node
type Node interface{} // XXX - can be a string, or NodeElement

// NodeElement type
//
// http://telegra.ph/api#NodeElement
type NodeElement struct {
	Tag      string            `json:"tag"`
	Attrs    map[string]string `json:"attrs,omitempty"`
	Children []Node            `json:"children,omitempty"`
}

// Page type
//
// http://telegra.ph/api#Page
type Page struct {
	Path        string `json:"path"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorName  string `json:"author_name,omitempty"`
	AuthorURL   string `json:"author_url,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Content     []Node `json:"content,omitempty"`
	Views       int    `json:"views"`
	CanEdit     bool   `json:"can_edit,omitempty"`
}

// PageList type
//
// http://telegra.ph/api#PageList
type PageList struct {
	TotalCount int    `json:"total_count"`
	Pages      []Page `json:"pages"`
}

// PageViews type
//
// http://telegra.ph/api#PageViews
type PageViews struct {
	Views int `json:"views"`
}
