package window

import (
	"sync"
	"testing"
	"time"
)

func TestNewClipboard(t *testing.T) {
	c := NewClipboard()
	if c == nil {
		t.Fatal("NewClipboard returned nil")
	}
}

func TestClipboardReadText(t *testing.T) {
	c := NewClipboard()

	text, err := c.ReadText()
	if err != nil {
		t.Errorf("ReadText() returned error: %v", err)
	}
	if text != "" {
		t.Errorf("ReadText() = %q, want empty string", text)
	}
}

func TestClipboardWriteText(t *testing.T) {
	c := NewClipboard()

	err := c.WriteText("Hello, World!")
	if err != nil {
		t.Errorf("WriteText() returned error: %v", err)
	}

	text, err := c.ReadText()
	if err != nil {
		t.Errorf("ReadText() returned error: %v", err)
	}
	if text != "Hello, World!" {
		t.Errorf("ReadText() = %q, want %q", text, "Hello, World!")
	}
}

func TestClipboardWriteTextOverwrite(t *testing.T) {
	c := NewClipboard()

	c.WriteText("First")
	c.WriteText("Second")

	text, _ := c.ReadText()
	if text != "Second" {
		t.Errorf("ReadText() = %q, want %q", text, "Second")
	}
}

func TestClipboardClear(t *testing.T) {
	c := NewClipboard()

	c.WriteText("Some text")
	c.Clear()

	text, _ := c.ReadText()
	if text != "" {
		t.Errorf("After Clear(), ReadText() = %q, want empty string", text)
	}
}

func TestClipboardOnChange(t *testing.T) {
	c := NewClipboard()

	var receivedText string
	var mu sync.Mutex
	c.OnChange(func(text string) {
		mu.Lock()
		receivedText = text
		mu.Unlock()
	})

	c.WriteText("New clipboard content")

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if receivedText != "New clipboard content" {
		t.Errorf("OnChange received %q, want %q", receivedText, "New clipboard content")
	}
}

func TestClipboardOnChangeOnClear(t *testing.T) {
	c := NewClipboard()

	var received []string
	var mu sync.Mutex
	c.OnChange(func(text string) {
		mu.Lock()
		received = append(received, text)
		mu.Unlock()
	})

	c.WriteText("Some text")
	c.Clear()

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	// Both callbacks should have fired - one with "Some text" and one with ""
	if len(received) != 2 {
		t.Errorf("OnChange called %d times, want 2", len(received))
	}
	if received[0] != "Some text" {
		t.Errorf("First OnChange received %q, want %q", received[0], "Some text")
	}
	if received[1] != "" {
		t.Errorf("Second OnChange received %q, want empty string", received[1])
	}
}

func TestClipboardMultipleWrites(t *testing.T) {
	c := NewClipboard()

	texts := []string{"First", "Second", "Third"}
	for _, text := range texts {
		c.WriteText(text)
	}

	lastText, _ := c.ReadText()
	if lastText != "Third" {
		t.Errorf("ReadText() = %q, want %q", lastText, "Third")
	}
}
