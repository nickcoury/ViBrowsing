package fetch

import (
	"strings"

	"github.com/nickcoury/ViBrowsing/internal/html"
)

// ExtractTitle finds the <title> element content in the DOM.
func ExtractTitle(dom *html.Node) string {
	if dom == nil {
		return ""
	}
	return findTitle(dom)
}

func findTitle(n *html.Node) string {
	if n == nil {
		return ""
	}
	if strings.EqualFold(n.TagName, "title") {
		return strings.TrimSpace(n.InnerText())
	}
	for _, child := range n.Children {
		if t := findTitle(child); t != "" {
			return t
		}
	}
	return ""
}

// ExtractFaviconURL finds the favicon URL from <link rel="icon"> or defaults to /favicon.ico.
func ExtractFaviconURL(dom *html.Node, baseURL string) string {
	if dom == nil {
		return resolveURL("/favicon.ico", baseURL)
	}
	links := dom.QuerySelectorAll("link")
	for _, link := range links {
		rel := strings.ToLower(link.GetAttribute("rel"))
		if strings.Contains(rel, "icon") || strings.Contains(rel, "shortcut") {
			href := link.GetAttribute("href")
			if href != "" {
				return resolveURL(href, baseURL)
			}
		}
	}
	return resolveURL("/favicon.ico", baseURL)
}

// resolveURL resolves a potentially relative URL against a base URL.
func resolveURL(rawURL, baseURL string) string {
	if rawURL == "" {
		return baseURL
	}
	// If it looks like an absolute URL, return as-is
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}
	// Parse base URL
	base, err := parseURL(baseURL)
	if err != nil {
		return rawURL
	}
	// Handle protocol-relative
	if strings.HasPrefix(rawURL, "//") {
		return base.Scheme + ":" + rawURL
	}
	// Handle absolute path
	if strings.HasPrefix(rawURL, "/") {
		return base.Scheme + "://" + base.Host + rawURL
	}
	// Handle relative path — get directory of base path
	path := base.Path
	if !strings.HasSuffix(path, "/") {
		// Get directory
		dirEnd := strings.LastIndex(path, "/")
		if dirEnd >= 0 {
			path = path[:dirEnd+1]
		} else {
			path = "/"
		}
	}
	return base.Scheme + "://" + base.Host + path + rawURL
}

type parsedURL struct {
	Scheme string
	Host   string
	Path   string
}

func parseURL(rawURL string) (parsedURL, error) {
	var p parsedURL
	// Simple URL parser without net/url dependency changes
	// Find scheme
	idx := strings.Index(rawURL, "://")
	if idx >= 0 {
		p.Scheme = rawURL[:idx]
		rest := rawURL[idx+3:]
		// Find first /
		slashIdx := strings.Index(rest, "/")
		if slashIdx >= 0 {
			p.Host = rest[:slashIdx]
			p.Path = rest[slashIdx:]
		} else {
			p.Host = rest
			p.Path = "/"
		}
	} else {
		p.Path = rawURL
	}
	return p, nil
}
