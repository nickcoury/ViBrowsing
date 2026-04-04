package css

import (
	"fmt"
	"strings"
	"testing"
)

// ParserHarness is a comprehensive test suite for the CSS parser.
// Covers rule parsing, declaration parsing, selector handling,
// and all implemented CSS properties.

func TestParser_Rules(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantN    int
		checkRule func(r Rule) error
	}{
		{
			name:  "single rule",
			input: "div { color: red; }",
			wantN: 1,
			checkRule: func(r Rule) error {
				if r.Selector != "div" {
					t.Errorf("selector=%q want %q", r.Selector, "div")
				}
				return nil
			},
		},
		{
			name:  "multiple rules",
			input: "div { color: red; } p { font-size: 16px; }",
			wantN: 2,
			checkRule: func(r Rule) error {
				return nil
			},
		},
		{
			name:  "comma selectors",
			input: "div, p, span { color: blue; }",
			wantN: 3,
			checkRule: func(r Rule) error {
				return nil
			},
		},
		{
			name:  "class selector",
			input: ".container { width: 100%; }",
			wantN: 1,
			checkRule: func(r Rule) error {
				if r.Selector != ".container" {
					t.Errorf("got %q", r.Selector)
				}
				return nil
			},
		},
		{
			name:  "id selector",
			input: "#main { margin: 0; }",
			wantN: 1,
			checkRule: func(r Rule) error {
				if r.Selector != "#main" {
					t.Errorf("got %q", r.Selector)
				}
				return nil
			},
		},
		{
			name:  "descendant selector",
			input: "div p { line-height: 1.5; }",
			wantN: 1,
			checkRule: func(r Rule) error {
				if r.Selector != "div p" {
					t.Errorf("got %q", r.Selector)
				}
				return nil
			},
		},
		{
			name:  "child selector",
			input: "div > p { color: black; }",
			wantN: 1,
			checkRule: func(r Rule) error {
				return nil
			},
		},
		{
			name:  "attribute selector",
			input: "input[type=text] { border: 1px; }",
			wantN: 1,
			checkRule: func(r Rule) error {
				return nil
			},
		},
		{
			name:  "pseudo class",
			input: "a:hover { text-decoration: underline; }",
			wantN: 1,
			checkRule: func(r Rule) error {
				return nil
			},
		},
		{
			name:  "empty rules skipped",
			input: "div { } span { color: red; }",
			wantN: 1,
			checkRule: func(r Rule) error {
				if r.Selector != "span" {
					t.Errorf("expected span, got %q", r.Selector)
				}
				return nil
			},
		},
		{
			name:  "at rules skipped",
			input: "@import url(foo.css); div { color: red; }",
			wantN: 1,
			checkRule: func(r Rule) error {
				return nil
			},
		},
		{
			name:  "comments stripped",
			input: "/* comment */ div { color: red; } /* another */",
			wantN: 1,
			checkRule: func(r Rule) error {
				return nil
			},
		},
		{
			name:  "multiple declarations",
			input: "div { color: red; font-size: 16px; margin: 10px; }",
			wantN: 1,
			checkRule: func(r Rule) error {
				if len(r.Declarations) != 3 {
					t.Errorf("expected 3 decls, got %d", len(r.Declarations))
				}
				return nil
			},
		},
		{
			name:  "important flag ignored",
			input: "div { color: red !important; }",
			wantN: 1,
			checkRule: func(r Rule) error {
				if len(r.Declarations) != 1 {
					t.Errorf("expected 1 decl, got %d", len(r.Declarations))
				}
				return nil
			},
		},
		{
			name:  "windows line endings",
			input: "div { color: red;\r\n }",
			wantN: 1,
			checkRule: func(r Rule) error {
				return nil
			},
		},
		{
			name:  "mac line endings",
			input: "div { color: red;\r }",
			wantN: 1,
			checkRule: func(r Rule) error {
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rules := Parse(tc.input)
			if len(rules) != tc.wantN {
				t.Errorf("got %d rules, want %d", len(rules), tc.wantN)
				return
			}
			if tc.checkRule != nil {
				if err := tc.checkRule(rules[0]); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestParser_Declarations(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantProp   string
		wantValue  string
	}{
		{"basic", "color: red", "color", "red"},
		{"with spaces", "  color  :  red  ", "color", "red"},
		{"empty value", "color: ", "color", ""},
		{"empty prop", ": red", "", ""},
		{"no colon", "color red", "", ""},
		{"multiple semicolons", "a: 1;;; b: 2", "a", "1"},
		{"important", "color: red ! important", "color", "red"},
		{"url value", "background: url(foo.png)", "background", "url(foo.png)"},
		{"function value", "width: calc(100% - 20px)", "width", "calc(100% - 20px)"},
		{"color function", "color: rgb(255, 0, 0)", "color", "rgb(255, 0, 0)"},
		{"font family", "font-family: Arial, sans-serif", "font-family", "Arial, sans-serif"},
		{"border shorthand", "border: 1px solid red", "border", "1px solid red"},
		{"dash in prop", "font-size: 16px", "font-size", "16px"},
		{"underscore prop", "margin_top: 10px", "margin_top", "10px"},
		{"star prefix prop", "*zoom: 1", "*zoom", "1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			decls := ParseInline(tc.input)
			if tc.wantProp == "" {
				return
			}
			var found *Declaration
			for _, d := range decls {
				if d.Property == tc.wantProp {
					found = &d
					break
				}
			}
			if found == nil {
				t.Errorf("property %q not found in %v", tc.wantProp, decls)
				return
			}
			if found.Value != tc.wantValue {
				t.Errorf("value=%q want %q", found.Value, tc.wantValue)
			}
		})
	}
}

func TestParser_Comments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wantN int
	}{
		{"block comment", "/* hello */ div { color: red; }", 1},
		{"multiple comments", "/* a */ /* b */ div { color: red; }", 1},
		{"comment between rules", "div { color: red; } /* c */ p { }", 1},
		{"comment in decl", "div { /* inside */ color: red; }", 1},
		{"unclosed comment", "div { color: red; } /* unclosed", 1},
		{"nested stars", "/***** test *****/", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rules := Parse(tc.input)
			if len(rules) != tc.wantN {
				t.Errorf("got %d rules, want %d", len(rules), tc.wantN)
			}
		})
	}
}

