package render

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/nickcoury/ViBrowsing/internal/css"
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

	// Background (margin box area)
	bgColor := css.ParseColor(box.Style["background"])
	if opacity < 1 {
		bgColor = applyOpacity(bgColor, opacity)
	}
	marginLeft := int(css.ParseLength(box.Style["margin-left"]).Value)
	marginTop := int(css.ParseLength(box.Style["margin-top"]).Value)
	marginRight := int(css.ParseLength(box.Style["margin-right"]).Value)
	marginBottom := int(css.ParseLength(box.Style["margin-bottom"]).Value)
	marginX := contentX - marginLeft
	marginY := contentY - marginTop
	marginW := marginLeft + contentW + marginRight
	marginH := marginTop + contentH + marginBottom
	c.FillRect(marginX, marginY, marginW, marginH, bgColor)

	// Padding (fill the padding box, drawn before border so border is on top)
	paddingTop := int(css.ParseLength(box.Style["padding-top"]).Value)
	paddingRight := int(css.ParseLength(box.Style["padding-right"]).Value)
	paddingBottom := int(css.ParseLength(box.Style["padding-bottom"]).Value)
	paddingLeft := int(css.ParseLength(box.Style["padding-left"]).Value)

	paddingX := contentX + paddingLeft
	paddingY := contentY + paddingTop
	paddingW := contentW - paddingLeft - paddingRight
	paddingH := contentH - paddingTop - paddingBottom

	if paddingW > 0 && paddingH > 0 {
		c.FillRect(paddingX, paddingY, paddingW, paddingH, bgColor)
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

	// Children (only if not hidden)
	if !isHidden {
		// Sort children by z-index for proper stacking order
		children := make([]*layout.Box, len(box.Children))
		copy(children, box.Children)
		c.drawChildrenSorted(children)
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

// DrawBorder draws a border around a rectangle.
func (c *Canvas) DrawBorder(x, y, w, h, bw int, col color.Color) {
	c.FillRect(x, y, w, bw, col)
	c.FillRect(x, y+h-bw, w, bw, col)
	c.FillRect(x, y, bw, h, col)
	c.FillRect(x+w-bw, y, bw, h, col)
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

// SavePNG saves the canvas to a PNG file.
func (c *Canvas) SavePNG(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, c.Pixels)
}

// ToImage returns the underlying *image.RGBA.
func (c *Canvas) ToImage() *image.RGBA {
	return c.Pixels
}
