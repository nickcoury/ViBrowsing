package render

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/fetch"
	"github.com/nickcoury/ViBrowsing/internal/html"
	"github.com/nickcoury/ViBrowsing/internal/layout"
)

// Canvas represents a pixel buffer for rendering.
type Canvas struct {
	Width  int
	Height int
	Pixels *image.RGBA

	// clipStack tracks clipping rectangles for overflow:hidden/scroll/auto
	clipStack []struct{ X, Y, W, H int }

	// scrollOffset is the current scroll position (in pixels)
	// This is applied to all content when rendering
	ScrollX int
	ScrollY int

	// pageHeight is the total height of the laid-out page content
	// Used to calculate scroll limits
	PageHeight int
}

// NewCanvas creates a new canvas with the given dimensions.
func NewCanvas(width, height int) *Canvas {
	return &Canvas{
		Width:  width,
		Height: height,
		Pixels: image.NewRGBA(image.Rect(0, 0, width, height)),
	}
}

// Clear fills the canvas with white.
func (c *Canvas) Clear() {
	white := color.RGBA{255, 255, 255, 255}
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			c.Pixels.Set(x, y, white)
		}
	}
}

// SetPixel sets a single pixel.
func (c *Canvas) SetPixel(x, y int, col color.Color) {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return
	}
	c.Pixels.Set(x, y, col)
}

// PushClip restricts drawing to the given rectangle (intersected with current clip).
func (c *Canvas) PushClip(x, y, w, h int) {
	if len(c.clipStack) > 0 {
		// Intersect with current clip
		cur := c.clipStack[len(c.clipStack)-1]
		x2 := x + w
		y2 := y + h
		cx2 := cur.X + cur.W
		cy2 := cur.Y + cur.H
		if x < cur.X {
			x = cur.X
		}
		if y < cur.Y {
			y = cur.Y
		}
		if x2 > cx2 {
			x2 = cx2
		}
		if y2 > cy2 {
			y2 = cy2
		}
		x = x
		y = y
		w = x2 - x
		h = y2 - y
	}
	c.clipStack = append(c.clipStack, struct{ X, Y, W, H int }{x, y, w, h})
}

// PopClip restores the previous clipping rectangle.
func (c *Canvas) PopClip() {
	if len(c.clipStack) > 0 {
		c.clipStack = c.clipStack[:len(c.clipStack)-1]
	}
}

// FillRect fills a rectangle with a color (clipped to current clip rect).
func (c *Canvas) FillRect(x, y, w, h int, col color.Color) {
	if len(c.clipStack) == 0 {
		// Fast path: no clipping
		for dy := 0; dy < h; dy++ {
			for dx := 0; dx < w; dx++ {
				c.Pixels.Set(x+dx, y+dy, col)
			}
		}
		return
	}
	// Clipped path
	clip := c.clipStack[len(c.clipStack)-1]
	cx2 := clip.X + clip.W
	cy2 := clip.Y + clip.H
	x2 := x + w
	y2 := y + h
	// Compute intersection with clip rect
	if x < clip.X {
		x = clip.X
	}
	if y < clip.Y {
		y = clip.Y
	}
	if x2 > cx2 {
		x2 = cx2
	}
	if y2 > cy2 {
		y2 = cy2
	}
	if x >= x2 || y >= y2 {
		return
	}
	for dy := y; dy < y2; dy++ {
		for dx := x; dx < x2; dx++ {
			c.Pixels.Set(dx, dy, col)
		}
	}
}

// applyOpacity blends a CSS color with an opacity factor (0-1).
func applyOpacity(col css.Color, opacity float64) css.Color {
	if opacity <= 0 {
		return css.Color{R: 0, G: 0, B: 0, A: 0}
	}
	if opacity >= 1 {
		return col
	}
	// col is already a css.Color, which stores 8-bit values
	r := uint8(float64(col.R) * opacity)
	g := uint8(float64(col.G) * opacity)
	b := uint8(float64(col.B) * opacity)
	a := uint8(float64(col.A) * opacity)
	return css.Color{R: r, G: g, B: b, A: a}
}

