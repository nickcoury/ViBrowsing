package fetch

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Cookie represents an HTTP cookie.
type Cookie struct {
	Name     string
	Value    string
	Domain   string
	Path     string
	Expires  time.Time
	Secure   bool
	HttpOnly bool
}

// CookieJar manages cookies for a browser session.
type CookieJar struct {
	mu  sync.RWMutex
	jar *cookiejar.Jar
}

// NewCookieJar creates a new CookieJar.
func NewCookieJar() *CookieJar {
	jar, _ := cookiejar.New(nil)
	return &CookieJar{jar: jar}
}

// SetCookies parses Set-Cookie header values and stores them.
func (j *CookieJar) SetCookies(rawURL string, setCookieHeaders []string) {
	if len(setCookieHeaders) == 0 {
		return
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	for _, header := range setCookieHeaders {
		c := parseSetCookieHeader(header)
		if c.Name == "" {
			continue
		}
		if c.Domain == "" {
			c.Domain = u.Host
		}
		if c.Path == "" {
			c.Path = "/"
		}
		j.jar.SetCookies(u, []*http.Cookie{
			{
				Name:     c.Name,
				Value:    c.Value,
				Domain:   c.Domain,
				Path:     c.Path,
				Expires:  c.Expires,
				Secure:   c.Secure,
				HttpOnly: c.HttpOnly,
			},
		})
	}
}

// GetCookies returns cookies for the given URL as a Cookie header string.
func (j *CookieJar) GetCookies(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	cookies := j.jar.Cookies(u)
	if len(cookies) == 0 {
		return ""
	}
	var parts []string
	for _, c := range cookies {
		parts = append(parts, c.Name+"="+c.Value)
	}
	return strings.Join(parts, "; ")
}

// GetCookieObjects returns raw Cookie structs for a URL.
func (j *CookieJar) GetCookieObjects(rawURL string) []Cookie {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	cookies := j.jar.Cookies(u)
	var result []Cookie
	for _, c := range cookies {
		result = append(result, Cookie{
			Name:    c.Name,
			Value:   c.Value,
			Domain:  c.Domain,
			Path:    c.Path,
			Expires: c.Expires,
			Secure:  c.Secure,
		})
	}
	return result
}

// parseSetCookieHeader parses a Set-Cookie header value.
func parseSetCookieHeader(header string) Cookie {
	c := Cookie{}
	parts := strings.Split(header, ";")
	if len(parts) == 0 {
		return c
	}
	// First part is Name=Value
	nameValue := strings.TrimSpace(parts[0])
	idx := strings.Index(nameValue, "=")
	if idx < 0 {
		return c
	}
	c.Name = strings.TrimSpace(nameValue[:idx])
	c.Value = strings.TrimSpace(nameValue[idx+1:])
	for i := 1; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		lower := strings.ToLower(part)
		if strings.HasPrefix(lower, "domain=") {
			c.Domain = strings.Trim(strings.TrimPrefix(part, "domain="), "\"")
		} else if strings.HasPrefix(lower, "path=") {
			c.Path = strings.Trim(strings.TrimPrefix(part, "path="), "\"")
		} else if strings.HasPrefix(lower, "expires=") {
			expStr := strings.TrimPrefix(part, "expires=")
			if t, err := time.Parse(time.RFC1123, strings.Trim(expStr, "\"")); err == nil {
				c.Expires = t
			}
		} else if lower == "secure" {
			c.Secure = true
		} else if lower == "httponly" {
			c.HttpOnly = true
		}
	}
	return c
}
