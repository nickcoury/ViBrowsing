package html

import (
	"fmt"
	"strings"
)

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
	TagName    string // e.g. "div", "span", "html"
	Data       string // text content for text nodes
	Attributes []Attribute
	Parent     *Node
	Children   []*Node

	// TemplateContent is the inert document fragment for <template> elements.
	// The content inside <template> is parsed but not rendered until
	// JavaScript activates it (typically via template.content).
	TemplateContent *Node
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

// matchSelector returns true if node matches the given selector.
// Tag name matching is case-insensitive per HTML5 spec.
// Handles: tagname, #id, .class, [attr], [attr=value], :not(), :first-child, :last-child, :only-child
func matchSelector(n *Node, selector string) bool {
	if n.Type != NodeElement {
		return false
	}

	sel := strings.TrimSpace(selector)
	tagName := strings.ToLower(n.TagName)

	// Handle compound selectors like "tag:not(...)" or "tag.class"
	// We need to check ALL parts of the selector
	remaining := sel

	for len(remaining) > 0 {
		remaining = strings.TrimSpace(remaining)
		if len(remaining) == 0 {
			break
		}

		switch remaining[0] {
		case ':':
			// Pseudo-class or pseudo-element
			// Find the end of the pseudo
			var pseudoName, pseudoArg string
			if idx := strings.Index(remaining[1:], "("); idx >= 0 && idx < strings.Index(remaining[1:], ")") {
				// Has argument
				pseudoName = remaining[1 : idx+1]
				argStart := idx + 2
				depth := 1
				for i := argStart; i < len(remaining); i++ {
					if remaining[i] == '(' {
						depth++
					} else if remaining[i] == ')' {
						depth--
						if depth == 0 {
							pseudoArg = remaining[argStart:i]
							remaining = remaining[i+1:]
							break
						}
					}
				}
			} else {
				// No argument - find end of pseudo name
				end := 1
				for end < len(remaining) && (remaining[end] == ':' || remaining[end] == '-' || (remaining[end] >= 'a' && remaining[end] <= 'z') || (remaining[end] >= 'A' && remaining[end] <= 'Z') || (remaining[end] >= '0' && remaining[end] <= '9')) {
					end++
				}
				pseudoName = remaining[1:end]
				remaining = remaining[end:]
			}

			// Evaluate pseudo-class
			if !matchPseudoClass(n, pseudoName, pseudoArg) {
				return false
			}

		case '#':
			// ID selector
			end := 1
			for end < len(remaining) && remaining[end] != '.' && remaining[end] != ':' && remaining[end] != '[' && remaining[end] != ' ' {
				end++
			}
			id := remaining[1:end]
			if n.GetAttribute("id") != id {
				return false
			}
			remaining = remaining[end:]

		case '.':
			// Class selector
			end := 1
			for end < len(remaining) && remaining[end] != '.' && remaining[end] != '#' && remaining[end] != ':' && remaining[end] != '[' && remaining[end] != ' ' {
				end++
			}
			class := remaining[1:end]
			classList := splitClasses(n.GetAttribute("class"))
			found := false
			for _, c := range classList {
				if c == class {
					found = true
					break
				}
			}
			if !found {
				return false
			}
			remaining = remaining[end:]

		case '[':
			// Attribute selector
			end := strings.Index(remaining[1:], "]") + 2
			if end < 2 {
				return false
			}
			attrSel := remaining[1 : end-1]
			remaining = remaining[end:]
			if !matchAttributeSelector(n, attrSel) {
				return false
			}

		case '*':
			// Universal selector
			remaining = remaining[1:]

		default:
			// Tag name selector
			end := 0
			for end < len(remaining) && remaining[end] != '.' && remaining[end] != '#' && remaining[end] != ':' && remaining[end] != '[' && remaining[end] != ' ' {
				end++
			}
			tag := strings.ToLower(remaining[:end])
			if tag != "*" && tagName != tag {
				return false
			}
			remaining = remaining[end:]
		}
	}

	return true
}

