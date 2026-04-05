package window

import (
	"sync"
	"testing"
	"time"
)

func TestNewAbortController(t *testing.T) {
	ac := NewAbortController()
	if ac == nil {
		t.Fatal("NewAbortController returned nil")
	}
	if ac.Signal() == nil {
		t.Error("Signal() returned nil")
	}
}

func TestAbortControllerAbort(t *testing.T) {
	ac := NewAbortController()
	signal := ac.Signal()

	if signal.Aborted() {
		t.Error("Signal should not be aborted initially")
	}

	ac.Abort()

	if !signal.Aborted() {
		t.Error("Signal should be aborted after Abort()")
	}
}

func TestAbortSignalOnAbort(t *testing.T) {
	ac := NewAbortController()
	signal := ac.Signal()

	var called bool
	var mu sync.Mutex
	signal.SetOnAbort(func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	ac.Abort()

	// Give goroutine time to execute
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !called {
		t.Error("OnAbort handler should have been called")
	}
}

func TestAbortSignalMultipleAbort(t *testing.T) {
	ac := NewAbortController()
	signal := ac.Signal()

	callCount := 0
	var mu sync.Mutex
	signal.SetOnAbort(func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	ac.Abort()
	ac.Abort()
	ac.Abort()

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	// Handler should only be called once (first abort)
	if callCount != 1 {
		t.Errorf("OnAbort called %d times, want 1", callCount)
	}
}

func TestAbortSignalThrowIfAborted(t *testing.T) {
	ac := NewAbortController()
	signal := ac.Signal()

	// Should not error when not aborted
	err := signal.ThrowIfAborted()
	if err != nil {
		t.Errorf("ThrowIfAborted() returned error when not aborted: %v", err)
	}

	// After abort, should return error
	ac.Abort()
	err = signal.ThrowIfAborted()
	if err == nil {
		t.Error("ThrowIfAborted() should return error after abort")
	}
	if _, ok := err.(*AbortError); !ok {
		t.Errorf("Error should be *AbortError, got %T", err)
	}
}

func TestAbortSignalAbortedState(t *testing.T) {
	ac := NewAbortController()
	signal := ac.Signal()

	if signal.Aborted() {
		t.Error("Should not be aborted initially")
	}

	ac.Abort()

	if !signal.Aborted() {
		t.Error("Should be aborted after Abort()")
	}
}

func TestAbortSignalIndependentControllers(t *testing.T) {
	ac1 := NewAbortController()
	ac2 := NewAbortController()

	ac1.Abort()

	if !ac1.Signal().Aborted() {
		t.Error("ac1 signal should be aborted")
	}
	if ac2.Signal().Aborted() {
		t.Error("ac2 signal should not be aborted when ac1 is aborted")
	}
}

func TestAbortSignalOnAbortCalledOnce(t *testing.T) {
	ac := NewAbortController()
	signal := ac.Signal()

	done := make(chan bool)
	signal.SetOnAbort(func() {
		// Simulate some work
		time.Sleep(5 * time.Millisecond)
		done <- true
	})

	ac.Abort()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("OnAbort handler did not complete in time")
	}
}
