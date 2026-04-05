package layout

import (
	"sync"
	"testing"

	"github.com/nickcoury/ViBrowsing/internal/js"
)

// TestInputEvent tests that input events are created and dispatched correctly.
func TestInputEvent(t *testing.T) {
	event := js.NewInputEvent("hello world")

	if event.Type != js.EventInput {
		t.Errorf("NewInputEvent().Type = %v, want %v", event.Type, js.EventInput)
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

// TestBoxSetValue tests that SetValue updates the internal value.
func TestBoxSetValue(t *testing.T) {
	box := &Box{
		value: "",
	}

	// Set value via SetValue method
	box.SetValue("hello")

	if box.value != "hello" {
		t.Errorf("box.value = %q, want %q", box.value, "hello")
	}
}

// TestBoxSetValueAndEvent tests that SetValue dispatches an input event when there's a listener.
func TestBoxSetValueAndEvent(t *testing.T) {
	box := &Box{
		value: "",
	}

	var receivedType js.EventType
	var mu sync.Mutex

	// Add an input event listener
	box.AddEventListener(js.EventInput, func(e *js.Event) {
		mu.Lock()
		receivedType = e.Type
		mu.Unlock()
	})

	// Set value - should trigger input event
	box.SetValue("test value")

	mu.Lock()
	if receivedType != js.EventInput {
		t.Errorf("receivedType = %v, want %v", receivedType, js.EventInput)
	}
	mu.Unlock()

	// Value should be updated
	if box.value != "test value" {
		t.Errorf("box.value = %q, want %q", box.value, "test value")
	}
}

// TestBoxSetValueNoChange tests that SetValue doesn't dispatch event if value unchanged.
func TestBoxSetValueNoChange(t *testing.T) {
	box := &Box{
		value: "same",
	}

	count := 0
	var mu sync.Mutex

	box.AddEventListener(js.EventInput, func(e *js.Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	box.SetValue("same")
	box.SetValue("same")

	mu.Lock()
	if count != 0 {
		t.Errorf("count = %d, want 0 (no event should fire for same value)", count)
	}
	mu.Unlock()
}

// TestBoxGetValue tests the GetValue method.
func TestBoxGetValue(t *testing.T) {
	box := &Box{
		value: "initial",
	}

	if box.GetValue() != "initial" {
		t.Errorf("GetValue() = %q, want %q", box.GetValue(), "initial")
	}
}

// TestBoxGetValueEmpty tests GetValue on box with no value set.
func TestBoxGetValueEmpty(t *testing.T) {
	box := &Box{}

	if box.GetValue() != "" {
		t.Errorf("GetValue() = %q, want empty string", box.GetValue())
	}
}

// TestBoxHasInputListeners tests the HasInputListeners method.
func TestBoxHasInputListeners(t *testing.T) {
	box := &Box{}

	if box.HasInputListeners() {
		t.Error("HasInputListeners() = true with no listeners, want false")
	}

	box.AddEventListener(js.EventInput, func(e *js.Event) {})

	if !box.HasInputListeners() {
		t.Error("HasInputListeners() = false after adding listener, want true")
	}
}

// TestBoxDispatchInputEvent tests DispatchInputEvent method.
func TestBoxDispatchInputEvent(t *testing.T) {
	box := &Box{
		value: "test",
	}

	var receivedType js.EventType
	var mu sync.Mutex

	box.AddEventListener(js.EventInput, func(e *js.Event) {
		mu.Lock()
		receivedType = e.Type
		mu.Unlock()
	})

	box.DispatchInputEvent()

	mu.Lock()
	if receivedType != js.EventInput {
		t.Errorf("receivedType = %v, want %v", receivedType, js.EventInput)
	}
	mu.Unlock()
}

// TestBoxDispatchInputEventNoListeners tests that DispatchInputEvent doesn't panic with no listeners.
func TestBoxDispatchInputEventNoListeners(t *testing.T) {
	box := &Box{
		value: "test",
	}

	// Should not panic
	box.DispatchInputEvent()
}

// TestBoxInputEventFromEmpty tests that SetValue on initially empty value dispatches event.
func TestBoxInputEventFromEmpty(t *testing.T) {
	box := &Box{
		value: "",
	}

	eventCount := 0
	var mu sync.Mutex

	box.AddEventListener(js.EventInput, func(e *js.Event) {
		mu.Lock()
		eventCount++
		mu.Unlock()
	})

	box.SetValue("a")
	box.SetValue("ab")
	box.SetValue("abc")

	mu.Lock()
	if eventCount != 3 {
		t.Errorf("eventCount = %d, want 3", eventCount)
	}
	mu.Unlock()
}

// TestBoxInputEventMultibyte tests that multibyte characters work correctly.
func TestBoxInputEventMultibyte(t *testing.T) {
	box := &Box{
		value: "",
	}

	var lastValue string
	var mu sync.Mutex

	box.AddEventListener(js.EventInput, func(e *js.Event) {
		mu.Lock()
		lastValue = box.value
		mu.Unlock()
	})

	testCases := []string{
		"日本語",
		"🎉",
		"emoji 🎉 and text",
		"mixed: abc 123",
	}

	for _, tc := range testCases {
		box.SetValue(tc)

		mu.Lock()
		if lastValue != tc {
			t.Errorf("lastValue = %q, want %q for input %q", lastValue, tc, tc)
		}
		mu.Unlock()
	}
}
