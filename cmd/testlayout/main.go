package main

import (
    "fmt"
    "github.com/nickcoury/ViBrowsing/internal/css"
    "github.com/nickcoury/ViBrowsing/internal/html"
    "github.com/nickcoury/ViBrowsing/internal/layout"
)

func main() {
    htmlData := []byte(`<!DOCTYPE html>
<html>
<head><style>body{margin:0}</style></head>
<body>
<h1 style="color:red">Hello World</h1>
<p style="color:blue">This is a test.</p>
</body>
</html>`)

    doc := html.Parse(htmlData)
    cssRules := css.Parse("")

    box := layout.BuildLayoutTree(doc, cssRules, 800, 600)

    // Monkey-patch: add debug to the layout by wrapping
    // Just run layout and print
    layout.LayoutBlock(box, 800)

    fmt.Printf("\nFull box tree:\n")
    printTree(box, 0)
}

func printTree(b *layout.Box, indent int) {
    prefix := ""
    for i := 0; i < indent; i++ {
        prefix += "  "
    }
    tag := "(root)"
    if b.Node != nil { tag = b.Node.TagName }
    typ := "?"
    if b.Type == 0 { typ = "block" }
    if b.Type == 1 { typ = "inline" }
    if b.Type == 2 { typ = "text" }
    fmt.Printf("%s[%s] %s: ContentX=%.0f ContentY=%.0f ContentW=%.0f ContentH=%.0f\n",
        prefix, typ, tag, b.ContentX, b.ContentY, b.ContentW, b.ContentH)
    for _, c := range b.Children {
        printTree(c, indent+1)
    }
}
