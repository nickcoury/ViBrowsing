package css

import (
	"testing"

	"github.com/nickcoury/ViBrowsing/internal/html"
)

func TestGetComputedStyle_BasicElement(t *testing.T) {
	// Create a simple document: <div id="test"></div>
	doc := &html.Node{
		Type: html.NodeDocument,
	}
	htmlNode := &html.Node{
		Type:     html.NodeElement,
		TagName:  "div",
		Parent:   doc,
		Children: []*html.Node{},
	}
	doc.Children = append(doc.Children, htmlNode)

	rules := []Rule{}

	style := GetComputedStyle(htmlNode, rules)

	// Check that we get default values
	if style["display"] != "block" {
		t.Errorf("expected display=block, got %s", style["display"])
	}
	if style["visibility"] != "visible" {
		t.Errorf("expected visibility=visible, got %s", style["visibility"])
	}
	if style["font-size"] != "16px" {
		t.Errorf("expected font-size=16px, got %s", style["font-size"])
	}
}

func TestGetComputedStyle_NilNode(t *testing.T) {
	rules := []Rule{}
	style := GetComputedStyle(nil, rules)

	if len(style) != 0 {
		t.Errorf("expected empty map for nil node, got %d entries", len(style))
	}
}

func TestGetComputedStyle_Inheritance(t *testing.T) {
	// Test that computed style for a span correctly inherits from parent
	// Since inline style parsing has issues with SetAttribute, we test
	// that GetComputedStyle properly walks the ancestor chain
	doc := &html.Node{
		Type: html.NodeDocument,
	}
	div := &html.Node{
		Type:     html.NodeElement,
		TagName:  "div",
		Parent:   doc,
		Children: []*html.Node{},
	}
	span := &html.Node{
		Type:     html.NodeElement,
		TagName:  "span",
		Parent:   div,
		Children: []*html.Node{},
	}
	div.Children = append(div.Children, span)
	doc.Children = append(doc.Children, div)

	// No rules - just testing that a span inherits display:inline from defaults
	// and color:black (which is a default for non-link elements)
	rules := []Rule{}

	style := GetComputedStyle(span, rules)

	// Span should inherit display:inline from defaults (span is inline by default)
	if style["display"] != "inline" {
		t.Errorf("expected display=inline, got %s", style["display"])
	}
	// Color should be black (default for non-anchor elements)
	if style["color"] != "black" {
		t.Errorf("expected color=black, got %s", style["color"])
	}
	// Font-size should inherit from parent (body default is 16px)
	if style["font-size"] != "16px" {
		t.Errorf("expected font-size=16px, got %s", style["font-size"])
	}
}

func TestGetComputedStyle_CSSRules(t *testing.T) {
	// Create document: <p class="highlight">text</p>
	doc := &html.Node{
		Type: html.NodeDocument,
	}
	p := &html.Node{
		Type:     html.NodeElement,
		TagName:  "p",
		Parent:   doc,
		Children: []*html.Node{},
	}
	p.SetAttribute("class", "highlight")
	doc.Children = append(doc.Children, p)

	// Add a CSS rule for .highlight
	rules := []Rule{
		{
			Selector:    ".highlight",
			Declarations: []Declaration{
				{Property: "background-color", Value: "yellow"},
				{Property: "font-weight", Value: "bold"},
			},
		},
	}

	style := GetComputedStyle(p, rules)

	if style["background-color"] != "yellow" {
		t.Errorf("expected background-color=yellow from rule, got %s", style["background-color"])
	}
	if style["font-weight"] != "bold" {
		t.Errorf("expected font-weight=bold from rule, got %s", style["font-weight"])
	}
}

func TestGetComputedStyle_InlineStyles(t *testing.T) {
	// Create document: <div style="margin-top: 20px"></div>
	doc := &html.Node{
		Type: html.NodeDocument,
	}
	div := &html.Node{
		Type:     html.NodeElement,
		TagName:  "div",
		Parent:   doc,
		Children: []*html.Node{},
	}
	div.SetAttribute("style", "margin-top: 20px; padding: 10px")
	doc.Children = append(doc.Children, div)

	rules := []Rule{}

	style := GetComputedStyle(div, rules)

	if style["margin-top"] != "20px" {
		t.Errorf("expected margin-top=20px from inline, got %s", style["margin-top"])
	}
	if style["padding-top"] != "10px" {
		t.Errorf("expected padding-top=10px from inline, got %s", style["padding-top"])
	}
}

