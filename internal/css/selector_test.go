package css

import (
	"testing"

	"github.com/nickcoury/ViBrowsing/internal/html"
)

// TestParseNthChild tests the ParseNthChild function with various formulas.
func TestParseNthChild(t *testing.T) {
	tests := []struct {
		formula string
		wantA   int
		wantB   int
		wantErr bool
	}{
		// Keywords
		{"odd", 2, 1, false},
		{"even", 2, 0, false},

		// Simple integers
		{"0", 0, 0, false},
		{"5", 0, 5, false},
		{"1", 0, 1, false},

		// Negative integers
		{"-1", 0, -1, false},
		{"-5", 0, -5, false},

		// Formulas with n
		{"n", 1, 0, false},
		{"N", 1, 0, false}, // case insensitive
		{"+n", 1, 0, false},
		{"-n", -1, 0, false},

		// Formulas with coefficient
		{"2n", 2, 0, false},
		{"3n", 3, 0, false},
		{"10n", 10, 0, false},

		// Formulas with negative coefficient
		{"-2n", -2, 0, false},
		{"-3n", -3, 0, false},

		// Formulas with n and offset
		{"2n+1", 2, 1, false},
		{"2n-1", 2, -1, false},
		{"3n+1", 3, 1, false},
		{"3n-1", 3, -1, false},
		{"2n+0", 2, 0, false},
		{"2n-0", 2, 0, false},

		// Just offset with sign (no n)
		{"+5", 0, 5, false},
		{"-5", 0, -5, false},

		// Edge cases
		{"0n", 0, 0, false},    // 0n means 0
		{"0n+1", 0, 1, false},   // 0n+1 means 1

		// Invalid formulas
		{"", 0, 0, true},
		{"abc", 0, 0, true},
		{"n+", 0, 0, true},
		{"2n+", 0, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.formula, func(t *testing.T) {
			a, b, err := ParseNthChild(tc.formula)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ParseNthChild(%q) expected error, got nil", tc.formula)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseNthChild(%q) unexpected error: %v", tc.formula, err)
				return
			}
			if a != tc.wantA || b != tc.wantB {
				t.Errorf("ParseNthChild(%q) = (%d, %d), want (%d, %d)",
					tc.formula, a, b, tc.wantA, tc.wantB)
			}
		})
	}
}

