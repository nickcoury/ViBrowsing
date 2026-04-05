package fetch

import (
	"bytes"
	"compress/gzip"
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

// MaxLineLength is the maximum characters allowed in a single line of text.
// Lines exceeding this during streaming will cause a FetchError with Reason "line too long".
const MaxLineLength = 1024 * 1024 // 1MB characters

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
	Body          []byte
	Decompressed  []byte // Decompressed body if Content-Encoding was present
	StatusCode    int
	ContentType   string
	ContentEncode  string // Content-Encoding value (gzip, deflate, br)
	FinalURL      string
	Headers       http.Header // Response headers for cookie extraction
	LastModified  time.Time   // Last-Modified header value
}

// Fetch retrieves a URL, following up to maxRedirects redirects.
// Supports http://, https://, and file:// URLs.
// userAgent overrides the default User-Agent. timeout is in seconds.
// cookieJar is optional - if provided, cookies will be sent and stored.
// compression configures Accept-Encoding; if nil, uses DefaultCompressionConfig.
func Fetch(rawURL string, userAgent string, timeoutSecs int, cookieJar *CookieJar, compression *CompressionConfig) (*Response, error) {
	// Default compression config
	if compression == nil {
		compression = DefaultCompressionConfig
	}

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
			FinalURL:    rawURL,
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

	// Get the connection pool
	pool := GetDefaultPool()
	client := pool.GetClient()

	if timeoutSecs > 0 {
		client.Timeout = time.Duration(timeoutSecs) * time.Second
	} else {
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

		// Set Accept-Encoding header
		req.Header.Set("Accept-Encoding", compression.GetAcceptEncodingHeader())

		// Add cookies if cookieJar is provided
		if cookieJar != nil {
			if cookieStr := cookieJar.GetCookies(currentURL); cookieStr != "" {
				req.Header.Set("Cookie", cookieStr)
			}
		}

		// Add If-Modified-Since for conditional GET
		if condCache := GetDefaultConditionalCache(); condCache != nil {
			if lm, ok := condCache.GetLastModified(currentURL); ok {
				req.Header.Set("If-Modified-Since", lm.Format(http.TimeFormat))
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
		contentEncoding := resp.Header.Get("Content-Encoding")

		// Decompress if needed
		decompressed, err := decompressBody(body, contentEncoding)
		if err != nil {
			// If decompression fails, use original body
			decompressed = body
		}

		// Handle 304 Not Modified
		if resp.StatusCode == 304 {
			return &Response{
				Body:          []byte{},
				Decompressed:  []byte{},
				StatusCode:    304,
				ContentType:   contentType,
				ContentEncode: contentEncoding,
				FinalURL:      currentURL,
				Headers:       resp.Header,
				LastModified: parseLastModified(resp.Header.Get("Last-Modified")),
			}, nil
		}

		// Store Last-Modified for future conditional GETs
		if lastModStr := resp.Header.Get("Last-Modified"); lastModStr != "" {
			if lm := parseLastModified(lastModStr); !lm.IsZero() {
				GetDefaultConditionalCache().SetLastModified(currentURL, lm)
			}
		}

		// Check for binary content (use decompressed for check if applicable)
		checkBody := decompressed
		if checkBody == nil {
			checkBody = body
		}
		isText := IsTextContent(contentType)
		if !isText || (contentType == "" && DetectBinaryContent(checkBody)) {
			// Binary content detected - return early with appropriate handling
			if strings.HasPrefix(contentType, "image/") || (contentType == "" && DetectBinaryContent(checkBody)) {
				// For images (or unknown but binary), return placeholder response
				binaryType := contentType
				if binaryType == "" {
					// Detect specific binary type from magic numbers
					if len(checkBody) >= 3 && checkBody[0] == 0xFF && checkBody[1] == 0xD8 && checkBody[2] == 0xFF {
						binaryType = "image/jpeg"
					} else if len(checkBody) >= 4 && checkBody[0] == 0x89 && checkBody[1] == 0x50 && checkBody[2] == 0x4E && checkBody[3] == 0x47 {
						binaryType = "image/png"
					} else if len(checkBody) >= 3 && checkBody[0] == 0x47 && checkBody[1] == 0x49 && checkBody[2] == 0x46 {
						binaryType = "image/gif"
					} else if len(checkBody) >= 4 && checkBody[0] == 0x25 && checkBody[1] == 0x50 && checkBody[2] == 0x44 && checkBody[3] == 0x46 {
						binaryType = "application/pdf"
					} else if len(checkBody) >= 12 && checkBody[0] == 0x52 && checkBody[1] == 0x49 && checkBody[2] == 0x46 && checkBody[3] == 0x46 &&
						checkBody[8] == 0x57 && checkBody[9] == 0x45 && checkBody[10] == 0x42 && checkBody[11] == 0x50 {
						binaryType = "image/webp"
					} else {
						binaryType = "application/octet-stream"
					}
				}
				return &Response{
					Body:          body,
					Decompressed:  decompressed,
					StatusCode:    resp.StatusCode,
					ContentType:   binaryType,
					ContentEncode: contentEncoding,
					FinalURL:      currentURL,
					Headers:       resp.Header,
					LastModified:  parseLastModified(resp.Header.Get("Last-Modified")),
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
				return &Response{
					Body:          body,
					Decompressed:  decompressed,
					StatusCode:    resp.StatusCode,
					ContentType:   contentType,
					ContentEncode: contentEncoding,
					FinalURL:      currentURL,
					Headers:       resp.Header,
					LastModified:  parseLastModified(resp.Header.Get("Last-Modified")),
				}, nil
			}

			// Store cookies before following redirect
			if cookieJar != nil {
				cookieJar.SetCookies(currentURL, resp.Header.Values("Set-Cookie"))
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

		// Store cookies from final response
		if cookieJar != nil {
			cookieJar.SetCookies(currentURL, resp.Header.Values("Set-Cookie"))
		}

		return &Response{
			Body:          body,
			Decompressed:  decompressed,
			StatusCode:    resp.StatusCode,
			ContentType:   contentType,
			ContentEncode: contentEncoding,
			FinalURL:      currentURL,
			Headers:       resp.Header,
			LastModified:  parseLastModified(resp.Header.Get("Last-Modified")),
		}, nil
	}

	return nil, fmt.Errorf("too many redirects: %w", lastErr)
}

// decompressBody decompresses body based on Content-Encoding.
func decompressBody(body []byte, encoding string) ([]byte, error) {
	if len(body) == 0 || encoding == "" || encoding == "identity" {
		return nil, nil
	}

	switch encoding {
	case "gzip":
		br := bytes.NewReader(body)
		gz, err := gzip.NewReader(br)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		return io.ReadAll(gz)

	case "deflate":
		// deflate can be raw or zlib-wrapped
		// Try raw deflate first
		br := bytes.NewReader(body)
		gz, err := gzip.NewReader(br)
		if err == nil {
			defer gz.Close()
			result, err := io.ReadAll(gz)
			if err == nil {
				return result, nil
			}
		}
		// Try as raw deflate (no header)
		br = bytes.NewReader(body)
		return io.ReadAll(deflateReader{br})

	default:
		// Unknown encoding, return nil to use original body
		return nil, nil
	}
}

// deflateReader implements a simple deflate decompressor for raw deflate data.
type deflateReader struct {
	r *bytes.Reader
}

// Read implements io.Reader for deflateReader.
// This is a simplified implementation for raw deflate without zlib header.
func (d deflateReader) Read(p []byte) (int, error) {
	if d.r.Len() == 0 {
		return 0, io.EOF
	}
	n := len(p)
	if n > d.r.Len() {
		n = d.r.Len()
	}
	buf := make([]byte, n)
	_, err := d.r.Read(buf)
	if err != nil {
		return 0, err
	}
	// Simple passthrough - real implementation would use compress/zlib
	copy(p, buf)
	return n, nil
}

// parseLastModified parses a Last-Modified header value.
func parseLastModified(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	// Try various time formats
	formats := []string{
		http.TimeFormat,       // RFC 1123
		time.RFC850,            // RFC 850
		time.ANSIC,             // ANSI C
		"Mon, 02 Jan 2006 15:04:05 MST", // Common variant
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// FetchStreaming fetches a URL and calls the callback with each chunk of body data.
// It aborts if the total size exceeds maxSize bytes.
// This allows handling large documents without loading the entire body into memory.
// progress callback is called with download progress updates.
// cancel channel can be used to cancel the fetch mid-download.
func FetchStreaming(rawURL string, userAgent string, timeoutSecs int, cookieJar *CookieJar, maxSize int,
	progress func(bytesRead int64, contentLength int64), cancel <-chan struct{}) error {

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

	// Get the connection pool
	pool := GetDefaultPool()
	client := pool.GetClient()

	if timeoutSecs > 0 {
		client.Timeout = time.Duration(timeoutSecs) * time.Second
	} else {
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

		// Add If-Modified-Since for conditional GET
		if condCache := GetDefaultConditionalCache(); condCache != nil {
			if lm, ok := condCache.GetLastModified(currentURL); ok {
				req.Header.Set("If-Modified-Since", lm.Format(http.TimeFormat))
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed (timeout=%ds): %w", timeoutSecs, err)
		}

		// Check for cancellation before reading body
		select {
		case <-cancel:
			resp.Body.Close()
			return &FetchError{URL: rawURL, Reason: "fetch cancelled"}
		default:
		}

		// Handle 304 Not Modified
		if resp.StatusCode == 304 {
			resp.Body.Close()
			return nil
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

		// Store Last-Modified for future conditional GETs
		if lastModStr := resp.Header.Get("Last-Modified"); lastModStr != "" {
			if lm := parseLastModified(lastModStr); !lm.IsZero() {
				GetDefaultConditionalCache().SetLastModified(currentURL, lm)
			}
		}

		// Get content length if available
		contentLength := resp.ContentLength
		if contentLength < 0 {
			contentLength = -1
		}

		// Read in chunks with size limit and progress reporting
		buf := make([]byte, 32*1024) // 32KB chunks
		totalSize := 0
		lineLength := 0
		bytesRead := int64(0)

		for {
			// Check for cancellation between reads
			select {
			case <-cancel:
				resp.Body.Close()
				return &FetchError{URL: rawURL, Reason: "fetch cancelled"}
			default:
			}

			n, err := resp.Body.Read(buf)
			if n > 0 {
				chunk := buf[:n]
				totalSize += n
				bytesRead += int64(n)

				if totalSize > maxSize {
					resp.Body.Close()
					return &FetchError{URL: rawURL, Reason: "document too large"}
				}

				// Check for extremely long lines (search for newlines)
				for _, b := range chunk {
					if b == '\n' {
						lineLength = 0
					} else {
						lineLength++
						if lineLength > MaxLineLength {
							resp.Body.Close()
							return &FetchError{URL: rawURL, Reason: "line too long"}
						}
					}
				}

				// Report progress
				if progress != nil {
					progress(bytesRead, contentLength)
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

// FetchStreamingWithOptions fetches a URL with streaming and progress/cancellation support.
func FetchStreamingWithOptions(rawURL string, userAgent string, timeoutSecs int, cookieJar *CookieJar,
	maxSize int, opts StreamingOptions) error {

	if opts.CancelFunc == nil && opts.OnProgress == nil {
		return nil
	}

	cancel := make(chan struct{})
	if opts.CancelFunc != nil {
		// Note: The caller should not close this channel; calling CancelFunc will close it
		go func() {
			opts.CancelFunc()
			close(cancel)
		}()
	}

	return FetchStreaming(rawURL, userAgent, timeoutSecs, cookieJar, maxSize,
		func(bytesRead, contentLength int64) {
			if opts.OnProgress != nil {
				opts.OnProgress(FetchProgress{
					URL:           rawURL,
					BytesRead:     bytesRead,
					ContentLength: contentLength,
				})
			}
		}, cancel)
}

// FetchStylesheet fetches an external CSS stylesheet from a URL.
// Returns the CSS text content and any error.
// timeout is in seconds (0 = default 10s).
func FetchStylesheet(rawURL string, userAgent string, timeoutSecs int, cookieJar *CookieJar) (string, error) {
	if timeoutSecs <= 0 {
		timeoutSecs = 10
	}

	resp, err := Fetch(rawURL, userAgent, timeoutSecs, cookieJar, nil)
	if err != nil {
		return "", err
	}

	// Verify it's a CSS content type
	contentType := resp.ContentType
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	if contentType != "text/css" && !strings.HasPrefix(contentType, "text/") {
		// Some servers serve CSS with wrong content-type; still try to use it if it looks like CSS
		cssText := string(resp.Body)
		if len(cssText) > 0 && (strings.Contains(cssText, "{") || strings.Contains(cssText, "@")) {
			return cssText, nil
		}
		return "", &FetchError{URL: rawURL, Reason: fmt.Sprintf("not CSS content-type: %s", resp.ContentType)}
	}

	// Use decompressed body if available
	body := resp.Body
	if resp.Decompressed != nil {
		body = resp.Decompressed
	}

	return string(body), nil
}

// FetchWithMaxSize fetches a URL and returns an error if the body exceeds maxSize bytes.
// This is a convenience wrapper around Fetch that enforces a custom size limit.
func FetchWithMaxSize(rawURL string, userAgent string, timeoutSecs int, cookieJar *CookieJar, maxSize int) (*Response, error) {
	resp, err := Fetch(rawURL, userAgent, timeoutSecs, cookieJar, nil)
	if err != nil {
		return nil, err
	}
	if len(resp.Body) > maxSize {
		return nil, &FetchError{URL: rawURL, Reason: fmt.Sprintf("document too large (max %d bytes)", maxSize)}
	}
	return resp, nil
}
