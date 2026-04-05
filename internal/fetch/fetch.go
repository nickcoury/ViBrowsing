package fetch

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// MaxDocumentSize is the maximum document size (10MB) that can be fetched.
// Documents larger than this will cause a FetchError with Reason "document too large".
const MaxDocumentSize = 10 * 1024 * 1024

// IsTextContent returns false for binary content types that shouldn't be parsed as HTML.
func IsTextContent(contentType string) bool {
	if contentType == "" {
		return false
	}
	// Strip charset parameter if present (e.g., "text/html; charset=utf-8")
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	contentType = strings.ToLower(contentType)

	// Explicitly non-text types
	nonTextPrefixes := []string{
		"image/",
		"audio/",
		"video/",
		"application/pdf",
		"application/octet-stream",
		"application/zip",
		"application/javascript",
		"application/x-javascript",
		"application/json",
		"application/xml",
		"application/msword",
		"application/vnd.",
		"font/",
	}

	for _, prefix := range nonTextPrefixes {
		if strings.HasPrefix(contentType, prefix) {
			return false
		}
	}

	// Text types that are safe to parse
	textTypes := []string{
		"text/html",
		"text/plain",
		"text/xml",
		"text/css",
		"text/xhtml",
		"application/xhtml+xml",
	}

	for _, t := range textTypes {
		if contentType == t {
			return true
		}
	}

	// Default: if it starts with "text/", treat as text
	if strings.HasPrefix(contentType, "text/") {
		return true
	}

	return false
}

// DetectBinaryContent checks the first few bytes of data for magic numbers
// to detect binary content when Content-Type header is missing or wrong.
func DetectBinaryContent(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// JPEG: FF D8 FF
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return true
	}

	// PNG: 89 50 4E 47
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return true
	}

	// GIF: 47 49 46 (GIF)
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return true
	}

	// PDF: 25 50 44 46 (%PDF)
	if data[0] == 0x25 && data[1] == 0x50 && data[2] == 0x44 && data[3] == 0x46 {
		return true
	}

	// WebP: 52 49 46 46 ... 57 45 42 50 (RIFF....WEBP)
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return true
	}

	// ZIP-based formats: 50 4B 03 04 (PK..)
	if data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x03 && data[3] == 0x04 {
		return true
	}

	// MP3/ audio: starts with ID3 (49 44 33) or MP3 frame sync (FF FB, FF F3, FF F4)
	if data[0] == 0x49 && data[1] == 0x44 && data[2] == 0x33 { // ID3
		return true
	}
	if data[0] == 0xFF && (data[1]&0xE0) == 0xE0 { // MP3 frame sync
		return true
	}

	// MP4/video: 00 00 00 XX 66 74 79 70 (ftyp) or 00 00 00 XX 6D 6F 6F 76 (moov)
	if len(data) >= 8 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 {
		ftyp := []byte{0x66, 0x74, 0x79, 0x70}
		moov := []byte{0x6D, 0x6F, 0x6F, 0x76}
		if data[4] == ftyp[0] && data[5] == ftyp[1] && data[6] == ftyp[2] && data[7] == ftyp[3] {
			return true
		}
		if data[4] == moov[0] && data[5] == moov[1] && data[6] == moov[2] && data[7] == moov[3] {
			return true
		}
	}

	return false
}

// Response holds the result of a fetch operation.
type Response struct {
	Body        []byte
	StatusCode  int
	ContentType string
	FinalURL    string
	Headers     http.Header // Response headers for cookie extraction
}

