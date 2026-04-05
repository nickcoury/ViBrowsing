package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/fetch"
	"github.com/nickcoury/ViBrowsing/internal/html"
	"github.com/nickcoury/ViBrowsing/internal/layout"
	"github.com/nickcoury/ViBrowsing/internal/render"
)

// Browser holds the browser state
type Browser struct {
	CookieJar       *fetch.CookieJar
	CurrentURL      string
	Page            *PageState
	ViewportW       int
	ViewportH       int
	ScrollY         int
	MaxScrollY      int
	URLBarText      string
	URLBarFocused   bool
	URLBarCursor    int // cursor position in URL bar
	Loading         bool
	ErrorMsg        string
	HoveredLinkIdx  int
	History         []string
	HistoryIndex    int
	PendingNav      string // URL being navigated to (for async fetch)
}

// PageState holds the current page data
type PageState struct {
	URL    string
	DOM    *html.Node
	Layout *layout.Box
	Canvas *render.Canvas
	Links  []LinkInfo
}

// LinkInfo stores the bounding box and URL for a clickable link
type LinkInfo struct {
	Box  *layout.Box
	HREF string
	Rect image.Rectangle
}

const (
	urlBarHeight  = 0  // integrated into nav bar
	navBarHeight  = 40
	statusBarH    = 24
	totalChrome   = navBarHeight + statusBarH
	contentOffset = navBarHeight
)

func main() {
	ebiten.SetWindowTitle("ViBrowsing")
	ebiten.SetWindowSize(1024, 768)

	browser := &Browser{
		ViewportW:    1024,
		ViewportH:    768,
		CookieJar:    fetch.NewCookieJar(),
		URLBarText:   "https://example.com",
		URLBarFocused: false,
		HoveredLinkIdx: -1,
		History:     []string{},
		HistoryIndex: -1,
	}

	if err := ebiten.RunGame(browser); err != nil {
		panic(err)
	}
}

// Layout implements ebiten.Game's Layout
func (b *Browser) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	b.ViewportW = outsideWidth
	b.ViewportH = outsideHeight
	return outsideWidth, outsideHeight
}

// Update runs every frame
func (b *Browser) Update() error {
	// Check if we need to start a navigation
	if b.PendingNav != "" {
		url := b.PendingNav
		b.PendingNav = ""
		b.navigateTo(url)
		return nil
	}

	// Scroll with keyboard
	scrollSpeed := 60
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyPageDown) {
		b.ScrollY += scrollSpeed
		if b.ScrollY > b.MaxScrollY {
			b.ScrollY = b.MaxScrollY
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyPageUp) {
		b.ScrollY -= scrollSpeed
		if b.ScrollY < 0 {
			b.ScrollY = 0
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyHome) {
		b.ScrollY = 0
	}
	if ebiten.IsKeyPressed(ebiten.KeyEnd) {
		b.ScrollY = b.MaxScrollY
	}

	// Back/Forward with Alt+Arrow
	if ebiten.IsKeyPressed(ebiten.KeyAlt) {
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) && b.HistoryIndex > 0 {
			b.HistoryIndex--
			b.navigateTo(b.History[b.HistoryIndex])
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) && b.HistoryIndex < len(b.History)-1 {
			b.HistoryIndex++
			b.navigateTo(b.History[b.HistoryIndex])
		}
	}

	// Mouse wheel scrolling
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		b.ScrollY -= int(wheelY * 50)
		if b.ScrollY < 0 {
			b.ScrollY = 0
		}
		if b.ScrollY > b.MaxScrollY {
			b.ScrollY = b.MaxScrollY
		}
	}

	// Check if mouse is in URL bar area
	mx, my := ebiten.CursorPosition()
	_ = mx

	// URL bar click detection
	if my >= 0 && my < navBarHeight {
		// Check for nav button clicks
		if mx >= 10 && mx <= 50 && !b.URLBarFocused {
			// Back button
			if b.HistoryIndex > 0 {
				b.HistoryIndex--
				b.navigateTo(b.History[b.HistoryIndex])
			}
		} else if mx >= 55 && mx <= 95 && !b.URLBarFocused {
			// Forward button
			if b.HistoryIndex < len(b.History)-1 {
				b.HistoryIndex++
				b.navigateTo(b.History[b.HistoryIndex])
			}
		} else if mx >= 100 && mx <= b.ViewportW-10 {
			// URL bar
			b.URLBarFocused = true
		}
	} else {
		b.URLBarFocused = false
	}

	// Update hovered link
	b.HoveredLinkIdx = -1
	if b.Page != nil && b.Page.Canvas != nil && my >= contentOffset {
		relY := my - contentOffset + b.ScrollY
		for i, link := range b.Page.Links {
			if link.Rect.Min.X <= mx && mx <= link.Rect.Max.X &&
				link.Rect.Min.Y <= relY && relY <= link.Rect.Max.Y {
				b.HoveredLinkIdx = i
				break
			}
		}
	}

	// Click to navigate
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && b.HoveredLinkIdx >= 0 {
		link := b.Page.Links[b.HoveredLinkIdx]
		b.navigateTo(link.HREF)
	}

	return nil
}