// matchPseudoClass evaluates a pseudo-class
func matchPseudoClass(n *Node, pseudoName, pseudoArg string) bool {
	switch strings.ToLower(pseudoName) {
	case "not":
		// :not(.foo, .bar) handles comma-separated selectors
		selectors := strings.Split(pseudoArg, ",")
		for _, innerSel := range selectors {
			innerSel = strings.TrimSpace(innerSel)
			// If ANY of the inner selectors match, the :not() fails
			if matchSelector(n, innerSel) {
				return false
			}
		}
		return true
	case "first-child":
		if n.Parent == nil {
			return false
		}
		for _, child := range n.Parent.Children {
			if child.Type == NodeElement {
				return child == n
			}
		}
		return false
	case "last-child":
		if n.Parent == nil {
			return false
		}
		var lastElement *Node
		for _, child := range n.Parent.Children {
			if child.Type == NodeElement {
				lastElement = child
			}
		}
		return lastElement == n
	case "only-child":
		if n.Parent == nil {
			return false
		}
		count := 0
		for _, child := range n.Parent.Children {
			if child.Type == NodeElement {
				count++
			}
		}
		return count == 1
	case "disabled":
		// :disabled matches form elements with disabled attribute
		tag := strings.ToLower(n.TagName)
		switch tag {
		case "input", "button", "select", "textarea", "fieldset", "optgroup", "option":
			return hasAttribute(n, "disabled")
		}
		return false
	case "enabled":
		// :enabled matches form elements that are NOT disabled
		tag := strings.ToLower(n.TagName)
		switch tag {
		case "input", "button", "select", "textarea", "fieldset", "optgroup", "option":
			return !hasAttribute(n, "disabled")
		}
		return true
	case "checked":
		// :checked matches checkbox/radio with checked attribute
		tag := strings.ToLower(n.TagName)
		if tag == "input" {
			typ := strings.ToLower(getAttribute(n, "type"))
			if typ == "checkbox" || typ == "radio" {
				return hasAttribute(n, "checked")
			}
		}
		return false
	case "focus-visible":
		// :focus-visible: matches elements that should show focus indicator
		return hasAttribute(n, "tabindex") || hasAttribute(n, "contenteditable")
	case "lang":
		// :lang(en) matches elements with matching lang attribute
		if pseudoArg == "" {
			return false
		}
		lang := strings.ToLower(getAttribute(n, "lang"))
		code := strings.ToLower(pseudoArg)
		return lang == code || strings.HasPrefix(lang, code+"-")
	default:
		// Unknown pseudo-class - treat as matching for forward compatibility
		return true
	}
}

// hasAttribute returns true if the node has the given attribute (value doesn't matter)
func hasAttribute(n *Node, key string) bool {
	for _, attr := range n.Attributes {
		if attr.Key == key {
			return true
		}
	}
	return false
}

// getAttribute returns the value of an attribute or empty string
func getAttribute(n *Node, key string) string {
	for _, attr := range n.Attributes {
		if attr.Key == key {
			return attr.Value
		}
	}
	return ""
}

