package html

import (
	"regexp"
	"strings"
)

// DetectCharset tries to detect the character encoding from HTML content.
// It checks:
// 1. BOM (Byte Order Mark) at the start
// 2. <meta charset="..."> HTML5 style
// 3. <meta http-equiv="Content-Type" content="text/html; charset=..."> legacy style
// Returns "utf-8" as the default if nothing is found.
func DetectCharset(data []byte) string {
	// BOM check — UTF-8 BOM is EF BB BF
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return "utf-8"
	}

	s := string(data)

	// HTML5 style: <meta charset="UTF-8"> or <meta charset=utf-8>
	m := regexp.MustCompile(`(?i)<meta[^>]+charset=["']?([^"'\s>]+)`).FindStringSubmatch(s[:min(1024, len(s))])
	if len(m) >= 2 {
		return strings.ToLower(m[1])
	}

	// Legacy http-equiv style
	m = regexp.MustCompile(`(?i)<meta[^>]+http-equiv=["']?Content-Type["']?[^>]+content=["'][^"']*charset=([^"'\s;]+)`).FindStringSubmatch(s[:min(2048, len(s))])
	if len(m) >= 2 {
		return strings.ToLower(m[1])
	}

	// Default
	return "utf-8"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