// Draw renders the browser UI
func (b *Browser) Draw(screen *ebiten.Image) {
	// Draw background
	screen.Fill(color.RGBA{R: 255, G: 255, B: 255, A: 255})

	// Draw navigation bar
	navBarImg := ebiten.NewImage(b.ViewportW, navBarHeight)
	navBarImg.Fill(color.RGBA{R: 45, G: 45, B: 55, A: 255})
	screen.DrawImage(navBarImg, &ebiten.DrawImageOptions{})

	// Back button
	backColor := color.RGBA{R: 65, G: 65, B: 75, A: 255}
	if b.HistoryIndex > 0 {
		backColor = color.RGBA{R: 70, G: 120, B: 200, A: 255}
	}
	drawRect(screen, 10, 8, 45, 20, backColor)

	// Forward button
	fwdColor := color.RGBA{R: 65, G: 65, B: 75, A: 255}
	if b.HistoryIndex < len(b.History)-1 && len(b.History) > 0 {
		fwdColor = color.RGBA{R: 70, G: 120, B: 200, A: 255}
	}
	drawRect(screen, 55, 8, 45, 20, fwdColor)

	// Reload button
	reloadColor := color.RGBA{R: 70, G: 120, B: 200, A: 255}
	if b.Loading {
		reloadColor = color.RGBA{R: 150, G: 150, B: 160, A: 255}
	}
	drawRect(screen, 100, 8, 45, 20, reloadColor)

	// URL bar
	urlBarBg := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	drawRect(screen, 155, 8, b.ViewportW-170, 24, urlBarBg)

	// Draw URL text using a simple approach
	urlText := b.URLBarText
	if urlText == "" {
		urlText = "Enter URL..."
	}
	// For simplicity, draw a colored rectangle as placeholder for text
	urlTextColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	// Draw loading spinner or URL
	if b.Loading {
		drawTextAt(screen, "Loading...", 162, navBarHeight/2-6, urlTextColor)
	} else {
		drawTextAt(screen, urlText, 162, navBarHeight/2-6, urlTextColor)
	}

	// Draw page content
	if b.Page != nil && b.Page.Canvas != nil && !b.Loading {
		// Create ebiten image from the canvas pixels
		ebitenImg := ebiten.NewImageFromImage(b.Page.Canvas.Pixels)
		opts := &ebiten.DrawImageOptions{}
		contentGeoM := ebiten.GeoM{}
		contentGeoM.Translate(0, float64(contentOffset))
		contentGeoM.Translate(0, -float64(b.ScrollY))
		opts.GeoM = contentGeoM
		screen.DrawImage(ebitenImg, opts)

		// Highlight hovered link
		if b.HoveredLinkIdx >= 0 {
			link := b.Page.Links[b.HoveredLinkIdx]
			r := link.Rect
			r.Min.Y -= b.ScrollY
			r.Max.Y -= b.ScrollY
			// Only draw if visible
			if r.Max.Y >= contentOffset && r.Min.Y <= b.ViewportH-statusBarH {
				highlightRect := image.Rect(r.Min.X, r.Min.Y+contentOffset, r.Max.X, r.Max.Y+contentOffset)
				drawRectOutline(screen, highlightRect, color.RGBA{R: 0, G: 100, B: 255, A: 100})
			}
		}
	} else if b.ErrorMsg != "" {
		drawTextAt(screen, b.ErrorMsg, 10, contentOffset+20, color.RGBA{R: 200, G: 0, B: 0, A: 255})
	} else if !b.Loading {
		drawTextAt(screen, "ViBrowsing", 10, contentOffset+20, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		drawTextAt(screen, "Enter a URL above", 10, contentOffset+50, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	}

	// Draw status bar
	statusBarImg := ebiten.NewImage(b.ViewportW, statusBarH)
	statusBarImg.Fill(color.RGBA{R: 230, G: 230, B: 235, A: 255})
	statusBarGeoM := ebiten.GeoM{}
	statusBarGeoM.Translate(0, float64(b.ViewportH-statusBarH))
	screen.DrawImage(statusBarImg, &ebiten.DrawImageOptions{GeoM: statusBarGeoM})

	// Status text
	statusText := ""
	if b.Page != nil && b.Page.URL != "" {
		statusText = b.Page.URL
	}
	if b.HoveredLinkIdx >= 0 && b.HoveredLinkIdx < len(b.Page.Links) {
		statusText = b.Page.Links[b.HoveredLinkIdx].HREF
	}
	drawTextAt(screen, statusText, 10, b.ViewportH-statusBarH+6, color.RGBA{R: 60, G: 60, B: 60, A: 255})

	// Scrollbar
	if b.MaxScrollY > 0 {
		viewH := b.ViewportH - totalChrome
		scrollbarH := viewH * viewH / (b.MaxScrollY + viewH)
		if scrollbarH < 20 {
			scrollbarH = 20
		}
		scrollbarThumbY := contentOffset + viewH - int(float64(viewH)*float64(b.ScrollY)/float64(b.MaxScrollY+viewH))
		drawRect(screen, b.ViewportW-12, scrollbarThumbY, 8, int(scrollbarH), color.RGBA{R: 180, G: 180, B: 180, A: 255})
	}
}

func drawRect(screen *ebiten.Image, x, y, w, h int, col color.Color) {
	if w <= 0 || h <= 0 {
		return
	}
	img := ebiten.NewImage(w, h)
	img.Fill(col)
	geoM := ebiten.GeoM{}
	geoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, &ebiten.DrawImageOptions{GeoM: geoM})
}

func drawRectOutline(screen *ebiten.Image, rect image.Rectangle, col color.Color) {
	// Top
	drawRect(screen, rect.Min.X, rect.Min.Y, rect.Dx(), 2, col)
	// Bottom
	drawRect(screen, rect.Min.X, rect.Max.Y-2, rect.Dx(), 2, col)
	// Left
	drawRect(screen, rect.Min.X, rect.Min.Y, 2, rect.Dy(), col)
	// Right
	drawRect(screen, rect.Max.X-2, rect.Min.Y, 2, rect.Dy(), col)
}

func drawTextAt(screen *ebiten.Image, text string, x, y int, col color.Color) {
	// Use a simple approach: draw as debug message at a position
	// For actual text we'd need font rendering, but for now use a white background rect
	// and the ebitenutil will draw on top
	_ = text
	_ = x
	_ = y
	_ = col
	// We'll use ebitenutil.DebugPrint at specific positions instead
}

// navigateTo navigates to a URL (async)
func (b *Browser) navigateTo(rawURL string) {
	if rawURL == "" {
		return
	}

	// Normalize URL
	url := rawURL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		if strings.Contains(url, "."){
			url = "https://" + url
		} else {
			url = "https://www.google.com/search?q=" + url
		}
	}

	// Update history
	if b.HistoryIndex < len(b.History)-1 {
		b.History = b.History[:b.HistoryIndex+1]
	}
	if len(b.History) == 0 || b.History[len(b.History)-1] != url {
		b.History = append(b.History, url)
	}
	b.HistoryIndex = len(b.History) - 1

	b.Loading = true
	b.ErrorMsg = ""
	b.URLBarText = url
	b.ScrollY = 0
	b.MaxScrollY = 0
	b.Page = nil
	b.HoveredLinkIdx = -1
	b.CurrentURL = url

	// Fetch and render in a goroutine
	go b.fetchAndRender(url)
}

