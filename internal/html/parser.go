package html

import "strings"

// voidElements are self-closing by nature in HTML.
var voidElements = map[string]bool{
	"br": true, "hr": true, "img": true, "input": true,
	"meta": true, "link": true, "area": true, "base": true,
	"col": true, "embed": true, "param": true,
	"source": true, "track": true, "wbr": true,
}

// tableFosterTags are tags that trigger foster parenting when inside a table
var tableFosterTags = map[string]bool{
	"table": true, "tbody": true, "thead": true, "tfoot": true, "tr": true,
}

// blockTags are block-level elements that implicitly close open <p> tags
var blockTags = map[string]bool{
	"div": true, "p": true, "ul": true, "ol": true, "li": true,
	"table": true, "tr": true, "td": true, "th": true, "thead": true,
	"tbody": true, "tfoot": true, "h1": true, "h2": true, "h3": true,
	"h4": true, "h5": true, "h6": true, "form": true, "header": true,
	"footer": true, "nav": true, "section": true, "article": true,
	"aside": true, "main": true, "figure": true, "figcaption": true,
}

// tableRelatedTags are table structure tags
var tableRelatedTags = map[string]bool{
	"table": true, "tbody": true, "thead": true, "tfoot": true, "tr": true,
	"td": true, "th": true, "caption": true,
}

// Parser builds a DOM tree from HTML tokens.
type Parser struct {
	tokens []*Token
	pos    int

	// openStack is the list of currently open elements
	openStack []*Node

	// fosterParenting tracks whether we're inside a table
	fosterParenting int // count of open table-related elements
}

// NewParser creates a new parser from tokens.
func NewParser(tokens []*Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

// Parse builds and returns the DOM tree.
func (p *Parser) Parse() *Node {
	doc := NewDocument()
	p.pos = 0
	p.openStack = nil
	p.fosterParenting = 0

	// Bootstrap: synthetic html, head, body containers
	htmlNode := NewElement("html")
	headNode := NewElement("head")
	bodyNode := NewElement("body")
	htmlNode.AppendChild(headNode)
	htmlNode.AppendChild(bodyNode)
	doc.AppendChild(htmlNode)

	// Stack: current open elements
	p.openStack = []*Node{htmlNode}

	for p.pos < len(p.tokens) {
		token := p.pp()
		if token == nil {
			break
		}

		switch token.Type {
		case TokenDOCTYPE:
			p.advance()

		case TokenStartTag:
			tagName := strings.ToLower(token.TagName)

			// Skip synthetic bootstrap tags
			if tagName == "html" || tagName == "head" || tagName == "body" {
				// Use the synthetic node if it exists as a child
				parent := p.openStack[len(p.openStack)-1]
				if child := parent.FindChildByTagName(tagName); child != nil {
					p.openStack = append(p.openStack, child)
				}
				p.advance()
				continue
			}

			// Create element node
			node := NewElement(tagName)
			for _, attr := range token.Attributes {
				node.Attributes = append(node.Attributes, attr)
			}

			// Implicit <p> close before block elements
			if blockTags[tagName] && len(p.openStack) > 0 && p.currentIs("p") {
				p.popTo("p")
			}

			// Foster parenting: insert at table's foster parent instead of inside table
			if p.fosterParenting > 0 && !tableRelatedTags[tagName] {
				// Find the nearest table ancestor
				fosterIdx := p.findFosterParentIndex()
				if fosterIdx >= 0 {
					// Insert before the table element
					parent := p.openStack[fosterIdx-1]
					if parent == nil && fosterIdx > 0 {
						parent = p.openStack[fosterIdx-1]
					}
					// Actually insert before table in current parent
					tableNode := p.openStack[fosterIdx]
					tableParent := tableNode.Parent
					if tableParent != nil {
						// Find table's position and insert before it
						for i, c := range tableParent.Children {
							if c == tableNode {
								// Insert before table
								children := tableParent.Children
								tableParent.Children = append(children[:i:i], append([]*Node{node}, children[i:]...)...)
								// Don't push to stack (node is not "open")
								p.advance()
								continue
							}
						}
					}
				}
			}

			// Append to parent
			if len(p.openStack) > 0 {
				p.openStack[len(p.openStack)-1].AppendChild(node)
			}

			// Track table context
			if tableRelatedTags[tagName] {
				p.fosterParenting++
			}

			// Void elements don't go on the stack
			if !voidElements[tagName] {
				p.openStack = append(p.openStack, node)
			}

			p.advance()

		case TokenEndTag:
			tagName := strings.ToLower(token.TagName)

			// Ignore synthetic end tags
			if tagName == "html" || tagName == "head" || tagName == "body" {
				p.advance()
				continue
			}

			// Foster parenting end tag: if closing a table, flush fosters
			if tableRelatedTags[tagName] {
				if p.fosterParenting > 0 {
					p.fosterParenting--
				}
				// Pop until we find the matching tag
				for len(p.openStack) > 0 {
					top := p.openStack[len(p.openStack)-1]
					if top.TagName == tagName {
						p.openStack = p.openStack[:len(p.openStack)-1]
						break
					}
					p.openStack = p.openStack[:len(p.openStack)-1]
				}
				p.advance()
				continue
			}

			// Generic end tag: pop until we find matching tag
			// If not found, skip (malformed HTML)
			for len(p.openStack) > 0 {
				top := p.openStack[len(p.openStack)-1]
				if top.TagName == tagName {
					p.openStack = p.openStack[:len(p.openStack)-1]
					break
				}
				// Pop this element — it's not closed
				p.openStack = p.openStack[:len(p.openStack)-1]
			}
			// Unknown end tag: skip it (don't crash on malformed HTML)
			p.advance()

		case TokenCharacter:
			text := strings.TrimSpace(token.Data)
			if text != "" {
				// Foster parenting: text inside table context goes to foster parent
				if p.fosterParenting > 0 {
					fosterIdx := p.findFosterParentIndex()
					if fosterIdx > 0 {
						fosterParent := p.openStack[fosterIdx-1]
						fosterParent.AppendChild(NewText(text))
					} else if len(p.openStack) > 0 {
						p.openStack[len(p.openStack)-1].AppendChild(NewText(text))
					}
				} else if len(p.openStack) > 0 {
					p.openStack[len(p.openStack)-1].AppendChild(NewText(text))
				}
			}
			p.advance()

		case TokenComment:
			p.advance()

		default:
			p.advance()
		}
	}

	return doc
}

// findFosterParentIndex returns the index of the nearest table in openStack.
// Returns -1 if no table found.
func (p *Parser) findFosterParentIndex() int {
	for i := len(p.openStack) - 1; i >= 0; i-- {
		if p.openStack[i].TagName == "table" {
			return i
		}
	}
	return -1
}

// currentIs returns true if the top of the open stack has the given tag name.
func (p *Parser) currentIs(tagName string) bool {
	if len(p.openStack) == 0 {
		return false
	}
	return p.openStack[len(p.openStack)-1].TagName == tagName
}

// popTo pops elements until the named tag is at the top of the stack.
// Returns true if the tag was found and popped to.
func (p *Parser) popTo(tagName string) bool {
	for len(p.openStack) > 0 {
		if p.openStack[len(p.openStack)-1].TagName == tagName {
			p.openStack = p.openStack[:len(p.openStack)-1]
			return true
		}
		p.openStack = p.openStack[:len(p.openStack)-1]
	}
	return false
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
