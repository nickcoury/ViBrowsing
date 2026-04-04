package css

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nickcoury/ViBrowsing/internal/html"
)

// ParseNthChild parses an :nth-child formula like "2n+1", "odd", "even", "3n", "+5", "-4", "5".
// Returns a (multiplier) and b (offset), or an error if the formula is invalid.
// The formula follows the pattern: an + b where n is 0, 1, 2, 3, ...
//
// Examples:
//   - "odd" or "2n+1" → a=2, b=1
//   - "even" or "2n" → a=2, b=0
//   - "3n" → a=3, b=0
//   - "3n+1" → a=3, b=1
//   - "5" → a=0, b=5 (selects only the 5th element)
//   - "-n+3" → a=-1, b=3 (selects first 3 elements)
func ParseNthChild(formula string) (a, b int, err error) {
	formula = strings.TrimSpace(formula)
	formula = strings.ToLower(formula)

	switch formula {
	case "odd":
		return 2, 1, nil
	case "even":
		return 2, 0, nil
	}

	// Try to parse as integer first (e.g., "5")
	if v, e := strconv.Atoi(formula); e == nil {
		return 0, v, nil
	}

	// Parse complex formula: [a]n[[+|-]b]
	sign := 1
	i := 0

	// Handle leading sign
	if i < len(formula) && (formula[i] == '+' || formula[i] == '-') {
		if formula[i] == '-' {
			sign = -1
		}
		i++
	}

	// Parse a (the multiplier)
	a = 1
	if i < len(formula) && formula[i] == 'n' {
		// Just "n" or "-n" means a=1
		i++
	} else {
		// Parse digits before 'n'
		start := i
		for i < len(formula) && formula[i] >= '0' && formula[i] <= '9' {
			i++
		}
		if i > start {
			if parsed, e := strconv.Atoi(formula[start:i]); e == nil {
				a = parsed
			}
		}
		if i < len(formula) && formula[i] == 'n' {
			i++
		} else if start == i {
			// No 'n' found and no digits, this might be just a number with sign
			// Like "-5" which should be a=0, b=-5
			// Only valid if the ENTIRE remaining string is a number
			remaining := formula[i:]
			if _, err := strconv.Atoi(remaining); err == nil {
				return 0, sign * a, nil
			}
			// Otherwise this is an invalid formula (e.g., "abc", "")
			return 0, 0, fmt.Errorf("invalid nth-child formula: %s", formula)
		}
	}

	// Apply sign to multiplier
	a *= sign

	// Parse b (the offset)
	b = 0
	if i < len(formula) {
		// Skip optional + or - (already handled leading sign)
		bSign := 1
		if formula[i] == '+' || formula[i] == '-' {
			if formula[i] == '-' {
				bSign = -1
			}
			i++
		}
		start := i
		for i < len(formula) && formula[i] >= '0' && formula[i] <= '9' {
			i++
		}
		if i > start {
			if parsed, e := strconv.Atoi(formula[start:i]); e == nil {
				b = parsed * bSign
			}
		} else {
			// No digits found after sign — this is invalid (e.g., "n+", "2n+")
			return 0, 0, fmt.Errorf("invalid nth-child formula: %s", formula)
		}
	}

	return a, b, nil
}

// MatchNthChild checks if a node matches the :nth-child() selector.
// This handles :nth-child, :nth-last-child, :nth-of-type, and :nth-last-of-type.
//
// Arguments:
//   - node: the HTML node to check
//   - formula: the nth formula (e.g., "2n+1", "odd", "even")
//   - ofType: if true, only count siblings of the same element type
//   - last: if true, count from the end (for :nth-last-child)
func MatchNthChild(node *html.Node, formula string, ofType, last bool) bool {
	if node.Parent == nil {
		return false
	}
	if formula == "" {
		return false
	}

	// Parse the formula
	a, b, err := ParseNthChild(formula)
	if err != nil {
		return false
	}

	// Collect matching siblings
	var siblings []*html.Node
	tag := node.TagName
	for _, child := range node.Parent.Children {
		if child.Type != html.NodeElement {
			continue
		}
		if ofType && child.TagName != tag {
			continue
		}
		siblings = append(siblings, child)
	}

	// Find the index of this node among matching siblings
	var index int
	for i, sib := range siblings {
		if sib == node {
			if last {
				// For :nth-last-child, reverse the index
				index = len(siblings) - 1 - i
			} else {
				index = i
			}
			break
		}
	}

	// Apply an + b formula using 1-indexed positions
	// CSS nth-child: 1st child = position 1, 2nd = position 2, etc.
	// DOM array is 0-indexed, so position = index + 1
	if a == 0 {
		return index+1 == b
	}
	// We need to find n >= 0 such that pos = a*n + b
	// Rearranging: pos - b = a*n, so n = (pos - b) / a
	// For n to be valid (non-negative integer), (pos - b) must be divisible by a
	// and the result must be >= 0.
	// Use (b - pos) to handle negative a values: n = (b - pos) / (-a)
	// The sign of a determines the direction.
	pos := index + 1
	diff := pos - b
	if diff == 0 {
		return true // n = 0 is always valid
	}
	if a > 0 {
		// For positive a: pos must be >= b and (pos-b) % a == 0
		return diff > 0 && diff%a == 0
	}
	// For negative a: pos must be <= b and (b-pos) % (-a) == 0
	return diff < 0 && (-diff)%(-a) == 0
}

