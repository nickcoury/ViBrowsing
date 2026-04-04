//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/nickcoury/ViBrowsing/internal/html"
)

func main() {
	data, err := os.ReadFile("sample_pages/test1.html")
	if err != nil {
		panic(err)
	}

	tokens := html.Tokenize(data)
	for i, tok := range tokens {
		if tok.Type == 0 {
			fmt.Printf("ZERO TOKEN at %d!\n", i)
			break
		}
		fmt.Printf("%3d: type=%d tag=%q data=%q self=%v attrs=%d\n",
			i, tok.Type, tok.TagName, tok.Data, tok.SelfClosing, len(tok.Attributes))
		if i > 40 {
			fmt.Printf("... (stopping at %d tokens)\n", i+1)
			break
		}
	}
}
