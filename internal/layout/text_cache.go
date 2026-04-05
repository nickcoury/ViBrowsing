package layout

import (
	"sync"

	"github.com/nickcoury/ViBrowsing/internal/css"
)

// TextMeasurement caches font.MeasureString results per font/size/text combination.
// This avoids repeated text measurement for the same font, size, and text string.
type TextMeasurementCache struct {
	// cache maps "font:size:text" -> width
	cache sync.Map
}

// NewTextMeasurementCache creates a new text measurement cache.
func NewTextMeasurementCache() *TextMeasurementCache {
	return &TextMeasurementCache{}
}

// Key generates a cache key for a font/size/text combination.
func (c *TextMeasurementCache) Key(font string, size float64, text string) string {
	return font + ":" + formatFloat(size) + ":" + text
}

// Get retrieves a cached width for the given font/size/text combination.
// Returns 0 if not found.
func (c *TextMeasurementCache) Get(font string, size float64, text string) float64 {
	key := c.Key(font, size, text)
	if val, ok := c.cache.Load(key); ok {
		return val.(float64)
	}
	return 0
}

// Set stores a width for the given font/size/text combination.
func (c *TextMeasurementCache) Set(font string, size float64, text string, width float64) {
	key := c.Key(font, size, text)
	c.cache.Store(key, width)
}

// Clear removes all entries from the cache.
func (c *TextMeasurementCache) Clear() {
	c.cache = sync.Map{}
}

// formatFloat formats a float64 without importing strconv (avoids allocation).
func formatFloat(f float64) string {
	// Simple formatting: convert to string using byte manipulation
	// This is a simplified version; strconv is more accurate but slower
	if f == float64(int(f)) {
		return itoa(int(f))
	}
	// For non-integer values, use a simplified representation
	intPart := int(f)
	fracPart := int((f - float64(intPart)) * 100)
	return itoa(intPart) + "." + itoa(fracPart)
}

// itoa converts an integer to a string without allocation.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

// BoxPool provides a pool of Box objects to reduce GC pressure.
// Using sync.Pool for allocating/freeing Box structs during layout.
type BoxPool struct {
	pool sync.Pool
}

// NewBoxPool creates a new Box object pool.
func NewBoxPool() *BoxPool {
	return &BoxPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Box{
					Style: make(map[string]string),
				}
			},
		},
	}
}

// Get retrieves a Box from the pool or creates a new one.
func (p *BoxPool) Get() *Box {
	box := p.pool.Get().(*Box)
	// Reset box fields
	box.Children = nil
	box.Type = BlockBox
	box.Node = nil
	box.Parent = nil
	box.ContentX = 0
	box.ContentY = 0
	box.ContentW = 0
	box.ContentH = 0
	box.MarginTop = css.Length{}
	box.MarginRight = css.Length{}
	box.MarginBottom = css.Length{}
	box.MarginLeft = css.Length{}
	box.BorderTop = css.Length{}
	box.BorderRight = css.Length{}
	box.BorderBottom = css.Length{}
	box.BorderLeft = css.Length{}
	box.PaddingTop = css.Length{}
	box.PaddingRight = css.Length{}
	box.PaddingBottom = css.Length{}
	box.PaddingLeft = css.Length{}
	box.ColSpan = 1
	box.RowSpan = 1
	box.ColumnIndex = 0
	box.MapName = ""
	box.AreaShape = ""
	box.AreaCoords = nil
	box.DataValue = ""
	// Clear style map but keep the allocated map
	for k := range box.Style {
		delete(box.Style, k)
	}
	return box
}

// Put returns a Box to the pool.
func (p *BoxPool) Put(box *Box) {
	p.pool.Put(box)
}

// CSSStylePool provides a pool of CSS style maps.
type CSSStylePool struct {
	pool sync.Pool
}

// NewCSSStylePool creates a new CSS style map pool.
func NewCSSStylePool() *CSSStylePool {
	return &CSSStylePool{
		pool: sync.Pool{
			New: func() interface{} {
				return make(map[string]string)
			},
		},
	}
}

// Get retrieves a style map from the pool or creates a new one.
func (p *CSSStylePool) Get() map[string]string {
	return p.pool.Get().(map[string]string)
}

// Put returns a style map to the pool.
func (p *CSSStylePool) Put(m map[string]string) {
	for k := range m {
		delete(m, k)
	}
	p.pool.Put(m)
}