// TestMatchNthOfType tests the MatchNthChild function with ofType=true for :nth-of-type pseudo-class.
// This tests that only siblings of the same element type are counted.
func TestMatchNthOfType(t *testing.T) {
	// Build a mixed DOM tree:
	// parent
	//   - div (index 0)
	//   - span (index 1)
	//   - div (index 2)
	//   - p (index 3)
	//   - div (index 4)
	//   - span (index 5)
	parent := html.NewElement("parent")
	elements := []string{"div", "span", "div", "p", "div", "span"}
	for i, tag := range elements {
		el := html.NewElement(tag)
		el.SetAttribute("id", "el"+string(rune('0'+i)))
		parent.AppendChild(el)
	}

	children := parent.Children

	tests := []struct {
		name    string
		target  int // which child (0-5) to test
		formula string
		last    bool
		want    bool
	}{
		// For div elements: indices 0, 2, 4 → positions among divs are 1, 2, 3
		// :nth-of-type(odd) on divs → 1st and 3rd div
		{"div:nth-of-type(odd) on 1st div", 0, "odd", false, true},
		{"div:nth-of-type(odd) on 2nd div", 2, "odd", false, false},
		{"div:nth-of-type(odd) on 3rd div", 4, "odd", false, true},

		// :nth-of-type(even) on divs → 2nd div only
		{"div:nth-of-type(even) on 1st div", 0, "even", false, false},
		{"div:nth-of-type(even) on 2nd div", 2, "even", false, true},
		{"div:nth-of-type(even) on 3rd div", 4, "even", false, false},

		// :nth-of-type(2n+1) = odd on divs
		{"div:nth-of-type(2n+1) 1st", 0, "2n+1", false, true},
		{"div:nth-of-type(2n+1) 2nd", 2, "2n+1", false, false},
		{"div:nth-of-type(2n+1) 3rd", 4, "2n+1", false, true},

		// :nth-of-type(3n) on divs → 3rd div only (position 3)
		{"div:nth-of-type(3n) 1st", 0, "3n", false, false},
		{"div:nth-of-type(3n) 2nd", 2, "3n", false, false},
		{"div:nth-of-type(3n) 3rd", 4, "3n", false, true},

		// :nth-of-type(2) on divs → 2nd div only
		{"div:nth-of-type(2) 1st", 0, "2", false, false},
		{"div:nth-of-type(2) 2nd", 2, "2", false, true},
		{"div:nth-of-type(2) 3rd", 4, "2", false, false},

		// For span elements: indices 1, 5 → positions among spans are 1, 2
		// :nth-of-type(odd) on spans → 1st span
		{"span:nth-of-type(odd) on 1st span", 1, "odd", false, true},
		{"span:nth-of-type(odd) on 2nd span", 5, "odd", false, false},

		// :nth-of-type(even) on spans → 2nd span
		{"span:nth-of-type(even) on 1st span", 1, "even", false, false},
		{"span:nth-of-type(even) on 2nd span", 5, "even", false, true},

		// :nth-last-of-type (counting from end)
		// For divs from end: last div is position 1, second-to-last is position 2, etc.
		// div:nth-last-of-type(1) → last div (index 4)
		{"div:nth-last-of-type(1) last div", 4, "1", true, true},
		{"div:nth-last-of-type(1) not last div", 0, "1", true, false},
		// div:nth-last-of-type(2) → second-to-last div (index 2)
		{"div:nth-last-of-type(2) middle div", 2, "2", true, true},
		{"div:nth-last-of-type(2) last div", 4, "2", true, false},

		// :nth-last-of-type(odd) on divs from end
		// Last div = position 1 (odd), second-to-last = position 2 (even)
		{"div:nth-last-of-type(odd) last div", 4, "odd", true, true},
		{"div:nth-last-of-type(odd) middle div", 2, "odd", true, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Use ofType=true for :nth-of-type matching
			got := MatchNthChild(children[tc.target], tc.formula, true, tc.last)
			if got != tc.want {
				t.Errorf("MatchNthChild(node at index %d (%s), %q, ofType=true, last=%v) = %v, want %v",
					tc.target, children[tc.target].TagName, tc.formula, tc.last, got, tc.want)
			}
		})
	}
}

