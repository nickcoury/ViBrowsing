package render

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/fetch"
	"github.com/nickcoury/ViBrowsing/internal/html"
	"github.com/nickcoury/ViBrowsing/internal/layout"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// isEmoji returns true if the rune is an emoji character.
// Emoji ranges include:
// - U+1F300 to U+1F9FF (misc symbols, dingbats, emoji)
// - U+2600 to U+26FF (misc symbols)
// - U+2700 to U+27BF (dingbats)
// - Supplemental emoji: U+1F600 to U+1F64F, U+1F680 to U+1F6FF, U+1F900 to U+1F9FF
func isEmoji(r rune) bool {
	// Miscellaneous Symbols and Pictographs (U+1F300 to U+1F5FF)
	if r >= 0x1F300 && r <= 0x1F5FF {
		return true
	}
	// Emoticons (U+1F600 to U+1F64F)
	if r >= 0x1F600 && r <= 0x1F64F {
		return true
	}
	// Transport and Map Symbols (U+1F680 to U+1F6FF)
	if r >= 0x1F680 && r <= 0x1F6FF {
		return true
	}
	// Supplemental Symbols and Pictographs (U+1F900 to U+1F9FF)
	if r >= 0x1F900 && r <= 0x1F9FF {
		return true
	}
	// Miscellaneous Symbols (U+2600 to U+26FF)
	if r >= 0x2600 && r <= 0x26FF {
		return true
	}
	// Dingbats (U+2700 to U+27BF)
	if r >= 0x2700 && r <= 0x27BF {
		return true
	}
	return false
}

// drawEmojiPlaceholder draws a colored circle as a placeholder for an emoji character.
// The color is derived from the emoji's codepoint to provide visual distinction.
func (c *Canvas) drawEmojiPlaceholder(x, y float64, size float64, r rune, col color.Color) {
	// Derive color components from the codepoint
	// Use a hash of the codepoint to generate distinct colors
	cp := int(r)
	// Generate hue from codepoint (0-360)
	hue := float64(cp%360) * 2 // Multiply to spread colors better
	// Convert HSV to RGB for a colorful circle
	saturation := 0.7
	value := 0.9

	// Convert HSV to RGB
	h := hue / 60.0
	hf := math.Floor(h)
	hi := int(hf) % 6
	f := h - hf

	v := value * 255
	p := value * (1 - saturation) * 255
	q := value * (1 - saturation*f) * 255
	t := value * (1 - saturation*(1-f)) * 255

	var cr, cg, cb uint8
	switch hi {
	case 0:
		cr, cg, cb = uint8(v), uint8(t), uint8(p)
	case 1:
		cr, cg, cb = uint8(q), uint8(v), uint8(p)
	case 2:
		cr, cg, cb = uint8(p), uint8(v), uint8(t)
	case 3:
		cr, cg, cb = uint8(p), uint8(q), uint8(v)
	case 4:
		cr, cg, cb = uint8(t), uint8(p), uint8(v)
	case 5:
		cr, cg, cb = uint8(v), uint8(p), uint8(q)
	}

	emojiColor := color.RGBA{R: cr, G: cg, B: cb, A: 255}

	// Draw a filled circle as the emoji placeholder
	centerX := int(x + size/2)
	centerY := int(y + size/2)
	radius := int(size / 2)

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				px := centerX + dx
				py := centerY + dy
				if px >= 0 && px < c.Width && py >= 0 && py < c.Height {
					c.Pixels.SetRGBA(px, py, emojiColor)
				}
			}
		}
	}
}

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

	// fontCache caches loaded fonts by family+size for reuse
	fontCache map[string]*truetype.Font

	// advanceWidths caches avg advance widths per font+size
	advanceWidths map[string]float64

	// textMeasureCache caches text width measurements per font/size/text
	// Key: "fontFamily:fontSize:text"
	textMeasureCache sync.Map

	// Default font
	defaultFont *truetype.Font
	defaultAdvWidth float64
}

// fontPaths maps CSS font-family names to system font file paths
var fontPaths = map[string]string{
	"serif":      "/usr/share/fonts/truetype/liberation/LiberationSerif-Regular.ttf",
	"sans-serif": "/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
	"monospace":  "/usr/share/fonts/truetype/liberation/LiberationMono-Regular.ttf",
}

// NewCanvas creates a new canvas with the given dimensions.
func NewCanvas(width, height int) *Canvas {
	c := &Canvas{
		Width:          width,
		Height:         height,
		Pixels:         image.NewRGBA(image.Rect(0, 0, width, height)),
		fontCache:      make(map[string]*truetype.Font),
		advanceWidths: make(map[string]float64),
	}
	c.loadFonts()
	return c
}

// loadFonts loads system fonts for text rendering
func (c *Canvas) loadFonts() {
	for name, path := range fontPaths {
		f, err := loadFontFile(path)
		if err == nil {
			c.fontCache[name] = f
			// Precompute advance width at 16px as default
			c.advanceWidths[name+"_16"] = c.computeAdvanceWidth(f, 16)
		}
	}
	// Set default font
	if f, ok := c.fontCache["sans-serif"]; ok {
		c.defaultFont = f
		c.defaultAdvWidth = c.advanceWidths["sans-serif_16"]
	} else if len(c.fontCache) > 0 {
		for name, f := range c.fontCache {
			c.defaultFont = f
			c.defaultAdvWidth = c.advanceWidths[name+"_16"]
			break
		}
	}
}

// computeAdvanceWidth computes the average advance width for a font at a given size
func (c *Canvas) computeAdvanceWidth(f *truetype.Font, fontSize float64) float64 {
	if f == nil {
		return fontSize * 0.6
	}
	// Convert fontSize to 26.6 fixed point scale
	scale := fixed.Int26_6(fontSize * 64)
	
	// Sample a-z to get average width
	var total fixed.Int26_6
	count := 0
	for i := rune('a'); i <= rune('z'); i++ {
		idx := f.Index(i)
		if idx != 0 {
			hm := f.HMetric(scale, idx)
			total += hm.AdvanceWidth
			count++
		}
	}
	if count > 0 {
		avgFixed := total / fixed.Int26_6(count)
		return float64(avgFixed) / 64.0
	}
	return fontSize * 0.6
}

