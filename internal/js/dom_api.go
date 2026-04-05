package js

import (
	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/html"
)

// DOM APIs exposed to JavaScript

// GetComputedStyle returns the computed style for a DOM element.
// This is the global window.getComputedStyle() function.
// It returns a map of CSS property names to their computed values.
func GetComputedStyle(element interface{}) map[string]string {
	var node *html.Node

	switch e := element.(type) {
	case *html.Node:
		node = e
	case html.Node:
		node = &e
	default:
		return map[string]string{}
	}

	// Use the CSS package's GetComputedStyle
	return css.GetComputedStyle(node, nil)
}

// GetComputedStyleWithRules returns the computed style using provided CSS rules.
func GetComputedStyleWithRules(element interface{}, rules []css.Rule) map[string]string {
	var node *html.Node

	switch e := element.(type) {
	case *html.Node:
		node = e
	case html.Node:
		node = &e
	default:
		return map[string]string{}
	}

	return css.GetComputedStyle(node, rules)
}

// InnerText returns the inner text content of an element.
// This is similar to element.innerText in the DOM API.
// It returns the visible text content, respecting CSS display and visibility.
func InnerText(element interface{}) string {
	var node *html.Node

	switch e := element.(type) {
	case *html.Node:
		node = e
	case html.Node:
		node = &e
	default:
		return ""
	}

	return node.InnerText()
}

// TextContent returns the text content of an element.
// This is the standard DOM textContent property.
func TextContent(element interface{}) string {
	var node *html.Node

	switch e := element.(type) {
	case *html.Node:
		node = e
	case html.Node:
		node = &e
	default:
		return ""
	}

	return node.TextContent()
}

// GetElementById returns the element with the given ID.
// This is the document.getElementById() function.
func GetElementById(doc interface{}, id string) interface{} {
	var root *html.Node

	switch d := doc.(type) {
	case *html.Node:
		root = d
	case html.Node:
		root = &d
	default:
		return nil
	}

	return root.GetElementById(id)
}

// QuerySelector returns the first element matching the CSS selector.
// This is the document.querySelector() function.
func QuerySelector(element interface{}, selector string) interface{} {
	var node *html.Node

	switch e := element.(type) {
	case *html.Node:
		node = e
	case html.Node:
		node = &e
	default:
		return nil
	}

	results := node.QuerySelectorAll(selector)
	if len(results) > 0 {
		return results[0]
	}
	return nil
}

// QuerySelectorAll returns all elements matching the CSS selector.
// This is the document.querySelectorAll() function.
func QuerySelectorAll(element interface{}, selector string) []*html.Node {
	var node *html.Node

	switch e := element.(type) {
	case *html.Node:
		node = e
	case html.Node:
		node = &e
	default:
		return nil
	}

	return node.QuerySelectorAll(selector)
}

// GetElementsByClassName returns all elements with the given class name.
// This is the document.getElementsByClassName() function.
func GetElementsByClassName(element interface{}, className string) []*html.Node {
	var node *html.Node

	switch e := element.(type) {
	case *html.Node:
		node = e
	case html.Node:
		node = &e
	default:
		return nil
	}

	return node.GetElementsByClassName(className)
}

// GetElementsByTagName returns all elements with the given tag name.
// This is the document.getElementsByTagName() function.
func GetElementsByTagName(element interface{}, tagName string) []*html.Node {
	var node *html.Node

	switch e := element.(type) {
	case *html.Node:
		node = e
	case html.Node:
		node = &e
	default:
		return nil
	}

	return node.GetElementsByTagName(tagName)
}

// ScrollIntoView scrolls the element into view.
// Options can be a boolean or an object with block, inline, behavior properties.
func ScrollIntoView(element interface{}, options ...interface{}) {
	var box interface{ ScrollIntoView(options interface{}) }

	switch e := element.(type) {
	case interface{ ScrollIntoView(interface{}) }:
		box = e
	default:
		return
	}

	if len(options) > 0 {
		box.ScrollIntoView(options[0])
	} else {
		box.ScrollIntoView(nil)
	}
}