// TestParser_Properties covers all implemented CSS properties.
func TestParser_Properties(t *testing.T) {
	// Each test is a CSS property → expected stored value
	tests := []struct {
		prop   string
		value  string
		check  func(d Declaration) error
	}{
		// Colors
		{"color", "red", checkDecl("color", "red")},
		{"color", "#fff", checkDecl("color", "#fff")},
		{"color", "#ffffff", checkDecl("color", "#ffffff")},
		{"color", "rgb(255,128,0)", checkDecl("color", "rgb(255,128,0)")},
		{"color", "rgba(255,128,0,0.5)", checkDecl("color", "rgba(255,128,0,0.5)")},
		{"color", "hsl(120,50%,50%)", checkDecl("color", "hsl(120,50%,50%)")},
		{"background-color", "blue", checkDecl("background-color", "blue")},
		{"border-color", "red", checkDecl("border-color", "red")},

		// Layout
		{"display", "block", checkDecl("display", "block")},
		{"display", "flex", checkDecl("display", "flex")},
		{"display", "inline", checkDecl("display", "inline")},
		{"display", "none", checkDecl("display", "none")},
		{"display", "inline-block", checkDecl("display", "inline-block")},

		// Sizing
		{"width", "100px", checkDecl("width", "100px")},
		{"width", "50%", checkDecl("width", "50%")},
		{"width", "auto", checkDecl("width", "auto")},
		{"height", "200px", checkDecl("height", "200px")},
		{"height", "auto", checkDecl("height", "auto")},
		{"max-width", "800px", checkDecl("max-width", "800px")},
		{"min-height", "100px", checkDecl("min-height", "100px")},

		// Margin & Padding
		{"margin", "10px", checkDecl("margin", "10px")},
		{"margin-top", "5px", checkDecl("margin-top", "5px")},
		{"padding", "8px", checkDecl("padding", "8px")},
		{"padding-left", "4px", checkDecl("padding-left", "4px")},

		// Border
		{"border", "1px solid black", checkDecl("border", "1px solid black")},
		{"border-width", "2px", checkDecl("border-width", "2px")},
		{"border-style", "dashed", checkDecl("border-style", "dashed")},
		{"border-color", "red", checkDecl("border-color", "red")},
		{"border-radius", "5px", checkDecl("border-radius", "5px")},
		{"border-radius", "10px 5px", checkDecl("border-radius", "10px 5px")},
		{"border-radius", "1px 2px 3px 4px", checkDecl("border-radius", "1px 2px 3px 4px")},

		// Box
		{"box-shadow", "2px 2px 4px rgba(0,0,0,0.5)", checkDecl("box-shadow", "2px 2px 4px rgba(0,0,0,0.5)")},
		{"outline", "1px solid blue", checkDecl("outline", "1px solid blue")},

		// Typography
		{"font-size", "16px", checkDecl("font-size", "16px")},
		{"font-size", "1.2em", checkDecl("font-size", "1.2em")},
		{"font-size", "large", checkDecl("font-size", "large")},
		{"font-weight", "bold", checkDecl("font-weight", "bold")},
		{"font-weight", "700", checkDecl("font-weight", "700")},
		{"font-style", "italic", checkDecl("font-style", "italic")},
		{"font-family", "Arial", checkDecl("font-family", "Arial")},
		{"line-height", "1.5", checkDecl("line-height", "1.5")},
		{"text-align", "center", checkDecl("text-align", "center")},
		{"text-decoration", "underline", checkDecl("text-decoration", "underline")},
		{"vertical-align", "middle", checkDecl("vertical-align", "middle")},
		{"white-space", "pre", checkDecl("white-space", "pre")},
		{"white-space", "nowrap", checkDecl("white-space", "nowrap")},
		{"letter-spacing", "2px", checkDecl("letter-spacing", "2px")},
		{"word-spacing", "4px", checkDecl("word-spacing", "4px")},
		{"text-indent", "2em", checkDecl("text-indent", "2em")},
		{"text-transform", "uppercase", checkDecl("text-transform", "uppercase")},

		// Flexbox
		{"flex-direction", "row", checkDecl("flex-direction", "row")},
		{"flex-direction", "column-reverse", checkDecl("flex-direction", "column-reverse")},
		{"justify-content", "space-between", checkDecl("justify-content", "space-between")},
		{"align-items", "center", checkDecl("align-items", "center")},
		{"align-self", "flex-end", checkDecl("align-self", "flex-end")},
		{"flex-grow", "1", checkDecl("flex-grow", "1")},
		{"flex-basis", "100px", checkDecl("flex-basis", "100px")},
		{"gap", "10px", checkDecl("gap", "10px")},

		// Positioning
		{"position", "absolute", checkDecl("position", "absolute")},
		{"position", "relative", checkDecl("position", "relative")},
		{"position", "fixed", checkDecl("position", "fixed")},
		{"top", "10px", checkDecl("top", "10px")},
		{"left", "20px", checkDecl("left", "20px")},
		{"right", "0", checkDecl("right", "0")},
		{"bottom", "5px", checkDecl("bottom", "5px")},
		{"z-index", "100", checkDecl("z-index", "100")},

		// Float
		{"float", "left", checkDecl("float", "left")},
		{"float", "right", checkDecl("float", "right")},
		{"clear", "both", checkDecl("clear", "both")},

		// Visibility
		{"visibility", "hidden", checkDecl("visibility", "hidden")},
		{"visibility", "visible", checkDecl("visibility", "visible")},
		{"opacity", "0.5", checkDecl("opacity", "0.5")},
		{"overflow", "hidden", checkDecl("overflow", "hidden")},
		{"overflow", "auto", checkDecl("overflow", "auto")},
		{"overflow-x", "scroll", checkDecl("overflow-x", "scroll")},

		// Background shorthand
		{"background", "red", checkDecl("background", "red")},
		{"background", "#fff url(foo.png) no-repeat center top", checkDecl("background", "#fff url(foo.png) no-repeat center top")},

		// Transform
		{"transform", "rotate(45deg)", checkDecl("transform", "rotate(45deg)")},
		{"transform", "scale(1.5)", checkDecl("transform", "scale(1.5)")},
		{"transform", "translate(10px, 20px)", checkDecl("transform", "translate(10px, 20px)")},
		{"transform", "rotate(15deg) scale(1.2)", checkDecl("transform", "rotate(15deg) scale(1.2)")},

		// Cursor
		{"cursor", "pointer", checkDecl("cursor", "pointer")},
		{"cursor", "text", checkDecl("cursor", "text")},
		{"cursor", "wait", checkDecl("cursor", "wait")},

		// Word wrap
		{"word-wrap", "break-word", checkDecl("word-wrap", "break-word")},
		{"overflow-wrap", "break-word", checkDecl("overflow-wrap", "break-word")},

		// List
		{"list-style-type", "disc", checkDecl("list-style-type", "disc")},
		{"list-style-type", "square", checkDecl("list-style-type", "square")},
		{"list-style-type", "decimal", checkDecl("list-style-type", "decimal")},
		{"list-style-position", "inside", checkDecl("list-style-position", "inside")},
	}
	for _, tc := range tests {
		t.Run(tc.prop, func(t *testing.T) {
			css := tc.prop + ": " + tc.value + ";"
			decls := ParseInline(css)
			if len(decls) == 0 {
				t.Fatalf("no decls parsed from %q", css)
			}
			if err := tc.check(decls[0]); err != nil {
				t.Error(err)
			}
		})
	}
}