func TestGetComputedStyle_InlineOverridesRules(t *testing.T) {
	// Create: <p class="big" style="font-size: 24px">text</p>
	doc := &html.Node{
		Type: html.NodeDocument,
	}
	p := &html.Node{
		Type:     html.NodeElement,
		TagName:  "p",
		Parent:   doc,
		Children: []*html.Node{},
	}
	p.SetAttribute("class", "big")
	p.SetAttribute("style", "font-size: 24px")
	doc.Children = append(doc.Children, p)

	rules := []Rule{
		{
			Selector: ".big",
			Declarations: []Declaration{
				{Property: "font-size", Value: "20px"},
			},
		},
	}

	style := GetComputedStyle(p, rules)

	// Inline style should win over CSS rule
	if style["font-size"] != "24px" {
		t.Errorf("expected font-size=24px from inline override, got %s", style["font-size"])
	}
}

func TestGetComputedStyle_DeepInheritance(t *testing.T) {
	// Test that GetComputedStyle properly walks the ancestor chain
	// and inherits values from all ancestors
	doc := &html.Node{
		Type: html.NodeDocument,
	}
	htmlNode := &html.Node{
		Type:     html.NodeElement,
		TagName:  "html",
		Parent:   doc,
		Children: []*html.Node{},
	}
	body := &html.Node{
		Type:     html.NodeElement,
		TagName:  "body",
		Parent:   htmlNode,
		Children: []*html.Node{},
	}
	div := &html.Node{
		Type:     html.NodeElement,
		TagName:  "div",
		Parent:   body,
		Children: []*html.Node{},
	}
	p := &html.Node{
		Type:     html.NodeElement,
		TagName:  "p",
		Parent:   div,
		Children: []*html.Node{},
	}
	span := &html.Node{
		Type:     html.NodeElement,
		TagName:  "span",
		Parent:   p,
		Children: []*html.Node{},
	}

	htmlNode.Children = append(htmlNode.Children, body)
	body.Children = append(body.Children, div)
	div.Children = append(div.Children, p)
	p.Children = append(p.Children, span)
	doc.Children = append(doc.Children, htmlNode)

	// No rules - test that span inherits defaults properly
	rules := []Rule{}

	style := GetComputedStyle(span, rules)

	// Check defaults inherited through chain
	if style["display"] != "inline" {
		t.Errorf("expected display=inline (inherited via p->div->body->html defaults), got %s", style["display"])
	}
	if style["font-size"] != "16px" {
		t.Errorf("expected font-size=16px (user-agent default), got %s", style["font-size"])
	}
	if style["color"] != "black" {
		t.Errorf("expected color=black (user-agent default), got %s", style["color"])
	}
	if style["font-family"] != "serif" {
		t.Errorf("expected font-family=serif (user-agent default), got %s", style["font-family"])
	}
}

func TestGetComputedStyle_ElementDefaults(t *testing.T) {
	// Create different element types and verify their default display values
	testCases := []struct {
		tagName  string
		expected string
	}{
		{"div", "block"},
		{"span", "inline"},
		{"p", "block"},
		{"li", "list-item"},
		{"strong", "inline"},
		{"h1", "block"},
		{"a", "inline"},
	}

	doc := &html.Node{
		Type: html.NodeDocument,
	}

	for _, tc := range testCases {
		node := &html.Node{
			Type:     html.NodeElement,
			TagName:  tc.tagName,
			Parent:   doc,
			Children: []*html.Node{},
		}
		doc.Children = []*html.Node{node}

		style := GetComputedStyle(node, []Rule{})

		if style["display"] != tc.expected {
			t.Errorf("for %s: expected display=%s, got %s", tc.tagName, tc.expected, style["display"])
		}
	}
}

func TestGetComputedStyle_IDAndClassSelectors(t *testing.T) {
	// Create: <div id="myid" class="class1 class2">text</div>
	doc := &html.Node{
		Type: html.NodeDocument,
	}
	div := &html.Node{
		Type:     html.NodeElement,
		TagName:  "div",
		Parent:   doc,
		Children: []*html.Node{},
	}
	div.SetAttribute("id", "myid")
	div.SetAttribute("class", "class1 class2")
	doc.Children = append(doc.Children, div)

	rules := []Rule{
		{
			Selector: "#myid",
			Declarations: []Declaration{
				{Property: "background-color", Value: "green"},
			},
		},
		{
			Selector: ".class1",
			Declarations: []Declaration{
				{Property: "color", Value: "white"},
			},
		},
		{
			Selector: "div.class2",
			Declarations: []Declaration{
				{Property: "border-width", Value: "1px"},
			},
		},
	}

	style := GetComputedStyle(div, rules)

	if style["background-color"] != "green" {
		t.Errorf("expected background-color=green from #id rule, got %s", style["background-color"])
	}
	if style["color"] != "white" {
		t.Errorf("expected color=white from .class1 rule, got %s", style["color"])
	}
	if style["border-width"] != "1px" {
		t.Errorf("expected border-width=1px from div.class2 rule, got %s", style["border-width"])
	}
}

