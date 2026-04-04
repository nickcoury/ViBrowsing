package css

import (
	"strconv"
	"strings"
)

// Rule represents a CSS rule: selector { declarations }.
type Rule struct {
	Selector    string
	Declarations []Declaration
	MediaQuery  string // e.g., "@media (max-width: 768px)" or "" for all
}

// Declaration represents a CSS declaration: property: value.
type Declaration struct {
	Property string
	Value    string
}

// Parse parses a CSS stylesheet and returns a list of rules.
func Parse(sheet string) []Rule {
	var rules []Rule
	sheet = strings.ReplaceAll(sheet, "\r\n", "\n")
	sheet = strings.ReplaceAll(sheet, "\r", "\n")

	i := 0
	for i < len(sheet) {
		// Skip whitespace and comments
		i = skipWhitespace(sheet, i)
		if i >= len(sheet) {
			break
		}

		// Handle @-rules
		if sheet[i] == '@' {
			// Find the end of the @ rule — could be a semicolon or an opening {
			j := i + 1
			for j < len(sheet) && sheet[j] != ';' && sheet[j] != '{' && sheet[j] != '}' {
				j++
			}
			if j >= len(sheet) {
				break
			}
			if sheet[j] == '{' {
				// Block @ rule like @media { ... } — parse it
				atName := strings.TrimSpace(sheet[i+1 : j])
				if strings.HasPrefix(atName, "media") {
					// Parse @media query — extract condition from "@media ..."
					mediaStart := j + 1
					mediaDepth := 1
					for mediaDepth > 0 && mediaStart < len(sheet) {
						if sheet[mediaStart] == '{' {
							mediaDepth++
						} else if sheet[mediaStart] == '}' {
							mediaDepth--
						}
						mediaStart++
					}
					mediaContent := sheet[j+1 : mediaStart-1]
					mediaCond := strings.TrimSpace(mediaContent)
					// Recursively parse the rules inside @media
					innerRules := Parse(mediaContent)
					for _, r := range innerRules {
						r.MediaQuery = mediaCond
						rules = append(rules, r)
					}
					i = mediaStart
					continue
				}
				// Skip the block
				i = j + 1
				i = skipBlock(sheet, i)
			} else if sheet[j] == ';' {
				// Single-line @ rule like @import url(foo.css); — just skip to semicolon
				i = j + 1
			} else {
				// No semicolon or brace found, skip past what we scanned and continue
				i = j
			}
			continue
		}

		// Find selector (until {)
		selectorStart := i
		for i < len(sheet) && sheet[i] != '{' {
			i++
		}
		selector := strings.TrimSpace(sheet[selectorStart:i])
		if i >= len(sheet) {
			break
		}
		i++ // skip '{'

		// Find declarations block
		decls, next := parseDeclarations(sheet, i)
		i = next

		if selector != "" && len(decls) > 0 {
			// Split multiple selectors (comma-separated)
			for _, sel := range strings.Split(selector, ",") {
				sel = strings.TrimSpace(sel)
				if sel != "" {
					rules = append(rules, Rule{
						Selector:    sel,
						Declarations: decls,
					})
				}
			}
		}
	}

	return rules
}

// MatchesMediaQuery returns true if the rule's media query matches the viewport.
func (r Rule) MatchesMediaQuery(viewportWidth, viewportHeight float64) bool {
	if r.MediaQuery == "" {
		return true
	}
	cond := r.MediaQuery
	if strings.Contains(cond, "max-width") {
		parts := strings.Split(cond, "max-width")
		if len(parts) > 1 {
			numStr := strings.TrimSpace(parts[1])
			numStr = strings.Trim(numStr, ":)")
			if w, err := strconv.ParseFloat(numStr, 64); err == nil {
				if viewportWidth > w {
					return false
				}
			}
		}
	}
	if strings.Contains(cond, "min-width") {
		parts := strings.Split(cond, "min-width")
		if len(parts) > 1 {
			numStr := strings.TrimSpace(parts[1])
			numStr = strings.Trim(numStr, ":)")
			if w, err := strconv.ParseFloat(numStr, 64); err == nil {
				if viewportWidth < w {
					return false
				}
			}
		}
	}
	return true
}

// ParseInline parses an inline style attribute value.
func ParseInline(style string) []Declaration {
	decls, _ := parseDeclarations(style, 0)
	return decls
}

func parseDeclarations(sheet string, start int) ([]Declaration, int) {
	var decls []Declaration
	i := start

	for i < len(sheet) {
		i = skipWhitespace(sheet, i)
		if i >= len(sheet) {
			break
		}

		// End of block
		if sheet[i] == '}' {
			return decls, i + 1
		}

		// Find property name
		propStart := i
		for i < len(sheet) && sheet[i] != ':' && sheet[i] != '!' {
			i++
		}
		prop := strings.TrimSpace(sheet[propStart:i])

		if i >= len(sheet) || sheet[i] != ':' {
			// Skip to next semicolon or end of block
			i = skipUntil(sheet, i, ';')
			if i < len(sheet) && sheet[i] == ';' {
				i++
			}
			continue
		}
		i++ // skip ':'

		// Find value
		valueStart := i
		for i < len(sheet) && sheet[i] != ';' && sheet[i] != '!' && sheet[i] != '}' {
			i++
		}
		value := strings.TrimSpace(sheet[valueStart:i])

		if i < len(sheet) && sheet[i] == '!' {
			// !important — skip it
			for i < len(sheet) && sheet[i] != ';' && sheet[i] != '}' {
				i++
			}
		}

		if prop != "" {
			decls = append(decls, Declaration{Property: prop, Value: value})
		}

		if i < len(sheet) && sheet[i] == ';' {
			i++
		}
	}

	return decls, i
}

func skipWhitespace(s string, i int) int {
	for i < len(s) {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' {
			i++
		} else if c == '/' {
			// Comment
			if i+1 < len(s) && s[i+1] == '*' {
				i += 2
				for i+1 < len(s) && !(s[i] == '*' && s[i+1] == '/') {
					i++
				}
				i += 2
			} else {
				break
			}
		} else {
			break
		}
	}
	return i
}

func skipUntil(s string, i int, target byte) int {
	for i < len(s) && s[i] != target && s[i] != '}' {
		i++
	}
	return i
}

func skipBlock(s string, i int) int {
	depth := 0
	for i < len(s) {
		if s[i] == '{' {
			depth++
			i++
		} else if s[i] == '}' {
			if depth == 0 {
				return i + 1
			}
			depth--
			i++
		} else {
			i++
		}
	}
	return i
}
