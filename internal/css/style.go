package css

import (
	"strconv"
	"strings"

	"github.com/nickcoury/ViBrowsing/internal/html"
)

// inheritedProperties is the set of CSS properties that inherit from parent to child.
// See https://www.w3.org/TR/CSS21/propidx.html and CSS3 specification.
var inheritedProperties = map[string]bool{
	"azimuth":               true,
	"border-collapse":        true,
	"border-spacing":         true,
	"caption-side":           true,
	"color":                  true,
	"cursor":                 true,
	"direction":              true,
	"elevation":              true,
	"empty-cells":            true,
	"font-family":            true,
	"font-size":              true,
	"font-style":             true,
	"font-variant":           true,
	"font-weight":            true,
	"font":                   true,
	"letter-spacing":         true,
	"line-height":            true,
	"list-style-image":       true,
	"list-style-position":    true,
	"list-style-type":        true,
	"list-style":             true,
	"orphans":                true,
	"overflow-wrap":          true,
	"pitch-range":            true,
	"pitch":                  true,
	"quotes":                 true,
	"richness":               true,
	"speak-header":           true,
	"speak-numeral":          true,
	"speak-punctuation":      true,
	"speak":                  true,
	"speech-rate":            true,
	"stress":                 true,
	"text-align":             true,
	"text-indent":            true,
	"text-transform":         true,
	"visibility":             true,
	"voice-family":           true,
	"volume":                 true,
	"white-space":            true,
	"widows":                 true,
	"word-spacing":           true,
	"writing-mode":           true,
}

// InheritStyle copies inherited properties from the parent style to the child.
// Properties already set on child (author or inline) take precedence.
func InheritStyle(parent, child map[string]string) map[string]string {
	if parent == nil {
		return child
	}
	// Copy child so we don't mutate the original
	result := make(map[string]string)
	for k, v := range child {
		result[k] = v
	}
	for k, v := range parent {
		if _, ok := result[k]; !ok && inheritedProperties[k] {
			result[k] = v
		}
	}
	return result
}

// Specificity represents a CSS selector's specificity (a, b, c).
type Specificity struct {
	ID, Class, Tag int
}

// ComputeStyle computes the cascaded style for an element given its DOM node
// and a list of CSS rules.
func ComputeStyle(tagName string, class string, id string, inlineStyles []Declaration, rules []Rule) map[string]string {
	// Start with user-agent defaults
	props := map[string]string{
		"display":         "inline",
		"visibility":      "visible",
		"color":           "black",
		"background":      "transparent",
		"font-size":       "16px",
		"font-family":     "serif",
		"font-weight":     "normal",
		"font-style":      "normal",
		"text-decoration": "none",
		"margin-top":      "0",
		"margin-right":    "0",
		"margin-bottom":   "0",
		"margin-left":     "0",
		"padding-top":     "0",
		"padding-right":   "0",
		"padding-bottom":  "0",
		"padding-left":    "0",
		"border-width":    "0",
		"border-style":   "none",
		"border-color":   "black",
		"width":           "auto",
		"height":          "auto",
		"text-align":      "left",
		"text-indent":     "0",
		"line-height":     "1.2",
		"letter-spacing":  "0",
		"word-spacing":    "0",
		"text-transform":  "none",
		"text-justify":    "auto",
		"hyphens":         "manual",
		"font-variant":    "normal",
		"unicode-bidi":   "normal",
		"unicode-range":   "U+0-FFFF",
		"direction":      "ltr",
		"writing-mode":   "horizontal-tb",
		"tab-size":       "8",
		"quotes":         "auto",
		"vertical-align":  "baseline",
		"opacity":         "1",
		"white-space":     "normal",
		"overflow":        "visible",
		"overflow-x":      "visible",
		"overflow-y":      "visible",
		"word-wrap":       "normal",
		"position":        "static",
		"top":             "auto",
		"right":           "auto",
		"bottom":          "auto",
		"left":            "auto",
		"float":           "none",
		"z-index":         "auto",
		"flex-direction":  "row",
		"flex-wrap":       "nowrap",
		"flex-flow":       "row nowrap",
		"justify-content": "flex-start",
		"align-items":     "stretch",
		"align-content":   "normal",
		"gap":             "0",
		"order":           "0",
		"flex-grow":       "0",
		"flex-shrink":     "1",
		"flex-basis":      "auto",
		"border-radius":   "0",
		"background-color": "transparent",
		"background-image": "none",
		"background-repeat": "repeat",
		"background-position": "0 0",
		"background-size":  "auto auto",
		"list-style-type":  "disc",
		"list-style-position": "outside",
		"list-style-image": "none",
		"outline-width":   "0",
		"outline-style":   "none",
		"outline-color":   "black",
		"outline-offset":  "0",
		"box-shadow":      "none",
		"cursor":          "auto",
		"transform":       "none",
		"text-shadow":     "none",
		"text-overflow":   "clip",
		"content":         "normal",
		// Animation properties
		"animation-name":        "none",
		"animation-duration":    "0s",
		"animation-timing-function": "ease",
		"animation-delay":       "0s",
		"animation-iteration-count": "1",
		"animation-direction":   "normal",
		"animation-fill-mode":   "none",
		// Image sizing properties
		"aspect-ratio":    "auto",
		"object-fit":       "fill",
		"object-position":  "50% 50%",
		// Filter and effect properties
		"filter":           "none",
		"backdrop-filter":  "none",
		"clip-path":        "none",
		"clip":             "auto",
		// Column properties
		"column-width":        "auto",
		"column-count":        "1",
		"column-gap":          "normal",
		"column-rule-width":   "medium",
		"column-rule-style":  "none",
		"column-rule-color":   "black",
		// Break properties
		"break-inside":     "auto",
		"break-before":     "auto",
		"break-after":      "auto",
		// Transition
		"transition-property":           "none",
		"transition-duration":           "0s",
		"transition-timing-function":    "ease",
		"transition-delay":              "0s",
		// New CSS properties
		"resize":              "none",
		"pointer-events":      "auto",
		"overscroll-behavior": "auto",
		"scroll-behavior":     "auto",
		"text-decoration-line":   "none",
		"text-decoration-color":  "currentColor",
		"text-decoration-style": "solid",
		"text-decoration-thickness": "auto",
		"text-underline-offset": "0",
		"text-decoration-skip-ink": "auto",
		"will-change":         "auto",
		"image-rendering":     "auto",
		"caption-side":        "top",
		"empty-cells":         "show",
		"caret-color":         "auto",
		"appearance":          "auto",
		"contain":             "none",
		"mix-blend-mode":      "normal",
		"hanging-punctuation": "none",
		"font-stretch":         "normal",
		"transform-box":        "view-box",
		"place-items":          "normal",
		"place-self":           "normal",
		"justify-items":        "normal",
		"justify-self":         "auto",
		"user-select":         "auto",
	}

	// Element-specific default styles (HTML5 user agent defaults)
	// These are overridden by any matching CSS rules
	switch tagName {
	case "strong", "b", "th":
		props["font-weight"] = "bold"
	case "em", "i", "cite", "var":
		props["font-style"] = "italic"
	case "code", "kbd", "samp", "pre":
		props["font-family"] = "monospace"
		if tagName == "pre" {
			props["white-space"] = "pre"
		}
	case "blockquote":
		props["display"] = "block"
		props["margin-left"] = "40px"
		props["margin-right"] = "40px"
		props["font-style"] = "italic"
	case "address":
		props["display"] = "block"
		props["font-style"] = "italic"
	case "header", "footer", "nav", "section", "article", "aside", "main", "figure", "figcaption", "details", "summary":
		props["display"] = "block"
	case "noscript":
		props["display"] = "block"
	case "hr":
		props["display"] = "block"
		props["border-width"] = "1px"
		props["border-style"] = "solid"
		props["border-color"] = "gray"
		props["margin-top"] = "8px"
		props["margin-bottom"] = "8px"
		props["height"] = "1px" // Use height for hr thickness since it's a void element
	case "input":
		props["display"] = "inline-block"
		props["border-width"] = "1px"
		props["border-style"] = "solid"
		props["border-color"] = "gray"
		props["padding-top"] = "2px"
		props["padding-bottom"] = "2px"
		props["padding-left"] = "4px"
		props["padding-right"] = "4px"
	case "button":
		props["display"] = "inline-block"
		props["border-width"] = "2px"
		props["border-style"] = "solid"
		props["border-color"] = "#444"
		props["padding-top"] = "4px"
		props["padding-bottom"] = "4px"
		props["padding-left"] = "12px"
		props["padding-right"] = "12px"
		props["background"] = "#f0f0f0"
	case "select":
		props["display"] = "inline-block"
		props["border-width"] = "1px"
		props["border-style"] = "solid"
		props["border-color"] = "gray"
		props["padding-top"] = "2px"
		props["padding-bottom"] = "2px"
		props["padding-left"] = "4px"
		props["padding-right"] = "4px"
		props["background"] = "white"
	case "textarea":
		props["display"] = "inline-block"
		props["border-width"] = "1px"
		props["border-style"] = "solid"
		props["border-color"] = "gray"
		props["padding-top"] = "4px"
		props["padding-bottom"] = "4px"
		props["padding-left"] = "4px"
		props["padding-right"] = "4px"
		props["background"] = "white"
		props["width"] = "200px"
		props["height"] = "100px"
	case "video", "audio":
		props["display"] = "inline-block"
		props["background"] = "#000"
		props["color"] = "#fff"
	case "del", "s", "strike":
		// Deleted text — typically rendered with strikethrough
		props["text-decoration"] = "line-through"
	case "ins", "u":
		// Inserted text — typically rendered with underline
		props["text-decoration"] = "underline"
	}

	// Apply rules in order (later rules win for same specificity)
	for _, rule := range rules {
		if matchSelector(tagName, class, id, rule.Selector) {
			for _, decl := range rule.Declarations {
				applyDecl(props, decl)
			}
		}
	}

	// Inline styles have highest priority
	for _, decl := range inlineStyles {
		applyDecl(props, decl)
	}

	return props
}

