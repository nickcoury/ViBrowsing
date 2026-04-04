package layout

import (
	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/html"
)

// BoxType represents the type of layout box.
type BoxType int

const (
	BlockBox BoxType = iota
	InlineBox
	InlineBlock
	TextBox
)

// Box represents a CSS box in the layout tree.
type Box struct {
	Type       BoxType
	Node       *html.Node // source DOM node

	// Content area
	ContentX, ContentY float64
	ContentW, ContentH float64

	// Box dimensions (CSS box model)
	MarginTop, MarginRight, MarginBottom, MarginLeft css.Length
	BorderTop, BorderRight, BorderBottom, BorderLeft css.Length
	PaddingTop, PaddingRight, PaddingBottom, PaddingLeft css.Length

	// Computed style
	Style map[string]string

	// Children
	Children []*Box
}

// GetWidth returns the computed width of the content area.
func (b *Box) GetWidth() float64 {
	if w, ok := b.Style["width"]; ok {
		l := css.ParseLength(w)
		if !l.IsAuto {
			return l.Value
		}
	}
	return b.ContentW
}

// GetHeight returns the computed height of the content area.
func (b *Box) GetHeight() float64 {
	if h, ok := b.Style["height"]; ok {
		l := css.ParseLength(h)
		return l.Value
	}
	return b.ContentH
}

// GetMargin returns the margin lengths.
func (b *Box) GetMargin() (top, right, bottom, left css.Length) {
	return b.MarginTop, b.MarginRight, b.MarginBottom, b.MarginLeft
}

// GetPadding returns the padding lengths.
func (b *Box) GetPadding() (top, right, bottom, left css.Length) {
	return b.PaddingTop, b.PaddingRight, b.PaddingBottom, b.PaddingLeft
}

// GetBorder returns the border widths.
func (b *Box) GetBorder() (top, right, bottom, left css.Length) {
	return b.BorderTop, b.BorderRight, b.BorderBottom, b.BorderLeft
}

// TotalWidth returns the full width including margin, border, padding.
func (b *Box) TotalWidth() float64 {
	m := css.ParseLength(b.Style["margin-left"]).Value +
		css.ParseLength(b.Style["margin-right"]).Value
	w := css.ParseLength(b.Style["width"]).Value
	t := css.ParseLength(b.Style["border-width"]).Value * 2
	p := css.ParseLength(b.Style["padding-left"]).Value +
		css.ParseLength(b.Style["padding-right"]).Value
	return m + w + t + p
}

// TotalHeight returns the full height including margin, border, padding.
func (b *Box) TotalHeight() float64 {
	return b.ContentH
}

// BuildLayoutTree converts a DOM tree to a layout tree.
func BuildLayoutTree(doc *html.Node, rules []css.Rule) *Box {
	// Find body
	body := findBody(doc)
	if body == nil {
		return nil
	}

	// Build layout tree recursively
	box := buildBox(body, rules, 0)
	return box
}

func findBody(n *html.Node) *html.Node {
	if n.Type == html.NodeElement && n.TagName == "body" {
		return n
	}
	for _, child := range n.Children {
		if b := findBody(child); b != nil {
			return b
		}
	}
	return nil
}

func buildBox(node *html.Node, rules []css.Rule, depth int) *Box {
	if node == nil {
		return nil
	}

	// Text nodes
	if node.Type == html.NodeText {
		return &Box{
			Type:  TextBox,
			Node:  node,
			Style: map[string]string{},
			Children: []*Box{},
		}
	}

	// Skip non-element nodes
	if node.Type != html.NodeElement {
		return nil
	}

	// Compute style
	class := node.GetAttribute("class")
	id := node.GetAttribute("id")
	tagName := node.TagName
	inlineDecls := css.ParseInline(node.GetAttribute("style"))
	style := css.ComputeStyle(tagName, class, id, inlineDecls, rules)

	display := style["display"]

	// Create box
	box := &Box{
		Type:  BlockBox,
		Node:  node,
		Style: style,
	}

	// Override display type
	switch display {
	case "block":
		box.Type = BlockBox
	case "inline":
		box.Type = InlineBox
	case "none":
		// Hidden — don't create a box
		return nil
	}

	// Default block elements
	switch tagName {
	case "div", "p", "h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li", "table", "tr", "form", "pre",
		"blockquote", "address", "article", "aside",
		"footer", "header", "main", "nav", "section",
		"figure", "figcaption":
		box.Type = BlockBox
	case "span", "a", "strong", "em", "b", "i", "code", "small":
		box.Type = InlineBox
	}

	// Parse box model properties
	box.MarginTop = css.ParseLength(style["margin-top"])
	box.MarginRight = css.ParseLength(style["margin-right"])
	box.MarginBottom = css.ParseLength(style["margin-bottom"])
	box.MarginLeft = css.ParseLength(style["margin-left"])
	box.PaddingTop = css.ParseLength(style["padding-top"])
	box.PaddingRight = css.ParseLength(style["padding-right"])
	box.PaddingBottom = css.ParseLength(style["padding-bottom"])
	box.PaddingLeft = css.ParseLength(style["padding-left"])

	borderWidth := css.ParseLength(style["border-width"])
	box.BorderTop = borderWidth
	box.BorderRight = borderWidth
	box.BorderBottom = borderWidth
	box.BorderLeft = borderWidth

	// Recurse for children
	for _, child := range node.Children {
		childBox := buildBox(child, rules, depth+1)
		if childBox != nil {
			box.Children = append(box.Children, childBox)
		}
	}

	return box
}
