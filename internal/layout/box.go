package layout

import (
	"fmt"

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
	FlexBox
	PositionedBox
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
	m := b.MarginLeft.Value + b.MarginRight.Value
	w := b.ContentW
	t := b.BorderLeft.Value + b.BorderRight.Value
	p := b.PaddingLeft.Value + b.PaddingRight.Value
	return m + w + t + p
}

// TotalHeight returns the full height including margin, border, padding.
func (b *Box) TotalHeight() float64 {
	m := b.MarginTop.Value + b.MarginBottom.Value
	h := b.ContentH
	t := b.BorderTop.Value + b.BorderBottom.Value
	p := b.PaddingTop.Value + b.PaddingBottom.Value
	return m + h + t + p
}

// OuterWidth returns the width from outer edge of margin to outer edge of margin.
func (b *Box) OuterWidth() float64 {
	return b.MarginLeft.Value + b.ContentW + b.BorderLeft.Value +
		b.PaddingLeft.Value + b.PaddingRight.Value + b.BorderRight.Value +
		b.MarginRight.Value
}

// OuterHeight returns the height from outer edge of margin to outer edge of margin.
func (b *Box) OuterHeight() float64 {
	return b.MarginTop.Value + b.ContentH + b.BorderTop.Value +
		b.PaddingTop.Value + b.PaddingBottom.Value + b.BorderBottom.Value +
		b.MarginBottom.Value
}

// BuildLayoutTree converts a DOM tree to a layout tree.
func BuildLayoutTree(doc *html.Node, rules []css.Rule) *Box {
	// Find body
	body := findBody(doc)
	if body == nil {
		return nil
	}

	// Build layout tree recursively
	box := buildBox(body, rules, 0, nil)
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

func buildBox(node *html.Node, rules []css.Rule, depth int, parentStyle map[string]string) *Box {
	if node == nil {
		return nil
	}

	// Text nodes inherit parent style
	if node.Type == html.NodeText {
		style := map[string]string{}
		if parentStyle != nil {
			for k, v := range parentStyle {
				style[k] = v
			}
		}
		return &Box{
			Type:  TextBox,
			Node:  node,
			Style: style,
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

	// Flex container
	if display == "flex" {
		box.Type = FlexBox
	}

	// Positioned elements
	if position := style["position"]; position == "absolute" || position == "relative" || position == "fixed" {
		box.Type = PositionedBox
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

	// Recurse for children (pass our computed style as parent style)
	for _, child := range node.Children {
		childBox := buildBox(child, rules, depth+1, style)
		if childBox != nil {
			box.Children = append(box.Children, childBox)
		}
	}

	return box
}

// String returns a formatted string representation of the layout tree.
func (b *Box) String() string {
	return b.stringWithIndent(0)
}

func (b *Box) stringWithIndent(indent int) string {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	var tagName string
	if b.Node != nil {
		tagName = b.Node.TagName
	} else {
		tagName = "(root)"
	}

	var boxType string
	switch b.Type {
	case BlockBox:
		boxType = "block"
	case InlineBox:
		boxType = "inline"
	case InlineBlock:
		boxType = "inline-block"
	case TextBox:
		boxType = "text"
	}

	s := prefix + fmt.Sprintf("[%s] %s (%.0fx%.0f at (%.0f,%.0f))\n",
		boxType, tagName, b.ContentW, b.ContentH, b.ContentX, b.ContentY)

	for _, child := range b.Children {
		s += child.stringWithIndent(indent + 1)
	}

	return s
}
