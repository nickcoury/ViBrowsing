package main

import (
	"fmt"
	"os"

	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/html"
	"github.com/nickcoury/ViBrowsing/internal/layout"
)

func main() {
	data, err := os.ReadFile("sample_pages/test1.html")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	dom := html.Parse(data)
	cssRules := css.Parse("")
	
	styleNodes := dom.QuerySelectorAll("style")
	for _, node := range styleNodes {
		sheet := node.InnerText()
		if sheet != "" {
			cssRules = append(cssRules, css.Parse(sheet)...)
		}
	}
	
	layoutBox := layout.BuildLayoutTree(dom, cssRules)
	if layoutBox == nil {
		fmt.Println("No layout tree")
		return
	}
	
	// Run layout pass
	layout.LayoutBlock(layoutBox, 800)
	
	printTree(layoutBox, 0)
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
	
	disp := b.Style["display"]
	marginTop := b.Style["margin-top"]
	marginLeft := b.Style["margin-left"]
	
	fmt.Printf("%s[%s] %s display=%s margin-top=%s margin-left=%s ContentX=%.0f ContentY=%.0f ContentW=%.0f ContentH=%.0f\n",
		prefix, boxType, tagName, disp, marginTop, marginLeft, b.ContentX, b.ContentY, b.ContentW, b.ContentH)
	
	for _, c := range b.Children {
		printTree(c, indent+1)
	}
}
