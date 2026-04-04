package fetch

import (
	"net/url"
	"strings"
)

// ResolveURL resolves a potentially relative URL against a base URL.
// If baseURL is empty, the rawURL is returned as-is.
// If rawURL is already absolute (has scheme), it's returned as-is.
func ResolveURL(rawURL, baseURL string) string {
	if rawURL == "" {
		return rawURL
	}

	// Protocol-relative URL (//example.com/path)
	if strings.HasPrefix(rawURL, "//") {
		if baseURL == "" {
			return "https:" + rawURL
		}
		// Use scheme from base URL
		if baseParsed, err := url.Parse(baseURL); err == nil && baseParsed.Scheme != "" {
			return baseParsed.Scheme + ":" + rawURL
		}
		return "https:" + rawURL
	}

	// If rawURL is already absolute (has scheme), return it
	parsed, _ := url.Parse(rawURL)
	if parsed != nil && parsed.Scheme != "" {
		return rawURL
	}

	// If no base URL, just return rawURL as-is
	if baseURL == "" {
		return rawURL
	}

	// Parse the base URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return rawURL
	}

	// Resolve the relative URL against the base
	resolved := base.ResolveReference(parsed).String()
	return resolved
}