// ComputeStyleForNode computes the cascaded style for an HTML node element.
// This version supports attribute selectors by having access to the full node.
func ComputeStyleForNode(node *html.Node, rules []Rule) map[string]string {
	if node == nil || node.Type != html.NodeElement {
		return map[string]string{}
	}

	tagName := node.TagName
	class := node.GetAttribute("class")
	id := node.GetAttribute("id")
	inlineDecls := ParseInline(node.GetAttribute("style"))

	// Start with user-agent defaults
	props := map[string]string{
		"display":         "inline",
		"visibility":      "visible",
		"color":           "black",
		"background":      "transparent",
		"font-size":       "16px",
		"font-family":     "serif",
		"font-weight":     "normal",
		"font-style":      "normal",
		"text-decoration": "none",
		"margin-top":      "0",
		"margin-right":    "0",
		"margin-bottom":   "0",
		"margin-left":     "0",
		"padding-top":     "0",
		"padding-right":   "0",
		"padding-bottom":  "0",
		"padding-left":    "0",
		"border-width":    "0",
		"border-style":   "none",
		"border-color":   "black",
		"width":           "auto",
		"height":          "auto",
		"text-align":      "left",
		"text-indent":     "0",
		"line-height":     "1.2",
		"letter-spacing":  "0",
		"word-spacing":    "0",
		"text-transform":  "none",
		"text-justify":    "auto",
		"hyphens":         "manual",
		"font-variant":    "normal",
		"unicode-bidi":   "normal",
		"unicode-range":   "U+0-FFFF",
		"direction":      "ltr",
		"writing-mode":   "horizontal-tb",
		"tab-size":       "8",
		"quotes":         "auto",
		"vertical-align":  "baseline",
		"opacity":         "1",
		"white-space":     "normal",
		"overflow":        "visible",
		"overflow-x":      "visible",
		"overflow-y":      "visible",
		"word-wrap":       "normal",
		"position":        "static",
		"top":             "auto",
		"right":           "auto",
		"bottom":          "auto",
		"left":            "auto",
		"float":           "none",
		"z-index":         "auto",
		"flex-direction":  "row",
		"flex-wrap":       "nowrap",
		"flex-flow":       "row nowrap",
		"justify-content": "flex-start",
		"align-items":     "stretch",
		"align-content":   "normal",
		"gap":             "0",
		"order":           "0",
		"flex-grow":       "0",
		"flex-shrink":     "1",
		"flex-basis":      "auto",
		"border-radius":   "0",
		"background-color": "transparent",
		"background-image": "none",
		"background-repeat": "repeat",
		"background-position": "0 0",
		"background-size":  "auto auto",
		"list-style-type":  "disc",
		"list-style-position": "outside",
		"list-style-image": "none",
		"outline-width":   "0",
		"outline-style":   "none",
		"outline-color":   "black",
		"outline-offset":  "0",
		"box-shadow":      "none",
		"cursor":          "auto",
		"transform":       "none",
		"text-shadow":     "none",
		"text-overflow":   "clip",
		"content":         "normal",
		"animation-name":        "none",
		"animation-duration":    "0s",
		"animation-timing-function": "ease",
		"animation-delay":       "0s",
		"animation-iteration-count": "1",
		"animation-direction":   "normal",
		"animation-fill-mode":   "none",
		"aspect-ratio":    "auto",
		"object-fit":       "fill",
		"object-position":  "50% 50%",
		"filter":           "none",
		"backdrop-filter":  "none",
		"clip-path":        "none",
		"clip":             "auto",
		"column-width":        "auto",
		"column-count":        "1",
		"column-gap":          "normal",
		"column-rule-width":   "medium",
		"column-rule-style":  "none",
		"column-rule-color":   "black",
		"break-inside":     "auto",
		"break-before":     "auto",
		"break-after":      "auto",
		"transition-property":           "none",
		"transition-duration":           "0s",
		"transition-timing-function":    "ease",
		"transition-delay":              "0s",
		"resize":              "none",
		"pointer-events":      "auto",
		"overscroll-behavior": "auto",
		"scroll-behavior":     "auto",
		"text-decoration-line":   "none",
		"text-decoration-color":  "currentColor",
		"text-decoration-style": "solid",
		"text-decoration-thickness": "auto",
		"text-underline-offset": "0",
		"text-decoration-skip-ink": "auto",
		"will-change":         "auto",
		"image-rendering":     "auto",
		"caption-side":        "top",
		"empty-cells":         "show",
		"caret-color":         "auto",
		"appearance":          "auto",
		"contain":             "none",
		"mix-blend-mode":      "normal",
		"hanging-punctuation": "none",
		"font-stretch":         "normal",
		"transform-box":        "view-box",
		"place-items":          "normal",
		"place-self":           "normal",
		"justify-items":        "normal",
		"justify-self":         "auto",
		"user-select":         "auto",
	}

	// Element-specific defaults
	switch tagName {
	case "html":
		props["display"] = "block"
	case "body":
		props["display"] = "block"
		props["margin-top"] = "8px"
		props["margin-right"] = "8px"
		props["margin-bottom"] = "8px"
		props["margin-left"] = "8px"
	case "strong", "b", "th":
		props["font-weight"] = "bold"
	case "em", "i", "cite", "var":
		props["font-style"] = "italic"
	case "code", "kbd", "samp", "pre":
		props["font-family"] = "monospace"
		if tagName == "pre" {
			props["white-space"] = "pre"
		}
	case "h1":
		props["display"] = "block"
		props["font-size"] = "2em"
		props["font-weight"] = "bold"
		props["margin-top"] = "0.67em"
		props["margin-bottom"] = "0.67em"
	case "h2":
		props["display"] = "block"
		props["font-size"] = "1.5em"
		props["font-weight"] = "bold"
		props["margin-top"] = "0.83em"
		props["margin-bottom"] = "0.83em"
	case "h3":
		props["display"] = "block"
		props["font-size"] = "1.17em"
		props["font-weight"] = "bold"
		props["margin-top"] = "1em"
		props["margin-bottom"] = "1em"
	case "h4":
		props["display"] = "block"
		props["font-size"] = "1em"
		props["font-weight"] = "bold"
		props["margin-top"] = "1.33em"
		props["margin-bottom"] = "1.33em"
	case "h5":
		props["display"] = "block"
		props["font-size"] = "0.83em"
		props["font-weight"] = "bold"
		props["margin-top"] = "1.67em"
		props["margin-bottom"] = "1.67em"
	case "h6":
		props["display"] = "block"
		props["font-size"] = "0.67em"
		props["font-weight"] = "bold"
		props["margin-top"] = "2.33em"
		props["margin-bottom"] = "2.33em"
	case "p":
		props["display"] = "block"
		props["margin-top"] = "1em"
		props["margin-bottom"] = "1em"
	case "div", "header", "footer", "nav", "section", "article", "aside", "main", "figure", "figcaption", "details", "summary":
		props["display"] = "block"
	case "ul", "ol":
		props["display"] = "block"
	case "li":
		props["display"] = "list-item"
	case "blockquote":
		props["display"] = "block"
		props["margin-left"] = "40px"
		props["margin-right"] = "40px"
		props["font-style"] = "italic"
	case "address":
		props["display"] = "block"
		props["font-style"] = "italic"
	case "noscript":
		props["display"] = "block"
	case "hr":
		props["display"] = "block"
		props["border-width"] = "1px"
		props["border-style"] = "solid"
		props["border-color"] = "gray"
		props["margin-top"] = "8px"
		props["margin-bottom"] = "8px"
		props["height"] = "1px"
	case "input":
		props["display"] = "inline-block"
		props["border-width"] = "1px"
		props["border-style"] = "solid"
		props["border-color"] = "gray"
		props["padding-top"] = "2px"
		props["padding-bottom"] = "2px"
		props["padding-left"] = "4px"
		props["padding-right"] = "4px"
	case "button":
		props["display"] = "inline-block"
		props["border-width"] = "2px"
		props["border-style"] = "solid"
		props["border-color"] = "#444"
		props["padding-top"] = "4px"
		props["padding-bottom"] = "4px"
		props["padding-left"] = "12px"
		props["padding-right"] = "12px"
		props["background"] = "#f0f0f0"
	case "select":
		props["display"] = "inline-block"
		props["border-width"] = "1px"
		props["border-style"] = "solid"
		props["border-color"] = "gray"
		props["padding-top"] = "2px"
		props["padding-bottom"] = "2px"
		props["padding-left"] = "4px"
		props["padding-right"] = "4px"
		props["background"] = "white"
	case "textarea":
		props["display"] = "inline-block"
		props["border-width"] = "1px"
		props["border-style"] = "solid"
		props["border-color"] = "gray"
		props["padding-top"] = "4px"
		props["padding-bottom"] = "4px"
		props["padding-left"] = "4px"
		props["padding-right"] = "4px"
		props["background"] = "white"
		props["width"] = "200px"
		props["height"] = "100px"
	case "video", "audio":
		props["display"] = "inline-block"
		props["background"] = "#000"
		props["color"] = "#fff"
	case "del", "s", "strike":
		props["text-decoration"] = "line-through"
	case "ins", "u":
		props["text-decoration"] = "underline"
	case "img":
		props["display"] = "inline-block"
	case "a":
		props["display"] = "inline"
		props["color"] = "blue"
		props["text-decoration"] = "underline"
	case "span", "sup", "sub", "small", "mark":
		props["display"] = "inline"
	}

	for _, rule := range rules {
		// Use node-aware selector matching for attribute selectors
		if matchSelector(tagName, class, id, rule.Selector) {
			// Also try node-aware matching for attribute selectors
			if MatchNodeSelector(node, rule.Selector) {
				for _, decl := range rule.Declarations {
					applyDecl(props, decl)
				}
			}
		}
	}

	// Inline styles have highest priority
	for _, decl := range inlineDecls {
		applyDecl(props, decl)
	}

	return props
}

