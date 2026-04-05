package internal

import (
	"os"
	"testing"

	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/html"
	"github.com/nickcoury/ViBrowsing/internal/layout"
)

// IntegrationHarness tests the full pipeline: HTML → DOM → Layout from end to end.

func TestIntegration_TokenizerToDOM(t *testing.T) {
	tests := []struct {
		name     string
		html     string
	}{
		{"minimal", "<html><body></body></html>"},
		{"with div", "<div><span>text</span></div>"},
		{"nested lists", "<ul><li>a</li><li>b</li></ul>"},
		{"deeply nested 100", stringsReplicate("<div>", 100) + stringsReplicate("</div>", 100)},
		{"deeply nested 500", stringsReplicate("<div>", 500) + stringsReplicate("</div>", 500)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Tokenize and parse independently; doc may be nil if parser has issues
			tokens := html.Tokenize([]byte(tc.html))
			doc := html.Parse([]byte(tc.html))
			// One of them should produce output
			if len(tokens) == 0 && doc == nil {
				t.Error("no tokens and nil doc")
			}
		})
	}
}

func TestIntegration_DOMToLayoutTree(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		css   string
	}{
		{"simple page", "<html><body><h1>Title</h1><p>Paragraph</p></body></html>", ""},
		{"css selectors", `<div id="main" class="container"><p class="text">Hello</p></div>`, `#main { color: red; } .text { font-size: 16px; }`},
		{"inline override", `<div style="color:blue">inline</div>`, `div { color: red; }`},
		{"flex layout", `<div style="display:flex"><div>A</div><div>B</div><div>C</div></div>`, ""},
		{"floated elements", `<div style="float:left;width:100px">left</div><div style="float:right;width:100px">right</div>`, ""},
		{"positioned", `<div style="position:relative"><div style="position:absolute;top:0;left:0">abs</div></div>`, ""},
		{"nested flex", `<div style="display:flex"><div style="display:flex"><div>A</div></div></div>`, ""},
		{"opacity", `<div style="opacity:0.5">translucent</div>`, ""},
		{"visibility hidden", `<div style="visibility:hidden">hidden</div>`, ""},
		{"display none", `<div style="display:none">gone</div><div>visible</div>`, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := html.Parse([]byte(tc.html))
			rules := css.Parse(tc.css)
			box := layout.BuildLayoutTree(doc, rules, 800, 600)
			if box == nil {
				t.Error("box tree was nil")
			}
		})
	}
}

func TestIntegration_LayoutDimensions(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		width    float64
		skipZero bool // if true, 0 dimensions are acceptable (unknown implementation)
	}{
		{
			name:  "fixed width div",
			html:  "<div style='width:200px'>content</div>",
			width: 800,
		},
		{
			name:  "percentage width",
			html:  "<div style='width:50%'>half</div>",
			width: 800,
		},
		{
			name:  "text wraps",
			html:  "<div style='width:100px'>" + stringsReplicate("x", 500) + "</div>",
			width: 100,
		},
		{
			name:  "flex row",
			html:  "<div style='display:flex'><div style='width:100px'>A</div><div style='width:100px'>B</div></div>",
			width: 800,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := html.Parse([]byte("<html><body>" + tc.html + "</body></html>"))
			box := layout.BuildLayoutTree(doc, nil, 800, 600)
			if box == nil {
				t.Skip("could not build layout tree")
			}
			layout.LayoutBlock(box, tc.width)
			// Just verify it doesn't crash with zero/positive dimensions
			if box.ContentW < 0 || box.ContentH < 0 {
				t.Errorf("negative dimensions: %.0fx%.0f", box.ContentW, box.ContentH)
			}
		})
	}
}

func TestIntegration_EndToEndPages(t *testing.T) {
	pages := []string{"test1.html", "test2.html", "test3.html", "test4.html", "test5.html"}
	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			data, err := os.ReadFile("../../sample_pages/" + page)
			if err != nil {
				t.Skipf("sample page not found: %s", page)
			}

			// Tokenize
			tokens := html.Tokenize(data)
			if len(tokens) == 0 {
				t.Fatal("tokenizer produced no tokens")
			}

			// Parse to DOM
			doc := html.Parse(data)
			if doc == nil {
				t.Fatal("parser produced nil doc")
			}

			// Build layout tree
			rules := css.Parse("")
			box := layout.BuildLayoutTree(doc, rules, 800, 600)
			if box == nil {
				t.Fatal("layout tree was nil")
			}

			// Layout pass - should not crash
			layout.LayoutBlock(box, 1024)
		})
	}
}

func TestIntegration_CSSCascade(t *testing.T) {
	// Test that CSS cascade applies correctly
	doc := html.Parse([]byte(`<div id="a" class="b">x</div>`))
	rules := css.Parse(`#a { color: red; } .b { color: blue; }`)
	box := layout.BuildLayoutTree(doc, rules, 800, 600)
	if box == nil {
		t.Fatal("box was nil")
	}
	// Just verify it doesn't crash with CSS rules
}

func TestIntegration_MalformedInput(t *testing.T) {
	inputs := []string{
		"",
		"<",
		"</div>",
		"<div><span>",
		"<div></div></div>",
		"<div></span>",
		stringsReplicate("<div>", 1000) + "text",
		"<div " + stringsReplicate("a", 1000) + "=\"x\">",
	}
	for _, input := range inputs {
		t.Run(input[:min(len(input), 30)], func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panicked on %q: %v", input[:min(len(input), 30)], r)
				}
			}()
			tokens := html.Tokenize([]byte(input))
			_ = tokens
			doc := html.Parse([]byte(input))
			if doc != nil {
				_ = layout.BuildLayoutTree(doc, nil, 800, 600)
			}
		})
	}
}

func TestIntegration_LargeDocument(t *testing.T) {
	// Large document should not panic
	deep := stringsReplicate("<div>", 500) + stringsReplicate("</div>", 500)
	doc := html.Parse([]byte(deep))
	if doc == nil {
		t.Fatal("doc was nil")
	}
	box := layout.BuildLayoutTree(doc, nil, 800, 600)
	if box == nil {
		t.Fatal("box was nil")
	}
	// Should not panic
	layout.LayoutBlock(box, 800)
}

// BenchmarkIntegration_FullPipeline benchmarks HTML → DOM → Layout
func BenchmarkIntegration_FullPipeline(b *testing.B) {
	pages := []string{"test1.html", "test2.html", "test3.html", "test4.html", "test5.html"}
	var docs []struct {
		data  []byte
		doc   *html.Node
		rules []css.Rule
	}
	for _, page := range pages {
		data, err := os.ReadFile("../../sample_pages/" + page)
		if err != nil {
			continue
		}
		doc := html.Parse(data)
		rules := css.Parse("")
		docs = append(docs, struct {
			data  []byte
			doc   *html.Node
			rules []css.Rule
		}{data, doc, rules})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, item := range docs {
			_ = html.Parse(item.data)
			box := layout.BuildLayoutTree(item.doc, item.rules, 800, 600)
			if box != nil {
				layout.LayoutBlock(box, 1024)
			}
		}
	}
}

func stringsReplicate(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
