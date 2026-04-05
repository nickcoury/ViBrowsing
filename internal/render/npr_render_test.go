package render

import (
    "testing"
    "os"
    "strings"
    
    "image/color"
    "image/png"
    
    "github.com/nickcoury/ViBrowsing/internal/css"
    "github.com/nickcoury/ViBrowsing/internal/fetch"
    "github.com/nickcoury/ViBrowsing/internal/html"
    "github.com/nickcoury/ViBrowsing/internal/layout"
)

func TestFontGlyphDrawing(t *testing.T) {
    canvas := NewCanvas(100, 100)
    canvas.Clear()
    
    // Check if default font is loaded
    t.Logf("Default font: %v", canvas.defaultFont != nil)
    if canvas.defaultFont == nil {
        t.Log("ERROR: No default font loaded!")
        return
    }
    
    // Try drawing a glyph
    t.Log("Attempting to draw glyph 'A' at (50, 50)")
    canvas.DrawGlyph(canvas.defaultFont, 16, 50, 50, color.RGBA{0, 0, 0, 255}, 'A')
    
    // Check pixels around that area
    for dy := -5; dy <= 5; dy++ {
        for dx := -5; dx <= 5; dx++ {
            idx := ((50+dy) * canvas.Width + (50+dx)) * 4
            r := canvas.Pixels.Pix[idx]
            g := canvas.Pixels.Pix[idx+1]
            b := canvas.Pixels.Pix[idx+2]
            a := canvas.Pixels.Pix[idx+3]
            if a != 255 || r != 255 || g != 255 || b != 255 {
                t.Logf("  Pixel at (%d, %d): RGBA=%d,%d,%d,%d", 50+dx, 50+dy, r, g, b, a)
            }
        }
    }
    
    // Also check first pixel
    t.Logf("First pixel after DrawGlyph: RGBA=%d,%d,%d,%d", 
        canvas.Pixels.Pix[0], canvas.Pixels.Pix[1], canvas.Pixels.Pix[2], canvas.Pixels.Pix[3])
}

func TestNPRStyleDebug(t *testing.T) {
    resp, err := fetch.Fetch("https://text.npr.org", "", 10, nil, nil)
    if err != nil {
        t.Fatalf("Fetch error: %v", err)
    }
    
    body := resp.Body
    if resp.Decompressed != nil {
        body = resp.Decompressed
    }
    
    dom := html.Parse(body)
    
    styleNodes := dom.QuerySelectorAll("style")
    t.Logf("Style nodes found: %d", len(styleNodes))
    
    for i, n := range styleNodes {
        t.Logf("Style node %d: TagName=%q, Type=%d, Children=%d", i, n.TagName, n.Type, len(n.Children))
        
        var directText string
        for _, c := range n.Children {
            if c.Type == html.NodeText {
                directText += c.Data
            }
        }
        t.Logf("  Direct text: %d chars, first 200: %q", len(directText), directText[:200])
        
        if strings.Contains(directText, "body") {
            t.Logf("  Contains 'body' selector - CSS FOUND")
        }
    }
}

func TestNPRPageRenderDebug(t *testing.T) {
    resp, err := fetch.Fetch("https://text.npr.org", "", 10, nil, nil)
    if err != nil {
        t.Fatalf("Fetch error: %v", err)
    }
    
    body := resp.Body
    if resp.Decompressed != nil {
        body = resp.Decompressed
    }
    t.Logf("Body len: %d (decompressed=%v)", len(body), resp.Decompressed != nil)
    t.Logf("ContentType: %s", resp.ContentType)
    t.Logf("First 300 chars: %q", string(body[:300]))
    
    dom := html.Parse(body)
    
    // Count nodes
    totalNodes := 0
    textNodes := 0
    var walk func(n *html.Node)
    walk = func(n *html.Node) {
        totalNodes++
        if n.Type == html.NodeText {
            textNodes++
        }
        for _, c := range n.Children {
            walk(c)
        }
    }
    for _, c := range dom.Children {
        walk(c)
    }
    t.Logf("Total nodes: %d, Text nodes: %d", totalNodes, textNodes)
    
    // Collect CSS - USE DIRECT TEXT ACCESS
    var cssRules []css.Rule
    for _, node := range dom.QuerySelectorAll("style") {
        var cssText string
        for _, c := range node.Children {
            if c.Type == html.NodeText {
                cssText += c.Data
            }
        }
        t.Logf("Found style with %d chars of CSS", len(cssText))
        if cssText != "" {
            rules := css.Parse(cssText)
            t.Logf("  Parsed %d CSS rules", len(rules))
            cssRules = append(cssRules, rules...)
        }
    }
    t.Logf("Total CSS rules: %d", len(cssRules))
    
    // Build layout
    layoutBox := layout.BuildLayoutTree(dom, cssRules, 1024, 732)
    if layoutBox == nil {
        t.Fatal("BuildLayoutTree returned nil")
    }
    
    // Count boxes before layout
    totalBoxes := 0
    textBoxes := 0
    var walkBoxes func(b *layout.Box)
    walkBoxes = func(b *layout.Box) {
        totalBoxes++
        if b.Type == layout.TextBox {
            textBoxes++
        }
        for _, c := range b.Children {
            walkBoxes(c)
        }
    }
    walkBoxes(layoutBox)
    t.Logf("Before layout - boxes: %d, textBoxes: %d", totalBoxes, textBoxes)
    t.Logf("Root box: Type=%d, ContentW=%.0f, ContentH=%.0f", layoutBox.Type, layoutBox.ContentW, layoutBox.ContentH)
    
    // Run layout
    layout.LayoutBlock(layoutBox, 1024)
    t.Logf("After layout - ContentH=%.0f", layoutBox.ContentH)
    
    // Create canvas
    canvas := NewCanvas(1024, 732)
    canvas.Clear()
    t.Logf("Canvas created, first pixels before DrawBox: RGBA=%d,%d,%d,%d", 
        canvas.Pixels.Pix[0], canvas.Pixels.Pix[1], canvas.Pixels.Pix[2], canvas.Pixels.Pix[3])
    
    // Draw
    canvas.DrawBox(layoutBox)
    
    // Check pixels after draw
    t.Logf("First pixels after DrawBox: RGBA=%d,%d,%d,%d", 
        canvas.Pixels.Pix[0], canvas.Pixels.Pix[1], canvas.Pixels.Pix[2], canvas.Pixels.Pix[3])
    
    // Sample some pixels from content area
    t.Log("Pixel samples from content area:")
    for _, y := range []int{50, 100, 200, 300, 400} {
        idx := (y * canvas.Width + 100) * 4
        t.Logf("  (%d, %d): RGBA=%d,%d,%d,%d", 100, y, 
            canvas.Pixels.Pix[idx], canvas.Pixels.Pix[idx+1], 
            canvas.Pixels.Pix[idx+2], canvas.Pixels.Pix[idx+3])
    }
    
    // Save to file
    f, err := os.Create("/tmp/npr_canvas_debug.png")
    if err != nil {
        t.Fatalf("Failed to create: %v", err)
    }
    err = png.Encode(f, canvas.Pixels)
    f.Close()
    if err != nil {
        t.Fatalf("Failed to encode: %v", err)
    }
    t.Logf("Saved to /tmp/npr_canvas_debug.png")
}