// TestMatchNthChild tests the MatchNthChild function.
func TestMatchNthChild(t *testing.T) {
	// Build a simple DOM tree for testing:
	// parent
	//   - div1 (index 0)
	//   - div2 (index 1)
	//   - div3 (index 2)
	//   - div4 (index 3)
	//   - div5 (index 4)
	parent := html.NewElement("parent")
	for i := 1; i <= 5; i++ {
		div := html.NewElement("div")
		div.SetAttribute("id", "div"+string(rune('0'+i)))
		parent.AppendChild(div)
	}

	children := parent.Children

	tests := []struct {
		name    string
		target  int // which child (0-4) to test
		formula string
		ofType  bool
		last    bool
		want    bool
	}{
		// :nth-child(odd) - 1st, 3rd, 5th child (1-indexed)
		{"odd div1", 0, "odd", false, false, true},
		{"odd div2", 1, "odd", false, false, false},
		{"odd div3", 2, "odd", false, false, true},
		{"odd div4", 3, "odd", false, false, false},
		{"odd div5", 4, "odd", false, false, true},

		// :nth-child(even) - 2nd, 4th child (1-indexed)
		{"even div1", 0, "even", false, false, false},
		{"even div2", 1, "even", false, false, true},
		{"even div3", 2, "even", false, false, false},
		{"even div4", 3, "even", false, false, true},
		{"even div5", 4, "even", false, false, false},

		// :nth-child(2n+1) = odd
		{"2n+1 div1", 0, "2n+1", false, false, true},
		{"2n+1 div2", 1, "2n+1", false, false, false},
		{"2n+1 div3", 2, "2n+1", false, false, true},

		// :nth-child(3n) - 3rd child only
		{"3n div1", 0, "3n", false, false, false},
		{"3n div2", 1, "3n", false, false, false},
		{"3n div3", 2, "3n", false, false, true},
		{"3n div4", 3, "3n", false, false, false},
		{"3n div5", 4, "3n", false, false, false},

		// :nth-child(3n+1) - 1st, 4th child
		{"3n+1 div1", 0, "3n+1", false, false, true},
		{"3n+1 div4", 3, "3n+1", false, false, true},

		// :nth-child(2) - 2nd child only
		{"2 div1", 0, "2", false, false, false},
		{"2 div2", 1, "2", false, false, true},

		// :nth-child(-n+3) - first 3 children
		{"-n+3 div1", 0, "-n+3", false, false, true},
		{"-n+3 div2", 1, "-n+3", false, false, true},
		{"-n+3 div3", 2, "-n+3", false, false, true},
		{"-n+3 div4", 3, "-n+3", false, false, false},
		{"-n+3 div5", 4, "-n+3", false, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MatchNthChild(children[tc.target], tc.formula, tc.ofType, tc.last)
			if got != tc.want {
				t.Errorf("MatchNthChild(node at index %d, %q, ofType=%v, last=%v) = %v, want %v",
					tc.target, tc.formula, tc.ofType, tc.last, got, tc.want)
			}
		})
	}
}

// TestMatchNot tests the MatchNot function.
func TestMatchNot(t *testing.T) {
	// Build a simple DOM tree
	parent := html.NewElement("div")
	parent.SetAttribute("class", "container")

	child1 := html.NewElement("p")
	child1.SetAttribute("class", "special")
	parent.AppendChild(child1)

	child2 := html.NewElement("p")
	parent.AppendChild(child2)

	span := html.NewElement("span")
	span.SetAttribute("id", "myspan")
	parent.AppendChild(span)

	tests := []struct {
		name     string
		node     *html.Node
		selector string
		want     bool
	}{
		// :not(tag) - should NOT match the tag
		{"p:not(p)", child1, "p", false},
		{"p:not(span)", child1, "span", true},
		{"span:not(p)", span, "p", true},

		// :not(.class)
		{"p:not(.special)", child1, ".special", false},
		{"p:not(.special)", child2, ".special", true},
		{"p:not(.nonexistent)", child1, ".nonexistent", true},

		// :not(#id)
		{"span:not(#myspan)", span, "#myspan", false},
		{"p:not(#myspan)", child1, "#myspan", true},

		// :not([attr])
		{"p:not([class])", child1, "[class]", false},
		{"p:not([class])", child2, "[class]", true},

		// Empty selector always matches
		{"any:not()", child1, "", true},

		// Complex selectors
		{"p:not(.special, .nonexistent)", child1, ".special, .nonexistent", false},
		{"p:not(.nonexistent, .alsononexistent)", child1, ".nonexistent, .alsononexistent", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MatchNot(tc.node, tc.selector)
			if got != tc.want {
				t.Errorf("MatchNot(%v, %q) = %v, want %v",
					tc.node.TagName, tc.selector, got, tc.want)
			}
		})
	}
}

