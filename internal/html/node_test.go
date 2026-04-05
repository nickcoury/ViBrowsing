package html

import (
	"testing"
)

// TestDOMRectProperties tests that DOMRect properties are calculated correctly
func TestDOMRectProperties(t *testing.T) {
	rect := &DOMRect{
		X:      10,
		Y:      20,
		Width:  100,
		Height: 50,
		Top:    20,
		Right:  110,
		Bottom: 70,
		Left:   10,
	}

	if rect.X != 10 {
		t.Errorf("expected X=10, got %v", rect.X)
	}
	if rect.Y != 20 {
		t.Errorf("expected Y=20, got %v", rect.Y)
	}
	if rect.Width != 100 {
		t.Errorf("expected Width=100, got %v", rect.Width)
	}
	if rect.Height != 50 {
		t.Errorf("expected Height=50, got %v", rect.Height)
	}

	// Check derived properties
	if rect.Top != rect.Y {
		t.Errorf("expected Top=%v, got %v", rect.Y, rect.Top)
	}
	if rect.Left != rect.X {
		t.Errorf("expected Left=%v, got %v", rect.X, rect.Left)
	}
	if rect.Right != rect.X+rect.Width {
		t.Errorf("expected Right=%v, got %v", rect.X+rect.Width, rect.Right)
	}
	if rect.Bottom != rect.Y+rect.Height {
		t.Errorf("expected Bottom=%v, got %v", rect.Y+rect.Height, rect.Bottom)
	}
}

// TestGetBoundingClientRectNilNode tests that nil node returns nil
func TestGetBoundingClientRectNilNode(t *testing.T) {
	var node *Node = nil
	rect := node.GetBoundingClientRectWithBox(nil)
	if rect != nil {
		t.Error("expected nil rect for nil node")
	}
}

// TestGetBoundingClientRectNilBox tests that nil box returns nil
func TestGetBoundingClientRectNilBox(t *testing.T) {
	doc := NewDocument()
	rect := doc.GetBoundingClientRectWithBox(nil)
	if rect != nil {
		t.Error("expected nil rect for nil box")
	}
}

// mockBox is a test implementation of BoxNodeInterface
type mockBox struct {
	node    *Node
	x, y    float64
	w, h    float64
	children []BoxNodeInterface
}

func (m *mockBox) GetNode() *Node { return m.node }
func (m *mockBox) GetContentX() float64 { return m.x }
func (m *mockBox) GetContentY() float64 { return m.y }
func (m *mockBox) GetContentW() float64 { return m.w }
func (m *mockBox) GetContentH() float64 { return m.h }
func (m *mockBox) GetChildren() []BoxNodeInterface { return m.children }

// TestGetBoundingClientRectWithBox tests basic bounding rect calculation
func TestGetBoundingClientRectWithBox(t *testing.T) {
	// Create a simple DOM structure
	doc := NewDocument()
	body := NewElement("body")
	doc.AppendChild(body)

	div := NewElement("div")
	body.AppendChild(div)

	// Create a mock box that corresponds to the div
	mockRoot := &mockBox{
		node: body,
		x:    0, y: 0, w: 800, h: 600,
		children: []BoxNodeInterface{
			&mockBox{
				node: div,
				x:    100, y: 50, w: 200, h: 100,
				children: []BoxNodeInterface{},
			},
		},
	}

	// Get bounding rect for div
	rect := div.GetBoundingClientRectWithBox(mockRoot)
	if rect == nil {
		t.Fatal("expected non-nil rect")
	}

	if rect.X != 100 {
		t.Errorf("expected X=100, got %v", rect.X)
	}
	if rect.Y != 50 {
		t.Errorf("expected Y=50, got %v", rect.Y)
	}
	if rect.Width != 200 {
		t.Errorf("expected Width=200, got %v", rect.Width)
	}
	if rect.Height != 100 {
		t.Errorf("expected Height=100, got %v", rect.Height)
	}
	if rect.Left != 100 {
		t.Errorf("expected Left=100, got %v", rect.Left)
	}
	if rect.Top != 50 {
		t.Errorf("expected Top=50, got %v", rect.Top)
	}
	if rect.Right != 300 {
		t.Errorf("expected Right=300, got %v", rect.Right)
	}
	if rect.Bottom != 150 {
		t.Errorf("expected Bottom=150, got %v", rect.Bottom)
	}
}