// DrawBox renders a layout box and its children.
func (c *Canvas) DrawBox(box *layout.Box) {
	if box == nil {
		return
	}

	// Skip display:none boxes entirely (no space occupied)
	if display, ok := box.Style["display"]; ok && display == "none" {
		return
	}

	// Apply scroll offset to content position
	contentX := int(box.ContentX) - c.ScrollX
	contentY := int(box.ContentY) - c.ScrollY
	contentW := int(box.ContentW)
	contentH := int(box.ContentH)

	if contentW <= 0 || contentH <= 0 {
		contentW = 800
		contentH = 600
	}

	// Opacity
	opacity := 1.0
	if op, ok := box.Style["opacity"]; ok {
		if v, err := strconv.ParseFloat(op, 64); err == nil {
			opacity = v
			if v < 0 {
				opacity = 0
			}
			if v > 1 {
				opacity = 1
			}
		}
	}

	// Background
	bgColor := css.ParseColor(box.Style["background"])
	if opacity < 1 {
		bgColor = applyOpacity(bgColor, opacity)
	}

	// Background (margin box area) — use rounded rect if border-radius set
	marginLeft := int(css.ParseLength(box.Style["margin-left"]).Value)
	marginTop := int(css.ParseLength(box.Style["margin-top"]).Value)
	marginRight := int(css.ParseLength(box.Style["margin-right"]).Value)
	marginBottom := int(css.ParseLength(box.Style["margin-bottom"]).Value)
	marginX := contentX - marginLeft
	marginY := contentY - marginTop
	marginW := marginLeft + contentW + marginRight
	marginH := marginTop + contentH + marginBottom

	borderRadius := css.ParseBorderRadius(box.Style["border-radius"])
	borderRadiusPx := borderRadius.TopLeft

	// Parse padding once (used in both branches and later for overflow clip)
	paddingTop := int(css.ParseLength(box.Style["padding-top"]).Value)
	paddingRight := int(css.ParseLength(box.Style["padding-right"]).Value)
	paddingBottom := int(css.ParseLength(box.Style["padding-bottom"]).Value)
	paddingLeft := int(css.ParseLength(box.Style["padding-left"]).Value)
	paddingX := contentX + paddingLeft
	paddingY := contentY + paddingTop
	paddingW := contentW - paddingLeft - paddingRight
	paddingH := contentH - paddingTop - paddingBottom

	// Background image drawing
	bgImage := box.Style["background-image"]
	if bgImage != "" && bgImage != "none" {
		if strings.HasPrefix(bgImage, "url(") {
			// Extract URL from url(...)
			start := strings.Index(bgImage, "(")
			end := strings.LastIndex(bgImage, ")")
			if start > 0 && end > start {
				url := bgImage[start+1 : end]
				// Remove quotes if present
				url = strings.Trim(url, "'\"")
				img := c.loadBackgroundImage(url)
				if img != nil && paddingW > 0 && paddingH > 0 {
					// Get background-repeat
					bgRepeat := box.Style["background-repeat"]
					if bgRepeat == "" {
						bgRepeat = "repeat"
					}
					// Get background-position
					bgPos := box.Style["background-position"]
					// Get background-size
					bgSize := box.Style["background-size"]
					c.drawBackgroundImage(img, paddingX, paddingY, paddingW, paddingH, bgRepeat, bgPos, bgSize, opacity)
				}
			}
		} else if strings.HasPrefix(bgImage, "linear-gradient(") {
			// Draw linear gradient background
			stops, direction, err := css.ParseLinearGradient(bgImage)
			if err == nil && len(stops) >= 2 && paddingW > 0 && paddingH > 0 {
				c.drawLinearGradient(paddingX, paddingY, paddingW, paddingH, stops, direction, opacity)
			}
		} else if strings.HasPrefix(bgImage, "radial-gradient(") {
			// Draw radial gradient background
			stops, shape, position, err := css.ParseRadialGradient(bgImage)
			if err == nil && len(stops) >= 2 && paddingW > 0 && paddingH > 0 {
				c.drawRadialGradient(paddingX, paddingY, paddingW, paddingH, stops, shape, position, opacity)
			}
		}
	}

	// Use rounded rect if border-radius is set
	if borderRadiusPx.Value > 0 && borderRadiusPx.Unit == css.UnitPx {
		cornerR := int(borderRadiusPx.Value)
		c.DrawRoundedRect(marginX, marginY, marginW, marginH, cornerR, bgColor)
		// Padding area (smaller rounded rect inside)
		// Reduce corner radius for inner layers (approximation)
		innerR := cornerR / 2
		if paddingW > 0 && paddingH > 0 && innerR > 0 {
			c.DrawRoundedRect(paddingX, paddingY, paddingW, paddingH, innerR, bgColor)
		}
	} else {
		c.FillRect(marginX, marginY, marginW, marginH, bgColor)

		if paddingW > 0 && paddingH > 0 {
			c.FillRect(paddingX, paddingY, paddingW, paddingH, bgColor)
		}
	}

	// Border (drawn last so it's on top of padding/background)
	borderWidth := int(css.ParseLength(box.Style["border-width"]).Value)
	borderColor := css.ParseColor(box.Style["border-color"])
	borderStyle := box.Style["border-style"]
	if borderStyle == "none" || borderStyle == "" || borderWidth == 0 {
		borderColor = css.Color{R: 0, G: 0, B: 0, A: 0}
	}
	if borderWidth > 0 {
		if opacity < 1 {
			borderColor = applyOpacity(borderColor, opacity)
		}
		c.DrawBorder(contentX, contentY, contentW, contentH, borderWidth, borderColor)
	}

	// Box shadow (drawn after background, before border)
	boxShadow := box.Style["box-shadow"]
	if boxShadow != "" && boxShadow != "none" {
		// Parse box-shadow: offsetX offsetY blurRadius color
		// Format: <offset-x> <offset-y> <blur-radius> <color>
		parts := strings.Fields(boxShadow)
		if len(parts) >= 2 {
			shadowX := int(css.ParseLength(parts[0]).Value)
			shadowY := int(css.ParseLength(parts[1]).Value)
			shadowColor := css.Color{R: 0, G: 0, B: 0, A: 128}
			if len(parts) >= 4 {
				shadowColor = css.ParseColor(parts[3])
			} else if len(parts) >= 3 {
				shadowColor = css.ParseColor(parts[2])
			}
			if opacity < 1 {
				shadowColor = applyOpacity(shadowColor, opacity)
			}
			// Draw shadow rectangle offset from content area
			shadowDrawX := contentX + shadowX
			shadowDrawY := contentY + shadowY
			c.FillRect(shadowDrawX, shadowDrawY, contentW, contentH, shadowColor)
		}
	}

	// Outline (drawn outside border box, like border but doesn't affect layout)
	outlineWidth := int(css.ParseLength(box.Style["outline-width"]).Value)
	outlineColor := css.ParseColor(box.Style["outline-color"])
	outlineStyle := box.Style["outline-style"]
	if outlineStyle == "none" || outlineStyle == "" || outlineWidth == 0 {
		outlineColor = css.Color{R: 0, G: 0, B: 0, A: 0}
	}
	if outlineWidth > 0 {
		if opacity < 1 {
			outlineColor = applyOpacity(outlineColor, opacity)
		}
		// Outline is drawn around the outer edge of the border area
		ox := contentX - outlineWidth
		oy := contentY - outlineWidth
		ow := contentW + outlineWidth*2
		oh := contentH + outlineWidth*2
		c.DrawBorder(ox, oy, ow, oh, outlineWidth, outlineColor)
	}

	// Handle overflow clipping (clip to content box for hidden/scroll/auto)
	overflow := box.Style["overflow"]
	overflowX := box.Style["overflow-x"]
	overflowY := box.Style["overflow-y"]
	clipContent := overflow == "hidden" || overflow == "scroll" || overflow == "auto" ||
		overflowX == "hidden" || overflowX == "scroll" || overflowX == "auto" ||
		overflowY == "hidden" || overflowY == "scroll" || overflowY == "auto"

	if clipContent {
		// Clip to the content box area (inside padding)
		clipX := contentX + paddingLeft
		clipY := contentY + paddingTop
		clipW := contentW - paddingLeft - paddingRight
		clipH := contentH - paddingTop - paddingBottom
		if clipW <= 0 {
			clipW = contentW
		}
		if clipH <= 0 {
			clipH = contentH
		}
		c.PushClip(clipX, clipY, clipW, clipH)
		defer c.PopClip()
	}

	// Check visibility - hidden boxes paint background/border/padding but not content
	visibility, _ := box.Style["visibility"]
	isHidden := visibility == "hidden"

	// Text content (only if not hidden)
	if !isHidden && box.Type == layout.TextBox && box.Node != nil {
		c.DrawText(box)
	}

	// Horizontal rule
	if !isHidden && box.Type == layout.HorizontalRuleBox {
		c.DrawHorizontalRule(box)
	}

	// Image content
	if !isHidden && box.Type == layout.ImageBox && box.Node != nil {
		c.DrawImage(box)
	}

	// Input element
	if !isHidden && box.Type == layout.InlineBox && box.Node != nil && box.Node.TagName == "input" {
		c.DrawInput(box)
	}

	// Button element
	if !isHidden && box.Type == layout.ButtonBox && box.Node != nil {
		c.DrawButton(box)
	}

	// Select element
	if !isHidden && box.Type == layout.SelectBox && box.Node != nil {
		c.DrawSelect(box)
	}

	// Textarea element
	if !isHidden && box.Type == layout.TextAreaBox && box.Node != nil {
		c.DrawTextArea(box)
	}

	// Label element
	if !isHidden && box.Type == layout.LabelBox && box.Node != nil {
		c.DrawLabel(box)
	}

	// Video/Audio element
	if !isHidden && box.Type == layout.MediaBox && box.Node != nil {
		c.DrawMedia(box)
	}

	// Progress element
	if !isHidden && box.Type == layout.ProgressBox && box.Node != nil {
		c.DrawProgress(box)
	}

	// Meter element
	if !isHidden && box.Type == layout.MeterBox && box.Node != nil {
		c.DrawMeter(box)
	}

	// Dialog element
	if !isHidden && box.Type == layout.DialogBox && box.Node != nil {
		c.DrawDialog(box)
	}

	// Children (only if not hidden)
	if !isHidden {
		// Sort children by z-index for proper stacking order
		children := make([]*layout.Box, len(box.Children))
		copy(children, box.Children)
		c.drawChildrenSorted(children)
	}

	// Draw list marker if this is a list item
	if box.Type == layout.ListItemBox {
		c.DrawListMarker(box)
	}
}

// drawChildrenSorted draws children in z-index order, then paint order.
func (c *Canvas) drawChildrenSorted(children []*layout.Box) {
	// Separate positioned and non-positioned children
	type zChild struct {
		box  *layout.Box
		z    int
		pos  bool
	}
	var normal, positioned []zChild

	for _, child := range children {
		z, _ := strconv.Atoi(child.Style["z-index"])
		pos := child.Style["position"] != "" && child.Style["position"] != "static"
		if pos {
			positioned = append(positioned, zChild{child, z, true})
		} else {
			normal = append(normal, zChild{child, z, false})
		}
	}

	// Sort each group by z-index
	var sortZIndex []zChild
	sortZIndex = append(sortZIndex, normal...)
	sortZIndex = append(sortZIndex, positioned...)
	// Simple bubble sort (children list is usually small)
	for i := 0; i < len(sortZIndex)-1; i++ {
		for j := i + 1; j < len(sortZIndex); j++ {
			if sortZIndex[j].z < sortZIndex[i].z {
				sortZIndex[i], sortZIndex[j] = sortZIndex[j], sortZIndex[i]
			}
		}
	}

	// Draw in z-index order
	for _, zc := range sortZIndex {
		c.DrawBox(zc.box)
	}
}

// DrawBorder draws a border around a rectangle with rounded corners.
func (c *Canvas) DrawBorder(x, y, w, h, bw int, col color.Color) {
	// Draw corners with arcs, straight edges with rects
	// For simplicity when bw is small relative to box, use basic rects
	// For larger bw with border-radius, draw properly

	// Draw four corners as quarter arcs
	// Top-left corner
	c.drawCornerArc(x+bw/2, y+bw/2, bw/2, col, 3)
	// Top-right corner
	c.drawCornerArc(x+w-bw/2, y+bw/2, bw/2, col, 2)
	// Bottom-right corner
	c.drawCornerArc(x+w-bw/2, y+h-bw/2, bw/2, col, 1)
	// Bottom-left corner
	c.drawCornerArc(x+bw/2, y+h-bw/2, bw/2, col, 4)

	// Draw four edges (each edge is a rect that doesn't overlap corners)
	// Top edge (between corners)
	if w > bw*2 {
		c.FillRect(x+bw/2, y, w-bw*2, bw/2, col)
		c.FillRect(x+bw/2, y+bw/2, w-bw*2, bw/2, col)
	}
	// Bottom edge
	if w > bw*2 {
		c.FillRect(x+bw/2, y+h-bw/2, w-bw*2, bw/2, col)
		c.FillRect(x+bw/2, y+h-bw, w-bw*2, bw/2, col)
	}
	// Left edge (between corners)
	if h > bw*2 {
		c.FillRect(x, y+bw/2, bw/2, h-bw*2, col)
		c.FillRect(x+bw/2, y+bw/2, bw/2, h-bw*2, col)
	}
	// Right edge (between corners)
	if h > bw*2 {
		c.FillRect(x+w-bw, y+bw/2, bw/2, h-bw*2, col)
		c.FillRect(x+w-bw/2, y+bw/2, bw/2, h-bw*2, col)
	}
}