// TestIsValid tests the IsValid function for form validation.
func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		attrs    map[string]string
		want     bool
	}{
		// Not a form element - always valid
		{"div no attrs", "div", map[string]string{}, true},
		{"span no attrs", "span", map[string]string{}, true},

		// Input with no attributes - valid
		{"input no attrs", "input", map[string]string{}, true},

		// Required but empty - invalid
		{"input required empty", "input", map[string]string{"required": ""}, false},
		{"textarea required empty", "textarea", map[string]string{"required": ""}, false},

		// Required with value - valid
		{"input required with value", "input", map[string]string{"required": "", "value": "hello"}, true},

		// Type email validation
		{"email valid", "input", map[string]string{"type": "email", "value": "test@example.com"}, true},
		{"email invalid (no @)", "input", map[string]string{"type": "email", "value": "testexample.com"}, false},

		// Type url validation
		{"url valid http", "input", map[string]string{"type": "url", "value": "http://example.com"}, true},
		{"url valid https", "input", map[string]string{"type": "url", "value": "https://example.com"}, true},
		{"url invalid", "input", map[string]string{"type": "url", "value": "example.com"}, false},

		// Type number validation
		{"number valid", "input", map[string]string{"type": "number", "value": "42"}, true},
		{"number invalid", "input", map[string]string{"type": "number", "value": "abc"}, false},
		{"number with min", "input", map[string]string{"type": "number", "value": "5", "min": "10"}, false},
		{"number with max", "input", map[string]string{"type": "number", "value": "15", "max": "10"}, false},

		// Minlength validation
		{"text minlength valid", "input", map[string]string{"type": "text", "value": "hello", "minlength": "3"}, true},
		{"text minlength invalid", "input", map[string]string{"type": "text", "value": "hi", "minlength": "5"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			node := html.NewElement(tc.tag)
			for k, v := range tc.attrs {
				node.SetAttribute(k, v)
			}

			got := IsValid(node)
			if got != tc.want {
				t.Errorf("IsValid(%s %v) = %v, want %v", tc.tag, tc.attrs, got, tc.want)
			}
		})
	}
}

// TestIsInvalid tests the IsInvalid function.
func TestIsInvalid(t *testing.T) {
	// Basic test - just inverse of IsValid
	node := html.NewElement("input")
	node.SetAttribute("required", "")
	node.SetAttribute("type", "email")
	node.SetAttribute("value", "invalid-email")

	if IsInvalid(node) != true {
		t.Error("IsInvalid should return true for invalid email input")
	}

	node2 := html.NewElement("input")
	node2.SetAttribute("type", "email")
	node2.SetAttribute("value", "valid@example.com")

	if IsInvalid(node2) != false {
		t.Error("IsInvalid should return false for valid email input")
	}
}

// TestPlaceholderShown tests the PlaceholderShown function.
func TestPlaceholderShown(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		attrs     map[string]string
		wantShown bool
	}{
		// Non-form elements - never show placeholder
		{"div no attrs", "div", map[string]string{}, false},
		{"span no attrs", "span", map[string]string{}, false},

		// Input with placeholder but no value - shown
		{"input placeholder no value", "input", map[string]string{"placeholder": "Enter name"}, true},

		// Input with placeholder AND value - not shown
		{"input placeholder with value", "input", map[string]string{"placeholder": "Enter name", "value": "John"}, false},

		// Input with value but no placeholder - not shown
		{"input no placeholder with value", "input", map[string]string{"value": "John"}, false},

		// Empty placeholder - not shown
		{"input empty placeholder", "input", map[string]string{"placeholder": ""}, false},

		// Textarea with placeholder - shown when no value
		{"textarea placeholder no value", "textarea", map[string]string{"placeholder": "Enter text"}, true},

		// Textarea with placeholder and value - not shown
		{"textarea placeholder with value", "textarea", map[string]string{"placeholder": "Enter text", "value": "Some text"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			node := html.NewElement(tc.tag)
			for k, v := range tc.attrs {
				node.SetAttribute(k, v)
			}

			got := PlaceholderShown(node)
			if got != tc.wantShown {
				t.Errorf("PlaceholderShown(%s %v) = %v, want %v",
					tc.tag, tc.attrs, got, tc.wantShown)
			}
		})
	}
}