// MeasureText returns the width in pixels of the text for the given font family and size.
// Results are cached to avoid repeated measurements of the same text.
func (c *Canvas) MeasureText(fontFamily string, fontSize float64, text string) float64 {
	// Normalize font family name (same logic as GetFont)
	family := strings.ToLower(fontFamily)
	if family == "serif" || family == "times" || family == "georgia" {
		family = "serif"
	} else if family == "monospace" || family == "courier" || family == "consolas" {
		family = "monospace"
	} else {
		family = "sans-serif"
	}

	// Create cache key: "fontFamily:fontSize:text"
	cacheKey := family + ":" + strconv.FormatFloat(fontSize, 'f', 1, 64) + ":" + text

	// Check cache
	if cached, ok := c.textMeasureCache.Load(cacheKey); ok {
		return cached.(float64)
	}

	// Get font
	f, _ := c.GetFont(fontFamily, fontSize)
	if f == nil {
		return fontSize * 0.6 * float64(len(text))
	}

	// Measure the text by summing advance widths of each character
	face := truetype.NewFace(f, &truetype.Options{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	var totalWidth fixed.Int26_6
	prevRune := rune(-1)

	for _, r := range text {
		if prevRune >= 0 {
			// Add kerning between previous and current character
			totalWidth += face.Kern(prevRune, r)
		}
		_, _, _, adv, ok := face.Glyph(fixed.Point26_6{}, r)
		if !ok {
			// Use average advance width if glyph not found
			adv = fixed.Int26_6(fontSize * 64 * 6 / 10)
		}
		totalWidth += adv
		prevRune = r
	}

	// Convert from 26.6 fixed point to float
	result := float64(totalWidth) / 64.0

	// Store in cache
	c.textMeasureCache.Store(cacheKey, result)

	return result
}

// loadFontFile loads a TrueType font file and returns the parsed font
func loadFontFile(path string) (*truetype.Font, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return truetype.Parse(data)
}

// GetFont returns a font and its average advance width for the given CSS font properties
func (c *Canvas) GetFont(fontFamily string, fontSize float64) (*truetype.Font, float64) {
	// Normalize font family name
	family := strings.ToLower(fontFamily)
	if family == "serif" || family == "times" || family == "georgia" {
		family = "serif"
	} else if family == "monospace" || family == "courier" || family == "consolas" {
		family = "monospace"
	} else {
		family = "sans-serif"
	}

	// Check cache
	if f, ok := c.fontCache[family]; ok {
		return f, c.computeAdvanceWidth(f, fontSize)
	}

	return c.defaultFont, c.defaultAdvWidth
}

// DrawGlyph draws a single glyph at the specified position using the freetype rasterizer
func (c *Canvas) DrawGlyph(f *truetype.Font, fontSize, x, y float64, col color.Color, ch rune) {
	if f == nil {
		return
	}
	
	// Create a font face for this size
	face := truetype.NewFace(f, &truetype.Options{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	
	// The dot position in 26.6 fixed point (baseline of glyph)
	dot := fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6(y * 64),
	}
	
	// Get the glyph mask - Glyph returns 5 values: dr, mask, maskp, advance, ok
	dr, mask, maskp, _, ok := face.Glyph(dot, ch)
	if !ok {
		return
	}
	
	// Create a temporary RGBA image for the color
	tmp := image.NewRGBA(dr)
	draw.Draw(tmp, dr, &image.Uniform{col}, image.ZP, draw.Src)
	
	// Draw the glyph mask onto our canvas using draw.DrawMask
	// maskp is the offset within the mask where the glyph starts
	dstRect := dr.Sub(dr.Min).Add(image.Point{
		X: int(x) + maskp.X - dr.Min.X,
		Y: int(y) + maskp.Y - dr.Min.Y,
	})
	if dstRect.Min.X < 0 {
		mask = mask.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(image.Rectangle{
			Min: image.Point{X: -dstRect.Min.X, Y: 0},
			Max: mask.Bounds().Max,
		})
		dstRect.Min.X = 0
	}
	if dstRect.Min.Y < 0 {
		mask = mask.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(image.Rectangle{
			Min: image.Point{X: 0, Y: -dstRect.Min.Y},
			Max: mask.Bounds().Max,
		})
		dstRect.Min.Y = 0
	}
	
	// Use draw.DrawMask to composite the glyph
	srcRect := tmp.Bounds()
	draw.DrawMask(c.Pixels, dstRect, tmp, srcRect.Min, mask, maskp, draw.Over)
}

// InnerWidth returns the viewport width (window.innerWidth equivalent).
func (c *Canvas) InnerWidth() int {
	return c.Width
}

// InnerHeight returns the viewport height (window.innerHeight equivalent).
func (c *Canvas) InnerHeight() int {
	return c.Height
}

// CursorType represents the cursor style for CSS cursor property.
type CursorType string

const (
	CursorAuto       CursorType = "auto"
	CursorDefault    CursorType = "default"
	CursorPointer    CursorType = "pointer"
	CursorText       CursorType = "text"
	CursorWait       CursorType = "wait"
	CursorCrosshair  CursorType = "crosshair"
	CursorMove       CursorType = "move"
	CursorHelp       CursorType = "help"
	CursorNResize    CursorType = "n-resize"
	CursorEResize    CursorType = "e-resize"
	CursorSResize    CursorType = "s-resize"
	CursorWResize    CursorType = "w-resize"
	CursorNEResize   CursorType = "ne-resize"
	CursorNWResize   CursorType = "nw-resize"
	CursorSEResize   CursorType = "se-resize"
	CursorSWResize   CursorType = "sw-resize"
	CursorNotAllowed CursorType = "not-allowed"
	CursorGrab       CursorType = "grab"
	CursorGrabbing   CursorType = "grabbing"
)

// ParseCursor parses the CSS cursor property and returns a CursorType.
func ParseCursor(cursorStr string) CursorType {
	switch cursorStr {
	case "auto":
		return CursorAuto
	case "default":
		return CursorDefault
	case "pointer":
		return CursorPointer
	case "text":
		return CursorText
	case "wait":
		return CursorWait
	case "crosshair":
		return CursorCrosshair
	case "move":
		return CursorMove
	case "help":
		return CursorHelp
	case "n-resize", "ns-resize":
		return CursorNResize
	case "e-resize", "ew-resize":
		return CursorEResize
	case "s-resize":
		return CursorSResize
	case "w-resize":
		return CursorWResize
	case "ne-resize", "nesw-resize":
		return CursorNEResize
	case "nw-resize", "nwse-resize":
		return CursorNWResize
	case "se-resize":
		return CursorSEResize
	case "sw-resize":
		return CursorSWResize
	case "not-allowed":
		return CursorNotAllowed
	case "grab":
		return CursorGrab
	case "grabbing":
		return CursorGrabbing
	default:
		return CursorAuto
	}
}

// DrawCursorIndicator draws a cursor indicator at the specified position.
// The cursor style is determined by the element type and CSS cursor property.
func (c *Canvas) DrawCursorIndicator(x, y int, cursorStr string, elementType string) {
	cursor := ParseCursor(cursorStr)

	// Determine actual cursor based on element type if cursor is auto
	if cursor == CursorAuto {
		switch elementType {
		case "a", "button", "input", "select", "textarea", "label":
			cursor = CursorPointer
		case "text", "p", "h1", "h2", "h3", "h4", "h5", "h6", "li":
			cursor = CursorText
		default:
			cursor = CursorDefault
		}
	}

	// Don't draw cursor indicator for default/auto (no visible cursor)
	if cursor == CursorAuto || cursor == CursorDefault {
		return
	}

	color := css.Color{R: 0, G: 0, B: 0, A: 200}

	switch cursor {
	case CursorPointer:
		// Draw a small pointing hand (triangle arrow)
		// Arrow pointing up-left from cursor position
		c.SetPixel(x, y, color)
		c.SetPixel(x+1, y, color)
		c.SetPixel(x+2, y, color)
		c.SetPixel(x, y+1, color)
		c.SetPixel(x, y+2, color)
		c.SetPixel(x+1, y+1, color)
		c.SetPixel(x, y+3, color)
		c.SetPixel(x+3, y, color)

	case CursorText:
		// Draw a text cursor (vertical bar)
		c.FillRect(x, y, 2, 16, color)

	case CursorWait:
		// Draw a busy/wait cursor (hourglass shape approximated)
		c.SetPixel(x, y, color)
		c.SetPixel(x+4, y, color)
		c.SetPixel(x+1, y+1, color)
		c.SetPixel(x+3, y+1, color)
		c.SetPixel(x+2, y+2, color)
		c.SetPixel(x+1, y+3, color)
		c.SetPixel(x+3, y+3, color)
		c.SetPixel(x, y+4, color)
		c.SetPixel(x+4, y+4, color)

	case CursorCrosshair:
		// Draw a crosshair
		c.SetPixel(x+2, y, color)
		c.SetPixel(x+2, y+1, color)
		c.SetPixel(x+2, y+3, color)
		c.SetPixel(x+2, y+4, color)
		c.SetPixel(x, y+2, color)
		c.SetPixel(x+1, y+2, color)
		c.SetPixel(x+3, y+2, color)
		c.SetPixel(x+4, y+2, color)

	case CursorMove:
		// Draw a move cursor (four-way arrow)
		c.SetPixel(x+2, y, color)
		c.SetPixel(x+2, y+4, color)
		c.SetPixel(x, y+2, color)
		c.SetPixel(x+4, y+2, color)
		c.SetPixel(x+1, y+1, color)
		c.SetPixel(x+3, y+1, color)
		c.SetPixel(x+1, y+3, color)
		c.SetPixel(x+3, y+3, color)

	case CursorHelp:
		// Draw a help cursor (question mark)
		c.SetPixel(x+1, y, color)
		c.SetPixel(x+2, y+1, color)
		c.SetPixel(x+2, y+2, color)
		c.SetPixel(x+1, y+3, color)
		c.SetPixel(x+2, y+4, color)
		c.SetPixel(x+2, y+5, color)

	case CursorNResize, CursorEResize, CursorSResize, CursorWResize:
		// Draw resize cursors (arrows)
		// N (up arrow)
		c.SetPixel(x+2, y, color)
		c.SetPixel(x+1, y+1, color)
		c.SetPixel(x+3, y+1, color)
		c.SetPixel(x+2, y+2, color)
		c.SetPixel(x+2, y+3, color)

	case CursorNEResize:
		// NE resize (arrow pointing up-right)
		c.SetPixel(x+3, y, color)
		c.SetPixel(x+2, y+1, color)
		c.SetPixel(x+1, y+2, color)
		c.SetPixel(x, y+3, color)
		c.SetPixel(x+3, y+3, color)

	case CursorNWResize:
		// NW resize (arrow pointing up-left)
		c.SetPixel(x, y, color)
		c.SetPixel(x+1, y+1, color)
		c.SetPixel(x+2, y+2, color)
		c.SetPixel(x+3, y+3, color)

	case CursorSEResize:
		// SE resize (arrow pointing down-right)
		c.SetPixel(x, y, color)
		c.SetPixel(x+3, y+1, color)
		c.SetPixel(x+2, y+2, color)
		c.SetPixel(x+1, y+3, color)

	case CursorSWResize:
		// SW resize (arrow pointing down-left)
		c.SetPixel(x+3, y, color)
		c.SetPixel(x, y+1, color)
		c.SetPixel(x+1, y+2, color)
		c.SetPixel(x+2, y+3, color)

	case CursorNotAllowed:
		// Draw a not-allowed/crossed circle
		c.SetPixel(x+2, y, color)
		c.SetPixel(x, y+2, color)
		c.SetPixel(x+4, y+2, color)
		c.SetPixel(x+2, y+4, color)
		c.SetPixel(x, y, color)
		c.SetPixel(x+4, y, color)
		c.SetPixel(x, y+4, color)
		c.SetPixel(x+4, y+4, color)

	case CursorGrab:
		// Draw a grab cursor (open hand)
		c.SetPixel(x, y, color)
		c.SetPixel(x+3, y, color)
		c.SetPixel(x+1, y+1, color)
		c.SetPixel(x+2, y+1, color)
		c.SetPixel(x+1, y+2, color)
		c.SetPixel(x+2, y+2, color)

	case CursorGrabbing:
		// Draw a grabbing cursor (closed hand)
		c.SetPixel(x, y, color)
		c.SetPixel(x+3, y, color)
		c.SetPixel(x, y+1, color)
		c.SetPixel(x+3, y+1, color)
		c.SetPixel(x+1, y+2, color)
		c.SetPixel(x+2, y+2, color)
	}
}

// GetCursorForElement returns the appropriate cursor string for an element.
func GetCursorForElement(box *layout.Box) string {
	// First check CSS cursor property
	if cursor, ok := box.Style["cursor"]; ok && cursor != "" {
		return cursor
	}

	// Fall back to element-based defaults
	if box.Node != nil {
		switch box.Node.TagName {
		case "a":
			return "pointer"
		case "button", "input", "select", "textarea", "label":
			return "pointer"
		case "text", "p", "h1", "h2", "h3", "h4", "h5", "h6", "li", "span":
			return "text"
		case "img":
			return "pointer"
		}
	}

	return "auto"
}

// ScrollTo scrolls to the given position. Supports both ScrollTo(x, y) and
// ScrollTo({left, top, behavior}) syntax.
func (c *Canvas) ScrollTo(x interface{}, yOpt ...interface{}) {
	switch arg := x.(type) {
	case float64:
		// ScrollTo(x, y)
		c.ScrollX = int(arg)
		if len(yOpt) > 0 {
			if y, ok := yOpt[0].(float64); ok {
				c.ScrollY = int(y)
			}
		}
	case int:
		// ScrollTo(x, y) with int
		c.ScrollX = arg
		if len(yOpt) > 0 {
			if y, ok := yOpt[0].(int); ok {
				c.ScrollY = y
			}
		}
	case map[string]interface{}:
		// ScrollTo({left, top, behavior})
		if left, ok := arg["left"].(float64); ok {
			c.ScrollX = int(left)
		} else if left, ok := arg["left"].(int); ok {
			c.ScrollX = left
		}
		if top, ok := arg["top"].(float64); ok {
			c.ScrollY = int(top)
		} else if top, ok := arg["top"].(int); ok {
			c.ScrollY = top
		}
		// behavior is ignored (no smooth scrolling in this renderer)
	}
	// Clamp to valid scroll range
	c.clampScroll()
}

// ScrollBy scrolls by the given delta. Supports both ScrollBy(dx, dy) and
// ScrollBy({left, top, behavior}) syntax.
func (c *Canvas) ScrollBy(x interface{}, dyOpt ...interface{}) {
	switch arg := x.(type) {
	case float64:
		// ScrollBy(dx, dy)
		c.ScrollX += int(arg)
		if len(dyOpt) > 0 {
			if dy, ok := dyOpt[0].(float64); ok {
				c.ScrollY += int(dy)
			}
		}
	case int:
		// ScrollBy(dx, dy) with int
		c.ScrollX += arg
		if len(dyOpt) > 0 {
			if dy, ok := dyOpt[0].(int); ok {
				c.ScrollY += dy
			}
		}
	case map[string]interface{}:
		// ScrollBy({left, top, behavior})
		if left, ok := arg["left"].(float64); ok {
			c.ScrollX += int(left)
		} else if left, ok := arg["left"].(int); ok {
			c.ScrollX += left
		}
		if top, ok := arg["top"].(float64); ok {
			c.ScrollY += int(top)
		} else if top, ok := arg["top"].(int); ok {
			c.ScrollY += top
		}
	}
	// Clamp to valid scroll range
	c.clampScroll()
}

// clampScroll ensures scroll position stays within valid bounds.
func (c *Canvas) clampScroll() {
	maxScrollY := 0
	if c.PageHeight > c.Height {
		maxScrollY = c.PageHeight - c.Height
	}
	if c.ScrollY < 0 {
		c.ScrollY = 0
	}
	if c.ScrollY > maxScrollY {
		c.ScrollY = maxScrollY
	}
}

// ScrollIntoView scrolls the element's box into the viewport.
// The options parameter can be:
//   - bool: true means top of element aligns with top of viewport, false means bottom
//   - map with "block": "start", "center", "end", or "nearest"
//   - map with "inline": "start", "center", "end", or "nearest"
//   - map with "behavior": "smooth" or "auto" (smooth is ignored)
//
// This method finds the box corresponding to the node and scrolls to bring
// it into view within the canvas's scrollable area.
func (c *Canvas) ScrollIntoView(node *html.Node, options interface{}) {
	// Find the box for this node
	box := c.findBoxByNode(node)
	if box == nil {
		return
	}

	c.ScrollBoxIntoView(box, options)
}

// ScrollBoxIntoView scrolls the given box into the viewport.
// This is an alternative to ScrollIntoView when you already have the box.
// The options parameter can be:
//   - bool: true means top of element aligns with top of viewport, false means bottom
//   - map with "block": "start", "center", "end", or "nearest"
//   - map with "inline": "start", "center", "end", or "nearest"
//   - map with "behavior": "smooth" or "auto" (smooth is ignored)
func (c *Canvas) ScrollBoxIntoView(box *layout.Box, options interface{}) {
	if box == nil {
		return
	}

	// Get the scroll target from the box
	targetY := box.ScrollIntoView(options)

	// Adjust for viewport centering or bottom alignment
	if opts, ok := options.(map[string]interface{}); ok {
		if block, ok := opts["block"].(string); ok {
			switch block {
			case "center":
				targetY -= c.Height / 2
			case "end":
				targetY -= c.Height - int(box.ContentH)
			case "nearest":
				// For nearest, we'd need to know current scroll position
				// Simplified: scroll minimum to bring element into view
				if c.ScrollY > targetY {
					// Element is above viewport, scroll up
				} else if c.ScrollY+c.Height < targetY+int(box.ContentH) {
					// Element is below viewport, scroll down to bottom of element
					targetY = targetY + int(box.ContentH) - c.Height
				}
			}
		}
	} else if alignBool, ok := options.(bool); ok {
		if !alignBool {
			// false means align to bottom
			targetY -= c.Height - int(box.ContentH)
		}
	}

	c.ScrollTo(float64(c.ScrollX), float64(targetY))
}

// findBoxByNode finds a layout box by traversing the layout tree.
// This is a helper that would typically use the layout tree.
func (c *Canvas) findBoxByNode(node *html.Node) *layout.Box {
	// This method needs access to the layout tree
	// In the current architecture, Canvas doesn't store the layout tree
	// We need to either:
	// 1. Store a reference to the layout tree in Canvas
	// 2. Have a separate method that takes the box directly
	// For now, return nil - the ScrollIntoView on Box can be called directly
	return nil
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

// applyClipPath applies a CSS clip-path to the canvas.
// For inset: uses PushClip with inset rectangle
// For circle/ellipse: draws content to temp image and masks with ellipse
// For polygon: uses bounding box (full polygon clipping requires GPU)
func (c *Canvas) applyClipPath(box *layout.Box) (restore func(), err error) {
	clipPath := box.Style["clip-path"]
	if clipPath == "" || clipPath == "none" {
		return func() {}, nil
	}

	shape, err := css.ParseClipPath(clipPath)
	if err != nil || shape.Type == "none" {
		return func() {}, err
	}

	contentX := int(box.ContentX) - c.ScrollX
	contentY := int(box.ContentY) - c.ScrollY
	contentW := int(box.ContentW)
	contentH := int(box.ContentH)

	if contentW <= 0 || contentH <= 0 {
		contentW = 800
		contentH = 600
	}

	switch shape.Type {
	case "inset":
		// inset(top right bottom left) - inset from box edges
		insetTop := int(shape.InsetTop)
		insetRight := int(shape.InsetRight)
		insetBottom := int(shape.InsetBottom)
		insetLeft := int(shape.InsetLeft)

		clipX := contentX + insetLeft
		clipY := contentY + insetTop
		clipW := contentW - insetLeft - insetRight
		clipH := contentH - insetTop - insetBottom

		if clipW < 0 {
			clipW = 0
		}
		if clipH < 0 {
			clipH = 0
		}

		if clipW > 0 && clipH > 0 {
			c.PushClip(clipX, clipY, clipW, clipH)
			restore = func() { c.PopClip() }
		} else {
			restore = func() {}
		}

	case "circle", "ellipse":
		// Create elliptical clip using alpha masking
		rx := shape.RadiusX
		ry := shape.RadiusY
		cx := shape.CenterX // percentage (0-100)
		cy := shape.CenterY // percentage (0-100)

		// Default radius for circle
		if shape.Type == "circle" {
			if rx == 0 {
				rx = math.Min(float64(contentW), float64(contentH)) / 2
			}
			ry = rx // circle has equal radii
		}

		// If ry is still 0, use min dimension
		if ry == 0 {
			ry = math.Min(rx, float64(contentH))
		}

		// Convert percentages to pixels relative to content box
		centerX := contentX + int(cx*float64(contentW)/100)
		centerY := contentY + int(cy*float64(contentH)/100)

		// Create a temporary RGBA image for masking
		maskImg := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))

		// Draw white ellipse on black background (the mask)
		rxInt := int(rx)
		ryInt := int(ry)
		if rxInt <= 0 {
			rxInt = 1
		}
		if ryInt <= 0 {
			ryInt = 1
		}

		// Draw filled ellipse using mid-point ellipse algorithm
		drawEllipseFill(maskImg, centerX, centerY, rxInt, ryInt, color.RGBA{255, 255, 255, 255})

		// Create result image
		resultImg := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))

		// Apply mask: copy pixels only where mask is white
		bounds := maskImg.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				maskPixel := maskImg.RGBAAt(x, y)
				if maskPixel.R > 128 { // In ellipse
					resultImg.SetRGBA(x, y, c.Pixels.RGBAAt(x, y))
				} else { // Outside ellipse - transparent/black
					resultImg.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
				}
			}
		}

		// Copy result back to canvas pixels
		for y := 0; y < c.Height; y++ {
			for x := 0; x < c.Width; x++ {
				c.Pixels.Set(x, y, resultImg.RGBAAt(x, y))
			}
		}

		restore = func() {}

	case "polygon":
		// Polygon clipping is complex without GPU - use bounding box approximation
		if len(shape.Points) >= 3 {
			// Find bounding box
			minX, maxX := shape.Points[0][0], shape.Points[0][0]
			minY, maxY := shape.Points[0][1], shape.Points[0][1]
			for _, p := range shape.Points[1:] {
				if p[0] < minX {
					minX = p[0]
				}
				if p[0] > maxX {
					maxX = p[0]
				}
				if p[1] < minY {
					minY = p[1]
				}
				if p[1] > maxY {
					maxY = p[1]
				}
			}
			bx := contentX + int(minX*float64(contentW)/100)
			by := contentY + int(minY*float64(contentH)/100)
			bw := int((maxX - minX) * float64(contentW) / 100)
			bh := int((maxY - minY) * float64(contentH) / 100)

			if bw > 0 && bh > 0 {
				c.PushClip(bx, by, bw, bh)
				restore = func() { c.PopClip() }
			} else {
				restore = func() {}
			}
		} else {
			restore = func() {}
		}

	default:
		restore = func() {}
	}

	return restore, nil
}

