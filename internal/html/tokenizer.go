package html

import (
	"bufio"
	"io"
	"strings"
)

// MaxLineLength is the maximum characters allowed in a single line of text.
// Lines exceeding this are split to prevent memory issues with very large HTML.
const MaxLineLength = 1024 * 1024 // 1MB characters

// MaxTokenLength is the maximum size of a single text token.
// Tokens larger than this are split into multiple tokens.
const MaxTokenLength = 64 * 1024 // 64KB

// TokenType represents the type of an HTML token.
type TokenType int

const (
	TokenDOCTYPE TokenType= iota
	TokenStartTag
	TokenEndTag
	TokenComment
	TokenCharacter
)

// Token represents an HTML token.
type Token struct {
	Type        TokenType
	Data       string
	TagName    string
	Attributes []Attribute
	SelfClosing bool
}

// Attribute represents an HTML attribute.
type Attribute struct {
	Key   string
	Value string
}

// Tokenizer is an HTML5 tokenizer.
// It can operate on byte input (via input field) or read from an io.Reader (via reader field).
// When using ReadFrom, the input field is used as a buffer and reader provides additional data.
type Tokenizer struct {
	input    []rune
	pos      int
	done     bool
	reader   io.Reader
	bufReader *bufio.Reader
	remainingText string // overflow from large text tokens
}

// NewTokenizer creates a new tokenizer from byte input.
func NewTokenizer(input []byte) *Tokenizer {
	return &Tokenizer{input: []rune(string(input)), pos: 0}
}

// NewTokenizerFromReader creates a new tokenizer that reads from an io.Reader.
// The reader will be used to fetch more data as needed.
func NewTokenizerFromReader(r io.Reader) *Tokenizer {
	return &Tokenizer{
		input:    nil,
		pos:      0,
		done:     false,
		reader:   r,
		bufReader: bufio.NewReaderSize(r, 32*1024),
	}
}

// ReadFrom reads data from the reader into the tokenizer's input buffer.
// This allows incremental parsing of large documents.
func (t *Tokenizer) ReadFrom() error {
	if t.reader == nil {
		return nil
	}

	// Read more data from the reader
	reader := t.bufReader
	if reader == nil {
		reader = bufio.NewReaderSize(t.reader, 32*1024)
	}

	// Expand input buffer
	data := make([]byte, 32*1024)
	n, err := reader.Read(data)
	if n > 0 {
		newRunes := []rune(string(data[:n]))
		t.input = append(t.input, newRunes...)
	}
	if err != nil && err != io.EOF {
		return err
	}
	if err == io.EOF {
		t.reader = nil
	}
	return nil
}