// drawCornerArc draws a quarter circle arc at corner (cx, cy) with radius r.
// quadrant: 1=bottom-right, 2=top-right, 3=top-left, 4=bottom-left
func (c *Canvas) drawCornerArc(cx, cy, r int, col color.Color, quadrant int) {
	if r <= 0 {
		return
	}
	// Draw filled arc by iterating pixels
	// Using midpoint circle algorithm for each quadrant
	r2 := r * r
	for py := 0; py <= r; py++ {
		for px := 0; px <= r; px++ {
			if px*px+py*py <= r2 {
				var dx, dy int
				switch quadrant {
				case 1: // bottom-right
					dx, dy = px, py
				case 2: // top-right
					dx, dy = px, -py
				case 3: // top-left
					dx, dy = -px, -py
				case 4: // bottom-left
					dx, dy = -px, py
				}
				c.SetPixel(cx+dx, cy+dy, col)
			}
		}
	}
}

// DrawRoundedRect draws a filled rectangle with rounded corners.
func (c *Canvas) DrawRoundedRect(x, y, w, h int, radius int, col color.Color) {
	if w <= 0 || h <= 0 || radius <= 0 {
		return
	}

	// Clamp radius to not exceed dimensions
	if radius > w/2 {
		radius = w / 2
	}
	if radius > h/2 {
		radius = h / 2
	}

	// Draw center rectangle (without corners)
	centerX := x + radius
	centerW := w - radius*2
	if centerW > 0 && h > 0 {
		c.FillRect(centerX, y, centerW, radius, col)
		c.FillRect(centerX, y+h-radius, centerW, radius, col)
		c.FillRect(centerX, y+radius, centerW, h-radius*2, col)
	}
	// Draw left rectangle
	if radius > 0 && h > radius*2 {
		c.FillRect(x, y+radius, radius, h-radius*2, col)
	}
	// Draw right rectangle
	if radius > 0 && h > radius*2 {
		c.FillRect(x+w-radius, y+radius, radius, h-radius*2, col)
	}

	// Draw four corner arcs
	r := radius
	// Top-left
	c.drawFilledCorner(x+r, y+r, r, col, 3)
	// Top-right
	c.drawFilledCorner(x+w-r, y+r, r, col, 2)
	// Bottom-right
	c.drawFilledCorner(x+w-r, y+h-r, r, col, 1)
	// Bottom-left
	c.drawFilledCorner(x+r, y+h-r, r, col, 4)
}

// drawFilledCorner draws a filled quarter circle at corner (cx, cy).
// quadrant: 1=bottom-right, 2=top-right, 3=top-left, 4=bottom-left
func (c *Canvas) drawFilledCorner(cx, cy, r int, col color.Color, quadrant int) {
	if r <= 0 {
		return
	}
	r2 := r * r
	for py := 0; py <= r; py++ {
		for px := 0; px <= r; px++ {
			if px*px+py*py <= r2 {
				var dx, dy int
				switch quadrant {
				case 1: // bottom-right
					dx, dy = px, py
				case 2: // top-right
					dx, dy = px, -py
				case 3: // top-left
					dx, dy = -px, -py
				case 4: // bottom-left
					dx, dy = -px, py
				}
				c.SetPixel(cx+dx, cy+dy, col)
			}
		}
	}
}

// DrawText renders text content (placeholder: colored rectangles).
func (c *Canvas) DrawText(box *layout.Box) {
	if box.Node == nil || box.Node.Type != 2 { // NodeText
		return
	}

	text := box.Node.Data
	fontSize := css.ParseLength(box.Style["font-size"]).Value
	if fontSize == 0 {
		fontSize = 16
	}

	// Apply text-transform
	textTransform := box.Style["text-transform"]
	switch textTransform {
	case "uppercase":
		text = strings.ToUpper(text)
	case "lowercase":
		text = strings.ToLower(text)
	case "capitalize":
		text = strings.Title(strings.ToLower(text))
	}

	textColor := css.ParseColor(box.Style["color"])
	x := int(box.ContentX)
	y := int(box.ContentY)
	containerWidth := int(box.ContentW)
	if containerWidth <= 0 {
		containerWidth = 800
	}

	// Font style (italic)
	fontStyle := box.Style["font-style"]

	// Font weight (bold chars are wider, light chars are narrower)
	fontWeight := box.Style["font-weight"]

	// Letter spacing
	letterSpacing := css.ParseLength(box.Style["letter-spacing"]).Value

	// Word spacing
	wordSpacing := css.ParseLength(box.Style["word-spacing"]).Value

	// Text indent (applied on first line only)
	textIndent := css.ParseLength(box.Style["text-indent"]).Value

	charWidth := fontSize * 0.6
	if fontWeight == "bold" || fontWeight == "700" || fontWeight == "800" || fontWeight == "900" {
		charWidth = fontSize * 0.65
	} else if fontWeight == "light" || fontWeight == "100" || fontWeight == "200" || fontWeight == "300" {
		charWidth = fontSize * 0.55
	}

	// Italic slant: make chars slightly wider
	italicSlant := 1.0
	if fontStyle == "italic" || fontStyle == "oblique" {
		italicSlant = 1.1
	}

	lineHeight := fontSize * 1.2

	currentX := float64(x) + textIndent
	firstLineStartX := x
	isFirstLine := true
	currentY := float64(y)
	isPre := box.Style["white-space"] == "pre" || box.Style["white-space"] == "pre-wrap"

	// Check for text-overflow ellipsis
	// text-overflow applies when overflow-x is hidden and text overflows
	overflow := box.Style["overflow"]
	overflowX := box.Style["overflow-x"]
	textOverflow := box.Style["text-overflow"]
	showEllipsis := (overflow == "hidden" || overflowX == "hidden") && textOverflow == "ellipsis"

	// Parse text-shadow
	shadowColor := textColor
	shadowX := 0
	shadowY := 0
	hasTextShadow := false
	if shadow := box.Style["text-shadow"]; shadow != "" && shadow != "none" {
		hasTextShadow = true
		parts := strings.Fields(shadow)
		if len(parts) >= 2 {
			shadowX = int(css.ParseLength(parts[0]).Value)
			shadowY = int(css.ParseLength(parts[1]).Value)
			if len(parts) >= 4 {
				shadowColor = css.ParseColor(parts[3])
			} else if len(parts) >= 3 {
				shadowColor = css.ParseColor(parts[2])
			}
		}
	}

	// For text-overflow ellipsis, calculate the maximum chars that fit
	maxChars := len(text)
	if showEllipsis {
		// Calculate how many chars fit in the container (accounting for textIndent on first line)
		availWidth := float64(containerWidth)
		if isFirstLine {
			availWidth -= textIndent
		}
		ellipsisWidth := charWidth * italicSlant * 2 // "…" is roughly 2 chars wide
		availWidth -= ellipsisWidth
		if availWidth < 0 {
			availWidth = 0
		}
		maxChars = int(availWidth / (charWidth * italicSlant))
		if maxChars < 0 {
			maxChars = 0
		} else if maxChars > len(text) {
			maxChars = len(text)
		}
	}

	for i := 0; i < len(text); i++ {
		// Stop if we've reached the ellipsis limit
		if showEllipsis && i >= maxChars {
			// Draw ellipsis at the end
			ellipsisCW := charWidth * italicSlant * 2 // "…" is roughly 2 chars wide
			ellipsisH := fontSize
			if italicSlant > 1 {
				ellipsisH = fontSize * italicSlant
			}
			// Draw text-shadow for ellipsis
			if hasTextShadow {
				c.FillRect(int(currentX)+shadowX, int(currentY)+shadowY, int(ellipsisCW), int(ellipsisH), shadowColor)
			}
			c.FillRect(int(currentX), int(currentY), int(ellipsisCW), int(ellipsisH), textColor)
			break
		}

		ch := rune(text[i])

		// Handle explicit newlines in pre mode
		if ch == '\n' {
			if isPre {
				currentX = float64(x)
				currentY += lineHeight
				isFirstLine = false
				currentX = float64(x)
				continue
			}
			// In normal mode, newline collapses
		}
		if ch == '\r' {
			continue
		}
		if ch == '\t' {
			ch = ' '
		}

		// Extra spacing for letter-spacing
		extraLetter := letterSpacing

		// Word spacing: add after space (but not for first space on line)
		extraWord := 0.0
		if ch == ' ' && !isPre {
			extraWord = wordSpacing
			if int(currentX) == x || int(currentX) == firstLineStartX {
				extraWord = 0
			}
		}

		cw := charWidth*italicSlant + extraLetter

		// Wrap line if needed (not in pre mode)
		canWrap := !isPre
		wrapX := float64(x)
		if isFirstLine && textIndent > 0 {
			wrapX = float64(x) + textIndent
		}
		if canWrap && currentX+cw > float64(x)+float64(containerWidth) && currentX > wrapX {
			currentX = float64(x)
			currentY += lineHeight
			isFirstLine = false
			if ch == ' ' {
				continue
			}
		}

		charH := fontSize
		if italicSlant > 1 {
			charH = fontSize * italicSlant
		}

		// Draw text-shadow first (behind the text)
		if hasTextShadow {
			c.FillRect(int(currentX)+shadowX, int(currentY)+shadowY, int(cw), int(charH), shadowColor)
		}

		c.FillRect(int(currentX), int(currentY), int(cw), int(charH), textColor)

		// Draw text-decoration (underline, overline, line-through)
		decorationLine := box.Style["text-decoration-line"]
		if decorationLine == "" {
			decorationLine = box.Style["text-decoration"]
		}
		if decorationLine != "" && decorationLine != "none" {
			decorationColor := textColor
			if dc := box.Style["text-decoration-color"]; dc != "" && dc != "currentColor" {
				decorationColor = css.ParseColor(dc)
			}

			// Underline: 1px line at baseline (below text)
			if decorationLine == "underline" || strings.Contains(decorationLine, "underline") {
				decY := int(currentY + fontSize + 1)
				decH := 1
				c.FillRect(int(currentX), decY, int(cw), decH, decorationColor)
			}
			// Overline: 1px line above text
			if decorationLine == "overline" || strings.Contains(decorationLine, "overline") {
				decY := int(currentY)
				decH := 1
				c.FillRect(int(currentX), decY, int(cw), decH, decorationColor)
			}
			// Line-through: 1px line at middle of text
			if decorationLine == "line-through" || strings.Contains(decorationLine, "line-through") {
				decY := int(currentY + fontSize/2)
				decH := 1
				c.FillRect(int(currentX), decY, int(cw), decH, decorationColor)
			}
		}

		currentX += cw + extraWord
	}
}

