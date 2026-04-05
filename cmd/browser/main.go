package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/fetch"
	"github.com/nickcoury/ViBrowsing/internal/html"
	"github.com/nickcoury/ViBrowsing/internal/layout"
	"github.com/nickcoury/ViBrowsing/internal/render"
)

// HistoryEntry represents a single history entry
type HistoryEntry struct {
	URL   string
	Title string
}

// BrowserState holds browser session state
type BrowserState struct {
	CookieJar *fetch.CookieJar
	History   []HistoryEntry
	HistoryIndex int
}

// PageState holds the current page data for navigation and selection
type PageState struct {
	URL        string
	DOM        *html.Node
	Layout     *layout.Box
	Canvas     *render.Canvas
	ViewportW  int
	ViewportH  int
	Links      []LinkInfo
	Selection  *Selection
	ErrorPage  bool
}

// Rect represents a rectangle for link hit-testing
type Rect struct {
	X, Y, W, H int
}

// LinkTarget represents the target frame/window for a link
type LinkTarget int

const (
	LinkTargetSelf   LinkTarget = iota // Current frame/window (default)
	LinkTargetBlank                    // New window/tab
	LinkTargetParent                   // Parent frame
	LinkTargetTop                      // Top-level window
	LinkTargetNamed                    // Named frame/window
)

// LinkInfo stores the bounding box and URL for a clickable link
type LinkInfo struct {
	Box    *layout.Box
	HREF   string
	Rect   Rect
	Target LinkTarget
	TargetName string // For named targets (e.g., target="myframe")
}

// Selection stores text selection state
type Selection struct {
	StartBox *layout.Box
	StartOffset int
	EndBox   *layout.Box
	EndOffset   int
	Text     string
	StartX, StartY int
	EndX, EndY     int
}

var (
	flagDumpDOM    = flag.Bool("dump-dom", false, "Print the parsed DOM tree to stdout")
	flagDumpLayout = flag.Bool("dump-layout", false, "Print the layout box tree to stdout")
	flagViewport   = flag.String("viewport", "800x600", "Viewport size as WxH (e.g. 375x667)")
	flagDebug      = flag.Bool("debug", false, "Enable verbose debug output")
	flagUserAgent  = flag.String("user-agent", "", "Set the User-Agent header")
	flagOutput     = flag.String("output", "output.png", "Output file path")
)

func main() {
	flag.Parse()

	args := flag.Args()
	url := "https://example.com"
	if len(args) > 0 {
		url = args[0]
		// If it looks like a local file path, prefix with file://
		if _, err := os.Stat(url); err == nil && !strings.HasPrefix(url, "http") {
			url = "file://" + url
		}
	}

	// Parse viewport
	viewportW, viewportH := 800, 600
	if v := *flagViewport; v != "" {
		parts := strings.Split(v, "x")
		if len(parts) == 2 {
			if w, err := strconv.Atoi(parts[0]); err == nil {
				viewportW = w
			}
			if h, err := strconv.Atoi(parts[1]); err == nil {
				viewportH = h
			}
		}
	}

	// Initialize browser state
	browser := &BrowserState{
		CookieJar:    fetch.NewCookieJar(),
		History:      make([]HistoryEntry, 0),
		HistoryIndex: -1,
	}

	// Navigate to the URL
	page := navigateToURL(browser, url, *flagUserAgent, viewportW, viewportH)
	if page == nil {
		os.Exit(1)
	}

	// Handle history navigation commands
	// For now, just render the page (history navigation would be interactive)

	// Also print extracted text as sanity check
	body := page.DOM.QuerySelectorAll("body")
	if len(body) > 0 {
		text := strings.TrimSpace(body[0].InnerText())
		if len(text) > 200 {
			text = text[:200] + "..."
		}
		fmt.Printf("\nExtracted text: %s\n", text)
	}
}