// GetComputedStyle returns the computed style for an element, including all
// CSS properties with their final values after the cascade and inheritance.
// It takes a DOM node and a list of CSS rules, then walks up the parent chain
// to apply inherited properties.
func GetComputedStyle(node *html.Node, rules []Rule) map[string]string {
	if node == nil {
		return map[string]string{}
	}

	// Build up inherited properties by walking from root to immediate parent
	// Each step applies the element's own style on top of inherited values
	props := map[string]string{}
	ancestor := node.Parent

	// First pass: collect ancestors in order from root to immediate parent
	var ancestors []*html.Node
	for anc := ancestor; anc != nil; anc = anc.Parent {
		if anc.Type == html.NodeElement {
			ancestors = append(ancestors, anc)
		}
	}

	// Reverse to process from root to immediate parent
	// props accumulates inherited values as we go
	for i := len(ancestors) - 1; i >= 0; i-- {
		anc := ancestors[i]
		// Compute style for this ancestor (their own cascade, without inheritance from props)
		ancStyle := ComputeStyleForNode(anc, rules)
		// Inherit from it
		inheritCSSProps(props, ancStyle)
	}

	// Finally, apply this element's own style (defaults + rules + inline)
	ownStyle := ComputeStyleForNode(node, rules)
	for k, v := range ownStyle {
		props[k] = v
	}

	return props
}

// inheritCSSProps copies inheritable CSS properties from parentStyle to props.
// Only properties that are not yet set in props (or are explicitly "inherit")
// will be inherited from parentStyle.
func inheritCSSProps(props, parentStyle map[string]string) {
	// List of CSS properties that are inherited by default
	inheritableProps := map[string]bool{
		"color":                        true,
		"font-family":                  true,
		"font-size":                    true,
		"font-style":                   true,
		"font-weight":                  true,
		"font-variant":                 true,
		"font-stretch":                 true,
		"font":                         true,
		"letter-spacing":               true,
		"word-spacing":                 true,
		"line-height":                  true,
		"text-align":                   true,
		"text-indent":                  true,
		"text-transform":               true,
		"text-justify":                 true,
		"direction":                    true,
		"unicode-bidi":                 true,
		"visibility":                   true,
		"quotes":                       true,
		"list-style-type":              true,
		"list-style-position":          true,
		"list-style-image":             true,
		"list-style":                   true,
		"cursor":                       true,
		"white-space":                  true,
		"caption-side":                  true,
		"empty-cells":                  true,
		"caret-color":                  true,
		"appearance":                   true,
		"hyphens":                      true,
		"tab-size":                     true,
		"text-decoration-line":         true,
		"text-decoration-color":        true,
		"text-decoration-style":        true,
		"text-decoration-thickness":    true,
		"text-underline-offset":        true,
		"text-decoration-skip-ink":    true,
		"writing-mode":                 true,
		"text-orientation":             true,
		"user-select":                  true,
		"pointer-events":               true,
	}

	for prop, inheritable := range inheritableProps {
		if inheritable {
			if v, ok := parentStyle[prop]; ok {
				// Only inherit if current value is "inherit" or not set
				if currentVal, exists := props[prop]; !exists || currentVal == "inherit" {
					props[prop] = v
				}
			}
		}
	}
}

// matchSelector returns true if the element matches the CSS selector.
// Supports: tag, .class, #id, tag.class, tag#id, [attr], [attr=value], [attr~=value], [attr|=value]
// Also supports combinators: descendant (space), child (>), adjacent sibling (+), general sibling (~)
func matchSelector(tagName, class, id, selector string) bool {
	sel := selector

	// Parse selector parts
	var selTag, selClass, selID string
	var seenTag, seenClass, seenID bool

	for len(sel) > 0 {
		if sel[0] == '*' {
			sel = sel[1:]
			continue
		}

		// Attribute selector: [attr], [attr=value], [attr~=value], [attr|=value]
		// Note: attribute selector matching is deferred to matchAttributeSelector
		// This just skips over the attribute selector in the selector string
		if sel[0] == '[' {
			end := strings.Index(sel[1:], "]")
			if end > 0 {
				sel = sel[end+2:]
				continue
			}
		}

		if sel[0] == '.' {
			sel = sel[1:]
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' && sel[end] != '[' && sel[end] != ':' && sel[end] != ' ' {
				end++
			}
			selClass = sel[:end]
			sel = sel[end:]
			seenClass = true
		} else if sel[0] == '#' {
			sel = sel[1:]
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' && sel[end] != '[' && sel[end] != ':' && sel[end] != ' ' {
				end++
			}
			selID = sel[:end]
			sel = sel[end:]
			seenID = true
		} else if sel[0] == ' ' {
			// Descendant combinator — skip whitespace
			sel = sel[1:]
		} else if sel[0] == ':' {
			// Pseudo-class or pseudo-element (::before, ::after, ::marker) — skip
			sel = sel[1:]
			if len(sel) > 0 && sel[0] == ':' {
				sel = sel[1:]
			}
			// Skip function arguments if present
			if len(sel) > 0 && sel[0] == '(' {
				depth := 1
				sel = sel[1:]
				for len(sel) > 0 && depth > 0 {
					if sel[0] == '(' {
						depth++
					} else if sel[0] == ')' {
						depth--
					}
					sel = sel[1:]
				}
			}
		} else {
			// Tag name
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' && sel[end] != '[' && sel[end] != ':' && sel[end] != ' ' {
				end++
			}
			selTag = sel[:end]
			sel = sel[end:]
			seenTag = true
		}
	}

	if seenTag && !tagMatch(tagName, selTag) {
		return false
	}
	if seenClass && !hasClass(class, selClass) {
		return false
	}
	if seenID && id != selID {
		return false
	}

	// Attribute selectors always pass at this level since we don't have node attributes here
	// The actual attribute matching is done at a higher level via MatchNodeSelector
	return true
}

