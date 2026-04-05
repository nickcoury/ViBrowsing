package window

// Window represents the browser window state accessible via JavaScript APIs.
type Window struct {
	// Viewport dimensions (innerWidth, innerHeight)
	ViewportWidth  int
	ViewportHeight int

	// Scroll position
	ScrollX int
	ScrollY int

	// Page content dimensions (for scroll limits)
	ContentWidth  int
	ContentHeight int
}

// NewWindow creates a new Window with default values.
func NewWindow(viewportWidth, viewportHeight int) *Window {
	return &Window{
		ViewportWidth:  viewportWidth,
		ViewportHeight: viewportHeight,
		ScrollX:        0,
		ScrollY:        0,
		ContentWidth:   viewportWidth,
		ContentHeight:  viewportHeight,
	}
}

// InnerWidth returns the viewport width (window.innerWidth).
func (w *Window) InnerWidth() int {
	return w.ViewportWidth
}

// InnerHeight returns the viewport height (window.innerHeight).
func (w *Window) InnerHeight() int {
	return w.ViewportHeight
}

// ScrollTo scrolls the window to the given position.
// It accepts either (x, y) or an options object with scrollX and scrollY.
// behavior can be "smooth" or "auto" (smooth is ignored in this implementation).
func (w *Window) ScrollTo(x interface{}, yOpt ...interface{}) {
	switch arg := x.(type) {
	case float64:
		// ScrollTo(x, y)
		w.ScrollX = int(arg)
		if len(yOpt) > 0 {
			if y, ok := yOpt[0].(float64); ok {
				w.ScrollY = int(y)
			}
		}
	case map[string]interface{}:
		// ScrollTo({left, top, behavior})
		if left, ok := arg["left"].(float64); ok {
			w.ScrollX = int(left)
		}
		if top, ok := arg["top"].(float64); ok {
			w.ScrollY = int(top)
		}
		// behavior is ignored (no smooth scrolling in this renderer)
	}
}

// ScrollBy scrolls the window by the given delta.
// It accepts either (dx, dy) or an options object with deltas.
func (w *Window) ScrollBy(x interface{}, dyOpt ...interface{}) {
	switch arg := x.(type) {
	case float64:
		// ScrollBy(dx, dy)
		w.ScrollX += int(arg)
		if len(dyOpt) > 0 {
			if dy, ok := dyOpt[0].(float64); ok {
				w.ScrollY += int(dy)
			}
		}
	case map[string]interface{}:
		// ScrollBy({left, top, behavior})
		if left, ok := arg["left"].(float64); ok {
			w.ScrollX += int(left)
		}
		if top, ok := arg["top"].(float64); ok {
			w.ScrollY += int(top)
		}
	}
}

// ScrollX returns the current horizontal scroll position.
func (w *Window) ScrollX_() int {
	return w.ScrollX
}

// ScrollY returns the current vertical scroll position.
func (w *Window) ScrollY_() int {
	return w.ScrollY
}

// SetScrollBounds sets the content dimensions for scroll limit calculation.
func (w *Window) SetScrollBounds(contentWidth, contentHeight int) {
	w.ContentWidth = contentWidth
	w.ContentHeight = contentHeight
}

// GetScrollLimits returns the maximum scroll positions.
func (w *Window) GetScrollLimits() (maxScrollX, maxScrollY int) {
	maxScrollX = 0
	if w.ContentWidth > w.ViewportWidth {
		maxScrollX = w.ContentWidth - w.ViewportWidth
	}
	maxScrollY = 0
	if w.ContentHeight > w.ViewportHeight {
		maxScrollY = w.ContentHeight - w.ViewportHeight
	}
	return maxScrollX, maxScrollY
}

// ClampScroll ensures scroll position is within valid bounds.
func (w *Window) ClampScroll() {
	maxX, maxY := w.GetScrollLimits()
	if w.ScrollX < 0 {
		w.ScrollX = 0
	}
	if w.ScrollX > maxX {
		w.ScrollX = maxX
	}
	if w.ScrollY < 0 {
		w.ScrollY = 0
	}
	if w.ScrollY > maxY {
		w.ScrollY = maxY
	}
}
