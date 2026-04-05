package css

import (
	"strings"
)

// Supports checks if a CSS property/value pair is supported by the browser.
// This implements the CSS.supports() API.
// Usage: CSS.supports("display", "grid") or CSS.supports("(max-width: 768px)")
func Supports(property, value string) bool {
	property = strings.TrimSpace(property)
	value = strings.TrimSpace(value)

	// Handle condition syntax: CSS.supports("(max-width: 768px)")
	if strings.HasPrefix(property, "(") {
		return supportsCondition(property)
	}

	// Handle property: value syntax
	return supportsPropertyValue(property, value)
}

// supportsCondition checks if a media query condition is supported.
func supportsCondition(condition string) bool {
	condition = strings.TrimSpace(condition)
	if strings.HasPrefix(condition, "(max-width:") {
		return true // we support max-width media queries
	}
	if strings.HasPrefix(condition, "(min-width:") {
		return true // we support min-width media queries
	}
	if strings.HasPrefix(condition, "(width:") {
		return true // we support width media queries
	}
	if strings.HasPrefix(condition, "(orientation:") {
		return true // we support orientation
	}
	if strings.HasPrefix(condition, "(aspect-ratio:") {
		return true // we support aspect-ratio
	}
	if strings.HasPrefix(condition, "(color:") {
		return true // we support color media query
	}
	if strings.HasPrefix(condition, "(hover:") {
		return true // we support hover media query
	}
	if strings.HasPrefix(condition, "(pointer:") {
		return true // we support pointer media query
	}
	// Default: assume supported for forward compatibility
	return true
}

// supportsPropertyValue checks if a property/value combination is supported.
func supportsPropertyValue(property, value string) bool {
	property = strings.ToLower(property)
	value = strings.ToLower(value)

	// Check for known supported properties
	switch property {
	case "display":
		return supportsDisplay(value)
	case "position":
		return supportsPosition(value)
	case "float":
		return supportsFloat(value)
	case "overflow":
		return supportsOverflow(value)
	case "visibility":
		return value == "visible" || value == "hidden" || value == "collapse"
	case "opacity":
		return true // any number 0-1 is valid
	case "color":
		return ParseColor(value).A > 0 || value == "transparent" || value == "inherit" || value == "initial" || value == "currentcolor"
	case "background", "background-color":
		return true
	case "font-size":
		return true
	case "font-weight":
		return supportsFontWeight(value)
	case "font-style":
		return value == "normal" || value == "italic" || value == "oblique"
	case "text-decoration":
		return true
	case "border", "border-width", "border-style", "border-color":
		return true
	case "margin", "padding":
		return true
	case "width", "height":
		return true
	case "max-width", "min-width", "max-height", "min-height":
		return true
	case "flex", "flex-direction", "flex-wrap", "flex-grow", "flex-shrink", "flex-basis":
		return true
	case "align-items", "align-content", "justify-content":
		return true
	case "gap":
		return true
	case "grid", "grid-template-columns", "grid-template-rows":
		return true
	case "top", "left", "right", "bottom":
		return true
	case "z-index":
		return true
	case "transform":
		return supportsTransform(value)
	case "filter":
		return supportsFilter(value)
	case "clip-path":
		return supportsClipPath(value)
	case "backdrop-filter":
		return supportsFilter(value) // same syntax as filter
	case "border-radius":
		return true
	case "box-shadow":
		return true
	case "text-shadow":
		return true
	case "background-image":
		return supportsBackgroundImage(value)
	case "background-repeat":
		return supportsBackgroundRepeat(value)
	case "background-position":
		return true
	case "background-size":
		return true
	case "transition":
		return true
	case "animation":
		return true
	case "white-space":
		return supportsWhiteSpace(value)
	case "text-align":
		return true
	case "vertical-align":
		return true
	case "line-height":
		return true
	case "letter-spacing", "word-spacing":
		return true
	case "text-indent":
		return true
	case "text-transform":
		return true
	case "content":
		return true
	case "quotes":
		return true
	case "list-style-type", "list-style-position", "list-style-image":
		return true
	case "outline":
		return true
	case "cursor":
		return true
	case "resize":
		return true
	case "overflow-x", "overflow-y":
		return supportsOverflow(value)
	case "object-fit":
		return value == "fill" || value == "contain" || value == "cover" || value == "none" || value == "scale-down"
	case "aspect-ratio":
		return true
	case "font-family":
		return true
	case "font-stretch":
		return supportsFontStretch(value)
	case "font-variant":
		return true
	case "direction":
		return value == "ltr" || value == "rtl"
	case "writing-mode":
		return value == "horizontal-tb" || value == "vertical-rl" || value == "vertical-lr"
	case "unicode-bidi":
		return true
	case "tab-size":
		return true
	case "hyphens":
		return true
	case "break-inside", "break-before", "break-after":
		return true
	case "column-width", "column-count", "column-gap", "column-rule":
		return true
	case "contain":
		return true
	case "mix-blend-mode":
		return supportsBlendMode(value)
	case "hanging-punctuation":
		return true
	case "transform-box":
		return value == "view-box" || value == "fill-box" || value == "content-box" || value == "border-box"
	case "place-items", "place-self", "place-content":
		return true
	case "user-select":
		return true
	case "pointer-events":
		return true
	case "overscroll-behavior":
		return true
	case "scroll-behavior":
		return true
	case "caret-color":
		return true
	case "appearance":
		return true
	case "will-change":
		return true
	case "image-rendering":
		return true
	case "caption-side":
		return value == "top" || value == "bottom"
	case "empty-cells":
		return value == "show" || value == "hide"
	default:
		// Unknown property - for forward compatibility, assume supported
		return true
	}
}