// TestGetBoundingClientRectNotFound tests when node has no corresponding box
func TestGetBoundingClientRectNotFound(t *testing.T) {
	// Create a DOM node without a corresponding box
	div := NewElement("div")

	mockRoot := &mockBox{
		node: NewElement("other"),
		x:    0, y: 0, w: 100, h: 100,
		children: []BoxNodeInterface{},
	}

	rect := div.GetBoundingClientRectWithBox(mockRoot)
	if rect != nil {
		t.Error("expected nil rect when node not found in box tree")
	}
}

// TestGetBoundingClientRectDeepNesting tests with deeply nested elements
func TestGetBoundingClientRectDeepNesting(t *testing.T) {
	// Create nested DOM: body > div > p > span
	body := NewElement("body")
	doc := NewDocument()
	doc.AppendChild(body)

	div := NewElement("div")
	body.AppendChild(div)

	p := NewElement("p")
	div.AppendChild(p)

	span := NewElement("span")
	p.AppendChild(span)

	// Create corresponding box tree
	mockSpan := &mockBox{
		node: span,
		x:    50, y: 100, w: 80, h: 20,
		children: []BoxNodeInterface{},
	}
	mockP := &mockBox{
		node: p,
		x:    50, y: 80, w: 200, h: 60,
		children: []BoxNodeInterface{mockSpan},
	}
	mockDiv := &mockBox{
		node: div,
		x:    10, y: 20, w: 300, h: 200,
		children: []BoxNodeInterface{mockP},
	}
	mockBody := &mockBox{
		node: body,
		x:    0, y: 0, w: 800, h: 600,
		children: []BoxNodeInterface{mockDiv},
	}

	// Get rect for span
	rect := span.GetBoundingClientRectWithBox(mockBody)
	if rect == nil {
		t.Fatal("expected non-nil rect for span")
	}

	if rect.X != 50 || rect.Y != 100 || rect.Width != 80 || rect.Height != 20 {
		t.Errorf("unexpected rect values: X=%v Y=%v Width=%v Height=%v",
			rect.X, rect.Y, rect.Width, rect.Height)
	}

	// Get rect for div
	rect = div.GetBoundingClientRectWithBox(mockBody)
	if rect == nil {
		t.Fatal("expected non-nil rect for div")
	}

	if rect.X != 10 || rect.Y != 20 || rect.Width != 300 || rect.Height != 200 {
		t.Errorf("unexpected rect values for div: X=%v Y=%v Width=%v Height=%v",
			rect.X, rect.Y, rect.Width, rect.Height)
	}
}

// TestInnerTextBasic tests basic InnerText functionality
func TestInnerTextBasic(t *testing.T) {
	doc := NewDocument()
	body := NewElement("body")
	doc.AppendChild(body)

	p := NewElement("p")
	body.AppendChild(p)
	p.AppendChild(NewText("Hello"))

	if innerText := body.InnerText(); innerText != "Hello" {
		t.Errorf("expected 'Hello', got %q", innerText)
	}
}

// TestInnerTextHiddenDisplayNone tests that InnerText skips display:none elements
func TestInnerTextHiddenDisplayNone(t *testing.T) {
	doc := NewDocument()
	body := NewElement("body")
	doc.AppendChild(body)

	visible := NewElement("div")
	body.AppendChild(visible)
	visible.AppendChild(NewText("Visible"))

	hidden := NewElement("div")
	body.AppendChild(hidden)
	hidden.SetAttribute("display", "none")
	hidden.AppendChild(NewText("Hidden"))

	if innerText := body.InnerText(); innerText != "Visible" {
		t.Errorf("expected 'Visible', got %q", innerText)
	}
}

