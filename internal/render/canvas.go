package render

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/nickcoury/ViBrowsing/internal/css"
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

	contentX := int(box.ContentX)
	contentY := int(box.ContentY)
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

	// Image content
	if !isHidden && box.Type == layout.ImageBox && box.Node != nil {
		c.DrawImage(box)
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

	for i := 0; i < len(text); i++ {
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
		c.FillRect(int(currentX), int(currentY), int(cw), int(charH), textColor)
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

	// Draw the image, scaled to fit the content box
	c.drawImageScaled(img, x, y, w, h)
}

// loadImage fetches and decodes an image from a URL.
// Returns the image or an error. Currently a stub — returns nil.
func (c *Canvas) loadImage(src string) (image.Image, error) {
	// TODO: implement actual image fetching and decoding
	// For now, return nil so alt text / broken image icon is shown
	return nil, nil
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
