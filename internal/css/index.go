package css

import (
	"strings"

	"github.com/nickcoury/ViBrowsing/internal/html"
)

// SelectorIndex provides O(1) lookup for CSS selectors by building
// indexes for class, id, and tag name lookups.
type SelectorIndex struct {
	byClass map[string][]*html.Node // class name -> nodes
	byID    map[string][]*html.Node // id -> nodes
	byTag   map[string][]*html.Node // tag name -> nodes
}

// BuildSelectorIndex walks the DOM tree and builds index maps for
// fast lookup of elements by class, id, and tag name.
func BuildSelectorIndex(doc *html.Node) *SelectorIndex {
	index := &SelectorIndex{
		byClass: make(map[string][]*html.Node),
		byID:    make(map[string][]*html.Node),
		byTag:   make(map[string][]*html.Node),
	}
	walkNode(doc, index)
	return index
}

// walkNode recursively traverses the DOM tree and populates the index.
func walkNode(node *html.Node, index *SelectorIndex) {
	if node == nil {
		return
	}

	if node.Type == html.NodeElement {
		// Index by class (elements may have multiple space-separated classes)
		class := node.GetAttribute("class")
		if class != "" {
			for _, c := range splitClasses(class) {
				index.byClass[c] = append(index.byClass[c], node)
			}
		}

		// Index by id
		id := node.GetAttribute("id")
		if id != "" {
			index.byID[id] = append(index.byID[id], node)
		}

		// Index by tag name (case-insensitive, store lowercase)
		tagName := strings.ToLower(node.TagName)
		index.byTag[tagName] = append(index.byTag[tagName], node)
	}

	// Recursively index children
	for _, child := range node.Children {
		walkNode(child, index)
	}
}

// GetElementsByClass returns all nodes with the given class name.
// Returns an empty slice if no elements match.
func GetElementsByClass(index *SelectorIndex, class string) []*html.Node {
	if index == nil || index.byClass == nil {
		return []*html.Node{}
	}
	nodes := index.byClass[class]
	if nodes == nil {
		return []*html.Node{}
	}
	return nodes
}

// GetElementByID returns the first node with the given id attribute.
// Returns nil if no element with that id is found.
func GetElementByID(index *SelectorIndex, id string) *html.Node {
	if index == nil || index.byID == nil {
		return nil
	}
	nodes := index.byID[id]
	if len(nodes) == 0 {
		return nil
	}
	return nodes[0]
}

// GetElementsByTag returns all nodes with the given tag name (case-insensitive).
// Returns an empty slice if no elements match.
func GetElementsByTag(index *SelectorIndex, tag string) []*html.Node {
	if index == nil || index.byTag == nil {
		return []*html.Node{}
	}
	tagName := strings.ToLower(tag)
	nodes := index.byTag[tagName]
	if nodes == nil {
		return []*html.Node{}
	}
	return nodes
}
