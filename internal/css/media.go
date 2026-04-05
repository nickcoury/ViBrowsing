package css

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
