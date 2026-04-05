package html

import (
	"testing"
)

func TestNewDOMParser(t *testing.T) {
	dp := NewDOMParser()
	if dp == nil {
		t.Fatal("NewDOMParser returned nil")
	}
}

func TestDOMParserParseFromStringHTML(t *testing.T) {
	dp := NewDOMParser()

	html := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body><p>Hello World</p></body>
</html>`

	doc := dp.ParseFromString(html, "text/html")
	if doc == nil {
		t.Fatal("ParseFromString returned nil")
	}
	if doc.Type != NodeDocument {
		t.Errorf("Root should be NodeDocument, got %v", doc.Type)
	}

	// Check that body was parsed
	body := doc.QuerySelector("body")
	if body == nil {
		t.Error("body element not found")
	}

	p := doc.QuerySelector("p")
	if p == nil {
		t.Error("p element not found")
	}

	if len(p.Children) > 0 && p.Children[0].Type == NodeText {
		if p.Children[0].Data != "Hello World" {
			t.Errorf("p content = %q, want %q", p.Children[0].Data, "Hello World")
		}
	}
}

func TestDOMParserParseFromStringSimpleHTML(t *testing.T) {
	dp := NewDOMParser()

	doc := dp.ParseFromString("<div>Test</div>", "text/html")
	if doc == nil {
		t.Fatal("ParseFromString returned nil")
	}

	div := doc.QuerySelector("div")
	if div == nil {
		t.Fatal("div element not found")
	}

	if len(div.Children) > 0 && div.Children[0].Type == NodeText {
		if div.Children[0].Data != "Test" {
			t.Errorf("div content = %q, want %q", div.Children[0].Data, "Test")
		}
	}
}

func TestDOMParserParseFromStringSVG(t *testing.T) {
	dp := NewDOMParser()

	svg := `<svg width="100" height="100">
  <circle cx="50" cy="50" r="40" fill="blue"/>
</svg>`

	doc := dp.ParseFromString(svg, "image/svg+xml")
	if doc == nil {
		t.Fatal("ParseFromString returned nil")
	}

	// SVG should be parsed as XML
	svgEl := doc.QuerySelector("svg")
	if svgEl == nil {
		t.Error("svg element not found")
	}

	circle := doc.QuerySelector("circle")
	if circle == nil {
		t.Error("circle element not found")
	}

	if circle.GetAttribute("fill") != "blue" {
		t.Errorf("circle fill = %q, want %q", circle.GetAttribute("fill"), "blue")
	}
}

func TestDOMParserParseFromStringXML(t *testing.T) {
	dp := NewDOMParser()

	xml := `<?xml version="1.0"?>
<root>
  <child attr="value">text content</child>
</root>`

	doc := dp.ParseFromString(xml, "text/xml")
	if doc == nil {
		t.Fatal("ParseFromString returned nil")
	}

	root := doc.QuerySelector("root")
	if root == nil {
		t.Error("root element not found")
	}
}

func TestDOMParserParseFromStringXMLNoDeclaration(t *testing.T) {
	dp := NewDOMParser()

	xml := `<note>
  <to>Tove</to>
  <from>Jani</from>
  <heading>Reminder</heading>
  <body>Don't forget me this weekend!</body>
</note>`

	doc := dp.ParseFromString(xml, "application/xml")
	if doc == nil {
		t.Fatal("ParseFromString returned nil")
	}

	note := doc.QuerySelector("note")
	if note == nil {
		t.Error("note element not found")
	}
}

func TestDOMParserParseFromStringXHTML(t *testing.T) {
	dp := NewDOMParser()

	xhtml := `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>XHTML Test</title></head>
<body><p>Content</p></body>
</html>`

	doc := dp.ParseFromString(xhtml, "application/xhtml+xml")
	if doc == nil {
		t.Fatal("ParseFromString returned nil")
	}

	title := doc.QuerySelector("title")
	if title == nil {
		t.Error("title element not found")
	}
}

func TestDOMParserParseFromStringUnknownMimeType(t *testing.T) {
	dp := NewDOMParser()

	// Unknown mime types should default to HTML parsing
	doc := dp.ParseFromString("<span>Test</span>", "application/unknown")
	if doc == nil {
		t.Fatal("ParseFromString returned nil for unknown mime type")
	}

	span := doc.QuerySelector("span")
	if span == nil {
		t.Error("span element not found - should default to HTML parsing")
	}
}

func TestDOMParserGetDocument(t *testing.T) {
	dp := NewDOMParser()

	doc1 := dp.ParseFromString("<p>First</p>", "text/html")
	if dp.GetDocument() != doc1 {
		t.Error("GetDocument should return the last parsed document")
	}

	doc2 := dp.ParseFromString("<p>Second</p>", "text/html")
	if dp.GetDocument() != doc2 {
		t.Error("GetDocument should return the new document after second parse")
	}
}

func TestDOMParserParseFromStringCaseInsensitiveMimeType(t *testing.T) {
	dp := NewDOMParser()

	doc := dp.ParseFromString("<div>Test</div>", "TEXT/HTML")
	if doc == nil {
		t.Fatal("ParseFromString returned nil for uppercase mime type")
	}

	div := doc.QuerySelector("div")
	if div == nil {
		t.Error("div element not found with uppercase mime type")
	}
}

func TestDOMParserParseFromStringEmpty(t *testing.T) {
	dp := NewDOMParser()

	doc := dp.ParseFromString("", "text/html")
	if doc == nil {
		t.Fatal("ParseFromString should not return nil for empty string")
	}
	// Empty HTML should produce a valid document structure
	if doc.Type != NodeDocument {
		t.Errorf("Root should be NodeDocument, got %v", doc.Type)
	}
}
