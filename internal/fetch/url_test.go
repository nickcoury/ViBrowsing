package fetch

import "testing"

func TestResolveURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		baseURL  string
		expected string
	}{
		{
			name:     "absolute URL unchanged",
			rawURL:   "https://example.com/page",
			baseURL:  "https://other.com",
			expected: "https://example.com/page",
		},
		{
			name:     "relative path resolved against base",
			rawURL:   "/path/page",
			baseURL:  "https://example.com/dir/",
			expected: "https://example.com/path/page",
		},
		{
			name:     "relative path with parent",
			rawURL:   "../other/page",
			baseURL:  "https://example.com/dir/subdir/",
			expected: "https://example.com/dir/other/page",
		},
		{
			name:     "protocol-relative URL with base",
			rawURL:   "//cdn.example.com/image.png",
			baseURL:  "https://example.com/page",
			expected: "https://cdn.example.com/image.png",
		},
		{
			name:     "protocol-relative URL without base",
			rawURL:   "//cdn.example.com/image.png",
			baseURL:  "",
			expected: "https://cdn.example.com/image.png",
		},
		{
			name:     "empty raw URL",
			rawURL:   "",
			baseURL:  "https://example.com/page",
			expected: "",
		},
		{
			name:     "query string only",
			rawURL:   "?search=term",
			baseURL:  "https://example.com/page",
			expected: "https://example.com/page?search=term",
		},
		{
			name:     "fragment only",
			rawURL:   "#section",
			baseURL:  "https://example.com/page",
			expected: "https://example.com/page#section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveURL(tt.rawURL, tt.baseURL)
			if result != tt.expected {
				t.Errorf("ResolveURL(%q, %q) = %q, want %q", tt.rawURL, tt.baseURL, result, tt.expected)
			}
		})
	}
}