// drawEllipseFill draws a filled ellipse at (cx, cy) with radii (rx, ry)
func drawEllipseFill(img *image.RGBA, cx, cy, rx, ry int, col color.Color) {
	if rx <= 0 || ry <= 0 {
		return
	}

	// Use mid-point ellipse algorithm for efficiency
	rx2 := float64(rx * rx)
	ry2 := float64(ry * ry)
	twoRx2 := 2 * rx2
	twoRy2 := 2 * ry2

	// Region 1
	x := 0
	y := ry
	px := 0.0
	py := twoRx2 * float64(y)

	// Plot initial points
	plotEllipsePoints(img, cx, cy, x, y, col)

	// Region 1 decision parameter
	p1 := ry2 - rx2*float64(ry) + 0.25*rx2

	for px < py {
		x++
		px += twoRy2
		if p1 < 0 {
			p1 += ry2 + px
		} else {
			y--
			py -= twoRx2
			p1 += ry2 + px - py
		}
		plotEllipsePoints(img, cx, cy, x, y, col)
	}

	// Region 2 decision parameter
	p2 := ry2*float64((x+1)*(x+1)) + rx2*float64((y-1)*(y-1)) - rx2*ry2

	for y > 0 {
		y--
		py -= twoRx2
		if p2 > 0 {
			p2 += rx2 - py
		} else {
			x++
			px += twoRy2
			p2 += rx2 - py + px
		}
		plotEllipsePoints(img, cx, cy, x, y, col)
	}
}

