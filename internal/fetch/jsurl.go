package fetch

import (
	"net/url"
	"strings"
)

// URL represents a URL object as per the URL standard.
// Implements the URL interface: https://developer.mozilla.org/en-US/docs/Web/API/URL
type JSURL struct {
	parsed *url.URL
}

// NewJSURL creates a new URL object from a raw URL string.
// If rawURL is a relative URL, base must be provided.
func NewJSURL(rawURL string, base string) (*JSURL, error) {
	var u *url.URL
	var err error

	// First, try to parse as absolute URL
	u, err = url.Parse(rawURL)
	if err != nil {
		return nil, &URLParseError{RawURL: rawURL, Reason: err.Error()}
	}

	// If it has no scheme and we have a base, resolve against base
	if u.Scheme == "" && base != "" {
		baseParsed, err := url.Parse(base)
		if err != nil {
			return nil, &URLParseError{RawURL: rawURL, Reason: "invalid base URL: " + err.Error()}
		}
		u = baseParsed.ResolveReference(u)
	}

	if u.Scheme == "" && u.Host == "" && u.Path == "" {
		return nil, &URLParseError{RawURL: rawURL, Reason: "invalid URL: no scheme, host, or path"}
	}

	return &JSURL{parsed: u}, nil
}

// Href returns the full URL string.
func (u *JSURL) Href() string {
	return u.parsed.String()
}

// Protocol returns the protocol scheme (e.g., "https:").
func (u *JSURL) Protocol() string {
	if u.parsed.Scheme == "" {
		return ""
	}
	return u.parsed.Scheme + ":"
}

// Host returns the host and port (e.g., "example.com:8080").
func (u *JSURL) Host() string {
	host := u.parsed.Host
	// Add default port if applicable
	if u.parsed.Port() != "" {
		return host
	}
	// Return standard ports for known schemes
	switch u.parsed.Scheme {
	case "http":
		if u.parsed.Host != "" && !strings.HasSuffix(host, ":80") {
			return host
		}
	case "https":
		if u.parsed.Host != "" && !strings.HasSuffix(host, ":443") {
			return host
		}
	}
	return host
}

// Hostname returns just the hostname (e.g., "example.com").
func (u *JSURL) Hostname() string {
	return u.parsed.Hostname()
}

// Port returns the port number as a string.
func (u *JSURL) Port() string {
	return u.parsed.Port()
}

// Pathname returns the path component (e.g., "/path/to/page").
func (u *JSURL) Pathname() string {
	if u.parsed.Path == "" {
		return "/"
	}
	return u.parsed.Path
}

// Search returns the query string including the leading "?" (e.g., "?foo=bar").
func (u *JSURL) Search() string {
	if u.parsed.RawQuery == "" {
		return ""
	}
	return "?" + u.parsed.RawQuery
}

// SearchParams returns a URLSearchParams object for manipulating the query string.
func (u *JSURL) SearchParams() *URLSearchParams {
	return NewURLSearchParams(u.parsed.RawQuery)
}

// Hash returns the fragment identifier including "#".
func (u *JSURL) Hash() string {
	if u.parsed.Fragment == "" {
		return ""
	}
	return "#" + u.parsed.Fragment
}

// Origin returns the origin (scheme + host + port) of the URL.
func (u *JSURL) Origin() string {
	if u.parsed.Host == "" {
		return ""
	}
	host := u.parsed.Host
	port := u.parsed.Port()
	switch u.parsed.Scheme {
	case "http":
		if port == "80" {
			host = u.parsed.Hostname()
		}
	case "https":
		if port == "443" {
			host = u.parsed.Hostname()
		}
	}
	return u.parsed.Scheme + "://" + host
}

// Username returns the username part of the URL.
func (u *JSURL) Username() string {
	return u.parsed.User.Username()
}

// Password returns the password part of the URL.
func (u *JSURL) Password() string {
	password, _ := u.parsed.User.Password()
	return password
}