// TestMatchNodeSelectorNot tests the full :not() integration with selector matching.
func TestMatchNodeSelectorNot(t *testing.T) {
	doc := html.NewDocument()
	root := html.NewElement("html")
	doc.AppendChild(root)

	body := html.NewElement("body")
	body.SetAttribute("class", "main-content")
	root.AppendChild(body)

	// Add form elements
	form := html.NewElement("form")
	form.SetAttribute("id", "myform")
	body.AppendChild(form)

	input1 := html.NewElement("input")
	input1.SetAttribute("type", "text")
	input1.SetAttribute("class", "required")
	form.AppendChild(input1)

	input2 := html.NewElement("input")
	input2.SetAttribute("type", "email")
	input2.SetAttribute("class", "optional")
	form.AppendChild(input2)

	div := html.NewElement("div")
	div.SetAttribute("class", "notice")
	body.AppendChild(div)

	tests := []struct {
		selector string
		matches  int // how many elements should match
	}{
		{"input:not([type=email])", 1}, // only input1
		{"input:not(.optional)", 1},    // only input1
		{"input:not(.required)", 1},     // only input2
		{".required:not(input)", 0},     // no element matches
		{"*:not(body)", 5},              // html, form, input1, input2, div = 5 non-body elements
	}

	for _, tc := range tests {
		t.Run(tc.selector, func(t *testing.T) {
			matches := doc.QuerySelectorAll(tc.selector)
			if len(matches) != tc.matches {
				t.Errorf("QuerySelectorAll(%q) returned %d matches, want %d",
					tc.selector, len(matches), tc.matches)
			}
		})
	}
}

// TestFirstLastOnlyChild tests :first-child, :last-child, and :only-child pseudo-classes.
func TestFirstLastOnlyChild(t *testing.T) {
	doc := html.NewDocument()
	root := html.NewElement("html")
	doc.AppendChild(root)

	// Build test structure:
	// body
	//   - div#first (first child)
	//   - div#middle (middle child)
	//   - div#last (last child)
	//   - span#only (only child of its parent)
	body := html.NewElement("body")
	root.AppendChild(body)

	divFirst := html.NewElement("div")
	divFirst.SetAttribute("id", "first")
	body.AppendChild(divFirst)

	divMiddle := html.NewElement("div")
	divMiddle.SetAttribute("id", "middle")
	body.AppendChild(divMiddle)

	divLast := html.NewElement("div")
	divLast.SetAttribute("id", "last")
	body.AppendChild(divLast)

	spanOnly := html.NewElement("span")
	spanOnly.SetAttribute("id", "only")
	divMiddle.AppendChild(spanOnly)

	tests := []struct {
		selector string
		matches  int
	}{
		// :first-child
		{"#first:first-child", 1},     // first div is first child
		{"#middle:first-child", 0},    // middle div is not first child
		{"#last:first-child", 0},      // last div is not first child
		{"#only:first-child", 1},      // span is first (and only) child of div

		// :last-child
		{"#first:last-child", 0},      // first div is not last child
		{"#middle:last-child", 0},     // middle div is not last child
		{"#last:last-child", 1},       // last div is last child
		{"#only:last-child", 1},       // span is last (and only) child of div

		// :only-child
		{"#first:only-child", 0},      // first div has siblings
		{"#middle:only-child", 0},     // middle div has siblings
		{"#last:only-child", 0},       // last div has siblings
		{"#only:only-child", 1},       // span is only child of its parent

		// Combined selectors
		{"div:first-child", 1},        // only one div is first child
		{"div:last-child", 1},          // only one div is last child
		{"div:only-child", 0},          // no div is only child (they all have siblings)
	}

	for _, tc := range tests {
		t.Run(tc.selector, func(t *testing.T) {
			matches := doc.QuerySelectorAll(tc.selector)
			if len(matches) != tc.matches {
				t.Errorf("QuerySelectorAll(%q) returned %d matches, want %d", tc.selector, len(matches), tc.matches)
			}
		})
	}
}
