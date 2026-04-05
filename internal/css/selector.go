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

// parseSelectorList splits a selector list by commas, respecting nested parentheses.
// This is used by :is(), :where(), and :not() pseudo-classes.
func parseSelectorList(formula string) []string {
	var selectors []string
	var current strings.Builder
	depth := 0
	for i := 0; i < len(formula); i++ {
		c := formula[i]
		if c == '(' {
			depth++
			current.WriteByte(c)
		} else if c == ')' {
			depth--
			current.WriteByte(c)
		} else if c == ',' && depth == 0 {
			selectors = append(selectors, current.String())
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}
	if s := strings.TrimSpace(current.String()); s != "" {
		selectors = append(selectors, s)
	}
	return selectors
}

// MatchIsWherePseudoClass matches :is() and :where() pseudo-classes.
// These are "forgiving" selector list pseudo-classes that match if ANY selector
// in the list matches. :where() has 0 specificity (handled at specificity calculation)
// but the matching logic is the same.
// The isWhere parameter is just for debugging/logging and doesn't affect matching.
func MatchIsWherePseudoClass(node *html.Node, formula string) bool {
	if formula == "" {
		return false
	}
	// Split by comma, but respect nested parentheses
	selectors := parseSelectorList(formula)
	for _, sel := range selectors {
		sel = strings.TrimSpace(sel)
		if sel != "" && MatchNodeSelector(node, sel) {
			return true
		}
	}
	return false
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

// MatchHasSelector checks if a node matches the :has() pseudo-class with a relative selector.
// The relative selector is evaluated against descendants/ancestors/siblings of the current node.
// Returns true if ANY element matching the relative selector's starting point exists in the
// appropriate relationship to the current node.
//
// Examples:
//   - div:has(p) — div contains a p descendant
//   - img:has(+ p) — img has an adjacent sibling p immediately before it
//   - p:has(~ div) — p has a div sibling after it (general sibling)
func MatchHasSelector(node *html.Node, relativeSelector string) bool {
	if node == nil || node.Type != html.NodeElement {
		return false
	}
	if relativeSelector == "" {
		return false
	}

	relSel := strings.TrimSpace(relativeSelector)
	if relSel == "" {
		return false
	}

	// Parse the relative selector to find the first combinator and the initial selector
	// The initial selector (before any combinator) is matched against potential target nodes
	// found by traversing from node in the direction specified by the combinator.
	//
	// Combinators:
	//   " " (space) - descendant: any descendant of node matches the initial selector
	//   ">"           - child: any direct child of node matches the initial selector
	//   "+"           - adjacent sibling: immediately preceding sibling matches
	//   "~"           - general sibling: any preceding sibling matches

	// Find the first combinator
	firstCombinator := " " // Default is descendant
	combinatorIdx := -1
	inAttr := false
	inParen := 0

	for i := 0; i < len(relSel); i++ {
		c := relSel[i]
		if c == '[' {
			inAttr = true
		} else if c == ']' {
			inAttr = false
		} else if c == '(' {
			inParen++
		} else if c == ')' {
			inParen--
		} else if !inAttr && inParen == 0 {
			if c == '>' || c == '+' || c == '~' {
				firstCombinator = string(c)
				combinatorIdx = i
				break
			}
		}
	}

	var initialSelector string
	var remainingSelectors string

	if combinatorIdx >= 0 {
		// There is a combinator
		initialSelector = strings.TrimSpace(relSel[:combinatorIdx])
		remainingSelectors = strings.TrimSpace(relSel[combinatorIdx+1:])
	} else {
		// No combinator - default to descendant
		initialSelector = relSel
		remainingSelectors = ""
	}

	if initialSelector == "" {
		return false
	}

	switch firstCombinator {
	case ">":
		// Direct child: check immediate children of node
		for _, child := range node.Children {
			if child.Type == html.NodeElement && matchSimpleSelectorHas(child, initialSelector) {
				// If there are remaining selectors, verify them
				if remainingSelectors != "" {
					if matchRelativeChain(child, remainingSelectors) {
						return true
					}
				} else {
					return true
				}
			}
		}
		return false

	case "+":
		// Adjacent sibling: immediately preceding sibling
		prev := findPrecedingSiblingElement(node)
		if prev != nil && matchSimpleSelectorHas(prev, initialSelector) {
			if remainingSelectors != "" {
				return matchRelativeChain(prev, remainingSelectors)
			}
			return true
		}
		return false

	case "~":
		// General sibling: any preceding sibling
		prev := findAnyPrecedingSiblingElement(node)
		for prev != nil {
			if prev.Type == html.NodeElement && matchSimpleSelectorHas(prev, initialSelector) {
				if remainingSelectors != "" {
					if matchRelativeChain(prev, remainingSelectors) {
						return true
					}
				} else {
					return true
				}
			}
			// Move to previous sibling
			prev = findPrecedingSiblingElement(prev)
		}
		return false

	default:
		// Descendant (space): any descendant
		return hasMatchingDescendant(node, initialSelector, remainingSelectors)
	}
}

// findPrecedingSiblingElement returns the immediately preceding element sibling.
func findPrecedingSiblingElement(node *html.Node) *html.Node {
	if node.Parent == nil {
		return nil
	}
	found := false
	for _, sib := range node.Parent.Children {
		if sib == node {
			found = true
			break
		}
		if sib.Type == html.NodeElement {
			if found {
				return sib
			}
		}
	}
	return nil
}

// findAnyPrecedingSiblingElement returns the closest preceding element sibling.
func findAnyPrecedingSiblingElement(node *html.Node) *html.Node {
	if node.Parent == nil {
		return nil
	}
	var prev *html.Node
	for _, sib := range node.Parent.Children {
		if sib == node {
			break
		}
		if sib.Type == html.NodeElement {
			prev = sib
		}
	}
	return prev
}

// hasMatchingDescendant checks if any descendant of node matches the given selector.
func hasMatchingDescendant(node *html.Node, selector, remaining string) bool {
	for _, child := range node.Children {
		if child.Type != html.NodeElement {
			continue
		}
		if matchSimpleSelectorHas(child, selector) {
			if remaining != "" {
				if matchRelativeChain(child, remaining) {
					return true
				}
			} else {
				return true
			}
		}
		// Recursively check descendants
		if hasMatchingDescendant(child, selector, remaining) {
			return true
		}
	}
	return false
}

// matchSimpleSelectorHas is a specialized version of matchSimpleSelector
// for use in :has() matching. It only handles simple selectors (tag, class, id, attribute).
func matchSimpleSelectorHas(node *html.Node, sel string) bool {
	// Universal selector * matches everything
	if sel == "*" {
		return true
	}

	tagName := strings.ToLower(node.TagName)
	sel = strings.TrimSpace(sel)

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
			// Pseudo-class - for :has() simple matching, we handle :not() specially
			// Other pseudo-classes are treated as not matching for simplicity
			sel = sel[1:]
			if len(sel) == 0 {
				return false
			}
			if sel[0] == ':' {
				// Pseudo-element
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
				end := 0
				for end < len(sel) && (isAlphanumeric(sel[end]) || sel[end] == '-') {
					end++
				}
				pseudoName = sel[:end]
				sel = sel[end:]
			}
			pseudoName = strings.ToLower(pseudoName)

			// Handle :not() specially - it must match for the overall selector to match
			if pseudoName == "not" {
				if !MatchNot(node, pseudoArg) {
					return false
				}
			} else {
				// For other pseudo-classes in a :has() context,
				// we treat them as not matching since they require
				// more complex evaluation
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

// matchRelativeChain matches a chain of selectors with combinators.
// This is used for the part of the :has() selector after the initial selector.
func matchRelativeChain(node *html.Node, relSel string) bool {
	// Parse and match the relative selector chain
	parts := splitSelectorParts(relSel)
	if len(parts) == 0 {
		return false
	}

	// Start with the first selector
	curr := node
	if !matchSimpleSelectorHas(curr, parts[0]) {
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
			next = curr.Parent
		case "+":
			// Adjacent sibling: immediately preceding sibling
			next = findPrecedingSiblingElement(curr)
		case "~":
			// General sibling: any preceding sibling
			next = findAnyPrecedingSiblingElement(curr)
		default:
			// Descendant (space): any ancestor
			next = curr.Parent
		}

		if next == nil {
			return false
		}
		curr = next

		// If combinator was descendant, we need to walk up and find a matching ancestor
		if combinator == " " {
			found := false
			for curr != nil {
				if matchSimpleSelectorHas(curr, selector) {
					found = true
					break
				}
				curr = curr.Parent
			}
			if !found {
				return false
			}
		} else {
			if !matchSimpleSelectorHas(curr, selector) {
				return false
			}
		}
	}
	return true
}
