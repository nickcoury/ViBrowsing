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

// KeyframeRule represents a @keyframes rule: name { percentage { props } ... }.
type KeyframeRule struct {
	Name     string
	Keyframes map[float64]map[string]string // percentage (0-100) -> properties
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
				if strings.HasPrefix(atName, "keyframes") {
					// Parse @keyframes block
					kf := parseKeyframes(sheet, i)
					if kf != nil {
						Keyframes = append(Keyframes, *kf)
					}
					i = skipBlock(sheet, j+1)
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

// Keyframes holds all parsed @keyframes rules.
var Keyframes []KeyframeRule

// parseKeyframes parses a @keyframes block starting at '@'.
func parseKeyframes(sheet string, start int) *KeyframeRule {
	// Find the @keyframes name
	j := start + 1
	for j < len(sheet) && sheet[j] != '{' && sheet[j] != ' ' && sheet[j] != '\t' {
		j++
	}
	atName := strings.TrimSpace(sheet[start+1 : j])
	// Extract keyframes name (skip "@keyframes" prefix)
	name := strings.TrimPrefix(atName, "keyframes")
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}

	// Find opening brace of keyframes block
	for j < len(sheet) && sheet[j] != '{' {
		j++
	}
	if j >= len(sheet) {
		return nil
	}
	j++ // skip '{'

	kf := &KeyframeRule{
		Name:     name,
		Keyframes: make(map[float64]map[string]string),
	}

	// Parse keyframe blocks: percentages followed by { declarations }
	for j < len(sheet) {
		j = skipWhitespace(sheet, j)
		if j >= len(sheet) {
			break
		}
		if sheet[j] == '}' {
			break
		}

		// Parse keyframe selectors: "from", "to", or "XX%"
		selStart := j
		for j < len(sheet) && sheet[j] != '{' && sheet[j] != '}' {
			j++
		}
		selStr := strings.TrimSpace(sheet[selStart:j])
		if sheet[j] != '{' {
			break
		}
		j++ // skip '{'

		// Parse declarations inside this keyframe block
		decls, next := parseDeclarations(sheet, j)
		j = next

		// Map each percentage in the selector to the declarations
		percentages := parseKeyframeSelectors(selStr)
		for _, pct := range percentages {
			if _, exists := kf.Keyframes[pct]; !exists {
				kf.Keyframes[pct] = make(map[string]string)
			}
			for _, decl := range decls {
				kf.Keyframes[pct][decl.Property] = decl.Value
			}
		}
	}

	return kf
}

// parseKeyframeSelectors parses a keyframe selector string like "from", "to", "50%", "0%, 100%", "25%, 75%".
func parseKeyframeSelectors(sel string) []float64 {
	var percentages []float64
	parts := strings.Split(sel, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		lower := strings.ToLower(part)
		if lower == "from" {
			percentages = append(percentages, 0)
		} else if lower == "to" {
			percentages = append(percentages, 100)
		} else if strings.HasSuffix(part, "%") {
			if val, err := strconv.ParseFloat(strings.TrimSuffix(part, "%"), 64); err == nil {
				percentages = append(percentages, val)
			}
		}
	}
	return percentages
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
