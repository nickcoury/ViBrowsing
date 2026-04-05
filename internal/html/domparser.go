package html

import (
	"strings"
)

// DOMParser parses HTML/XML strings into DOM documents.
// Implements the DOMParser interface: https://developer.mozilla.org/en-US/docs/Web/API/DOMParser
type DOMParser struct {
	// document is the parsed document (set after parsing)
	document *Node
}

// NewDOMParser creates a new DOMParser instance.
func NewDOMParser() *DOMParser {
	return &DOMParser{}
}

// ParseFromString parses a string and returns a Document.
// Supported mimeTypes:
//   - "text/html": Parses the string as HTML, returns an HTML document
//   - "image/svg+xml": Parses the string as SVG
//   - "text/xml": Parses the string as XML
//   - "application/xml": Parses the string as XML
//   - "application/xhtml+xml": Parses the string as XHTML
func (dp *DOMParser) ParseFromString(str string, mimeType string) *Node {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))

	switch mimeType {
	case "text/html":
		return dp.parseHTML(str)
	case "image/svg+xml", "image/svg":
		return dp.parseXML(str)
	case "text/xml":
		return dp.parseXML(str)
	case "application/xml":
		return dp.parseXML(str)
	case "application/xhtml+xml":
		return dp.parseXML(str)
	default:
		// Default to HTML parsing
		return dp.parseHTML(str)
	}
}

// parseHTML parses HTML string into a Document.
func (dp *DOMParser) parseHTML(str string) *Node {
	// Use the existing HTML parser
	doc := Parse([]byte(str))
	dp.document = doc
	return doc
}

// parseXML parses XML/SVG/XHTML string into a Document.
func (dp *DOMParser) parseXML(str string) *Node {
	// For XML, we need a different approach since the HTML parser
	// is lenient. For now, we'll use a basic XML-like parsing approach.
	doc := NewDocument()
	dp.document = doc

	// Simple XML parsing - handle root element and children
	str = strings.TrimSpace(str)
	if strings.HasPrefix(str, "<?xml") {
		// Skip XML declaration
		if idx := strings.Index(str, "?>"); idx != -1 {
			str = strings.TrimSpace(str[idx+2:])
		}
	}

	// Find root element
	str = strings.TrimSpace(str)
	if strings.HasPrefix(str, "<") {
		// Extract root element
		root, _ := dp.parseXMLElement(str)
		if root != nil {
			doc.AppendChild(root)
		}
	}

	return doc
}

// parseXMLElement parses a single XML element from the string.
// Returns the remaining string after parsing.
func (dp *DOMParser) parseXMLElement(s string) (node *Node, remaining string) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "<") {
		return nil, s
	}

	// Find element name
	nameEnd := 0
	for nameEnd < len(s) {
		c := s[nameEnd]
		if c == ' ' || c == '>' || c == '/' || c == '\n' || c == '\t' {
			break
		}
		nameEnd++
	}

	if nameEnd == 0 {
		return nil, s
	}

	tagName := s[1:nameEnd]
	element := NewElement(tagName)

	// Parse attributes
	attrsStr := s[nameEnd:]
	attrsStr = strings.TrimSpace(attrsStr)

	for len(attrsStr) > 0 && !strings.HasPrefix(attrsStr, ">") && !strings.HasPrefix(attrsStr, "/>") {
		// Find next attribute
		eqIdx := strings.Index(attrsStr, "=")
		if eqIdx == -1 {
			break
		}

		attrName := strings.TrimSpace(attrsStr[:eqIdx])
		attrsStr = attrsStr[eqIdx+1:]

		// Get quoted value
		if len(attrsStr) == 0 {
			break
		}

		var attrValue string
		quote := attrsStr[0]
		if quote == '"' || quote == '\'' {
			// Find closing quote
			closeIdx := strings.Index(attrsStr[1:], string(quote))
			if closeIdx == -1 {
				break
			}
			attrValue = attrsStr[1 : closeIdx+1]
			attrsStr = attrsStr[closeIdx+2:]
		} else {
			// Unquoted value
			spaceIdx := strings.IndexAny(attrsStr, " \t\n>")
			if spaceIdx == -1 {
				attrValue = attrsStr
				attrsStr = ""
			} else {
				attrValue = attrsStr[:spaceIdx]
				attrsStr = attrsStr[spaceIdx:]
			}
		}

		element.SetAttribute(attrName, attrValue)
		attrsStr = strings.TrimSpace(attrsStr)
	}

	// Handle self-closing tags
	if strings.HasPrefix(attrsStr, "/>") {
		return element, attrsStr[2:]
	}

	// Skip to closing bracket
	if strings.HasPrefix(attrsStr, ">") {
		attrsStr = attrsStr[1:]
	}

	// Check for empty element (no children)
	if strings.HasPrefix(attrsStr, "</") {
		return element, attrsStr
	}

	// Parse content until closing tag
	content := ""
	for len(attrsStr) > 0 {
		if strings.HasPrefix(attrsStr, "</") {
			break
		}
		if strings.HasPrefix(attrsStr, "<") {
			// Child element
			child, rem := dp.parseXMLElement(attrsStr)
			if child != nil {
				element.AppendChild(child)
			}
			attrsStr = rem
		} else {
			// Text content
			ltIdx := strings.Index(attrsStr, "<")
			if ltIdx == -1 {
				content += attrsStr
				break
			}
			content += attrsStr[:ltIdx]
			attrsStr = attrsStr[ltIdx:]
		}
	}

	// Trim and add text content
	content = strings.TrimSpace(content)
	if content != "" {
		text := NewText(content)
		element.AppendChild(text)
	}

	// Skip closing tag
	if strings.HasPrefix(attrsStr, "</") {
		closeIdx := strings.Index(attrsStr, ">")
		if closeIdx != -1 {
			attrsStr = attrsStr[closeIdx+1:]
		}
	}

	return element, attrsStr
}

// GetDocument returns the last parsed document.
func (dp *DOMParser) GetDocument() *Node {
	return dp.document
}
