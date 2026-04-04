package layout

import (
	"github.com/nickcoury/ViBrowsing/internal/css"
)

// LayoutContext holds layout state for a single layout pass.
type LayoutContext struct {
	Width  float64 // containing block width
	Height float64
	X, Y   float64 // cursor position
}

// LayoutBlock performs block-level layout on a box tree.
// Returns the total height consumed.
func LayoutBlock(root *Box, containingWidth float64) {
	ctx := &LayoutContext{
		Width: containingWidth,
		X:     0,
		Y:     0,
	}

	layoutChildren(root, ctx)
}

func layoutChildren(box *Box, ctx *LayoutContext) {
	for _, child := range box.Children {
		switch child.Type {
		case BlockBox:
			layoutBlockChild(child, ctx)
		case InlineBox, TextBox:
			layoutInlineChild(child, box, ctx)
		}
	}
}

func layoutBlockChild(box *Box, ctx *LayoutContext) {
	// Compute width
	width := computeWidth(box, ctx.Width)

	// Margin collapsing: top margin of first block child collapses with parent's top
	marginTop := css.ParseLength(box.Style["margin-top"]).Value
	marginBottom := css.ParseLength(box.Style["margin-bottom"]).Value

	// Position
	box.ContentX = ctx.X + css.ParseLength(box.Style["margin-left"]).Value
	box.ContentY = ctx.Y + marginTop

	// Compute content height
	childCtx := &LayoutContext{
		Width: width,
		X:     box.ContentX,
		Y:     box.ContentY,
	}
	layoutChildren(box, childCtx)

	// Content height = position of last child
	box.ContentW = width
	box.ContentH = computeHeight(box, childCtx)

	// Advance cursor
	ctx.Y += marginTop + box.ContentH + marginBottom
}

func layoutInlineChild(box *Box, parent *Box, ctx *LayoutContext) {
	// Inline layout: wrap at container width
	contentWidth := parent.ContentW
	if contentWidth <= 0 {
		contentWidth = 800
	}

	if box.Type == TextBox {
		text := box.Node.Data
		whiteSpace := box.Style["white-space"]
		fontSize := css.ParseLength(box.Style["font-size"]).Value
		if fontSize == 0 {
			fontSize = 16
		}
		lineHeight := css.ParseLength(box.Style["line-height"]).Value
		if lineHeight == 0 {
			lineHeight = 1.2
		}
		lineHeightPx := lineHeight * fontSize
		charWidth := fontSize * 0.6

		x := ctx.X
		startX := x
		startY := ctx.Y
		isPre := whiteSpace == "pre" || whiteSpace == "pre-wrap"
		wrapPrevWord := false

		for i := 0; i < len(text); i++ {
			c := text[i]

			// Handle explicit newlines in pre mode
			if c == '\n' {
				if isPre {
					x = ctx.X
					ctx.Y += lineHeightPx
					x = ctx.X
					wrapPrevWord = false
					continue
				}
				// In normal mode, newline collapses to a space
				c = ' '
			}
			if c == '\r' {
				continue
			}
			if c == '\t' {
				c = ' '
			}

			cw := charWidth
			if c == ' ' {
				if !isPre && wrapPrevWord {
					// Skip leading whitespace after a wrap
					continue
				}
				wrapPrevWord = false
			}

			// Wrap line if needed (not in pre mode)
			canWrap := !isPre
			if canWrap && x+cw > contentWidth && x > ctx.X {
				x = ctx.X
				ctx.Y += lineHeightPx
				wrapPrevWord = false
				// Skip space at start of new line (normal mode)
				if c == ' ' {
					continue
				}
			}

			x += cw
			if c != ' ' {
				wrapPrevWord = true
			}
		}

		// Set box to cover the entire text range
		box.ContentX = startX
		box.ContentY = startY
		box.ContentW = x - startX
		if box.ContentW < charWidth {
			box.ContentW = charWidth
		}
		box.ContentH = ctx.Y + lineHeightPx - startY
		if box.ContentH < lineHeightPx {
			box.ContentH = lineHeightPx
		}

		// Advance cursor past the text
		if x > ctx.X {
			ctx.X = x
		}
	} else {
		// Inline element box
		marginLeft := css.ParseLength(box.Style["margin-left"]).Value
		marginRight := css.ParseLength(box.Style["margin-right"]).Value
		width := computeWidth(box, ctx.Width)

		box.ContentX = ctx.X + marginLeft
		box.ContentY = ctx.Y
		box.ContentW = width
		box.ContentH = computeHeight(box, nil)

		ctx.X += marginLeft + width + marginRight
	}
}

func computeWidth(box *Box, containingWidth float64) float64 {
	widthStr := box.Style["width"]

	// Handle auto
	if widthStr == "auto" || widthStr == "" {
		marginLeft := css.ParseLength(box.Style["margin-left"]).Value
		marginRight := css.ParseLength(box.Style["margin-right"]).Value
		paddingLeft := css.ParseLength(box.Style["padding-left"]).Value
		paddingRight := css.ParseLength(box.Style["padding-right"]).Value
		borderLeft := css.ParseLength(box.Style["border-width"]).Value
		borderRight := css.ParseLength(box.Style["border-width"]).Value

		return containingWidth - marginLeft - marginRight - paddingLeft - paddingRight - borderLeft - borderRight
	}

	l := css.ParseLength(widthStr)
	switch l.Unit {
	case css.UnitPercent:
		return containingWidth * l.Value / 100
	default:
		return l.Value
	}
}

func computeHeight(box *Box, childCtx *LayoutContext) float64 {
	hStr := box.Style["height"]
	if hStr != "auto" && hStr != "" {
		l := css.ParseLength(hStr)
		if l.Unit == css.UnitPercent {
			return l.Value // Use as-is for now
		}
		return l.Value
	}

	// Auto height: sum of children
	if childCtx != nil {
		return childCtx.Y
	}
	return 0
}