// DrawImage renders an img element. Shows alt text if image unavailable/loading.
func (c *Canvas) DrawImage(box *layout.Box) {
	if box.Node == nil || box.Node.Type != html.NodeElement {
		return
	}

	src := box.Node.GetAttribute("src")
	alt := box.Node.GetAttribute("alt")

	// Default dimensions
	w := int(box.ContentW)
	h := int(box.ContentH)
	if w <= 0 {
		w = 150
		box.ContentW = 150
	}
	if h <= 0 {
		h = 150
		box.ContentH = 150
	}

	x := int(box.ContentX)
	y := int(box.ContentY)

	if src == "" {
		// No src — show alt text in a placeholder box
		c.FillRect(x, y, w, h, css.Color{R: 240, G: 240, B: 240, A: 255})
		c.DrawBorder(x, y, w, h, 1, css.Color{R: 200, G: 200, B: 200, A: 255})
		// Draw alt text if available
		if alt != "" {
			// Render alt text as simple colored rects (no actual font rendering)
			fontSize := 12.0
			charWidth := fontSize * 0.6
			maxChars := w / int(charWidth)
			if maxChars > len(alt) {
				maxChars = len(alt)
			}
			textColor := css.Color{R: 100, G: 100, B: 100, A: 255}
			for i := 0; i < maxChars; i++ {
				c.FillRect(x+int(float64(i)*charWidth), y+h/2-int(fontSize/2), int(charWidth), int(fontSize), textColor)
			}
		}
		return
	}

	// Attempt to load and decode the image
	img, err := c.loadImage(src)
	if err != nil || img == nil {
		// Failed to load — show alt or placeholder
		c.FillRect(x, y, w, h, css.Color{R: 240, G: 240, B: 240, A: 255})
		c.DrawBorder(x, y, w, h, 1, css.Color{R: 200, G: 200, B: 200, A: 255})
		// Draw a broken-image icon (simple X)
		lineColor := css.Color{R: 180, G: 180, B: 180, A: 255}
		c.FillRect(x+w/4, y+h/4, w/2, 1, lineColor)
		c.FillRect(x+w/4, y+h*3/4, w/2, 1, lineColor)
		c.FillRect(x+w/4, y+h/4, 1, h/2, lineColor)
		c.FillRect(x+w*3/4, y+h/4, 1, h/2, lineColor)
		return
	}

	// Get object-fit and object-position from style
	objectFit := box.Style["object-fit"]
	if objectFit == "" {
		objectFit = "fill"
	}
	objectPosition := box.Style["object-position"]
	if objectPosition == "" {
		objectPosition = "50% 50%"
	}

	// Draw the image with object-fit and object-position
	c.drawImageFitted(img, x, y, w, h, objectFit, objectPosition)
}

// loadImage fetches and decodes an image from a URL.
// Returns the image or an error.
func (c *Canvas) loadImage(src string) (image.Image, error) {
	// TODO: implement actual image fetching and decoding
	// For now, return nil so alt text / broken image icon is shown
	return nil, nil
}

// loadBackgroundImage fetches and decodes a background image URL.
// Returns the image or nil if loading failed.
func (c *Canvas) loadBackgroundImage(url string) image.Image {
	resp, err := fetch.Fetch(url, "", 30)
	if err != nil || resp == nil || resp.StatusCode != 200 {
		return nil
	}

	// Decode based on content type or magic bytes
	contentType := resp.ContentType
	if contentType == "" {
		// Try to detect from magic bytes
		if len(resp.Body) >= 4 {
			if resp.Body[0] == 0x89 && resp.Body[1] == 0x50 && resp.Body[2] == 0x4E && resp.Body[3] == 0x47 {
				contentType = "image/png"
			} else if resp.Body[0] == 0xFF && resp.Body[1] == 0xD8 {
				contentType = "image/jpeg"
			} else if resp.Body[0] == 0x47 && resp.Body[1] == 0x49 && resp.Body[2] == 0x46 {
				contentType = "image/gif"
			} else if len(resp.Body) >= 12 && string(resp.Body[0:12]) == "RIFF" && string(resp.Body[8:12]) == "WEBP" {
				contentType = "image/webp"
			}
		}
	}

	reader := bytes.NewReader(resp.Body)

	switch contentType {
	case "image/png":
		img, err := png.Decode(reader)
		if err != nil {
			return nil
		}
		return img
	case "image/jpeg", "image/jpg":
		img, err := jpeg.Decode(reader)
		if err != nil {
			return nil
		}
		return img
	case "image/gif":
		img, err := gif.Decode(reader)
		if err != nil {
			return nil
		}
		return img
	case "image/webp":
		// WebP decode not in stdlib, return nil for now
		return nil
	default:
		// Try PNG as fallback
		reader.Seek(0, 0)
		if img, err := png.Decode(reader); err == nil {
			return img
		}
		return nil
	}
}

