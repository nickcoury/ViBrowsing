package window

import (
	"sync"
	"time"
)

// EventTarget represents an object that can receive events and have event listeners.
type EventTarget struct {
	mu           sync.RWMutex
	eventListeners map[string][]EventHandler
}

// EventHandler is a function that handles an event.
type EventHandler func(*Event)

// Event represents a DOM event.
type Event struct {
	Type       string
	Target     *EventTarget
	Bubbles    bool
	Cancelable bool
	DefaultPrevented bool
	TimeStamp  time.Time
}

// WheelEvent represents a wheel event as per DOM Level 3.
type WheelEvent struct {
	Event
	// Delta values indicate the scroll amount in the respective axis.
	// These are similar to the deltaY attribute in the DOM WheelEvent interface.
	// Positive values indicate scrolling down/right, negative values indicate up/left.
	DeltaX float64 // Horizontal scroll delta
	DeltaY float64 // Vertical scroll delta
	DeltaZ float64 // Z-axis scroll delta (for specialized devices)
	// DeltaMode indicates the units of measurement for the delta values.
	// 0 = pixels, 1 = lines, 2 = pages, 3 = degrees (for trackpad momentum scrolling)
	DeltaMode int
}

// NewWheelEvent creates a new WheelEvent with the given deltas and default values.
func NewWheelEvent(deltaX, deltaY, deltaZ float64) *WheelEvent {
	return &WheelEvent{
		Event: Event{
			Type:       "wheel",
			Bubbles:    true,
			Cancelable: true,
			TimeStamp:  time.Now(),
		},
		DeltaX:    deltaX,
		DeltaY:    deltaY,
		DeltaZ:    deltaZ,
		DeltaMode: 0, // DOM_DELTA_PIXEL = 0
	}
}

// PreventDefault prevents the default action for this event.
func (e *WheelEvent) PreventDefault() {
	e.DefaultPrevented = true
}

// NewEventTarget creates a new EventTarget.
func NewEventTarget() *EventTarget {
	return &EventTarget{
		eventListeners: make(map[string][]EventHandler),
	}
}

// AddEventListener adds an event handler for the specified event type.
func (t *EventTarget) AddEventListener(eventType string, handler EventHandler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.eventListeners[eventType] = append(t.eventListeners[eventType], handler)
}

// RemoveEventListener removes an event handler for the specified event type.
func (t *EventTarget) RemoveEventListener(eventType string, handler EventHandler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	handlers := t.eventListeners[eventType]
	for i, h := range handlers {
		// Compare function pointers (works if they're the same function reference)
		if &h == &handler {
			t.eventListeners[eventType] = append(handlers[:i], handlers[i+1:]...)
			return
		}
	}
}

// DispatchEvent dispatches an event to the target.
// For the WheelEvent, the Target field should be set before calling this.
func (t *EventTarget) DispatchEvent(event *Event) {
	t.mu.RLock()
	handlers := t.eventListeners[event.Type]
	t.mu.RUnlock()

	// Set target if not already set
	if event.Target == nil {
		event.Target = t
	}

	// Call handlers in order
	for _, handler := range handlers {
		handler(event)
	}
}

// WheelEventHandler is a type alias for handlers specifically for wheel events.
type WheelEventHandler func(*WheelEvent)

// AddWheelEventListener adds a handler for wheel events.
func (t *EventTarget) AddWheelEventListener(handler WheelEventHandler) {
	t.AddEventListener("wheel", func(e *Event) {
		if we, ok := e.(*WheelEvent); ok {
			handler(we)
		}
	})
}

// GetScrollableParent finds the nearest scrollable ancestor box.
// overflow: auto or overflow: scroll containers are scrollable.
// Returns nil if no scrollable parent is found.
func (t *EventTarget) GetScrollableParent(box interface{}) interface{} {
	// This is a placeholder - actual implementation would traverse the box tree
	// to find a parent with overflow: auto or overflow: scroll
	return nil
}
