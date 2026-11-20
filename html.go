package telegraph

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// allowedTags are the tags the API accepts in NodeElement.Tag.
// https://telegra.ph/api#NodeElement
var allowedTags = map[string]bool{
	"a": true, "aside": true, "b": true, "blockquote": true, "br": true,
	"code": true, "em": true, "figcaption": true, "figure": true, "h3": true,
	"h4": true, "hr": true, "i": true, "iframe": true, "img": true, "li": true,
	"ol": true, "p": true, "pre": true, "s": true, "strong": true, "u": true,
	"ul": true, "video": true,
}

// allowedAttrs are the attributes the API accepts in NodeElement.Attrs.
var allowedAttrs = map[string]bool{"href": true, "src": true}

// tagAliases maps unsupported tags to the closest supported one.
var tagAliases = map[string]string{
	"h1": "h3", "h2": "h3", "h5": "h4", "h6": "h4",
	"del": "s", "strike": "s",
}

// droppedTags are removed together with their content; their text is not article text.
var droppedTags = map[string]bool{
	"head": true, "noscript": true, "script": true, "style": true,
	"template": true, "title": true,
}

// NewNodesWithHTML converts an HTML fragment into nodes the API accepts.
//
//   - Supported tags are kept, with only href and src attributes.
//   - h1/h2 become h3, h5/h6 become h4, del/strike become s.
//   - script, style, noscript, template and SVG/MathML elements are removed with their content.
//   - Other unsupported tags (div, span, font, ...) are unwrapped: their children are kept.
//
// Text, whitespace included, is passed through unchanged.
func NewNodesWithHTML(htmlContent string) (Nodes, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("telegraph: parse HTML: %w", err)
	}
	body := findElement(doc, atom.Body)
	if body == nil {
		return Nodes{}, nil
	}
	return convertChildren(body), nil
}

func findElement(n *html.Node, a atom.Atom) *html.Node {
	if n.Type == html.ElementNode && n.DataAtom == a {
		return n
	}
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		if found := findElement(ch, a); found != nil {
			return found
		}
	}
	return nil
}

func convertChildren(n *html.Node) Nodes {
	out := Nodes{}
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		out = append(out, convertNode(ch)...)
	}
	return out
}

func convertNode(n *html.Node) Nodes {
	switch n.Type {
	case html.TextNode:
		return Nodes{n.Data}
	case html.ElementNode:
		if n.Namespace != "" || droppedTags[n.Data] { // SVG/MathML, script, ...
			return nil
		}
		tag := n.Data
		if alias, ok := tagAliases[tag]; ok {
			tag = alias
		}
		children := convertChildren(n)
		if !allowedTags[tag] {
			return children
		}

		el := NodeElement{Tag: tag}
		for _, attr := range n.Attr {
			if attr.Namespace == "" && allowedAttrs[attr.Key] {
				if el.Attrs == nil {
					el.Attrs = map[string]string{}
				}
				el.Attrs[attr.Key] = attr.Val
			}
		}
		if len(children) > 0 {
			el.Children = children
		}
		return Nodes{el}
	default: // comments, doctype
		return nil
	}
}

// NodesToHTML renders nodes, e.g. Page.Content, as an HTML fragment.
// Text and attribute values are escaped; attributes are written in key order.
//
// Only the tags and attributes Telegraph accepts are rendered; any other is an
// error, because html.Render writes tag and attribute names verbatim. URLs in
// href and src are written as given: check their scheme (e.g. javascript:)
// before showing nodes from an untrusted source.
func NodesToHTML(nodes Nodes) (string, error) {
	hns, err := toHTMLNodes(nodes)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, hn := range hns {
		if err := html.Render(&sb, hn); err != nil {
			return "", fmt.Errorf("telegraph: render HTML: %w", err)
		}
	}
	return sb.String(), nil
}

func toHTMLNodes(nodes Nodes) ([]*html.Node, error) {
	var out []*html.Node
	for i, n := range nodes {
		var el NodeElement
		switch n := n.(type) {
		case string:
			out = append(out, &html.Node{Type: html.TextNode, Data: n})
			continue
		case NodeElement:
			el = n
		case *NodeElement:
			if n == nil {
				return nil, fmt.Errorf("telegraph: node #%d is a nil *NodeElement", i)
			}
			el = *n
		default:
			return nil, fmt.Errorf("telegraph: node #%d has unsupported type %T (want string or NodeElement)", i, n)
		}

		children, err := toHTMLNodes(el.Children)
		if err != nil {
			return nil, err
		}
		if el.Tag == "" { // a tag-less element is a fragment, as in the API's nodeToDom example
			out = append(out, children...)
			continue
		}
		if !allowedTags[el.Tag] {
			return nil, fmt.Errorf("telegraph: node #%d: tag %q is not supported by Telegraph", i, el.Tag)
		}
		hn := &html.Node{Type: html.ElementNode, Data: el.Tag, DataAtom: atom.Lookup([]byte(el.Tag))}
		for _, key := range slices.Sorted(maps.Keys(el.Attrs)) {
			if !allowedAttrs[key] {
				return nil, fmt.Errorf("telegraph: node #%d: attribute %q is not supported by Telegraph", i, key)
			}
			hn.Attr = append(hn.Attr, html.Attribute{Key: key, Val: el.Attrs[key]})
		}
		for _, ch := range children {
			hn.AppendChild(ch)
		}
		out = append(out, hn)
	}
	return out, nil
}

// NewNodeWithString creates a text node.
func NewNodeWithString(str string) Node {
	return Node(str)
}

// NewNodeWithElement creates an element node.
func NewNodeWithElement(tag string, attrs map[string]string, children []Node) Node {
	return Node(NodeElement{
		Tag:      tag,
		Attrs:    attrs,
		Children: children,
	})
}
