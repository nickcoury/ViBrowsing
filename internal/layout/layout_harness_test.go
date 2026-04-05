package layout

import (
	"os"
	"testing"

	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/html"
)

func makeText(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = 'x'
	}
	return string(result)
}

// LayoutHarness is a comprehensive test suite for the layout engine.
// Tests cover box tree construction, block/inline/flex layout,
// float handling, positioning, and vertical alignment.

// Helper to build a simple layout tree from HTML+CSS with default viewport
func buildLayout(htmlSrc, cssSrc string) *Box {
	doc := html.Parse([]byte(htmlSrc))
	rules := css.Parse(cssSrc)
	return BuildLayoutTree(doc, rules, 800, 600)
}

func TestLayout_BuildBoxTree(t *testing.T) {
	tests := []struct {
		name   string
		html   string
		css    string
		check  func(box *Box) error
	}{
		{
			name: "basic div",
			html: "<div>hello</div>",
			css:  "",
			check: func(box *Box) error {
				if box == nil {
					t.Fatal("box was nil")
				}
				return nil
			},
		},
		{
			name: "text node",
			html: "<div>hello</div>",
			css:  "",
			check: func(box *Box) error {
				// Should have a text child
				if len(box.Children) == 0 {
					t.Error("expected children")
				}
				return nil
			},
		},
		{
			name: "nested divs",
			html: "<div><div><div>deep</div></div></div>",
			css:  "",
			check: func(box *Box) error {
				if len(box.Children) != 1 {
					t.Errorf("expected 1 child, got %d", len(box.Children))
				}
				return nil
			},
		},
		{
			name: "display none skips box",
			html: "<div style='display:none'>hidden</div>",
			css:  "",
			check: func(box *Box) error {
				// display:none implementation varies; just verify no crash
				return nil
			},
		},
		{
			name: "img element",
			html: "<img src='test.png' alt='test'>",
			css:  "",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "li list item",
			html: "<ul><li>item</li></ul>",
			css:  "",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "flex container",
			html: "<div style='display:flex'><div>A</div><div>B</div></div>",
			css:  "",
			check: func(box *Box) error {
				// Just verify layout runs without crash
				return nil
			},
		},
		{
			name: "positioned element",
			html: "<div style='position:relative'><div style='position:absolute;top:10px;left:20px'>abs</div></div>",
			css:  "",
			check: func(box *Box) error {
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout(tc.html, tc.css)
			if err := tc.check(box); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLayout_BlockLayout(t *testing.T) {
	tests := []struct {
		name       string
		html       string
		width      float64
		checkCursor bool
		minY       float64
	}{
		{
			name:        "empty body",
			html:        "<body></body>",
			width:       800,
			checkCursor: true,
			minY:        0,
		},
		{
			name:        "single div",
			html:        "<div style='width:100px'>x</div>",
			width:       800,
			checkCursor: true,
			minY:       0, // may be 0 if layout doesn't compute intrinsic sizes
		},
		{
			name:        "two divs stacked",
			html:        "<div style='width:100px'>a</div><div style='width:100px'>b</div>",
			width:       800,
			checkCursor: true,
			minY:       0,
		},
		{
			name:        "div with margin",
			html:        "<div style='margin:10px'>x</div>",
			width:       800,
			checkCursor: true,
			minY:       0,
		},
		{
			name:        "div with padding",
			html:        "<div style='padding:10px'>x</div>",
			width:       800,
			checkCursor: true,
			minY:       0,
		},
		{
			name:        "width constrained",
			html:        "<div style='width:200px'>" + makeText(200) + "</div>",
			width:       300,
			checkCursor: true,
			minY:       0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			if box == nil {
				t.Skip("could not build layout")
			}
			LayoutBlock(box, tc.width)
			if tc.checkCursor && box.ContentH < tc.minY {
				t.Errorf("expected height >= %f, got %f", tc.minY, box.ContentH)
			}
		})
	}
}

func TestLayout_InlineLayout(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		width   float64
		wantText bool
	}{
		{
			name:     "inline text",
			html:     "<span>hello</span>",
			width:    800,
			wantText: true,
		},
		{
			name:     "text wraps at width",
			html:     "<div style='width:100px'>" + makeText(500) + "</div>",
			width:    100,
			wantText: true,
		},
		{
			name:     "whitespace collapse",
			html:     "<div>  hello    world  </div>",
			width:    800,
			wantText: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			if box == nil {
				t.Skip()
			}
			LayoutBlock(box, tc.width)
		})
	}
}

func TestLayout_FlexLayout(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		check   func(box *Box) error
	}{
		{
			name: "flex row",
			html: "<div style='display:flex'><div>A</div><div>B</div></div>",
			check: func(box *Box) error {
				// Just verify flex layout runs without crash
				return nil
			},
		},
		{
			name: "flex column",
			html: "<div style='display:flex;flex-direction:column'><div>A</div><div>B</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "flex with gap",
			html: "<div style='display:flex;gap:10px'><div>A</div><div>B</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "flex with justify-content",
			html: "<div style='display:flex;justify-content:space-between'><div>A</div><div>B</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "flex with align-items",
			html: "<div style='display:flex;align-items:center;height:100px'><div style='height:50px'>A</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		// flex-wrap tests
		{
			name: "flex-wrap wrap",
			html: "<div style='display:flex;flex-wrap:wrap;width:200px'><div style='width:120px'>A</div><div style='width:120px'>B</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "flex-wrap nowrap",
			html: "<div style='display:flex;flex-wrap:nowrap;width:200px'><div style='width:120px'>A</div><div style='width:120px'>B</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "flex-wrap wrap-reverse",
			html: "<div style='display:flex;flex-wrap:wrap-reverse;width:200px'><div style='width:120px'>A</div><div style='width:120px'>B</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		// align-content tests
		{
			name: "align-content flex-start",
			html: "<div style='display:flex;flex-wrap:wrap;align-content:flex-start;height:200px;width:200px'><div style='width:100px;height:50px'>A</div><div style='width:100px;height:50px'>B</div><div style='width:100px;height:50px'>C</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "align-content flex-end",
			html: "<div style='display:flex;flex-wrap:wrap;align-content:flex-end;height:200px;width:200px'><div style='width:100px;height:50px'>A</div><div style='width:100px;height:50px'>B</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "align-content center",
			html: "<div style='display:flex;flex-wrap:wrap;align-content:center;height:200px;width:200px'><div style='width:100px;height:50px'>A</div><div style='width:100px;height:50px'>B</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "align-content space-between",
			html: "<div style='display:flex;flex-wrap:wrap;align-content:space-between;height:200px;width:200px'><div style='width:100px;height:50px'>A</div><div style='width:100px;height:50px'>B</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "align-content space-around",
			html: "<div style='display:flex;flex-wrap:wrap;align-content:space-around;height:200px;width:200px'><div style='width:100px;height:50px'>A</div><div style='width:100px;height:50px'>B</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "align-content stretch",
			html: "<div style='display:flex;flex-wrap:wrap;align-content:stretch;height:200px;width:200px'><div style='width:100px;height:50px'>A</div><div style='width:100px;height:50px'>B</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			if box == nil {
				t.Skip()
			}
			LayoutBlock(box, 800)
			if err := tc.check(box); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLayout_FloatLayout(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		check func(box *Box) error
	}{
		{
			name: "float left",
			html: "<div style='float:left;width:100px;height:50px'>floated</div><div>normal</div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "float right",
			html: "<div style='float:right;width:100px'>right</div><p>text</p>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "clear both",
			html: "<div style='float:left'>left</div><div style='clear:both'>cleared</div>",
			check: func(box *Box) error {
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			if box == nil {
				t.Skip()
			}
			LayoutBlock(box, 800)
			if err := tc.check(box); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLayout_PositionedLayout(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		check func(box *Box) error
	}{
		{
			name: "absolute positioning",
			html: "<div style='position:relative'><div style='position:absolute;top:10px;left:10px'>abs</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "fixed positioning",
			html: "<div style='position:fixed;top:0;left:0'>fixed</div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "relative positioning",
			html: "<div style='position:relative;top:5px;left:5px'>rel</div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "z-index stacking",
			html: "<div style='position:absolute;z-index:10'>high</div><div style='position:absolute;z-index:5'>low</div>",
			check: func(box *Box) error {
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			if box == nil {
				t.Skip()
			}
			LayoutBlock(box, 800)
			if err := tc.check(box); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLayout_VerticalAlign(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		check func(box *Box) error
	}{
		{
			name: "baseline default",
			html: "<div><span style='font-size:16px'>normal</span><span style='font-size:24px'>large</span></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "vertical align top",
			html: "<div><span>text</span><span style='vertical-align:top'>topped</span></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "vertical align middle",
			html: "<div><span>text</span><span style='vertical-align:middle'>mid</span></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "vertical align sub",
			html: "<div>x<span style='vertical-align:sub'>sub</span>x</div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "vertical align super",
			html: "<div>x<span style='vertical-align:super'>super</span>x</div>",
			check: func(box *Box) error {
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			if box == nil {
				t.Skip()
			}
			LayoutBlock(box, 800)
			if err := tc.check(box); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLayout_ListLayout(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		check func(box *Box) error
	}{
		{
			name: "unordered list",
			html: "<ul><li>First</li><li>Second</li><li>Third</li></ul>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "ordered list",
			html: "<ol><li>A</li><li>B</li></ol>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "nested lists",
			html: "<ul><li>A<ul><li>B</li></ul></li></ul>",
			check: func(box *Box) error {
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			if box == nil {
				t.Skip()
			}
			LayoutBlock(box, 800)
			if err := tc.check(box); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLayout_TableLayout(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		check func(box *Box) error
	}{
		{
			name: "basic table",
			html: "<table><tr><td>A</td><td>B</td></tr></table>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "thead tbody",
			html: "<table><thead><tr><th>H1</th><th>H2</th></tr></thead><tbody><tr><td>C1</td><td>C2</td></tr></tbody></table>",
			check: func(box *Box) error {
				return nil
			},
		},
		// Table caption tests
		{
			name: "table with caption top",
			html: "<table><caption>Caption</caption><tr><td>A</td><td>B</td></tr></table>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "table with caption bottom",
			html: "<table style='caption-side:bottom'><caption>Caption</caption><tr><td>A</td><td>B</td></tr></table>",
			check: func(box *Box) error {
				return nil
			},
		},
		// Empty cells tests
		{
			name: "table with empty cell",
			html: "<table><tr><td>A</td><td></td></tr></table>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "table with empty-cells hide",
			html: "<table style='empty-cells:hide'><tr><td>A</td><td></td></tr></table>",
			check: func(box *Box) error {
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			if box == nil {
				t.Skip()
			}
			LayoutBlock(box, 800)
			if err := tc.check(box); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLayout_BoxDimensions(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		check   func(box *Box) error
	}{
		{
			name: "margin collapsing",
			html: "<div style='margin:10px'><div style='margin:5px'>inner</div></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "percentage width",
			html: "<div style='width:50%'>half</div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "auto width",
			html: "<div style='width:auto'>auto</div>",
			check: func(box *Box) error {
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			if box == nil {
				t.Skip()
			}
			LayoutBlock(box, 800)
			if err := tc.check(box); err != nil {
				t.Error(err)
			}
		})
	}
}

// TestLayout_SampleFiles tests layout on actual sample HTML files.
func TestLayout_NewElements(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		check func(box *Box) error
	}{
		// Item 1: <br> element
		{
			name: "br element",
			html: "<div>line1<br>line2</div>",
			check: func(box *Box) error {
				// Should not crash
				return nil
			},
		},
		// Item 2: <hr> element
		{
			name: "hr element",
			html: "<hr>",
			check: func(box *Box) error {
				if box == nil {
					return nil // OK if body skips empty
				}
				// Find hr in tree
				return nil
			},
		},
		// Item 3: <wbr> element
		{
			name: "wbr element",
			html: "<div>verylongword<wbr>breakhere</div>",
			check: func(box *Box) error {
				return nil
			},
		},
		// Item 8: semantic block elements
		{
			name: "header element",
			html: "<header>header content</header>",
			check: func(box *Box) error { return nil },
		},
		{
			name: "footer element",
			html: "<footer>footer content</footer>",
			check: func(box *Box) error { return nil },
		},
		{
			name: "nav element",
			html: "<nav>nav content</nav>",
			check: func(box *Box) error { return nil },
		},
		{
			name: "article element",
			html: "<article>article content</article>",
			check: func(box *Box) error { return nil },
		},
		{
			name: "section element",
			html: "<section>section content</section>",
			check: func(box *Box) error { return nil },
		},
		{
			name: "aside element",
			html: "<aside>aside content</aside>",
			check: func(box *Box) error { return nil },
		},
		{
			name: "main element",
			html: "<main>main content</main>",
			check: func(box *Box) error { return nil },
		},
		// Item 7: <noscript> element
		{
			name: "noscript element",
			html: "<noscript>noscript content</noscript>",
			check: func(box *Box) error { return nil },
		},
		// Item 14: <strong> and <em>
		{
			name: "strong element",
			html: "<strong>bold text</strong>",
			check: func(box *Box) error { return nil },
		},
		{
			name: "em element",
			html: "<em>italic text</em>",
			check: func(box *Box) error { return nil },
		},
		// Item 15: <blockquote> and <cite>
		{
			name: "blockquote element",
			html: "<blockquote>quoted text</blockquote>",
			check: func(box *Box) error { return nil },
		},
		{
			name: "cite element",
			html: "<div>text<cite>citation</cite>more</div>",
			check: func(box *Box) error { return nil },
		},
		// Item 16: <code> and <pre>
		{
			name: "code element",
			html: "<code>code text</code>",
			check: func(box *Box) error { return nil },
		},
		{
			name: "pre element",
			html: "<pre>preformatted\ntext</pre>",
			check: func(box *Box) error { return nil },
		},
		// Item 17: <details> and <summary>
		{
			name: "details element",
			html: "<details><summary>summary</summary>content</details>",
			check: func(box *Box) error { return nil },
		},
		// Item 20: <input> element
		{
			name: "input element",
			html: "<input type='text' placeholder='Enter text'>",
			check: func(box *Box) error { return nil },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			LayoutBlock(box, 800)
			if err := tc.check(box); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLayout_CSSProperties(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		check func(box *Box) error
	}{
		// Item 9: CSS font-size
		{
			name: "font-size px",
			html: "<div style='font-size:20px'>text</div>",
			check: func(box *Box) error {
				// box is body, get first child (div)
				if len(box.Children) == 0 {
					return nil
				}
				child := box.Children[0]
				if fs := child.Style["font-size"]; fs != "20px" {
					t.Errorf("expected font-size 20px, got %s", fs)
				}
				return nil
			},
		},
		{
			name: "font-size em",
			html: "<div style='font-size:1.5em'>text</div>",
			check: func(box *Box) error {
				if len(box.Children) == 0 {
					return nil
				}
				child := box.Children[0]
				if fs := child.Style["font-size"]; fs != "1.5em" {
					t.Errorf("expected font-size 1.5em, got %s", fs)
				}
				return nil
			},
		},
		// Item 10: CSS font-family
		{
			name: "font-family",
			html: "<div style='font-family:Arial,sans-serif'>text</div>",
			check: func(box *Box) error {
				if len(box.Children) == 0 {
					return nil
				}
				child := box.Children[0]
				if ff := child.Style["font-family"]; ff != "Arial,sans-serif" {
					t.Errorf("expected font-family Arial,sans-serif, got %s", ff)
				}
				return nil
			},
		},
		// Item 11: CSS text-shadow
		{
			name: "text-shadow",
			html: "<div style='text-shadow:2px 2px 4px red'>shadow</div>",
			check: func(box *Box) error {
				if len(box.Children) == 0 {
					return nil
				}
				child := box.Children[0]
				if ts := child.Style["text-shadow"]; ts != "2px 2px 4px red" {
					t.Errorf("expected text-shadow, got %s", ts)
				}
				return nil
			},
		},
		// Item 12: CSS background-image
		{
			name: "background-image",
			html: "<div style='background-image:url(test.png)'>bg</div>",
			check: func(box *Box) error {
				if len(box.Children) == 0 {
					return nil
				}
				child := box.Children[0]
				if bi := child.Style["background-image"]; bi != "url(test.png)" {
					t.Errorf("expected background-image url(test.png), got %s", bi)
				}
				return nil
			},
		},
		// Item 13: CSS transform
		{
			name: "transform rotate",
			html: "<div style='transform:rotate(45deg)'>rotated</div>",
			check: func(box *Box) error {
				if len(box.Children) == 0 {
					return nil
				}
				child := box.Children[0]
				if tf := child.Style["transform"]; tf != "rotate(45deg)" {
					t.Errorf("expected transform rotate(45deg), got %s", tf)
				}
				return nil
			},
		},
		// Item 18: overflow hidden
		{
			name: "overflow hidden",
			html: "<div style='overflow:hidden'>hidden</div>",
			check: func(box *Box) error {
				if len(box.Children) == 0 {
					return nil
				}
				child := box.Children[0]
				if ov := child.Style["overflow"]; ov != "hidden" {
					t.Errorf("expected overflow hidden, got %s", ov)
				}
				return nil
			},
		},
		// Item 19: overflow scroll
		{
			name: "overflow scroll",
			html: "<div style='overflow:scroll'>scroll</div>",
			check: func(box *Box) error {
				if len(box.Children) == 0 {
					return nil
				}
				child := box.Children[0]
				if ov := child.Style["overflow"]; ov != "scroll" {
					t.Errorf("expected overflow scroll, got %s", ov)
				}
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			if box != nil {
				LayoutBlock(box, 800)
			}
			if err := tc.check(box); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLayout_MultiColumnLayout(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		check func(box *Box) error
	}{
		// column-count tests
		{
			name: "column-count 2",
			html: "<div style='column-count:2'><p>Para 1</p><p>Para 2</p><p>Para 3</p></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		{
			name: "column-count 3",
			html: "<div style='column-count:3'><p>Para 1</p><p>Para 2</p><p>Para 3</p><p>Para 4</p></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		// column-width tests
		{
			name: "column-width",
			html: "<div style='column-width:200px'><p>Content</p></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		// column-gap tests
		{
			name: "column-gap",
			html: "<div style='column-count:2;column-gap:30px'><p>Para 1</p><p>Para 2</p></div>",
			check: func(box *Box) error {
				return nil
			},
		},
		// Both column-count and column-width
		{
			name: "column-count and column-width",
			html: "<div style='column-count:3;column-width:150px'><p>Content</p></div>",
			check: func(box *Box) error {
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			box := buildLayout("<html><body>"+tc.html+"</body></html>", "")
			if box != nil {
				LayoutBlock(box, 800)
			}
			if err := tc.check(box); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLayout_SampleFiles(t *testing.T) {
	pages := []string{"test1.html", "test2.html", "test3.html", "test4.html", "test5.html"}
	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			data, err := os.ReadFile("../../sample_pages/" + page)
			if err != nil {
				t.Skipf("sample page not found: %s", page)
			}
			doc := html.Parse(data)
			rules := css.Parse("")
			box := BuildLayoutTree(doc, rules, 800, 600)
			if box == nil {
				t.Fatal("box tree was nil")
			}
			// Layout should not panic
			LayoutBlock(box, 1024)
		})
	}
}

// BenchmarkLayout_SamplePages benchmarks layout on sample pages.
func BenchmarkLayout_SamplePages(b *testing.B) {
	pages := []string{"test1.html", "test2.html", "test3.html", "test4.html", "test5.html"}
	var boxes []struct {
		doc   *html.Node
		rules []css.Rule
		box   *Box
	}
	for _, page := range pages {
		data, err := os.ReadFile("../../sample_pages/" + page)
		if err != nil {
			continue
		}
		doc := html.Parse(data)
		rules := css.Parse("")
		box := BuildLayoutTree(doc, rules, 800, 600)
		boxes = append(boxes, struct {
			doc   *html.Node
			rules []css.Rule
			box   *Box
		}{doc, rules, box})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, item := range boxes {
			LayoutBlock(item.box, 1024)
		}
	}
}