func supportsDisplay(value string) bool {
	switch value {
	case "block", "inline", "none", "flex", "inline-block", "inline-flex",
		"grid", "inline-grid", "list-item", "run-in", "contents",
		"table", "inline-table", "table-row", "table-cell",
		"table-row-group", "table-header-group", "table-footer-group",
		"table-caption":
		return true
	}
	return false
}

func supportsPosition(value string) bool {
	return value == "static" || value == "relative" || value == "absolute" || value == "fixed" || value == "sticky"
}

func supportsFloat(value string) bool {
	return value == "none" || value == "left" || value == "right"
}

func supportsOverflow(value string) bool {
	return value == "visible" || value == "hidden" || value == "scroll" || value == "auto"
}

func supportsFontWeight(value string) bool {
	switch value {
	case "normal", "bold", "bolder", "lighter":
		return true
	}
	// Numeric weights
	if v, err := ParseFloat(value, 64); err == nil {
		return v >= 1 && v <= 1000
	}
	return false
}

func supportsTransform(value string) bool {
	if value == "none" {
		return true
	}
	// Check for transform functions
	transforms := []string{"rotate", "scale", "translate", "skew", "matrix", "perspective"}
	for _, t := range transforms {
		if strings.Contains(value, t+"(") {
			return true
		}
	}
	return false
}

func supportsFilter(value string) bool {
	if value == "none" {
		return true
	}
	// Check for filter functions
	filters := []string{"blur", "brightness", "contrast", "grayscale", "sepia", "hue-rotate", "drop-shadow", "opacity"}
	for _, f := range filters {
		if strings.Contains(value, f+"(") {
			return true
		}
	}
	return false
}

func supportsClipPath(value string) bool {
	if value == "none" {
		return true
	}
	// Check for clip-path functions
	clipFuncs := []string{"inset", "circle", "ellipse", "polygon", "url"}
	for _, c := range clipFuncs {
		if strings.Contains(value, c+"(") {
			return true
		}
	}
	return false
}

func supportsBackgroundImage(value string) bool {
	if value == "none" || value == "inherit" || value == "initial" || value == "unset" {
		return true
	}
	if strings.HasPrefix(value, "url(") {
		return true
	}
	if strings.HasPrefix(value, "linear-gradient(") {
		return true
	}
	if strings.HasPrefix(value, "radial-gradient(") {
		return true
	}
	if strings.HasPrefix(value, "conic-gradient(") {
		return true
	}
	return false
}

func supportsBackgroundRepeat(value string) bool {
	switch value {
	case "repeat", "no-repeat", "space", "round":
		return true
	}
	if strings.Contains(value, "repeat-x") || strings.Contains(value, "repeat-y") {
		return true
	}
	return false
}