func TestGlyphDetails(t *testing.T) {
    canvas := NewCanvas(200, 100)
    canvas.Clear()
    
    if canvas.defaultFont == nil {
        t.Fatal("No default font!")
    }
    
    f := canvas.defaultFont
    fontSize := 16.0
    x, y := 50.0, 50.0
    ch := 'A'
    
    // Use same approach as DrawGlyph
    face := truetype.NewFace(f, &truetype.Options{
        Size:    fontSize,
        DPI:     72,
        Hinting: 1, // font.HintingFull
    })
    
    dot := fixed.Point26_6{
        X: fixed.Int26_6(x * 64),
        Y: fixed.Int26_6(y * 64),
    }
    
    dr, mask, maskp, adv, ok := face.Glyph(dot, ch)
    t.Logf("Glyph 'A': dr=%v, mask bounds=%v, maskp=%v, adv=%v, ok=%v", dr, mask.Bounds(), maskp, adv, ok)
    
    if !ok {
        t.Log("ERROR: Glyph not found!")
    } else {
        // Check if mask has any content
        t.Logf("  maskpix at origin: %v", maskPix(mask, 0, 0))
        
        // Try drawing with direct pixel manipulation
        tmp := image.NewRGBA(dr)
        draw.Draw(tmp, dr, &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.ZP, draw.Src)
        
        dstRect := dr.Sub(dr.Min).Add(image.Point{
            X: int(x) + maskp.X - dr.Min.X,
            Y: int(y) + maskp.Y - dr.Min.Y,
        })
        t.Logf("  dstRect: %v", dstRect)
        
        // Check if dstRect is valid
        if dstRect.Min.X >= 0 && dstRect.Min.Y >= 0 && dstRect.Max.X <= canvas.Width && dstRect.Max.Y <= canvas.Height {
            t.Log("  dstRect is within canvas bounds")
            
            // Draw mask
            srcRect := tmp.Bounds()
            draw.DrawMask(canvas.Pixels, dstRect, tmp, srcRect.Min, mask, maskp, draw.Over)
            
            t.Logf("After draw: first pixel RGBA=%d,%d,%d,%d", 
                canvas.Pixels.Pix[0], canvas.Pixels.Pix[1], canvas.Pixels.Pix[2], canvas.Pixels.Pix[3])
            
            // Check some pixels in dstRect area
            for py := dstRect.Min.Y; py < dstRect.Max.Y; py++ {
                for px := dstRect.Min.X; px < dstRect.Max.X; px++ {
                    if px >= 0 && px < canvas.Width && py >= 0 && py < canvas.Height {
                        idx := (py * canvas.Width + px) * 4
                        if canvas.Pixels.Pix[idx] != 255 || canvas.Pixels.Pix[idx+1] != 255 || canvas.Pixels.Pix[idx+2] != 255 {
                            t.Logf("  Modified pixel at (%d, %d): RGBA=%d,%d,%d,%d", px, py,
                                canvas.Pixels.Pix[idx], canvas.Pixels.Pix[idx+1], canvas.Pixels.Pix[idx+2], canvas.Pixels.Pix[idx+3])
                        }
                    }
                }
            }
        } else {
            t.Logf("  dstRect is OUT OF BOUNDS: %v (canvas: %dx%d)", dstRect, canvas.Width, canvas.Height)
        }
    }
}

func maskPix(m image.Image, x, y int) byte {
    // Simple check if pixel at x,y is set in mask
    r, _, _, _ := m.At(x, y).RGBA()
    return byte(r >> 8)
}