// drawBackgroundImage draws a background image with repeat, position, and size.
func (c *Canvas) drawBackgroundImage(img image.Image, x, y, w, h int, repeat, position, size string, opacity float64) {
	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()
	if imgW == 0 || imgH == 0 {
		return
	}

	// Parse background-size
	imgDrawW := imgW
	imgDrawH := imgH
	if size != "" && size != "auto" {
		parts := strings.Fields(size)
		if len(parts) >= 1 {
			if parts[0] == "cover" || parts[0] == "contain" {
				// Scale to fit/cover while maintaining aspect ratio
				scaleW := float64(w) / float64(imgW)
				scaleH := float64(h) / float64(imgH)
				scale := scaleW
				if parts[0] == "cover" {
					if scaleH > scale {
						scale = scaleH
					}
				} else { // contain
					if scaleH < scale {
						scale = scaleH
					}
				}
				imgDrawW = int(float64(imgW) * scale)
				imgDrawH = int(float64(imgH) * scale)
			} else {
				l := css.ParseLength(parts[0])
				if l.Unit == css.UnitPercent {
					imgDrawW = int(float64(w) * l.Value / 100)
				} else {
					imgDrawW = int(l.Value)
				}
				if len(parts) >= 2 {
					l2 := css.ParseLength(parts[1])
					if l2.Unit == css.UnitPercent {
						imgDrawH = int(float64(h) * l2.Value / 100)
					} else {
						imgDrawH = int(l2.Value)
					}
				} else {
					// Auto height - maintain aspect ratio
					imgDrawH = int(float64(imgDrawW) * float64(imgH) / float64(imgW))
				}
			}
		}
	}

	// Parse background-position
	posX := 0
	posY := 0
	if position != "" {
		parts := strings.Fields(position)
		if len(parts) >= 2 {
			lx := css.ParseLength(parts[0])
			ly := css.ParseLength(parts[1])
			if lx.Unit == css.UnitPercent {
				posX = int(float64(w-imgDrawW) * lx.Value / 100)
			} else {
				posX = int(lx.Value)
			}
			if ly.Unit == css.UnitPercent {
				posY = int(float64(h-imgDrawH) * ly.Value / 100)
			} else {
				posY = int(ly.Value)
			}
		} else if parts[0] == "center" {
			posX = (w - imgDrawW) / 2
			posY = (h - imgDrawH) / 2
		} else if parts[0] == "top" || parts[0] == "bottom" {
			posY = 0
			if parts[0] == "bottom" {
				posY = h - imgDrawH
			}
			if len(parts) >= 2 {
				lx := css.ParseLength(parts[1])
				if lx.Unit == css.UnitPercent {
					posX = int(float64(w-imgDrawW) * lx.Value / 100)
				} else {
					posX = int(lx.Value)
				}
			}
		} else if parts[0] == "left" || parts[0] == "right" {
			posX = 0
			if parts[0] == "right" {
				posX = w - imgDrawW
			}
			if len(parts) >= 2 {
				ly := css.ParseLength(parts[1])
				if ly.Unit == css.UnitPercent {
					posY = int(float64(h-imgDrawH) * ly.Value / 100)
				} else {
					posY = int(ly.Value)
				}
			}
		}
	}

	// Handle repeat
	switch repeat {
	case "no-repeat":
		// Draw single image at position
		c.drawImagePart(img, x+posX, y+posY, imgDrawW, imgDrawH, 0, 0, imgW, imgH, opacity)
	case "repeat-x":
		// Repeat horizontally
		for px := x + posX; px < x+w; px += imgDrawW {
			remaining := x + w - px
			drawW := imgDrawW
			srcX := 0
			if px < x {
				srcX = x - px
				drawW = imgDrawW - srcX
				px = x
			}
			if drawW > remaining {
				drawW = remaining
			}
			if drawW > 0 {
				c.drawImagePart(img, px, y+posY, drawW, imgDrawH, srcX, 0, srcX+drawW, imgH, opacity)
			}
		}
	case "repeat-y":
		// Repeat vertically
		for py := y + posY; py < y+h; py += imgDrawH {
			remaining := y + h - py
			drawH := imgDrawH
			srcY := 0
			if py < y {
				srcY = y - py
				drawH = imgDrawH - srcY
				py = y
			}
			if drawH > remaining {
				drawH = remaining
			}
			if drawH > 0 {
				c.drawImagePart(img, x+posX, py, imgDrawW, drawH, 0, srcY, imgW, srcY+drawH, opacity)
			}
		}
	case "space":
		// Space to fill evenly - treat as simple repeat for now
		fallthrough
	case "round":
		fallthrough
	case "repeat":
		fallthrough
	default:
		// Tile both directions
		for py := y + posY; py < y+h; py += imgDrawH {
			for px := x + posX; px < x+w; px += imgDrawW {
				// Clip to box bounds
				drawX := px
				drawY := py
				drawW := imgDrawW
				drawH := imgDrawH
				srcX := 0
				srcY := 0

				if px < x {
					srcX = x - px
					drawW = imgDrawW - srcX
					drawX = x
				}
				if py < y {
					srcY = y - py
					drawH = imgDrawH - srcY
					drawY = y
				}
				rightEdge := drawX + drawW
				bottomEdge := drawY + drawH
				if rightEdge > x+w {
					drawW = x + w - drawX
				}
				if bottomEdge > y+h {
					drawH = y + h - drawY
				}
				if drawW > 0 && drawH > 0 {
					c.drawImagePart(img, drawX, drawY, drawW, drawH, srcX, srcY, srcX+drawW, srcY+drawH, opacity)
				}
			}
		}
	}
}

// drawImagePart draws a portion of an image at specified coordinates.
func (c *Canvas) drawImagePart(img image.Image, dx, dy, dw, dh int, sx, sy, sw, sh int, opacity float64) {
	if dw <= 0 || dh <= 0 || sw <= 0 || sh <= 0 {
		return
	}
	scaleX := float64(sw) / float64(dw)
	scaleY := float64(sh) / float64(dh)

	for py := 0; py < dh; py++ {
		srcY := sy + int(float64(py)*scaleY)
		if srcY >= sh {
			srcY = sh - 1
		}
		for px := 0; px < dw; px++ {
			srcX := sx + int(float64(px)*scaleX)
			if srcX >= sw {
				srcX = sw - 1
			}
			r, g, b, a := img.At(srcX, srcY).RGBA()
			// Convert from 16-bit to 8-bit
			col := color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}
			if opacity < 1 {
				col.R = uint8(float64(col.R) * opacity)
				col.G = uint8(float64(col.G) * opacity)
				col.B = uint8(float64(col.B) * opacity)
				col.A = uint8(float64(col.A) * opacity)
			}
			c.Pixels.Set(dx+px, dy+py, col)
		}
	}
}

// drawImageScaled draws an image scaled to fit within (x, y, w, h).
func (c *Canvas) drawImageScaled(img image.Image, x, y, w, h int) {
	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()
	if imgW == 0 || imgH == 0 {
		return
	}

	// Scale to fit within the box while preserving aspect ratio
	scaleX := float64(w) / float64(imgW)
	scaleY := float64(h) / float64(imgH)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	scaledW := int(float64(imgW) * scale)
	scaledH := int(float64(imgH) * scale)

	// Center in the box
	drawX := x + (w-scaledW)/2
	drawY := y + (h-scaledH)/2

	// Nearest-neighbor scaling by pixel copying
	for py := 0; py < scaledH; py++ {
		srcY := int(float64(py) / scale)
		if srcY >= imgH {
			srcY = imgH - 1
		}
		for px := 0; px < scaledW; px++ {
			srcX := int(float64(px) / scale)
			if srcX >= imgW {
				srcX = imgW - 1
			}
			r, g, b, a := img.At(srcX, srcY).RGBA()
			// Convert from 16-bit to 8-bit
			col := color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}
			c.Pixels.Set(drawX+px, drawY+py, col)
		}
	}
}

// drawImageFitted draws an image within a bounding box using object-fit and object-position.
// object-fit: fill, contain, cover, none, scale-down
// object-position: x y (keywords, percentages, or lengths)
func (c *Canvas) drawImageFitted(img image.Image, x, y, w, h int, objectFit, objectPosition string) {
	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()
	if imgW == 0 || imgH == 0 || w <= 0 || h <= 0 {
		return
	}

	var srcX, srcY, srcW, srcH int // Source rectangle to draw from image

	switch objectFit {
	case "contain":
		// Scale to fit within box, maintaining aspect ratio, leaving empty space
		scaleX := float64(w) / float64(imgW)
		scaleY := float64(h) / float64(imgH)
		scale := scaleX
		if scaleY < scale {
			scale = scaleY
		}
		scaledW := int(float64(imgW) * scale)
		scaledH := int(float64(imgH) * scale)
		// Center in box, fill remainder with background (already drawn)
		srcX = 0
		srcY = 0
		srcW = imgW
		srcH = imgH
		drawX := x + (w-scaledW)/2
		drawY := y + (h-scaledH)/2
		c.drawImagePart(img, drawX, drawY, scaledW, scaledH, srcX, srcY, srcW, srcH, 1.0)

	case "cover":
		// Scale to cover box, maintaining aspect ratio, cropping excess
		scaleX := float64(w) / float64(imgW)
		scaleY := float64(h) / float64(imgH)
		scale := scaleX
		if scaleY > scale {
			scale = scaleY
		}
		scaledW := int(float64(imgW) * scale)
		scaledH := int(float64(imgH) * scale)
		// Apply object-position
		posX, posY := parseObjectPosition(objectPosition, w, h, scaledW, scaledH)
		// Clip to box and draw
		c.drawImagePartClipped(img, x, y, w, h, posX, posY, scaledW, scaledH)

	case "none":
		// Use intrinsic size, positioned with object-position
		srcW = imgW
		srcH = imgH
		posX, posY := parseObjectPosition(objectPosition, w, h, imgW, imgH)
		c.drawImagePartClipped(img, x, y, w, h, posX, posY, imgW, imgH)

	case "scale-down":
		// Act like none or contain, whichever results in smaller image
		if imgW > w || imgH > h {
			// Use contain behavior
			scaleX := float64(w) / float64(imgW)
			scaleY := float64(h) / float64(imgH)
			scale := scaleX
			if scaleY < scale {
				scale = scaleY
			}
			scaledW := int(float64(imgW) * scale)
			scaledH := int(float64(imgH) * scale)
			srcX = 0
			srcY = 0
			srcW = imgW
			srcH = imgH
			drawX := x + (w-scaledW)/2
			drawY := y + (h-scaledH)/2
			c.drawImagePart(img, drawX, drawY, scaledW, scaledH, srcX, srcY, srcW, srcH, 1.0)
		} else {
			// Use intrinsic size
			posX, posY := parseObjectPosition(objectPosition, w, h, imgW, imgH)
			c.drawImagePartClipped(img, x, y, w, h, posX, posY, imgW, imgH)
		}

	case "fill":
		// Default: stretch to fill box (ignore aspect ratio)
		srcX = 0
		srcY = 0
		srcW = imgW
		srcH = imgH
		c.drawImagePart(img, x, y, w, h, srcX, srcY, srcW, srcH, 1.0)

	default:
		// Default to fill
		srcX = 0
		srcY = 0
		srcW = imgW
		srcH = imgH
		c.drawImagePart(img, x, y, w, h, srcX, srcY, srcW, srcH, 1.0)
	}
}