// plotEllipsePoints plots the four symmetric points of an ellipse
func plotEllipsePoints(img *image.RGBA, cx, cy, x, y int, col color.Color) {
	// Draw horizontal line at y offset to fill the ellipse
	for dx := -x; dx <= x; dx++ {
		px := cx + dx
		py1 := cy + y
		py2 := cy - y

		if px >= 0 && px < img.Bounds().Dx() && py1 >= 0 && py1 < img.Bounds().Dy() {
			img.Set(px, py1, col)
		}
		if px >= 0 && px < img.Bounds().Dx() && py2 >= 0 && py2 < img.Bounds().Dy() {
			img.Set(px, py2, col)
		}
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

// blendColors applies a CSS blend mode to blend a foreground color with a background color.
// The blendMode should be a CSS mix-blend-mode value like "multiply", "screen", etc.
func blendColors(fg, bg css.Color, blendMode string) css.Color {
	r1 := float64(fg.R) / 255.0
	g1 := float64(fg.G) / 255.0
	b1 := float64(fg.B) / 255.0
	r2 := float64(bg.R) / 255.0
	g2 := float64(bg.G) / 255.0
	b2 := float64(bg.B) / 255.0

	var r, g, b float64

	switch blendMode {
	case "multiply":
		r = r1 * r2
		g = g1 * g2
		b = b1 * b2
	case "screen":
		r = 1 - (1-r1)*(1-r2)
		g = 1 - (1-g1)*(1-g2)
		b = 1 - (1-b1)*(1-b2)
	case "overlay":
		if r2 < 0.5 {
			r = 2 * r1 * r2
		} else {
			r = 1 - 2*(1-r1)*(1-r2)
		}
		if g2 < 0.5 {
			g = 2 * g1 * g2
		} else {
			g = 1 - 2*(1-g1)*(1-g2)
		}
		if b2 < 0.5 {
			b = 2 * b1 * b2
		} else {
			b = 1 - 2*(1-b1)*(1-b2)
		}
	case "darken":
		r = math.Min(r1, r2)
		g = math.Min(g1, g2)
		b = math.Min(b1, b2)
	case "lighten":
		r = math.Max(r1, r2)
		g = math.Max(g1, g2)
		b = math.Max(b1, b2)
	case "color-dodge":
		if r1 >= 1 {
			r = 1
		} else {
			r = math.Min(1, r2/(1-r1))
		}
		if g1 >= 1 {
			g = 1
		} else {
			g = math.Min(1, g2/(1-g1))
		}
		if b1 >= 1 {
			b = 1
		} else {
			b = math.Min(1, b2/(1-b1))
		}
	case "color-burn":
		if r1 <= 0 {
			r = 0
		} else {
			r = math.Max(0, 1-(1-r2)/r1)
		}
		if g1 <= 0 {
			g = 0
		} else {
			g = math.Max(0, 1-(1-g2)/g1)
		}
		if b1 <= 0 {
			b = 0
		} else {
			b = math.Max(0, 1-(1-b2)/b1)
		}
	case "hard-light":
		if r1 < 0.5 {
			r = 2 * r1 * r2
		} else {
			r = 1 - 2*(1-r1)*(1-r2)
		}
		if g1 < 0.5 {
			g = 2 * g1 * g2
		} else {
			g = 1 - 2*(1-g1)*(1-g2)
		}
		if b1 < 0.5 {
			b = 2 * b1 * b2
		} else {
			b = 1 - 2*(1-b1)*(1-b2)
		}
	case "soft-light":
		if r1 < 0.5 {
			r = r2 - (1-2*r1)*r2*(1-r2)
		} else {
			if r2 < 0.25 {
				r = r2 + (2*r1-1)*(r2 - r2*((16*r2-12)*r2+3)/8)
			} else {
				r = r2 + (2*r1-1)*(math.Sqrt(r2)-r2)
			}
		}
		if g1 < 0.5 {
			g = g2 - (1-2*g1)*g2*(1-g2)
		} else {
			if g2 < 0.25 {
				g = g2 + (2*g1-1)*(g2 - g2*((16*g2-12)*g2+3)/8)
			} else {
				g = g2 + (2*g1-1)*(math.Sqrt(g2)-g2)
			}
		}
		if b1 < 0.5 {
			b = b2 - (1-2*b1)*b2*(1-b2)
		} else {
			if b2 < 0.25 {
				b = b2 + (2*b1-1)*(b2 - b2*((16*b2-12)*b2+3)/8)
			} else {
				b = b2 + (2*b1-1)*(math.Sqrt(b2)-b2)
			}
		}
	case "difference":
		if r1 > r2 {
			r = r1 - r2
		} else {
			r = r2 - r1
		}
		if g1 > g2 {
			g = g1 - g2
		} else {
			g = g2 - g1
		}
		if b1 > b2 {
			b = b1 - b2
		} else {
			b = b2 - b1
		}
	case "exclusion":
		r = r1 + r2 - 2*r1*r2
		g = g1 + g2 - 2*g1*g2
		b = b1 + b2 - 2*b1*b2
	case "hue", "saturation", "color", "luminosity":
		// Use temporary variables to avoid compiler complaints
		var h1Val, s1Val, l1Val, h2Val, s2Val, l2Val float64
		h1Val, s1Val, l1Val = rgbToHSL(r1, g1, b1)
		h2Val, s2Val, l2Val = rgbToHSL(r2, g2, b2)
		var rH, gH, bH float64
		switch blendMode {
		case "hue":
			rH, gH, bH = hslToRGB(h1Val, s2Val, l2Val)
		case "saturation":
			rH, gH, bH = hslToRGB(h2Val, s1Val, l2Val)
		case "color":
			rH, gH, bH = hslToRGB(h1Val, s2Val, l2Val)
		case "luminosity":
			rH, gH, bH = hslToRGB(h2Val, s2Val, l1Val)
		}
		r = rH
		g = gH
		b = bH
	case "normal", "":
		r = r1
		g = g1
		b = b1
	default:
		r = r1
		g = g1
		b = b1
	}

	// Clamp values
	r = math.Max(0, math.Min(1, r))
	g = math.Max(0, math.Min(1, g))
	b = math.Max(0, math.Min(1, b))

	return css.Color{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: fg.A,
	}
}

// rgbToHSL converts RGB (0-1) to HSL.
func rgbToHSL(r, g, b float64) (h, s, l float64) {
	max := r
	if g > max {
		max = g
	}
	if b > max {
		max = b
	}
	min := r
	if g < min {
		min = g
	}
	if b < min {
		min = b
	}

	l = (max + min) / 2

	if max == min {
		h = 0
		s = 0
		return
	}

	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}

	if max == r {
		h = (g - b) / d
		if g < b {
			h += 6
		}
	} else if max == g {
		h = (b-r)/d + 2
	} else {
		h = (r-g)/d + 4
	}
	h /= 6
	return
}

// hslToRGB converts HSL to RGB (0-1).
func hslToRGB(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		r = l
		g = l
		b = l
		return
	}

	q := l * (1 + s)
	if l > 0.5 {
		q = l + s - l*s
	}
	p := 2*l - q

	r = hueToRGB(p, q, h+1/3)
	g = hueToRGB(p, q, h)
	b = hueToRGB(p, q, h-1/3)
	return
}

// hueToRGB helper for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1/6 {
		return p + (q-p)*6*t
	}
	if t < 1/2 {
		return q
	}
	if t < 2/3 {
		return p + (q-p)*(2/3-t)*6
	}
	return p
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

