package window

import (
	"sync"
	"time"
)

// EventTarget represents an object that can receive events and have event listeners.
type EventTarget struct {
	mu            sync.RWMutex
	eventListeners map[string][]EventHandler
	wheelHandlers  []WheelEventHandler
}

// EventHandler is a function that handles an event.
type EventHandler func(*Event)

// Event represents a DOM event.
type Event struct {
	Type           string
	Target         *EventTarget
	Bubbles         bool
	Cancelable      bool
	DefaultPrevented bool
	TimeStamp       time.Time
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

// InputEvent represents an input event (including beforeinput).
type InputEvent struct {
	Event
	// Data is the data of the input (for beforeinput/input events)
	Data string
	// InputType is the type of input (insertText, insertParagraph, deleteContentBackward, etc.)
	InputType string
}

// NewInputEvent creates a new InputEvent.
func NewInputEvent(eventType, data, inputType string) *InputEvent {
	return &InputEvent{
		Event: Event{
			Type:       eventType,
			Bubbles:    true,
			Cancelable: eventType == "beforeinput",
			TimeStamp:  time.Now(),
		},
		Data:       data,
		InputType:  inputType,
	}
}

// ChangeEvent represents a change event fired when a form element's value changes.
type ChangeEvent struct {
	Event
	// Data is the new value of the element.
	Data string
}

// NewChangeEvent creates a new ChangeEvent.
func NewChangeEvent(data string) *ChangeEvent {
	return &ChangeEvent{
		Event: Event{
			Type:       "change",
			Bubbles:    true,
			Cancelable: false,
			TimeStamp:  time.Now(),
		},
		Data: data,
	}
}

// DragEvent represents a drag/drop event.
type DragEvent struct {
	Event
	// Data is the data being dragged.
	Data string
	// DragX, DragY are the current drag coordinates.
	DragX, DragY float64
	// EffectAllowed indicates what operations are allowed.
	EffectAllowed string
}

// NewDragEvent creates a new DragEvent.
func NewDragEvent(eventType string, data string, x, y float64) *DragEvent {
	return &DragEvent{
		Event: Event{
			Type:       eventType,
			Bubbles:    true,
			Cancelable: eventType == "dragstart" || eventType == "drag",
			TimeStamp:  time.Now(),
		},
		Data:    data,
		DragX:   x,
		DragY:   y,
	}
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
		wheelHandlers:  make([]WheelEventHandler, 0),
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
	t.mu.Lock()
	defer t.mu.Unlock()
	t.wheelHandlers = append(t.wheelHandlers, handler)
}

// RemoveWheelEventListener removes a wheel event handler.
func (t *EventTarget) RemoveWheelEventListener(handler WheelEventHandler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, h := range t.wheelHandlers {
		if &h == &handler {
			t.wheelHandlers = append(t.wheelHandlers[:i], t.wheelHandlers[i+1:]...)
			return
		}
	}
}

// DispatchWheelEvent dispatches a wheel event to wheel handlers.
func (t *EventTarget) DispatchWheelEvent(event *WheelEvent) {
	t.mu.RLock()
	handlers := t.wheelHandlers
	t.mu.RUnlock()

	// Set target if not already set
	if event.Target == nil {
		event.Target = t
	}

	for _, handler := range handlers {
		handler(event)
	}
}

// GetScrollableParent finds the nearest scrollable ancestor box.
// overflow: auto or overflow: scroll containers are scrollable.
// Returns nil if no scrollable parent is found.
func (t *EventTarget) GetScrollableParent(box interface{}) interface{} {
	// This is a placeholder - actual implementation would traverse the box tree
	// to find a parent with overflow: auto or overflow: scroll
	return nil
}