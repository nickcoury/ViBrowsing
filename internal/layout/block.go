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

	if box.Type == TextBox {
		text := box.Node.Data
		fontSize := css.ParseLength(box.Style["font-size"]).Value
		if fontSize == 0 {
			fontSize = 16
		}

		// Approximate: 0.6 * font-size per character
		charWidth := fontSize * 0.6
		lineHeight := css.ParseLength(box.Style["line-height"]).Value * fontSize

		// Wrap text
		x := ctx.X

		for range text {
			if x+charWidth > contentWidth && x > ctx.X {
				// Wrap to next line
				x = ctx.X
				ctx.Y += lineHeight
			}
			box.ContentX = x
			box.ContentY = ctx.Y
			box.ContentW = charWidth
			box.ContentH = lineHeight
			x += charWidth
		}

		// Update cursor
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