// Fetch retrieves a URL, following up to maxRedirects redirects.
// Supports http://, https://, and file:// URLs.
// userAgent overrides the default User-Agent. timeout is in seconds.
// cookieJar is optional - if provided, cookies will be sent and stored.
func Fetch(rawURL string, userAgent string, timeoutSecs int, cookieJar *CookieJar) (*Response, error) {
	// Validate URL before processing
	if !IsValidURL(rawURL) {
		return nil, &FetchError{URL: rawURL, Reason: "invalid URL format"}
	}

	// Sanitize and normalize the URL
	sanitized, err := SanitizeURL(rawURL)
	if err != nil {
		return nil, err
	}
	rawURL = sanitized

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
			return nil, &FetchError{URL: rawURL, Reason: fmt.Sprintf("file not found: %v", err)}
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
		return nil, &FetchError{URL: rawURL, Reason: fmt.Sprintf("failed to parse URL: %v", err)}
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

		// Add cookies if cookieJar is provided
		if cookieJar != nil {
			if cookieStr := cookieJar.GetCookies(currentURL); cookieStr != "" {
				req.Header.Set("Cookie", cookieStr)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed (timeout=%ds): %w", timeoutSecs, err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}

		// Check document size limit
		if len(body) > MaxDocumentSize {
			return nil, &FetchError{URL: rawURL, Reason: "document too large"}
		}

		contentType := resp.Header.Get("Content-Type")

		// Check for binary content
		isText := IsTextContent(contentType)
		if !isText || (contentType == "" && DetectBinaryContent(body)) {
			// Binary content detected - return early with appropriate handling
			if strings.HasPrefix(contentType, "image/") || (contentType == "" && DetectBinaryContent(body)) {
				// For images (or unknown but binary), return placeholder response
				binaryType := contentType
				if binaryType == "" {
					// Detect specific binary type from magic numbers
					if len(body) >= 3 && body[0] == 0xFF && body[1] == 0xD8 && body[2] == 0xFF {
						binaryType = "image/jpeg"
					} else if len(body) >= 4 && body[0] == 0x89 && body[1] == 0x50 && body[2] == 0x4E && body[3] == 0x47 {
						binaryType = "image/png"
					} else if len(body) >= 3 && body[0] == 0x47 && body[1] == 0x49 && body[2] == 0x46 {
						binaryType = "image/gif"
					} else if len(body) >= 4 && body[0] == 0x25 && body[1] == 0x50 && body[2] == 0x44 && body[3] == 0x46 {
						binaryType = "application/pdf"
					} else if len(body) >= 12 && body[0] == 0x52 && body[1] == 0x49 && body[2] == 0x46 && body[3] == 0x46 &&
						body[8] == 0x57 && body[9] == 0x45 && body[10] == 0x42 && body[11] == 0x50 {
						binaryType = "image/webp"
					} else {
						binaryType = "application/octet-stream"
					}
				}
				return &Response{
					Body:        body,
					StatusCode:  resp.StatusCode,
					ContentType: binaryType,
					FinalURL:    currentURL,
					Headers:     resp.Header,
				}, nil
			}
			// For other binary types, return error
			return nil, &FetchError{
				URL:    rawURL,
				Reason: fmt.Sprintf("binary content type not renderable: %s", contentType),
			}
		}

		// Handle redirects
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			redirectURL := resp.Header.Get("Location")
			if redirectURL == "" {
				resp.Body.Close()
				return &Response{
					Body:        body,
					StatusCode:  resp.StatusCode,
					ContentType: contentType,
					FinalURL:    currentURL,
					Headers:     resp.Header,
				}, nil
			}

			// Store cookies before following redirect
			if cookieJar != nil {
				cookieJar.SetCookies(currentURL, resp.Header.Values("Set-Cookie"))
			}

			// Resolve relative redirects
			absURL, err := resp.Request.URL.Parse(redirectURL)
			if err != nil {
				resp.Body.Close()
				return nil, fmt.Errorf("invalid redirect URL: %w", err)
			}
			currentURL = absURL.String()
			lastErr = fmt.Errorf("redirect loop exceeded")
			resp.Body.Close()
			continue
		}

		// Store cookies from final response
		if cookieJar != nil {
			cookieJar.SetCookies(currentURL, resp.Header.Values("Set-Cookie"))
		}

		resp.Body.Close()
		return &Response{
			Body:        body,
			StatusCode:  resp.StatusCode,
			ContentType: contentType,
			FinalURL:    currentURL,
			Headers:     resp.Header,
		}, nil
	}

	return nil, fmt.Errorf("too many redirects: %w", lastErr)
}

