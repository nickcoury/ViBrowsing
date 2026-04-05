package window

import (
	"sync"
	"testing"
	"time"
)

func TestNewHistory(t *testing.T) {
	h := NewHistory("https://example.com")
	if h == nil {
		t.Fatal("NewHistory returned nil")
	}
	if h.Length() != 1 {
		t.Errorf("Length = %d, want 1", h.Length())
	}
}

func TestHistoryPushState(t *testing.T) {
	h := NewHistory("https://example.com")

	h.PushState(map[string]interface{}{"page": 1}, "Page 1", "/page1")

	if h.Length() != 2 {
		t.Errorf("Length = %d, want 2", h.Length())
	}

	state := h.State()
	if state == nil {
		t.Fatal("State() returned nil")
	}
	stateMap, ok := state.(map[string]interface{})
	if !ok {
		t.Fatal("State is not a map")
	}
	if stateMap["page"] != 1 {
		t.Errorf("State[page] = %v, want 1", stateMap["page"])
	}

	current := h.Current()
	if current.URL != "/page1" {
		t.Errorf("Current URL = %q, want /page1", current.URL)
	}
}

func TestHistoryReplaceState(t *testing.T) {
	h := NewHistory("https://example.com")

	h.PushState(map[string]interface{}{"page": 1}, "Page 1", "/page1")
	h.ReplaceState(map[string]interface{}{"page": 2}, "Page 2", "/page2")

	if h.Length() != 2 {
		t.Errorf("Length = %d, want 2 (replace shouldn't add entry)", h.Length())
	}

	state := h.State()
	stateMap := state.(map[string]interface{})
	if stateMap["page"] != 2 {
		t.Errorf("State[page] = %v, want 2", stateMap["page"])
	}

	current := h.Current()
	if current.URL != "/page2" {
		t.Errorf("Current URL = %q, want /page2", current.URL)
	}
}

func TestHistoryBackForward(t *testing.T) {
	h := NewHistory("https://example.com")

	h.PushState(1, "Page 1", "/page1")
	h.PushState(2, "Page 2", "/page2")
	h.PushState(3, "Page 3", "/page3")

	// Go back
	err := h.Back()
	if err != nil {
		t.Errorf("Back() returned error: %v", err)
	}

	state := h.State()
	if state != 2 {
		t.Errorf("State after Back() = %v, want 2", state)
	}

	// Go forward
	err = h.Forward()
	if err != nil {
		t.Errorf("Forward() returned error: %v", err)
	}

	state = h.State()
	if state != 3 {
		t.Errorf("State after Forward() = %v, want 3", state)
	}

	// Go back twice
	h.Back()
	h.Back()
	state = h.State()
	if state != 1 {
		t.Errorf("State after 2x Back() = %v, want 1", state)
	}
}

func TestHistoryGo(t *testing.T) {
	h := NewHistory("https://example.com")

	h.PushState(1, "Page 1", "/page1")
	h.PushState(2, "Page 2", "/page2")
	h.PushState(3, "Page 3", "/page3")

	// Go to index 0 via Back twice
	err := h.Go(-2)
	if err != nil {
		t.Errorf("Go(-2) returned error: %v", err)
	}

	state := h.State()
	if state != 1 {
		t.Errorf("State after Go(-2) = %v, want 1", state)
	}

	// Go forward twice
	err = h.Go(2)
	if err != nil {
		t.Errorf("Go(2) returned error: %v", err)
	}

	state = h.State()
	if state != 3 {
		t.Errorf("State after Go(2) = %v, want 3", state)
	}

	// Go out of bounds
	err = h.Go(10)
	if err == nil {
		t.Error("Go(10) should return error for out of bounds")
	}

	// Go way back
	err = h.Go(-5)
	if err == nil {
		t.Error("Go(-5) should return error for out of bounds")
	}
}

func TestHistoryOnPopState(t *testing.T) {
	h := NewHistory("https://example.com")

	h.PushState(1, "Page 1", "/page1")
	h.PushState(2, "Page 2", "/page2")

	var receivedState interface{}
	var mu sync.Mutex
	h.OnPopState(func(state *HistoryState) {
		mu.Lock()
		receivedState = state.Data
		mu.Unlock()
	})

	h.Back()

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if receivedState != 1 {
		t.Errorf("PopState received state = %v, want 1", receivedState)
	}
}

func TestHistoryOnChange(t *testing.T) {
	h := NewHistory("https://example.com")

	var events []string
	var mu sync.Mutex
	h.OnChange(func(event string) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
	})

	h.PushState(1, "Page 1", "/page1")
	h.PushState(2, "Page 2", "/page2")
	h.ReplaceState(3, "Page 3", "/page3")

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	expected := []string{"push", "push", "replace"}
	if len(events) != len(expected) {
		t.Fatalf("Got %d events, want %d: %v", len(events), len(expected), events)
	}
	for i, e := range expected {
		if events[i] != e {
			t.Errorf("Event[%d] = %q, want %q", i, events[i], e)
		}
	}
	mu.Unlock()
}

func TestHistoryState(t *testing.T) {
	h := NewHistory("https://example.com")

	if h.State() != nil {
		t.Error("Initial State() should be nil")
	}

	h.PushState("mystate", "title", "/url")

	state := h.State()
	if state != "mystate" {
		t.Errorf("State() = %v, want mystate", state)
	}
}

func TestHistoryTruncation(t *testing.T) {
	// Create history with small maxEntries to test truncation
	h := &History{
		entries:    make([]*HistoryState, 0, 10),
		current:    -1,
		maxEntries: 10,
	}

	// Add many entries
	for i := 0; i < 20; i++ {
		h.mu.Lock()
		h.pushEntry(&HistoryState{Data: i, Title: "", URL: "/page"})
		h.mu.Unlock()
	}

	h.mu.RLock()
	length := len(h.entries)
	h.mu.RUnlock()

	if length != 10 {
		t.Errorf("Length after 20 pushes with maxEntries=10 = %d, want 10", length)
	}
}
