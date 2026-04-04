package html

import (
	"fmt"
	"os"
	"testing"
)

func TestTokenizerDebug(t *testing.T) {
	data, err := os.ReadFile("../../sample_pages/test1.html")
	if err != nil {
		t.Fatal(err)
	}

	tokens := Tokenize(data)
	fmt.Printf("Total tokens: %d\n", len(tokens))
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

func TestEntityDecoding(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"&amp;", "&"},
		{"&lt;", "<"},
		{"&gt;", ">"},
		{"&quot;", "\""},
		{"&apos;", "'"},
		{"&nbsp;", "\u00A0"},
		{"&#65;", "A"},
		{"&#x41;", "A"},
		{"&#x3C;", "<"},
		{"&mdash;", "\u2014"},
		{"&ndash;", "\u2013"},
		{"&hellip;", "\u2026"},
		{"no entities here", "no entities here"},
		{"&amp; &lt; &gt;", "& < >"},
		{"&unknown;", "&unknown;"},
		{"&#;", "&#;"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			tokens := Tokenize([]byte(tc.input))
			var text string
			for _, tok := range tokens {
				if tok.Type == TokenCharacter {
					text += tok.Data
				}
			}
			if text != tc.expected {
				t.Errorf("input=%q got=%q want=%q", tc.input, text, tc.expected)
			}
		})
	}
}