// Pathname sets the path component.
func (u *JSURL) SetPathname(path string) {
	u.parsed.Path = path
}

// Search sets the query string (without the leading "?").
func (u *JSURL) SetSearch(search string) {
	if strings.HasPrefix(search, "?") {
		search = search[1:]
	}
	u.parsed.RawQuery = search
}

// Hash sets the fragment identifier (without the leading "#").
func (u *JSURL) SetHash(hash string) {
	if strings.HasPrefix(hash, "#") {
		hash = hash[1:]
	}
	u.parsed.Fragment = hash
}

// Hostname setter
func (u *JSURL) SetHostname(hostname string) {
	u.parsed.Host = hostname + u.parsed.Port()
}

// Port setter
func (u *JSURL) SetPort(port string) {
	host := u.parsed.Hostname()
	if port != "" {
		u.parsed.Host = host + ":" + port
	} else {
		u.parsed.Host = host
	}
}

// Protocol setter
func (u *JSURL) SetProtocol(protocol string) {
	// Remove trailing colon if present
	protocol = strings.TrimSuffix(protocol, ":")
	u.parsed.Scheme = protocol
}

// ToString returns the string representation of the URL.
func (u *JSURL) ToString() string {
	return u.parsed.String()
}

// String implements fmt.Stringer.
func (u *JSURL) String() string {
	return u.parsed.String()
}

// URLParseError represents an error when parsing a URL.
type URLParseError struct {
	RawURL string
	Reason string
}

func (e *URLParseError) Error() string {
	return "url parse error: " + e.Reason + " (raw URL: " + e.RawURL + ")"
}

// URLSearchParams represents a collection of key/value pairs for URL query strings.
// Implements the URLSearchParams interface: https://developer.mozilla.org/en-US/docs/Web/API/URLSearchParams
type URLSearchParams struct {
	params map[string][]string
}

// NewURLSearchParams creates a new URLSearchParams object.
// It accepts either a query string, an object of key/value pairs, or nothing.
func NewURLSearchParams(query interface{}) *URLSearchParams {
	usp := &URLSearchParams{
		params: make(map[string][]string),
	}

	switch q := query.(type) {
	case string:
		usp.parseQueryString(q)
	case map[string]string:
		for k, v := range q {
			usp.params[k] = append(usp.params[k], v)
		}
	case map[string]interface{}:
		for k, v := range q {
			if vs, ok := v.(string); ok {
				usp.params[k] = append(usp.params[k], vs)
			}
		}
	}

	return usp
}

// parseQueryString parses a query string into key/value pairs.
func (usp *URLSearchParams) parseQueryString(query string) {
	// Remove leading ? if present
	if strings.HasPrefix(query, "?") {
		query = query[1:]
	}

	if query == "" {
		return
	}

	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}

		kv := strings.SplitN(pair, "=", 2)
		key := kv[0]
		value := ""
		if len(kv) == 2 {
			value = kv[1]
		}

		// URL decode
		key = decodePercentEncoding(key)
		value = decodePercentEncoding(value)

		usp.params[key] = append(usp.params[key], value)
	}
}

// decodePercentEncoding decodes percent-encoded characters.
func decodePercentEncoding(s string) string {
	// Use url.QueryEscape would encode, not decode
	// We need to decode manually
	result := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		if s[i] == '%' && i+2 < len(s) {
			// Parse hex digits
			var hex byte
			for j := 0; j < 2; j++ {
				hex <<= 4
				c := s[i+1+j]
				switch {
				case c >= '0' && c <= '9':
					hex |= c - '0'
				case c >= 'A' && c <= 'F':
					hex |= c - 'A' + 10
				case c >= 'a' && c <= 'f':
					hex |= c - 'a' + 10
				default:
					// Invalid hex, keep as-is
					hex = 0
					break
				}
			}
			result = append(result, hex)
			i += 3
		} else {
			result = append(result, s[i])
			i++
		}
	}
	return string(result)
}

