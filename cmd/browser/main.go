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

// LinkInfo stores the bounding box and URL for a clickable link
type LinkInfo struct {
	Box  *layout.Box
	HREF string
	Rect Rect
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

	if *flagDebug {
		fmt.Printf("ViBrowsing fetching: %s\n", url)
	}

	// Fetch the page
	resp, err := fetch.Fetch(url, *flagUserAgent, 10)
	if err != nil {
		fmt.Printf("Fetch error: %v\n", err)
		os.Exit(1)
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

	// Build layout tree
	layoutBox := layout.BuildLayoutTree(dom, cssRules)
	if layoutBox == nil {
		fmt.Println("No layout (no body found)")
		os.Exit(1)
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
		os.Exit(1)
	}

	if *flagDebug {
		fmt.Printf("Rendered to %s\n", outputFile)
	}

	// Also print extracted text as sanity check
	body := dom.QuerySelectorAll("body")
	if len(body) > 0 {
		text := strings.TrimSpace(body[0].InnerText())
		if len(text) > 200 {
			text = text[:200] + "..."
		}
		fmt.Printf("\nExtracted text: %s\n", text)
	}
}
