package window

import (
	"sync"
)

// HistoryState represents the history state object stored in the history stack.
type HistoryState struct {
	Data  interface{}
	Title string
	URL   string
}

// History represents the browser's history stack.
// Implements the History API: https://developer.mozilla.org/en-US/docs/Web/API/History
type History struct {
	mu sync.RWMutex
	// entries is the stack of history entries
	entries []*HistoryState
	// current is the index of the current entry
	current int
	// maxEntries is the maximum number of entries to keep
	maxEntries int
	// popstateHandlers are called when the history changes due to back/forward navigation
	popstateHandlers []func(*HistoryState)
	// onchangeHandlers are called when any history change occurs (push/replace/back/forward)
	onchangeHandlers []func(string) // arg is "push" | "replace" | "pop"
}

// NewHistory creates a new History instance with an initial entry.
func NewHistory(initialURL string) *History {
	h := &History{
		entries:    make([]*HistoryState, 0, 100),
		current:    -1,
		maxEntries: 1000,
	}
	h.pushEntry(&HistoryState{Data: nil, Title: "", URL: initialURL})
	return h
}

// pushEntry adds a new entry to the history stack.
// Must be called with mutex held.
func (h *History) pushEntry(state *HistoryState) {
	h.current++
	if h.current >= len(h.entries) {
		h.entries = append(h.entries, state)
	} else {
		// Truncate forward history when pushing new entry
		h.entries = h.entries[:h.current]
		h.entries = append(h.entries, state)
	}
	// Limit history size
	if len(h.entries) > h.maxEntries {
		// Remove oldest entries
		removeCount := len(h.entries) - h.maxEntries
		h.entries = h.entries[removeCount:]
		h.current -= removeCount
	}
}

// Length returns the number of entries in the history stack.
func (h *History) Length() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.entries)
}

// Current returns the current history entry.
func (h *History) Current() *HistoryState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.current < 0 || h.current >= len(h.entries) {
		return nil
	}
	return h.entries[h.current]
}

// State returns the state object of the current history entry.
func (h *History) State() interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.current < 0 || h.current >= len(h.entries) {
		return nil
	}
	return h.entries[h.current].Data
}

// PushState adds a new entry to the history stack.
// data is the state object to associate with the entry.
// title is a string (currently ignored by most browsers).
// url is the new URL for the entry.
func (h *History) PushState(data interface{}, title string, url string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pushEntry(&HistoryState{Data: data, Title: title, URL: url})
	h.notifyOnChange("push")
}

// ReplaceState replaces the current history entry.
// data is the state object to associate with the entry.
// title is a string (currently ignored by most browsers).
// url is the new URL for the entry.
func (h *History) ReplaceState(data interface{}, title string, url string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.current >= 0 && h.current < len(h.entries) {
		h.entries[h.current] = &HistoryState{Data: data, Title: title, URL: url}
		h.notifyOnChange("replace")
	}
}

// Go navigates to a specific offset in the history stack.
// delta of -1 goes back, 1 goes forward, etc.
func (h *History) Go(delta int) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check bounds BEFORE computing new position
	if h.current+delta < 0 || h.current+delta >= len(h.entries) {
		return &HistoryError{Message: "history index out of bounds"}
	}
	h.current = h.current + delta
	h.notifyPopstate()
	h.notifyOnChange("pop")
	return nil
}

// Back navigates to the previous page in the history stack.
func (h *History) Back() error {
	return h.Go(-1)
}

// Forward navigates to the next page in the history stack.
func (h *History) Forward() error {
	return h.Go(1)
}

// OnPopState registers a handler to be called when the history changes
// due to back/forward navigation.
func (h *History) OnPopState(handler func(*HistoryState)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.popstateHandlers = append(h.popstateHandlers, handler)
}

// OnChange registers a handler to be called whenever the history changes.
// eventType is "push", "replace", or "pop".
func (h *History) OnChange(handler func(string)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onchangeHandlers = append(h.onchangeHandlers, handler)
}

// notifyPopstate calls all popstate handlers with the current state.
// Must be called with mutex held.
func (h *History) notifyPopstate() {
	state := h.entries[h.current]
	for _, handler := range h.popstateHandlers {
		handler(state)
	}
}

// notifyOnChange calls all onchange handlers with the event type.
// Must be called with mutex held.
func (h *History) notifyOnChange(eventType string) {
	for _, handler := range h.onchangeHandlers {
		handler(eventType)
	}
}

// GetEntry returns the history entry at the given index.
func (h *History) GetEntry(index int) *HistoryState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if index < 0 || index >= len(h.entries) {
		return nil
	}
	return h.entries[index]
}

// HistoryError represents an error from the History API.
type HistoryError struct {
	Message string
}

func (e *HistoryError) Error() string {
	return e.Message
}