// MatchNot checks if a node does NOT match the given selector.
// This properly handles complex selectors including selector lists.
func MatchNot(node *html.Node, selector string) bool {
	if selector == "" {
		return true
	}
	// Handle selector lists (comma-separated)
	// :not(.foo, .bar) matches if the element matches NEITHER .foo NOR .bar
	selectors := strings.Split(selector, ",")
	for _, sel := range selectors {
		sel = strings.TrimSpace(sel)
		if sel != "" && MatchNodeSelector(node, sel) {
			return false
		}
	}
	return true
}

// IsValid checks if a form element has valid content.
// For input/textarea elements with required attribute, checks:
// - Required attribute is present and value is non-empty
// - For specific types (email, url, number), validates format
func IsValid(node *html.Node) bool {
	if node.Type != html.NodeElement {
		return false
	}

	tagName := strings.ToLower(node.TagName)
	if tagName != "input" && tagName != "textarea" {
		return true
	}

	required := node.HasAttribute("required")
	value := node.GetAttribute("value")
	inputType := strings.ToLower(node.GetAttribute("type"))

	// If required attribute is present (even with empty value) AND value is empty, the input is invalid
	if required && strings.TrimSpace(value) == "" {
		return false
	}

	// If no value and not required, it's valid
	if value == "" {
		return true
	}

	// Type-specific validation
	switch inputType {
	case "email":
		// Basic email validation - must contain @
		if !strings.Contains(value, "@") {
			return false
		}
	case "url":
		// Basic URL validation - must start with http:// or https://
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			return false
		}
	case "number":
		// Check if it's a valid number
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return false
		}
		// Check min/max constraints
		if min := node.GetAttribute("min"); min != "" {
			if minVal, err := strconv.ParseFloat(min, 64); err == nil {
				if val, err := strconv.ParseFloat(value, 64); err == nil && val < minVal {
					return false
				}
			}
		}
		if max := node.GetAttribute("max"); max != "" {
			if maxVal, err := strconv.ParseFloat(max, 64); err == nil {
				if val, err := strconv.ParseFloat(value, 64); err == nil && val > maxVal {
					return false
				}
			}
		}
	case "text":
		// Check minlength constraint
		if minlength := node.GetAttribute("minlength"); minlength != "" {
			if minLen, err := strconv.Atoi(minlength); err == nil && len(value) < minLen {
				return false
			}
		}
		// Check maxlength constraint
		if maxlength := node.GetAttribute("maxlength"); maxlength != "" {
			if maxLen, err := strconv.Atoi(maxlength); err == nil && len(value) > maxLen {
				return false
			}
		}
	}

	return true
}

// IsInvalid returns true if the form element has invalid content.
// This is the inverse of IsValid.
func IsInvalid(node *html.Node) bool {
	return !IsValid(node)
}

// PlaceholderShown checks if the placeholder text is currently visible.
// Placeholder is shown when: element has placeholder attribute AND value is empty.
func PlaceholderShown(node *html.Node) bool {
	if node.Type != html.NodeElement {
		return false
	}

	tagName := strings.ToLower(node.TagName)
	if tagName != "input" && tagName != "textarea" {
		return false
	}

	placeholder := node.GetAttribute("placeholder")
	value := node.GetAttribute("value")

	// Placeholder is shown when there's a placeholder attribute and no value
	return placeholder != "" && value == ""
}
