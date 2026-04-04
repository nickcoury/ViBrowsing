package html

import (
	"fmt"
	"os"
	"testing"
)

func TestParserDebug(t *testing.T) {
	data, err := os.ReadFile("../../sample_pages/test1.html")
	if err != nil {
		t.Fatal(err)
	}

	dom := Parse(data)
	fmt.Printf("DOM:\n%s\n", dom.String())
}

func TestForeignContentSVG(t *testing.T) {
	// SVG inside HTML should not trigger foster parenting or implicit <p> close
	doc := Parse([]byte(`<html><body><p>Hello<svg><rect><g><text>Inside SVG</text></g></rect></svg>World</p></body></html>`))
	body := doc.QuerySelectorAll("body")[0]
	text := body.InnerText()
	// Should have "Hello Inside SVG World" with no odd paragraph breaks
	if text != "HelloInside SVGWorld" {
		t.Errorf("unexpected text in SVG context: %q", text)
	}
}

func TestForeignContentMath(t *testing.T) {
	// MathML inside HTML should not trigger foster parenting
	doc := Parse([]byte(`<html><body><p>Formula:<math><mi>x</mi></math>End</p></body></html>`))
	body := doc.QuerySelectorAll("body")[0]
	text := body.InnerText()
	if text != "Formula:xEnd" {
		t.Errorf("unexpected text in MathML context: %q", text)
	}
}

func TestForeignContentNested(t *testing.T) {
	// Nested foreign content
	doc := Parse([]byte(`<html><body><svg><foreignObject><p>This p should not close early</p></foreignObject></svg></body></html>`))
	// The "This p should not close early" should all be inside foreignObject
	body := doc.QuerySelectorAll("body")[0]
	svg := body.QuerySelectorAll("svg")[0]
	if len(svg.Children) == 0 {
		t.Error("svg should have children")
	}
}
