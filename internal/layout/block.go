package layout

import (
	"strconv"

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
		case FlexBox:
			layoutFlexContainer(child, ctx)
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

// layoutFlexContainer performs flex layout for a flex container box.
func layoutFlexContainer(box *Box, ctx *LayoutContext) {
	width := computeWidth(box, ctx.Width)
	marginTop := css.ParseLength(box.Style["margin-top"]).Value
	marginBottom := css.ParseLength(box.Style["margin-bottom"]).Value

	box.ContentX = ctx.X + css.ParseLength(box.Style["margin-left"]).Value
	box.ContentY = ctx.Y + marginTop
	box.ContentW = width

	flexDirection := box.Style["flex-direction"]
	justifyContent := box.Style["justify-content"]
	alignItems := box.Style["align-items"]
	gapStr := box.Style["gap"]

	isRow := flexDirection == "row" || flexDirection == "" || flexDirection == "row-reverse"
	isReverse := flexDirection == "row-reverse" || flexDirection == "column-reverse"

	// Parse gap
	gap := 0.0
	if gapStr != "" {
		gap = css.ParseLength(gapStr).Value
	}

	// Collect flex items and calculate total flex and available space
	type flexItem struct {
		box    *Box
		minW   float64
		flexW  float64
		flexG  float64
	}
	var items []flexItem
	for _, child := range box.Children {
		if child.Style["display"] == "none" {
			continue
		}
		flexGrow := 0.0
		if fg, ok := child.Style["flex-grow"]; ok {
			flexGrow, _ = strconv.ParseFloat(fg, 64)
		}
		flexBasis := css.ParseLength(child.Style["flex-basis"]).Value
		if flexBasis == 0 && child.Style["flex-basis"] != "0" {
			flexBasis = css.ParseLength(child.Style["width"]).Value
		}
		itemMinW := css.ParseLength(child.Style["width"]).Value
		if itemMinW == 0 {
			itemMinW = 50 // fallback min size
		}
		items = append(items, flexItem{
			box:    child,
			minW:   itemMinW,
			flexW:  flexBasis,
			flexG:  flexGrow,
		})
	}

	if len(items) == 0 {
		box.ContentH = 0
		ctx.Y += marginTop + marginBottom
		return
	}

	// Two-pass flex layout
	// Pass 1: determine initial sizes (flex-basis or min width)
	// Pass 2: resolve flexible lengths

	// For row flex, mainAxis = width, crossAxis = height
	// For column flex, mainAxis = height, crossAxis = width
	mainSize := width
	if !isRow {
		mainSize = ctx.Height
		if mainSize == 0 {
			mainSize = 600
		}
	}

	// First pass: calculate flex base sizes
	totalFlex := 0.0
	for _, item := range items {
		if item.flexW > 0 {
			totalFlex += item.flexW
		}
	}

	// Calculate available space
	usedSpace := 0.0
	for _, item := range items {
		if item.flexW > 0 {
			usedSpace += item.flexW
		} else {
			usedSpace += item.minW
		}
	}
	available := mainSize - usedSpace - gap*float64(len(items)-1)
	if available < 0 {
		available = 0
	}

	// Second pass: resolve flexible items
	for i := range items {
		item := &items[i]
		if item.flexG > 0 && totalFlex > 0 {
			item.flexW += available * item.flexG / totalFlex
		} else if item.flexW == 0 {
			item.flexW = item.minW
		}
	}

	// Calculate total main axis size
	totalMain := 0.0
	for _, item := range items {
		totalMain += item.flexW
	}
	totalMain += gap * float64(len(items)-1)

	// Calculate cross axis size (max of item heights for row flex)
	maxCross := 0.0
	for _, item := range items {
		itemH := computeHeight(item.box, nil)
		if itemH > maxCross {
			maxCross = itemH
		}
	}

	// Calculate initial offset for justify-content
	var mainCursor float64

	switch justifyContent {
	case "center":
		mainCursor = (mainSize - totalMain) / 2
	case "flex-end":
		mainCursor = mainSize - totalMain
	case "space-between":
		if len(items) > 1 {
			mainCursor = 0
		}
	case "space-around":
		mainCursor = (mainSize - totalMain) / 2
	default: // "flex-start" or ""
		mainCursor = 0
	}

	if isReverse {
		mainCursor = mainSize - mainCursor
	}

	// Layout each flex item
	itemCtx := &LayoutContext{
		Width:  width,
		Height: maxCross,
		X:      box.ContentX,
		Y:      box.ContentY,
	}

	for i, item := range items {
		child := item.box

		if isRow {
			// Row flex: horizontal main axis
			childX := box.ContentX + mainCursor
			if isReverse {
				childX = box.ContentX + mainSize - mainCursor - item.flexW
			}
			child.ContentX = childX
			child.ContentY = box.ContentY
			child.ContentW = item.flexW
			child.ContentH = maxCross

			// Align self
			align := child.Style["align-self"]
			if align == "" {
				align = alignItems
			}
			switch align {
			case "center":
				child.ContentY = box.ContentY + (maxCross-child.ContentH)/2
			case "flex-end":
				child.ContentY = box.ContentY + maxCross - child.ContentH
			}

			// Position children within flex item
			for _, grandchild := range child.Children {
				layoutChild(grandchild, child, itemCtx)
			}

			if isReverse {
				mainCursor -= gap + item.flexW
			} else {
				mainCursor += gap + item.flexW
			}
		} else {
			// Column flex: vertical main axis
			childY := box.ContentY + mainCursor
			if isReverse {
				childY = box.ContentY + mainSize - mainCursor - item.flexW
			}
			child.ContentX = box.ContentX
			child.ContentY = childY
			child.ContentW = width
			child.ContentH = item.flexW

			// Align self
			align := child.Style["align-self"]
			if align == "" {
				align = alignItems
			}
			switch align {
			case "center":
				child.ContentX = box.ContentX + (width-child.ContentW)/2
			case "flex-end":
				child.ContentX = box.ContentX + width - child.ContentW
			}

			// Position children within flex item
			itemCtx.Width = width
			itemCtx.X = child.ContentX
			itemCtx.Y = child.ContentY
			for _, grandchild := range child.Children {
				layoutChild(grandchild, child, itemCtx)
			}

			if isReverse {
				mainCursor -= gap + item.flexW
			} else {
				mainCursor += gap + item.flexW
			}
		}
		_ = i // silence unused variable
	}

	box.ContentH = maxCross
	if !isRow {
		box.ContentH = mainCursor
		if mainCursor < 0 {
			box.ContentH = mainSize
		}
	}

	ctx.Y += marginTop + box.ContentH + marginBottom
}

// layoutChild dispatches layout based on child box type.
func layoutChild(child *Box, parent *Box, ctx *LayoutContext) {
	if child == nil {
		return
	}
	if child.Style["display"] == "none" {
		return
	}
	switch child.Type {
	case BlockBox:
		layoutBlockChild(child, ctx)
	case FlexBox:
		layoutFlexContainer(child, ctx)
	default:
		layoutInlineChild(child, parent, ctx)
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
