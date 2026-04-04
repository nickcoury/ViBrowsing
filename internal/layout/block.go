package layout

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nickcoury/ViBrowsing/internal/css"
)

// LayoutContext holds layout state for a single layout pass.
type LayoutContext struct {
	Width  float64 // containing block width
	Height float64
	X, Y   float64 // cursor position
	FontSize float64 // current font size for em unit resolution

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
		FontSize:        16,
	}

	layoutChildren(root, ctx)

	// Set root dimensions after laying out children
	// For the root (body), ContentW is the containing width and ContentH is the total height
	root.ContentW = containingWidth
	root.ContentH = ctx.Y
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
			// Reset X cursor to content area edge for block children
			// Inline content may have advanced ctx.X past the content area
			ctx.X = box.ContentX
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
		case TableBox:
			applyVerticalAlignments(ctx)
			ctx.LineBoxBaseline = 0
			ctx.LineBoxMaxAscent = 0
			ctx.LineBoxMaxDescent = 0
			ctx.LineBoxStartY = ctx.Y
			ctx.LineBoxChildren = nil
			layoutTableContainer(child, ctx)
		case TableRowBox, TableSectionBox:
			applyVerticalAlignments(ctx)
			ctx.LineBoxBaseline = 0
			ctx.LineBoxMaxAscent = 0
			ctx.LineBoxMaxDescent = 0
			ctx.LineBoxStartY = ctx.Y
			ctx.LineBoxChildren = nil
			layoutTableRow(child, ctx)
		case TableCellBox, TableCaptionBox:
			applyVerticalAlignments(ctx)
			ctx.LineBoxBaseline = 0
			ctx.LineBoxMaxAscent = 0
			ctx.LineBoxMaxDescent = 0
			ctx.LineBoxStartY = ctx.Y
			ctx.LineBoxChildren = nil
			layoutTableCell(child, ctx)
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
		case ButtonBox, SelectBox, TextAreaBox, LabelBox, MediaBox:
			// These are inline-level elements treated like inline boxes
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

	// Check for multi-column layout (CSS columns)
	columnCount := 1
	columnWidth := 0.0
	columnGap := 0.0
	if cc, ok := box.Style["column-count"]; ok && cc != "" && cc != "auto" {
		if c, err := strconv.Atoi(cc); err == nil && c > 0 {
			columnCount = c
		}
	}
	if cw, ok := box.Style["column-width"]; ok && cw != "" && cw != "auto" {
		if l := css.ParseLength(cw); l.Value > 0 {
			columnWidth = l.Value
		}
	}
	if cg, ok := box.Style["column-gap"]; ok && cg != "" && cg != "normal" {
		if l := css.ParseLength(cg); l.Value > 0 {
			columnGap = l.Value
		}
	} else {
		columnGap = 16 // CSS normal gap
	}

	// Handle multi-column layout
	if columnWidth > 0 || columnCount > 1 {
		layoutMultiColumn(box, ctx, columnCount, columnWidth, columnGap)
		return
	}

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

	// Resolve em units in margins using the parent's font-size
	parentFontSize := 16.0
	if ctx.FontSize > 0 {
		parentFontSize = ctx.FontSize
	}
	if marginTop > 0 && css.ParseLength(box.Style["margin-top"]).Unit == css.UnitEm {
		marginTop = marginTop * parentFontSize
	}
	if marginBottom > 0 && css.ParseLength(box.Style["margin-bottom"]).Unit == css.UnitEm {
		marginBottom = marginBottom * parentFontSize
	}

	// Compute y position: start at FloatBottom if we're below floats
	// (ctx.Y might be below floats already, or at original cursor)
	box.ContentY = ctx.Y + marginTop

	// Compute x position: reset to 0 to avoid inline cursor leaking into block positioning
	xPos := 0.0
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
	// DEBUG: trace ctx.Y for h1 and div
	if box.Node != nil && (box.Node.TagName == "h1" || box.Node.TagName == "div") {
		fmt.Printf("DEBUG [%s] ENTRY: ctx.Y=%.1f marginTop=%.1f\n", box.Node.TagName, ctx.Y, marginTop)
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

		// Resolve font-size to pixels, handling em units
		var getAncestorFontSize func(b *Box) float64
		getAncestorFontSize = func(b *Box) float64 {
			cur := b.Parent
			for cur != nil {
				if cur.Type == BlockBox || cur.Type == FlexBox || cur.Type == PositionedBox {
					l := css.ParseLength(cur.Style["font-size"])
					switch l.Unit {
					case css.UnitEm:
						parentFS := getAncestorFontSize(cur)
						return l.Value * parentFS
					case css.UnitRem:
						return l.Value * 16
					default:
						return l.Value
					}
				}
				cur = cur.Parent
			}
			return 16
		}
		effectiveParentFontSize := getAncestorFontSize(box)
		if effectiveParentFontSize == 0 {
			effectiveParentFontSize = 16
		}
		l := css.ParseLength(box.Style["font-size"])
		var fontSize float64
		switch l.Unit {
		case css.UnitEm:
			fontSize = l.Value * effectiveParentFontSize
		case css.UnitRem:
			fontSize = l.Value * 16
		default:
			fontSize = l.Value
		}
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
	alignContent := box.Style["align-content"]
	gapStr := box.Style["gap"]
	flexWrap := box.Style["flex-wrap"]

	isRow := flexDirection == "row" || flexDirection == "" || flexDirection == "row-reverse"
	isReverse := flexDirection == "row-reverse" || flexDirection == "column-reverse"
	isWrap := flexWrap == "wrap" || flexWrap == "wrap-reverse"
	isWrapReverse := flexWrap == "wrap-reverse"

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
		flexS  float64
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
		flexShrink := 1.0
		if fs, ok := child.Style["flex-shrink"]; ok {
			flexShrink, _ = strconv.ParseFloat(fs, 64)
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
			flexS:  flexShrink,
		})
	}

	if len(items) == 0 {
		box.ContentH = 0
		ctx.Y += marginTop + marginBottom
		return
	}

	// For row flex, mainAxis = width, crossAxis = height
	// For column flex, mainAxis = height, crossAxis = width
	mainSize := width
	crossSize := ctx.Height
	if !isRow {
		mainSize = ctx.Height
		if mainSize == 0 {
			mainSize = 600
		}
	}
	if crossSize == 0 {
		crossSize = 600
	}

	// Build flex lines (when wrapping is enabled)
	type flexLine struct {
		items        []flexItem
		mainSize     float64
		crossSize    float64
		totalFlexG   float64
		available    float64
	}

	var lines []flexLine

	if isWrap {
		// Multi-line: collect items into lines
		var currentLine []flexItem
		currentLineMain := 0.0
		for _, item := range items {
			// Check if item fits in current line
			itemMain := item.flexW
			if itemMain == 0 {
				itemMain = item.minW
			}
			if currentLineMain+itemMain+gap > mainSize && len(currentLine) > 0 {
				// Start new line
				lines = append(lines, flexLine{
					items:     currentLine,
					mainSize:  currentLineMain - gap,
					crossSize: 0,
				})
				currentLine = nil
				currentLineMain = 0
			}
			if len(currentLine) > 0 {
				currentLineMain += gap
			}
			currentLine = append(currentLine, item)
			currentLineMain += itemMain
		}
		if len(currentLine) > 0 {
			lines = append(lines, flexLine{
				items:    currentLine,
				mainSize: currentLineMain - gap,
				crossSize: 0,
			})
		}
	} else {
		// Single line (no-wrap)
		totalMain := 0.0
		for _, item := range items {
			itemMain := item.flexW
			if itemMain == 0 {
				itemMain = item.minW
			}
			totalMain += itemMain
		}
		totalMain += gap * float64(len(items)-1)
		lines = append(lines, flexLine{
			items:     items,
			mainSize:  totalMain,
			crossSize: 0,
		})
	}

	// Calculate cross size for each line and resolve flexible lengths
	for li := range lines {
		line := &lines[li]
		totalFlex := 0.0
		usedSpace := 0.0
		for _, item := range line.items {
			if item.flexW > 0 {
				totalFlex += item.flexW
				usedSpace += item.flexW
			} else {
				usedSpace += item.minW
			}
		}
		line.totalFlexG = totalFlex
		line.available = mainSize - usedSpace - gap*float64(len(line.items)-1)
		if line.available < 0 {
			line.available = 0
		}

		// Resolve flexible lengths
		for i := range line.items {
			item := &line.items[i]
			if item.flexG > 0 && totalFlex > 0 {
				item.flexW += line.available * item.flexG / totalFlex
			} else if item.flexW == 0 {
				item.flexW = item.minW
			}
		}

		// Calculate line cross size (max of item cross-axis sizes)
		for _, item := range line.items {
			itemH := computeHeight(item.box, nil)
			if itemH > line.crossSize {
				line.crossSize = itemH
			}
		}
	}

	// Total cross size (sum of all line cross sizes + gaps between lines)
	totalCross := 0.0
	for _, line := range lines {
		totalCross += line.crossSize
	}
	if len(lines) > 1 {
		totalCross += gap * float64(len(lines)-1)
	}

	// Calculate cross-axis offset for align-content
	// First stretch lines if needed, then calculate offsets
	if alignContent == "stretch" && totalCross < crossSize {
		// Stretch lines to fill cross axis
		stretchExtra := (crossSize - totalCross) / float64(len(lines))
		for li := range lines {
			lines[li].crossSize += stretchExtra
		}
		totalCross = crossSize
	}

	var crossCursor float64
	switch alignContent {
	case "flex-end":
		crossCursor = crossSize - totalCross
	case "center":
		crossCursor = (crossSize - totalCross) / 2
	case "space-between":
		// space-between: distribute lines evenly, first at start, last at end
		// crossCursor starts at 0, handled in line positioning loop
		crossCursor = 0
	case "space-around":
		// space-around: equal space around each line
		space := (crossSize - totalCross) / float64(len(lines)*2)
		crossCursor = space
	case "stretch":
		crossCursor = 0
	default: // "flex-start" or "normal" or ""
		crossCursor = 0
	}

	if isWrapReverse {
		// wrap-reverse flips the cross-axis, so last line is at cross-start
		crossCursor = crossSize - totalCross
	}

	// Layout each flex item
	itemCtx := &LayoutContext{
		Width:  width,
		Height: crossSize,
		X:      box.ContentX,
		Y:      box.ContentY,
	}

	box.ContentH = totalCross

	// Position each line
	for li, line := range lines {
		// Calculate main-axis offset for justify-content
		var mainCursor float64
		totalMain := 0.0
		for _, item := range line.items {
			totalMain += item.flexW
		}
		totalMain += gap * float64(len(line.items)-1)

		switch justifyContent {
		case "center":
			mainCursor = (mainSize - totalMain) / 2
		case "flex-end":
			mainCursor = mainSize - totalMain
		case "space-between":
			if len(line.items) > 1 {
				mainCursor = 0
			}
		case "space-around":
			mainCursor = (mainSize - totalMain) / 2
		case "space-evenly":
			if len(line.items) > 1 {
				mainCursor = (mainSize - totalMain) / float64(len(line.items)+1)
			} else {
				mainCursor = (mainSize - totalMain) / 2
			}
		default: // "flex-start" or ""
			mainCursor = 0
		}

		if isReverse {
			mainCursor = mainSize - mainCursor
		}

		// Position each item in this line
		for _, item := range line.items {
			child := item.box
			itemCtx.Height = line.crossSize

			if isRow {
				childX := box.ContentX + mainCursor
				if isReverse {
					childX = box.ContentX + mainSize - mainCursor - item.flexW
				}
				child.ContentX = childX
				child.ContentY = box.ContentY + crossCursor
				child.ContentW = item.flexW
				child.ContentH = line.crossSize

				// Align self
				align := child.Style["align-self"]
				if align == "" {
					align = alignItems
				}
				switch align {
				case "center":
					child.ContentY = box.ContentY + crossCursor + (line.crossSize-child.ContentH)/2
				case "flex-end":
					child.ContentY = box.ContentY + crossCursor + line.crossSize - child.ContentH
				case "stretch":
					if child.ContentH < line.crossSize {
						child.ContentH = line.crossSize
					}
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
				childY := box.ContentY + crossCursor + mainCursor
				if isReverse {
					childY = box.ContentY + crossCursor + mainSize - mainCursor - item.flexW
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
				case "stretch":
					if child.ContentW < width {
						child.ContentW = width
					}
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
		}

		// Advance crossCursor for next line (for space-between/space-around)
		if alignContent == "space-between" || alignContent == "space-around" {
			// For space-between/space-around, lines are evenly distributed
			// The crossCursor was set for the initial offset, now advance by line size + gap
			if isWrapReverse {
				crossCursor -= gap + line.crossSize
			} else {
				crossCursor += gap + line.crossSize
			}
		} else if isWrapReverse {
			crossCursor -= gap + line.crossSize
		} else {
			crossCursor += gap + line.crossSize
		}
		_ = li // silence unused variable
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

// layoutPositionedChild handles position:absolute/relative/fixed/sticky layout.
func layoutPositionedChild(box *Box, ctx *LayoutContext) {
	position := box.Style["position"]

	// visibility:collapse on table rows/cells removes them from the layout
	if visibility := box.Style["visibility"]; visibility == "collapse" {
		// For table rows and cells, collapse removes the space
		// For other elements, treat as hidden
		tag := box.Node.TagName
		if tag == "tr" || tag == "thead" || tag == "tbody" || tag == "tfoot" || tag == "td" || tag == "th" {
			// Collapsed table elements take no space
			box.ContentW = 0
			box.ContentH = 0
			return
		}
	}

	if position == "fixed" {
		// Fixed is relative to viewport
		ctx = &LayoutContext{
			Width: 800,
			Height: 600,
			X:     0,
			Y:     0,
		}
	}

	// position:sticky acts like relative, but is positioned based on scroll offset
	// For simplicity in batch rendering, treat sticky as relative with special flags
	if position == "sticky" {
		// Store sticky offset values in style for render-time adjustment
		box.Style["_stickyTop"] = box.Style["top"]
		box.Style["_stickyLeft"] = box.Style["left"]
		box.Style["_stickyRight"] = box.Style["right"]
		box.Style["_stickyBottom"] = box.Style["bottom"]
		// Treat as relative for layout purposes
		position = "relative"
	}

	top := css.ParseLength(box.Style["top"])
	left := css.ParseLength(box.Style["left"])
	right := css.ParseLength(box.Style["right"])
	bottom := css.ParseLength(box.Style["bottom"])

	width := computeWidth(box, ctx.Width)

	if position == "relative" {
		// Relative: offset from normal position
		box.ContentX = ctx.X + css.ParseLength(box.Style["margin-left"]).Value
		box.ContentY = ctx.Y + css.ParseLength(box.Style["margin-top"]).Value
		if !left.IsAuto {
			box.ContentX = ctx.X + left.Value
		} else if !right.IsAuto {
			// right offset with auto left: position from right edge
			box.ContentX = ctx.X + ctx.Width - width - right.Value
		}
		if !top.IsAuto {
			box.ContentY = ctx.Y + top.Value
		} else if !bottom.IsAuto {
			// bottom offset with auto top: position from bottom edge
			box.ContentY = ctx.Y + ctx.Height - box.ContentH - bottom.Value
		}
	} else {
		// Absolute: removed from flow, positioned relative to containing block
		box.ContentX = ctx.X
		box.ContentY = ctx.Y
		if !left.IsAuto {
			box.ContentX = ctx.X + left.Value
		} else if !right.IsAuto {
			// right offset with auto left: position from right edge
			box.ContentX = ctx.X + ctx.Width - width - right.Value
		}
		if !top.IsAuto {
			box.ContentY = ctx.Y + top.Value
		} else if !bottom.IsAuto {
			// bottom offset with auto top: position from bottom edge
			box.ContentY = ctx.Y + ctx.Height - box.ContentH - bottom.Value
		}
	}

	box.ContentW = width
	box.ContentH = computeHeight(box, ctx)

	// For absolute with right/bottom but auto left/right: fit width between left and right
	if position == "absolute" && !right.IsAuto && left.IsAuto {
		box.ContentW = ctx.Width - (box.ContentX - ctx.X) - right.Value
		if box.ContentW < 0 {
			box.ContentW = 0
		}
	}

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

	// Handle calc()
	if strings.HasPrefix(widthStr, "calc(") {
		val, err := css.ParseCalc(widthStr, func(unit string) float64 {
			return containingWidth
		})
		if err == nil {
			return val
		}
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
		// Handle calc()
		if strings.HasPrefix(hStr, "calc(") {
			val, err := css.ParseCalc(hStr, func(unit string) float64 {
				if childCtx != nil {
					return childCtx.Height
				}
				return 0
			})
			if err == nil {
				return val
			}
		}
		l := css.ParseLength(hStr)
		if l.Unit == css.UnitPercent {
			if childCtx != nil {
				return childCtx.Height * l.Value / 100
			}
			return l.Value
		}
		return l.Value
	}

	// Auto height: find the bottommost edge among all children
	if childCtx != nil {
		maxBottom := childCtx.Y
		for _, child := range box.Children {
			bottom := child.ContentY + child.ContentH
			if bottom > maxBottom {
				maxBottom = bottom
			}
		}
		return maxBottom - box.ContentY
	}
	return 0
}

// layoutTableContainer handles layout for a table box.
// It manages the overall table dimensions and hosts table sections/rows.
func layoutTableContainer(box *Box, ctx *LayoutContext) {
	width := computeWidth(box, ctx.Width)
	marginTop := css.ParseLength(box.Style["margin-top"]).Value
	marginLeft := css.ParseLength(box.Style["margin-left"]).Value
	marginBottom := css.ParseLength(box.Style["margin-bottom"]).Value

	box.ContentX = ctx.X + marginLeft
	box.ContentY = ctx.Y + marginTop
	box.ContentW = width

	captionSide := box.Style["caption-side"]
	if captionSide == "" {
		captionSide = "top"
	}

	// Find caption and table rows/sections
	var caption *Box
	var tableRows []*Box
	for _, child := range box.Children {
		if child.Type == TableCaptionBox {
			caption = child
		} else if child.Type == TableRowBox {
			tableRows = append(tableRows, child)
		} else if child.Type == TableSectionBox {
			// For section boxes, collect their row children
			for _, secChild := range child.Children {
				if secChild.Type == TableRowBox {
					tableRows = append(tableRows, secChild)
				}
			}
		}
	}

	// Layout caption if present
	captionHeight := 0.0
	if caption != nil {
		captionCtx := &LayoutContext{
			Width: width,
			X:     box.ContentX,
			Y:     box.ContentY,
		}
		if captionSide == "top" {
			layoutCaption(caption, captionCtx, "top")
			captionHeight = caption.ContentH
		}
	}

	// Layout each row and compute total table height
	totalHeight := 0.0
	for _, row := range tableRows {
		layoutTableRow(row, &LayoutContext{
			Width: width,
			X:     box.ContentX,
			Y:     box.ContentY + captionHeight + totalHeight,
		})
		totalHeight += row.ContentH
	}

	// Layout caption on bottom if caption-side is bottom
	if caption != nil && captionSide == "bottom" {
		captionCtx := &LayoutContext{
			Width: width,
			X:     box.ContentX,
			Y:     box.ContentY + captionHeight + totalHeight,
		}
		layoutCaption(caption, captionCtx, "bottom")
		captionHeight += caption.ContentH
	}

	box.ContentH = captionHeight + totalHeight

	// Update cursor
	nextY := box.ContentY + box.ContentH + marginBottom
	ctx.Y = nextY
}

// layoutCaption handles layout for a table caption.
func layoutCaption(box *Box, ctx *LayoutContext, side string) {
	width := computeWidth(box, ctx.Width)
	marginTop := css.ParseLength(box.Style["margin-top"]).Value
	marginBottom := css.ParseLength(box.Style["margin-bottom"]).Value

	box.ContentX = ctx.X
	box.ContentY = ctx.Y + marginTop
	box.ContentW = width

	// Layout children within caption
	childCtx := &LayoutContext{
		Width: width,
		X:     box.ContentX,
		Y:     box.ContentY,
	}
	layoutChildren(box, childCtx)
	box.ContentH = computeHeight(box, childCtx) + marginTop + marginBottom
}

// layoutTableRow handles layout for a table-row box.
// It positions its cells horizontally.
func layoutTableRow(box *Box, ctx *LayoutContext) {
	width := ctx.Width

	// Collect cells from this row's children
	var cells []*Box
	for _, child := range box.Children {
		if child.Type == TableCellBox {
			cells = append(cells, child)
		}
	}

	// Calculate column widths based on cells
	// Simple approach: divide available width among columns
	numCols := 0
	for _, cell := range cells {
		numCols += cell.ColSpan
	}
	if numCols == 0 {
		numCols = 1
	}
	colWidth := width / float64(numCols)

	// Position cells horizontally
	x := ctx.X
	maxHeight := 0.0
	colIndex := 0

	for _, cell := range cells {
		cellW := colWidth * float64(cell.ColSpan)
		cellH := computeHeight(cell, nil)
		if cellH > maxHeight {
			maxHeight = cellH
		}

		cell.ContentX = x
		cell.ContentY = ctx.Y
		cell.ContentW = cellW
		cell.ColumnIndex = colIndex

		// Layout cell contents
		cellCtx := &LayoutContext{
			Width: cellW,
			X:     x,
			Y:     ctx.Y,
		}
		layoutChildren(cell, cellCtx)
		cell.ContentH = computeHeight(cell, cellCtx)
		if cell.ContentH > maxHeight {
			maxHeight = cell.ContentH
		}

		x += cellW
		colIndex += cell.ColSpan
	}

	box.ContentX = ctx.X
	box.ContentY = ctx.Y
	box.ContentW = width
	box.ContentH = maxHeight

	// Update cursor
	ctx.Y += maxHeight
}

// layoutTableCell handles layout for a table-cell box.
// It's positioned by its parent row.
func layoutTableCell(box *Box, ctx *LayoutContext) {
	width := computeWidth(box, ctx.Width)
	height := computeHeight(box, nil)

	box.ContentX = ctx.X
	box.ContentY = ctx.Y
	box.ContentW = width
	box.ContentH = height

	// Check if cell is empty (no visible content)
	// A cell is considered empty if it has no text or inline content children
	isEmpty := true
	for _, child := range box.Children {
		if child.Type == TextBox && strings.TrimSpace(child.Node.Data) != "" {
			isEmpty = false
			break
		}
		if child.Type != TextBox {
			isEmpty = false
			break
		}
	}
	// Store empty state for rendering (empty-cells property affects rendering)
	box.Style["_empty"] = strconv.FormatBool(isEmpty)

	// Layout children
	childCtx := &LayoutContext{
		Width: width,
		X:     ctx.X,
		Y:     ctx.Y,
	}
	layoutChildren(box, childCtx)
	box.ContentH = computeHeight(box, childCtx)
}

// layoutMultiColumn handles CSS multi-column layout (column-count, column-width, column-gap).
func layoutMultiColumn(box *Box, ctx *LayoutContext, columnCount int, columnWidth float64, columnGap float64) {
	width := computeWidth(box, ctx.Width)
	marginTop := css.ParseLength(box.Style["margin-top"]).Value
	marginBottom := css.ParseLength(box.Style["margin-bottom"]).Value

	box.ContentX = ctx.X + css.ParseLength(box.Style["margin-left"]).Value
	box.ContentY = ctx.Y + marginTop
	box.ContentW = width

	// Calculate actual column count and width
	// If column-width is specified and column-count is auto, calculate column count
	// If column-count is specified and column-width is auto, calculate column width
	// If both are specified, use whichever creates fewer columns
	actualCount := columnCount
	actualWidth := columnWidth

	if columnWidth > 0 && columnCount == 1 {
		// column-width is specified, calculate column count
		// formula: n = max(1, floor((availableWidth + columnGap) / (columnWidth + columnGap)))
		if columnGap > 0 {
			actualCount = int((width + columnGap) / (columnWidth + columnGap))
		} else {
			actualCount = int(width / columnWidth)
		}
		if actualCount < 1 {
			actualCount = 1
		}
		actualWidth = columnWidth
	} else if columnCount > 1 && columnWidth == 0 {
		// column-count is specified, calculate column width
		// total gap space = columnGap * (actualCount - 1)
		// columnWidth = (width - totalGapSpace) / actualCount
		gapSpace := columnGap * float64(columnCount-1)
		actualWidth = (width - gapSpace) / float64(columnCount)
		actualCount = columnCount
	} else if columnWidth > 0 && columnCount > 1 {
		// Both specified - use whichever creates fewer columns
		// Calculate columns from width
		countFromWidth := int((width + columnGap) / (columnWidth + columnGap))
		if countFromWidth < 1 {
			countFromWidth = 1
		}
		// Use min of specified count and calculated count
		if countFromWidth < columnCount {
			actualCount = countFromWidth
			actualWidth = columnWidth
		} else {
			actualCount = columnCount
			gapSpace := columnGap * float64(columnCount-1)
			actualWidth = (width - gapSpace) / float64(columnCount)
		}
	}

	if actualCount < 1 {
		actualCount = 1
	}
	if actualWidth < 10 {
		actualWidth = width / float64(actualCount)
	}

	// Calculate column gap
	totalGapSpace := columnGap * float64(actualCount-1)
	if totalGapSpace < 0 {
		totalGapSpace = 0
	}
	actualColWidth := (width - totalGapSpace) / float64(actualCount)

	// Create column boxes and distribute children
	// For simplicity, we create virtual column boxes that hold the distributed content
	var columns []*Box
	for i := 0; i < actualCount; i++ {
		col := &Box{
			Type:   ColumnBox,
			Node:   box.Node,
			Parent: box,
			Style:  make(map[string]string),
			Children: []*Box{},
		}
		col.ContentX = box.ContentX + float64(i)*(actualColWidth+columnGap)
		col.ContentY = box.ContentY
		col.ContentW = actualColWidth
		col.ContentH = 0
		columns = append(columns, col)
	}

	// Distribute children across columns
	// Simple distribution: round-robin based on content type
	// Block-level children go into columns sequentially
	colIndex := 0
	for _, child := range box.Children {
		if columns[colIndex].ContentH > 0 {
			// Add some height for the separator between block children in same column
			// Actually, we just append and layout - spacing handled in child layout
		}
		child.Parent = columns[colIndex]
		columns[colIndex].Children = append(columns[colIndex].Children, child)
		colIndex = (colIndex + 1) % actualCount
	}

	// Layout content in each column
	maxColHeight := 0.0
	for i, col := range columns {
		childCtx := &LayoutContext{
			Width: actualColWidth,
			X:     col.ContentX,
			Y:     col.ContentY,
		}
		// Layout children of the column
		for _, child := range col.Children {
			layoutChild(child, col, childCtx)
		}
		col.ContentH = childCtx.Y - col.ContentY
		if col.ContentH > maxColHeight {
			maxColHeight = col.ContentH
		}
		// Update column position based on max height found so far
		columns[i].ContentH = col.ContentH
	}

	// Set box dimensions
	box.ContentH = maxColHeight

	// Note: columns are not stored in box.Children to avoid double-layout
	// The column information is stored for rendering in a special style property
	box.Style["_columnCount"] = strconv.Itoa(actualCount)
	box.Style["_columnWidth"] = strconv.FormatFloat(actualColWidth, 'f', 2, 64)
	box.Style["_columnGap"] = strconv.FormatFloat(columnGap, 'f', 2, 64)

	// Advance cursor
	nextY := box.ContentY + box.ContentH + marginBottom
	ctx.Y = nextY
}
