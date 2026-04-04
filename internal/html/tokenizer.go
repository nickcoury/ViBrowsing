package html

import "strings"

// TokenType represents the type of an HTML token.
type TokenType int

const (
	TokenDOCTYPE TokenType = iota
	TokenStartTag
	TokenEndTag
	TokenComment
	TokenCharacter
)

// Token represents an HTML token.
type Token struct {
	Type       TokenType
	Data      string
	TagName   string
	Attributes []Attribute
	SelfClosing bool
}

// Attribute represents an HTML attribute.
type Attribute struct {
	Key   string
	Value string
}

// Tokenizer is an HTML5 tokenizer (simplified).
type Tokenizer struct {
	input       []rune
	pos         int
	current     Token
	charBuff    string
}

// NewTokenizer creates a new tokenizer.
func NewTokenizer(input []byte) *Tokenizer {
	return &Tokenizer{input: []rune(string(input)), pos: 0}
}

// Next returns the next token, or nil at end of input.
func (t *Tokenizer) Next() *Token {
	// If there's buffered character data, emit it first
	if t.charBuff != "" {
		data := t.charBuff
		t.charBuff = ""
		return &Token{Type: TokenCharacter, Data: data}
	}

	for t.pos < len(t.input) {
		t.step()
		// After each step, flush buffered character data
		if t.charBuff != "" {
			data := t.charBuff
			t.charBuff = ""
			return &Token{Type: TokenCharacter, Data: data}
		}
		if t.current.Type != 0 {
			token := t.current
			t.current = Token{}
			return &token
		}
	}

	// At end of input: flush remaining buffer
	if t.charBuff != "" {
		data := t.charBuff
		t.charBuff = ""
		return &Token{Type: TokenCharacter, Data: data}
	}

	return nil
}

// step advances the state machine by one step.
func (t *Tokenizer) step() {
	if t.pos >= len(t.input) {
		return
	}
	c := t.input[t.pos]

	switch {
	// dataState
	case t.current.Type == 0 && t.current.TagName == "" && !strings.HasPrefix(string(t.input[t.pos:]), "<"):
		if c == '<' || c == 0 {
			return
		}
		t.charBuff += string(c)
		t.pos++

	case strings.HasPrefix(string(t.input[t.pos:]), "<!"):
		t.pos += 2
		t.handleMarkupDeclaration()

	case strings.HasPrefix(string(t.input[t.pos:]), "</"):
		t.pos += 2
		t.current = Token{Type: TokenEndTag}
		t.tagName()

	case strings.HasPrefix(string(t.input[t.pos:]), "<"):
		t.pos++
		t.current = Token{Type: TokenStartTag}
		t.tagName()

	default:
		t.pos++
	}
}

func (t *Tokenizer) handleMarkupDeclaration() {
	remaining := string(t.input[t.pos:])
	if strings.HasPrefix(remaining, "DOCTYPE") || strings.HasPrefix(remaining, "doctype") {
		t.pos += 8 // len("DOCTYPE")
		// Consume until '>'
		for t.pos < len(t.input) && t.input[t.pos] != '>' {
			t.pos++
		}
		if t.pos < len(t.input) {
			t.pos++ // consume '>'
		}
		t.current = Token{Type: TokenDOCTYPE, TagName: "html"}
		return
	}
	if strings.HasPrefix(remaining, "--") {
		t.pos += 2
		t.charBuff = ""
		// Consume until -->
		end := strings.Index(string(t.input[t.pos:]), "-->")
		if end >= 0 {
			t.current = Token{Type: TokenComment, Data: string(t.input[t.pos : t.pos+end])}
			t.pos += end + 3
		} else {
			t.current = Token{Type: TokenComment, Data: string(t.input[t.pos:])}
			t.pos = len(t.input)
		}
		return
	}
	// Bogus comment
	end := strings.Index(string(t.input[t.pos:]), ">")
	if end >= 0 {
		t.current = Token{Type: TokenComment, Data: string(t.input[t.pos : t.pos+end])}
		t.pos += end + 1
	} else {
		t.current = Token{Type: TokenComment, Data: string(t.input[t.pos:])}
		t.pos = len(t.input)
	}
}

func (t *Tokenizer) tagName() {
	start := t.pos
	for t.pos < len(t.input) {
		c := t.input[t.pos]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f' {
			break
		}
		if c == '>' || c == '/' {
			break
		}
		t.pos++
	}
	t.current.TagName = strings.ToLower(string(t.input[start:t.pos]))

	// Self-closing: <tag ... />
	if t.pos < len(t.input) && t.input[t.pos] == '/' {
		t.current.SelfClosing = true
		t.pos++
	}

	// End of tag
	if t.pos < len(t.input) && t.input[t.pos] == '>' {
		t.pos++
		return
	}

	// Parse attributes
	if t.pos < len(t.input) && t.input[t.pos] != '>' {
		t.attrs()
	}

	// Consume '>'
	if t.pos < len(t.input) && t.input[t.pos] == '>' {
		t.pos++
	}
}

func (t *Tokenizer) attrs() {
	for t.pos < len(t.input) {
		c := t.input[t.pos]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f' {
			t.pos++
			continue
		}
		if c == '>' || c == '/' {
			return
		}
		// Attribute name
		start := t.pos
		for t.pos < len(t.input) {
			c := t.input[t.pos]
			if c == '=' || c == ' ' || c == '\n' || c == '\t' || c == '\f' || c == '>' || c == '/' {
				break
			}
			t.pos++
		}
		attr := Attribute{Key: string(t.input[start:t.pos])}
		t.current.Attributes = append(t.current.Attributes, attr)

		// Maybe =
		for t.pos < len(t.input) && t.input[t.pos] == ' ' || t.pos < len(t.input) && t.input[t.pos] == '\t' || t.pos < len(t.input) && t.input[t.pos] == '\n' {
			t.pos++
		}
		if t.pos < len(t.input) && t.input[t.pos] == '=' {
			t.pos++
			// Attribute value
			for t.pos < len(t.input) && (t.input[t.pos] == ' ' || t.input[t.pos] == '\t' || t.input[t.pos] == '\n' || t.input[t.pos] == '\r' || t.input[t.pos] == '\f') {
				t.pos++
			}
			var quote rune
			if t.pos < len(t.input) && (t.input[t.pos] == '"' || t.input[t.pos] == '\'') {
				quote = t.input[t.pos]
				t.pos++
			}
			valStart := t.pos
			for t.pos < len(t.input) {
				c := t.input[t.pos]
				if quote != 0 {
					if c == quote {
						t.pos++
						break
					}
				} else {
					if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f' || c == '>' {
						break
					}
				}
				t.pos++
			}
			t.current.Attributes[len(t.current.Attributes)-1].Value = string(t.input[valStart:t.pos])
			if quote != 0 && t.pos < len(t.input) {
				t.pos++ // consume closing quote
			}
		}
	}
}

// Tokenize is a convenience function that tokenizes input and returns all tokens.
func Tokenize(input []byte) []*Token {
	t := NewTokenizer(input)
	var tokens []*Token
	for {
		token := t.Next()
		if token == nil {
			break
		}
		tokens = append(tokens, token)
	}
	return tokens
}
