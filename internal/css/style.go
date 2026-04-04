package css

import (
	"strconv"
	"strings"
)

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
		"font-variant":    "normal",
		"unicode-bidi":   "normal",
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

// matchSelector returns true if the element matches the CSS selector.
// Supports: tag, .class, #id, tag.class, tag#id, [attr], [attr=value], [attr~=value], [attr|=value]
// Also supports combinators: descendant (space), child (>), adjacent sibling (+), general sibling (~)
func matchSelector(tagName, class, id, selector string) bool {
	sel := selector

	// Parse selector parts
	var selTag, selClass, selID string
	var seenTag, seenClass, seenID bool
	var attrName, attrValue string
	var attrOp string // "", "=", "~=", "|="

	for len(sel) > 0 {
		if sel[0] == '*' {
			sel = sel[1:]
			continue
		}

		// Attribute selector: [attr], [attr=value], [attr~=value], [attr|=value]
		if sel[0] == '[' {
			end := strings.Index(sel[1:], "]")
			if end > 0 {
				attrPart := sel[1 : end+1]
				sel = sel[end+2:]

				// Parse attribute selector
				eqIdx := strings.Index(attrPart, "=")
				if eqIdx > 0 {
					attrName = attrPart[:eqIdx]
					attrValue = attrPart[eqIdx+1:]
					attrValue = strings.Trim(attrValue, "\"")
					if strings.HasPrefix(attrPart, attrName+"~=") {
						attrOp = "~="
					} else if strings.HasPrefix(attrPart, attrName+"|=") {
						attrOp = "|="
					} else {
						attrOp = "="
					}
				} else {
					attrName = attrPart
					attrOp = ""
					attrValue = ""
				}
				continue
			}
		}

		if sel[0] == '.' {
			sel = sel[1:]
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' && sel[end] != '[' {
				end++
			}
			selClass = sel[:end]
			sel = sel[end:]
			seenClass = true
		} else if sel[0] == '#' {
			sel = sel[1:]
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' && sel[end] != '[' {
				end++
			}
			selID = sel[:end]
			sel = sel[end:]
			seenID = true
		} else {
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' && sel[end] != '[' {
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
	// The actual attribute matching is done at a higher level via matchAttributeSelector
	return true
}

// matchAttributeSelector checks if an attribute value matches the selector.
func matchAttributeSelector(attrValue, op, selector string) bool {
	switch op {
	case "":
		return attrValue != ""
	case "=":
		return attrValue == selector
	case "~=":
		// Space-separated list contains value
		for _, v := range strings.Fields(attrValue) {
			if v == selector {
				return true
			}
		}
		return false
	case "|=":
		// Value or value followed by hyphen
		return attrValue == selector || strings.HasPrefix(attrValue, selector+"-")
	}
	return false
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
			for len(sel) > 0 && sel[0] != '.' && sel[0] != '#' && sel[0] != ' ' {
				sel = sel[1:]
			}
		case '.':
			b++
			sel = sel[1:]
			for len(sel) > 0 && sel[0] != '.' && sel[0] != '#' && sel[0] != ' ' {
				sel = sel[1:]
			}
		case ' ':
			sel = sel[1:]
		default:
			c++
			for len(sel) > 0 && sel[0] != '.' && sel[0] != '#' && sel[0] != ' ' {
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
		// Store color in "background" (backward compat for canvas.go)
		parts := strings.Fields(value)
		for _, part := range parts {
			lower := strings.ToLower(part)
			if strings.HasPrefix(lower, "url(") {
				props["background-image"] = part
				props["background"] = value // keep full shorthand too
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
			} else if lower == "transparent" || lower == "inherit" {
				props["background"] = part
			} else if ParseColor(part).A > 0 {
				props["background"] = part
			}
		}
		// If no specific part matched, store the whole value in background
		if props["background"] == "" {
			props["background"] = value
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
	case "flex-direction":
		props["flex-direction"] = value
	case "flex-wrap":
		props["flex-wrap"] = value
	case "align-content":
		props["align-content"] = value
	case "order":
		props["order"] = value
	case "flex-grow":
		props["flex-grow"] = value
	case "flex-shrink":
		props["flex-shrink"] = value
	case "flex-basis":
		props["flex-basis"] = value
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

		if !setDuration && (strings.HasSuffix(part, "s") || strings.HasSuffix(part, "ms")) {
			props["animation-duration"] = part
			setDuration = true
		} else if !setTimingFunc && timingFuncs[lower] {
			props["animation-timing-function"] = part
			setTimingFunc = true
		} else if !setDelay && (strings.HasSuffix(part, "s") || strings.HasSuffix(part, "ms")) && !setDuration {
			// This would be a second duration — treat as delay
			props["animation-delay"] = part
			setDelay = true
		} else if !setDelay && (strings.HasSuffix(part, "s") || strings.HasSuffix(part, "ms")) && setDuration && !setTimingFunc {
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
