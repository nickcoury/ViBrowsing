package fetch

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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
// userAgent overrides the default User-Agent. timeout is in seconds.
func Fetch(rawURL string, userAgent string, timeoutSecs int) (*Response, error) {
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

	client := &http.Client{
		Timeout: time.Duration(timeoutSecs) * time.Second,
	}
	if timeoutSecs <= 0 {
		client.Timeout = 30 * time.Second
	}

	for i := 0; i <= 10; i++ {
		req, err := http.NewRequest("GET", currentURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		ua := userAgent
		if ua == "" {
			ua = "HallucinHTML/1.0 (+https://github.com/nickcoury/ViBrowsing)"
		}
		req.Header.Set("User-Agent", ua)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed (timeout=%ds): %w", timeoutSecs, err)
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