// parseObjectPosition parses object-position value and returns offsetX, offsetY.
// object-position: <position> where <position> = [left|center|right] || [top|center|bottom]
// Also supports percentages and pixel values.
func parseObjectPosition(pos string, boxW, boxH, imgW, imgH int) (offsetX, offsetY int) {
	// Default to center
	offsetX = (boxW - imgW) / 2
	offsetY = (boxH - imgH) / 2

	if pos == "" {
		return offsetX, offsetY
	}

	parts := strings.Fields(pos)
	if len(parts) == 0 {
		return offsetX, offsetY
	}

	// Parse horizontal position (first value: left, center, right, or length/%)
	first := parts[0]
	// Parse vertical position (second value: top, center, bottom, or length/%)
	var second string
	if len(parts) >= 2 {
		second = parts[1]
	} else {
		// Single value: could be a keyword like "center" or "50%"
		// If it's a single keyword, it applies to both axes
		first = pos
	}

	// Parse first value (horizontal)
	switch strings.ToLower(first) {
	case "left":
		offsetX = 0
	case "right":
		offsetX = boxW - imgW
	case "center":
		offsetX = (boxW - imgW) / 2
	default:
		// Could be a percentage or length
		if strings.HasSuffix(first, "%") {
			if v, err := strconv.ParseFloat(strings.TrimSuffix(first, "%"), 64); err == nil {
				offsetX = int(float64(boxW-imgW) * v / 100)
			}
		} else if l := css.ParseLength(first); l.Unit == css.UnitPx {
			offsetX = int(l.Value)
		}
	}

	// Parse second value (vertical) if present
	if second != "" {
		switch strings.ToLower(second) {
		case "top":
			offsetY = 0
		case "bottom":
			offsetY = boxH - imgH
		case "center":
			offsetY = (boxH - imgH) / 2
		default:
			// Could be a percentage or length
			if strings.HasSuffix(second, "%") {
				if v, err := strconv.ParseFloat(strings.TrimSuffix(second, "%"), 64); err == nil {
					offsetY = int(float64(boxH-imgH) * v / 100)
				}
			} else if l := css.ParseLength(second); l.Unit == css.UnitPx {
				offsetY = int(l.Value)
			}
		}
	} else if len(parts) == 1 {
		// Single value is vertical if first was horizontal keyword
		if strings.ToLower(first) == "left" || strings.ToLower(first) == "right" {
			// Second value is vertical
			switch strings.ToLower(first) {
			case "top":
				offsetY = 0
			case "bottom":
				offsetY = boxH - imgH
			}
		}
	}

	return offsetX, offsetY
}

// drawImagePartClipped draws an image portion with clipping to fit the target bounds.
func (c *Canvas) drawImagePartClipped(img image.Image, x, y, w, h int, srcX, srcY, srcW, srcH int) {
	if srcW <= 0 || srcH <= 0 || w <= 0 || h <= 0 {
		return
	}

	// Compute how much of the source image extends beyond the box
	// and should be clipped
	scaleX := float64(w) / float64(srcW)
	scaleY := float64(h) / float64(srcH)

	// Draw the image, clipping to box bounds
	for py := 0; py < h; py++ {
		// Source Y in image pixels
		srcYPos := srcY + int(float64(py)/scaleY)
		if srcYPos < 0 {
			srcYPos = 0
		}
		if srcYPos >= img.Bounds().Dy() {
			srcYPos = img.Bounds().Dy() - 1
		}
		for px := 0; px < w; px++ {
			// Source X in image pixels
			srcXPos := srcX + int(float64(px)/scaleX)
			if srcXPos < 0 {
				srcXPos = 0
			}
			if srcXPos >= img.Bounds().Dx() {
				srcXPos = img.Bounds().Dx() - 1
			}
			r, g, b, a := img.At(srcXPos, srcYPos).RGBA()
			col := color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}
			c.Pixels.Set(x+px, y+py, col)
		}
	}
}

// SavePNG saves the canvas to a PNG file.
func (c *Canvas) SavePNG(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, c.Pixels)
}

// DrawListMarker renders the list marker (bullet or number) for a list item.
func (c *Canvas) DrawListMarker(box *layout.Box) {
	marker := box.Style["_marker"]
	markerWidthStr := box.Style["_markerWidth"]
	markerWidth := 20.0
	if markerWidthStr != "" {
		if w, err := strconv.ParseFloat(markerWidthStr, 64); err == nil {
			markerWidth = w
		}
	}

	x := int(box.ContentX - markerWidth)
	y := int(box.ContentY)
	h := int(box.ContentH)
	if h <= 0 {
		h = 20
	}

	textColor := css.ParseColor(box.Style["color"])

	// Draw marker based on type
	switch marker {
	case "none":
		// No marker
	case "square", "circle", "disc", "":
		// Bullet point — filled circle
		r := int(float64(h) * 0.3)
		if r < 3 {
			r = 3
		}
		c.drawFilledCorner(x+int(markerWidth)/2, y+h/2, r, textColor, 1)
	case "decimal", "ol:decimal":
		// Number — draw simple number representation
		fontSize := 14.0
		charWidth := fontSize * 0.6
		c.FillRect(x+2, y+h/4, int(charWidth*3), int(fontSize), textColor)
	case "lower-alpha", "upper-alpha":
		// Letter
		fontSize := 14.0
		charWidth := fontSize * 0.6
		c.FillRect(x+2, y+h/4, int(charWidth), int(fontSize), textColor)
	default:
		// Default bullet
		r := int(float64(h) * 0.25)
		if r < 3 {
			r = 3
		}
		c.drawFilledCorner(x+int(markerWidth)/2, y+h/2, r, textColor, 1)
	}
}

// ToImage returns the underlying *image.RGBA.
func (c *Canvas) ToImage() *image.RGBA {
	return c.Pixels
}

// DrawHorizontalRule renders an <hr> element as a thin horizontal line.
func (c *Canvas) DrawHorizontalRule(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	if w <= 0 {
		w = 200
	}

	borderColor := css.ParseColor(box.Style["border-color"])
	borderWidth := int(css.ParseLength(box.Style["border-width"]).Value)
	if borderWidth <= 0 {
		borderWidth = 1
	}

	// Draw the line
	c.FillRect(x, y, w, borderWidth, borderColor)
}

// DrawInput renders an <input> element as a rectangle with border.
func (c *Canvas) DrawInput(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 {
		w = 150
		box.ContentW = 150
	}
	if h <= 0 {
		h = 20
		box.ContentH = 20
	}

	borderColor := css.ParseColor(box.Style["border-color"])
	borderWidth := int(css.ParseLength(box.Style["border-width"]).Value)
	if borderWidth <= 0 {
		borderWidth = 1
	}

	bgColor := css.ParseColor(box.Style["background"])
	if bgColor.A == 0 {
		bgColor = css.Color{R: 255, G: 255, B: 255, A: 255}
	}

	// Draw background
	c.FillRect(x, y, w, h, bgColor)

	// Draw border
	c.DrawBorder(x, y, w, h, borderWidth, borderColor)

	// Draw placeholder text if present
	placeholder := ""
	if box.Node != nil {
		placeholder = box.Node.GetAttribute("placeholder")
	}
	if placeholder != "" {
		textColor := css.Color{R: 169, G: 169, B: 169, A: 255}
		fontSize := 14.0
		charWidth := fontSize * 0.6
		maxChars := w / int(charWidth)
		if maxChars > len(placeholder) {
			maxChars = len(placeholder)
		}
		for i := 0; i < maxChars; i++ {
			c.FillRect(x+int(float64(i)*charWidth)+2, y+h/2-int(fontSize/2), int(charWidth), int(fontSize), textColor)
		}
	}
}