// matchAttributeSelector checks if an attribute value matches the selector.
func matchAttributeSelector(attrValue, op, selector string) bool {
	switch op {
	case "":
		// [attr] — attribute exists
		return attrValue != ""
	case "=":
		// [attr=value] — exact match
		return attrValue == selector
	case "~=":
		// [attr~=value] — space-separated list contains value
		for _, v := range strings.Fields(attrValue) {
			if v == selector {
				return true
			}
		}
		return false
	case "|=":
		// [attr|=value] — value or value followed by hyphen
		return attrValue == selector || strings.HasPrefix(attrValue, selector+"-")
	case "^=":
		// [attr^=value] — starts with value
		return strings.HasPrefix(attrValue, selector)
	case "$=":
		// [attr$=value] — ends with value
		return strings.HasSuffix(attrValue, selector)
	case "*=":
		// [attr*=value] — contains value
		return strings.Contains(attrValue, selector)
	}
	return false
}

// parseAttributeSelector parses an attribute selector string like [attr], [attr=value],
// [attr~=value], [attr|=value], [attr^=value], [attr$=value], [attr*=value]
// and returns (attrName, operator, value).
func parseAttributeSelector(selector string) (string, string, string) {
	// selector is like [attr], [attr=value], [attr="value"], etc.
	// Remove surrounding brackets
	s := selector
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")

	// Find the operator
	var attrName, op, value string

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '=' && op == "" {
			attrName = strings.TrimSpace(s[:i])
			op = "="
			value = strings.TrimSpace(s[i+1:])
			// Handle quoted values
			if len(value) >= 2 {
				if value[0] == '"' || value[0] == '\'' {
					value = value[1 : len(value)-1]
				}
			}
			break
		}
		if c == '~' && i+1 < len(s) && s[i+1] == '=' && op == "" {
			attrName = strings.TrimSpace(s[:i])
			op = "~="
			value = strings.TrimSpace(s[i+2:])
			break
		}
		if c == '|' && i+1 < len(s) && s[i+1] == '=' && op == "" {
			attrName = strings.TrimSpace(s[:i])
			op = "|="
			value = strings.TrimSpace(s[i+2:])
			break
		}
		if c == '^' && i+1 < len(s) && s[i+1] == '=' && op == "" {
			attrName = strings.TrimSpace(s[:i])
			op = "^="
			value = strings.TrimSpace(s[i+2:])
			break
		}
		if c == '$' && i+1 < len(s) && s[i+1] == '=' && op == "" {
			attrName = strings.TrimSpace(s[:i])
			op = "$="
			value = strings.TrimSpace(s[i+2:])
			break
		}
		if c == '*' && i+1 < len(s) && s[i+1] == '=' && op == "" {
			attrName = strings.TrimSpace(s[:i])
			op = "*="
			value = strings.TrimSpace(s[i+2:])
			break
		}
	}

	if op == "" {
		// No operator found — just [attr]
		attrName = strings.TrimSpace(s)
		op = ""
		value = ""
	}

	// Remove quotes from value if present
	if len(value) >= 2 {
		if (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
	}

	return attrName, op, value
}

// MatchNodeSelector returns true if the element node matches the full CSS selector.
// This handles tag names, IDs, classes, attribute selectors ([attr], [attr=value], etc.),
// pseudo-classes (:hover, :first-child, :not(), :nth-child()), pseudo-elements (::before, ::after, ::marker),
// and combinators (descendant space, child >, adjacent sibling +, general sibling ~).
func MatchNodeSelector(node *html.Node, selector string) bool {
	if node == nil || node.Type != html.NodeElement {
		return false
	}

	sel := strings.TrimSpace(selector)
	if sel == "" {
		return false
	}

	// Split by combinators while preserving them
	// Combinators: " " (descendant), ">" (child), "+" (adjacent sibling), "~" (general sibling)
	parts := splitSelectorParts(sel)
	if len(parts) == 0 {
		return false
	}

	// parts[0] is the first simple selector, parts[1+] are (combinator, simple selector) pairs
	if len(parts) == 1 {
		return matchSimpleSelector(node, parts[0])
	}

	// Multi-part selector: walk the DOM starting from this node
	return matchSelectorChain(node, parts)
}

// splitSelectorParts splits a selector into simple selector parts separated by combinators.
func splitSelectorParts(sel string) []string {
	var parts []string
	var current strings.Builder
	inAttr := false
	inParen := 0
	i := 0
	for i < len(sel) {
		c := sel[i]
		if c == '[' {
			inAttr = true
			current.WriteByte(c)
			i++
		} else if c == ']' {
			inAttr = false
			current.WriteByte(c)
			i++
		} else if c == '(' {
			inParen++
			current.WriteByte(c)
			i++
		} else if c == ')' {
			inParen--
			current.WriteByte(c)
			i++
		} else if (c == ' ' || c == '>' || c == '+' || c == '~') && !inAttr && inParen == 0 {
			// Found a combinator
			part := strings.TrimSpace(current.String())
			if part != "" {
				parts = append(parts, part)
			}
			// Skip whitespace before combinator
			for i < len(sel) && sel[i] == ' ' {
				i++
			}
			// Add the combinator as its own part
			if i < len(sel) && (sel[i] == '>' || sel[i] == '+' || sel[i] == '~') {
				parts = append(parts, string(sel[i]))
				i++
			}
			// Skip whitespace after combinator
			for i < len(sel) && sel[i] == ' ' {
				i++
			}
			current.Reset()
		} else {
			current.WriteByte(c)
			i++
		}
	}
	part := strings.TrimSpace(current.String())
	if part != "" {
		parts = append(parts, part)
	}
	return parts
}

// matchSelectorChain matches a chain of simple selectors separated by combinators.
func matchSelectorChain(node *html.Node, parts []string) bool {
	if len(parts) == 0 {
		return false
	}

	// Start with the first simple selector
	curr := node
	if !matchSimpleSelector(curr, parts[0]) {
		return false
	}

	// Process remaining (combinator, selector) pairs
	i := 1
	for i < len(parts) {
		if i+1 >= len(parts) {
			break
		}
		combinator := parts[i]
		selector := parts[i+1]
		i += 2

		var next *html.Node
		switch combinator {
		case ">":
			// Child: direct parent
			next = findDirectParent(curr)
		case "+":
			// Adjacent sibling: immediately preceding sibling
			next = findPrecedingSibling(curr)
		case "~":
			// General sibling: any preceding sibling
			next = findAnyPrecedingSibling(curr)
		default:
			// Descendant (space): any ancestor
			next = findAncestor(curr)
		}

		if next == nil {
			return false
		}
		curr = next

		// If combinator was descendant, we need to walk up and find a matching ancestor
		if combinator == " " {
			found := false
			for curr != nil {
				if matchSimpleSelector(curr, selector) {
					found = true
					break
				}
				curr = curr.Parent
			}
			if !found {
				return false
			}
		} else {
			if !matchSimpleSelector(curr, selector) {
				return false
			}
		}
	}
	return true
}

func findDirectParent(node *html.Node) *html.Node {
	return node.Parent
}

func findPrecedingSibling(node *html.Node) *html.Node {
	if node.Parent == nil {
		return nil
	}
	for _, sib := range node.Parent.Children {
		if sib == node {
			return nil // No preceding sibling found (we're the first)
		}
		if sib.Type == html.NodeElement {
			// Return the element immediately before us
			return sib
		}
	}
	return nil
}

