package css

import "strings"

// MatchMediaQuery checks if a media query string matches the given viewport dimensions.
// Returns true if the media query matches, false otherwise.
// Supports: media types (screen, print, all), width/height (with min-/max- prefixes),
// orientation (portrait/landscape), aspect-ratio.
func MatchMediaQuery(mediaQuery string, viewportWidth int, viewportHeight int) bool {
	vw := float64(viewportWidth)
	vh := float64(viewportHeight)
	return evaluateMediaQuery(mediaQuery, vw, vh)
}

// FilterRulesByMedia filters a slice of CSS rules, returning only those whose
// @media condition matches the given viewport dimensions. Rules without a
// MediaQuery (i.e., MediaQuery == "") always match.
func FilterRulesByMedia(rules []Rule, viewportWidth int, viewportHeight int) []Rule {
	var filtered []Rule
	for _, rule := range rules {
		if rule.MediaQuery == "" {
			// No media query - always applies
			filtered = append(filtered, rule)
		} else if MatchMediaQuery(rule.MediaQuery, viewportWidth, viewportHeight) {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

// ContainerQuery represents a parsed container query condition.
type ContainerQuery struct {
	Name string // optional container name
	// Conditions for size queries
	MinWidth  float64
	MaxWidth  float64
	MinHeight float64
	MaxHeight float64
	Width     float64 // exact width (0 if not set)
	Height    float64 // exact height (0 if not set)
}

// ParseContainerQuery parses a container query condition string.
// Supports: (min-width: X), (max-width: X), (min-height: X), (max-height: X),
// (width: X), (height: X), and queries with optional container name.
func ParseContainerQuery(query string) *ContainerQuery {
	cq := &ContainerQuery{}

	// Parse optional container name: @container name (condition)
	// or just: @container (condition)
	query = strings.TrimSpace(query)

	// Check for container name before parenthesis
	if idx := strings.Index(query, "("); idx > 0 {
		name := strings.TrimSpace(query[:idx])
		// Check if this looks like a name and not just a condition
		if !strings.HasPrefix(name, "min-") && !strings.HasPrefix(name, "max-") &&
			!strings.HasPrefix(name, "width") && !strings.HasPrefix(name, "height") &&
			name != "not" && name != "and" {
			cq.Name = name
			query = strings.TrimSpace(query[idx:])
		}
	}

	// Remove surrounding parentheses
	query = strings.TrimSpace(query)
	if strings.HasPrefix(query, "(") && strings.HasSuffix(query, ")") {
		query = query[1 : len(query)-1]
	}

	// Handle "not" prefix
	if strings.HasPrefix(query, "not ") {
		// For simplicity, we'll parse the condition and negate later
		query = strings.TrimSpace(query[4:])
	}

	// Parse conditions
	// Split by "and" but keep it simple for now
	conditions := strings.Split(query, "and")
	for _, cond := range conditions {
		cond = strings.TrimSpace(cond)
		cq.parseCondition(cond)
	}

	return cq
}

// parseCondition parses a single container query condition.
func (cq *ContainerQuery) parseCondition(cond string) {
	cond = strings.TrimSpace(cond)

	// Remove surrounding parentheses if present
	if strings.HasPrefix(cond, "(") && strings.HasSuffix(cond, ")") {
		cond = cond[1 : len(cond)-1]
	}

	// min-width
	if strings.HasPrefix(cond, "min-width:") {
		val := strings.TrimSpace(cond[9:])
		if v := parseContainerLength(val); v > 0 {
			cq.MinWidth = v
		}
		return
	}
	// max-width
	if strings.HasPrefix(cond, "max-width:") {
		val := strings.TrimSpace(cond[10:])
		if v := parseContainerLength(val); v > 0 {
			cq.MaxWidth = v
		}
		return
	}
	// min-height
	if strings.HasPrefix(cond, "min-height:") {
		val := strings.TrimSpace(cond[11:])
		if v := parseContainerLength(val); v > 0 {
			cq.MinHeight = v
		}
		return
	}
	// max-height
	if strings.HasPrefix(cond, "max-height:") {
		val := strings.TrimSpace(cond[11:])
		if v := parseContainerLength(val); v > 0 {
			cq.MaxHeight = v
		}
		return
	}
	// width (exact)
	if strings.HasPrefix(cond, "width:") {
		val := strings.TrimSpace(cond[6:])
		if v := parseContainerLength(val); v > 0 {
			cq.Width = v
		}
		return
	}
	// height (exact)
	if strings.HasPrefix(cond, "height:") {
		val := strings.TrimSpace(cond[7:])
		if v := parseContainerLength(val); v > 0 {
			cq.Height = v
		}
		return
	}
}

// parseContainerLength parses a length value (e.g., "100px", "10rem", "50%").
func parseContainerLength(val string) float64 {
	val = strings.TrimSpace(val)
	val = strings.TrimSuffix(val, "px")
	val = strings.TrimSuffix(val, "rem")
	val = strings.TrimSuffix(val, "em")
	val = strings.TrimSuffix(val, "%")
	val = strings.TrimSpace(val)

	var result float64
	for _, c := range val {
		if c >= '0' && c <= '9' || c == '.' || c == '-' {
			continue
		}
		return 0
	}

	// Simple parse
	negative := false
	if strings.HasPrefix(val, "-") {
		negative = true
		val = val[1:]
	}
	for _, c := range val {
		if c == '.' {
			continue
		}
		if c < '0' || c > '9' {
			return 0
		}
		result = result*10 + float64(c-'0')
	}
	if negative {
		result = -result
	}
	return result
}

// MatchContainerQuery checks if a container query matches the given container dimensions.
func (cq *ContainerQuery) Match(containerWidth, containerHeight float64) bool {
	// Check min-width
	if cq.MinWidth > 0 && containerWidth < cq.MinWidth {
		return false
	}
	// Check max-width
	if cq.MaxWidth > 0 && containerWidth > cq.MaxWidth {
		return false
	}
	// Check min-height
	if cq.MinHeight > 0 && containerHeight < cq.MinHeight {
		return false
	}
	// Check max-height
	if cq.MaxHeight > 0 && containerHeight > cq.MaxHeight {
		return false
	}
	// Check exact width
	if cq.Width > 0 && containerWidth != cq.Width {
		return false
	}
	// Check exact height
	if cq.Height > 0 && containerHeight != cq.Height {
		return false
	}
	// If no conditions set, it matches (for default behavior)
	return true
}

// MatchContainerQueryString parses and matches a container query string against dimensions.
func MatchContainerQueryString(query string, containerWidth, containerHeight float64) bool {
	cq := ParseContainerQuery(query)
	return cq.Match(containerWidth, containerHeight)
}

// IsContainerQuery returns true if the string looks like a container query.
func IsContainerQuery(s string) bool {
	s = strings.TrimSpace(s)
	// Container queries start with @container or contain container-related conditions
	if strings.HasPrefix(s, "@container") {
		return true
	}
	// Check for container-size related keywords
	lower := strings.ToLower(s)
	if strings.Contains(lower, "min-width:") || strings.Contains(lower, "max-width:") ||
		strings.Contains(lower, "min-height:") || strings.Contains(lower, "max-height:") {
		// To distinguish from media queries, check for container-specific patterns
		// For simplicity, if it has container-type or container-name, it's definitely container
		return true
	}
	return false
}