// DrawButton renders a <button> element as a bordered box with text.
func (c *Canvas) DrawButton(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 {
		w = 100
		box.ContentW = 100
	}
	if h <= 0 {
		h = 30
		box.ContentH = 30
	}

	borderColor := css.ParseColor(box.Style["border-color"])
	borderWidth := int(css.ParseLength(box.Style["border-width"]).Value)
	if borderWidth <= 0 {
		borderWidth = 2
	}

	bgColor := css.ParseColor(box.Style["background"])
	if bgColor.A == 0 {
		bgColor = css.Color{R: 240, G: 240, B: 240, A: 255}
	}

	// Draw background
	c.FillRect(x, y, w, h, bgColor)

	// Draw border
	c.DrawBorder(x, y, w, h, borderWidth, borderColor)

	// Draw button text
	textColor := css.ParseColor(box.Style["color"])
	if textColor.A == 0 {
		textColor = css.Color{R: 0, G: 0, B: 0, A: 255}
	}
	fontSize := 14.0
	charWidth := fontSize * 0.6
	padding := 8

	// Get button text from inner text
	buttonText := ""
	if box.Node != nil {
		buttonText = box.Node.InnerText()
	}
	if buttonText == "" {
		buttonText = "Button"
	}

	// Center text horizontally in button
	textWidth := float64(len(buttonText)) * charWidth
	textStartX := float64(x) + (float64(w)-textWidth)/2
	if textStartX < float64(x)+float64(padding) {
		textStartX = float64(x + padding)
	}

	// Draw each character as a simple block representation
	// (This is a simplified text rendering - actual font rendering would be better)
	for i := 0; i < len(buttonText); i++ {
		charX := int(textStartX + float64(i)*charWidth)
		charY := y + h/2 - int(fontSize/2)
		// Draw character block with slight gaps for readability
		if charX+int(charWidth) < x+w-padding {
			c.FillRect(charX, charY, int(charWidth)-1, int(fontSize), textColor)
		}
	}
}

// DrawSelect renders a <select> element as a bordered box with option text.
func (c *Canvas) DrawSelect(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 {
		w = 150
		box.ContentW = 150
	}
	if h <= 0 {
		h = 24
		box.ContentH = 24
	}

	borderColor := css.ParseColor(box.Style["border-color"])
	borderWidth := int(css.ParseLength(box.Style["border-width"]).Value)
	if borderWidth <= 0 {
		borderWidth = 1
	}

	bgColor := css.ParseColor(box.Style["background"])
	if bgColor.A == 0 {
		bgColor = css.Color{R: 255, G: 255, B: 255, A: 255}
	}

	// Draw background
	c.FillRect(x, y, w, h, bgColor)

	// Draw border
	c.DrawBorder(x, y, w, h, borderWidth, borderColor)

	// Get selected option text
	selectText := "Select..."
	if box.Node != nil {
		// Look for selected option
		for _, child := range box.Node.Children {
			if child.TagName == "option" {
				selectText = child.InnerText()
				break
			}
		}
	}

	// Draw select text
	textColor := css.Color{R: 0, G: 0, B: 0, A: 255}
	fontSize := 14.0
	charWidth := fontSize * 0.6
	padding := 6
	arrowSpace := 20 // Space reserved for dropdown arrow

	maxChars := (w - padding*2 - arrowSpace) / int(charWidth)
	if maxChars > len(selectText) {
		maxChars = len(selectText)
	}

	// Draw each character
	for i := 0; i < maxChars; i++ {
		charX := x + padding + int(float64(i)*charWidth)
		charY := y + h/2 - int(fontSize/2)
		c.FillRect(charX, charY, int(charWidth)-1, int(fontSize), textColor)
	}

	// Draw dropdown arrow (simple triangle pointing down)
	arrowX := x + w - 14
	arrowY := y + h/2
	// Triangle: three horizontal lines forming a down arrow
	c.FillRect(arrowX, arrowY-2, 8, 2, textColor)
	c.FillRect(arrowX+1, arrowY, 6, 2, textColor)
	c.FillRect(arrowX+2, arrowY+2, 4, 2, textColor)
}

// DrawTextArea renders a <textarea> element as a bordered multiline box.
func (c *Canvas) DrawTextArea(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 {
		w = 200
		box.ContentW = 200
	}
	if h <= 0 {
		h = 100
		box.ContentH = 100
	}

	borderColor := css.ParseColor(box.Style["border-color"])
	borderWidth := int(css.ParseLength(box.Style["border-width"]).Value)
	if borderWidth <= 0 {
		borderWidth = 1
	}

	bgColor := css.ParseColor(box.Style["background"])
	if bgColor.A == 0 {
		bgColor = css.Color{R: 255, G: 255, B: 255, A: 255}
	}

	// Draw background
	c.FillRect(x, y, w, h, bgColor)

	// Draw border
	c.DrawBorder(x, y, w, h, borderWidth, borderColor)

	// Draw placeholder text area indicator (light gray lines)
	placeholderColor := css.Color{R: 200, G: 200, B: 200, A: 255}
	fontSize := 14.0
	lineHeight := fontSize * 1.2
	numLines := int(float64(h) / lineHeight)
	for i := 0; i < numLines; i++ {
		lineY := y + int(float64(i)*lineHeight) + int(fontSize)
		if lineY < y+h-4 {
			c.FillRect(x+4, lineY, w-8, 1, placeholderColor)
		}
	}
}

// DrawLabel renders a <label> element with visual association to its form element.
func (c *Canvas) DrawLabel(box *layout.Box) {
	// Labels are rendered by their text children, no special rendering needed
	// The text boxes inside will be rendered normally

	// Add visual association if label contains a form element as descendant
	// This handles the case: <label>Name: <input id="name"></label>
	associatedBox := c.findFormElementInLabel(box)
	if associatedBox == nil {
		return
	}

	// Draw a subtle connector line between the last text of label and the form element
	labelX := int(box.ContentX)
	labelY := int(box.ContentY + box.ContentH)
	labelW := int(box.ContentW)

	assocX := int(associatedBox.ContentX)
	assocY := int(associatedBox.ContentY)

	// Draw a connecting line from label to the associated element
	connectorColor := css.Color{R: 100, G: 149, B: 237, A: 200} // Cornflower blue

	// If label and element are on similar Y positions, draw a horizontal connector
	if absInt(assocY - labelY) < 30 || absInt(assocY - int(box.ContentY)) < 30 {
		if labelX + labelW < assocX {
			// Label is to the left, connect right edge to left edge of element
			midY := labelY
			if absInt(assocY - int(box.ContentY)) < absInt(labelY - int(box.ContentY)) {
				midY = int(box.ContentY) + int(box.ContentH/2)
			}
			// Vertical line from label down to element's Y
			c.FillRect(labelX+labelW-2, midY, 2, assocY-midY+4, connectorColor)
			// Horizontal line to element
			c.FillRect(labelX+labelW-2, assocY, assocX-(labelX+labelW-2), 2, connectorColor)
		}
	}
}

// findFormElementInLabel searches for a form element (input, select, textarea, button)
// that is a descendant of the label box.
func (c *Canvas) findFormElementInLabel(box *layout.Box) *layout.Box {
	if box == nil {
		return nil
	}

	// Check if this box is a form element
	if box.Type == layout.ButtonBox || box.Type == layout.SelectBox ||
		box.Type == layout.TextAreaBox || (box.Type == layout.InlineBox && box.Node != nil && box.Node.TagName == "input") {
		return box
	}

	// Search children recursively
	for _, child := range box.Children {
		if found := c.findFormElementInLabel(child); found != nil {
			return found
		}
	}

	return nil
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// DrawMedia renders a <video> or <audio> element with controls UI.
func (c *Canvas) DrawMedia(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	isVideo := box.Node != nil && box.Node.TagName == "video"
	if isVideo {
		if w <= 0 {
			w = 320
			box.ContentW = 320
		}
		if h <= 0 {
			h = 240
			box.ContentH = 240
		}
	} else {
		// Audio
		if w <= 0 {
			w = 300
			box.ContentW = 300
		}
		if h <= 0 {
			h = 40
			box.ContentH = 40
		}
	}

	bgColor := css.Color{R: 0, G: 0, B: 0, A: 255}

	// Draw background
	c.FillRect(x, y, w, h, bgColor)

	// Draw media type text
	textColor := css.Color{R: 255, G: 255, B: 255, A: 255}
	fontSize := 16.0
	charWidth := fontSize * 0.6

	mediaType := "Audio"
	if isVideo {
		mediaType = "Video"
	}
	for i := 0; i < len(mediaType); i++ {
		c.FillRect(x+w/2-int(len(mediaType)*int(charWidth)/2)+int(float64(i)*charWidth), y+h/2-int(fontSize/2), int(charWidth), int(fontSize), textColor)
	}

	// Draw controls UI
	controlsY := y + h - 30
	controlsH := 24

	// Play/pause button
	playX := x + 10
	playW := 30
	c.FillRect(playX, controlsY, playW, controlsH, css.Color{R: 60, G: 60, B: 60, A: 255})
	// Play triangle
	c.FillRect(playX+10, controlsY+6, 10, 12, css.Color{R: 200, G: 200, B: 200, A: 255})

	// Progress bar background
	progressX := playX + playW + 10
	progressW := w - (progressX - x) - 80
	if progressW < 0 {
		progressW = 0
	}
	c.FillRect(progressX, controlsY+8, progressW, 8, css.Color{R: 80, G: 80, B: 80, A: 255})
	// Progress fill (50% for placeholder)
	c.FillRect(progressX, controlsY+8, progressW/2, 8, css.Color{R: 100, G: 150, B: 255, A: 255})

	// Volume control
	volX := x + w - 60
	volW := 50
	c.FillRect(volX, controlsY, volW, controlsH, css.Color{R: 60, G: 60, B: 60, A: 255})
	// Volume bars
	for i := 0; i < 3; i++ {
		barH := 8 + i*4
		c.FillRect(volX+8+i*14, controlsY+controlsH-barH-4, 8, barH, css.Color{R: 200, G: 200, B: 200, A: 255})
	}
}

// DrawProgress renders a <progress> element as a progress bar.
func (c *Canvas) DrawProgress(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 {
		w = 200
		box.ContentW = 200
	}
	if h <= 0 {
		h = 20
		box.ContentH = 20
	}

	// Get value and max attributes
	value := 0.0
	max := 100.0
	if box.Node != nil {
		if valStr := box.Node.GetAttribute("value"); valStr != "" {
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				value = v
			}
		}
		if maxStr := box.Node.GetAttribute("max"); maxStr != "" {
			if m, err := strconv.ParseFloat(maxStr, 64); err == nil && m > 0 {
				max = m
			}
		}
	}

	// Draw track (background)
	trackColor := css.Color{R: 220, G: 220, B: 220, A: 255}
	c.FillRect(x, y, w, h, trackColor)

	// Draw progress bar
	progressWidth := int(float64(w) * (value / max))
	if progressWidth > w {
		progressWidth = w
	}
	barColor := css.Color{R: 0, G: 122, B: 255, A: 255}
	c.FillRect(x, y, progressWidth, h, barColor)

	// Draw border
	borderColor := css.Color{R: 180, G: 180, B: 180, A: 255}
	c.DrawBorder(x, y, w, h, 1, borderColor)
}