// Next returns the next token, or nil at end of input.
func (t *Tokenizer) Next() *Token {
	for !t.done && t.pos < len(t.input) {
		c := t.input[t.pos]

		// Text: accumulate until '<'
		if c != '<' {
			start := t.pos
			for t.pos < len(t.input) && t.input[t.pos] != '<' {
				t.pos++
			}
			text := decodeEntities(string(t.input[start:t.pos]))

			// Prepend any remaining text from previous oversized token
			if t.remainingText != "" {
				text = t.remainingText + text
				t.remainingText = ""
			}

			// Check if text exceeds MaxLineLength and split if necessary
			if len(text) > MaxLineLength {
				return t.emitTextToken(text)
			}

			return &Token{Type: TokenCharacter, Data: text}
		}

		// Check for DOCTYPE first (must be before comment since "<!DOCTYPE" starts with "<!")
		if t.match("<!DOCTYPE") || t.match("<!doctype") {
			t.pos += 9
			// Collect DOCTYPE data (everything between DOCTYPE and >)
			doctypeStart := t.pos
			for t.pos < len(t.input) && t.input[t.pos] != '>' {
				t.pos++
			}
			doctypeData := string(t.input[doctypeStart:t.pos])
			if t.pos < len(t.input) {
				t.pos++ // consume '>'
			}
			// Skip following whitespace
			for t.pos < len(t.input) {
				c := t.input[t.pos]
				if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f' {
					t.pos++
				} else {
					break
				}
			}
			// Skip synthetic root elements: parser bootstraps html/head/body
			// This handles: <!DOCTYPE ...>\n<html> — skip the html that follows DOCTYPE
			if t.match("<html") || t.match("<HTML") {
				t.skipTag()
			}
			if t.match("<head") || t.match("<HEAD") {
				t.skipTag()
			}
			if t.match("<body") || t.match("<BODY") {
				t.skipTag()
			}
			if t.match("</html>") || t.match("</HTML>") {
				t.skipTag()
			}
			if t.match("</head>") || t.match("</HEAD>") {
				t.skipTag()
			}
			if t.match("</body>") || t.match("</BODY>") {
				t.skipTag()
			}
			// Emit DOCTYPE token
			return &Token{Type: TokenDOCTYPE, Data: doctypeData}
		}

		// Comment: <!-- ... --> (must be before general "<!" case)
		if t.match("<!--") {
			t.pos += 4
			start := t.pos
			for t.pos < len(t.input) {
				if t.match("-->") {
					data := string(t.input[start:t.pos])
					t.pos += 3
					return &Token{Type: TokenComment, Data: data}
				}
				t.pos++
			}
			// EOF in comment
			return &Token{Type: TokenComment, Data: string(t.input[start:])}
		}

		// End tag: </name>
		if t.match("</") {
			t.pos += 2
			tagName := t.readTagName()
			// Consume the '>' after the tag name
			if t.pos < len(t.input) && t.input[t.pos] == '>' {
				t.pos++
			}
			// Skip synthetic root end tags: parser bootstraps html/head/body
			switch strings.ToLower(tagName) {
			case "html", "head", "body":
				continue
			}
			return &Token{Type: TokenEndTag, TagName: strings.ToLower(tagName)}
		}

		// Start tag: <name ...>
		if t.match("<") {
			t.pos++
			tagName := t.readTagName()
			tagNameLower := strings.ToLower(tagName)
			attrs := t.readAttributes()
			selfClosing := false
			if t.pos < len(t.input) && t.input[t.pos] == '/' {
				selfClosing = true
				t.pos++
			}
			if t.pos < len(t.input) && t.input[t.pos] == '>' {
				t.pos++
			}
			// Skip synthetic root elements: parser bootstraps html/head/body
			// These tags are created by the parser; don't emit tokens for them
			switch tagNameLower {
			case "html", "head", "body":
				continue
			}
			// Skip script content: don't emit script tag or its content
			// This prevents JS execution (we don't execute JS in this browser)
			if tagNameLower == "script" {
				// Skip until </script> tag
				for t.pos < len(t.input) {
					if t.match("</script>") || t.match("</SCRIPT>") {
						t.pos += 9 // skip </script>
						break
					}
					t.pos++
				}
				continue
			}
			return &Token{
				Type:        TokenStartTag,
				TagName:     tagNameLower,
				Attributes:  attrs,
				SelfClosing: selfClosing,
			}
		}

		// Default: consume the stray '<' character
		t.pos++
	}

	return nil
}

// emitTextToken handles text that exceeds MaxLineLength by splitting it into
// tokens of at most MaxTokenLength characters.
func (t *Tokenizer) emitTextToken(text string) *Token {
	// If text is very long, split it into chunks
	if len(text) > MaxTokenLength {
		// Take first chunk
		chunk := text[:MaxTokenLength]
		// Store remaining for next call
		t.remainingText = text[MaxTokenLength:]
		return &Token{Type: TokenCharacter, Data: chunk}
	}

	// Text is within MaxTokenLength but exceeds MaxLineLength
	// Just return it as-is (the parser will handle truncation)
	return &Token{Type: TokenCharacter, Data: text}
}

// match checks if the input at current position starts with s.
func (t *Tokenizer) match(s string) bool {
	if t.pos >= len(t.input) {
		return false
	}
	input := string(t.input[t.pos:])
	return strings.HasPrefix(input, s)
}

// readTagName reads a tag name starting at the current position.
func (t *Tokenizer) readTagName() string {
	start := t.pos
	for t.pos < len(t.input) {
		c := t.input[t.pos]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f' || c == '>' || c == '/' {
			break
		}
		t.pos++
	}
	return string(t.input[start:t.pos])
}