// TestInnerTextHiddenVisibility tests that InnerText skips visibility:hidden elements
func TestInnerTextHiddenVisibility(t *testing.T) {
	doc := NewDocument()
	body := NewElement("body")
	doc.AppendChild(body)

	visible := NewElement("span")
	body.AppendChild(visible)
	visible.AppendChild(NewText("Showing"))

	hidden := NewElement("span")
	body.AppendChild(hidden)
	hidden.SetAttribute("visibility", "hidden")
	hidden.AppendChild(NewText("Not showing"))

	if innerText := body.InnerText(); innerText != "Showing" {
		t.Errorf("expected 'Showing', got %q", innerText)
	}
}

// TestInnerTextHiddenStyleDisplayNone tests style attribute for display:none
func TestInnerTextHiddenStyleDisplayNone(t *testing.T) {
	doc := NewDocument()
	body := NewElement("body")
	doc.AppendChild(body)

	visible := NewElement("div")
	body.AppendChild(visible)
	visible.AppendChild(NewText("Seen"))

	hidden := NewElement("div")
	body.AppendChild(hidden)
	hidden.SetAttribute("style", "display: none")
	hidden.AppendChild(NewText("Unseen"))

	if innerText := body.InnerText(); innerText != "Seen" {
		t.Errorf("expected 'Seen', got %q", innerText)
	}
}

// TestInnerTextBlockElements tests that newlines are added between block elements
func TestInnerTextBlockElements(t *testing.T) {
	doc := NewDocument()
	body := NewElement("body")
	doc.AppendChild(body)

	h1 := NewElement("h1")
	body.AppendChild(h1)
	h1.AppendChild(NewText("Title"))

	p := NewElement("p")
	body.AppendChild(p)
	p.AppendChild(NewText("Paragraph"))

	expected := "Title\nParagraph"
	if innerText := body.InnerText(); innerText != expected {
		t.Errorf("expected %q, got %q", expected, innerText)
	}
}

// TestInnerTextVsTextContent tests that InnerText and TextContent differ for hidden elements
func TestInnerTextVsTextContent(t *testing.T) {
	doc := NewDocument()
	body := NewElement("body")
	doc.AppendChild(body)

	visible := NewElement("div")
	body.AppendChild(visible)
	visible.AppendChild(NewText("Text 1"))

	hidden := NewElement("div")
	body.AppendChild(hidden)
	hidden.SetAttribute("display", "none")
	hidden.AppendChild(NewText("Text 2"))

	// TextContent returns all text
	if tc := body.TextContent(); tc != "Text 1Text 2" {
		t.Errorf("TextContent: expected 'Text 1Text 2', got %q", tc)
	}

	// InnerText skips hidden
	if it := body.InnerText(); it != "Text 1" {
		t.Errorf("InnerText: expected 'Text 1', got %q", it)
	}
}

// TestInnerTextDeeplyHidden tests nested hidden elements
func TestInnerTextDeeplyHidden(t *testing.T) {
	doc := NewDocument()
	body := NewElement("body")
	doc.AppendChild(body)

	outer := NewElement("div")
	body.AppendChild(outer)
	outer.AppendChild(NewText("Outer text "))

	inner := NewElement("div")
	outer.AppendChild(inner)
	inner.SetAttribute("display", "none")
	inner.AppendChild(NewText("Inner hidden"))

	if innerText := body.InnerText(); innerText != "Outer text " {
		t.Errorf("expected 'Outer text ', got %q", innerText)
	}
}

// TestInnerTextEmpty tests InnerText on empty elements
func TestInnerTextEmpty(t *testing.T) {
	div := NewElement("div")
	if innerText := div.InnerText(); innerText != "" {
		t.Errorf("expected empty string, got %q", innerText)
	}
}

// TestInnerTextNestedBlockAndInline tests mixed block/inline nesting
func TestInnerTextNestedBlockAndInline(t *testing.T) {
	doc := NewDocument()
	body := NewElement("body")
	doc.AppendChild(body)

	div := NewElement("div")
	body.AppendChild(div)

	span1 := NewElement("span")
	div.AppendChild(span1)
	span1.AppendChild(NewText("A"))

	br := NewElement("br")
	div.AppendChild(br)

	span2 := NewElement("span")
	div.AppendChild(span2)
	span2.AppendChild(NewText("B"))

	expected := "A\nB"
	if innerText := body.InnerText(); innerText != expected {
		t.Errorf("expected %q, got %q", expected, innerText)
	}
}
