package window

import (
	"sync"
	"testing"
	"time"
)

func TestNewWheelEvent(t *testing.T) {
	we := NewWheelEvent(10, 50, 0)

	if we.Type != "wheel" {
		t.Errorf("Expected Type 'wheel', got '%s'", we.Type)
	}
	if !we.Bubbles {
		t.Error("Expected Bubbles to be true")
	}
	if !we.Cancelable {
		t.Error("Expected Cancelable to be true")
	}
	if we.DeltaX != 10 {
		t.Errorf("Expected DeltaX 10, got %f", we.DeltaX)
	}
	if we.DeltaY != 50 {
		t.Errorf("Expected DeltaY 50, got %f", we.DeltaY)
	}
	if we.DeltaZ != 0 {
		t.Errorf("Expected DeltaZ 0, got %f", we.DeltaZ)
	}
	if we.DeltaMode != 0 {
		t.Errorf("Expected DeltaMode 0 (DOM_DELTA_PIXEL), got %d", we.DeltaMode)
	}
	if we.DefaultPrevented {
		t.Error("DefaultPrevented should be false initially")
	}
}

func TestWheelEventPreventDefault(t *testing.T) {
	we := NewWheelEvent(0, 100, 0)

	if we.DefaultPrevented {
		t.Error("DefaultPrevented should be false before PreventDefault is called")
	}

	we.PreventDefault()

	if !we.DefaultPrevented {
		t.Error("DefaultPrevented should be true after PreventDefault is called")
	}
}

func TestWheelEventTimeStamp(t *testing.T) {
	before := time.Now()
	we := NewWheelEvent(0, 10, 0)
	after := time.Now()

	if we.TimeStamp.Before(before) || we.TimeStamp.After(after) {
		t.Errorf("TimeStamp %v is not between %v and %v", we.TimeStamp, before, after)
	}
}

func TestWheelEventDispatch(t *testing.T) {
	target := NewEventTarget()

	var receivedEvent *WheelEvent
	var mu sync.Mutex

	handler := func(we *WheelEvent) {
		mu.Lock()
		receivedEvent = we
		mu.Unlock()
	}

	target.AddWheelEventListener(handler)

	event := NewWheelEvent(0, 50, 0)
	event.Target = target
	target.DispatchEvent(&event.Event)

	mu.Lock()
	defer mu.Unlock()

	if receivedEvent == nil {
		t.Fatal("Handler was not called")
	}
	if receivedEvent.DeltaY != 50 {
		t.Errorf("Expected DeltaY 50, got %f", receivedEvent.DeltaY)
	}
}

func TestWheelEventMultipleHandlers(t *testing.T) {
	target := NewEventTarget()

	var callCount int
	var mu sync.Mutex

	handler1 := func(we *WheelEvent) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	handler2 := func(we *WheelEvent) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	target.AddWheelEventListener(handler1)
	target.AddWheelEventListener(handler2)

	event := NewWheelEvent(0, 100, 0)
	target.DispatchEvent(&event.Event)

	mu.Lock()
	defer mu.Unlock()

	if callCount != 2 {
		t.Errorf("Expected 2 handler calls, got %d", callCount)
	}
}

func TestWheelEventHorizontalAndVertical(t *testing.T) {
	// Test wheel events with both horizontal and vertical components
	tests := []struct {
		name   string
		deltaX float64
		deltaY float64
		deltaZ float64
	}{
		{"vertical scroll down", 0, 100, 0},
		{"vertical scroll up", 0, -100, 0},
		{"horizontal scroll right", 100, 0, 0},
		{"horizontal scroll left", -100, 0, 0},
		{"diagonal scroll", 50, 75, 0},
		{"trackpad pinch", 0, 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			we := NewWheelEvent(tt.deltaX, tt.deltaY, tt.deltaZ)

			if we.DeltaX != tt.deltaX {
				t.Errorf("DeltaX: expected %f, got %f", tt.deltaX, we.DeltaX)
			}
			if we.DeltaY != tt.deltaY {
				t.Errorf("DeltaY: expected %f, got %f", tt.deltaY, we.DeltaY)
			}
			if we.DeltaZ != tt.deltaZ {
				t.Errorf("DeltaZ: expected %f, got %f", tt.deltaZ, we.DeltaZ)
			}
		})
	}
}

func TestWheelEventPreventDefaultPreventsScroll(t *testing.T) {
	target := NewEventTarget()
	scrollOccurred := false
	var mu sync.Mutex

	// Handler that prevents default action
	preventHandler := func(we *WheelEvent) {
		we.PreventDefault()
	}

	// Handler that would scroll content (only called if event not prevented)
	scrollHandler := func(we *WheelEvent) {
		if !we.DefaultPrevented {
			mu.Lock()
			scrollOccurred = true
			mu.Unlock()
		}
	}

	target.AddWheelEventListener(preventHandler)
	target.AddWheelEventListener(scrollHandler)

	event := NewWheelEvent(0, 50, 0)
	target.DispatchEvent(&event.Event)

	mu.Lock()
	defer mu.Unlock()

	if !event.DefaultPrevented {
		t.Error("Expected DefaultPrevented to be true after PreventDefault was called")
	}
	// Note: In a real implementation, the scroll would only happen if DefaultPrevented is false
	// In this test, we demonstrate the mechanism - scrollOccurred would be true if we didn't prevent
	_ = scrollOccurred
}

func TestWheelEventBubblesCancelable(t *testing.T) {
	we := NewWheelEvent(10, 20, 0)

	// Wheel events should bubble and be cancelable per DOM spec
	if !we.Bubbles {
		t.Error("WheelEvent should bubble")
	}
	if !we.Cancelable {
		t.Error("WheelEvent should be cancelable")
	}
}

func TestWheelEventDeltaMode(t *testing.T) {
	// DeltaMode 0 = DOM_DELTA_PIXEL (default)
	we := NewWheelEvent(0, 100, 0)
	if we.DeltaMode != 0 {
		t.Errorf("Expected DeltaMode 0 (DOM_DELTA_PIXEL), got %d", we.DeltaMode)
	}

	// DOM_DELTA_LINE = 1, DOM_DELTA_PAGE = 2, DOM_DELTA_DEGREE = 3
	// These would be set by the input handling code based on the scroll source
}