// readAttributes reads zero or more attributes after a tag name.
func (t *Tokenizer) readAttributes() []Attribute {
	var attrs []Attribute
	for t.pos < len(t.input) {
		c := t.input[t.pos]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f' {
			t.pos++
			continue
		}
		if c == '>' || c == '/' {
			break
		}
		// Attribute name
		nameStart := t.pos
		for t.pos < len(t.input) {
			c := t.input[t.pos]
			if c == '=' || c == ' ' || c == '\n' || c == '\t' || c == '\f' || c == '>' || c == '/' {
				break
			}
			t.pos++
		}
		attr := Attribute{Key: string(t.input[nameStart:t.pos])}
		attrs = append(attrs, attr)

		// Skip whitespace before =
		for t.pos < len(t.input) && (t.input[t.pos] == ' ' || t.input[t.pos] == '\t' || t.input[t.pos] == '\n') {
			t.pos++
		}
		if t.pos < len(t.input) && t.input[t.pos] == '=' {
			t.pos++
		} else {
			continue
		}
		// Skip whitespace after =
		for t.pos < len(t.input) && (t.input[t.pos] == ' ' || t.input[t.pos] == '\t' || t.input[t.pos] == '\n' || t.input[t.pos] == '\r' || t.input[t.pos] == '\f') {
			t.pos++
		}
		// Attribute value
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
					break
				}
			} else {
				if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f' || c == '>' {
					break
				}
			}
			t.pos++
		}
		attrs[len(attrs)-1].Value = string(t.input[valStart:t.pos])
		if quote != 0 && t.pos < len(t.input) && t.input[t.pos] == quote {
			t.pos++ // skip closing quote
		}
		// Skip any trailing whitespace before the next attribute
		for t.pos < len(t.input) && (t.input[t.pos] == ' ' || t.input[t.pos] == '\t' || t.input[t.pos] == '\n' || t.input[t.pos] == '\r' || t.input[t.pos] == '\f') {
			t.pos++
		}
	}
	return attrs
}

// namedEntities maps HTML named entity names to their decoded runes.
var namedEntities = map[string]rune{
	"amp":  '&',
	"lt":   '<',
	"gt":   '>',
	"quot": '"',
	"apos": '\'',
	"nbsp": '\u00A0', // non-breaking space
	"ndash": '\u2013',
	"mdash": '\u2014',
	"lsquo": '\u2018',
	"rsquo": '\u2019',
	"ldquo": '\u201C',
	"rdquo": '\u201D',
	"hellip": '\u2026',
	"copy":  '\u00A9',
	"reg":   '\u00AE',
	"trade": '\u2122',
	"deg":   '\u00B0',
	"plusmn": '\u00B1',
	"times": '\u00D7',
	"divide": '\u00F7',
	"frac12": '\u00BD',
	"frac14": '\u00BC',
	"frac34": '\u00BE',
}

// decodeEntities decodes HTML named and numeric character references in a string.
// Handles: &name; &#nnn; &#xhh;
func decodeEntities(s string) string {
	var result []rune
	runes := []rune(s)
	i := 0
	for i < len(runes) {
		if runes[i] == '&' {
			// Collect the entity name
			start := i + 1
			end := start
			 for end < len(runes) && end < start+10 {
				if runes[end] == ';' {
					break
				}
				end++
			}
			if end < len(runes) && runes[end] == ';' {
				entity := string(runes[start:end])
				if runes[start] == '#' {
					// Numeric reference
					numStr := entity[1:] // remove #
					base := 10
					if len(numStr) > 1 && numStr[0] == 'x' {
						base = 16
						numStr = numStr[1:]
					}
					if val, err := parseUint(numStr, base); err == nil && val > 0 && val < 0x10FFFF {
						result = append(result, rune(val))
						i = end + 1
						continue
					}
				} else {
					// Named entity
					if r, ok := namedEntities[entity]; ok {
						result = append(result, r)
						i = end + 1
						continue
					}
				}
			}
			// Not a valid entity, emit '&' as-is
			result = append(result, '&')
			i++
		} else {
			result = append(result, runes[i])
			i++
		}
	}
	return string(result)
}

func parseUint(s string, base int) (uint64, error) {
	var val uint64
	for i := 0; i < len(s); i++ {
		c := s[i]
		var digit uint64
		switch {
		case c >= '0' && c <= '9':
			digit = uint64(c - '0')
		case c >= 'a' && c <= 'f' && base == 16:
			digit = uint64(c - 'a' + 10)
		case c >= 'A' && c <= 'F' && base == 16:
			digit = uint64(c - 'A' + 10)
		default:
			return 0, nil
		}
		val = val*uint64(base) + digit
	}
	return val, nil
}

// skipTag skips a full tag including attributes and closing >.
// Used when we want to consume a tag without emitting a token.
func (t *Tokenizer) skipTag() {
	// Consume until '>'
	for t.pos < len(t.input) && t.input[t.pos] != '>' {
		t.pos++
	}
	if t.pos < len(t.input) {
		t.pos++ // consume '>'
	}
}

// Tokenize returns all tokens from the input.
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