// matchAttributeSelector checks if an attribute selector matches
func matchAttributeSelector(n *Node, attrSel string) bool {
	var attrName, op, value string
	for i, c := range attrSel {
		if c == '=' && i > 0 {
			attrName = attrSel[:i]
			value = attrSel[i+1:]
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			op = "="
			break
		}
	}
	if attrName == "" {
		attrName = attrSel
		op = ""
		value = ""
	}

	// Check if attribute exists/matches
	attrValue := ""
	hasAttr := false
	for _, attr := range n.Attributes {
		if attr.Key == attrName {
			hasAttr = true
			attrValue = attr.Value
			break
		}
	}

	switch op {
	case "":
		return hasAttr
	case "=":
		if !hasAttr {
			return false
		}
		return attrValue == value
	case "~=":
		if !hasAttr {
			return false
		}
		for _, c := range strings.Fields(attrValue) {
			if c == value {
				return true
			}
		}
		return false
	case "|=":
		if !hasAttr {
			return false
		}
		return attrValue == value || strings.HasPrefix(attrValue, value+"-")
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

// SetAttribute sets or replaces an attribute on the node.
func (n *Node) SetAttribute(key, value string) {
	for i, attr := range n.Attributes {
		if attr.Key == key {
			n.Attributes[i].Value = value
			return
		}
	}
	n.Attributes = append(n.Attributes, Attribute{Key: key, Value: value})
}

// RemoveAttribute removes an attribute from the node.
func (n *Node) RemoveAttribute(key string) {
	for i, attr := range n.Attributes {
		if attr.Key == key {
			n.Attributes = append(n.Attributes[:i], n.Attributes[i+1:]...)
			return
		}
	}
}

// HasAttribute returns true if the node has the given attribute.
func (n *Node) HasAttribute(key string) bool {
	for _, attr := range n.Attributes {
		if attr.Key == key {
			return true
		}
	}
	return false
}

// GetElementsByClassName returns all descendant elements with the given class name.
func (n *Node) GetElementsByClassName(className string) []*Node {
	var results []*Node
	n.getElementsByClassName(className, &results)
	return results
}

func (n *Node) getElementsByClassName(className string, results *[]*Node) {
	if n.Type == NodeElement {
		class := n.GetAttribute("class")
		for _, c := range splitClasses(class) {
			if c == className {
				*results = append(*results, n)
				break
			}
		}
	}
	for _, child := range n.Children {
		child.getElementsByClassName(className, results)
	}
}

// GetElementsByTagName returns all descendant elements with the given tag name.
func (n *Node) GetElementsByTagName(tagName string) []*Node {
	var results []*Node
	n.getElementsByTagName(strings.ToLower(tagName), &results)
	return results
}

func (n *Node) getElementsByTagName(tagName string, results *[]*Node) {
	if n.Type == NodeElement {
		if strings.ToLower(n.TagName) == tagName {
			*results = append(*results, n)
		}
	}
	for _, child := range n.Children {
		child.getElementsByTagName(tagName, results)
	}
}

// QuerySelectorAll returns all descendant elements matching the CSS selector.
// Supports: tagname, #id, .class, [attr], [attr=value], [attr~=value], [attr|=value],
// and comma-separated selectors.
func (n *Node) QuerySelectorAll(selector string) []*Node {
	var results []*Node
	// Split comma-separated selectors
	for _, sel := range strings.Split(selector, ",") {
		sel = strings.TrimSpace(sel)
		if sel == "" {
			continue
		}
		n.querySelectorAll(sel, &results)
	}
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

// QuerySelector returns the first descendant element matching the CSS selector.
func (n *Node) QuerySelector(selector string) *Node {
	nodes := n.QuerySelectorAll(selector)
	if len(nodes) > 0 {
		return nodes[0]
	}
	return nil
}

// GetElementById returns the first descendant element with the given id attribute.
func (n *Node) GetElementById(id string) *Node {
	if n.Type == NodeElement {
		if n.GetAttribute("id") == id {
			return n
		}
	}
	for _, child := range n.Children {
		if found := child.GetElementById(id); found != nil {
			return found
		}
	}
	return nil
}

// DOMRect represents a rectangle in viewport coordinates (as returned by getBoundingClientRect).
type DOMRect struct {
	X, Y      float64 // viewport-relative x and y coordinates
	Width     float64 // width of the border box
	Height    float64 // height of the border box
	Top       float64 // distance from top edge of viewport to top edge of element
	Right     float64 // distance from left edge of viewport to right edge of element
	Bottom    float64 // distance from top edge of viewport to bottom edge of element
	Left      float64 // distance from left edge of viewport to left edge of element
}

// NewDOMRect creates a DOMRect with all properties properly computed from x, y, width, height.
func NewDOMRect(x, y, width, height float64) *DOMRect {
	return &DOMRect{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Top:    y,
		Right:  x + width,
		Bottom: y + height,
		Left:   x,
	}
}

// GetBoundingClientRect returns the bounding box of this node relative to the viewport.
// It traverses the provided root Box tree to find the Box corresponding to this node.
// Returns nil if no box is found for this node.
func (n *Node) GetBoundingClientRect(rootBox interface{}) *DOMRect {
	if n == nil {
		return nil
	}

	// Try to find the box for this node by traversing the box tree
	// The rootBox is expected to be a *layout.Box, but we use interface{} to avoid import cycle
	// We use recursive approach similar to GetElementById

	var findBox func(box interface{}, node *Node) interface{}
	findBox = func(box interface{}, node *Node) interface{} {
		if box == nil {
			return nil
		}

		// Use reflection to access Box struct fields
		// Box has: Type, Node *html.Node, ContentX, ContentY, ContentW, ContentH float64, Children []*Box
		b := box

		// Get the Node field from Box using type assertion approach
		// Since we can't import layout, we use a convention: the box has these methods/fields
		type boxInterface interface {
			GetNode() *Node
			GetContentX() float64
			GetContentY() float64
			GetContentW() float64
			GetContentH() float64
			GetChildren() []interface{}
		}

		if bi, ok := b.(boxInterface); ok {
			if bi.GetNode() == node {
				return b
			}
			for _, child := range bi.GetChildren() {
				if found := findBox(child, node); found != nil {
					return found
				}
			}
		}
		return nil
	}

	foundBox := findBox(rootBox, n)
	if foundBox == nil {
		return nil
	}

	// Extract dimensions from found box
	type boxInterface interface {
		GetContentX() float64
		GetContentY() float64
		GetContentW() float64
		GetContentH() float64
	}

	bi := foundBox.(boxInterface)
	x := bi.GetContentX()
	y := bi.GetContentY()
	w := bi.GetContentW()
	h := bi.GetContentH()

	return NewDOMRect(x, y, w, h)
}

// BoxNodeInterface is implemented by layout.Box to allow GetBoundingClientRect to work
// without importing the layout package (avoiding circular dependency).
type BoxNodeInterface interface {
	GetNode() *Node
	GetContentX() float64
	GetContentY() float64
	GetContentW() float64
	GetContentH() float64
	GetChildren() []BoxNodeInterface
}

// GetBoundingClientRectWithBox returns the bounding box of this node relative to the viewport.
// This version accepts the root layout box directly and uses the layout.Box type.
func (n *Node) GetBoundingClientRectWithBox(rootBox BoxNodeInterface) *DOMRect {
	if n == nil || rootBox == nil {
		return nil
	}

	var findBox func(box BoxNodeInterface) BoxNodeInterface
	findBox = func(box BoxNodeInterface) BoxNodeInterface {
		if box.GetNode() == n {
			return box
		}
		for _, child := range box.GetChildren() {
			if found := findBox(child); found != nil {
				return found
			}
		}
		return nil
	}

	foundBox := findBox(rootBox)
	if foundBox == nil {
		return nil
	}

	x := foundBox.GetContentX()
	y := foundBox.GetContentY()
	w := foundBox.GetContentW()
	h := foundBox.GetContentH()

	return NewDOMRect(x, y, w, h)
}

// ClassList returns a ClassList-like view of the element's classes.
// Each returned ClassEntry has Add, Remove, Toggle, Contains methods.
func (n *Node) ClassList() *ClassList {
	return &ClassList{node: n}
}

// ClassList provides DOM classList API for a node.
type ClassList struct {
	node *Node
}

// getClasses returns the current class list as a slice.
func (cl *ClassList) getClasses() []string {
	class := cl.node.GetAttribute("class")
	if class == "" {
		return nil
	}
	return splitClasses(class)
}

// String returns the class attribute value.
func (cl *ClassList) String() string {
	return cl.node.GetAttribute("class")
}

// Contains returns true if the class is present.
func (cl *ClassList) Contains(class string) bool {
	for _, c := range cl.getClasses() {
		if c == class {
			return true
		}
	}
	return false
}

// Add adds the given class names.
func (cl *ClassList) Add(classes ...string) {
	current := cl.getClasses()
	for _, c := range classes {
		if !containsString(current, c) {
			current = append(current, c)
		}
	}
	cl.node.SetAttribute("class", strings.Join(current, " "))
}

// Remove removes the given class names.
func (cl *ClassList) Remove(classes ...string) {
	current := cl.getClasses()
	var result []string
	for _, c := range current {
		if !containsString(classes, c) {
			result = append(result, c)
		}
	}
	cl.node.SetAttribute("class", strings.Join(result, " "))
}

// Toggle removes the class if present, adds it if not.
func (cl *ClassList) Toggle(class string) {
	if cl.Contains(class) {
		cl.Remove(class)
	} else {
		cl.Add(class)
	}
}

// Replace replaces one class with another.
func (cl *ClassList) Replace(oldClass, newClass string) {
	current := cl.getClasses()
	for i, c := range current {
		if c == oldClass {
			current[i] = newClass
			break
		}
	}
	cl.node.SetAttribute("class", strings.Join(current, " "))
}

// Item returns the class at the given index, or "".
func (cl *ClassList) Item(index int) string {
	classes := cl.getClasses()
	if index >= 0 && index < len(classes) {
		return classes[index]
	}
	return ""
}

// Length returns the number of classes.
func (cl *ClassList) Length() int {
	return len(cl.getClasses())
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// textContent returns the text content of the node and all descendants.
func (n *Node) textContent() string {
	if n.Type == NodeText {
		return n.Data
	}
	var text string
	for _, child := range n.Children {
		text += child.textContent()
	}
	return text
}

// setTextContent sets the text content, removing all children and adding a text node.
func (n *Node) setTextContent(text string) {
	// Remove all children
	n.Children = nil
	if text != "" {
		n.Children = []*Node{NewText(text)}
	}
}

// TextContent gets/sets the text content of the node.
func (n *Node) TextContent() string {
	return n.textContent()
}

// SetTextContent sets the text content of the node.
func (n *Node) SetTextContent(text string) {
	n.setTextContent(text)
}

// isHidden returns true if the element is hidden via CSS (display:none or visibility:hidden).
func (n *Node) isHidden() bool {
	if n.Type != NodeElement {
		return false
	}

	// Check visibility:hidden
	visibility := n.GetAttribute("visibility")
	if visibility == "hidden" {
		return true
	}

	// Check style attribute for visibility:hidden
	style := n.GetAttribute("style")
	if strings.Contains(style, "visibility:hidden") || strings.Contains(style, "visibility: hidden") {
		return true
	}

	// Check display:none
	display := n.GetAttribute("display")
	if display == "none" {
		return true
	}

	// Check style attribute for display:none
	if strings.Contains(style, "display:none") || strings.Contains(style, "display: none") {
		return true
	}

	return false
}

// isBlockElement returns true if the element is a block-level element.
func (n *Node) isBlockElement() bool {
	if n.Type != NodeElement {
		return false
	}
	tag := strings.ToLower(n.TagName)
	switch tag {
	case "address", "blockquote", "br", "center", "dir", "div", "dl", "dt",
		"fieldset", "figure", "form", "h1", "h2", "h3", "h4", "h5", "h6",
		"hr", "isindex", "li", "main", "menu", "nav", "ol", "p", "pre",
		"section", "table", "tbody", "td", "tfoot", "th", "thead", "tr",
		"ul":
		return true
	}
	return false
}

// InnerHTML returns the inner HTML of the node.
func (n *Node) InnerHTML() string {
	if n.Type == NodeText {
		return ""
	}
	var html string
	for _, child := range n.Children {
		html += child.OuterHTML()
	}
	return html
}

// SetInnerHTML sets the inner HTML by parsing the given HTML string.
func (n *Node) SetInnerHTML(htmlStr string) {
	// Remove all children
	n.Children = nil
	// Parse the HTML
	tokens := Tokenize([]byte(htmlStr))
	parser := NewParser(tokens)
	frag := parser.Parse()
	// Move children from parsed fragment to this node
	for _, child := range frag.Children {
		child.Parent = n
		n.Children = append(n.Children, child)
	}
}

// OuterHTML returns the outer HTML of the node.
func (n *Node) OuterHTML() string {
	if n.Type == NodeText {
		return n.Data
	}
	if n.Type == NodeComment {
		return "<!--" + n.Data + "-->"
	}
	var html string
	html += "<" + n.TagName
	for _, attr := range n.Attributes {
		html += " " + attr.Key + "=\"" + attr.Value + "\""
	}
	html += ">"
	for _, child := range n.Children {
		html += child.OuterHTML()
	}
	html += "</" + n.TagName + ">"
	return html
}

// RemoveChild removes the child node from the parent's children list.
func (n *Node) RemoveChild(child *Node) error {
	for i, c := range n.Children {
		if c == child {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			child.Parent = nil
			return nil
		}
	}
	return fmt.Errorf("node not found")
}

// InsertBefore inserts a node before the reference node in the parent's children list.
// If ref is nil, inserts at the end.
func (n *Node) InsertBefore(newNode, ref *Node) {
	if ref == nil {
		n.AppendChild(newNode)
		return
	}
	for i, c := range n.Children {
		if c == ref {
			newNode.Parent = n
			// Insert at position i
			n.Children = append(n.Children[:i], append([]*Node{newNode}, n.Children[i:]...)...)
			return
		}
	}
}

// ReplaceChild replaces a child node with a new node.
func (n *Node) ReplaceChild(newNode, oldNode *Node) {
	for i, c := range n.Children {
		if c == oldNode {
			newNode.Parent = n
			n.Children[i] = newNode
			oldNode.Parent = nil
			return
		}
	}
}

// CloneNode creates a shallow copy of the node.
func (n *Node) CloneNode() *Node {
	clone := &Node{
		Type:       n.Type,
		TagName:    n.TagName,
		Data:       n.Data,
		Attributes: make([]Attribute, len(n.Attributes)),
		Parent:     nil,
		Children:   nil,
	}
	copy(clone.Attributes, n.Attributes)
	return clone
}

// CloneNodeDeep creates a deep copy of the node and all descendants.
func (n *Node) CloneNodeDeep() *Node {
	clone := n.CloneNode()
	for _, child := range n.Children {
		clonedChild := child.CloneNodeDeep()
		clonedChild.Parent = clone
		clone.Children = append(clone.Children, clonedChild)
	}
	return clone
}

// dataset returns a map of data-* attributes for easy access.
func (n *Node) dataset() map[string]string {
	result := make(map[string]string)
	for _, attr := range n.Attributes {
		if strings.HasPrefix(attr.Key, "data-") {
			result[attr.Key[5:]] = attr.Value
		}
	}
	return result
}

// Dataset returns a Dataset-like map for data-* attributes.
func (n *Node) Dataset() *Dataset {
	return &Dataset{node: n}
}

// Dataset provides access to data-* attributes.
type Dataset struct {
	node *Node
}

// Get returns the value of a data-* attribute (without data- prefix).
func (d *Dataset) Get(name string) string {
	return d.node.GetAttribute("data-" + name)
}

// Set sets the value of a data-* attribute.
func (d *Dataset) Set(name, value string) {
	d.node.SetAttribute("data-"+name, value)
}

// Remove removes a data-* attribute.
func (d *Dataset) Remove(name string) {
	d.node.RemoveAttribute("data-" + name)
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

// InsideStyleOrScript returns true if this node is a descendant of a <style> or <script> element.
func (n *Node) InsideStyleOrScript() bool {
	current := n.Parent
	for current != nil {
		if current.Type == NodeElement {
			tag := current.TagName
			if tag == "style" || tag == "script" {
				return true
			}
		}
		current = current.Parent
	}
	return false
}

// InnerText returns the rendered text content of the node.
// Unlike TextContent, InnerText respects CSS styling and will not return
// text from elements with display:none or visibility:hidden.
// It also adds newlines between block elements for proper rendering.
func (n *Node) InnerText() string {
	if n.Type == NodeText {
		return n.Data
	}

	if n.Type == NodeElement {
		// Skip hidden elements
		if n.isHidden() {
			return ""
		}
	}

	var text string
	prevWasBlock := false
	for _, child := range n.Children {
		// Skip hidden elements before processing
		if child.Type == NodeElement && child.isHidden() {
			prevWasBlock = false // Reset so next visible block doesn't add extra newline
			continue
		}

		// Skip <style> and <script> elements and their content entirely
		if child.Type == NodeElement && (child.TagName == "style" || child.TagName == "script") {
			continue
		}

		// Skip text nodes inside style/script
		if child.Type == NodeText && child.InsideStyleOrScript() {
			continue
		}

		isBlock := child.Type == NodeElement && child.isBlockElement()
		childText := child.InnerText()

		// For block elements that produce empty text (like <br/>), add newline directly
		if isBlock && childText == "" {
			if text != "" && !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
			prevWasBlock = false
			continue
		}

		if childText == "" {
			continue
		}

		// Add newline BEFORE block elements (not after, to avoid trailing newlines)
		if prevWasBlock {
			if text != "" && !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
		}

		text += childText
		prevWasBlock = isBlock
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