// navigateToURL fetches a URL and builds the page state
func navigateToURL(browser *BrowserState, url, userAgent string, viewportW, viewportH int) *PageState {
	if *flagDebug {
		fmt.Printf("ViBrowsing fetching: %s\n", url)
	}

	// Fetch the page with cookie support
	resp, err := fetch.Fetch(url, userAgent, 10, browser.CookieJar, nil)
	if err != nil {
		fmt.Printf("Fetch error: %v\n", err)
		return nil
	}

	// Add to history (clear any forward history when navigating to new URL)
	if browser.HistoryIndex < len(browser.History)-1 {
		browser.History = browser.History[:browser.HistoryIndex+1]
	}
	browser.History = append(browser.History, HistoryEntry{URL: resp.FinalURL})
	browser.HistoryIndex = len(browser.History) - 1

	if *flagDebug {
		fmt.Printf("History: %d entries, current: %d\n", len(browser.History), browser.HistoryIndex)
	}

	// Build DOM tree
	dom := html.Parse(resp.Body)

	if resp.StatusCode >= 400 {
		// Render an error page instead of the error response
		errorPageHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>Error %d</title></head>
<body style="font-family: sans-serif; padding: 40px; background: #f5f5f5;">
<div style="background: white; border: 1px solid #ddd; border-radius: 4px; padding: 30px; max-width: 500px; margin: 0 auto;">
<h1 style="color: #d32f2f; margin: 0 0 20px;">Error %d</h1>
<p style="color: #333; margin: 0;">The requested URL could not be loaded.</p>
<p style="color: #666; font-size: 14px; margin: 10px 0 0;">URL: %s</p>
</div>
</body>
</html>`, resp.StatusCode, resp.StatusCode, resp.FinalURL)
		dom = html.Parse([]byte(errorPageHTML))
		resp.Body = []byte(errorPageHTML)
		resp.ContentType = "text/html"
		if *flagDebug {
			fmt.Printf("HTTP error: %d %s — rendering error page\n", resp.StatusCode, resp.FinalURL)
		}
	}

	if *flagDebug {
		fmt.Printf("Fetched: %d bytes, Content-Type: %s\n", len(resp.Body), resp.ContentType)
	}

	if *flagDumpDOM {
		fmt.Printf("DOM tree:\n%s\n", dom.String())
	}

	// Find stylesheets
	var cssRules []css.Rule

	// Extract inline styles and <style> tags
	styleNodes := dom.QuerySelectorAll("style")
	for _, node := range styleNodes {
		sheet := node.InnerText()
		if sheet != "" {
			cssRules = append(cssRules, css.Parse(sheet)...)
		}
	}

	// Fetch external stylesheets from <link rel="stylesheet"> tags
	linkNodes := dom.QuerySelectorAll("link")
	for _, node := range linkNodes {
		rel := strings.ToLower(node.GetAttribute("rel"))
		if rel != "stylesheet" {
			continue
		}
		href := node.GetAttribute("href")
		if href == "" {
			continue
		}
		// Resolve the stylesheet URL against the base URL
		sheetURL := fetch.ResolveURL(href, resp.FinalURL)
		if *flagDebug {
			fmt.Printf("Fetching stylesheet: %s\n", sheetURL)
		}
		cssText, err := fetch.FetchStylesheet(sheetURL, *flagUserAgent, 10, browser.CookieJar)
		if err != nil {
			if *flagDebug {
				fmt.Printf("  Failed to fetch stylesheet %s: %v\n", sheetURL, err)
			}
			continue
		}
		cssRules = append(cssRules, css.Parse(cssText)...)
	}

	// Build layout tree
	layoutBox := layout.BuildLayoutTree(dom, cssRules, viewportW, viewportH)
	if layoutBox == nil {
		fmt.Println("No layout (no body found)")
		return nil
	}

	// Layout the page
	layout.LayoutBlock(layoutBox, float64(viewportW))

	if *flagDumpLayout {
		fmt.Printf("Layout tree:\n%s\n", layoutBox.String())
	}

	// Render to canvas
	canvas := render.NewCanvas(viewportW, viewportH)
	canvas.Clear()
	canvas.DrawBox(layoutBox)

	// Save as PNG
	outputFile := *flagOutput
	if err := canvas.SavePNG(outputFile); err != nil {
		fmt.Printf("Failed to save PNG: %v\n", err)
		return nil
	}

	if *flagDebug {
		fmt.Printf("Rendered to %s\n", outputFile)
	}

	// Extract title for history
	title := fetch.ExtractTitle(dom)
	if title != "" {
		browser.History[browser.HistoryIndex].Title = title
	}

	// Extract links with their targets
	links := extractLinks(layoutBox, dom, resp.FinalURL)

	return &PageState{
		URL:        resp.FinalURL,
		DOM:        dom,
		Layout:     layoutBox,
		Canvas:     canvas,
		ViewportW:  viewportW,
		ViewportH:  viewportH,
		Links:      links,
		Selection:  nil,
		ErrorPage:  resp.StatusCode >= 400,
	}
}

// extractLinks extracts all links from the DOM and layout, with their targets
func extractLinks(layoutBox *layout.Box, dom *html.Node, baseURL string) []LinkInfo {
	var links []LinkInfo
	if dom == nil {
		return links
	}

	// Find all <a> elements
	anchorNodes := dom.QuerySelectorAll("a")
	for _, a := range anchorNodes {
		href := a.GetAttribute("href")
		if href == "" {
			continue
		}

		// Resolve the URL
		resolvedHREF := fetch.ResolveURL(href, baseURL)

		// Get the target attribute
		target := a.GetAttribute("target")
		linkTarget, targetName := parseLinkTarget(target)

		// Try to find the layout box for this anchor
		// For now, we store the info even without precise bounds
		linkInfo := LinkInfo{
			HREF:       resolvedHREF,
			Target:     linkTarget,
			TargetName: targetName,
		}
		links = append(links, linkInfo)
	}

	return links
}

// parseLinkTarget parses the target attribute and returns the LinkTarget and target name
func parseLinkTarget(target string) (LinkTarget, string) {
	switch strings.ToLower(target) {
	case "_blank":
		return LinkTargetBlank, ""
	case "_self":
		return LinkTargetSelf, ""
	case "_parent":
		return LinkTargetParent, ""
	case "_top":
		return LinkTargetTop, ""
	case "":
		return LinkTargetSelf, ""
	default:
		// Named target (frame or window name)
		return LinkTargetNamed, target
	}
}