func (b *Browser) fetchAndRender(rawURL string) {
	resp, err := fetch.Fetch(rawURL, "", 10, b.CookieJar)
	if err != nil {
		b.Loading = false
		b.ErrorMsg = fmt.Sprintf("Fetch error: %v", err)
		return
	}

	dom := html.Parse(resp.Body)

	if resp.StatusCode >= 400 {
		b.Loading = false
		b.ErrorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return
	}

	// Collect CSS rules
	var cssRules []css.Rule
	for _, node := range dom.QuerySelectorAll("style") {
		sheet := node.InnerText()
		if sheet != "" {
			cssRules = append(cssRules, css.Parse(sheet)...)
		}
	}
	for _, node := range dom.QuerySelectorAll("link") {
		rel := strings.ToLower(node.GetAttribute("rel"))
		if rel != "stylesheet" {
			continue
		}
		href := node.GetAttribute("href")
		if href == "" {
			continue
		}
		sheetURL := fetch.ResolveURL(href, resp.FinalURL)
		cssText, err := fetch.FetchStylesheet(sheetURL, "", 10, b.CookieJar)
		if err != nil {
			continue
		}
		cssRules = append(cssRules, css.Parse(cssText)...)
	}

	layoutBox := layout.BuildLayoutTree(dom, cssRules, b.ViewportW, b.ViewportH-contentOffset)
	if layoutBox == nil {
		b.Loading = false
		b.ErrorMsg = "No body element found"
		return
	}

	layout.LayoutBlock(layoutBox, float64(b.ViewportW))

	canvas := render.NewCanvas(b.ViewportW, b.ViewportH-contentOffset)
	canvas.Clear()
	canvas.DrawBox(layoutBox)

	links := extractLinks(layoutBox, dom, resp.FinalURL)

	pageH := int(layoutBox.ContentH)
	b.MaxScrollY = pageH - (b.ViewportH - totalChrome)
	if b.MaxScrollY < 0 {
		b.MaxScrollY = 0
	}

	b.Page = &PageState{
		URL:    resp.FinalURL,
		DOM:    dom,
		Layout: layoutBox,
		Canvas: canvas,
		Links:  links,
	}
	b.Loading = false
}

func extractLinks(layoutBox *layout.Box, dom *html.Node, baseURL string) []LinkInfo {
	var links []LinkInfo
	if dom == nil {
		return links
	}
	anchorNodes := dom.QuerySelectorAll("a")
	for _, a := range anchorNodes {
		href := a.GetAttribute("href")
		if href == "" {
			continue
		}
		resolvedHREF := fetch.ResolveURL(href, baseURL)
		linkBox := layoutBox.FindBoxByNode(a)
		var rect image.Rectangle
		if linkBox != nil {
			rect = image.Rect(
				int(linkBox.ContentX),
				int(linkBox.ContentY),
				int(linkBox.ContentX+linkBox.ContentW),
				int(linkBox.ContentY+linkBox.ContentH),
			)
		}
		links = append(links, LinkInfo{
			Box:  linkBox,
			HREF: resolvedHREF,
			Rect: rect,
		})
	}
	return links
}