// applyTransform applies a CSS transform to coordinates (x, y) relative to a box,
// given the box's width and height, and the transform origin (defaults to center).
// Returns the transformed coordinates (newX, newY).
func applyTransform(x, y float64, t css.Transform, boxW, boxH float64) (newX, newY float64) {
	// Default transform origin is center of the box
	originX := boxW / 2
	originY := boxH / 2

	// Handle "none" transform - no transformation, return original coordinates
	if t.Type == "none" {
		return x, y
	}

	// Translate to origin, apply transform, translate back
	// Step 1: Translate point to origin space
	px := x - originX
	py := y - originY

	var tx, ty float64

	switch t.Type {
	case "rotate":
		// Rotate around origin
		angleRad := t.Rotate * math.Pi / 180
		cosA := math.Cos(angleRad)
		sinA := math.Sin(angleRad)
		tx = px*cosA - py*sinA
		ty = px*sinA + py*cosA

	case "scale":
		tx = px * t.ScaleX
		ty = py * t.ScaleY

	case "translate":
		tx = px + t.TranslateX
		ty = py + t.TranslateY

	case "skew":
		// skewX shifts x based on y, skewY shifts y based on x
		skewXRad := t.SkewX * math.Pi / 180
		skewYRad := t.SkewY * math.Pi / 180
		tx = px + py*math.Tan(skewXRad)
		ty = py + px*math.Tan(skewYRad)

	case "matrix":
		// matrix(a, b, c, d, e, f) = [a c e] [x]
		//                           [b d f] [y]
		//                           [0 0 1] [1]
		m := t.Matrix
		tx = m[0]*px + m[2]*py + m[4]
		ty = m[1]*px + m[3]*py + m[5]

	default:
		tx, ty = px, py
	}

	// Step 3: Translate back from origin space
	return tx + originX, ty + originY
}

