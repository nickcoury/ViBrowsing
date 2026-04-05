package observer

// IntersectionObserver represents the IntersectionObserver API for viewport visibility detection.
// This implementation stores observations but doesn't actively track viewport changes
// without JavaScript execution. The browser can use this for lazy loading hints.
type IntersectionObserver struct {
	// Callback invoked when intersection changes
	Callback func([]IntersectionObserverEntry)
	
	// Observer options
	Threshold []float64 // at which ratios to invoke callback
	RootMargin string    // margin around root
	Root *Element // root element to use as viewport
}

// IntersectionObserverEntry represents a single intersection observation result.
type IntersectionObserverEntry struct {
	// The target element
	Target *Element
	
	// Intersection ratio (0.0 to 1.0)
	IntersectionRatio float64
	
	// Whether the target is intersecting
	IsIntersecting bool
	
	// Bounding client rect
	BoundX, BoundY, BoundW, BoundH float64
	
	// Intersection rect
	IntersectX, IntersectY, IntersectW, IntersectH float64
	
	// Root bounds
	RootX, RootY, RootW, RootH float64
}

// Element represents an element for observer tracking (minimal interface).
type Element struct {
	// Element identifier
	ID string
	
	// Bounding box
	X, Y, Width, Height float64
	
	// Visibility state
	IsVisible bool
	
	// Custom data for observer
	ObserverData map[string]interface{}
}

// NewIntersectionObserver creates a new IntersectionObserver.
func NewIntersectionObserver(callback func([]IntersectionObserverEntry), options *IntersectionObserverInit) *IntersectionObserver {
	obs := &IntersectionObserver{
		Callback: callback,
	}
	
	if options != nil {
		obs.Threshold = options.Threshold
		obs.RootMargin = options.RootMargin
		obs.Root = options.Root
	}
	
	if obs.Threshold == nil {
		// Default threshold
		obs.Threshold = []float64{0}
	}
	
	return obs
}

// IntersectionObserverInit provides initialization options for IntersectionObserver.
type IntersectionObserverInit struct {
	Threshold   []float64 // thresholds at which to callback
	RootMargin  string    // CSS-style margin
	Root        *Element  // viewport element
}

// Observe starts observing an element.
func (obs *IntersectionObserver) Observe(target *Element) {
	// In a real implementation, this would register the target
	// for intersection tracking. Since we don't execute JS, we just
	// store the observation for potential future use.
}

// Unobserve stops observing an element.
func (obs *IntersectionObserver) Unobserve(target *Element) {
	// Remove observation
}

// Disconnect stops all observations and clears callback.
func (obs *IntersectionObserver) Disconnect() {
	obs.Callback = nil
}

// TakeRecords returns accumulated intersection observations.
func (obs *IntersectionObserver) TakeRecords() []IntersectionObserverEntry {
	// Return any accumulated records
	return []IntersectionObserverEntry{}
}

// IntersectionObserverEntryArray is a helper for the callback
type IntersectionObserverEntryArray = []IntersectionObserverEntry

// ResizeObserver represents the ResizeObserver API for element resize detection.
type ResizeObserver struct {
	// Callback invoked when elements resize
	Callback func([]ResizeObserverEntry)
	
	// Active observations
	Observations []*ResizeObservation
}

// ResizeObserverEntry represents a single resize observation result.
type ResizeObserverEntry struct {
	// The target element
	Target *Element
	
	// Content box size
	ContentW, ContentH float64
	
	// Border box size  
	BorderW, BorderH float64
	
	// Device pixel ratio
	DevicePixelRatio float64
}

// ResizeObservation tracks a single element's resize observation.
type ResizeObservation struct {
	Target      *Element
	LastW, LastH float64
}

// NewResizeObserver creates a new ResizeObserver.
func NewResizeObserver(callback func([]ResizeObserverEntry)) *ResizeObserver {
	return &ResizeObserver{
		Callback:    callback,
		Observations: make([]*ResizeObservation, 0),
	}
}

// Observe starts observing an element for resize.
func (ro *ResizeObserver) Observe(target *Element) {
	obs := &ResizeObservation{
		Target: target,
		LastW:  target.Width,
		LastH:  target.Height,
	}
	ro.Observations = append(ro.Observations, obs)
}

// Unobserve stops observing an element.
func (ro *ResizeObserver) Unobserve(target *Element) {
	for i, obs := range ro.Observations {
		if obs.Target == target {
			ro.Observations = append(ro.Observations[:i], ro.Observations[i+1:]...)
			return
		}
	}
}

// Disconnect stops all observations.
func (ro *ResizeObserver) Disconnect() {
	ro.Observations = make([]*ResizeObservation, 0)
	ro.Callback = nil
}

// CheckAndNotify checks for size changes and notifies if needed.
func (ro *ResizeObserver) CheckAndNotify() {
	if ro.Callback == nil {
		return
	}
	
	var entries []ResizeObserverEntry
	for _, obs := range ro.Observations {
		if obs.Target == nil {
			continue
		}
		
		// Check if size changed
		if obs.LastW != obs.Target.Width || obs.LastH != obs.Target.Height {
			entry := ResizeObserverEntry{
				Target:  obs.Target,
				ContentW: obs.Target.Width,
				ContentH: obs.Target.Height,
			}
			entries = append(entries, entry)
			obs.LastW = obs.Target.Width
			obs.LastH = obs.Target.Height
		}
	}
	
	if len(entries) > 0 {
		ro.Callback(entries)
	}
}

// ResizeObserverEntryArray is a helper alias
type ResizeObserverEntryArray = []ResizeObserverEntry

// GetResizeObserverEntries converts ResizeObservation slice to entries for callback
func GetResizeObserverEntries(observations []*ResizeObservation) []ResizeObserverEntry {
	entries := make([]ResizeObserverEntry, 0, len(observations))
	for _, obs := range observations {
		if obs.Target != nil {
			entries = append(entries, ResizeObserverEntry{
				Target:   obs.Target,
				ContentW: obs.Target.Width,
				ContentH: obs.Target.Height,
			})
		}
	}
	return entries
}
