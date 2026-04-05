package fetch

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// IsValidURL validates a raw URL string without requiring it to be fully parsed.
// It checks for structural validity of common URL schemes.
func IsValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	// Check for suspicious characters that shouldn't be in URLs
	// Control characters, newlines, etc.
	suspiciousChars := [...]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	for _, c := range suspiciousChars {
		if strings.Contains(string(c), rawURL) {
			return false
		}
	}

	// Trim whitespace for validation purposes
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return false
	}

	// Protocol-relative URL (//example.com) - needs base URL to be valid
	if strings.HasPrefix(trimmed, "//") {
		return true // Valid structure, but requires context
	}

	// File URL
	if strings.HasPrefix(trimmed, "file://") {
		path := trimmed[7:]
		// file:// can be just "file://" or have a path
		if path == "" {
			return true
		}
		// Path should be absolute or relative and not contain suspicious chars
		return !containsSuspiciousPath(path)
	}

	// http/https URLs
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return false
		}
		// Must have a valid host
		if parsed.Host == "" {
			return false
		}
		// Host should not contain suspicious characters
		if containsSuspiciousHost(parsed.Host) {
			return false
		}
		return true
	}

	// No scheme - could be a domain name, check if it looks like one
	if !strings.Contains(trimmed, "://") {
		return looksLikeDomain(trimmed)
	}

	return false
}

// containsSuspiciousPath checks for control characters or other unsafe path chars
func containsSuspiciousPath(path string) bool {
	for _, c := range path {
		if c < 32 || c == 127 {
			return true
		}
	}
	return false
}

// containsSuspiciousHost checks for control characters in host
func containsSuspiciousHost(host string) bool {
	for _, c := range host {
		if c < 32 || c == 127 {
			return true
		}
	}
	return false
}

// looksLikeDomain checks if a string looks like a valid domain or hostname
func looksLikeDomain(s string) bool {
	if s == "" {
		return false
	}

	// Must not contain whitespace or control characters
	for _, c := range s {
		if c <= 32 || c == 127 {
			return false
		}
	}

	// Simple domain pattern check
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-.]*[a-zA-Z0-9])?$`)
	return domainRegex.MatchString(s) && len(s) <= 253
}

// SanitizeURL cleans and normalizes a URL string.
func SanitizeURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", &FetchError{URL: rawURL, Reason: "URL is empty"}
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "", &FetchError{URL: rawURL, Reason: "URL is empty after trimming whitespace"}
	}

	// Check for suspicious control characters
	for _, c := range trimmed {
		if c < 32 || c == 127 {
			return "", &FetchError{URL: rawURL, Reason: "URL contains control characters"}
		}
	}

	// Handle file:// URLs
	if strings.HasPrefix(trimmed, "file://") {
		path := trimmed[7:]
		if containsSuspiciousPath(path) {
			return "", &FetchError{URL: rawURL, Reason: "file URL contains invalid characters in path"}
		}
		return trimmed, nil
	}

	// Handle URLs with scheme
	if strings.Contains(trimmed, "://") {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return "", &FetchError{URL: rawURL, Reason: fmt.Sprintf("invalid URL structure: %v", err)}
		}

		// Validate scheme
		if parsed.Scheme != "http" && parsed.Scheme != "https" && parsed.Scheme != "file" {
			return "", &FetchError{URL: rawURL, Reason: fmt.Sprintf("unsupported scheme: %s", parsed.Scheme)}
		}

		// For http/https, validate host
		if (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host == "" {
			return "", &FetchError{URL: rawURL, Reason: "URL missing host"}
		}

		if containsSuspiciousHost(parsed.Host) {
			return "", &FetchError{URL: rawURL, Reason: "URL contains invalid characters in host"}
		}

		return trimmed, nil
	}

	// No scheme - check if it looks like a domain and add https://
	if looksLikeDomain(trimmed) {
		return "https://" + trimmed, nil
	}

	// Doesn't look like a valid domain
	return "", &FetchError{URL: rawURL, Reason: "invalid URL format, does not appear to be a valid domain"}
}

// FetchError represents an error that occurred during URL fetching.
type FetchError struct {
	URL    string
	Reason string
}

func (e *FetchError) Error() string {
	return fmt.Sprintf("fetch error: %s (URL: %s)", e.Reason, e.URL)
}

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
