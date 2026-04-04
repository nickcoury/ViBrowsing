package fetch

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Response holds the result of a fetch operation.
type Response struct {
	Body       []byte
	StatusCode int
	ContentType string
	FinalURL   string
}

// Fetch retrieves a URL, following up to maxRedirects redirects.
// Supports http://, https://, and file:// URLs.
func Fetch(rawURL string, maxRedirects int) (*Response, error) {
	// Handle file:// URLs
	if strings.HasPrefix(rawURL, "file://") {
		path := rawURL[7:]
		if strings.HasPrefix(path, "/") {
			// Unix absolute path
		} else {
			// Relative path
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		return &Response{
			Body:        data,
			StatusCode:  200,
			ContentType: "text/html",
			FinalURL:   rawURL,
		}, nil
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	currentURL := parsedURL.String()
	var lastErr error

	for i := 0; i <= maxRedirects; i++ {
		req, err := http.NewRequest("GET", currentURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent", "HallucinHTML/1.0 (+https://github.com/nickcoury/ViBrowsing)")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}

		contentType := resp.Header.Get("Content-Type")

		// Handle redirects
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			redirectURL := resp.Header.Get("Location")
			if redirectURL == "" {
				return &Response{
					Body:        body,
					StatusCode:  resp.StatusCode,
					ContentType: contentType,
					FinalURL:    currentURL,
				}, nil
			}

			// Resolve relative redirects
			absURL, err := resp.Request.URL.Parse(redirectURL)
			if err != nil {
				return nil, fmt.Errorf("invalid redirect URL: %w", err)
			}
			currentURL = absURL.String()
			lastErr = fmt.Errorf("redirect loop exceeded")
			continue
		}

		return &Response{
			Body:        body,
			StatusCode:  resp.StatusCode,
			ContentType: contentType,
			FinalURL:    currentURL,
		}, nil
	}

	return nil, fmt.Errorf("too many redirects: %w", lastErr)
}
