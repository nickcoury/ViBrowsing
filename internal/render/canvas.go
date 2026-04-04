package render

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/layout"
)

// Canvas represents a pixel buffer for rendering.
type Canvas struct {
	Width  int
	Height int
	Pixels *image.RGBA
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

// FillRect fills a rectangle with a color.
func (c *Canvas) FillRect(x, y, w, h int, col color.Color) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			c.SetPixel(x+dx, y+dy, col)
		}
	}
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

	// Background
	bgColor := css.ParseColor(box.Style["background"])
	c.FillRect(contentX, contentY, contentW, contentH, bgColor)

	// Border
	borderWidth := int(css.ParseLength(box.Style["border-width"]).Value)
	if borderWidth > 0 {
		borderColor := css.ParseColor(box.Style["border-color"])
		borderStyle := box.Style["border-style"]
		if borderStyle == "none" || borderStyle == "" {
			borderColor = css.Color{R: 0, G: 0, B: 0, A: 255}
		}
		c.DrawBorder(contentX, contentY, contentW, contentH, borderWidth, borderColor)
	}

	// Padding (fill)
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

	// Check visibility - hidden boxes paint background/border/padding but not content
	visibility, _ := box.Style["visibility"]
	isHidden := visibility == "hidden"

	// Text content (only if not hidden)
	if !isHidden && box.Type == layout.TextBox && box.Node != nil {
		c.DrawText(box)
	}

	// Children (only if not hidden)
	if !isHidden {
		for _, child := range box.Children {
			c.DrawBox(child)
		}
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

	textColor := css.ParseColor(box.Style["color"])
	x := int(box.ContentX)
	y := int(box.ContentY)

	charWidth := int(fontSize * 0.6)
	lineHeight := int(fontSize * 1.2)

	currentX := x
	currentY := y

	for range text {
		if currentX+charWidth > x+int(box.ContentW) && currentX > x {
			currentX = x
			currentY += lineHeight
		}
		c.FillRect(currentX, currentY, charWidth, int(fontSize), textColor)
		currentX += charWidth
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
