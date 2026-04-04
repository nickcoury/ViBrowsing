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
	if resp.StatusCode >= 400 {
		fmt.Printf("HTTP error: %d %s\n", resp.StatusCode, resp.FinalURL)
		os.Exit(1)
	}
	if *flagDebug {
		fmt.Printf("Fetched: %d bytes, Content-Type: %s\n", len(resp.Body), resp.ContentType)
	}

	// Build DOM tree
	dom := html.Parse(resp.Body)

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