func checkDecl(prop, value string) func(d Declaration) error {
	return func(d Declaration) error {
		if d.Property != prop {
			return fmt.Errorf("property=%q want %q", d.Property, prop)
		}
		if d.Value != value {
			return fmt.Errorf("value=%q want %q", d.Value, value)
		}
		return nil
	}
}

// TestParser_ShorthandProperties tests that shorthand properties
// don't interfere with longhand properties.
func TestParser_ShorthandOverride(t *testing.T) {
	css := `
		margin: 10px;
		margin-top: 20px;
		margin-right: 5px;
		padding: 5px;
		padding-bottom: 15px;
	`
	decls := ParseInline(css)
	m := make(map[string]string)
	for _, d := range decls {
		m[d.Property] = d.Value
	}

	if m["margin"] != "10px" {
		t.Errorf("margin=%q", m["margin"])
	}
	if m["margin-top"] != "20px" {
		t.Errorf("margin-top=%q", m["margin-top"])
	}
	if m["padding"] != "5px" {
		t.Errorf("padding=%q", m["padding"])
	}
	if m["padding-bottom"] != "15px" {
		t.Errorf("padding-bottom=%q", m["padding-bottom"])
	}
}

// TestParser_InvalidCSS ensures the parser doesn't crash on bad CSS.
func TestParser_InvalidCSS(t *testing.T) {
	inputs := []string{
		"",
		"div",
		"{ color: red }",
		"div { }",
		"div { color: }",
		"{ }",
		"///***//",
		strings.Repeat("a", 10000) + ": b;",
	}
	for _, input := range inputs {
		t.Run(input[:min(len(input), 30)], func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panicked on %q: %v", input, r)
				}
			}()
			Parse(input)
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// BenchmarkParser_CSSStylesheet benchmarks parsing of a large stylesheet.
func BenchmarkParser_CSSStylesheet(b *testing.B) {
	css := `
		html, body { margin: 0; padding: 0; font-family: Arial, sans-serif; }
		.container { width: 100%; max-width: 1200px; margin: 0 auto; padding: 20px; }
		.header { background: #fff; border-bottom: 1px solid #ddd; padding: 15px 0; }
		.nav { display: flex; gap: 20px; align-items: center; }
		.nav a { color: #333; text-decoration: none; font-size: 14px; }
		.nav a:hover { color: #007bff; text-decoration: underline; }
		.btn { display: inline-block; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
		.btn-primary { background: #007bff; color: white; border: none; }
		.btn-secondary { background: #6c757d; color: white; }
		.card { background: white; border: 1px solid #ddd; border-radius: 8px; padding: 20px; margin: 10px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.card h2 { font-size: 1.5em; margin-bottom: 10px; color: #222; }
		.card p { line-height: 1.6; color: #444; }
		.grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
		@media (max-width: 768px) { .grid { grid-template-columns: 1fr; } }
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse(css)
	}
}

// BenchmarkParser_AtomicDecls benchmarks parsing of many individual declarations.
func BenchmarkParser_AtomicDecls(b *testing.B) {
	decls := "color: red; font-size: 16px; margin: 10px; padding: 5px; background: white; border: 1px solid #ddd;"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse(decls)
	}
}

// TestParser_Keyframes tests @keyframes parsing.
func TestParser_Keyframes(t *testing.T) {
	// Reset global Keyframes before test
	Keyframes = nil

	tests := []struct {
		name         string
		input        string
		wantName     string
		wantPcts     []float64
		wantProps    map[float64]map[string]string
	}{
		{
			name:     "simple from-to",
			input:    "@keyframes fade { from { opacity: 0; } to { opacity: 1; } }",
			wantName: "fade",
			wantPcts: []float64{0, 100},
			wantProps: map[float64]map[string]string{
				0:   {"opacity": "0"},
				100: {"opacity": "1"},
			},
		},
		{
			name:     "percentage keyframes",
			input:    "@keyframes slide { 0% { left: 0; } 50% { left: 100px; } 100% { left: 0; } }",
			wantName: "slide",
			wantPcts: []float64{0, 50, 100},
			wantProps: map[float64]map[string]string{
				0:   {"left": "0"},
				50:  {"left": "100px"},
				100: {"left": "0"},
			},
		},
		{
			name:     "multiple selectors same block",
			input:    "@keyframes test { 0%, 100% { color: red; } 50% { color: blue; } }",
			wantName: "test",
			wantPcts: []float64{0, 100, 50},
			wantProps: map[float64]map[string]string{
				0:   {"color": "red"},
				100: {"color": "red"},
				50:  {"color": "blue"},
			},
		},
		{
			name:     "with webkit prefix",
			input:    "@-webkit-keyframes pulse { from { opacity: 1; } to { opacity: 0; } }",
			wantName: "pulse",
			wantPcts: []float64{0, 100},
			wantProps: map[float64]map[string]string{
				0:   {"opacity": "1"},
				100: {"opacity": "0"},
			},
		},
		{
			name:     "multiple properties per keyframe",
			input:    "@keyframes move { 0% { left: 0; top: 0; } 100% { left: 100px; top: 100px; } }",
			wantName: "move",
			wantPcts: []float64{0, 100},
			wantProps: map[float64]map[string]string{
				0:   {"left": "0", "top": "0"},
				100: {"left": "100px", "top": "100px"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Keyframes = nil // Reset
			Parse(tc.input)
			if len(Keyframes) != 1 {
				t.Fatalf("got %d keyframes, want 1", len(Keyframes))
			}
			kf := Keyframes[0]
			if kf.Name != tc.wantName {
				t.Errorf("name=%q want %q", kf.Name, tc.wantName)
			}
			for _, pct := range tc.wantPcts {
				if _, exists := kf.Keyframes[pct]; !exists {
					t.Errorf("missing keyframe at %f%%", pct)
				} else {
					for prop, wantVal := range tc.wantProps[pct] {
						if gotVal := kf.Keyframes[pct][prop]; gotVal != wantVal {
							t.Errorf("at %f%%, %s=%q want %q", pct, prop, gotVal, wantVal)
						}
					}
				}
			}
		})
	}
}

// TestParser_KeyframesGlobal tests that Keyframes accumulates across multiple Parse calls.
func TestParser_KeyframesGlobal(t *testing.T) {
	Keyframes = nil
	Parse("@keyframes a { from { opacity: 0; } to { opacity: 1; } }")
	Parse("@keyframes b { from { width: 0; } to { width: 100px; } }")
	if len(Keyframes) != 2 {
		t.Errorf("got %d keyframes, want 2", len(Keyframes))
	}
	if Keyframes[0].Name != "a" || Keyframes[1].Name != "b" {
		t.Errorf("unexpected keyframe names: %v", Keyframes)
	}
}

// TestParseKeyframeSelectors tests the selector parsing logic.
func TestParseKeyframeSelectors(t *testing.T) {
	tests := []struct {
		input    string
		expected []float64
	}{
		{"from", []float64{0}},
		{"to", []float64{100}},
		{"0%", []float64{0}},
		{"100%", []float64{100}},
		{"50%", []float64{50}},
		{"0%, 100%", []float64{0, 100}},
		{"25%, 50%, 75%", []float64{25, 50, 75}},
		{"FROM", []float64{0}},
		{"TO", []float64{100}},
		{"0%, 50%, 100%", []float64{0, 50, 100}},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parseKeyframeSelectors(tc.input)
			if len(result) != len(tc.expected) {
				t.Errorf("got %v len=%d want %v", result, len(result), tc.expected)
				return
			}
			for i, v := range result {
				if v != tc.expected[i] {
					t.Errorf("got %v want %v", result, tc.expected)
					return
				}
			}
		})
	}
}

// TestStyle_AnimationProperties tests that animation properties are in defaults and applied correctly.
func TestStyle_AnimationProperties(t *testing.T) {
	props := ComputeStyle("div", "", "", nil, nil)

	// Check defaults
	if props["animation-name"] != "none" {
		t.Errorf("animation-name default=%q want %q", props["animation-name"], "none")
	}
	if props["animation-duration"] != "0s" {
		t.Errorf("animation-duration default=%q want %q", props["animation-duration"], "0s")
	}
	if props["animation-timing-function"] != "ease" {
		t.Errorf("animation-timing-function default=%q want %q", props["animation-timing-function"], "ease")
	}
	if props["animation-delay"] != "0s" {
		t.Errorf("animation-delay default=%q want %q", props["animation-delay"], "0s")
	}
	if props["animation-iteration-count"] != "1" {
		t.Errorf("animation-iteration-count default=%q want %q", props["animation-iteration-count"], "1")
	}
	if props["animation-direction"] != "normal" {
		t.Errorf("animation-direction default=%q want %q", props["animation-direction"], "normal")
	}
	if props["animation-fill-mode"] != "none" {
		t.Errorf("animation-fill-mode default=%q want %q", props["animation-fill-mode"], "none")
	}
}

// TestStyle_AnimationShorthand tests the animation shorthand parsing.
func TestStyle_AnimationShorthand(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected map[string]string
	}{
		{
			name:  "name only",
			value: "fade",
			expected: map[string]string{
				"animation-name":        "fade",
				"animation-duration":    "0s",
				"animation-timing-function": "ease",
				"animation-delay":       "0s",
				"animation-iteration-count": "1",
				"animation-direction":   "normal",
				"animation-fill-mode":   "none",
			},
		},
		{
			name:  "name and duration",
			value: "slide 2s",
			expected: map[string]string{
				"animation-name":        "slide",
				"animation-duration":    "2s",
				"animation-timing-function": "ease",
				"animation-delay":       "0s",
				"animation-iteration-count": "1",
				"animation-direction":   "normal",
				"animation-fill-mode":   "none",
			},
		},
		{
			name:  "full animation",
			value: "fadeIn 1s ease-in 0.5s infinite alternate forwards",
			expected: map[string]string{
				"animation-name":        "fadeIn",
				"animation-duration":    "1s",
				"animation-timing-function": "ease-in",
				"animation-delay":       "0.5s",
				"animation-iteration-count": "infinite",
				"animation-direction":   "alternate",
				"animation-fill-mode":   "forwards",
			},
		},
		{
			name:  "with numeric iteration count",
			value: "spin 3s linear 1s 3 normal",
			expected: map[string]string{
				"animation-name":        "spin",
				"animation-duration":    "3s",
				"animation-timing-function": "linear",
				"animation-delay":       "1s",
				"animation-iteration-count": "3",
				"animation-direction":   "normal",
				"animation-fill-mode":   "none",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			props := map[string]string{
				"animation-name":        "none",
				"animation-duration":    "0s",
				"animation-timing-function": "ease",
				"animation-delay":       "0s",
				"animation-iteration-count": "1",
				"animation-direction":   "normal",
				"animation-fill-mode":   "none",
			}
			parseAnimationShorthand(props, tc.value)
			for k, v := range tc.expected {
				if props[k] != v {
					t.Errorf("animation shorthand %q: %s=%q want %q", tc.value, k, props[k], v)
				}
			}
		})
	}
}

// TestStyle_AnimationIndividualProps tests that individual animation properties override defaults.
func TestStyle_AnimationIndividualProps(t *testing.T) {
	css := `div { animation-name: pulse; animation-duration: 2s; animation-iteration-count: infinite; }`
	rules := Parse(css)
	props := ComputeStyle("div", "", "", nil, rules)
	if props["animation-name"] != "pulse" {
		t.Errorf("animation-name=%q want %q", props["animation-name"], "pulse")
	}
	if props["animation-duration"] != "2s" {
		t.Errorf("animation-duration=%q want %q", props["animation-duration"], "2s")
	}
	if props["animation-iteration-count"] != "infinite" {
		t.Errorf("animation-iteration-count=%q want %q", props["animation-iteration-count"], "infinite")
	}
}

// TestStyle_AnimationWithRules tests animation applied through rules.
func TestStyle_AnimationWithRules(t *testing.T) {
	Keyframes = nil
	css := `
@keyframes myanim {
	0% { opacity: 0; }
	100% { opacity: 1; }
}
div { animation: fade 1s ease; }
`
	rules := Parse(css)
	props := ComputeStyle("div", "", "", nil, rules)
	if props["animation-name"] != "fade" {
		t.Errorf("animation-name=%q want %q", props["animation-name"], "fade")
	}
	if props["animation-duration"] != "1s" {
		t.Errorf("animation-duration=%q want %q", props["animation-duration"], "1s")
	}
	// Verify keyframe was parsed
	if len(Keyframes) != 1 || Keyframes[0].Name != "myanim" {
		t.Errorf("keyframes not parsed correctly: %v", Keyframes)
	}
}
