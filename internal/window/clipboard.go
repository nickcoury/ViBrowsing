package window

import (
	"sync"
)

// Clipboard represents the Clipboard API (navigator.clipboard).
// Provides read and write access to the clipboard contents.
type Clipboard struct {
	mu        sync.RWMutex
	text      string
	onchange  func(string) // Called when clipboard content changes
}

// NewClipboard creates a new Clipboard instance.
func NewClipboard() *Clipboard {
	return &Clipboard{
		text: "",
	}
}

// ReadText reads the current clipboard text content.
func (c *Clipboard) ReadText() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// In a real implementation, this would use OS-level clipboard access.
	// For ViBrowsing, we return the internal clipboard state.
	return c.text, nil
}

// WriteText writes text to the clipboard.
func (c *Clipboard) WriteText(text string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.text = text
	if c.onchange != nil {
		c.onchange(text)
	}
	return nil
}

// OnChange registers a handler to be called when the clipboard content changes.
func (c *Clipboard) OnChange(handler func(string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onchange = handler
}

// Clear clears the clipboard.
func (c *Clipboard) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	oldText := c.text
	c.text = ""
	if c.onchange != nil && oldText != "" {
		c.onchange("")
	}
	return nil
}

// ClipboardError represents an error from the Clipboard API.
type ClipboardError struct {
	Message string
}

func (e *ClipboardError) Error() string {
	return e.Message
}

// ClipboardItem represents a single clipboard item with a MIME type and data.
type ClipboardItem struct {
	MimeType string
	Data     []byte
}

// ClipboardReadOptions provides options for reading clipboard data.
type ClipboardReadOptions struct {
	// If true, the user will not be prompted for permission.
	// Not used in this implementation.
	Anonymous bool
}

// ClipboardReadResult holds the result of a clipboard read operation.
type ClipboardReadResult struct {
	Items []ClipboardItem
}
