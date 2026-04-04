package html

import "strings"

// NodeType represents the type of a DOM node.
type NodeType int

const (
	NodeDocument NodeType = iota
	NodeElement
	NodeText
	NodeComment
)

// Node represents a DOM node.
type Node struct {
	Type       NodeType
	TagName    string // e.g., "div", "span", "html"
	Data       string // text content for text nodes
	Attributes []Attribute
	Parent     *Node
	Children   []*Node
}

// NewDocument creates a new document node.
func NewDocument() *Node {
	return &Node{Type: NodeDocument, TagName: "#document"}
}

// NewElement creates a new element node.
func NewElement(tagName string) *Node {
	return &Node{Type: NodeElement, TagName: tagName}
}

// FindChildByTagName returns the first child element with the given tag name, or nil.
func (n *Node) FindChildByTagName(tagName string) *Node {
	for _, child := range n.Children {
		if child.Type == NodeElement && child.TagName == tagName {
			return child
		}
	}
	return nil
}

// NewText creates a new text node.
func NewText(data string) *Node {
	return &Node{Type: NodeText, Data: data}
}

// NewComment creates a new comment node.
func NewComment(data string) *Node {
	return &Node{Type: NodeComment, Data: data}
}

// AppendChild adds a child node, updating parent pointers.
func (n *Node) AppendChild(child *Node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

// QuerySelectorAll returns all descendant elements matching the selector.
// Currently supports: tagname, #id, .class, and combinations.
func (n *Node) QuerySelectorAll(selector string) []*Node {
	var results []*Node
	n.querySelectorAll(selector, &results)
	return results
}

func (n *Node) querySelectorAll(selector string, results *[]*Node) {
	// Simple selector matching
	if matchSelector(n, selector) {
		*results = append(*results, n)
	}

	for _, child := range n.Children {
		child.querySelectorAll(selector, results)
	}
}

// matchSelector returns true if node matches the given selector.
// Tag name matching is case-insensitive per HTML5 spec.
func matchSelector(n *Node, selector string) bool {
	if n.Type != NodeElement {
		return false
	}

	// Tag name selector (case-insensitive per HTML5 spec)
	if strings.EqualFold(selector, n.TagName) {
		return true
	}

	// ID selector (#id)
	if len(selector) > 1 && selector[0] == '#' {
		id := selector[1:]
		for _, attr := range n.Attributes {
			if attr.Key == "id" && attr.Value == id {
				return true
			}
		}
	}

	// Class selector (.class)
	if len(selector) > 1 && selector[0] == '.' {
		class := selector[1:]
		for _, attr := range n.Attributes {
			if attr.Key == "class" {
				// Simple class matching
				for _, c := range splitClasses(attr.Value) {
					if c == class {
						return true
					}
				}
			}
		}
	}

	return false
}

func splitClasses(value string) []string {
	var classes []string
	var current []byte
	for _, c := range []byte(value) {
		if c == ' ' || c == '\t' || c == '\n' {
			if len(current) > 0 {
				classes = append(classes, string(current))
				current = nil
			}
		} else {
			current = append(current, c)
		}
	}
	if len(current) > 0 {
		classes = append(classes, string(current))
	}
	return classes
}

// GetAttribute returns the value of an attribute, or "" if not present.
func (n *Node) GetAttribute(key string) string {
	for _, attr := range n.Attributes {
		if attr.Key == key {
			return attr.Value
		}
	}
	return ""
}

// InsideTemplate returns true if this node is a descendant of a <template> element.
func (n *Node) InsideTemplate() bool {
	current := n.Parent
	for current != nil {
		if current.Type == NodeElement && current.TagName == "template" {
			return true
		}
		current = current.Parent
	}
	return false
}

// InnerText returns the concatenated text content of this node and all descendants.
func (n *Node) InnerText() string {
	if n.Type == NodeText {
		return n.Data
	}
	var text string
	for _, child := range n.Children {
		text += child.InnerText()
	}
	return text
}

// String returns a simple string representation of the node tree.
func (n *Node) String() string {
	return n.stringIndent(0)
}

func (n *Node) stringIndent(depth int) string {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	switch n.Type {
	case NodeDocument:
		return indent + "#document\n" + childrenString(n.Children, depth+1)
	case NodeElement:
		s := indent + "<" + n.TagName + ">\n"
		return s + childrenString(n.Children, depth+1)
	case NodeText:
		text := n.Data
		if len(text) > 50 {
			text = text[:50] + "..."
		}
		return indent + "\"" + text + "\"\n"
	case NodeComment:
		return indent + "<!-- " + n.Data + " -->\n"
	}
	return ""
}

func childrenString(children []*Node, depth int) string {
	if len(children) == 0 {
		return ""
	}
	s := ""
	for _, child := range children {
		s += child.stringIndent(depth)
	}
	return s
}
