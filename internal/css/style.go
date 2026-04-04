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
		"display":        "inline",
		"color":          "black",
		"background":     "transparent",
		"font-size":      "16px",
		"font-family":    "serif",
		"font-weight":    "normal",
		"margin-top":     "0",
		"margin-right":   "0",
		"margin-bottom":  "0",
		"margin-left":    "0",
		"padding-top":    "0",
		"padding-right":  "0",
		"padding-bottom": "0",
		"padding-left":   "0",
		"border-width":    "0",
		"border-style":   "none",
		"border-color":   "black",
		"width":          "auto",
		"height":         "auto",
		"text-align":     "left",
		"line-height":    "1.2",
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
	case "background-color":
		props["background"] = value
	case "background":
		props["background"] = value
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
	case "text-align":
		props["text-align"] = value
	case "line-height":
		props["line-height"] = value
	case "font-style":
		props["font-style"] = value
	}
}