// Append adds a new parameter with the given name and value.
func (usp *URLSearchParams) Append(name, value string) {
	usp.params[name] = append(usp.params[name], value)
}

// Delete removes all parameters with the given name.
func (usp *URLSearchParams) Delete(name string) {
	delete(usp.params, name)
}

// Get returns the first value associated with the given name.
func (usp *URLSearchParams) Get(name string) string {
	values, ok := usp.params[name]
	if !ok || len(values) == 0 {
		return ""
	}
	return values[0]
}

// GetAll returns all values associated with the given name.
func (usp *URLSearchParams) GetAll(name string) []string {
	values, ok := usp.params[name]
	if !ok {
		return []string{}
	}
	return values
}

// Has returns true if a parameter with the given name exists.
func (usp *URLSearchParams) Has(name string) bool {
	_, ok := usp.params[name]
	return ok
}

// Set sets the value for the given name, replacing all existing values.
func (usp *URLSearchParams) Set(name, value string) {
	usp.params[name] = []string{value}
}

// Sort sorts all key/value pairs by name.
func (usp *URLSearchParams) Sort() {
	// Create sorted slice of keys
	keys := make([]string, 0, len(usp.params))
	for k := range usp.params {
		keys = append(keys, k)
	}
	sortStrings(keys)

	// Rebuild map in sorted order
	sorted := make(map[string][]string, len(usp.params))
	for _, k := range keys {
		sorted[k] = usp.params[k]
	}
	usp.params = sorted
}

// sortStrings sorts a slice of strings in place.
func sortStrings(s []string) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// ToString returns the query string representation.
func (usp *URLSearchParams) ToString() string {
	return usp.Encode()
}

// Encode returns the URL-encoded query string.
func (usp *URLSearchParams) Encode() string {
	if len(usp.params) == 0 {
		return ""
	}

	var parts []string
	for name, values := range usp.params {
		encodedName := encodePercentEncoding(name)
		for _, value := range values {
			encodedValue := encodePercentEncoding(value)
			parts = append(parts, encodedName+"="+encodedValue)
		}
	}

	return strings.Join(parts, "&")
}

// encodePercentEncoding percent-encodes special characters.
func encodePercentEncoding(s string) string {
	// Use url.PathEscape for proper encoding
	// It encodes everything except unreserved characters
	const unreserved = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.~"
	result := make([]byte, 0, len(s)*3)
	for _, c := range []byte(s) {
		if strings.Contains(unreserved, string(c)) {
			result = append(result, c)
		} else {
			result = append(result, '%')
			result = append(result, "0123456789ABCDEF"[c>>4])
			result = append(result, "0123456789ABCDEF"[c&0x0F])
		}
	}
	return string(result)
}

// Keys returns an iterator over all parameter names.
func (usp *URLSearchParams) Keys() []string {
	keys := make([]string, 0, len(usp.params))
	for k := range usp.params {
		keys = append(keys, k)
	}
	// Sort keys alphabetically
	sortStrings(keys)
	return keys
}

// Values returns an iterator over all parameter values.
func (usp *URLSearchParams) Values() []string {
	values := make([]string, 0)
	for _, vs := range usp.params {
		values = append(values, vs...)
	}
	return values
}

// Entries returns each key/value pair as a slice of [2]string.
func (usp *URLSearchParams) Entries() [][2]string {
	entries := make([][2]string, 0)
	for name, values := range usp.params {
		for _, value := range values {
			entries = append(entries, [2]string{name, value})
		}
	}
	return entries
}

// ForEach calls the callback function for each key/value pair.
func (usp *URLSearchParams) ForEach(callback func(value, key string)) {
	for name, values := range usp.params {
		for _, value := range values {
			callback(value, name)
		}
	}
}
