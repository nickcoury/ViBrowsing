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

	// Float tracking: tracks the float boundaries for content wrapping
	FloatLeftEdge  float64 // right edge of the rightmost left-float (0 if none)
	FloatRightEdge float64 // left edge of the leftmost right-float (Width if none)
	FloatBottom    float64 // lowest bottom edge of all floats in current line

	// Line box tracking for vertical-align
	LineBoxBaseline float64 // Y baseline for current line box (relative to ctx.Y)
	LineBoxMaxAscent float64 // max height above baseline
	LineBoxMaxDescent float64 // max height below baseline
	LineBoxStartY float64 // Y where current line box started
	LineBoxChildren [] *Box // children in current line box for deferred vertical-align
}

// LayoutBlock performs block-level layout on a box tree.
// Returns the total height consumed.
func LayoutBlock(root *Box, containingWidth float64) {
	ctx := &LayoutContext{
		Width:  containingWidth,
		X:      0,
		Y:      0,
		FloatRightEdge: containingWidth,
	}

	layoutChildren(root, ctx)
}

// applyVerticalAlignments positions all children in the current line box
// according to their vertical-align values, relative to the baseline.
func applyVerticalAlignments(ctx *LayoutContext) {
	if ctx.LineBoxChildren == nil || len(ctx.LineBoxChildren) == 0 {
		return
	}

	baseline := ctx.LineBoxStartY + ctx.LineBoxMaxAscent

	for _, child := range ctx.LineBoxChildren {
		align := child.Style["vertical-align"]
		childH := child.ContentH
		childAscent := childH * 0.75 // approximate text ascent ratio
		_ = childH * 0.25 // childDescent kept for future use

		switch align {
		case "top":
			// Align to top of line box
			child.ContentY = ctx.LineBoxStartY
		case "bottom":
			// Align to bottom of line box
			child.ContentY = ctx.LineBoxStartY + ctx.LineBoxMaxAscent + ctx.LineBoxMaxDescent - childH
		case "middle":
			// Align to middle of line box
			child.ContentY = ctx.LineBoxStartY + ctx.LineBoxMaxAscent - childH/2
		case "text-top":
			// Align to top of parent's text
			child.ContentY = ctx.LineBoxStartY
		case "text-bottom":
			// Align to bottom of parent's text
			child.ContentY = ctx.LineBoxStartY + ctx.LineBoxMaxAscent + ctx.LineBoxMaxDescent - childH
		case "sub":
			// Subscript: lower than baseline
			subscriptOffset := childH * 0.3
			child.ContentY = baseline + subscriptOffset - childH + childH*0.1
		case "super":
			// Superscript: higher than baseline
			supOffset := childH * 0.4
			child.ContentY = baseline - supOffset
		case "baseline", "":
			// Default: align baseline
			child.ContentY = baseline - childAscent
		default:
			// Check for length value like "10px" or "1em"
			l := css.ParseLength(align)
			if l.Unit == css.UnitPx || l.Unit == css.UnitEm || l.Unit == css.UnitRem {
				offset := l.Value
				if l.Unit == css.UnitEm {
					offset *= 16 // assume 16px font-size
				}
				child.ContentY = baseline - childH*0.75 + offset
			} else {
				child.ContentY = baseline - childAscent
			}
		}
	}
}

func layoutChildren(box *Box, ctx *LayoutContext) {
	// Track line box state for vertical-align
	ctx.LineBoxBaseline = 0
	ctx.LineBoxMaxAscent = 0
	ctx.LineBoxMaxDescent = 0
	ctx.LineBoxStartY = ctx.Y
	ctx.LineBoxChildren = nil

	for _, child := range box.Children {
		switch child.Type {
		case BlockBox:
			// Flush pending line box before block child
			applyVerticalAlignments(ctx)
			ctx.LineBoxBaseline = 0
			ctx.LineBoxMaxAscent = 0
			ctx.LineBoxMaxDescent = 0
			ctx.LineBoxStartY = ctx.Y
			ctx.LineBoxChildren = nil
			layoutBlockChild(child, ctx)
		case FlexBox:
			applyVerticalAlignments(ctx)
			ctx.LineBoxBaseline = 0
			ctx.LineBoxMaxAscent = 0
			ctx.LineBoxMaxDescent = 0
			ctx.LineBoxStartY = ctx.Y
			ctx.LineBoxChildren = nil
			layoutFlexContainer(child, ctx)
		case PositionedBox:
			applyVerticalAlignments(ctx)
			ctx.LineBoxBaseline = 0
			ctx.LineBoxMaxAscent = 0
			ctx.LineBoxMaxDescent = 0
			ctx.LineBoxStartY = ctx.Y
			ctx.LineBoxChildren = nil
			layoutPositionedChild(child, ctx)
		case InlineBox, TextBox:
			prevY := ctx.Y
			layoutInlineChild(child, box, ctx)
			// Check if we wrapped to a new line
			if ctx.Y > prevY {
				// New line started — flush previous line box
				applyVerticalAlignments(ctx)
				ctx.LineBoxBaseline = 0
				ctx.LineBoxMaxAscent = 0
				ctx.LineBoxMaxDescent = 0
				ctx.LineBoxStartY = ctx.Y
				ctx.LineBoxChildren = nil
			}
		}
	}
	// Flush any remaining line box
	applyVerticalAlignments(ctx)
}