func findAnyPrecedingSibling(node *html.Node) *html.Node {
	if node.Parent == nil {
		return nil
	}
	var prev *html.Node
	for _, sib := range node.Parent.Children {
		if sib == node {
			return prev
		}
		if sib.Type == html.NodeElement {
			prev = sib
		}
	}
	return nil
}

func findAncestor(node *html.Node) *html.Node {
	return node.Parent
}

// matchSimpleSelector matches a simple selector (no combinators) against a node.
// Simple selector: tag#id.class[attr]:pseudo
func matchSimpleSelector(node *html.Node, selector string) bool {
	sel := selector

	// Universal selector * matches everything
	if sel == "*" {
		return true
	}

	// Track what we've matched
	tagName := strings.ToLower(node.TagName)

	// Parse the selector from left to right
	for len(sel) > 0 {
		sel = strings.TrimSpace(sel)
		if len(sel) == 0 {
			break
		}

		switch sel[0] {
		case '#':
			// ID selector
			sel = sel[1:]
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' && sel[end] != '[' && sel[end] != ':' {
				end++
			}
			id := sel[:end]
			sel = sel[end:]
			if node.GetAttribute("id") != id {
				return false
			}

		case '.':
			// Class selector
			sel = sel[1:]
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' && sel[end] != '[' && sel[end] != ':' {
				end++
			}
			class := sel[:end]
			sel = sel[end:]
			if !node.ClassList().Contains(class) {
				return false
			}

		case '[':
			// Attribute selector
			end := strings.Index(sel[1:], "]")
			if end < 0 {
				return false
			}
			attrSel := sel[:end+2]
			sel = sel[end+2:]

			attrName, op, value := parseAttributeSelector(attrSel)
			nodeAttr := node.GetAttribute(attrName)
			if !matchAttributeSelector(nodeAttr, op, value) {
				return false
			}

		case ':':
			// Pseudo-class or pseudo-element
			sel = sel[1:]
			if len(sel) == 0 {
				return false
			}
			if sel[0] == ':' {
				// Pseudo-element (::before, ::after, ::marker) — for now, treated as matching
				sel = sel[1:]
				continue
			}
			// Parse pseudo-class name and optional argument
			var pseudoName string
			var pseudoArg string
			if idx := strings.Index(sel, "("); idx >= 0 {
				pseudoName = sel[:idx]
				argStart := idx + 1
				depth := 1
				for i := argStart; i < len(sel); i++ {
					if sel[i] == '(' {
						depth++
					} else if sel[i] == ')' {
						depth--
						if depth == 0 {
							pseudoArg = sel[argStart:i]
							sel = sel[i+1:]
							break
						}
					}
				}
				if depth != 0 {
					return false
				}
			} else {
				// No argument — find end of pseudo name
				end := 0
				for end < len(sel) && (isAlphanumeric(sel[end]) || sel[end] == '-') {
					end++
				}
				pseudoName = sel[:end]
				sel = sel[end:]
			}
			pseudoName = strings.ToLower(pseudoName)

			// Evaluate the pseudo-class
			if !matchPseudoClass(node, pseudoName, pseudoArg) {
				return false
			}

		default:
			// Tag selector
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' && sel[end] != '[' && sel[end] != ':' && sel[end] != ' ' {
				end++
			}
			tag := sel[:end]
			sel = sel[end:]
			if tag != "*" && tagName != strings.ToLower(tag) {
				return false
			}
		}
	}

	return true
}

// isAlphanumeric returns true if c is an ASCII letter or digit.
func isAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// matchPseudoClass evaluates a pseudo-class against a node.
// pseudoName is the lowercased pseudo-class name (e.g., "first-child", "not", "nth-child").
// pseudoArg is the optional argument (e.g., "2n+1" for :nth-child).
func matchPseudoClass(node *html.Node, pseudoName, pseudoArg string) bool {
	switch pseudoName {
	case "first-child":
		if node.Parent == nil {
			return false
		}
		for _, child := range node.Parent.Children {
			if child.Type == html.NodeElement {
				return child == node
			}
		}
		return false
	case "last-child":
		if node.Parent == nil {
			return false
		}
		var lastElement *html.Node
		for _, child := range node.Parent.Children {
			if child.Type == html.NodeElement {
				lastElement = child
			}
		}
		return lastElement == node
	case "only-child":
		if node.Parent == nil {
			return false
		}
		count := 0
		for _, child := range node.Parent.Children {
			if child.Type == html.NodeElement {
				count++
			}
		}
		return count == 1
	case "nth-child":
		return MatchNthChild(node, pseudoArg, false, false)
	case "nth-of-type":
		return MatchNthChild(node, pseudoArg, true, false)
	case "nth-last-child":
		return MatchNthChild(node, pseudoArg, false, true)
	case "first-of-type":
		if node.Parent == nil {
			return false
		}
		tag := node.TagName
		for _, child := range node.Parent.Children {
			if child.Type == html.NodeElement && child.TagName == tag {
				return child == node
			}
		}
		return false
	case "last-of-type":
		if node.Parent == nil {
			return false
		}
		tag := node.TagName
		var last *html.Node
		for _, child := range node.Parent.Children {
			if child.Type == html.NodeElement && child.TagName == tag {
				last = child
			}
		}
		return last == node
	case "empty":
		for _, child := range node.Children {
			if child.Type == html.NodeElement {
				return false
			}
			if child.Type == html.NodeText && strings.TrimSpace(child.Data) != "" {
				return false
			}
		}
		return true
	case "not":
		// :not(selector) — matches if the argument selector does NOT match
		return MatchNot(node, pseudoArg)
	case "is":
		// :is(selector list) — matches if ANY selector in the list matches (forgiving)
		return MatchIsWherePseudoClass(node, pseudoArg)
	case "where":
		// :where(selector list) — same as :is() but has 0 specificity
		// Matching logic is identical to :is()
		return MatchIsWherePseudoClass(node, pseudoArg)
	case "hover", "focus", "active", "visited", "link":
		// State pseudo-classes — for now, treat as matching (no interactivity state tracking)
		return true
	case "checked":
		// For input[type=checkbox] and input[type=radio], matches if checked attribute is present
		typ := strings.ToLower(node.GetAttribute("type"))
		if typ == "checkbox" || typ == "radio" {
			return node.HasAttribute("checked")
		}
		return false
	case "disabled":
		// :disabled matches elements with the disabled attribute
		// Form elements: input, button, select, textarea, fieldset, optgroup, option
		tag := strings.ToLower(node.TagName)
		switch tag {
		case "input", "button", "select", "textarea", "fieldset", "optgroup", "option":
			return node.HasAttribute("disabled")
		}
		return false
	case "enabled":
		// :enabled matches elements that are not disabled
		tag := strings.ToLower(node.TagName)
		switch tag {
		case "input", "button", "select", "textarea", "fieldset", "optgroup", "option":
			return !node.HasAttribute("disabled")
		}
		// Non-form elements are implicitly enabled
		return true
	case "focus-visible":
		// :focus-visible matches elements with focus that should show a focus indicator.
		// Since we don't have real input system tracking, we match elements that are
		// keyboard-focusable: those with tabindex OR focusable form elements.
		if node.HasAttribute("tabindex") {
			return true
		}
		tag := strings.ToLower(node.TagName)
		switch tag {
		case "input", "button", "select", "textarea", "a", "area":
			return true
		}
		return false
	case "lang":
		// :lang(en) matches elements with lang attribute starting with the given language code
		if pseudoArg == "" {
			return false
		}
		lang := node.GetAttribute("lang")
		if lang == "" {
			return false
		}
		lang = strings.ToLower(lang)
		code := strings.ToLower(pseudoArg)
		// Match if lang equals code OR starts with code followed by '-'
		return lang == code || strings.HasPrefix(lang, code+"-")
	case "valid":
		// Form validation — checks proper formatting for required fields
		return IsValid(node)
	case "invalid":
		// Form validation — inverse of valid
		return IsInvalid(node)
	case "placeholder-shown":
		return PlaceholderShown(node)
	case "has":
		// :has(relative selector) — matches if the element has descendants/siblings
		// matching the relative selector
		return MatchHasSelector(node, pseudoArg)
	case "indeterminate":
		// :indeterminate matches checkbox elements in an indeterminate state
		// (JavaScript sets .indeterminate=true) or radio buttons in a group where
		// none are checked. Since we don't have JS, match <input type="checkbox">
		// (without checked attribute) and <input type="radio">.
		typ := strings.ToLower(node.GetAttribute("type"))
		if typ == "checkbox" {
			return !node.HasAttribute("checked")
		}
		if typ == "radio" {
			return true
		}
		return false
	case "default":
		// :default matches form elements that are the default in a group:
		// submit buttons, the initially-checked radio/checkbox, default option in select.
		// Match: <button type="submit">, <input type="submit">, <input type="image">,
		// and <input type="checkbox"> or <input type="radio"> with checked attribute.
		tag := strings.ToLower(node.TagName)
		typ := strings.ToLower(node.GetAttribute("type"))
		if tag == "button" && typ == "submit" {
			return true
		}
		if typ == "submit" || typ == "image" {
			return true
		}
		if typ == "checkbox" || typ == "radio" {
			return node.HasAttribute("checked")
		}
		return false
	default:
		// Unknown pseudo-class — treat as matching for forward compatibility
		return true
	}
}

