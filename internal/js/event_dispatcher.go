package js

import "sync"

// EventListener is a function that handles an event.
type EventListener func(*Event)

// EventTarget represents an object that can receive events.
type EventTarget struct {
	mu         sync.RWMutex
	listeners  map[EventType][]EventListener
}

// NewEventTarget creates a new EventTarget.
func NewEventTarget() *EventTarget {
	return &EventTarget{
		listeners: make(map[EventType][]EventListener),
	}
}

// AddEventListener adds an event listener for the given event type.
func (t *EventTarget) AddEventListener(eventType EventType, listener EventListener) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.listeners[eventType] = append(t.listeners[eventType], listener)
}

// RemoveEventListener removes an event listener for the given event type.
func (t *EventTarget) RemoveEventListener(eventType EventType, listener EventListener) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if listeners, ok := t.listeners[eventType]; ok {
		newListeners := make([]EventListener, 0, len(listeners))
		for _, l := range listeners {
			if &l != &listener {
				newListeners = append(newListeners, l)
			}
		}
		t.listeners[eventType] = newListeners
	}
}

// DispatchEvent dispatches an event to all registered listeners.
func (t *EventTarget) DispatchEvent(event *Event) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	event.Target = t

	if listeners, ok := t.listeners[event.Type]; ok {
		for _, listener := range listeners {
			listener(event)
		}
	}
}

// HasEventListeners returns true if there are listeners for the given event type.
func (t *EventTarget) HasEventListeners(eventType EventType) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.listeners[eventType]) > 0
}

// ScrollEventDispatcher is a helper to dispatch scroll events.
type ScrollEventDispatcher struct {
	*EventTarget
}

// NewScrollEventDispatcher creates a new ScrollEventDispatcher.
func NewScrollEventDispatcher() *ScrollEventDispatcher {
	return &ScrollEventDispatcher{
		EventTarget: NewEventTarget(),
	}
}

// DispatchScrollEvent dispatches a scroll event with the given scroll position.
func (d *ScrollEventDispatcher) DispatchScrollEvent(scrollLeft, scrollTop float64) {
	event := NewScrollEvent(scrollLeft, scrollTop)
	d.DispatchEvent(&event.Event)
}

// DispatchWheelEvent dispatches a wheel event with the given deltas.
func (d *ScrollEventDispatcher) DispatchWheelEvent(deltaX, deltaY, deltaZ float64) {
	event := NewWheelEvent(deltaX, deltaY, deltaZ)
	d.DispatchEvent(&event.Event)
}