// DrawMeter renders a <meter> element as a gauge/progress indicator.
func (c *Canvas) DrawMeter(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 {
		w = 200
		box.ContentW = 200
	}
	if h <= 0 {
		h = 20
		box.ContentH = 20
	}

	// Get value, min, max, low, high, optimum attributes
	value := 50.0
	min := 0.0
	max := 100.0
	low := 25.0
	high := 75.0
	optimum := 50.0

	if box.Node != nil {
		if valStr := box.Node.GetAttribute("value"); valStr != "" {
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				value = v
			}
		}
		if minStr := box.Node.GetAttribute("min"); minStr != "" {
			if m, err := strconv.ParseFloat(minStr, 64); err == nil {
				min = m
			}
		}
		if maxStr := box.Node.GetAttribute("max"); maxStr != "" {
			if m, err := strconv.ParseFloat(maxStr, 64); err == nil {
				max = m
			}
		}
		if lowStr := box.Node.GetAttribute("low"); lowStr != "" {
			if l, err := strconv.ParseFloat(lowStr, 64); err == nil {
				low = l
			}
		}
		if highStr := box.Node.GetAttribute("high"); highStr != "" {
			if h, err := strconv.ParseFloat(highStr, 64); err == nil {
				high = h
			}
		}
		if optStr := box.Node.GetAttribute("optimum"); optStr != "" {
			if o, err := strconv.ParseFloat(optStr, 64); err == nil {
				optimum = o
			}
		}
	}

	// Determine color based on value region
	barColor := css.Color{R: 0, G: 180, B: 0, A: 255} // green for optimum
	if value < low || value > high {
		if (optimum >= low && optimum <= high) {
			// optimum in middle - value is in outer region (bad)
			barColor = css.Color{R: 255, G: 100, B: 100, A: 255} // red
		}
	} else if value < optimum-10 || value > optimum+10 {
		barColor = css.Color{R: 255, G: 200, B: 0, A: 255} // yellow
	}

	// Draw track (background)
	trackColor := css.Color{R: 220, G: 220, B: 220, A: 255}
	c.FillRect(x, y, w, h, trackColor)

	// Draw meter bar
	meterWidth := int(float64(w) * ((value - min) / (max - min)))
	if meterWidth > w {
		meterWidth = w
	}
	if meterWidth < 0 {
		meterWidth = 0
	}
	c.FillRect(x, y, meterWidth, h, barColor)

	// Draw border
	borderColor := css.Color{R: 180, G: 180, B: 180, A: 255}
	c.DrawBorder(x, y, w, h, 1, borderColor)
}

// DrawDialog renders a <dialog> element with a modal backdrop.
func (c *Canvas) DrawDialog(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 {
		w = 400
		box.ContentW = 400
	}
	if h <= 0 {
		h = 200
		box.ContentH = 200
	}

	// Draw semi-transparent backdrop if open
	isOpen := true
	if box.Node != nil {
		if openStr := box.Node.GetAttribute("open"); openStr == "" {
			isOpen = false
		}
	}

	if isOpen {
		// Draw backdrop over entire viewport (approximation)
		backdropColor := css.Color{R: 0, G: 0, B: 0, A: 128}
		// Draw a subtle overlay behind the dialog
		c.FillRect(x-10, y-10, w+20, h+20, backdropColor)
	}

	// Draw dialog box background
	dialogBg := css.Color{R: 255, G: 255, B: 255, A: 255}
	c.FillRect(x, y, w, h, dialogBg)

	// Draw border
	borderColor := css.Color{R: 100, G: 100, B: 100, A: 255}
	c.DrawBorder(x, y, w, h, 2, borderColor)
}

// drawLinearGradient draws a linear gradient filling the given rectangle.
func (c *Canvas) drawLinearGradient(x, y, w, h int, stops []css.GradientStop, direction string, opacity float64) {
	if w <= 0 || h <= 0 || len(stops) < 2 {
		return
	}
	// Resolve first and last stop colors and offsets
	stop0Off := stops[0].Offset
	stop0Color := stops[0].Color
	stopN := len(stops) - 1
	stopNOff := stops[stopN].Offset
	stopNColor := stops[stopN].Color
	if len(stops) > 1 {
		if stops[1].Offset >= 0 {
			stopNColor = stops[1].Color
		}
	}
	if stop0Off < 0 {
		stop0Off = 0
	}
	if stopNOff < 0 {
		stopNOff = 1
	}

	for py := 0; py < h; py++ {
		t := float64(py) / float64(h)
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}

		var r, g, b, a uint8
		if t <= stop0Off {
			r, g, b, a = stop0Color.R, stop0Color.G, stop0Color.B, stop0Color.A
		} else if t >= stopNOff {
			r, g, b, a = stopNColor.R, stopNColor.G, stopNColor.B, stopNColor.A
		} else {
			rangeSize := stopNOff - stop0Off
			if rangeSize > 0 {
				localT := (t - stop0Off) / rangeSize
				r = uint8(float64(stop0Color.R)*(1-localT) + float64(stopNColor.R)*localT)
				g = uint8(float64(stop0Color.G)*(1-localT) + float64(stopNColor.G)*localT)
				b = uint8(float64(stop0Color.B)*(1-localT) + float64(stopNColor.B)*localT)
				a = uint8(float64(stop0Color.A)*(1-localT) + float64(stopNColor.A)*localT)
			}
		}

		if opacity < 1 {
			a = uint8(float64(a) * opacity)
		}
		rowColor := css.Color{R: r, G: g, B: b, A: a}
		c.FillRect(x, y+py, w, 1, rowColor)
	}
}

// drawRadialGradient draws a radial gradient filling the given rectangle.
func (c *Canvas) drawRadialGradient(x, y, w, h int, stops []css.GradientStop, shape, position string, opacity float64) {
	if w <= 0 || h <= 0 || len(stops) < 2 {
		return
	}
	cx := float64(w) / 2
	cy := float64(h) / 2
	maxRadius := math.Min(cx, cy)

	startColor := stops[0].Color
	endColor := stops[len(stops)-1].Color
	if len(stops) > 1 && stops[1].Offset >= 0 {
		endColor = stops[1].Color
	}

	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			dx := float64(px) - cx
			dy := float64(py) - cy
			dist := math.Sqrt(dx*dx + dy*dy)

			t := dist / maxRadius
			if t > 1 {
				t = 1
			}

			r := uint8(float64(startColor.R)*(1-t) + float64(endColor.R)*t)
			g := uint8(float64(startColor.G)*(1-t) + float64(endColor.G)*t)
			b := uint8(float64(startColor.B)*(1-t) + float64(endColor.B)*t)
			a := uint8(float64(startColor.A)*(1-t) + float64(endColor.A)*t)

			if opacity < 1 {
				a = uint8(float64(a) * opacity)
			}
			pixelColor := css.Color{R: r, G: g, B: b, A: a}
			c.FillRect(x+px, y+py, 1, 1, pixelColor)
		}
	}
}
