// +build ignore

package main

import (
	"fmt"
	"os"

	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/html"
	"github.com/nickcoury/ViBrowsing/internal/layout"
	"github.com/nickcoury/ViBrowsing/internal/render"
)

func main() {
	// Test 1: Minimal HTML - check body/h1 positioning
	testHTML := `<!DOCTYPE html>
<html>
<head><title>Test 1: Minimal</title></head>
<body style="margin: 20px;">
<h1 style="margin: 0;">Hello</h1>
<p style="margin: 0;">World</p>
</body>
</html>`

	fmt.Println("=== Test 1: Minimal (body margin:20px) ===")
	dom1 := html.Parse([]byte(testHTML))
	cssRules1 := css.Parse("")
	styleNodes1 := dom1.QuerySelectorAll("style")
	for _, node := range styleNodes1 {
		sheet := node.InnerText()
		if sheet != "" {
			cssRules1 = append(cssRules1, css.Parse(sheet)...)
		}
	}
	box1 := layout.BuildLayoutTree(dom1, cssRules1)
	if box1 != nil {
		layout.LayoutBlock(box1, 800)
		printTree(box1, 0)
	}

	// Test 2: With title/style elements in body (check head/body separation)
	testHTML2 := `<!DOCTYPE html>
<html>
<head><title>Test 2</title></head>
<body style="margin: 20px; background: #f0f0f0;">
<h1 style="margin: 0;">Hello World</h1>
<p style="margin: 0;">Some text here.</p>
</body>
</html>`

	fmt.Println("\n=== Test 2: Basic page layout ===")
	dom2 := html.Parse([]byte(testHTML2))
	cssRules2 := css.Parse("")
	box2 := layout.BuildLayoutTree(dom2, cssRules2)
	if box2 != nil {
		layout.LayoutBlock(box2, 800)
		printTree(box2, 0)
	}

	// Test 3: Render to PNG
	fmt.Println("\n=== Rendering test1 to PNG ===")
	canvas := render.NewCanvas(800, 600)
	canvas.Clear()
	if box1 != nil {
		canvas.DrawBox(box1)
	}
	if err := canvas.SavePNG("/home/nick/Repos/nickcoury/ViBrowsing/test1_output.png"); err != nil {
		fmt.Printf("Save error: %v\n", err)
	} else {
		fmt.Println("Saved test1_output.png")
	}

	fmt.Println("\n=== Rendering test2 to PNG ===")
	canvas2 := render.NewCanvas(800, 600)
	canvas2.Clear()
	if box2 != nil {
		canvas2.DrawBox(box2)
	}
	if err := canvas2.SavePNG("/home/nick/Repos/nickcoury/ViBrowsing/test2_output.png"); err != nil {
		fmt.Printf("Save error: %v\n", err)
	} else {
		fmt.Println("Saved test2_output.png")
	}
}

func printTree(b *layout.Box, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	tagName := "(root)"
	if b.Node != nil {
		tagName = b.Node.TagName
	}

	boxType := "?"
	switch b.Type {
	case layout.BlockBox:
		boxType = "block"
	case layout.InlineBox:
		boxType = "inline"
	case layout.TextBox:
		boxType = "text"
	}

	marginTop := b.Style["margin-top"]
	marginLeft := b.Style["margin-left"]

	fmt.Printf("%s[%s] %s margin-top=%s margin-left=%s ContentX=%.0f ContentY=%.0f ContentW=%.0f ContentH=%.0f\n",
		prefix, boxType, tagName, marginTop, marginLeft, b.ContentX, b.ContentY, b.ContentW, b.ContentH)

	for _, c := range b.Children {
		printTree(c, indent+1)
	}
}

var _ = os.Stdin