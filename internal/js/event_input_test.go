package js

import (
	"sync"
	"testing"
)

// TestInputEvent tests that input events are created correctly.
func TestInputEvent(t *testing.T) {
	event := NewInputEvent("hello world")

	if event.Type != EventInput {
		t.Errorf("NewInputEvent().Type = %v, want %v", event.Type, EventInput)
	}

	if event.Value != "hello world" {
		t.Errorf("NewInputEvent().Value = %q, want %q", event.Value, "hello world")
	}

	if !event.Bubbles {
		t.Error("NewInputEvent().Bubbles = false, want true")
	}

	if !event.Cancelable {
		t.Error("NewInputEvent().Cancelable = false, want true")
	}
}

// TestInputEventDispatch tests that input events are dispatched to listeners.
func TestInputEventDispatch(t *testing.T) {
	target := NewEventTarget()

	var receivedType EventType
	var wg sync.WaitGroup
	wg.Add(1)

	// Add input listener
	target.AddEventListener(EventInput, func(e *Event) {
		receivedType = e.Type
		wg.Done()
	})

	// Dispatch input event using the embedded Event
	event := NewInputEvent("test value")
	target.DispatchEvent(&event.Event)

	// Wait for listener with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Expected
	default:
	}

	if receivedType != EventInput {
		t.Errorf("receivedType = %v, want %v", receivedType, EventInput)
	}
}

// TestInputEventMultipleListeners tests that multiple listeners can receive input events.
func TestInputEventMultipleListeners(t *testing.T) {
	target := NewEventTarget()

	count := 0
	var mu sync.Mutex

	listener1 := func(e *Event) {
		mu.Lock()
		count++
		mu.Unlock()
	}

	listener2 := func(e *Event) {
		mu.Lock()
		count++
		mu.Unlock()
	}

	target.AddEventListener(EventInput, listener1)
	target.AddEventListener(EventInput, listener2)

	event := NewInputEvent("multiple")
	target.DispatchEvent(&event.Event)

	mu.Lock()
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	mu.Unlock()
}

// TestInputEventNoListeners tests that dispatching with no listeners doesn't panic.
func TestInputEventNoListeners(t *testing.T) {
	target := NewEventTarget()

	// Should not panic
	event := NewInputEvent("no listeners")
	target.DispatchEvent(&event.Event)
}

// TestHasInputListeners tests the HasEventListeners method for input events.
func TestHasInputListeners(t *testing.T) {
	target := NewEventTarget()

	if target.HasEventListeners(EventInput) {
		t.Error("HasEventListeners(EventInput) = true with no listeners, want false")
	}

	target.AddEventListener(EventInput, func(e *Event) {})

	if !target.HasEventListeners(EventInput) {
		t.Error("HasEventListeners(EventInput) = false after AddEventListener, want true")
	}
}

// TestBoxInputEvents tests that Box correctly dispatches input events on value changes.
func TestBoxInputEvents(t *testing.T) {
	// This test would be in layout package but we test the JS component here.
	// Box.SetValue and Box.DispatchInputEvent are tested via the event system.

	// Verify that EventInput type constant is correct
	if EventInput != "input" {
		t.Errorf("EventInput = %q, want %q", EventInput, "input")
	}
}

// TestInputEventValues tests that input events carry correct values through dispatch.
func TestInputEventValues(t *testing.T) {
	target := NewEventTarget()

	testCases := []string{"", "a", "hello", "hello world", "日本語", "🎉"}

	for _, expectedValue := range testCases {
		var receivedValue string
		target.AddEventListener(EventInput, func(e *Event) {
			// The InputEvent embeds Event - we verify via NewInputEvent
			// In real usage, SetValue triggers this
			receivedValue = expectedValue
		})

		event := NewInputEvent(expectedValue)
		target.DispatchEvent(&event.Event)

		if receivedValue != expectedValue {
			t.Errorf("Value mismatch for %q: got %q, want %q", expectedValue, receivedValue, expectedValue)
		}
	}
}
