package window

import (
	"sync"
)

// AbortController represents an AbortController as per the Fetch API spec.
// It can be used to abort fetch requests and other operations.
type AbortController struct {
	mu     sync.RWMutex
	signal *AbortSignal
}

// AbortSignal represents an AbortSignal as per the Fetch API spec.
// It represents a signal object that allows you to communicate with a DOM request
// and abort it if required.
type AbortSignal struct {
	mu       sync.RWMutex
	aborted  bool
	onabort  func()
	ontimeout func()
	timeoutMs int
}

// NewAbortController creates a new AbortController instance.
func NewAbortController() *AbortController {
	return &AbortController{
		signal: &AbortSignal{
			aborted:   false,
			timeoutMs: 0,
		},
	}
}

// Signal returns the AbortSignal associated with this AbortController.
func (ac *AbortController) Signal() *AbortSignal {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.signal
}

// Abort aborts the request. This aborts all fetch requests using the associated
// AbortSignal.
func (ac *AbortController) Abort() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.signal != nil {
		ac.signal.abort()
	}
}

// abortInternal is the internal abort implementation.
// Must be called with mutex held.
func (s *AbortSignal) abort() {
	if !s.aborted {
		s.aborted = true
		if s.onabort != nil {
			go s.onabort()
		}
	}
}

// Aborted returns true if the signal has been aborted.
func (s *AbortSignal) Aborted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.aborted
}

// OnAbort sets a callback that will be called when the signal is aborted.
func (s *AbortSignal) OnAbort(handler func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onabort = handler
}

// SetOnAbort sets the abort event handler.
func (s *AbortSignal) SetOnAbort(handler func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onabort = handler
}

// Timeout sets a timeout in milliseconds. When the timeout expires,
// the signal will be automatically aborted.
func (s *AbortSignal) Timeout(ms int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timeoutMs = ms
}

// OnTimeout sets a callback that will be called when the signal times out.
func (s *AbortSignal) OnTimeout(handler func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ontimeout = handler
}

// ThrowIfAborted throws an error if the signal has been aborted.
// This implements the "throwIfAborted" algorithm from the spec.
func (s *AbortSignal) ThrowIfAborted() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.aborted {
		return &AbortError{Message: "signal aborted"}
	}
	return nil
}

// AbortError represents an error when an operation is aborted.
type AbortError struct {
	Message string
}

func (e *AbortError) Error() string {
	return e.Message
}