// layoutFloatChild positions a floated element.
func layoutFloatChild(box *Box, ctx *LayoutContext) {
	float := box.Style["float"]
	marginLeft := css.ParseLength(box.Style["margin-left"]).Value
	marginRight := css.ParseLength(box.Style["margin-right"]).Value
	marginTop := css.ParseLength(box.Style["margin-top"]).Value
	_ = css.ParseLength(box.Style["margin-bottom"]).Value // not used for floats

	// Width of the float
	floatWidth := computeWidth(box, ctx.Width)
	if floatWidth > ctx.Width {
		floatWidth = ctx.Width
	}

	if float == "left" {
		// Float left: position at left edge, update left float boundary
		box.ContentX = ctx.X + marginLeft
		box.ContentY = ctx.Y + marginTop
		box.ContentW = floatWidth
		box.ContentH = computeHeight(box, nil)

		// Update float boundary
		newFloatEdge := ctx.X + floatWidth + marginLeft + marginRight
		if newFloatEdge > ctx.FloatLeftEdge {
			ctx.FloatLeftEdge = newFloatEdge
		}
		// Update float bottom
		floatBottom := box.ContentY + box.ContentH + marginTop
		if floatBottom > ctx.FloatBottom {
			ctx.FloatBottom = floatBottom
		}
	} else {
		// Float right: position at right edge, update right float boundary
		box.ContentW = floatWidth
		box.ContentX = ctx.X + ctx.Width - floatWidth - marginRight
		box.ContentY = ctx.Y + marginTop
		box.ContentH = computeHeight(box, nil)

		// Update float boundary
		newFloatEdge := ctx.X + ctx.Width - floatWidth - marginLeft - marginRight
		if newFloatEdge < ctx.FloatRightEdge {
			ctx.FloatRightEdge = newFloatEdge
		}
		// Update float bottom
		floatBottom := box.ContentY + box.ContentH + marginTop
		if floatBottom > ctx.FloatBottom {
			ctx.FloatBottom = floatBottom
		}
	}
}

func layoutBlockChild(box *Box, ctx *LayoutContext) {
	float := box.Style["float"]

	// Handle float elements
	if float == "left" || float == "right" {
		layoutFloatChild(box, ctx)
		return
	}

	// Compute width (accounting for floats if they exist)
	availWidth := ctx.Width
	if ctx.FloatLeftEdge > 0 {
		availWidth -= ctx.FloatLeftEdge
	}
	if ctx.FloatRightEdge < ctx.Width {
		availWidth -= ctx.Width - ctx.FloatRightEdge
	}
	if availWidth < 50 {
		availWidth = 50
	}
	width := computeWidth(box, availWidth)

	// Margin collapsing: top margin of first block child collapses with parent's top
	marginTop := css.ParseLength(box.Style["margin-top"]).Value
	marginBottom := css.ParseLength(box.Style["margin-bottom"]).Value

	// Compute y position: start at FloatBottom if we're below floats
	// (ctx.Y might be below floats already, or at original cursor)
	box.ContentY = ctx.Y + marginTop

	// Compute x position respecting floats
	xPos := ctx.X
	if ctx.FloatLeftEdge > 0 {
		xPos = ctx.FloatLeftEdge
	}
	box.ContentX = xPos + css.ParseLength(box.Style["margin-left"]).Value

	// Check if box overlaps floats — if so, move to next line below FloatBottom
	boxEndX := box.ContentX + width
	boxOverlapsLeftFloat := ctx.FloatLeftEdge > 0 && boxEndX > ctx.FloatLeftEdge
	boxOverlapsRightFloat := ctx.FloatRightEdge < ctx.Width && box.ContentX < ctx.FloatRightEdge
	if boxOverlapsLeftFloat || boxOverlapsRightFloat {
		// Clear floats and move below
		ctx.FloatLeftEdge = 0
		ctx.FloatRightEdge = ctx.Width
		ctx.FloatBottom = 0
		box.ContentY = ctx.Y + marginTop
		box.ContentX = ctx.X + css.ParseLength(box.Style["margin-left"]).Value
		// Recompute width after float clearing
		width = computeWidth(box, ctx.Width)
		box.ContentW = width
	}

	// Compute content height
	childCtx := &LayoutContext{
		Width:         width,
		X:             box.ContentX,
		Y:             box.ContentY,
		FloatLeftEdge: ctx.FloatLeftEdge,
		FloatRightEdge: ctx.FloatRightEdge,
		FloatBottom:   ctx.FloatBottom,
	}
	layoutChildren(box, childCtx)

	// Content height = position of last child
	box.ContentW = width
	box.ContentH = computeHeight(box, childCtx)

	// Advance cursor: use max of float bottom and block bottom
	nextY := box.ContentY + box.ContentH + marginBottom
	if ctx.FloatBottom > nextY {
		nextY = ctx.FloatBottom
	}
	ctx.Y = nextY

	// Clear floats after a block that was pushed below them
	if ctx.FloatBottom > 0 && box.ContentY+box.ContentH >= ctx.FloatBottom {
		// If this block ends at or below FloatBottom, clear floats
		// so subsequent blocks don't try to flow around cleared floats
	}
}