// FetchStreaming fetches a URL and calls the callback with each chunk of body data.
// It aborts if the total size exceeds maxSize bytes.
// This allows handling large documents without loading the entire body into memory.
func FetchStreaming(rawURL string, userAgent string, timeoutSecs int, cookieJar *CookieJar, maxSize int, callback func(chunk []byte) error) error {
	// Validate URL before processing
	if !IsValidURL(rawURL) {
		return &FetchError{URL: rawURL, Reason: "invalid URL format"}
	}

	// Sanitize and normalize the URL
	sanitized, err := SanitizeURL(rawURL)
	if err != nil {
		return err
	}
	rawURL = sanitized

	// Handle file:// URLs - not supported for streaming
	if strings.HasPrefix(rawURL, "file://") {
		return &FetchError{URL: rawURL, Reason: "streaming not supported for file:// URLs"}
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return &FetchError{URL: rawURL, Reason: fmt.Sprintf("failed to parse URL: %v", err)}
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
			return fmt.Errorf("failed to create request: %w", err)
		}

		ua := userAgent
		if ua == "" {
			ua = "HallucinHTML/1.0 (+https://github.com/nickcoury/ViBrowsing)"
		}
		req.Header.Set("User-Agent", ua)

		// Add cookies if cookieJar is provided
		if cookieJar != nil {
			if cookieStr := cookieJar.GetCookies(currentURL); cookieStr != "" {
				req.Header.Set("Cookie", cookieStr)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed (timeout=%ds): %w", timeoutSecs, err)
		}

		// Handle redirects
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			redirectURL := resp.Header.Get("Location")
			if redirectURL == "" {
				resp.Body.Close()
				return &FetchError{URL: rawURL, Reason: "redirect without location"}
			}

			// Store cookies before following redirect
			if cookieJar != nil {
				cookieJar.SetCookies(currentURL, resp.Header.Values("Set-Cookie"))
			}

			// Resolve relative redirects
			absURL, err := resp.Request.URL.Parse(redirectURL)
			if err != nil {
				resp.Body.Close()
				return fmt.Errorf("invalid redirect URL: %w", err)
			}
			currentURL = absURL.String()
			lastErr = fmt.Errorf("redirect loop exceeded")
			resp.Body.Close()
			continue
		}

		// Store cookies from final response
		if cookieJar != nil {
			cookieJar.SetCookies(currentURL, resp.Header.Values("Set-Cookie"))
		}

		// Read in chunks with size limit
		buf := make([]byte, 32*1024) // 32KB chunks
		totalSize := 0
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				chunk := buf[:n]
				totalSize += n
				if totalSize > maxSize {
					resp.Body.Close()
					return &FetchError{URL: rawURL, Reason: "document too large"}
				}
				if err := callback(chunk); err != nil {
					resp.Body.Close()
					return err
				}
			}
			if err != nil {
				if err == io.EOF {
					resp.Body.Close()
					return nil
				}
				resp.Body.Close()
				return fmt.Errorf("error reading body: %w", err)
			}
		}
	}

	return fmt.Errorf("too many redirects: %w", lastErr)
}

// FetchWithMaxSize fetches a URL and returns an error if the body exceeds maxSize bytes.
// This is a convenience wrapper around Fetch that enforces a custom size limit.
func FetchWithMaxSize(rawURL string, userAgent string, timeoutSecs int, cookieJar *CookieJar, maxSize int) (*Response, error) {
	// For small limits, use streaming to avoid loading large body into memory
	if maxSize < MaxDocumentSize {
		var body bytes.Buffer
		err := FetchStreaming(rawURL, userAgent, timeoutSecs, cookieJar, maxSize, func(chunk []byte) error {
			body.Write(chunk)
			return nil
		})
		if err != nil {
			return nil, err
		}
		resp, err := Fetch(rawURL, userAgent, timeoutSecs, cookieJar)
		if err != nil {
			return nil, err
		}
		resp.Body = body.Bytes()
		return resp, nil
	}
	return Fetch(rawURL, userAgent, timeoutSecs, cookieJar)
}
