package js

// EventType represents the type of a DOM event.
type EventType string

// Event type constants for common DOM events.
const (
	EventClick     EventType = "click"
	EventScroll    EventType = "scroll"
	EventWheel     EventType = "wheel"
	EventInput     EventType = "input"
	EventChange    EventType = "change"
	EventBeforeInput EventType = "beforeinput"
	EventKeyDown   EventType = "keydown"
	EventKeyUp     EventType = "keyup"
	EventKeyPress  EventType = "keypress"
	EventFocus     EventType = "focus"
	EventBlur      EventType = "blur"
	EventSubmit    EventType = "submit"
	EventLoad      EventType = "load"
	EventError     EventType = "error"
	EventMouseDown EventType = "mousedown"
	EventMouseUp   EventType = "mouseup"
	EventMouseMove EventType = "mousemove"
	EventResize    EventType = "resize"
	EventScrollEnd EventType = "scrollend"
)

// Event represents a DOM event.
type Event struct {
	Type     EventType
	Bubbles  bool
	Cancelable bool
	DefaultPrevented bool
	Target   interface{}
}

// NewEvent creates a new event with the given type.
func NewEvent(eventType EventType) *Event {
	return &Event{
		Type:     eventType,
		Bubbles:  true,
		Cancelable: true,
	}
}

// ScrollEvent represents a scroll event.
type ScrollEvent struct {
	Event
	ScrollLeft float64
	ScrollTop  float64
	ScrollX    float64
	ScrollY    float64
}

// NewScrollEvent creates a new scroll event.
func NewScrollEvent(scrollLeft, scrollTop float64) *ScrollEvent {
	return &ScrollEvent{
		Event: Event{
			Type:       EventScroll,
			Bubbles:    true,
			Cancelable: true,
		},
		ScrollLeft: scrollLeft,
		ScrollTop:  scrollTop,
		ScrollX:    scrollLeft,
		ScrollY:    scrollTop,
	}
}

// WheelEvent represents a mouse wheel event.
type WheelEvent struct {
	Event
	DeltaX float64
	DeltaY float64
	DeltaZ float64
}

// NewWheelEvent creates a new wheel event.
func NewWheelEvent(deltaX, deltaY, deltaZ float64) *WheelEvent {
	return &WheelEvent{
		Event: Event{
			Type:       EventWheel,
			Bubbles:    true,
			Cancelable: true,
		},
		DeltaX: deltaX,
		DeltaY: deltaY,
		DeltaZ: deltaZ,
	}
}

// InputEvent represents an input event fired when the value of an input/textarea element changes.
type InputEvent struct {
	Event
	// Value is the current value of the input/textarea at the time of the event.
	Value string
}

// NewInputEvent creates a new input event with the given value.
func NewInputEvent(value string) *InputEvent {
	return &InputEvent{
		Event: Event{
			Type:       EventInput,
			Bubbles:    true,
			Cancelable: true,
		},
		Value: value,
	}
}
