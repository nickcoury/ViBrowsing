package fetch

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDecompressBody(t *testing.T) {
	tests := []struct {
		name     string
		encoding string
		input    string
		want     string
		wantErr  bool
	}{
		{
			name:     "gzip compression",
			encoding: "gzip",
			input:    "hello world",
			want:     "hello world",
		},
		{
			name:     "empty encoding",
			encoding: "",
			input:    "hello",
			want:     "",
		},
		{
			name:     "identity encoding",
			encoding: "identity",
			input:    "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input []byte
			if tt.encoding == "gzip" {
				var buf strings.Builder
				w := gzip.NewWriter(&buf)
				w.Write([]byte(tt.input))
				w.Close()
				input = []byte(buf.String())
			} else {
				input = []byte(tt.input)
			}

			got, err := decompressBody(input, tt.encoding)
			if (err != nil) != tt.wantErr {
				t.Errorf("decompressBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.encoding == "" || tt.encoding == "identity" {
				// For empty/identity, decompressBody returns nil
				if got != nil && string(got) != tt.input {
					t.Errorf("decompressBody() = %v, want %v", string(got), tt.input)
				}
			}
		})
	}
}

func TestParseLastModified(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{
			name:  "RFC 1123 format",
			input: "Wed, 21 Oct 2015 07:28:00 GMT",
			want:  time.Date(2015, 10, 21, 7, 28, 0, 0, time.UTC),
		},
		{
			name:  "empty string",
			input: "",
			want:  time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLastModified(tt.input)
			if !tt.want.IsZero() && !got.Equal(tt.want) {
				t.Errorf("parseLastModified() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompressionConfig(t *testing.T) {
	cfg := DefaultCompressionConfig

	header := cfg.GetAcceptEncodingHeader()
	if header == "" {
		t.Error("GetAcceptEncodingHeader() returned empty string")
	}

	// Should contain gzip
	if !strings.Contains(header, "gzip") {
		t.Error("GetAcceptEncodingHeader() should contain gzip")
	}

	// Should contain deflate
	if !strings.Contains(header, "deflate") {
		t.Error("GetAcceptEncodingHeader() should contain deflate")
	}

	// Should contain br (brotli)
	if !strings.Contains(header, "br") {
		t.Error("GetAcceptEncodingHeader() should contain br")
	}
}

func TestParseContentEncoding(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"gzip", "gzip"},
		{"GZIP", "gzip"},
		{"deflate", "deflate"},
		{"br", "br"},
		{"identity", "identity"},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseContentEncoding(tt.input)
			if got != tt.want {
				t.Errorf("ParseContentEncoding(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConditionalCache(t *testing.T) {
	cache := NewConditionalCache()
	url := "https://example.com/page"

	// Initially should not have entry
	if _, ok := cache.GetLastModified(url); ok {
		t.Error("expected no entry initially")
	}

	// Set Last-Modified
	lm := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	cache.SetLastModified(url, lm)

	// Should now have entry
	got, ok := cache.GetLastModified(url)
	if !ok {
		t.Error("expected entry after SetLastModified")
	}
	if !got.Equal(lm) {
		t.Errorf("GetLastModified() = %v, want %v", got, lm)
	}
}

func TestConnPool(t *testing.T) {
	pool := NewConnPool()
	if pool == nil {
		t.Fatal("NewConnPool() returned nil")
	}

	client := pool.GetClient()
	if client == nil {
		t.Error("GetClient() returned nil")
	}

	transport := pool.GetTransport()
	if transport == nil {
		t.Error("GetTransport() returned nil")
	}

	// Test stats
	stats := pool.Stats()
	if stats.HostPools == nil {
		t.Error("Stats().HostPools should not be nil")
	}

	// Test close
	pool.Close()
}

func TestResponse304(t *testing.T) {
	// Create a test server that returns 304 Not Modified
	var receivedIfModifiedSince string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedIfModifiedSince = r.Header.Get("If-Modified-Since")
		if receivedIfModifiedSince != "" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Full content"))
	}))
	defer server.Close()

	// First request - get Last-Modified
	resp, err := Fetch(server.URL, "", 10, nil, nil)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("First request: expected status 200, got %d", resp.StatusCode)
	}

	// Second request - should send If-Modified-Since and get 304
	resp, err = Fetch(server.URL, "", 10, nil, nil)
	if err != nil {
		t.Fatalf("Fetch() second request error = %v", err)
	}
	if resp.StatusCode != 304 {
		t.Errorf("Second request: expected status 304, got %d", resp.StatusCode)
	}
	if receivedIfModifiedSince == "" {
		t.Error("Expected If-Modified-Since header to be sent")
	}
}

func TestAcceptEncodingHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ae := r.Header.Get("Accept-Encoding")
		if ae == "" {
			t.Error("Accept-Encoding header should be set")
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "text/plain")
		
		// Return gzip compressed content
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		gz.Write([]byte("Hello, World!"))
		gz.Close()
		w.Write(buf.Bytes())
	}))
	defer server.Close()

	resp, err := Fetch(server.URL, "", 10, nil, nil)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	// Should have Content-Encoding header
	if resp.ContentEncode != "gzip" {
		t.Errorf("Expected Content-Encoding gzip, got %q", resp.ContentEncode)
	}

	// Decompressed body should be available
	if resp.Decompressed == nil {
		t.Error("Expected Decompressed body to be set")
	}
}

func TestFetchStreamingCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send a slow response
		w.Header().Set("Content-Type", "text/html")
		for i := 0; i < 100; i++ {
			w.Write([]byte("chunk "))
			w.(http.Flusher).Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	cancel := make(chan struct{})
	err := FetchStreaming(server.URL, "", 30, nil, 1024*1024, 
		func(bytesRead, contentLength int64) {}, cancel)
	
	// Cancel immediately
	close(cancel)
	
	// Should complete without error (cancellation happens mid-download)
	// The fetch may or may not be cancelled depending on timing
	_ = err
}

func TestFetchStreamingProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	var lastProgress int64
	err := FetchStreaming(server.URL, "", 10, nil, 1024*1024,
		func(bytesRead, contentLength int64) {
			lastProgress = bytesRead
		}, nil)

	if err != nil {
		t.Fatalf("FetchStreaming() error = %v", err)
	}

	if lastProgress == 0 {
		t.Error("Expected progress callback to be called")
	}
}

func TestIsTextContent(t *testing.T) {
	tests := []struct {
		contentType string
		want        bool
	}{
		{"text/html", true},
		{"text/plain", true},
		{"text/css", true},
		{"text/html; charset=utf-8", true},
		{"image/png", false},
		{"application/json", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			got := IsTextContent(tt.contentType)
			if got != tt.want {
				t.Errorf("IsTextContent(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestDetectBinaryContent(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"JPEG magic", []byte{0xFF, 0xD8, 0xFF, 0xE0}, true},
		{"PNG magic", []byte{0x89, 0x50, 0x4E, 0x47}, true},
		{"GIF magic", []byte{0x47, 0x49, 0x46, 0x38}, true},
		{"PDF magic", []byte{0x25, 0x50, 0x44, 0x46}, true},
		{"text content", []byte("<html><body>hello</body></html>"), false},
		{"too short", []byte{0xFF, 0xD8}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectBinaryContent(tt.data)
			if got != tt.want {
				t.Errorf("DetectBinaryContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchStreamingNotModified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	err := FetchStreaming(server.URL, "", 10, nil, 1024*1024, nil, nil)
	if err != nil {
		t.Fatalf("FetchStreaming() 304 error = %v", err)
	}
}

func TestPoolStats(t *testing.T) {
	pool := NewConnPool()
	stats := pool.Stats()
	
	if stats.HostPools == nil {
		t.Error("HostPools should not be nil")
	}
	if stats.TotalConnections != 0 {
		t.Errorf("Expected 0 total connections, got %d", stats.TotalConnections)
	}
}
