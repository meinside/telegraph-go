package telegraph

import (
	"encoding/json"
	"testing"
)

func TestNewNodesWithHTMLKeepsOnlySupportedTagsAndAttrs(t *testing.T) {
	in := `<div class="wrap"><h1 id="t">Title</h1>` +
		`<p style="x">A <a href="https://example.com" target="_blank">link</a>` +
		` and <font color="red">red</font>.</p>` +
		`<script>alert(1)</script><style>p{}</style><!-- note -->` +
		`<img src="https://example.com/i.png" alt="i"><del>old</del><svg><text>v</text></svg></div>`

	nodes, err := NewNodesWithHTML(in)
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.Marshal(nodes)
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"tag":"h3","children":["Title"]},` +
		`{"tag":"p","children":["A ",{"tag":"a","attrs":{"href":"https://example.com"},"children":["link"]}," and ","red","."]},` +
		`{"tag":"img","attrs":{"src":"https://example.com/i.png"}},` +
		`{"tag":"s","children":["old"]}]`
	assertSameJSON(t, string(got), want)
}

func TestNewNodesWithHTMLKeepsWhitespace(t *testing.T) {
	nodes, err := NewNodesWithHTML("<pre>a\n  b</pre>\nline")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(nodes)
	assertSameJSON(t, string(got), `[{"tag":"pre","children":["a\n  b"]},"\nline"]`)
}

func TestNodesToHTML(t *testing.T) {
	nodes := Nodes{
		NodeElement{Tag: "p", Children: Nodes{
			"1 < 2 & ",
			&NodeElement{Tag: "a", Attrs: map[string]string{"href": `https://example.com/?a=1&b="2"`}, Children: Nodes{"link"}},
		}},
		NodeElement{Tag: "br"},
		NodeElement{Children: Nodes{"fragment"}},
	}

	got, err := NodesToHTML(nodes)
	if err != nil {
		t.Fatal(err)
	}
	want := `<p>1 &lt; 2 &amp; <a href="https://example.com/?a=1&amp;b=&#34;2&#34;">link</a></p><br/>fragment`
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestHTMLRoundTrip(t *testing.T) {
	in := `<p>Hello <b>bold</b> <a href="https://example.com">x</a></p><ul><li>one</li><li>two</li></ul><hr/>`
	nodes, err := NewNodesWithHTML(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := NodesToHTML(nodes)
	if err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Errorf("got  %s\nwant %s", out, in)
	}
}

// html.Render writes tag and attribute names verbatim, so only names Telegraph
// accepts may reach it; anything else could inject markup.
func TestNodesToHTMLRejectsUnsupportedTagsAndAttrs(t *testing.T) {
	cases := map[string]Nodes{
		"markup in tag":       {NodeElement{Tag: `img src=x onerror=alert(1)`}},
		"markup in attr key":  {NodeElement{Tag: "p", Attrs: map[string]string{`x onmouseover=alert(1) y`: "v"}}},
		"script":              {NodeElement{Tag: "script", Children: Nodes{"alert(1)"}}},
		"event handler attr":  {NodeElement{Tag: "img", Attrs: map[string]string{"onerror": "alert(1)"}}},
		"nested unsupported":  {NodeElement{Tag: "p", Children: Nodes{NodeElement{Tag: "style"}}}},
		"uppercase supported": {NodeElement{Tag: "P"}},
	}
	for name, nodes := range cases {
		if out, err := NodesToHTML(nodes); err == nil {
			t.Errorf("%s: rendered %q, want error", name, out)
		}
	}
}

func TestNodesToHTMLRejectsUnsupportedNode(t *testing.T) {
	if _, err := NodesToHTML(Nodes{NodeElement{Tag: "p", Children: Nodes{3.14}}}); err == nil {
		t.Error("want error for float node")
	}
}

func TestNodesJSONRejectsMalformedNodes(t *testing.T) {
	var ns Nodes
	if err := json.Unmarshal([]byte(`["ok", 1]`), &ns); err == nil {
		t.Error("want error for number node")
	}
	if err := json.Unmarshal([]byte(`null`), &ns); err != nil || ns != nil {
		t.Errorf("null: ns = %v, err = %v", ns, err)
	}
	if _, err := json.Marshal(Nodes{(*NodeElement)(nil)}); err == nil {
		t.Error("want error for nil *NodeElement")
	}
}
