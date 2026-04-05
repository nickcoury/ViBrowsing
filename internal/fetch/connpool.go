package fetch

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// MaxConnsPerHost is the maximum number of connections to keep per host.
const MaxConnsPerHost = 6

// MaxIdleConns is the maximum number of idle connections across all hosts.
const MaxIdleConns = 100

// ConnPool manages a pool of HTTP connections for keep-alive.
type ConnPool struct {
	mu       sync.RWMutex
	pools    map[string]*hostPool
	transport *http.Transport
	client   *http.Client
}

// hostPool manages connections for a specific host.
type hostPool struct {
	mu       sync.RWMutex
	conns    []*persistentConn
	maxConns int
}

// persistentConn represents a reusable connection to a host.
type persistentConn struct {
	host   string
	readAt time.Time
}

// NewConnPool creates a new connection pool.
func NewConnPool() *ConnPool {
	pool := &ConnPool{
		pools: make(map[string]*hostPool),
	}

	transport := &http.Transport{
		MaxIdleConns:        MaxIdleConns,
		MaxIdleConnsPerHost: MaxConnsPerHost,
		IdleConnTimeout:     90 * time.Second,
		// Set to false for HTTP/1.1 keep-alive (connections are reused)
		DisableKeepAlives: false,
	}

	pool.transport = transport
	pool.client = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return pool
}

// GetClient returns the HTTP client for this pool.
func (p *ConnPool) GetClient() *http.Client {
	return p.client
}

// GetTransport returns the transport used by this pool.
func (p *ConnPool) GetTransport() *http.Transport {
	return p.transport
}

// SetTimeout sets the timeout for the HTTP client.
func (p *ConnPool) SetTimeout(timeout time.Duration) {
	p.client.Timeout = timeout
}

// PoolStats contains statistics about the connection pool.
type PoolStats struct {
	TotalConnections int
	IdleConnections  int
	HostPools        map[string]int
}

// Stats returns current connection pool statistics.
func (p *ConnPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := PoolStats{
		HostPools: make(map[string]int),
	}

	for host, pool := range p.pools {
		pool.mu.Lock()
		stats.HostPools[host] = len(pool.conns)
		stats.TotalConnections += len(pool.conns)
		pool.mu.Unlock()
	}

	return stats
}

// Close closes all connections in the pool.
func (p *ConnPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, pool := range p.pools {
		pool.mu.Lock()
		for range pool.conns {
			// Connection cleanup handled by transport
		}
		pool.conns = nil
		pool.mu.Unlock()
	}

	p.transport.CloseIdleConnections()
}

// hostKey generates a unique key for a host (host:port).
func hostKey(scheme, host string) string {
	return scheme + "://" + host
}

// DefaultConnPool is the global connection pool instance.
var DefaultConnPool = NewConnPool()

// GetDefaultPool returns the default connection pool.
func GetDefaultPool() *ConnPool {
	return DefaultConnPool
}

// ConditionalRequest holds If-Modified-Since timestamp for conditional GET.
type ConditionalRequest struct {
	LastModified time.Time
	URL          string
}

// ConditionalCache stores conditional GET state per URL.
type ConditionalCache struct {
	mu       sync.RWMutex
	requests map[string]ConditionalRequest
}

// NewConditionalCache creates a new conditional GET cache.
func NewConditionalCache() *ConditionalCache {
	return &ConditionalCache{
		requests: make(map[string]ConditionalRequest),
	}
}

// SetLastModified stores the Last-Modified time for a URL.
func (c *ConditionalCache) SetLastModified(rawURL string, lastModified time.Time) {
	if lastModified.IsZero() {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requests[rawURL] = ConditionalRequest{
		LastModified: lastModified,
		URL:          rawURL,
	}
}

// GetLastModified returns the stored Last-Modified time for a URL.
func (c *ConditionalCache) GetLastModified(rawURL string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	req, ok := c.requests[rawURL]
	if !ok {
		return time.Time{}, false
	}
	return req.LastModified, true
}

// DefaultConditionalCache is the global conditional GET cache.
var DefaultConditionalCache = NewConditionalCache()

// GetDefaultConditionalCache returns the default conditional GET cache.
func GetDefaultConditionalCache() *ConditionalCache {
	return DefaultConditionalCache
}

// CompressionConfig holds compression settings.
type CompressionConfig struct {
	AcceptGzip    bool
	AcceptDeflate bool
	AcceptBr      bool // brotli
}

// DefaultCompressionConfig is the default compression settings (all enabled).
var DefaultCompressionConfig = &CompressionConfig{
	AcceptGzip:    true,
	AcceptDeflate: true,
	AcceptBr:      true,
}

// GetAcceptEncodingHeader returns the Accept-Encoding header value.
func (c *CompressionConfig) GetAcceptEncodingHeader() string {
	var encodings []string
	if c.AcceptGzip {
		encodings = append(encodings, "gzip")
	}
	if c.AcceptDeflate {
		encodings = append(encodings, "deflate")
	}
	if c.AcceptBr {
		encodings = append(encodings, "br")
	}
	return strings.Join(encodings, ", ")
}

// ParseContentEncoding parses Content-Encoding header value.
func ParseContentEncoding(encoding string) string {
	encoding = strings.TrimSpace(strings.ToLower(encoding))
	if encoding == "gzip" {
		return "gzip"
	}
	if encoding == "deflate" {
		return "deflate"
	}
	if encoding == "br" {
		return "br"
	}
	if encoding == "compress" {
		return "compress"
	}
	if encoding == "identity" {
		return "identity"
	}
	return ""
}

// FetchProgress holds progress information for streaming fetches.
type FetchProgress struct {
	URL           string
	BytesRead     int64
	ContentLength int64
}

// StreamingOptions contains options for streaming fetches.
type StreamingOptions struct {
	OnProgress func(progress FetchProgress)
	CancelFunc func() // Call to cancel the fetch
}