// FillRectTransformed fills a rectangle with a color, applying a CSS transform.
// The transform is applied around the box's center (transform-origin: center).
func (c *Canvas) FillRectTransformed(x, y, w, h int, col color.Color, t css.Transform) {
	if w <= 0 || h <= 0 {
		return
	}
	boxW := float64(w)
	boxH := float64(h)

	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			// Apply transform to this pixel's position relative to box
			tx, ty := applyTransform(float64(px), float64(py), t, boxW, boxH)
			// Convert back to integer pixel coordinates
			drawX := int(tx) + x
			drawY := int(ty) + y
			c.SetPixel(drawX, drawY, col)
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

	// CSS Transform
	transform := css.ParseTransform(box.Style["transform"])
	hasTransform := transform.Type != "none"

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
		if hasTransform {
			// For transformed elements, draw a transformed rectangle
			c.FillRectTransformed(marginX, marginY, marginW, marginH, bgColor, transform)
		} else {
			c.DrawRoundedRect(marginX, marginY, marginW, marginH, cornerR, bgColor)
		}
		// Padding area (smaller rounded rect inside)
		// Reduce corner radius for inner layers (approximation)
		innerR := cornerR / 2
		if paddingW > 0 && paddingH > 0 && innerR > 0 {
			if hasTransform {
				c.FillRectTransformed(paddingX, paddingY, paddingW, paddingH, bgColor, transform)
			} else {
				c.DrawRoundedRect(paddingX, paddingY, paddingW, paddingH, innerR, bgColor)
			}
		}
	} else {
		if hasTransform {
			c.FillRectTransformed(marginX, marginY, marginW, marginH, bgColor, transform)
			if paddingW > 0 && paddingH > 0 {
				c.FillRectTransformed(paddingX, paddingY, paddingW, paddingH, bgColor, transform)
			}
		} else {
			c.FillRect(marginX, marginY, marginW, marginH, bgColor)
			if paddingW > 0 && paddingH > 0 {
				c.FillRect(paddingX, paddingY, paddingW, paddingH, bgColor)
			}
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
		// Multiple box-shadows are comma-separated: box-shadow: 1px 1px black, 2px 2px blue
		// Split by comma but respect function parentheses
		shadowParts := strings.Split(boxShadow, ",")
		for _, shadow := range shadowParts {
			parts := strings.Fields(shadow)
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
		// Outline offset (gap between border and outline)
		outlineOffset := int(css.ParseLength(box.Style["outline-offset"]).Value)
		// Outline is drawn around the outer edge of the border area, offset by outline-offset
		ox := contentX - outlineWidth - outlineOffset
		oy := contentY - outlineWidth - outlineOffset
		ow := contentW + outlineWidth*2 + outlineOffset*2
		oh := contentH + outlineWidth*2 + outlineOffset*2
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

	// Handle clip-path CSS property - apply proper clip shapes
	if clipPath := box.Style["clip-path"]; clipPath != "" && clipPath != "none" {
		restoreClip, err := c.applyClipPath(box)
		if err == nil {
			defer restoreClip()
		}
	}

	// Check visibility - hidden boxes paint background/border/padding but not content
	visibility, _ := box.Style["visibility"]
	isHidden := visibility == "hidden"

	// Apply backdrop-filter for fixed/absolute positioned elements
	// This filters the pixels BEHIND the element before drawing it
	if !isHidden {
		c.ApplyBackdropFilter(box)
	}

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

	// Iframe content
	if !isHidden && box.Type == layout.IframeBox && box.Node != nil {
		c.DrawIframe(box)
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

	// Fieldset element
	if !isHidden && box.Type == layout.FieldsetBox && box.Node != nil {
		c.DrawFieldset(box)
	}

	// Legend element
	if !isHidden && box.Type == layout.LegendBox && box.Node != nil {
		c.DrawLegend(box)
	}

	// Details element
	if !isHidden && box.Type == layout.DetailsBox && box.Node != nil {
		c.DrawDetails(box)
	}

	// Summary element
	if !isHidden && box.Type == layout.SummaryBox && box.Node != nil {
		c.DrawSummary(box)
	}

	// Slot element (display:contents - renders nothing but children are projected)
	if !isHidden && box.Type == layout.SlotBox && box.Node != nil {
		c.DrawSlot(box)
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

	// Apply CSS filter effect to the content area after all content is drawn
	filterStr := box.Style["filter"]
	if filterStr != "" && filterStr != "none" {
		// Filter applies to the content box (inside padding)
		filterX := contentX + paddingLeft
		filterY := contentY + paddingTop
		filterW := contentW - paddingLeft - paddingRight
		filterH := contentH - paddingTop - paddingBottom
		if filterW <= 0 {
			filterW = contentW
		}
		if filterH <= 0 {
			filterH = contentH
		}
		if filterW > 0 && filterH > 0 {
			c.ApplyFilter(box, filterX, filterY, filterW, filterH)
		}
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

// DrawText renders text content with proper font rendering.
func (c *Canvas) DrawText(box *layout.Box) {
	if box.Node == nil || box.Node.Type != 2 { // NodeText
		return
	}

	text := box.Node.Data

	// Resolve font-size to pixels, handling em units
	// Walk up to ancestor block to find parent font size for em resolution
	var getAncestorFontSize func(b *layout.Box) float64
	getAncestorFontSize = func(b *layout.Box) float64 {
		cur := b.Parent
		for cur != nil {
			if cur.Type == layout.BlockBox || cur.Type == layout.FlexBox || cur.Type == layout.PositionedBox {
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

	// Font family
	fontFamily := box.Style["font-family"]
	if fontFamily == "" {
		fontFamily = "sans-serif"
	}

	// Font style (italic)
	fontStyle := box.Style["font-style"]

	// Letter spacing
	letterSpacing := css.ParseLength(box.Style["letter-spacing"]).Value

	// Word spacing
	wordSpacing := css.ParseLength(box.Style["word-spacing"]).Value

	// Text indent (applied on first line only)
	textIndent := css.ParseLength(box.Style["text-indent"]).Value

	// Get font for this text
	f, advWidth := c.GetFont(fontFamily, fontSize)
	if advWidth == 0 {
		advWidth = fontSize * 0.6
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
	overflow := box.Style["overflow"]
	overflowX := box.Style["overflow-x"]
	textOverflow := box.Style["text-overflow"]
	showEllipsis := (overflow == "hidden" || overflowX == "hidden") && textOverflow == "ellipsis"

	// Parse text-shadow (multiple comma-separated shadows)
	type textShadow struct {
		x     int
		y     int
		color css.Color
	}
	var textShadows []textShadow
	if shadow := box.Style["text-shadow"]; shadow != "" && shadow != "none" {
		shadowParts := splitShadowParts(shadow)
		for _, shadow := range shadowParts {
			parts := strings.Fields(shadow)
			if len(parts) >= 2 {
				ts := textShadow{
					x:     int(css.ParseLength(parts[0]).Value),
					y:     int(css.ParseLength(parts[1]).Value),
					color: textColor,
				}
				if len(parts) >= 4 {
					ts.color = css.ParseColor(parts[3])
				} else if len(parts) >= 3 {
					// If 3 parts and third isn't a color, use it as color
					if pc := css.ParseColor(parts[2]); pc.A > 0 || parts[2] == "transparent" {
						ts.color = pc
					}
				}
				textShadows = append(textShadows, ts)
			}
		}
	}

	// For text-overflow ellipsis, calculate the maximum chars that fit
	maxChars := len(text)
	if showEllipsis {
		availWidth := float64(containerWidth)
		if isFirstLine {
			availWidth -= textIndent
		}
		ellipsisWidth := advWidth * italicSlant * 3 // "…" is roughly 3 chars wide
		availWidth -= ellipsisWidth
		if availWidth < 0 {
			availWidth = 0
		}
		maxChars = int(availWidth / (advWidth * italicSlant))
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
			ellipsis := "…"
			for _, ech := range ellipsis {
				// Draw shadows first
				for _, ts := range textShadows {
					c.DrawGlyph(f, fontSize, currentX+float64(ts.x), currentY+float64(ts.y), ts.color, ech)
				}
				c.DrawGlyph(f, fontSize, currentX, currentY, textColor, ech)
				// Advance cursor
				face := truetype.NewFace(f, &truetype.Options{Size: fontSize, DPI: 72})
				dot := fixed.Point26_6{X: fixed.Int26_6(currentX * 64), Y: fixed.Int26_6(currentY * 64)}
				_, _, _, adv, _ := face.Glyph(dot, ech)
				currentX += float64(adv) / 64.0 * italicSlant
			}
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
			tabSize := 8 // default
			if ts := css.ParseLength(box.Style["tab-size"]).Value; ts > 0 {
				tabSize = int(ts)
			}
			tabWidth := float64(tabSize) * advWidth * italicSlant
			tabStop := math.Ceil((currentX - float64(x)) / tabWidth) * tabWidth
			currentX = float64(x) + tabStop
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

		// Get the glyph advance width for this character
		face := truetype.NewFace(f, &truetype.Options{Size: fontSize, DPI: 72})
		dot := fixed.Point26_6{X: fixed.Int26_6(currentX * 64), Y: fixed.Int26_6(currentY * 64)}
		_, _, _, adv, ok := face.Glyph(dot, ch)
		cw := advWidth * italicSlant
		if ok {
			cw = float64(adv)/64.0*italicSlant + extraLetter
		}

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
			// Recalculate glyph advance at new line position
			dot = fixed.Point26_6{X: fixed.Int26_6(currentX * 64), Y: fixed.Int26_6(currentY * 64)}
			_, _, _, adv, ok = face.Glyph(dot, ch)
			if ok {
				cw = float64(adv)/64.0*italicSlant + extraLetter
			}
		}

		// Draw text-shadow first (behind the text) - multiple shadows
		for _, ts := range textShadows {
			if isEmoji(ch) {
				c.drawEmojiPlaceholder(currentX+float64(ts.x), currentY+float64(ts.y), fontSize*0.8, ch, ts.color)
			} else {
				c.DrawGlyph(f, fontSize, currentX+float64(ts.x), currentY+float64(ts.y), ts.color, ch)
			}
		}

		// Draw the actual glyph with proper font rasterization
		// For emoji characters, use colored placeholder instead of font glyph
		if isEmoji(ch) {
			c.drawEmojiPlaceholder(currentX, currentY, fontSize*0.8, ch, textColor)
		} else {
			c.DrawGlyph(f, fontSize, currentX, currentY, textColor, ch)
		}

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
			decorationStyle := box.Style["text-decoration-style"]
			underlineOffset := css.ParseLength(box.Style["text-underline-offset"]).Value
			if underlineOffset == 0 {
				underlineOffset = 1 // default offset
			}

			// Underline: 1px line at baseline (below text)
			if decorationLine == "underline" || strings.Contains(decorationLine, "underline") {
				decY := int(currentY + fontSize + underlineOffset)
				decH := 1
				if dt := css.ParseLength(box.Style["text-decoration-thickness"]).Value; dt > 0 {
					decH = int(dt)
				}
				// Draw dotted/dashed underline by skipping every other pixel
				if decorationStyle == "dotted" {
					for dx := 0; dx < int(cw); dx += 2 {
						c.FillRect(int(currentX)+dx, decY, 1, decH, decorationColor)
					}
				} else if decorationStyle == "dashed" {
					for dx := 0; dx < int(cw); dx += 4 {
						c.FillRect(int(currentX)+dx, decY, 2, decH, decorationColor)
					}
				} else {
					c.FillRect(int(currentX), decY, int(cw), decH, decorationColor)
				}
			}
			// Overline: 1px line above text
			if decorationLine == "overline" || strings.Contains(decorationLine, "overline") {
				decY := int(currentY)
				decH := 1
				if dt := css.ParseLength(box.Style["text-decoration-thickness"]).Value; dt > 0 {
					decH = int(dt)
				}
				if decorationStyle == "dotted" {
					for dx := 0; dx < int(cw); dx += 2 {
						c.FillRect(int(currentX)+dx, decY, 1, decH, decorationColor)
					}
				} else if decorationStyle == "dashed" {
					for dx := 0; dx < int(cw); dx += 4 {
						c.FillRect(int(currentX)+dx, decY, 2, decH, decorationColor)
					}
				} else {
					c.FillRect(int(currentX), decY, int(cw), decH, decorationColor)
				}
			}
			// Line-through: 1px line at middle of text
			if decorationLine == "line-through" || strings.Contains(decorationLine, "line-through") {
				decY := int(currentY + fontSize/2)
				decH := 1
				if dt := css.ParseLength(box.Style["text-decoration-thickness"]).Value; dt > 0 {
					decH = int(dt)
				}
				if decorationStyle == "dotted" {
					for dx := 0; dx < int(cw); dx += 2 {
						c.FillRect(int(currentX)+dx, decY, 1, decH, decorationColor)
					}
				} else if decorationStyle == "dashed" {
					for dx := 0; dx < int(cw); dx += 4 {
						c.FillRect(int(currentX)+dx, decY, 2, decH, decorationColor)
					}
				} else {
					c.FillRect(int(currentX), decY, int(cw), decH, decorationColor)
				}
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

// DrawIframe renders an iframe with a placeholder for cross-origin iframes.
// For same-origin iframes, it could recursively render the nested content.
func (c *Canvas) DrawIframe(box *layout.Box) {
	if box.Node == nil || box.Node.Type != html.NodeElement {
		return
	}

	src := box.Node.GetAttribute("src")
	title := box.Node.GetAttribute("title")
	width := box.Node.GetAttribute("width")
	height := box.Node.GetAttribute("height")

	// Get dimensions from box (which has CSS/computed dimensions)
	w := int(box.ContentW)
	h := int(box.ContentH)
	if w <= 0 {
		if width != "" {
			if wVal, err := strconv.Atoi(width); err == nil {
				w = wVal
			}
		}
		if w <= 0 {
			w = 300 // browser default
		}
	}
	if h <= 0 {
		if height != "" {
			if hVal, err := strconv.Atoi(height); err == nil {
				h = hVal
			}
		}
		if h <= 0 {
			h = 150 // browser default
		}
	}

	x := int(box.ContentX)
	y := int(box.ContentY)

	// Draw placeholder background (light gray)
	c.FillRect(x, y, w, h, css.Color{R: 240, G: 240, B: 240, A: 255})

	// Draw border
	c.DrawBorder(x, y, w, h, 1, css.Color{R: 180, G: 180, B: 180, A: 255})

	// Determine label text
	label := "iframe"
	if title != "" {
		label = title
	} else if src != "" {
		// Truncate long URLs
		if len(src) > 30 {
			label = src[:27] + "..."
		} else {
			label = src
		}
	}

	// Draw lock icon and label text using simple rectangles
	// Draw a simple lock icon (🔒 placeholder as small rectangle)
	lockSize := 8
	lockX := x + 8
	lockY := y + h/2 - lockSize/2
	c.FillRect(lockX, lockY, lockSize, lockSize, css.Color{R: 100, G: 100, B: 100, A: 255})

	// Draw text label
	fontSize := 12.0
	charWidth := fontSize * 0.6
	textX := lockX + lockSize + 8
	maxTextWidth := w - int(textX-x) - 10
	maxChars := int(float64(maxTextWidth) / charWidth)
	if maxChars > len(label) {
		maxChars = len(label)
	}
	if maxChars > 0 {
		textColor := css.Color{R: 80, G: 80, B: 80, A: 255}
		for i := 0; i < maxChars; i++ {
			charX := textX + int(float64(i)*charWidth)
			c.FillRect(charX, y+h/2-int(fontSize/2), int(charWidth), int(fontSize), textColor)
		}
	}
}

// loadImage fetches and decodes an image from a URL.
// Returns the image or an error.
func (c *Canvas) loadImage(src string) (image.Image, error) {
	resp, err := fetch.Fetch(src, "", 30, nil)
	if err != nil || resp == nil || resp.StatusCode != 200 {
		return nil, err
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
			return nil, err
		}
		return img, nil
	case "image/jpeg", "image/jpg":
		img, err := jpeg.Decode(reader)
		if err != nil {
			return nil, err
		}
		return img, nil
	case "image/gif":
		img, err := gif.Decode(reader)
		if err != nil {
			return nil, err
		}
		return img, nil
	case "image/webp":
		// WebP decode not in stdlib, return nil for now
		return nil, nil
	default:
		// Try PNG as fallback
		reader.Seek(0, 0)
		if img, err := png.Decode(reader); err == nil {
			return img, nil
		}
		return nil, nil
	}
}

// loadBackgroundImage fetches and decodes a background image URL.
// Returns the image or nil if loading failed.
func (c *Canvas) loadBackgroundImage(url string) image.Image {
	resp, err := fetch.Fetch(url, "", 30, nil)
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
		// space: tile to fill, but distribute space evenly between images
		// Count how many images fit
		countX := 1
		if imgDrawW < w {
			countX = (w + imgDrawW - 1) / imgDrawW
		}
		countY := 1
		if imgDrawH < h {
			countY = (h + imgDrawH - 1) / imgDrawH
		}
		// Compute total space to distribute
		totalSpaceX := w - countX*imgDrawW
		totalSpaceY := h - countY*imgDrawH
		spaceX := 0
		spaceY := 0
		if countX > 1 {
			spaceX = totalSpaceX / (countX + 1)
		}
		if countY > 1 {
			spaceY = totalSpaceY / (countY + 1)
		}
		// Draw images with even spacing
		for py := y + spaceY; py < y+h; py += imgDrawH + spaceY {
			for px := x + spaceX; px < x+w; px += imgDrawW + spaceX {
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
	case "round":
		// round: tile and scale image to fill space (no gaps)
		countX := 1
		if imgDrawW < w {
			countX = (w + imgDrawW - 1) / imgDrawW
		}
		countY := 1
		if imgDrawH < h {
			countY = (h + imgDrawH - 1) / imgDrawH
		}
		// Scale images to fill gaps
		newImgW := w / countX
		newImgH := h / countY
		if newImgW < 1 {
			newImgW = 1
		}
		if newImgH < 1 {
			newImgH = 1
		}
		for py := y; py < y+h; py += newImgH {
			for px := x; px < x+w; px += newImgW {
				drawX := px
				drawY := py
				drawW := newImgW
				drawH := newImgH
				srcX := 0
				srcY := 0
				if px < x {
					srcX = x - px
					drawW = newImgW - srcX
					drawX = x
				}
				if py < y {
					srcY = y - py
					drawH = newImgH - srcY
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

	// Use marker-specific color if set (from li::marker { color: ... }), otherwise use element color
	textColor := css.ParseColor(box.Style["_marker_color"])
	if textColor.A == 0 {
		textColor = css.ParseColor(box.Style["color"])
	}

	// Get marker-specific font-size if set, otherwise use element font-size
	markerFontSize := 14.0
	if mfs := box.Style["_marker_font_size"]; mfs != "" {
		if fs, err := strconv.ParseFloat(mfs, 64); err == nil {
			markerFontSize = fs
		}
	}

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
		// Number — draw simple number representation using marker-specific font-size
		fontSize := markerFontSize
		charWidth := fontSize * 0.6
		c.FillRect(x+2, y+h/4, int(charWidth*3), int(fontSize), textColor)
	case "lower-alpha", "upper-alpha":
		// Letter using marker-specific font-size
		fontSize := markerFontSize
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

	// Draw caret (text cursor) with caret-color
	caretColor := c.parseCaretColor(box.Style["caret-color"], box.Style["color"])
	if caretColor.A > 0 {
		caretX := x + 4 // Position after left padding
		caretY := y + 2
		caretHeight := h - 4
		caretWidth := 2
		c.FillRect(caretX, caretY, caretWidth, caretHeight, caretColor)
	}
}

// parseCaretColor parses the caret-color CSS property.
// Returns the computed color or falls back to the color property or default black.
func (c *Canvas) parseCaretColor(caretColorStr, colorStr string) css.Color {
	caretColorStr = strings.TrimSpace(caretColorStr)
	if caretColorStr == "" || caretColorStr == "auto" {
		// auto falls back to the element's color property
		colorVal := css.ParseColor(colorStr)
		if colorVal.A > 0 {
			return colorVal
		}
		return css.Color{R: 0, G: 0, B: 0, A: 255}
	}
	if caretColorStr == "transparent" {
		return css.Color{R: 0, G: 0, B: 0, A: 0}
	}
	return css.ParseColor(caretColorStr)
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

	// Draw caret (text cursor) with caret-color
	caretColor := c.parseCaretColor(box.Style["caret-color"], box.Style["color"])
	if caretColor.A > 0 {
		caretX := x + 4 // Position after left padding
		caretY := y + int(fontSize) + 2
		caretHeight := int(fontSize)
		caretWidth := 2
		c.FillRect(caretX, caretY, caretWidth, caretHeight, caretColor)
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

// DrawDetails renders a <details> element with a disclosure triangle.
// The summary is always visible; the rest of the content is toggled by open attribute.
func (c *Canvas) DrawDetails(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 {
		w = 300
	}
	if h <= 0 {
		h = 100
	}

	// Determine open state
	isOpen := false
	if box.Node != nil {
		if openStr := box.Node.GetAttribute("open"); openStr != "" {
			isOpen = true
		}
	}

	// Draw background
	bgColor := css.ParseColor(box.Style["background-color"])
	if bgColor.A == 0 {
		bgColor = css.Color{R: 255, G: 255, B: 255, A: 255}
	}
	c.FillRect(x, y, w, h, bgColor)

	// Draw border
	borderColor := css.Color{R: 180, G: 180, B: 180, A: 255}
	c.DrawBorder(x, y, w, h, 1, borderColor)

	// Draw disclosure triangle indicator (▶ collapsed / ▼ expanded)
	triangleColor := css.Color{R: 80, G: 80, B: 80, A: 255}
	triSize := 12
	triX := x + 8
	triY := y + 10

	if isOpen {
		// Draw ▼ (expanded) using filled triangles
		for i := 0; i < triSize; i++ {
			c.FillRect(triX+i/2, triY+i, triSize-i, 1, triangleColor)
		}
	} else {
		// Draw ▶ (collapsed) using filled triangles
		for i := 0; i < triSize; i++ {
			c.FillRect(triX+i, triY+i/2, 1, triSize-i, triangleColor)
		}
	}
}

// DrawSummary renders a <summary> element inside a <details> element.
// It acts as the label/header that users click to toggle the details.
func (c *Canvas) DrawSummary(box *layout.Box) {
	// Summary is rendered as part of the details element.
	// Draw nothing special here - the text content will be drawn by the normal text rendering.
	// The key visual indicator (▶/▼) is drawn by DrawDetails.
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 {
		w = 300
	}
	if h <= 0 {
		h = 30
	}

	// Draw subtle background to distinguish summary from content
	bgColor := css.Color{R: 245, G: 245, B: 245, A: 255}
	c.FillRect(x, y, w, h, bgColor)
}

// DrawSlot renders a <slot> element for web components.
// Slots project content from the light DOM into the shadow DOM.
func (c *Canvas) DrawSlot(box *layout.Box) {
	// Slots are display:contents - they don't render themselves.
	// The content slotted into them is rendered as normal children.
	// This is a no-op for the slot itself.
}

// DrawTemplate renders a <template> element.
// Templates are parsed but their content is inert and not rendered.
func (c *Canvas) DrawTemplate(box *layout.Box) {
	// Template content is not rendered - it's stored as inert document fragment.
	// This is a no-op for the template element itself.
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

// splitShadowParts splits a CSS shadow value by comma, respecting function arguments.
// E.g., "1px 1px black, 2px 2px blue" -> ["1px 1px black", "2px 2px blue"]
// E.g., "drop-shadow(1px 1px black) drop-shadow(2px 2px blue)" -> ["drop-shadow(1px 1px black)", "drop-shadow(2px 2px blue)"]
func splitShadowParts(shadow string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	for i := 0; i < len(shadow); i++ {
		c := shadow[i]
		if c == '(' {
			depth++
			current.WriteByte(c)
		} else if c == ')' {
			depth--
			current.WriteByte(c)
		} else if c == ',' && depth == 0 {
			part := strings.TrimSpace(current.String())
			if part != "" {
				parts = append(parts, part)
			}
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}
	part := strings.TrimSpace(current.String())
	if part != "" {
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		parts = []string{shadow}
	}
	return parts
}

// DrawFieldset renders a <fieldset> element with a border.
// Fieldset draws a border around its content, with the legend positioned
// at the top-left corner overlapping the border.
func (c *Canvas) DrawFieldset(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 {
		w = 400
	}
	if h <= 0 {
		h = 100
	}

	// Draw fieldset border
	borderColor := css.ParseColor(box.Style["border-color"])
	if borderColor.A == 0 {
		borderColor = css.Color{R: 128, G: 128, B: 128, A: 255}
	}
	borderWidth := int(css.ParseLength(box.Style["border-width"]).Value)
	if borderWidth == 0 {
		borderWidth = 2
	}

	// Background
	bgColor := css.ParseColor(box.Style["background"])
	if bgColor.A == 0 {
		bgColor = css.Color{R: 255, G: 255, B: 255, A: 255}
	}
	c.FillRect(x, y, w, h, bgColor)

	// Draw border around fieldset
	c.DrawBorder(x, y, w, h, borderWidth, borderColor)

	// Find and position legend
	for _, child := range box.Children {
		if child.Type == layout.LegendBox {
			// Legend is positioned at top-left, overlapping the border
			// Adjust legend position to sit on the top border
			legendH := int(child.ContentH)
			child.ContentY = float64(y - legendH/2)
			if child.ContentY < float64(y) {
				child.ContentY = float64(y)
			}
		}
	}
}

// DrawLegend renders a <legend> element.
// Legend is rendered as a block-level element with padding and a background.
func (c *Canvas) DrawLegend(box *layout.Box) {
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 {
		w = 100
	}
	if h <= 0 {
		h = 24
	}

	// Background for legend
	bgColor := css.ParseColor(box.Style["background"])
	if bgColor.A == 0 {
		bgColor = css.Color{R: 240, G: 240, B: 240, A: 255}
	}

	// Padding
	paddingLeft := int(css.ParseLength(box.Style["padding-left"]).Value)
	paddingRight := int(css.ParseLength(box.Style["padding-right"]).Value)
	paddingTop := int(css.ParseLength(box.Style["padding-top"]).Value)
	paddingBottom := int(css.ParseLength(box.Style["padding-bottom"]).Value)

	// Border
	borderColor := css.ParseColor(box.Style["border-color"])
	if borderColor.A == 0 {
		borderColor = css.Color{R: 128, G: 128, B: 128, A: 255}
	}
	borderWidth := int(css.ParseLength(box.Style["border-width"]).Value)

	// Draw legend background
	legendX := x - paddingLeft
	legendY := y - paddingTop/2
	legendW := w + paddingLeft + paddingRight
	legendH := h + paddingTop + paddingBottom

	c.FillRect(legendX, legendY, legendW, legendH, bgColor)

	// Draw legend border
	if borderWidth > 0 {
		c.DrawBorder(legendX, legendY, legendW, legendH, borderWidth, borderColor)
	}
}

// ApplyFilter applies CSS filter effects to a region of the canvas.
// It parses the filter string from box.Style["filter"] and applies the effects.
func (c *Canvas) ApplyFilter(box *layout.Box, x, y, w, h int) {
	filterStr := box.Style["filter"]
	if filterStr == "" || filterStr == "none" {
		return
	}

	effects, err := css.ParseFilter(filterStr)
	if err != nil || len(effects) == 0 {
		return
	}

	// Create a temporary buffer to hold the original pixels
	// We need this because filters like blur need to read original pixel values
	// while writing modified values to the same region
	srcPixels := image.NewRGBA(image.Rect(0, 0, w, h))
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			srcX := x + dx
			srcY := y + dy
			if srcX >= 0 && srcX < c.Width && srcY >= 0 && srcY < c.Height {
				srcPixels.Set(dx, dy, c.Pixels.At(srcX, srcY))
			}
		}
	}

	// Apply each filter effect in order
	for _, effect := range effects {
		switch effect.Type {
		case "blur":
			radius := int(effect.Amount)
			if radius > 10 {
				radius = 10 // Cap at 10px for performance
			}
			if radius > 0 {
				c.applyBlur(srcPixels, 0, 0, w, h, radius)
			}
		case "brightness":
			c.applyBrightness(srcPixels, 0, 0, w, h, effect.Amount)
		case "contrast":
			c.applyContrast(srcPixels, 0, 0, w, h, effect.Amount)
		case "grayscale":
			c.applyGrayscale(srcPixels, 0, 0, w, h, effect.Amount)
		case "sepia":
			c.applySepia(srcPixels, 0, 0, w, h, effect.Amount)
		case "hue-rotate":
			// hue-rotate not implemented yet
		case "drop-shadow":
			// drop-shadow is typically applied differently, not as a pixel filter
			// It's usually drawn as a shadow behind the element
		}
	}

	// Copy filtered pixels back to canvas
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			dstX := x + dx
			dstY := y + dy
			if dstX >= 0 && dstX < c.Width && dstY >= 0 && dstY < c.Height {
				c.Pixels.Set(dstX, dstY, srcPixels.At(dx, dy))
			}
		}
	}
}

// ApplyBackdropFilter applies a CSS backdrop-filter effect to the area behind a fixed/absolute positioned element.
// It captures the pixels behind the element, applies the filter, and composites them back.
// This creates effects like blur-behind for modal dialogs.
func (c *Canvas) ApplyBackdropFilter(box *layout.Box) {
	backdropFilter := box.Style["backdrop-filter"]
	if backdropFilter == "" || backdropFilter == "none" {
		return
	}

	// Check if element is fixed or absolute positioned
	position := box.Style["position"]
	if position != "fixed" && position != "absolute" {
		return
	}

	// Get element dimensions
	x := int(box.ContentX)
	y := int(box.ContentY)
	w := int(box.ContentW)
	h := int(box.ContentH)

	if w <= 0 || h <= 0 {
		return
	}

	effects, err := css.ParseFilter(backdropFilter)
	if err != nil || len(effects) == 0 {
		return
	}

	// Create a buffer to hold the original pixels behind the element
	srcPixels := image.NewRGBA(image.Rect(0, 0, w, h))
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			srcX := x + dx
			srcY := y + dy
			if srcX >= 0 && srcX < c.Width && srcY >= 0 && srcY < c.Height {
				srcPixels.Set(dx, dy, c.Pixels.At(srcX, srcY))
			}
		}
	}

	// Apply each filter effect in order (same as ApplyFilter)
	for _, effect := range effects {
		switch effect.Type {
		case "blur":
			radius := int(effect.Amount)
			if radius > 10 {
				radius = 10 // Cap at 10px for performance
			}
			if radius > 0 {
				c.applyBlur(srcPixels, 0, 0, w, h, radius)
			}
		case "brightness":
			c.applyBrightness(srcPixels, 0, 0, w, h, effect.Amount)
		case "contrast":
			c.applyContrast(srcPixels, 0, 0, w, h, effect.Amount)
		case "grayscale":
			c.applyGrayscale(srcPixels, 0, 0, w, h, effect.Amount)
		case "sepia":
			c.applySepia(srcPixels, 0, 0, w, h, effect.Amount)
		case "hue-rotate":
			// hue-rotate not implemented yet
		case "drop-shadow":
			// drop-shadow is typically applied differently
		}
	}

	// Copy filtered pixels back to canvas (behind the element)
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			dstX := x + dx
			dstY := y + dy
			if dstX >= 0 && dstX < c.Width && dstY >= 0 && dstY < c.Height {
				c.Pixels.Set(dstX, dstY, srcPixels.At(dx, dy))
			}
		}
	}
}

// applyBlur applies a box blur filter to the specified region.
// It uses a simple 3x3 or 5x5 kernel based on radius.
func (c *Canvas) applyBlur(pixels *image.RGBA, x, y, w, h, radius int) {
	if w <= 0 || h <= 0 {
		return
	}

	// Create output buffer
	result := image.NewRGBA(pixels.Bounds())

	// Determine kernel size based on radius
	kernelSize := 3
	if radius > 2 {
		kernelSize = 5
	}
	offset := kernelSize / 2

	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			var rSum, gSum, bSum, aSum float64
			count := 0.0

			for ky := -offset; ky <= offset; ky++ {
				for kx := -offset; kx <= offset; kx++ {
					srcX := px + kx
					srcY := py + ky

					if srcX >= x && srcX < x+w && srcY >= y && srcY < y+h {
						r, g, b, a := pixels.At(srcX, srcY).RGBA()
						rSum += float64(r >> 8)
						gSum += float64(g >> 8)
						bSum += float64(b >> 8)
						aSum += float64(a >> 8)
						count++
					}
				}
			}

			if count > 0 {
				newR := uint8(rSum / count)
				newG := uint8(gSum / count)
				newB := uint8(bSum / count)
				newA := uint8(aSum / count)
				result.Set(px, py, color.RGBA{newR, newG, newB, newA})
			}
		}
	}

	// Copy result back
	copy(pixels.Pix, result.Pix)
}

// applyBrightness applies a brightness filter.
// amount > 1 brightens, amount < 1 darkens, amount = 1 is normal.
func (c *Canvas) applyBrightness(pixels *image.RGBA, x, y, w, h int, amount float64) {
	if w <= 0 || h <= 0 || amount <= 0 {
		return
	}

	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			r, g, b, a := pixels.At(px, py).RGBA()
			newR := float64(r>>8) * amount
			newG := float64(g>>8) * amount
			newB := float64(b>>8) * amount

			if newR > 255 {
				newR = 255
			}
			if newG > 255 {
				newG = 255
			}
			if newB > 255 {
				newB = 255
			}

			pixels.Set(px, py, color.RGBA{uint8(newR), uint8(newG), uint8(newB), uint8(a >> 8)})
		}
	}
}

// applyContrast adjusts the contrast of the image.
// amount = 1 is normal, amount > 1 increases contrast, amount = 0 is gray.
func (c *Canvas) applyContrast(pixels *image.RGBA, x, y, w, h int, amount float64) {
	if w <= 0 || h <= 0 || amount < 0 {
		return
	}

	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			r, g, b, a := pixels.At(px, py).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			// Apply contrast formula: output = (input - 0.5) * contrast + 0.5
			newR := (float64(r8)/255.0 - 0.5) * amount + 0.5
			newG := (float64(g8)/255.0 - 0.5) * amount + 0.5
			newB := (float64(b8)/255.0 - 0.5) * amount + 0.5

			// Clamp to 0-1 range and convert back to 0-255
			if newR < 0 {
				newR = 0
			}
			if newR > 1 {
				newR = 1
			}
			if newG < 0 {
				newG = 0
			}
			if newG > 1 {
				newG = 1
			}
			if newB < 0 {
				newB = 0
			}
			if newB > 1 {
				newB = 1
			}

			pixels.Set(px, py, color.RGBA{uint8(newR * 255), uint8(newG * 255), uint8(newB * 255), uint8(a >> 8)})
		}
	}
}

// applyGrayscale converts the image to grayscale.
// amount = 1 is fully grayscale, amount = 0 is unchanged.
func (c *Canvas) applyGrayscale(pixels *image.RGBA, x, y, w, h int, amount float64) {
	if w <= 0 || h <= 0 || amount < 0 || amount > 1 {
		return
	}

	if amount >= 1 {
		// Fast path: fully grayscale
		for py := y; py < y+h; py++ {
			for px := x; px < x+w; px++ {
				r, g, b, a := pixels.At(px, py).RGBA()
				// ITU-R BT.709 luminance coefficients
				luminance := 0.2126*float64(r>>8) + 0.7152*float64(g>>8) + 0.0722*float64(b>>8)
				gray := uint8(luminance)
				pixels.Set(px, py, color.RGBA{gray, gray, gray, uint8(a >> 8)})
			}
		}
	} else {
		// Partial grayscale
		invAmount := 1 - amount
		for py := y; py < y+h; py++ {
			for px := x; px < x+w; px++ {
				r, g, b, a := pixels.At(px, py).RGBA()
				r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
				luminance := 0.2126*float64(r8) + 0.7152*float64(g8) + 0.0722*float64(b8)
				newR := amount*luminance + invAmount*float64(r8)
				newG := amount*luminance + invAmount*float64(g8)
				newB := amount*luminance + invAmount*float64(b8)
				pixels.Set(px, py, color.RGBA{uint8(newR), uint8(newG), uint8(newB), uint8(a >> 8)})
			}
		}
	}
}

// applySepia applies a sepia tone filter.
// amount = 1 is fully sepia, amount = 0 is unchanged.
func (c *Canvas) applySepia(pixels *image.RGBA, x, y, w, h int, amount float64) {
	if w <= 0 || h <= 0 || amount < 0 || amount > 1 {
		return
	}

	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			r, g, b, a := pixels.At(px, py).RGBA()
			r8, g8, b8 := float64(r>>8), float64(g>>8), float64(b>>8)

			// Sepia transformation matrix
			// R = min(1, R*0.393 + G*0.769 + B*0.189)
			// G = min(1, R*0.349 + G*0.686 + B*0.168)
			// B = min(1, R*0.272 + G*0.534 + B*0.131)
			newR := r8*0.393 + g8*0.769 + b8*0.189
			newG := r8*0.349 + g8*0.686 + b8*0.168
			newB := r8*0.272 + g8*0.534 + b8*0.131

			if newR > 255 {
				newR = 255
			}
			if newG > 255 {
				newG = 255
			}
			if newB > 255 {
				newB = 255
			}

			// Blend with original based on amount
			invAmount := 1 - amount
			finalR := amount*newR + invAmount*r8
			finalG := amount*newG + invAmount*g8
			finalB := amount*newB + invAmount*b8

			pixels.Set(px, py, color.RGBA{uint8(finalR), uint8(finalG), uint8(finalB), uint8(a >> 8)})
		}
	}
}
