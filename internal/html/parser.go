package html

import "strings"

// voidElements are self-closing by nature in HTML.
var voidElements = map[string]bool{
	"br": true, "hr": true, "img": true, "input": true,
	"meta": true, "link": true, "area": true, "base": true,
	"col": true, "embed": true, "param": true,
	"source": true, "track": true, "wbr": true,
}

// Parser builds a DOM tree from HTML tokens.
type Parser struct {
	tokens []*Token
	pos    int
}

// NewParser creates a new parser from tokens.
func NewParser(tokens []*Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

// Parse builds and returns the DOM tree.
func (p *Parser) Parse() *Node {
	doc := NewDocument()
	p.pos = 0

	// Bootstrap: synthetic html, head, body containers
	htmlNode := NewElement("html")
	headNode := NewElement("head")
	bodyNode := NewElement("body")
	htmlNode.AppendChild(headNode)
	htmlNode.AppendChild(bodyNode)
	doc.AppendChild(htmlNode)

	// Stack: current open elements
	openStack := []*Node{htmlNode}

	for p.pos < len(p.tokens) {
		token := p.pp()
		if token == nil {
			break
		}

		switch token.Type {
		case TokenDOCTYPE:
			// Ignored
			p.advance()

		case TokenStartTag:
			tagName := strings.ToLower(token.TagName)

			// Create new element
			node := NewElement(tagName)
			for _, attr := range token.Attributes {
				node.Attributes = append(node.Attributes, attr)
			}

			// Append to parent
			if len(openStack) > 0 {
				openStack[len(openStack)-1].AppendChild(node)
			}

			// Void elements are leaf nodes
			if !voidElements[tagName] {
				openStack = append(openStack, node)
			}
			p.advance()

		case TokenEndTag:
			tagName := strings.ToLower(token.TagName)

			// Pop until we find the matching tag
			for len(openStack) > 0 {
				top := openStack[len(openStack)-1]
				if strings.ToLower(top.TagName) == tagName {
					openStack = openStack[:len(openStack)-1]
					break
				}
				// Tag not on stack: malformed HTML, skip it
				openStack = openStack[:len(openStack)-1]
			}
			p.advance()

		case TokenCharacter:
			text := strings.TrimSpace(token.Data)
			if text != "" && len(openStack) > 0 {
				openStack[len(openStack)-1].AppendChild(NewText(text))
			}
			p.advance()

		case TokenComment:
			// Ignored
			p.advance()

		default:
			p.advance()
		}
	}

	return doc
}

func (p *Parser) pp() *Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return nil
}

func (p *Parser) advance() {
	p.pos++
}

// Parse is a convenience function: tokenize then parse.
func Parse(input []byte) *Node {
	tokens := Tokenize(input)
	parser := NewParser(tokens)
	return parser.Parse()
}
