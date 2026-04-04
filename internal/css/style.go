package css

import "strings"

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
		"justify-content": "flex-start",
		"align-items":     "stretch",
		"flex-wrap":       "nowrap",
		"gap":             "0",
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
// Supports: tag, .class, #id, tag.class, tag#id
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

		if sel[0] == '.' {
			sel = sel[1:]
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' {
				end++
			}
			selClass = sel[:end]
			sel = sel[end:]
			seenClass = true
		} else if sel[0] == '#' {
			sel = sel[1:]
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' {
				end++
			}
			selID = sel[:end]
			sel = sel[end:]
			seenID = true
		} else {
			end := 0
			for end < len(sel) && sel[end] != '.' && sel[end] != '#' {
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

	return true
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
	}
}