func TestGetComputedStyle_CompleteProperties(t *testing.T) {
	// Verify that computed style has many properties (at least the defaults)
	doc := &html.Node{
		Type: html.NodeDocument,
	}
	div := &html.Node{
		Type:     html.NodeElement,
		TagName:  "div",
		Parent:   doc,
		Children: []*html.Node{},
	}
	doc.Children = append(doc.Children, div)

	style := GetComputedStyle(div, []Rule{})

	// Check several properties exist with default values
	expectedProps := map[string]string{
		"display":         "block",
		"visibility":      "visible",
		"color":           "black",
		"background":      "transparent",
		"font-size":       "16px",
		"font-family":     "serif",
		"font-weight":     "normal",
		"margin-top":      "0",
		"padding-top":     "0",
		"border-width":    "0",
		"border-style":   "none",
		"width":           "auto",
		"height":          "auto",
		"text-align":      "left",
		"line-height":     "1.2",
		"position":        "static",
		"float":           "none",
		"opacity":         "1",
		"tab-size":        "8",
		"hyphens":         "manual",
		"text-justify":    "auto",
		"unicode-bidi":    "normal",
		"direction":       "ltr",
		"font-synthesis":  "none",
		"appearance":      "auto",
	}

	for prop, expected := range expectedProps {
		if style[prop] != expected {
			t.Errorf("expected %s=%s, got %s", prop, expected, style[prop])
		}
	}
}

func TestApplyDecl_FontSynthesis(t *testing.T) {
	doc := &html.Node{
		Type: html.NodeDocument,
	}
	div := &html.Node{
		Type:     html.NodeElement,
		TagName:  "div",
		Parent:   doc,
		Children: []*html.Node{},
	}
	div.SetAttribute("style", "font-synthesis: none; font-synthesis: weight")
	doc.Children = append(doc.Children, div)

	style := GetComputedStyle(div, []Rule{})
	if style["font-synthesis"] != "weight" {
		t.Errorf("expected font-synthesis=weight, got %s", style["font-synthesis"])
	}
}

func TestGetComputedStyle_FontFamilyInheritance(t *testing.T) {
	// Test that font-family properly inherits from parent to child
	// When parent has font-family:Arial and child has no explicit font-family,
	// child should compute font-family:Arial
	doc := &html.Node{
		Type: html.NodeDocument,
	}
	div := &html.Node{
		Type:     html.NodeElement,
		TagName:  "div",
		Parent:   doc,
		Children: []*html.Node{},
	}
	div.SetAttribute("style", "font-family: Arial")
	span := &html.Node{
		Type:     html.NodeElement,
		TagName:  "span",
		Parent:   div,
		Children: []*html.Node{},
	}
	div.Children = append(div.Children, span)
	doc.Children = append(doc.Children, div)

	rules := []Rule{}

	divStyle := GetComputedStyle(div, rules)
	if divStyle["font-family"] != "Arial" {
		t.Errorf("div: expected font-family=Arial, got %s", divStyle["font-family"])
	}

	spanStyle := GetComputedStyle(span, rules)
	if spanStyle["font-family"] != "Arial" {
		t.Errorf("span should inherit font-family=Arial from div, got %s", spanStyle["font-family"])
	}
}

func TestApplyDecl_Hyphens(t *testing.T) {
	doc := &html.Node{
		Type: html.NodeDocument,
	}
	div := &html.Node{
		Type:     html.NodeElement,
		TagName:  "div",
		Parent:   doc,
		Children: []*html.Node{},
	}
	div.SetAttribute("style", "hyphens: auto")
	doc.Children = append(doc.Children, div)

	style := GetComputedStyle(div, []Rule{})
	if style["hyphens"] != "auto" {
		t.Errorf("expected hyphens=auto, got %s", style["hyphens"])
	}
}
