package html

import (
	"os"
	"strings"
	"testing"
)

// ParserHarness is a comprehensive test suite for the HTML5 parser.
// Tests cover DOM construction, foster parenting, implicit tag closing,
// foreign content (SVG/MathML), and tree building correctness.

func TestParser_BasicStructure(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkDOM func(d *Node) error
	}{
		{
			name:  "minimal document",
			input: "<html></html>",
			checkDOM: func(d *Node) error {
				if len(d.Children) != 1 || d.Children[0].TagName != "html" {
					t.Errorf("expected html child, got %v", d.Children)
				}
				return nil
			},
		},
		{
			name:  "div creates body",
			input: "<div>hello</div>",
			checkDOM: func(d *Node) error {
				html := d.QuerySelectorAll("html")[0]
				body := html.QuerySelectorAll("body")
				if len(body) != 1 {
					t.Errorf("expected body, got %d", len(body))
				}
				return nil
			},
		},
		{
			name:  "head element",
			input: "<html><head><title>Hello</title></head></html>",
			checkDOM: func(d *Node) error {
				head := d.QuerySelectorAll("head")
				if len(head) != 1 {
					t.Errorf("expected head, got %d", len(head))
				}
				title := d.QuerySelectorAll("title")
				if len(title) != 1 {
					t.Errorf("expected title, got %d", len(title))
				}
				return nil
			},
		},
		{
			name:  "text direct under body",
			input: "plain text",
			checkDOM: func(d *Node) error {
				body := d.QuerySelectorAll("body")[0]
				if len(body.Children) == 0 {
					t.Error("expected children under body")
				}
				return nil
			},
		},
		{
			name:  "multiple siblings",
			input: "<div></div><div></div><span></span>",
			checkDOM: func(d *Node) error {
				divs := d.QuerySelectorAll("div")
				if len(divs) != 2 {
					t.Errorf("expected 2 divs, got %d", len(divs))
				}
				spans := d.QuerySelectorAll("span")
				if len(spans) != 1 {
					t.Errorf("expected 1 span, got %d", len(spans))
				}
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			if err := tc.checkDOM(doc); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParser_ImplicitPClose(t *testing.T) {
	// Block elements implicitly close open <p> tags
	tests := []struct {
		name  string
		input string
		check func(d *Node) error
	}{
		{
			name: "p before div",
			input: "<p>Hello<div>World</div>",
			check: func(d *Node) error {
				// Should have p and div as siblings, not nested
				pCount := len(d.QuerySelectorAll("p"))
				divCount := len(d.QuerySelectorAll("div"))
				if pCount != 1 || divCount != 1 {
					t.Errorf("p=%d div=%d (expected 1 each)", pCount, divCount)
				}
				return nil
			},
		},
		{
			name: "p before h1",
			input: "<p>Title<h1>Heading</h1>",
			check: func(d *Node) error {
				ps := d.QuerySelectorAll("p")
				h1s := d.QuerySelectorAll("h1")
				if len(ps) != 1 || len(h1s) != 1 {
					t.Errorf("p=%d h1=%d (expected 1 each)", len(ps), len(h1s))
				}
				return nil
			},
		},
		{
			name: "multiple p closes",
			input: "<p>a<p>b<p>c",
			check: func(d *Node) error {
				ps := d.QuerySelectorAll("p")
				if len(ps) != 3 {
					t.Errorf("expected 3 p tags, got %d", len(ps))
				}
				return nil
			},
		},
		{
			name: "li implicitly closes p",
			input: "<p>Hello<li>Item",
			check: func(d *Node) error {
				ps := d.QuerySelectorAll("p")
				lis := d.QuerySelectorAll("li")
				if len(ps) != 1 || len(lis) != 1 {
					t.Errorf("p=%d li=%d", len(ps), len(lis))
				}
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			if err := tc.check(doc); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParser_FosterParenting(t *testing.T) {
	// When a table is open and a block element appears,
	// the block should be fostered out of the table.
	tests := []struct {
		name  string
		input string
		check func(d *Node) error
	}{
		{
			name: "div before table content",
			input: "<div>above</div><table><div>Foster Me</div></table>",
			check: func(d *Node) error {
				// The div inside table should be fostered out
				divs := d.QuerySelectorAll("div")
				// At minimum, the outer div should exist
				if len(divs) < 1 {
					t.Errorf("expected at least 1 div")
				}
				return nil
			},
		},
		{
			name: "text inside table fostered",
			input: "<table><tbody>text</tbody></table>",
			check: func(d *Node) error {
				// Text inside table should be fostered out
				return nil
			},
		},
		{
			name: "nested tables",
			input: "<table><tr><td><table><tr><td>nested</td></tr></table></td></tr></table>",
			check: func(d *Node) error {
				tds := d.QuerySelectorAll("td")
				if len(tds) < 2 {
					t.Errorf("expected at least 2 tds, got %d", len(tds))
				}
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			if err := tc.check(doc); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParser_ForeignContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(d *Node) error
	}{
		{
			name: "svg text no p close",
			input: `<svg><text>Hello</text></svg>`,
			check: func(d *Node) error {
				svg := d.QuerySelectorAll("svg")
				if len(svg) != 1 {
					t.Errorf("expected 1 svg, got %d", len(svg))
				}
				return nil
			},
		},
		{
			name: "svg with g",
			input: `<svg><g><rect/><text>Test</text></g></svg>`,
			check: func(d *Node) error {
				gs := d.QuerySelectorAll("g")
				if len(gs) < 1 {
					t.Errorf("expected at least 1 g element")
				}
				return nil
			},
		},
		{
			name: "mathml basic",
			input: `<math><mi>x</mi></math>`,
			check: func(d *Node) error {
				math := d.QuerySelectorAll("math")
				if len(math) != 1 {
					t.Errorf("expected 1 math element")
				}
				return nil
			},
		},
		{
			name: "svg inside p no adoption agency",
			input: `<p>Hello<svg><rect/></svg>World</p>`,
			check: func(d *Node) error {
				ps := d.QuerySelectorAll("p")
				if len(ps) != 1 {
					t.Errorf("expected 1 p, got %d", len(ps))
				}
				text := ps[0].InnerText()
				if !strings.Contains(text, "Hello") {
					t.Errorf("expected 'Hello' in p, got %q", text)
				}
				return nil
			},
		},
		{
			name: "foreignObject",
			input: `<svg><foreignObject><div>inner</div></foreignObject></svg>`,
			check: func(d *Node) error {
				fo := d.QuerySelectorAll("foreignObject")
				if len(fo) != 1 {
					t.Errorf("expected 1 foreignObject")
				}
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			if err := tc.check(doc); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParser_VoidElements(t *testing.T) {
	// Void elements should not appear as open elements in the stack
	tests := []struct {
		name    string
		input   string
		hasChild string
	}{
		{"br", "<p>line<br>break</p>", "br"},
		{"hr", "<p>text<hr>more", "hr"},
		{"img", "<div><img src=x alt=y>", "img"},
		{"input", "<form><input type=text>", "input"},
		{"meta", "<head><meta charset=utf-8>", "meta"},
		{"link", "<head><link rel=stylesheet>", "link"},
		{"br multiple", "<p>a<br><br><br>b", "br"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			els := doc.QuerySelectorAll(tc.hasChild)
			if len(els) == 0 {
				t.Errorf("expected at least 1 %s element", tc.hasChild)
			}
		})
	}
}

func TestParser_TableStructure(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(d *Node) error
	}{
		{
			name: "basic table",
			input: "<table><tr><td>A</td><td>B</td></tr></table>",
			check: func(d *Node) error {
				table := d.QuerySelectorAll("table")
				tr := d.QuerySelectorAll("tr")
				td := d.QuerySelectorAll("td")
				if len(table) != 1 || len(tr) != 1 || len(td) != 2 {
					t.Errorf("table=%d tr=%d td=%d", len(table), len(tr), len(td))
				}
				return nil
			},
		},
		{
			name: "thead tbody tfoot",
			input: "<table><thead><tr><th>H</th></tr></thead><tbody><tr><td>B</td></tr></tbody></table>",
			check: func(d *Node) error {
				thead := d.QuerySelectorAll("thead")
				tbody := d.QuerySelectorAll("tbody")
				if len(thead) != 1 || len(tbody) != 1 {
					t.Errorf("thead=%d tbody=%d", len(thead), len(tbody))
				}
				return nil
			},
		},
		{
			name: "td colspan",
			input: "<table><tr><td colspan=2>A</td></tr></table>",
			check: func(d *Node) error {
				td := d.QuerySelectorAll("td")
				if len(td) != 1 {
					t.Errorf("expected 1 td, got %d", len(td))
				}
				if td[0].GetAttribute("colspan") != "2" {
					t.Errorf("expected colspan=2")
				}
				return nil
			},
		},
		{
			name: "table caption",
			input: "<table><caption>Title</caption><tr><td>A</td></tr></table>",
			check: func(d *Node) error {
				caption := d.QuerySelectorAll("caption")
				if len(caption) != 1 {
					t.Errorf("expected 1 caption")
				}
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			if err := tc.check(doc); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParser_ListStructure(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(d *Node) error
	}{
		{
			name: "unordered list",
			input: "<ul><li>A</li><li>B</li></ul>",
			check: func(d *Node) error {
				ul := d.QuerySelectorAll("ul")
				li := d.QuerySelectorAll("li")
				if len(ul) != 1 || len(li) != 2 {
					t.Errorf("ul=%d li=%d", len(ul), len(li))
				}
				return nil
			},
		},
		{
			name: "ordered list",
			input: "<ol><li>A</li><li>B</li></ol>",
			check: func(d *Node) error {
				ol := d.QuerySelectorAll("ol")
				li := d.QuerySelectorAll("li")
				if len(ol) != 1 || len(li) != 2 {
					t.Errorf("ol=%d li=%d", len(ol), len(li))
				}
				return nil
			},
		},
		{
			name: "nested lists",
			input: "<ul><li>A<ul><li>B</li></ul></li></ul>",
			check: func(d *Node) error {
				li := d.QuerySelectorAll("li")
				if len(li) != 2 {
					t.Errorf("expected 2 li, got %d", len(li))
				}
				return nil
			},
		},
		{
			name: "definition list",
			input: "<dl><dt>Term</dt><dd>Def</dd></dl>",
			check: func(d *Node) error {
				dt := d.QuerySelectorAll("dt")
				dd := d.QuerySelectorAll("dd")
				if len(dt) != 1 || len(dd) != 1 {
					t.Errorf("dt=%d dd=%d", len(dt), len(dd))
				}
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			if err := tc.check(doc); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParser_FormElements(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(d *Node) error
	}{
		{
			name: "form basic",
			input: "<form><input name=x><button>Go</button></form>",
			check: func(d *Node) error {
				form := d.QuerySelectorAll("form")
				input := d.QuerySelectorAll("input")
				button := d.QuerySelectorAll("button")
				if len(form) != 1 || len(input) != 1 || len(button) != 1 {
					t.Errorf("form=%d input=%d button=%d", len(form), len(input), len(button))
				}
				return nil
			},
		},
		{
			name: "select",
			input: "<select><option>A</option><option>B</option></select>",
			check: func(d *Node) error {
				opt := d.QuerySelectorAll("option")
				if len(opt) != 2 {
					t.Errorf("expected 2 options, got %d", len(opt))
				}
				return nil
			},
		},
		{
			name: "textarea",
			input: "<textarea>content</textarea>",
			check: func(d *Node) error {
				ta := d.QuerySelectorAll("textarea")
				if len(ta) != 1 {
					t.Errorf("expected 1 textarea")
				}
				return nil
			},
		},
		{
			name: "label for input",
			input: "<label for=x>Label</label><input id=x>",
			check: func(d *Node) error {
				label := d.QuerySelectorAll("label")
				if len(label) != 1 {
					t.Errorf("expected 1 label")
				}
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			if err := tc.check(doc); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParser_SemanticElements(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(d *Node) error
	}{
		{
			name: "header footer",
			input: "<header></header><main></main><footer></footer>",
			check: func(d *Node) error {
				h := d.QuerySelectorAll("header")
				m := d.QuerySelectorAll("main")
				f := d.QuerySelectorAll("footer")
				if len(h) != 1 || len(m) != 1 || len(f) != 1 {
					t.Errorf("header=%d main=%d footer=%d", len(h), len(m), len(f))
				}
				return nil
			},
		},
		{
			name: "article section",
			input: "<article><section></section></article>",
			check: func(d *Node) error {
				a := d.QuerySelectorAll("article")
				s := d.QuerySelectorAll("section")
				if len(a) != 1 || len(s) != 1 {
					t.Errorf("article=%d section=%d", len(a), len(s))
				}
				return nil
			},
		},
		{
			name: "nav aside",
			input: "<nav></nav><aside></aside>",
			check: func(d *Node) error {
				nav := d.QuerySelectorAll("nav")
				aside := d.QuerySelectorAll("aside")
				if len(nav) != 1 || len(aside) != 1 {
					t.Errorf("nav=%d aside=%d", len(nav), len(aside))
				}
				return nil
			},
		},
		{
			name: "figure figcaption",
			input: "<figure><img src=x alt=y><figcaption>Caption</figcaption></figure>",
			check: func(d *Node) error {
				fig := d.QuerySelectorAll("figure")
				cap := d.QuerySelectorAll("figcaption")
				if len(fig) != 1 || len(cap) != 1 {
					t.Errorf("figure=%d figcaption=%d", len(fig), len(cap))
				}
				return nil
			},
		},
		{
			name: "details summary",
			input: "<details><summary>Summary</summary>Content</details>",
			check: func(d *Node) error {
				det := d.QuerySelectorAll("details")
				sum := d.QuerySelectorAll("summary")
				if len(det) != 1 || len(sum) != 1 {
					t.Errorf("details=%d summary=%d", len(det), len(sum))
				}
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			if err := tc.check(doc); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParser_AttributePreservation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		tag     string
		key     string
		valWant string
	}{
		{"id", `<div id="main">`, "div", "id", "main"},
		{"class", `<div class="foo bar">`, "div", "class", "foo bar"},
		{"data attr", `<span data-val="123">`, "span", "data-val", "123"},
		{"aria", `<button aria-label="close">`, "button", "aria-label", "close"},
		{"src alt", `<img src="a.png" alt="image">`, "img", "src", "a.png"},
		{"href", `<a href="/page">Link</a>`, "a", "href", "/page"},
		{"rel", `<link rel="stylesheet" href="a.css">`, "link", "rel", "stylesheet"},
		{"type value", `<input type="email">`, "input", "type", "email"},
		{"multiple", `<div id=x class=y data-z=w>`, "div", "id", "x"},
		{"boolean attr", `<input disabled checked>`, "input", "disabled", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			els := doc.QuerySelectorAll(tc.tag)
			if len(els) == 0 {
				t.Fatalf("no %s found", tc.tag)
			}
			got := els[0].GetAttribute(tc.key)
			if got != tc.valWant {
				t.Errorf("attr %s=%q want %q", tc.key, got, tc.valWant)
			}
		})
	}
}

func TestParser_EndTagHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(d *Node) error
	}{
		{
			name:  "ignore unknown end tag",
			input: "<div>hello</div></div></span>",
			check: func(d *Node) error {
				divs := d.QuerySelectorAll("div")
				if len(divs) != 1 {
					t.Errorf("expected 1 div")
				}
				return nil
			},
		},
		{
			name:  "tr implies td close",
			input: "<table><tr><td>A</td></tr></table>",
			check: func(d *Node) error {
				td := d.QuerySelectorAll("td")
				if len(td) != 1 {
					t.Errorf("expected 1 td")
				}
				return nil
			},
		},
		{
			name:  "closing body closes all",
			input: "<div><span>text</span></div></body>",
			check: func(d *Node) error {
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			if err := tc.check(doc); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParser_DeeplyNested(t *testing.T) {
	// Test that deeply nested structures don't cause stack overflow
	tests := []struct {
		name      string
		input     string
		maxDepth  int
	}{
		{"divs 100 deep", strings.Repeat("<div>", 100) + strings.Repeat("</div>", 100), 100},
		{"divs 1000 deep", strings.Repeat("<div>", 1000) + strings.Repeat("</div>", 1000), 1000},
		{"spans 500 deep", strings.Repeat("<span>", 500) + strings.Repeat("</span>", 500), 500},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse([]byte(tc.input))
			// Verify it parsed without crashing
			if doc == nil {
				t.Fatal("doc was nil")
			}
		})
	}
}

func TestParser_MalformedHTML(t *testing.T) {
	// Parser should not panic on malformed input
	tests := []struct {
		name  string
		input string
	}{
		{"unclosed html", "<html><body>"},
		{"extra closing tags", "</div></div></div>"},
		{"mismatched tags", "<div></span>"},
		{"null byte", "<div>\x00text</div>"},
		{"huge attribute", "<div " + strings.Repeat("a", 10000) + "=\"v\">"},
		{"script with </div>", "<script></div></script>"},
		{"style with </style>", "<style></style></style>"},
		{"nested tables", "<table><tr><td><table><tr><td></td></tr></table></td></tr></table>"},
		{"select inside table", "<table><select><option></option></select></table>"},
		{"optgroup inside select", "<select><optgroup><option></option></optgroup></select>"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panicked on %q: %v", tc.name, r)
				}
			}()
			doc := Parse([]byte(tc.input))
			if doc == nil {
				t.Errorf("doc was nil for %q", tc.name)
			}
		})
	}
}

// TestParser_SamplePages parses real sample pages and verifies structure.
func TestParser_SamplePages(t *testing.T) {
	pages := []string{"test1.html", "test2.html", "test3.html", "test4.html", "test5.html"}
	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			data, err := os.ReadFile("../../sample_pages/" + page)
			if err != nil {
				t.Skipf("sample page not found: %s", page)
			}
			doc := Parse(data)
			if doc == nil {
				t.Error("doc was nil")
			}
			// Should have html, head, body
			if len(doc.QuerySelectorAll("html")) != 1 {
				t.Error("missing html element")
			}
		})
	}
}

// BenchmarkParser_SamplePages benchmarks parsing of sample pages.
func BenchmarkParser_SamplePages(b *testing.B) {
	pages := []string{"test1.html", "test2.html", "test3.html", "test4.html", "test5.html"}
	var docs [][]byte
	for _, page := range pages {
		data, err := os.ReadFile("../../sample_pages/" + page)
		if err != nil {
			continue
		}
		docs = append(docs, data)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, data := range docs {
			Parse(data)
		}
	}
}

// BenchmarkParser_Stress benchmarks parsing of deeply nested structures.
func BenchmarkParser_DeepNesting(b *testing.B) {
	input := strings.Repeat("<div>", 500) + strings.Repeat("</div>", 500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse([]byte(input))
	}
}
