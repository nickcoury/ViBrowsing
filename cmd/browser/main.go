package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/fetch"
	"github.com/nickcoury/ViBrowsing/internal/html"
	"github.com/nickcoury/ViBrowsing/internal/layout"
	"github.com/nickcoury/ViBrowsing/internal/render"
)

func main() {
	url := "https://example.com"
	if len(os.Args) > 1 {
		url = os.Args[1]
		// If it looks like a local file path, prefix with file://
		if _, err := os.Stat(url); err == nil && !strings.HasPrefix(url, "http") {
			url = "file://" + url
		}
	}

	fmt.Printf("ViBrowsing fetching: %s\n", url)

	// Fetch the page
	resp, err := fetch.Fetch(url, 5)
	if err != nil {
		fmt.Printf("Fetch error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Fetched: %d bytes, Content-Type: %s\n", len(resp.Body), resp.ContentType)

	// Build DOM tree
	dom := html.Parse(resp.Body)
	fmt.Printf("DOM tree:\n%s\n", dom.String())

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
	viewportW := 800
	viewportH := 600
	layout.LayoutBlock(layoutBox, float64(viewportW))

	// Render to canvas
	canvas := render.NewCanvas(viewportW, viewportH)
	canvas.Clear()
	canvas.DrawBox(layoutBox)

	// Save as PNG
	outputFile := "output.png"
	if err := canvas.SavePNG(outputFile); err != nil {
		fmt.Printf("Failed to save PNG: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Rendered to %s\n", outputFile)

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