func tagMatch(elTag, selTag string) bool {
	if selTag == "" || selTag == "*" {
		return true
	}
	return elTag == selTag
}

func hasClass(elClass, selClass string) bool {
	if elClass == "" || selClass == "" {
		return selClass == ""
	}
	classes := splitClasses(elClass)
	for _, c := range classes {
		if c == selClass {
			return true
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

func computeSpecificity(selector string) Specificity {
	var a, b, c int
	sel := selector
	for len(sel) > 0 {
		switch sel[0] {
		case '#':
			a++
			sel = sel[1:]
			for len(sel) > 0 && sel[0] != '.' && sel[0] != '#' && sel[0] != ' ' && sel[0] != '[' && sel[0] != ':' {
				sel = sel[1:]
			}
		case '.':
			b++
			sel = sel[1:]
			for len(sel) > 0 && sel[0] != '.' && sel[0] != '#' && sel[0] != ' ' && sel[0] != '[' && sel[0] != ':' {
				sel = sel[1:]
			}
		case '[':
			// Attribute selector
			b++
			end := strings.Index(sel[1:], "]")
			if end > 0 {
				sel = sel[end+2:]
				for len(sel) > 0 && sel[0] == ' ' {
					sel = sel[1:]
				}
			} else {
				break
			}
		case ':':
			// Pseudo-class or pseudo-element
			b++
			sel = sel[1:]
			if sel[0] == ':' {
				// Pseudo-element (::before, ::marker) — CSS3 specificity treats these like class (b)
				sel = sel[1:]
			}
			for len(sel) > 0 && sel[0] != ' ' && sel[0] != '[' && sel[0] != '.' && sel[0] != '#' && sel[0] != ':' && sel[0] != '(' {
				sel = sel[1:]
			}
			if len(sel) > 0 && sel[0] == '(' {
				// Skip function arguments
				depth := 1
				sel = sel[1:]
				for len(sel) > 0 && depth > 0 {
					if sel[0] == '(' {
						depth++
					} else if sel[0] == ')' {
						depth--
					}
					sel = sel[1:]
				}
			}
		case ' ':
			sel = sel[1:]
		case '>':
			sel = sel[1:]
		case '+':
			sel = sel[1:]
		case '~':
			sel = sel[1:]
		default:
			c++
			for len(sel) > 0 && sel[0] != '.' && sel[0] != '#' && sel[0] != ' ' && sel[0] != '[' && sel[0] != ':' && sel[0] != '>' && sel[0] != '+' && sel[0] != '~' {
				sel = sel[1:]
			}
		}
	}
	return Specificity{a, b, c}
}

func applyDecl(props map[string]string, decl Declaration) {
	prop := strings.ToLower(decl.Property)
	value := decl.Value

	switch prop {
	case "display":
		props["display"] = value
	case "color":
		props["color"] = value
	case "background":
		// background shorthand: [color] [image] [repeat] [position] [/ size]
		// Parse space-separated values into component properties
		parts := strings.Fields(value)
		colorSet := false
		for _, part := range parts {
			lower := strings.ToLower(part)
			if strings.HasPrefix(lower, "url(") {
				props["background-image"] = part
			} else if part == "no-repeat" || part == "repeat" || part == "repeat-x" || part == "repeat-y" || part == "space" || part == "round" {
				props["background-repeat"] = part
			} else if part == "left" || part == "right" || part == "top" || part == "bottom" || part == "center" {
				props["background-position"] = part
			} else if strings.Contains(part, "/") {
				posParts := strings.Split(part, "/")
				if len(posParts) == 2 {
					props["background-position"] = strings.TrimSpace(posParts[0])
					props["background-size"] = strings.TrimSpace(posParts[1])
				}
			} else if lower == "transparent" {
				props["background-color"] = "transparent"
				colorSet = true
			} else if lower == "inherit" || lower == "initial" || lower == "unset" {
				props["background-color"] = part
				colorSet = true
			} else if strings.HasPrefix(lower, "linear-gradient") || strings.HasPrefix(lower, "radial-gradient") || strings.HasPrefix(lower, "conic-gradient") {
				props["background-image"] = part
			} else if col := ParseColor(part); col.A > 0 {
				props["background-color"] = part
				colorSet = true
			}
		}
		// Store the full shorthand value for canvas.go to parse
		props["background"] = value
		// If no color was identified, treat last color-like value as background-color
		if !colorSet && len(parts) > 0 {
			if ParseColor(parts[len(parts)-1]).A > 0 {
				props["background-color"] = parts[len(parts)-1]
			}
		}
	case "background-color":
		props["background-color"] = value
		props["background"] = value // sync for canvas.go
	case "background-repeat":
		props["background-repeat"] = value
	case "background-position":
		props["background-position"] = value
	case "background-size":
		props["background-size"] = value
	case "list-style-type":
		props["list-style-type"] = value
	case "list-style-position":
		props["list-style-position"] = value
	case "list-style-image":
		props["list-style-image"] = value
	case "list-style":
		// shorthand: type position image
		parts := strings.Fields(value)
		for _, part := range parts {
			if part == "inside" || part == "outside" {
				props["list-style-position"] = part
			} else if strings.HasPrefix(part, "url(") {
				props["list-style-image"] = part
			} else if part != "" {
				props["list-style-type"] = part
			}
		}
	case "font-size":
		props["font-size"] = value
	case "font-family":
		props["font-family"] = value
	case "font-weight":
		props["font-weight"] = value
	case "font-stretch":
		props["font-stretch"] = value
	case "font":
		// Font shorthand: [style] [variant] [weight] [stretch] size[/line-height] family
		// Examples: "12px Arial", "bold 16px sans-serif", "italic 12px/1.5 serif", "semi-condensed 16px Arial"
		parts := strings.Fields(value)
		if len(parts) >= 2 {
			// Try to identify size and family
			for i, part := range parts {
				// Check if this part contains font-size (with optional line-height)
				if strings.Contains(part, "px") || strings.Contains(part, "em") || strings.Contains(part, "pt") || strings.Contains(part, "%") {
					// This is the size part
					if idx := strings.Index(part, "/"); idx >= 0 {
						props["font-size"] = part[:idx]
						props["line-height"] = part[idx+1:]
					} else {
						props["font-size"] = part
					}
					// Everything after this is font-family
					if i+1 < len(parts) {
						props["font-family"] = strings.Join(parts[i+1:], ", ")
					}
					break
				}
				// Check for style
				if part == "italic" || part == "oblique" || part == "normal" {
					props["font-style"] = part
				}
				// Check for font-stretch keywords
				if part == "ultra-condensed" || part == "extra-condensed" || part == "condensed" ||
					part == "semi-condensed" || part == "semi-expanded" || part == "expanded" ||
					part == "extra-expanded" || part == "ultra-expanded" {
					props["font-stretch"] = part
				}
				// Check for weight
				if part == "bold" || part == "bolder" || part == "lighter" || part == "normal" {
					props["font-weight"] = part
				} else if _, err := strconv.Atoi(part); err == nil {
					props["font-weight"] = part
				}
			}
		}
	case "margin-top":
		props["margin-top"] = value
	case "margin-right":
		props["margin-right"] = value
	case "margin-bottom":
		props["margin-bottom"] = value
	case "margin-left":
		props["margin-left"] = value
	case "margin":
		props["margin-top"] = value
		props["margin-right"] = value
		props["margin-bottom"] = value
		props["margin-left"] = value
	case "padding-top":
		props["padding-top"] = value
	case "padding-right":
		props["padding-right"] = value
	case "padding-bottom":
		props["padding-bottom"] = value
	case "padding-left":
		props["padding-left"] = value
	case "padding":
		props["padding-top"] = value
		props["padding-right"] = value
		props["padding-bottom"] = value
		props["padding-left"] = value
	case "border-width":
		props["border-width"] = value
	case "border-color":
		props["border-color"] = value
	case "border-style":
		props["border-style"] = value
	case "width":
		props["width"] = value
	case "height":
		props["height"] = value
	case "visibility":
		props["visibility"] = value
	case "white-space":
		props["white-space"] = value
	case "letter-spacing":
		props["letter-spacing"] = value
	case "word-spacing":
		props["word-spacing"] = value
	case "text-indent":
		props["text-indent"] = value
	case "text-transform":
		props["text-transform"] = value
	case "text-justify":
		props["text-justify"] = value
	case "ruby-align":
		props["ruby-align"] = value
	case "ruby-position":
		props["ruby-position"] = value
	case "place-items":
		// Shorthand: align-items justify-items
		// If only one value is provided, it applies to both
		parts := strings.Fields(value)
		if len(parts) >= 1 {
			props["align-items"] = parts[0]
		}
		if len(parts) >= 2 {
			props["justify-items"] = parts[1]
		} else {
			props["justify-items"] = parts[0]
		}
		props["place-items"] = value
	case "place-self":
		// Shorthand: align-self justify-self
		// If only one value is provided, it applies to both
		parts := strings.Fields(value)
		if len(parts) >= 1 {
			props["align-self"] = parts[0]
		}
		if len(parts) >= 2 {
			props["justify-self"] = parts[1]
		} else {
			props["justify-self"] = parts[0]
		}
		props["place-self"] = value
	case "justify-items":
		props["justify-items"] = value
	case "justify-self":
		props["justify-self"] = value
	case "place-content":
		props["place-content"] = value
	case "text-wrap":
		props["text-wrap"] = value
	case "math-style":
		props["math-style"] = value
	case "view-transition-name":
		props["view-transition-name"] = value
	case "field-sizing":
		props["field-sizing"] = value
	case "container-type":
		props["container-type"] = value
	case "container-name":
		props["container-name"] = value
	case "container":
		props["container"] = value
	case "font-variant":
		props["font-variant"] = value
	case "unicode-bidi":
		props["unicode-bidi"] = value
	case "direction":
		props["direction"] = value
	case "writing-mode":
		props["writing-mode"] = value
	case "tab-size":
		props["tab-size"] = value
	case "quotes":
		props["quotes"] = value
	case "opacity":
		props["opacity"] = value
	case "vertical-align":
		props["vertical-align"] = value
	case "overflow":
		props["overflow"] = value
	case "overflow-x":
		props["overflow-x"] = value
	case "overflow-y":
		props["overflow-y"] = value
	case "word-wrap":
		props["word-wrap"] = value
	case "overflow-wrap":
		props["overflow-wrap"] = value
	case "cursor":
		props["cursor"] = value
	case "transform":
		props["transform"] = value
	case "transform-box":
		props["transform-box"] = value
	case "position":
		props["position"] = value
	case "top":
		props["top"] = value
	case "right":
		props["right"] = value
	case "bottom":
		props["bottom"] = value
	case "left":
		props["left"] = value
	case "float":
		props["float"] = value
	case "z-index":
		props["z-index"] = value
	case "flex-direction":
		props["flex-direction"] = value
	case "justify-content":
		props["justify-content"] = value
	case "align-items":
		props["align-items"] = value
	case "align-self":
		props["align-self"] = value
	case "flex-grow":
		props["flex-grow"] = value
	case "flex-shrink":
		props["flex-shrink"] = value
	case "flex-basis":
		props["flex-basis"] = value
	case "flex-wrap":
		props["flex-wrap"] = value
	case "gap":
		props["gap"] = value
	case "flex-flow":
		props["flex-flow"] = value
		// Parse into direction and wrap
		parts := strings.Fields(value)
		for _, part := range parts {
			lower := strings.ToLower(part)
			if lower == "row" || lower == "column" || lower == "row-reverse" || lower == "column-reverse" {
				props["flex-direction"] = part
			} else if lower == "nowrap" || lower == "wrap" || lower == "wrap-reverse" {
				props["flex-wrap"] = part
			}
		}
	case "align-content":
		props["align-content"] = value
	case "order":
		props["order"] = value
	case "border-radius":
		props["border-radius"] = value
	case "outline-width":
		props["outline-width"] = value
	case "outline-style":
		props["outline-style"] = value
	case "outline-color":
		props["outline-color"] = value
	case "outline":
		// outline is a shorthand: outline-width outline-style outline-color
		// Parse space-separated values
		parts := strings.Fields(value)
		if len(parts) >= 1 {
			props["outline-style"] = parts[0] // style is always first
		}
		if len(parts) >= 2 {
			props["outline-width"] = parts[1]
		}
		if len(parts) >= 3 {
			props["outline-color"] = parts[2]
		}
	case "box-shadow":
		props["box-shadow"] = value
	case "text-shadow":
		props["text-shadow"] = value
	case "background-image":
		props["background-image"] = value
	case "text-overflow":
		props["text-overflow"] = value
	case "content":
		props["content"] = value
	// Animation shorthand
	case "animation":
		parseAnimationShorthand(props, value)
	// Individual animation properties
	case "animation-name":
		props["animation-name"] = value
	case "animation-duration":
		props["animation-duration"] = value
	case "animation-timing-function":
		props["animation-timing-function"] = value
	case "animation-delay":
		props["animation-delay"] = value
	case "animation-iteration-count":
		props["animation-iteration-count"] = value
	case "animation-direction":
		props["animation-direction"] = value
	case "animation-fill-mode":
		props["animation-fill-mode"] = value
	case "aspect-ratio":
		props["aspect-ratio"] = value
	case "object-fit":
		// Valid values: fill, contain, cover, none, scale-down
		props["object-fit"] = value
	case "object-position":
		props["object-position"] = value
	case "filter":
		props["filter"] = value
	case "backdrop-filter":
		props["backdrop-filter"] = value
	case "clip-path":
		props["clip-path"] = value
	case "clip":
		props["clip"] = value
	case "column-width":
		props["column-width"] = value
	case "column-count":
		props["column-count"] = value
	case "column-gap":
		props["column-gap"] = value
	case "column-rule-width":
		props["column-rule-width"] = value
	case "column-rule-style":
		props["column-rule-style"] = value
	case "column-rule-color":
		props["column-rule-color"] = value
	case "column-rule":
		// shorthand: width style color
		parts := strings.Fields(value)
		for _, part := range parts {
			lower := strings.ToLower(part)
			if lower == "thin" || lower == "medium" || lower == "thick" {
				props["column-rule-width"] = part
			} else if part == "none" || part == "hidden" || part == "dotted" || part == "dashed" || part == "solid" || part == "double" || part == "groove" || part == "ridge" || part == "inset" || part == "outset" {
				props["column-rule-style"] = part
			} else if ParseColor(part).A > 0 || strings.HasPrefix(lower, "#") || strings.HasPrefix(lower, "rgb") {
				props["column-rule-color"] = part
			}
		}
	case "break-inside":
		props["break-inside"] = value
	case "break-before":
		props["break-before"] = value
	case "break-after":
		props["break-after"] = value
	case "pointer-events":
		// Valid values: auto, none, visiblePainted, visibleFill, visibleStroke, painted, fill, stroke, all
		props["pointer-events"] = value
	case "border-image":
		// border-image: source slice / width / outset repeat
		// Example: url(border.png) 30 round
		props["border-image"] = value
		parts := strings.Fields(value)
		for _, part := range parts {
			lower := strings.ToLower(part)
			if strings.HasPrefix(lower, "url(") || lower == "none" || lower == "linear-gradient" || strings.HasPrefix(lower, "radial-gradient") {
				props["border-image-source"] = part
			} else if strings.Contains(part, "/") {
				// slice/width/outset
				slices := strings.Split(part, "/")
				if len(slices) >= 1 {
					props["border-image-slice"] = slices[0]
				}
				if len(slices) >= 2 {
					props["border-image-width"] = slices[1]
				}
				if len(slices) >= 3 {
					props["border-image-outset"] = slices[2]
				}
			} else if lower == "stretch" || lower == "repeat" || lower == "round" || lower == "space" {
				props["border-image-repeat"] = lower
			}
		}
	case "border-image-source":
		props["border-image-source"] = value
	case "border-image-slice":
		props["border-image-slice"] = value
	case "border-image-width":
		props["border-image-width"] = value
	case "border-image-outset":
		props["border-image-outset"] = value
	case "border-image-repeat":
		props["border-image-repeat"] = value
	case "transition":
		parseTransitionShorthand(props, value)
	case "transition-property":
		props["transition-property"] = value
	case "transition-duration":
		props["transition-duration"] = value
	case "transition-timing-function":
		props["transition-timing-function"] = value
	case "transition-delay":
		props["transition-delay"] = value
	case "resize":
		props["resize"] = value
	case "caption-side":
		props["caption-side"] = value
	case "appearance":
		props["appearance"] = value
	case "overscroll-behavior":
		props["overscroll-behavior"] = value
	case "overscroll-behavior-x":
		props["overscroll-behavior-x"] = value
	case "overscroll-behavior-y":
		props["overscroll-behavior-y"] = value
	case "scroll-behavior":
		props["scroll-behavior"] = value
	case "text-decoration-line":
		props["text-decoration-line"] = value
	case "text-decoration-color":
		props["text-decoration-color"] = value
	case "text-decoration-style":
		props["text-decoration-style"] = value
	case "text-decoration":
		// Shorthand: line style color (e.g., "underline solid red")
		// Parse individual values
		parts := strings.Fields(value)
		for _, part := range parts {
			lower := strings.ToLower(part)
			if lower == "underline" || lower == "overline" || lower == "line-through" || lower == "blink" || lower == "none" {
				props["text-decoration-line"] = lower
			} else if lower == "solid" || lower == "double" || lower == "dotted" || lower == "dashed" || lower == "wavy" {
				props["text-decoration-style"] = lower
			} else {
				// Assume it's a color
				props["text-decoration-color"] = part
			}
		}
		// Default style to solid if only line is specified
		if props["text-decoration-style"] == "" {
			props["text-decoration-style"] = "solid"
		}
	case "will-change":
		props["will-change"] = value
	case "image-rendering":
		props["image-rendering"] = value
	case "contain":
		props["contain"] = value
	case "mix-blend-mode":
		props["mix-blend-mode"] = value
	case "hanging-punctuation":
		props["hanging-punctuation"] = value
	case "text-decoration-thickness":
		props["text-decoration-thickness"] = value
	case "text-underline-offset":
		props["text-underline-offset"] = value
	case "outline-offset":
		props["outline-offset"] = value
	case "text-decoration-skip-ink":
		props["text-decoration-skip-ink"] = value
	case "hyphens":
		props["hyphens"] = value
	}
}

// parseAnimationShorthand parses the CSS animation shorthand property.
// Syntax: name duration timing-function delay iteration-count direction fill-mode
// All values except name are optional, and order matters.
func parseAnimationShorthand(props map[string]string, value string) {
	// Common timing function keywords
	timingFuncs := map[string]bool{
		"ease": true, "linear": true, "ease-in": true, "ease-out": true,
		"ease-in-out": true, "step-start": true, "step-end": true,
	}
	// Common direction keywords
	directions := map[string]bool{
		"normal": true, "reverse": true, "alternate": true, "alternate-reverse": true,
	}
	// Common fill-mode keywords
	fillModes := map[string]bool{
		"none": true, "forwards": true, "backwards": true, "both": true,
	}
	// Common iteration-count values
	iterCounts := map[string]bool{
		"infinite": true,
	}

	parts := strings.Fields(value)
	if len(parts) == 0 {
		return
	}

	// Track which properties have been set
	setDuration := false
	setTimingFunc := false
	setDelay := false
	setIterCount := false
	setDirection := false
	setFillMode := false

	// Iterate through parts - first is always name if it doesn't look like a keyword
	// Since order is flexible after name, we try to identify each value by context
	i := 0
	// First, try to find the animation name (first unclassified value)
	if len(parts) > 0 {
		// Check if first part could be a duration (contains 's' or numeric)
		first := parts[0]
		lower := strings.ToLower(first)
		isDuration := strings.HasSuffix(first, "s") && !strings.HasSuffix(first, "ms")
		if isDuration || (strings.ContainsAny(first, "0123456789.") && !strings.HasSuffix(first, "%")) {
			// First part is a duration, not a name — name stays "none"
		} else if !timingFuncs[lower] && !directions[lower] && !fillModes[lower] && !iterCounts[lower] {
			// First part is the animation name
			props["animation-name"] = first
			i++
		} else {
			// First part is not a name, name stays "none"
		}
	}

	for ; i < len(parts); i++ {
		part := parts[i]
		lower := strings.ToLower(part)
		isTimeValue := strings.HasSuffix(part, "s") || strings.HasSuffix(part, "ms")

		// Track how many time values we've seen to determine meaning:
		// 1st duration-like → duration, 2nd → timing-function (rare), 3rd → delay
		timeValueCount := 0
		if setDuration {
			timeValueCount++
		}
		if setTimingFunc {
			timeValueCount++
		}
		if setDelay {
			timeValueCount++
		}

		if !setDuration && isTimeValue && timeValueCount == 0 {
			// First time value is always duration
			props["animation-duration"] = part
			setDuration = true
		} else if !setTimingFunc && isTimeValue && timeValueCount == 1 {
			// Second time value is timing-function (unusual but possible)
			props["animation-timing-function"] = part
			setTimingFunc = true
		} else if !setDelay && isTimeValue && timeValueCount >= 2 {
			// Third+ time value is delay
			props["animation-delay"] = part
			setDelay = true
		} else if !setTimingFunc && timingFuncs[lower] {
			props["animation-timing-function"] = part
			setTimingFunc = true
		} else if !setIterCount && iterCounts[lower] {
			props["animation-iteration-count"] = part
			setIterCount = true
		} else if !setIterCount {
			// Could be a number (iteration count)
			if _, err := strconv.ParseFloat(part, 64); err == nil {
				props["animation-iteration-count"] = part
				setIterCount = true
			}
		} else if !setDirection && directions[lower] {
			props["animation-direction"] = part
			setDirection = true
		} else if !setFillMode && fillModes[lower] {
			props["animation-fill-mode"] = part
			setFillMode = true
		}
		// If we still haven't set duration and this part looks like it could be timing function
		if !setTimingFunc && !setDuration && (timingFuncs[lower] || lower == "cubic-bezier" || lower == "steps") {
			// Could be timing function or name
			if !strings.ContainsAny(part, "0123456789") {
				props["animation-timing-function"] = part
				setTimingFunc = true
			}
		}
	}
}

// parseTransitionShorthand parses the CSS transition shorthand property.
// Syntax: property duration timing-function delay
func parseTransitionShorthand(props map[string]string, value string) {
	// timing functions
	timingFuncs := map[string]bool{
		"ease": true, "linear": true, "ease-in": true, "ease-out": true,
		"ease-in-out": true, "step-start": true, "step-end": true,
	}

	parts := strings.Fields(value)
	if len(parts) == 0 {
		return
	}

	setDuration := false
	setDelay := false
	setTimingFunc := false

	// Iterate through parts
	for i, part := range parts {
		lower := strings.ToLower(part)

		if !setDuration && (strings.HasSuffix(part, "s") || strings.HasSuffix(part, "ms")) {
			if !setDelay {
				props["transition-duration"] = part
				setDuration = true
			} else {
				props["transition-delay"] = part
			}
		} else if !setTimingFunc && (timingFuncs[lower] || strings.HasPrefix(lower, "cubic-bezier") || strings.HasPrefix(lower, "steps")) {
			props["transition-timing-function"] = part
			setTimingFunc = true
		} else if !strings.ContainsAny(part, "0123456789") && part != "none" && part != "all" {
			// This is the property name
			props["transition-property"] = part
		}

		// Check if this could be delay (comes after duration in the shorthand)
		if setDuration && !setDelay && i > 0 {
			// Look ahead for delay
		}
	}

	// If only one time value and no timing function, it could be duration or delay
	// Default timing function is ease
	if !setTimingFunc {
		props["transition-timing-function"] = "ease"
	}
}
