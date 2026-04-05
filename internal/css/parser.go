package css

import (
	"fmt"
	"strconv"
	"strings"
)

// Rule represents a CSS rule: selector { declarations }.
type Rule struct {
	Selector    string
	Declarations []Declaration
	MediaQuery  string // e.g., "@media (max-width: 768px)" or "" for all
	Layer       string // e.g., "utilities" or "" for unlayered rules. Anonymous layers use a generated unique name.
}

// LayerOrder tracks the declaration order of cascade layers.
// Earlier layers have lower priority than later layers.
type LayerOrder struct {
	layers []string // layer names in declaration order
}

// GetLayerPriority returns the priority for a layer (higher = later = higher priority).
// Unlayered rules (empty string) return math.MaxInt.
// Returns -1 if the layer is not yet registered.
func (lo *LayerOrder) GetLayerPriority(layer string) int {
	if layer == "" {
		return len(lo.layers) // unlayered has highest priority
	}
	for i, l := range lo.layers {
		if l == layer {
			return i
		}
	}
	return -1
}

// RegisterLayer adds a layer to the order if not already present.
func (lo *LayerOrder) RegisterLayer(layer string) {
	if layer == "" {
		return // don't register empty layer
	}
	for _, l := range lo.layers {
		if l == layer {
			return
		}
	}
	lo.layers = append(lo.layers, layer)
}

// KeyframeRule represents a @keyframes rule: name { percentage { props } ... }.
type KeyframeRule struct {
	Name     string
	Keyframes map[float64]map[string]string // percentage (0-100) -> properties
}

// anonymousLayerCounter is used to generate unique anonymous layer names.
var anonymousLayerCounter int

// nextAnonymousLayerName returns a unique name for an anonymous @layer block.
func nextAnonymousLayerName() string {
	anonymousLayerCounter++
	return fmt.Sprintf("__anonymous_layer_%d__", anonymousLayerCounter)
}

// RegisteredProperty represents a CSS @property custom property registration.
// See https://drafts.css-houdini.org/css-properties-values-api/#the-at-property-rule
type RegisteredProperty struct {
	Name         string
	Syntax       string // e.g., "<color>", "<number>", "<length>", "<percentage>", etc.
	Inherits     bool
	InitialValue string
}

// registeredProperties holds all @property custom property registrations.
var registeredProperties = make(map[string]RegisteredProperty)

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
				if strings.HasPrefix(atName, "keyframes") || strings.HasPrefix(atName, "-webkit-keyframes") {
					// Parse @keyframes block
					kf := parseKeyframes(sheet, i)
					if kf != nil {
						Keyframes = append(Keyframes, *kf)
					}
					i = skipBlock(sheet, j+1)
					continue
				}
			if strings.HasPrefix(atName, "layer") {
				// Parse @layer block
				layerName := extractLayerName(atName)
				layerContentStart := j + 1
				layerDepth := 1
				for layerDepth > 0 && layerContentStart < len(sheet) {
					if sheet[layerContentStart] == '{' {
						layerDepth++
					} else if sheet[layerContentStart] == '}' {
						layerDepth--
					}
					layerContentStart++
				}
				layerContent := sheet[j+1 : layerContentStart-1]
				// Recursively parse the rules inside @layer
				innerRules := parseLayerContent(layerContent, layerName)
				rules = append(rules, innerRules...)
				i = layerContentStart
				continue
			}
			if strings.HasPrefix(atName, "property") {
				// Parse @property custom property registration
				prop := parseProperty(sheet, i)
				if prop != nil {
					registeredProperties[prop.Name] = *prop
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
	// Find opening brace of keyframes block
	for j < len(sheet) && sheet[j] != '{' {
		j++
	}
	if j >= len(sheet) {
		return nil
	}

	// Extract keyframes name: content between "@keyframes" and "{"
	// e.g. "@keyframes fade {" → name = "fade"
	atName := strings.TrimSpace(sheet[start+1 : j])
	// Remove "@keyframes" or "@-webkit-keyframes" prefix
	name := strings.TrimPrefix(atName, "-webkit-")
	name = strings.TrimPrefix(name, "keyframes")
	name = strings.TrimSpace(name)
	if name == "" {
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

// parseProperty parses a @property custom property registration block.
// Format: @property --name { syntax: '<type>'; inherits: true|false; initial-value: '<value>'; }
func parseProperty(sheet string, start int) *RegisteredProperty {
	// Find the property name (starts with --)
	j := start + 1 // skip '@'
	for j < len(sheet) && (sheet[j] == ' ' || sheet[j] == '\t') {
		j++
	}
	if j+2 >= len(sheet) || sheet[j] != '-' || sheet[j+1] != '-' {
		return nil
	}
	j += 2 // skip '--'

	// Read the property name until whitespace or brace
	nameStart := j
	for j < len(sheet) && sheet[j] != ' ' && sheet[j] != '\t' && sheet[j] != '{' && sheet[j] != ';' {
		j++
	}
	propName := sheet[nameStart:j]
	propName = strings.TrimSpace(propName)
	if propName == "" || !strings.HasPrefix(propName, "--") {
		return nil
	}

	// Find the opening brace
	for j < len(sheet) && sheet[j] != '{' {
		j++
	}
	if j >= len(sheet) {
		return nil
	}

	// Parse the declarations inside the block
	decls, _ := parseDeclarations(sheet, j+1)

	prop := &RegisteredProperty{
		Name: propName,
	}

	// Extract syntax, inherits, and initial-value from declarations
	for _, d := range decls {
		propLower := strings.ToLower(d.Property)
		switch propLower {
		case "syntax":
			prop.Syntax = strings.Trim(d.Value, "'\"")
		case "inherits":
			prop.Inherits = strings.ToLower(d.Value) == "true"
		case "initial-value":
			prop.InitialValue = strings.Trim(d.Value, "'\"")
		}
	}

	// Validate required fields
	if prop.Syntax == "" || prop.InitialValue == "" {
		return nil
	}

	return prop
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

// extractLayerName extracts the layer name from an @layer directive.
// @layer           -> "" (anonymous layer with unique name)
// @layer name      -> "name"
// @layer name1,    -> "name1" (comma-separated, take first)
// @layer name.name -> "name.name" (nested layer)
func extractLayerName(atName string) string {
	// Remove @layer prefix
	name := strings.TrimPrefix(atName, "layer")
	name = strings.TrimSpace(name)

	if name == "" {
		// Anonymous @layer {} - generate unique name
		return nextAnonymousLayerName()
	}

	// Handle comma-separated layer names: @layer foo, bar, baz
	// Take the first one for simplicity
	if idx := strings.Index(name, ","); idx >= 0 {
		name = strings.TrimSpace(name[:idx])
	}

	return name
}

// parseLayerContent parses the content of a @layer block and assigns the layer name to all rules.
func parseLayerContent(content string, layerName string) []Rule {
	// Parse the content as regular CSS rules
	innerRules := Parse(content)
	// Assign the layer name to each rule
	for i := range innerRules {
		innerRules[i].Layer = layerName
	}
	return innerRules
}