func layoutInlineChild(box *Box, parent *Box, ctx *LayoutContext) {
	// Inline layout: wrap at container width
	contentWidth := parent.ContentW
	if contentWidth <= 0 {
		contentWidth = 800
	}

	// Handle <br> element - forced line break
	if box.Type == InlineBox && box.Node != nil && box.Node.TagName == "br" {
		fontSize := css.ParseLength(box.Style["font-size"]).Value
		if fontSize == 0 {
			fontSize = 16
		}
		lineHeight := css.ParseLength(box.Style["line-height"]).Value
		if lineHeight == 0 {
			lineHeight = 1.2
		}
		lineHeightPx := lineHeight * fontSize

		ctx.X = ctx.X
		ctx.Y += lineHeightPx
		// Flush previous line box
		applyVerticalAlignments(ctx)
		ctx.LineBoxBaseline = 0
		ctx.LineBoxMaxAscent = 0
		ctx.LineBoxMaxDescent = 0
		ctx.LineBoxStartY = ctx.Y
		ctx.LineBoxChildren = nil
		box.ContentW = 0
		box.ContentH = 0
		return
	}

	// Handle <wbr> element - zero-width break opportunity
	if box.Type == InlineBox && box.Node != nil && box.Node.TagName == "wbr" {
		// wbr is a zero-width break point - mark it but don't force break
		box.ContentW = 0
		box.ContentH = 0
		return
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
			wordWrap := box.Style["word-wrap"]
			overflowWrap := box.Style["overflow-wrap"]
			canWrap := !isPre
			// word-wrap: break-word allows breaking even without hyphens/spaces
			canBreakWord := wordWrap == "break-word" || overflowWrap == "break-word"

			if canWrap && x+cw > contentWidth && x > ctx.X {
				x = ctx.X
				ctx.Y += lineHeightPx
				wrapPrevWord = false
				// Skip space at start of new line (normal mode)
				if c == ' ' {
					continue
				}
			} else if canBreakWord && x+cw > contentWidth && x > ctx.X {
				// Break long word without spaces
				x = ctx.X
				ctx.Y += lineHeightPx
				wrapPrevWord = false
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
		box.ContentY = ctx.LineBoxStartY
		box.ContentW = x - startX
		if box.ContentW < charWidth {
			box.ContentW = charWidth
		}
		box.ContentH = lineHeightPx

		// Track line box metrics
		if box.ContentH > ctx.LineBoxMaxAscent+ctx.LineBoxMaxDescent {
			// New total line height — adjust
			extra := box.ContentH - (ctx.LineBoxMaxAscent + ctx.LineBoxMaxDescent)
			ctx.LineBoxMaxDescent += extra
		}

		// Add to line box
		ctx.LineBoxChildren = append(ctx.LineBoxChildren, box)

		// Advance cursor past the text
		if x > ctx.X {
			ctx.X = x
		}
	} else {
		// Inline element box
		marginLeft := css.ParseLength(box.Style["margin-left"]).Value
		marginRight := css.ParseLength(box.Style["margin-right"]).Value
		width := computeWidth(box, ctx.Width)

		// Check if we need to wrap first
		totalWidth := marginLeft + width + marginRight
		if ctx.X+totalWidth > contentWidth && ctx.X > 0 {
			// Wrap to next line
			ctx.Y += ctx.LineBoxMaxAscent + ctx.LineBoxMaxDescent
			ctx.X = 0
			// Flush previous line box
			applyVerticalAlignments(ctx)
			ctx.LineBoxBaseline = 0
			ctx.LineBoxMaxAscent = 0
			ctx.LineBoxMaxDescent = 0
			ctx.LineBoxStartY = ctx.Y
			ctx.LineBoxChildren = nil
		}

		box.ContentX = ctx.X + marginLeft
		box.ContentY = ctx.LineBoxStartY
		box.ContentW = width
		box.ContentH = computeHeight(box, nil)

		// Add to line box for vertical-align
		ctx.LineBoxChildren = append(ctx.LineBoxChildren, box)

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
	case PositionedBox:
		layoutPositionedChild(child, ctx)
	case ImageBox:
		layoutImageChild(child, ctx)
	case ListItemBox:
		layoutListItemChild(child, ctx)
	default:
		layoutInlineChild(child, parent, ctx)
	}
}

// layoutImageChild handles layout for img elements.
func layoutImageChild(box *Box, ctx *LayoutContext) {
	if box.Node == nil {
		return
	}

	width := computeWidth(box, ctx.Width)
	height := css.ParseLength(box.Style["height"]).Value

	box.ContentX = ctx.X + css.ParseLength(box.Style["margin-left"]).Value
	box.ContentY = ctx.Y + css.ParseLength(box.Style["margin-top"]).Value
	box.ContentW = width
	box.ContentH = height

	marginBottom := css.ParseLength(box.Style["margin-bottom"]).Value
	ctx.Y += box.ContentH + marginBottom
}

// layoutListItemChild handles layout for li elements with list markers.
func layoutListItemChild(box *Box, ctx *LayoutContext) {
	marginTop := css.ParseLength(box.Style["margin-top"]).Value
	marginBottom := css.ParseLength(box.Style["margin-bottom"]).Value
	marginLeft := css.ParseLength(box.Style["margin-left"]).Value

	// Marker width (space for bullet/number)
	markerWidth := 20.0
	listStyleType := box.Style["list-style-type"]

	// Determine if we're in an ordered list
	parentIsOl := false
	if box.Parent != nil && box.Parent.Node != nil && box.Parent.Node.TagName == "ol" {
		parentIsOl = true
	}

	// Simple marker rendering: draw bullet/number before content
	box.ContentX = ctx.X + marginLeft + markerWidth
	box.ContentY = ctx.Y + marginTop
	box.ContentW = computeWidth(box, ctx.Width) - markerWidth

	// Compute height from children
	childCtx := &LayoutContext{
		Width:         box.ContentW,
		X:             box.ContentX,
		Y:             box.ContentY,
		FloatLeftEdge: ctx.FloatLeftEdge,
		FloatRightEdge: ctx.FloatRightEdge,
		FloatBottom:   ctx.FloatBottom,
	}
	layoutChildren(box, childCtx)

	box.ContentH = computeHeight(box, childCtx)

	// Store marker info in style for rendering
	box.Style["_marker"] = listStyleType
	if parentIsOl {
		box.Style["_marker"] = "ol:" + listStyleType
	}
	box.Style["_markerWidth"] = strconv.FormatFloat(markerWidth, 'f', 0, 64)

	nextY := box.ContentY + box.ContentH + marginBottom
	if ctx.FloatBottom > nextY {
		nextY = ctx.FloatBottom
	}
	ctx.Y = nextY
}

// layoutPositionedChild handles position:absolute/relative/fixed layout.
func layoutPositionedChild(box *Box, ctx *LayoutContext) {
	position := box.Style["position"]

	if position == "fixed" {
		// Fixed is relative to viewport
		ctx = &LayoutContext{
			Width: 800,
			Height: 600,
			X:     0,
			Y:     0,
		}
	}

	top := css.ParseLength(box.Style["top"])
	left := css.ParseLength(box.Style["left"])
	_ = css.ParseLength(box.Style["right"])  // right/bottom not yet used
	_ = css.ParseLength(box.Style["bottom"])

	width := computeWidth(box, ctx.Width)

	if position == "relative" {
		// Relative: offset from normal position
		box.ContentX = ctx.X + css.ParseLength(box.Style["margin-left"]).Value
		box.ContentY = ctx.Y + css.ParseLength(box.Style["margin-top"]).Value
		if !left.IsAuto {
			box.ContentX = ctx.X + left.Value
		}
		if !top.IsAuto {
			box.ContentY = ctx.Y + top.Value
		}
	} else {
		// Absolute: removed from flow, positioned relative to containing block
		box.ContentX = ctx.X
		box.ContentY = ctx.Y
		if !left.IsAuto {
			box.ContentX = ctx.X + left.Value
		}
		if !top.IsAuto {
			box.ContentY = ctx.Y + top.Value
		}
	}

	box.ContentW = width
	box.ContentH = computeHeight(box, ctx)

	// Layout children within positioned box
	childCtx := &LayoutContext{
		Width: box.ContentW,
		X:     box.ContentX,
		Y:     box.ContentY,
	}
	layoutChildren(box, childCtx)
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