func supportsWhiteSpace(value string) bool {
	switch value {
	case "normal", "pre", "pre-wrap", "pre-line", "nowrap", "break-spaces":
		return true
	}
	return false
}

func supportsFontStretch(value string) bool {
	stretches := []string{"ultra-condensed", "extra-condensed", "condensed", "semi-condensed",
		"normal", "semi-expanded", "expanded", "extra-expanded", "ultra-expanded"}
	for _, s := range stretches {
		if value == s {
			return true
		}
	}
	// Percentage values
	if strings.HasSuffix(value, "%") {
		if v, err := ParseFloat(strings.TrimSuffix(value, "%"), 64); err == nil {
			return v >= 50 && v <= 200
		}
	}
	return false
}

func supportsBlendMode(value string) bool {
	modes := []string{"normal", "multiply", "screen", "overlay", "darken", "lighten",
		"color-dodge", "color-burn", "hard-light", "soft-light", "difference",
		"exclusion", "hue", "saturation", "color", "luminosity"}
	for _, m := range modes {
		if value == m {
			return true
		}
	}
	return false
}

// ParseFloat is a helper to parse a float from a string.
func ParseFloat(s string, bits int) (float64, error) {
	s = strings.TrimSpace(s)
	var sign float64 = 1
	if strings.HasPrefix(s, "-") {
		sign = -1
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	// Simple float parsing without strconv
	var result float64
	var decimal float64
	var afterDecimal bool
	for _, c := range s {
		if c >= '0' && c <= '9' {
			if afterDecimal {
				decimal = decimal*10 + float64(c-'0')
			} else {
				result = result*10 + float64(c-'0')
			}
		} else if c == '.' && !afterDecimal {
			afterDecimal = true
		} else {
			break
		}
	}
	// Apply decimal part
	if afterDecimal {
		// Count digits after decimal
		decStr := ""
		for _, c := range s[strings.Index(s, ".")+1:] {
			if c >= '0' && c <= '9' {
				decStr += string(c)
			} else {
				break
			}
		}
		decLen := len(decStr)
		for i := 0; i < decLen; i++ {
			decimal /= 10
		}
		result += decimal
	}
	return sign * result, nil
}

// MediaQueryList represents a media query list for window.matchMedia()
type MediaQueryList struct {
	MediaQuery string
	Matches    bool
}

// MatchMedia checks if a media query matches the given viewport dimensions.
func MatchMedia(mediaQuery string, viewportWidth, viewportHeight float64) *MediaQueryList {
	matches := evaluateMediaQuery(mediaQuery, viewportWidth, viewportHeight)
	return &MediaQueryList{
		MediaQuery: mediaQuery,
		Matches:    matches,
	}
}

// evaluateMediaQuery evaluates a media query string against viewport dimensions.
func evaluateMediaQuery(query string, vw, vh float64) bool {
	query = strings.TrimSpace(query)

	// Handle "not" at the beginning
	not := false
	if strings.HasPrefix(query, "not") {
		not = true
		query = strings.TrimSpace(query[3:])
	}

	// Handle "only" (equivalent to not for our purposes)
	if strings.HasPrefix(query, "only") {
		query = strings.TrimSpace(query[4:])
	}

	// Handle media types: screen, print, all, etc.
	// For "type and (cond)", idx points to " and", skip just " and" (3 chars) to keep type
	// For "(cond1) and (cond2)", we need to keep both conditions
	if idx := strings.Index(query, " and"); idx >= 0 {
		before := strings.TrimSpace(query[:idx])
		after := strings.TrimSpace(query[idx+4:])
		// If before looks like a condition (starts with '('), keep both
		if strings.HasPrefix(before, "(") {
			query = before + " " + after
		} else {
			query = after
		}
	} else {
		// No "and" found - check if it's just a media type
		lower := strings.ToLower(query)
		if lower == "screen" || lower == "print" || lower == "all" || lower == "speech" || lower == "tty" || lower == "tv" || lower == "projection" || lower == "handheld" || lower == "braille" || lower == "embossed" || lower == "aural" {
			return true
		}
	}

	matches := true

	// Parse conditions
	// Extract conditions like (max-width: 768px), (min-width: 320px), etc.
	for {
		start := strings.Index(query, "(")
		if start < 0 {
			break
		}
		end := strings.Index(query[start:], ")")
		if end < 0 {
			break
		}
		end += start

		condition := query[start+1 : end]
		if !evaluateCondition(condition, vw, vh) {
			matches = false
		}

		query = query[end+1:]
		query = strings.TrimSpace(query)
		if strings.HasPrefix(query, "and") {
			query = strings.TrimSpace(query[3:])
		}
	}

	if not {
		matches = !matches
	}

	return matches
}

func evaluateCondition(cond string, vw, vh float64) bool {
	cond = strings.TrimSpace(cond)

	// max-width
	if strings.HasPrefix(cond, "max-width:") {
		val := strings.TrimSpace(cond[10:])
		val = strings.TrimSuffix(val, ";")
		val = strings.TrimSpace(val)
		if w, err := ParseFloat(val, 64); err == nil {
			return vw <= w
		}
	}

	// min-width
	if strings.HasPrefix(cond, "min-width:") {
		val := strings.TrimSpace(cond[10:])
		val = strings.TrimSuffix(val, ";")
		val = strings.TrimSpace(val)
		if w, err := ParseFloat(val, 64); err == nil {
			return vw >= w
		}
	}

	// max-height
	if strings.HasPrefix(cond, "max-height:") {
		val := strings.TrimSpace(cond[11:])
		val = strings.TrimSuffix(val, ";")
		val = strings.TrimSpace(val)
		if h, err := ParseFloat(val, 64); err == nil {
			return vh <= h
		}
	}

	// min-height
	if strings.HasPrefix(cond, "min-height:") {
		val := strings.TrimSpace(cond[11:])
		val = strings.TrimSuffix(val, ";")
		val = strings.TrimSpace(val)
		if h, err := ParseFloat(val, 64); err == nil {
			return vh >= h
		}
	}

	// width
	if strings.HasPrefix(cond, "width:") {
		val := strings.TrimSpace(cond[6:])
		val = strings.TrimSuffix(val, ";")
		val = strings.TrimSpace(val)
		if w, err := ParseFloat(val, 64); err == nil {
			return vw == w
		}
	}

	// height
	if strings.HasPrefix(cond, "height:") {
		val := strings.TrimSpace(cond[7:])
		val = strings.TrimSuffix(val, ";")
		val = strings.TrimSpace(val)
		if h, err := ParseFloat(val, 64); err == nil {
			return vh == h
		}
	}

	// orientation
	if strings.HasPrefix(cond, "orientation:") {
		val := strings.TrimSpace(cond[12:])
		val = strings.TrimSuffix(val, ";")
		val = strings.TrimSpace(val)
		if val == "portrait" {
			return vh > vw
		}
		if val == "landscape" {
			return vw > vh
		}
	}

	// aspect-ratio
	if strings.HasPrefix(cond, "aspect-ratio:") {
		val := strings.TrimSpace(cond[12:])
		val = strings.TrimSuffix(val, ";")
		val = strings.TrimSpace(val)
		if vw > 0 && vh > 0 {
			ratio := vw / vh
			// Parse desired ratio like "16/9"
			if strings.Contains(val, "/") {
				parts := strings.Split(val, "/")
				if len(parts) == 2 {
					if w, err := ParseFloat(parts[0], 64); err == nil {
						if h, err := ParseFloat(parts[1], 64); err == nil && h > 0 {
							return abs(ratio-w/h) < 0.01
						}
					}
				}
			}
		}
	}

	// color (bits per color component)
	if strings.HasPrefix(cond, "color:") {
		val := strings.TrimSpace(cond[5:])
		val = strings.TrimSuffix(val, ";")
		val = strings.TrimSpace(val)
		if val == "0" {
			return true // monochrome display
		}
		if _, err := ParseFloat(val, 64); err == nil {
			return true // we support color
		}
	}

	// hover
	if strings.HasPrefix(cond, "hover:") {
		return true // we assume hover is supported
	}

	// pointer
	if strings.HasPrefix(cond, "pointer:") {
		val := strings.TrimSpace(cond[8:])
		if val == "fine" || val == "coarse" || val == "none" {
			return true
		}
	}

	return true // Default: assume supported
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
