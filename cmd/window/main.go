package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/fetch"
	"github.com/nickcoury/ViBrowsing/internal/html"
	"github.com/nickcoury/ViBrowsing/internal/layout"
	"github.com/nickcoury/ViBrowsing/internal/render"
	"github.com/nickcoury/ViBrowsing/internal/window"
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
	Typing          string // what user is typing
	TypingFocused   bool   // true when URL bar has focus and user can type
	LastKeyTime     int64  // for detecting first keypress
	linkClicked     bool   // prevents double-click navigation
	KeyHeld         map[ebiten.Key]bool // tracks which keys are currently held
	LastTypingLen   int    // length of Typing last frame (to detect changes)
	// Channel for communication from fetch goroutine to main loop
	pageResult      chan *PageState // nil means error with ErrorMsg set
	navURL          string          // URL being navigated to
	navDone         chan struct{}   // closed when navigation completes
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

	// Default URL
	startURL := "https://example.com"

	browser := &Browser{
		ViewportW:    1024,
		ViewportH:    768,
		CookieJar:    fetch.NewCookieJar(),
		URLBarText:   startURL,
		URLBarFocused: false,
		HoveredLinkIdx: -1,
		History:     []string{},
		HistoryIndex: -1,
		KeyHeld:     make(map[ebiten.Key]bool),
	}

	// Navigate to start URL
	browser.navigateTo(startURL)

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
	// Check if navigation completed
	if b.navDone != nil {
		select {
		case page := <-b.pageResult:
			if page != nil {
				b.Page = page
				pageH := int(page.Layout.ContentH)
				b.MaxScrollY = pageH - (b.ViewportH - totalChrome)
				if b.MaxScrollY < 0 {
					b.MaxScrollY = 0
				}
			}
			b.Loading = false
			b.navDone = nil
			b.pageResult = nil
		default:
			// Still loading
		}
	}

	// Check if we need to start a navigation
	if b.PendingNav != "" {
		url := b.PendingNav
		b.PendingNav = ""
		b.navigateTo(url)
		return nil
	}

	// Handle text typing in URL bar
	if b.TypingFocused {
		// Character keys - only add on first press (edge detection)
		charKeys := []struct {
			key ebiten.Key
			ch  rune
		}{
			{ebiten.KeyA, 'a'}, {ebiten.KeyB, 'b'}, {ebiten.KeyC, 'c'},
			{ebiten.KeyD, 'd'}, {ebiten.KeyE, 'e'}, {ebiten.KeyF, 'f'},
			{ebiten.KeyG, 'g'}, {ebiten.KeyH, 'h'}, {ebiten.KeyI, 'i'},
			{ebiten.KeyJ, 'j'}, {ebiten.KeyK, 'k'}, {ebiten.KeyL, 'l'},
			{ebiten.KeyM, 'm'}, {ebiten.KeyN, 'n'}, {ebiten.KeyO, 'o'},
			{ebiten.KeyP, 'p'}, {ebiten.KeyQ, 'q'}, {ebiten.KeyR, 'r'},
			{ebiten.KeyS, 's'}, {ebiten.KeyT, 't'}, {ebiten.KeyU, 'u'},
			{ebiten.KeyV, 'v'}, {ebiten.KeyW, 'w'}, {ebiten.KeyX, 'x'},
			{ebiten.KeyY, 'y'}, {ebiten.KeyZ, 'z'},
			{ebiten.Key0, '0'}, {ebiten.Key1, '1'}, {ebiten.Key2, '2'},
			{ebiten.Key3, '3'}, {ebiten.Key4, '4'}, {ebiten.Key5, '5'},
			{ebiten.Key6, '6'}, {ebiten.Key7, '7'}, {ebiten.Key8, '8'},
			{ebiten.Key9, '9'},
			{ebiten.KeySpace, ' '},
			{ebiten.KeyComma, ','}, {ebiten.KeyPeriod, '.'},
			{ebiten.KeySlash, '/'}, {ebiten.KeyMinus, '-'},
			{ebiten.KeyEqual, '='},
		}

		for _, ck := range charKeys {
			pressed := ebiten.IsKeyPressed(ck.key)
			if pressed && !b.KeyHeld[ck.key] {
				// First press - add character
				b.Typing += string(ck.ch)
			}
			b.KeyHeld[ck.key] = pressed
		}

		// Backspace - only on edge
		backspacePressed := ebiten.IsKeyPressed(ebiten.KeyBackspace)
		if backspacePressed && !b.KeyHeld[ebiten.KeyBackspace] && len(b.Typing) > 0 {
			// Handle UTF-8 rune deletion
			for i := len(b.Typing) - 1; i >= 0; {
				r, size := utf8.DecodeLastRuneInString(b.Typing[:i+1])
				b.Typing = b.Typing[:i]
				i -= size
				if r != utf8.RuneError || size == 1 {
					break
				}
			}
		}
		b.KeyHeld[ebiten.KeyBackspace] = backspacePressed

		// Enter - submit URL (edge)
		enterPressed := ebiten.IsKeyPressed(ebiten.KeyEnter)
		if enterPressed && !b.KeyHeld[ebiten.KeyEnter] {
			url := b.Typing
			if url != "" {
				b.TypingFocused = false
				b.navigateTo(url)
			}
			b.Typing = ""
		}
		b.KeyHeld[ebiten.KeyEnter] = enterPressed

		// Escape - cancel typing (edge)
		escapePressed := ebiten.IsKeyPressed(ebiten.KeyEscape)
		if escapePressed && !b.KeyHeld[ebiten.KeyEscape] {
			b.TypingFocused = false
			b.Typing = ""
		}
		b.KeyHeld[ebiten.KeyEscape] = escapePressed
	}

	// Scroll with keyboard (only when not typing)
	if !b.TypingFocused {
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

		// Mouse wheel scrolling - track deltaX and deltaY for wheel event
		wheelX, wheelY := ebiten.Wheel()
		if wheelX != 0 || wheelY != 0 {
			// Create and dispatch wheel event to target element
			target := b.getWheelEventTarget(mx, my)
			if target != nil {
				wheelEvent := b.createWheelEvent(wheelX, wheelY, target)
				// Default action: scroll content unless prevented
				if !wheelEvent.DefaultPrevented {
					if wheelY != 0 {
						b.ScrollY -= int(wheelY * 50)
						if b.ScrollY < 0 {
							b.ScrollY = 0
						}
						if b.ScrollY > b.MaxScrollY {
							b.ScrollY = b.MaxScrollY
						}
					}
					if wheelX != 0 {
						// Horizontal scroll - for now just log, horizontal scrolling is less common
						_ = wheelX // Reserved for future horizontal scroll support
					}
				}
			}
		}

		// Mouse click handling
		mx, my := ebiten.CursorPosition()

		// URL bar click detection
		if my >= 0 && my < navBarHeight {
			// Check for nav button clicks
			if mx >= 10 && mx <= 50 && !b.TypingFocused {
				// Back button
				if b.HistoryIndex > 0 {
					b.HistoryIndex--
					b.navigateTo(b.History[b.HistoryIndex])
				}
			} else if mx >= 55 && mx <= 95 && !b.TypingFocused {
				// Forward button
				if b.HistoryIndex < len(b.History)-1 {
					b.HistoryIndex++
					b.navigateTo(b.History[b.HistoryIndex])
				}
			} else if mx >= b.ViewportW-58 && mx <= b.ViewportW-10 {
				// Go button clicked
				if b.TypingFocused || b.Typing != "" {
					url := b.Typing
					if url != "" {
						b.navigateTo(url)
					}
					b.Typing = ""
					b.TypingFocused = false
				} else {
					// Start typing in URL bar
					b.TypingFocused = true
					b.Typing = b.URLBarText
					b.URLBarFocused = true
				}
			} else if mx >= 100 && mx <= b.ViewportW-60 {
				// URL bar clicked - focus for typing
				b.TypingFocused = true
				b.Typing = b.URLBarText
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
		if b.HoveredLinkIdx >= 0 && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			// Only navigate once per click (when first pressed)
			if !b.linkClicked {
				link := b.Page.Links[b.HoveredLinkIdx]
				b.navigateTo(link.HREF)
				b.linkClicked = true
			}
		} else {
			b.linkClicked = false
		}
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
	if b.TypingFocused {
		urlBarBg = color.RGBA{R: 255, G: 255, B: 220, A: 255}
	}
	drawRect(screen, 155, 8, b.ViewportW-220, 24, urlBarBg)

	// Draw Go button
	goBtnColor := color.RGBA{R: 70, G: 120, B: 200, A: 255}
	drawRect(screen, b.ViewportW-58, 8, 45, 24, goBtnColor)

	// Show what user is typing, or current URL
	displayText := b.Typing
	if !b.TypingFocused && b.Typing == "" {
		displayText = b.URLBarText
	}
	if displayText == "" {
		displayText = "Enter URL..."
	}
	urlTextColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	if !b.TypingFocused && b.Typing == "" && b.URLBarText == "" {
		urlTextColor = color.RGBA{R: 140, G: 140, B: 140, A: 255}
	}
	drawTextAt(screen, displayText, 162, navBarHeight/2-6, urlTextColor)

	// Draw "Go" text on button
	goTextColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	drawTextAt(screen, "Go", b.ViewportW-50, navBarHeight/2-6, goTextColor)

	// Draw page content
	if b.Page != nil && b.Page.Canvas != nil && !b.Loading {
		// Debug: print canvas info
		if b.Page.Canvas.Pixels != nil {
			// Check a few pixel values
			pix := b.Page.Canvas.Pixels
			idx0 := pix.Pix[0]
			idx1 := pix.Pix[1]
			idx2 := pix.Pix[2]
			idxMid := (b.Page.Canvas.Width*100 + 100) * 4
			var midR, midG, midB uint8
			if idxMid < len(pix.Pix) {
				midR = pix.Pix[idxMid]
				midG = pix.Pix[idxMid+1]
				midB = pix.Pix[idxMid+2]
			}
			println(fmt.Sprintf("Canvas: %dx%d, Pixels len: %d, First pixels: %d,%d,%d Mid(%d,%d): %d,%d,%d",
				b.Page.Canvas.Width, b.Page.Canvas.Height, len(pix.Pix),
				idx0, idx1, idx2, 100, 100, midR, midG, midB))
		}

		// Create ebiten image from the canvas pixels
		ebitenImg := ebiten.NewImageFromImage(b.Page.Canvas.Pixels)
		opts := &ebiten.DrawImageOptions{}
		contentGeoM := ebiten.GeoM{}
		contentGeoM.Translate(0, float64(contentOffset))
		contentGeoM.Translate(0, -float64(b.ScrollY))
		opts.GeoM = contentGeoM
		screen.DrawImage(ebitenImg, opts)

		// Highlight hovered link
		if b.HoveredLinkIdx >= 0 && b.HoveredLinkIdx < len(b.Page.Links) {
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
	} else if b.Page != nil && b.Page.Canvas != nil && b.Loading {
		drawTextAt(screen, "Loading page...", 10, contentOffset+20, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		drawTextAt(screen, fmt.Sprintf("Layout: %.0fx%.0f", b.Page.Layout.ContentW, b.Page.Layout.ContentH), 10, contentOffset+50, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	} else if b.ErrorMsg != "" {
		drawTextAt(screen, "Error: "+b.ErrorMsg, 10, contentOffset+20, color.RGBA{R: 200, G: 0, B: 0, A: 255})
	} else {
		drawTextAt(screen, "ViBrowsing - Enter a URL above", 10, contentOffset+20, color.RGBA{R: 0, G: 0, B: 0, A: 255})
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
	// Use ebitenutil.DebugPrintAt for simple text rendering
	// Note: DebugPrintAt doesn't support custom colors, so we draw a background rect
	// with the text color and use white text. For proper color support we'd need freetype.
	// For now just draw the text as-is
	ebitenutil.DebugPrintAt(screen, text, x, y)
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
	b.navURL = url

	// Create channel for result
	b.pageResult = make(chan *PageState, 1)
	b.navDone = make(chan struct{}, 1)

	// Fetch and render in a goroutine
	go func() {
		page, errMsg := b.fetchAndRenderSync(url)
		b.pageResult <- page
		b.ErrorMsg = errMsg
		close(b.navDone)
	}()
}

func (b *Browser) fetchAndRenderSync(rawURL string) (*PageState, string) {
	resp, err := fetch.Fetch(rawURL, "", 10, b.CookieJar, nil)
	if err != nil {
		return nil, fmt.Sprintf("Fetch error: %v", err)
	}

	// Use decompressed body if available (handles gzip compression)
	body := resp.Body
	if resp.Decompressed != nil {
		body = resp.Decompressed
	}

	dom := html.Parse(body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Sprintf("HTTP %d", resp.StatusCode)
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
		return nil, "No body element found"
	}

	layout.LayoutBlock(layoutBox, float64(b.ViewportW))

	canvas := render.NewCanvas(b.ViewportW, b.ViewportH-contentOffset)
	canvas.Clear()
	canvas.DrawBox(layoutBox)

	links := extractLinks(layoutBox, dom, resp.FinalURL)

	page := &PageState{
		URL:    resp.FinalURL,
		DOM:    dom,
		Layout: layoutBox,
		Canvas: canvas,
		Links:  links,
	}

	// Debug: save canvas to file
	if page.Canvas != nil && page.Canvas.Pixels != nil {
		saved := image_to_png(page.Canvas.Pixels, page.Canvas.Width, page.Canvas.Height)
		if saved {
			println(fmt.Sprintf("Canvas DEBUG: saved %dx%d image", page.Canvas.Width, page.Canvas.Height))
		}
	}

	return page, ""
}

// image_to_png saves RGBA pixels as PNG for debugging
func image_to_png(pixels *image.RGBA, width, height int) bool {
	f, err := os.Create(fmt.Sprintf("/tmp/vibrowsing_canvas_%d.png", time.Now().UnixNano()))
	if err != nil {
		return false
	}
	defer f.Close()

	// Create a new RGBA image from the pixel data
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	copy(img.Pix, pixels.Pix)

	err = png.Encode(f, img)
	if err != nil {
		return false
	}
	return true
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

// getWheelEventTarget returns the target element for a wheel event based on mouse position.
// Returns nil if the mouse is outside the content area.
func (b *Browser) getWheelEventTarget(mx, my int) *window.EventTarget {
	// Only dispatch wheel events to content area (below navbar)
	if my < contentOffset {
		return nil
	}
	// Return the document as the target for wheel events
	// In a full implementation, this would find the specific element at position (mx, my)
	return window.NewEventTarget()
}

// createWheelEvent creates a WheelEvent and dispatches it to the target.
// Returns the wheel event (with DefaultPrevented potentially set by handlers).
func (b *Browser) createWheelEvent(deltaX, deltaY float64, target *window.EventTarget) *window.WheelEvent {
	we := window.NewWheelEvent(deltaX, deltaY, 0)
	we.Target = target
	target.DispatchEvent(&we.Event)
	return we
}
