package css

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"
)

// Assert that Color implements color.Color at compile time.
var _ color.Color = Color{}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A uint8
}

// RGBA implements color.Color (8-bit per channel expanded to 16-bit big-endian per channel).
func (c Color) RGBA() (r, g, b, a uint32) {
	return uint32(c.R)<<8 | uint32(c.R),
		uint32(c.G)<<8 | uint32(c.G),
		uint32(c.B)<<8 | uint32(c.B),
		uint32(c.A)<<8 | uint32(c.A)
}

// LengthUnit represents a CSS length unit.
type LengthUnit int

const (
	UnitPx LengthUnit = iota
	UnitEm
	UnitRem
	UnitPercent
	UnitAuto
)

// Length represents a CSS length value.
type Length struct {
	Value  float64
	Unit   LengthUnit
	IsAuto bool
}

// ParseLength parses a length string like "10px", "1.5em", "50%".
func ParseLength(s string) Length {
	s = strings.TrimSpace(s)
	if s == "auto" {
		return Length{IsAuto: true}
	}

	var value float64
	var unitStr string

	// Parse number
	numStr := ""
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '+' {
			numStr += string(c)
		} else {
			unitStr = strings.TrimSpace(s[i:])
			break
		}
	}

	if numStr == "" {
		return Length{}
	}

	value, _ = strconv.ParseFloat(numStr, 64)

	switch strings.ToLower(unitStr) {
	case "px":
		return Length{Value: value, Unit: UnitPx}
	case "em":
		return Length{Value: value, Unit: UnitEm}
	case "rem":
		return Length{Value: value, Unit: UnitRem}
	case "%":
		return Length{Value: value, Unit: UnitPercent}
	default:
		return Length{Value: value, Unit: UnitPx}
	}
}

// ParseCalc parses a CSS calc() expression like "100% - 20px" or "50% + 10px".
// Returns the computed pixel value, using resolvePercent to resolve percentage values.
func ParseCalc(s string, resolvePercent func(unit string) float64) (float64, error) {
	// Strip "calc(" wrapper if present
	orig := strings.TrimSpace(s)
	s = orig
	// Reject inputs that look like function calls but aren't calc()
	// These include min(), max(), notcalc(), etc.
	if strings.HasPrefix(s, "min(") || strings.HasPrefix(s, "max(") {
		return 0, fmt.Errorf("not a calc() expression")
	}
	// Reject any input that has a function-call style "(" but doesn't start with "calc("
	if strings.Contains(s, "(") && !strings.HasPrefix(s, "calc(") {
		return 0, fmt.Errorf("not a calc() expression")
	}
	if strings.HasPrefix(s, "calc(") {
		// calc() requires the full "calc(...)" form with closing paren
		if !strings.HasSuffix(s, ")") {
			return 0, fmt.Errorf("calc expression missing closing paren")
		}
		s = s[5:] // skip "calc("
		s = s[:len(s)-1] // strip trailing ")"
		// Check for unbalanced parens in the inner expression
		depth := 0
		for i := 0; i < len(s); i++ {
			if s[i] == '(' {
				depth++
			} else if s[i] == ')' {
				depth--
				if depth < 0 {
					return 0, fmt.Errorf("unbalanced parentheses in calc expression")
				}
			}
		}
		if depth > 0 {
			return 0, fmt.Errorf("unclosed parentheses in calc expression")
		}
	}
	if s == orig && !strings.HasPrefix(orig, "calc(") {
		// If no calc() wrapper and no other prefix, try evaluating directly
		// but first check if it looks like it has unbalanced parens
		depth := 0
		for i := 0; i < len(s); i++ {
			if s[i] == '(' {
				depth++
			} else if s[i] == ')' {
				depth--
				if depth < 0 {
					return 0, fmt.Errorf("unbalanced parentheses")
				}
			}
		}
		if depth > 0 {
			return 0, fmt.Errorf("unclosed parentheses")
		}
	}
	s = strings.TrimSpace(s)
	return evaluateCalc(s, resolvePercent)
}

func evaluateCalc(s string, resolvePercent func(unit string) float64) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty calc expression")
	}

	// Find the lowest-precedence operator at the top level
	// + and - have lower precedence than * and /
	// We scan from left to right, tracking parentheses depth
	parenDepth := 0
	var ops []int
	var opChars []byte

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '(' {
			parenDepth++
		} else if c == ')' {
			parenDepth--
			if parenDepth < 0 {
				// Unbalanced closing paren
				return 0, fmt.Errorf("unbalanced parentheses in calc expression")
			}
		} else if parenDepth == 0 && (c == '+' || c == '-' || c == '*' || c == '/') {
			// Skip +/- that are unary signs, not binary operators.
			// A +/- is binary only if preceded by an operand (number, ), %, etc.)
			// e.g., "10px-5px" → binary (operand before -)
			//       "10px + -5px" → + is binary (operand before space), - is unary
			if c == '+' || c == '-' {
				isBinary := false
				// Scan left past whitespace to find if preceded by operand
				for j := i - 1; j >= 0; j-- {
					prev := s[j]
					if prev == ' ' || prev == '\t' {
						continue // skip whitespace
					}
					// Found non-whitespace
					if (prev >= 'a' && prev <= 'z') || (prev >= 'A' && prev <= 'Z') ||
						(prev >= '0' && prev <= '9') || prev == ')' || prev == '%' || prev == '"' || prev == '\'' {
						isBinary = true
					}
					break
				}
				if !isBinary {
					continue
				}
			}
			ops = append(ops, i)
			opChars = append(opChars, c)
		}
	}

	// Check for unclosed opening paren
	if parenDepth > 0 {
		return 0, fmt.Errorf("unbalanced parentheses in calc expression")
	}

	if len(ops) == 0 {
		// No top-level operators — parse as a single value
		s = strings.TrimSpace(s)
		// Check for parentheses grouping
		if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
			return evaluateCalc(s[1:len(s)-1], resolvePercent)
		}
		// Parse as a length
		l := ParseLength(s)
		if l.Unit == UnitPercent {
			return resolvePercent("%"), nil
		}
		return l.Value, nil
	}

	// Find last lowest-precedence operator (* and / bind tighter)
	lastOpIdx := ops[0]
	lastOp := opChars[0]
	isHighPrec := lastOp == '*' || lastOp == '/'

	for i := 1; i < len(ops); i++ {
		op := opChars[i]
		isHighPrecCur := op == '*' || op == '/'
		if !isHighPrec && isHighPrecCur {
			continue
		}
		if (!isHighPrec && !isHighPrecCur) || (isHighPrec && isHighPrecCur) {
			lastOpIdx = ops[i]
			lastOp = op
			isHighPrec = isHighPrecCur
		}
	}

	leftStr := strings.TrimSpace(s[:lastOpIdx])
	rightStr := strings.TrimSpace(s[lastOpIdx+1:])

	// Handle unary signs on operands: if operand starts with +/-, preserve it
	// as part of the value (ParseLength handles negative numbers correctly).
	// We only strip a leading sign if it's NOT at the start of the operand
	// (i.e., it's already a binary operator, not a unary sign).

	leftVal, err := evaluateCalc(leftStr, resolvePercent)
	if err != nil {
		return 0, err
	}
	rightVal, err := evaluateCalc(rightStr, resolvePercent)
	if err != nil {
		return 0, err
	}

	switch lastOp {
	case '+':
		return leftVal + rightVal, nil
	case '-':
		return leftVal - rightVal, nil
	case '*':
		return leftVal * rightVal, nil
	case '/':
		if rightVal == 0 {
			return 0, nil
		}
		return leftVal / rightVal, nil
	}
	return 0, nil
}

// BorderRadius represents CSS border-radius (1-4 values).
type BorderRadius struct {
	TopLeft, TopRight, BottomRight, BottomLeft Length
}

// ParseBorderRadius parses "10px", "10px 5px", "10px 5px 10px", "10px 5px 10px 5px".
func ParseBorderRadius(s string) BorderRadius {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return BorderRadius{}
	}
	if len(parts) == 1 {
		v := ParseLength(parts[0])
		return BorderRadius{v, v, v, v}
	}
	if len(parts) == 2 {
		v1 := ParseLength(parts[0])
		v2 := ParseLength(parts[1])
		return BorderRadius{v1, v2, v1, v2}
	}
	if len(parts) == 3 {
		v1 := ParseLength(parts[0])
		v2 := ParseLength(parts[1])
		v3 := ParseLength(parts[2])
		return BorderRadius{v1, v2, v3, v2}
	}
	// 4 values
	return BorderRadius{
		ParseLength(parts[0]),
		ParseLength(parts[1]),
		ParseLength(parts[2]),
		ParseLength(parts[3]),
	}
}

// ParseAspectRatio parses a CSS aspect-ratio value like "16/9" or "4/3".
// Returns the ratio as width/height (e.g., 16/9 = 1.777...).
// Returns 0 if the string is "auto" or invalid.
func ParseAspectRatio(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "auto" {
		return 0
	}

	// Handle ratio format: "width/height"
	if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		if len(parts) == 2 {
			width, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			height, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err1 == nil && err2 == nil && height != 0 {
				return width / height
			}
		}
		return 0
	}

	// Single number (could be a unitless ratio like "2" meaning 2/1)
	val, err := strconv.ParseFloat(s, 64)
	if err == nil && val != 0 {
		return val
	}

	return 0
}

// ParseColor parses a CSS color string.
func ParseColor(s string) Color {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	// Named colors (subset)
	named := map[string]Color{
		"black":       {0, 0, 0, 255},
		"white":       {255, 255, 255, 255},
		"red":         {255, 0, 0, 255},
		"green":       {0, 128, 0, 255},
		"blue":        {0, 0, 255, 255},
		"yellow":      {255, 255, 0, 255},
		"cyan":        {0, 255, 255, 255},
		"magenta":     {255, 0, 255, 255},
		"gray":        {128, 128, 128, 255},
		"grey":        {128, 128, 128, 255},
		"orange":      {255, 165, 0, 255},
		"purple":      {128, 0, 128, 255},
		"pink":        {255, 192, 203, 255},
		"brown":       {165, 42, 42, 255},
		"navy":        {0, 0, 128, 255},
		"teal":        {0, 128, 128, 255},
		"olive":       {128, 128, 0, 255},
		"maroon":      {128, 0, 0, 255},
		"silver":      {192, 192, 192, 255},
		"lime":        {0, 255, 0, 255},
		"aqua":        {0, 255, 255, 255},
		"fuchsia":     {255, 0, 255, 255},
		"transparent": {0, 0, 0, 0},
		"inherit":     {0, 0, 0, 255},
	}

	if c, ok := named[s]; ok {
		return c
	}

	// #RGB
	if len(s) == 4 && s[0] == '#' {
		r := parseHex(string(s[1]))
		g := parseHex(string(s[2]))
		b := parseHex(string(s[3]))
		return Color{r, g, b, 255}
	}

	// #RRGGBB
	if len(s) == 7 && s[0] == '#' {
		r := parseHex(s[1:3])
		g := parseHex(s[3:5])
		b := parseHex(s[5:7])
		return Color{r, g, b, 255}
	}

	// rgb(r, g, b)
	if strings.HasPrefix(s, "rgb(") && strings.HasSuffix(s, ")") {
		inner := s[4 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 3 {
			r := parseUint8(parts[0])
			g := parseUint8(parts[1])
			b := parseUint8(parts[2])
			return Color{r, g, b, 255}
		}
	}

	// hsl(h, s%, l%)
	if strings.HasPrefix(s, "hsl(") && strings.HasSuffix(s, ")") {
		inner := s[4 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 3 {
			r, g, b := hslToRgb(parts[0], parts[1], parts[2])
			return Color{r, g, b, 255}
		}
	}

	// hsla(h, s%, l%, a)
	if strings.HasPrefix(s, "hsla(") && strings.HasSuffix(s, ")") {
		inner := s[5 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 4 {
			r, g, b := hslToRgb(parts[0], parts[1], parts[2])
			a := parseFloatAlpha(parts[3])
			return Color{r, g, b, a}
		}
	}

	// rgba(r, g, b, a)
	if strings.HasPrefix(s, "rgba(") && strings.HasSuffix(s, ")") {
		inner := s[5 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 4 {
			r := parseUint8(parts[0])
			g := parseUint8(parts[1])
			b := parseUint8(parts[2])
			a := parseFloatAlpha(parts[3])
			return Color{r, g, b, a}
		}
	}

	return Color{0, 0, 0, 255} // default black
}

func parseHex(s string) uint8 {
	v, _ := strconv.ParseUint(s, 16, 8)
	return uint8(v)
}

func parseUint8(s string) uint8 {
	s = strings.TrimSpace(s)
	v, _ := strconv.ParseFloat(s, 64)
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	return uint8(v)
}

func parseFloat255(s string) uint8 {
	s = strings.TrimSpace(s)
	v, _ := strconv.ParseFloat(s, 64)
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	return uint8(v)
}

// hslToRgb converts HSL values (h in degrees, s and l in percentages) to RGB.
func hslToRgb(hStr, sStr, lStr string) (r, g, b uint8) {
	h, _ := strconv.ParseFloat(strings.TrimSpace(hStr), 64)
	sPct, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(sStr, "%")), 64)
	lPct, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(lStr, "%")), 64)
	s := sPct / 100
	l := lPct / 100

	// Normalize hue to [0, 360)
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	if s == 0 {
		v := uint8(l * 255)
		return v, v, v
	}

	var p, q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p = 2*l - q
	hh := h / 360

	hue2 := func(n float64) float64 {
		if n < 0 {
			n++
		}
		if n < 1.0/6 {
			return p + (q-p)*6*n
		}
		if n < 1.0/2 {
			return q
		}
		if n < 2.0/3 {
			return p + (q-p)*(2.0/3-n)*6
		}
		return p
	}

	rr := hue2(hh + 1.0/3)
	gg := hue2(hh)
	bb := hue2(hh - 1.0/3)

	if rr < 0 {
		rr = 0
	}
	if rr > 1 {
		rr = 1
	}
	if gg < 0 {
		gg = 0
	}
	if gg > 1 {
		gg = 1
	}
	if bb < 0 {
		bb = 0
	}
	if bb > 1 {
		bb = 1
	}

	return uint8(rr * 255), uint8(gg * 255), uint8(bb * 255)
}

func parseFloatAlpha(s string) uint8 {
	s = strings.TrimSpace(s)
	// Handle percentage: "50%" -> 0.5
	if strings.HasSuffix(s, "%") {
		pct, _ := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}
		return uint8(pct * 255 / 100)
	}
	// Handle 0-1 float
	v, _ := strconv.ParseFloat(s, 64)
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return uint8(v * 255)
}

// GradientStop represents a color stop in a CSS gradient.
type GradientStop struct {
	Color  Color
	Offset float64 // 0-1 percentage
}

// Transform represents a CSS 2D transform.
type Transform struct {
	Type     string    // "rotate", "scale", "translate", "none"
	Rotate   float64   // rotation in degrees
	ScaleX   float64   // scale X factor (1.0 = no scale)
	ScaleY   float64   // scale Y factor (1.0 = no scale)
	TranslateX float64 // translate X in pixels
	TranslateY float64 // translate Y in pixels
}

// ParseTransform parses a CSS transform value.
// Supports: rotate(<angle>), scale(<x>, <y>?), translate(<x>, <y>)
// Examples: rotate(45deg), scale(1.5), translate(10px, 20px)
func ParseTransform(s string) Transform {
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return Transform{Type: "none"}
	}

	t := Transform{Type: "rotate", Rotate: 0, ScaleX: 1, ScaleY: 1}

	// Handle rotate
	if strings.HasPrefix(s, "rotate(") && strings.HasSuffix(s, ")") {
		inner := s[7 : len(s)-1]
		inner = strings.TrimSpace(inner)
		if angle, err := strconv.ParseFloat(strings.TrimSuffix(inner, "deg"), 64); err == nil {
			t.Type = "rotate"
			t.Rotate = angle
			return t
		}
		// Try without "deg" suffix (just a number)
		if angle, err := strconv.ParseFloat(inner, 64); err == nil {
			t.Type = "rotate"
			t.Rotate = angle
			return t
		}
	}

	// Handle scale
	if strings.HasPrefix(s, "scale(") && strings.HasSuffix(s, ")") {
		inner := s[6 : len(s)-1]
		parts := strings.Split(inner, ",")
		t.Type = "scale"
		if len(parts) == 1 {
			v, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			t.ScaleX = v
			t.ScaleY = v
		} else if len(parts) >= 2 {
			vx, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			vy, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			t.ScaleX = vx
			t.ScaleY = vy
		}
		return t
	}

	// Handle translate
	if strings.HasPrefix(s, "translate(") && strings.HasSuffix(s, ")") {
		inner := s[10 : len(s)-1]
		parts := strings.Split(inner, ",")
		t.Type = "translate"
		if len(parts) >= 1 {
			vx := ParseLength(strings.TrimSpace(parts[0]))
			t.TranslateX = vx.Value
		}
		if len(parts) >= 2 {
			vy := ParseLength(strings.TrimSpace(parts[1]))
			t.TranslateY = vy.Value
		}
		return t
	}

	return t
}

// ParseLinearGradient parses a linear-gradient value.
// Format: linear-gradient(direction, color-stop1, color-stop2, ...)
func ParseLinearGradient(s string) ([]GradientStop, string, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "linear-gradient(") || !strings.HasSuffix(s, ")") {
		return nil, "", fmt.Errorf("not a linear-gradient")
	}
	inner := s[16 : len(s)-1] // Remove "linear-gradient(" and ")"
	
	parts := splitGradientParts(inner)
	if len(parts) < 2 {
		return nil, "", fmt.Errorf("not enough gradient stops")
	}
	
	// Parse direction (first part)
	direction := "to bottom"
	firstPart := strings.TrimSpace(parts[0])
	if firstPart != "" && !strings.Contains(firstPart, "(") {
		// Check for angle (e.g., "45deg") or direction keyword
		if strings.HasSuffix(firstPart, "deg") {
			direction = firstPart // Use as-is for now
		} else if strings.HasPrefix(firstPart, "to ") {
			direction = firstPart
		} else {
			// Default to "to bottom" if first part looks like a color
			// Direction is optional, so if first part is a color, use default
			if !isColorString(firstPart) {
				direction = firstPart
			}
		}
	}
	
	// Parse color stops
	var stops []GradientStop
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// Skip direction part if present
		if i == 0 && (strings.HasPrefix(part, "to ") || strings.HasSuffix(part, "deg") || isDirectionKeyword(part)) {
			continue
		}
		stop := parseGradientStop(part)
		stops = append(stops, stop)
	}
	
	return stops, direction, nil
}

func isDirectionKeyword(s string) bool {
	dirs := []string{"top", "bottom", "left", "right", "center"}
	for _, d := range dirs {
		if s == d {
			return true
		}
	}
	return false
}

func isColorString(s string) bool {
	// Check if string starts with a color indicator
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "#") || strings.HasPrefix(s, "rgb") || strings.HasPrefix(s, "hsl") {
		return true
	}
	named := []string{"black", "white", "red", "green", "blue", "yellow", "cyan", "magenta", "gray", "orange", "purple", "pink", "brown", "navy", "teal", "olive", "maroon", "silver", "lime", "aqua", "fuchsia", "transparent"}
	for _, c := range named {
		if s == c {
			return true
		}
	}
	return false
}

func splitGradientParts(inner string) []string {
	var parts []string
	var current []byte
	depth := 0
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		if c == '(' {
			depth++
			current = append(current, c)
		} else if c == ')' {
			depth--
			current = append(current, c)
		} else if c == ',' && depth == 0 {
			parts = append(parts, string(current))
			current = nil
		} else {
			current = append(current, c)
		}
	}
	if len(current) > 0 {
		parts = append(parts, string(current))
	}
	return parts
}

func parseGradientStop(part string) GradientStop {
	part = strings.TrimSpace(part)
	
	// Check for offset (e.g., "red 50%" or "0deg 50%")
	parts := strings.Fields(part)
	if len(parts) == 2 && strings.HasSuffix(parts[1], "%") {
		offset, _ := strconv.ParseFloat(strings.TrimSuffix(parts[1], "%"), 64)
		offset = offset / 100
		return GradientStop{Color: ParseColor(parts[0]), Offset: offset}
	}
	
	return GradientStop{Color: ParseColor(part), Offset: -1} // -1 means auto
}

// CalcExpression represents a parsed calc() expression.
type CalcExpression struct {
	Op       string // "+", "-", "*", "/"
	Left     interface{} // Length or *CalcExpression
	Right    interface{} // Length or *CalcExpression
	IsMin    bool
	IsMax    bool
	MinValues []Length // for min()/max() with multiple args
}

func parseCalcExpression(s string) (*CalcExpression, error) {
	s = strings.TrimSpace(s)
	depth := 0
	var ops []int
	var parenDepth []int

	// Find operators at the top level
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '(' {
			depth++
		} else if c == ')' {
			depth--
		} else if depth == 0 && (c == '+' || c == '-' || c == '*' || c == '/') {
			// Skip if it's a sign not an operator (e.g., negative number)
			if (c == '+' || c == '-') && i > 0 && !isOpChar(s[i-1]) && s[i-1] != ' ' && s[i-1] != '(' {
				continue
			}
			ops = append(ops, i)
			parenDepth = append(parenDepth, depth)
		}
	}

	if len(ops) == 0 {
		// No top-level operator, try to parse as a Length
		return parseLengthFromExpression(s)
	}

	// Find the last lowest-precedence operator at top level
	// * and / have higher precedence than + and -
	// We process left-to-right for now
	lastOpIdx := ops[0]
	for i := 1; i < len(ops); i++ {
		op := s[ops[i]]
		lastOp := s[lastOpIdx]
		if (op == '+' || op == '-') && (lastOp == '*' || lastOp == '/') {
			lastOpIdx = ops[i]
		}
	}

	op := s[lastOpIdx]
	left := strings.TrimSpace(s[:lastOpIdx])
	right := strings.TrimSpace(s[lastOpIdx+1:])

	leftExpr, err := parseLengthFromExpression(left)
	if err != nil {
		return nil, err
	}
	rightExpr, err := parseLengthFromExpression(right)
	if err != nil {
		return nil, err
	}

	return &CalcExpression{
		Op:    string(op),
		Left:  leftExpr,
		Right: rightExpr,
	}, nil
}

func isOpChar(c byte) bool {
	return c == '+' || c == '-' || c == '*' || c == '/'
}

func parseLengthFromExpression(s string) (*CalcExpression, error) {
	s = strings.TrimSpace(s)

	// Handle nested calc
	if strings.HasPrefix(s, "calc(") {
		return parseCalcExpression(s[5 : len(s)-1])
	}

	// Handle parenthesized expressions
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return parseCalcExpression(s[1 : len(s)-1])
	}

	// Parse as a Length
	length := ParseLength(s)
	if length.IsAuto {
		return nil, fmt.Errorf("auto not allowed in calc")
	}
	return &CalcExpression{Left: &length, Op: string('+'), Right: nil}, nil
}

// EvaluateCalc evaluates a CalcExpression to a Length.
// Note: This is a simplified evaluation that uses UnitPx for all results.
// A full implementation would need context (viewport size, font size, etc.)
func EvaluateCalc(expr *CalcExpression) Length {
	if expr == nil {
		return Length{}
	}

	if expr.IsMin {
		minVal := EvaluateCalcValue(expr.MinValues[0])
		for _, v := range expr.MinValues[1:] {
			candidate := EvaluateCalcValue(v)
			if candidate < minVal {
				minVal = candidate
			}
		}
		return Length{Value: minVal, Unit: UnitPx}
	}

	if expr.IsMax {
		maxVal := EvaluateCalcValue(expr.MinValues[0])
		for _, v := range expr.MinValues[1:] {
			candidate := EvaluateCalcValue(v)
			if candidate > maxVal {
				maxVal = candidate
			}
		}
		return Length{Value: maxVal, Unit: UnitPx}
	}

	leftVal := EvaluateCalcValue(expr.Left)
	rightVal := EvaluateCalcValue(expr.Right)

	var result float64
	switch expr.Op {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	case "*":
		result = leftVal * rightVal
	case "/":
		if rightVal == 0 {
			result = 0
		} else {
			result = leftVal / rightVal
		}
	}

	return Length{Value: result, Unit: UnitPx}
}

// EvaluateCalcValue evaluates a calc operand to a float pixel value.
func EvaluateCalcValue(v interface{}) float64 {
	switch val := v.(type) {
	case Length:
		return val.Value
	case *CalcExpression:
		return EvaluateCalc(val).Value
	}
	return 0
}

// ParseClamp parses a clamp(min, val, max) expression.
func ParseClamp(s string) (Length, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "clamp(") || !strings.HasSuffix(s, ")") {
		return Length{}, fmt.Errorf("not a clamp expression")
	}
	inner := s[6 : len(s)-1] // Remove "clamp(" and ")"

	parts := splitMathFuncParts(inner)
	if len(parts) != 3 {
		return Length{}, fmt.Errorf("clamp requires exactly 3 arguments")
	}

	minLen := ParseLength(strings.TrimSpace(parts[0]))
	valLen := ParseLength(strings.TrimSpace(parts[1]))
	maxLen := ParseLength(strings.TrimSpace(parts[2]))

	minVal := minLen.Value
	valVal := valLen.Value
	maxVal := maxLen.Value

	// Handle percentage and other units - convert to pixels assuming 1px = 1 unit
	// A full implementation would resolve units in context
	if minLen.Unit == UnitPercent {
		minVal = minLen.Value // keep as-is for now
	}
	if maxLen.Unit == UnitPercent {
		maxVal = maxLen.Value
	}

	// Clamp: max(min, min(val, max))
	if valVal < minVal {
		valVal = minVal
	}
	if valVal > maxVal {
		valVal = maxVal
	}

	return Length{Value: valVal, Unit: UnitPx}, nil
}

// ParseMin parses a min(val1, val2, ...) expression.
func ParseMin(s string) (Length, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "min(") || !strings.HasSuffix(s, ")") {
		return Length{}, fmt.Errorf("not a min expression")
	}
	inner := s[4 : len(s)-1] // Remove "min(" and ")"

	parts := splitMathFuncParts(inner)
	if len(parts) < 1 {
		return Length{}, fmt.Errorf("min requires at least 1 argument")
	}

	minVal := ParseLength(strings.TrimSpace(parts[0])).Value
	for i := 1; i < len(parts); i++ {
		v := ParseLength(strings.TrimSpace(parts[i])).Value
		if v < minVal {
			minVal = v
		}
	}

	return Length{Value: minVal, Unit: UnitPx}, nil
}

// ParseMax parses a max(val1, val2, ...) expression.
func ParseMax(s string) (Length, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "max(") || !strings.HasSuffix(s, ")") {
		return Length{}, fmt.Errorf("not a max expression")
	}
	inner := s[4 : len(s)-1] // Remove "max(" and ")"

	parts := splitMathFuncParts(inner)
	if len(parts) < 1 {
		return Length{}, fmt.Errorf("max requires at least 1 argument")
	}

	maxVal := ParseLength(strings.TrimSpace(parts[0])).Value
	for i := 1; i < len(parts); i++ {
		v := ParseLength(strings.TrimSpace(parts[i])).Value
		if v > maxVal {
			maxVal = v
		}
	}

	return Length{Value: maxVal, Unit: UnitPx}, nil
}

// splitMathFuncParts splits arguments in a math function like min(), max(), clamp().
// Respects nested parentheses.
func splitMathFuncParts(inner string) []string {
	var parts []string
	var current []byte
	depth := 0
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		if c == '(' {
			depth++
			current = append(current, c)
		} else if c == ')' {
			depth--
			current = append(current, c)
		} else if c == ',' && depth == 0 {
			parts = append(parts, strings.TrimSpace(string(current)))
			current = nil
		} else {
			current = append(current, c)
		}
	}
	if len(current) > 0 {
		parts = append(parts, strings.TrimSpace(string(current)))
	}
	return parts
}

// ParseRadialGradient parses a radial-gradient value.
func ParseRadialGradient(s string) ([]GradientStop, string, string, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "radial-gradient(") || !strings.HasSuffix(s, ")") {
		return nil, "", "", fmt.Errorf("not a radial-gradient")
	}
	inner := s[16 : len(s)-1]
	
	parts := splitGradientParts(inner)
	var stops []GradientStop
	shape := "ellipse"
	position := "center"
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "circle") || strings.Contains(part, "ellipse") {
			if strings.Contains(part, "circle") {
				shape = "circle"
			}
		} else if strings.Contains(part, "at ") {
			position = part
		} else if strings.Contains(part, "%") || strings.HasPrefix(part, "#") || strings.HasPrefix(part, "rgb") || isColorString(part) {
			stop := parseGradientStop(part)
			stops = append(stops, stop)
		}
	}
	
	return stops, shape, position, nil
}

// FilterEffect represents a CSS filter effect.
type FilterEffect struct {
	Type   string   // blur, brightness, contrast, grayscale, sepia, hue-rotate, drop-shadow
	Amount float64  // 0-1 or 0-N depending on filter type
	Color  Color    // for drop-shadow
	X, Y   float64  // offset for drop-shadow
	Blur   float64  // blur radius for drop-shadow
}

// ParseFilter parses a CSS filter value.
// Format: blur(px), brightness(%), contrast(%), grayscale(%), sepia(%),
//         hue-rotate(deg), drop-shadow(x y blur color)
func ParseFilter(s string) ([]FilterEffect, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return nil, nil
	}

	var effects []FilterEffect
	// Split by space (multiple filters)
	parts := strings.Fields(s)
	for _, part := range parts {
		effect := FilterEffect{}
		lower := strings.ToLower(part)

		if strings.HasPrefix(lower, "blur(") && strings.HasSuffix(lower, ")") {
			inner := part[5 : len(part)-1]
			if v, err := strconv.ParseFloat(inner, 64); err == nil {
				effect.Type = "blur"
				effect.Amount = v
				effects = append(effects, effect)
			}
		} else if strings.HasPrefix(lower, "brightness(") && strings.HasSuffix(lower, ")") {
			inner := part[11 : len(part)-1]
			if v, err := strconv.ParseFloat(strings.TrimSuffix(inner, "%"), 64); err == nil {
				effect.Type = "brightness"
				effect.Amount = v / 100
				effects = append(effects, effect)
			}
		} else if strings.HasPrefix(lower, "contrast(") && strings.HasSuffix(lower, ")") {
			inner := part[9 : len(part)-1]
			if v, err := strconv.ParseFloat(strings.TrimSuffix(inner, "%"), 64); err == nil {
				effect.Type = "contrast"
				effect.Amount = v / 100
				effects = append(effects, effect)
			}
		} else if strings.HasPrefix(lower, "grayscale(") && strings.HasSuffix(lower, ")") {
			inner := part[11 : len(part)-1]
			if v, err := strconv.ParseFloat(strings.TrimSuffix(inner, "%"), 64); err == nil {
				effect.Type = "grayscale"
				effect.Amount = v / 100
				effects = append(effects, effect)
			}
		} else if strings.HasPrefix(lower, "sepia(") && strings.HasSuffix(lower, ")") {
			inner := part[6 : len(part)-1]
			if v, err := strconv.ParseFloat(strings.TrimSuffix(inner, "%"), 64); err == nil {
				effect.Type = "sepia"
				effect.Amount = v / 100
				effects = append(effects, effect)
			}
		} else if strings.HasPrefix(lower, "hue-rotate(") && strings.HasSuffix(lower, ")") {
			inner := part[11 : len(part)-1]
			if v, err := strconv.ParseFloat(strings.TrimSuffix(inner, "deg"), 64); err == nil {
				effect.Type = "hue-rotate"
				effect.Amount = v
				effects = append(effects, effect)
			}
		} else if strings.HasPrefix(lower, "drop-shadow(") && strings.HasSuffix(lower, ")") {
			inner := part[12 : len(part)-1]
			effect.Type = "drop-shadow"
			effect.Color = Color{R: 0, G: 0, B: 0, A: 128}
			// Parse drop-shadow arguments
			args := strings.Fields(inner)
			if len(args) >= 2 {
				effect.X, _ = strconv.ParseFloat(args[0], 64)
				effect.Y, _ = strconv.ParseFloat(args[1], 64)
			}
			if len(args) >= 3 {
				effect.Blur, _ = strconv.ParseFloat(args[2], 64)
			}
			if len(args) >= 4 {
				effect.Color = ParseColor(args[3])
			}
			effects = append(effects, effect)
		}
	}
	return effects, nil
}

// ClipShape represents a clipping shape.
type ClipShape struct {
	Type     string   // "inset", "circle", "ellipse", "polygon", "none"
	InsetTop, InsetRight, InsetBottom, InsetLeft float64  // for inset
	RadiusX, RadiusY float64  // for circle/ellipse
	CenterX, CenterY float64  // for circle/ellipse
	Points   [][2]float64    // for polygon
}

// ParseClipPath parses a CSS clip-path value.
func ParseClipPath(s string) (ClipShape, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return ClipShape{Type: "none"}, nil
	}

	shape := ClipShape{}

	if strings.HasPrefix(s, "inset(") && strings.HasSuffix(s, ")") {
		shape.Type = "inset"
		inner := s[6 : len(s)-1]
		parts := strings.Fields(strings.ReplaceAll(inner, ",", " "))
		if len(parts) >= 1 {
			v, _ := strconv.ParseFloat(parts[0], 64)
			shape.InsetTop = v
			shape.InsetRight = v
			shape.InsetBottom = v
			shape.InsetLeft = v
		}
		if len(parts) >= 2 {
			v, _ := strconv.ParseFloat(parts[1], 64)
			shape.InsetRight = v
			shape.InsetLeft = v
		}
		if len(parts) >= 3 {
			v, _ := strconv.ParseFloat(parts[2], 64)
			shape.InsetBottom = v
		}
		if len(parts) >= 4 {
			v, _ := strconv.ParseFloat(parts[3], 64)
			shape.InsetLeft = v
		}
	} else if strings.HasPrefix(s, "circle(") && strings.HasSuffix(s, ")") {
		shape.Type = "circle"
		inner := s[7 : len(s)-1]
		parts := strings.Fields(strings.ReplaceAll(inner, "at", ","))
		if len(parts) >= 1 && parts[0] != "" {
			if v, err := strconv.ParseFloat(strings.TrimSuffix(parts[0], "px"), 64); err == nil {
				shape.RadiusX = v
				shape.RadiusY = v
			}
		}
		if len(parts) >= 3 {
			if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
				shape.CenterX = v
			}
			if v, err := strconv.ParseFloat(parts[2], 64); err == nil {
				shape.CenterY = v
			}
		}
	} else if strings.HasPrefix(s, "ellipse(") && strings.HasSuffix(s, ")") {
		shape.Type = "ellipse"
		inner := s[8 : len(s)-1]
		parts := strings.Fields(strings.ReplaceAll(inner, "at", ","))
		if len(parts) >= 2 {
			if v, err := strconv.ParseFloat(strings.TrimSuffix(parts[0], "px"), 64); err == nil {
				shape.RadiusX = v
			}
			if v, err := strconv.ParseFloat(strings.TrimSuffix(parts[1], "px"), 64); err == nil {
				shape.RadiusY = v
			}
		}
	} else if strings.HasPrefix(s, "polygon(") && strings.HasSuffix(s, ")") {
		shape.Type = "polygon"
		inner := s[9 : len(s)-1]
		parts := strings.Split(inner, ",")
		for _, part := range parts {
			coords := strings.Fields(part)
			if len(coords) >= 2 {
				x, _ := strconv.ParseFloat(coords[0], 64)
				y, _ := strconv.ParseFloat(coords[1], 64)
				shape.Points = append(shape.Points, [2]float64{x, y})
			}
		}
	}

	return shape, nil
}

// ParseBackdropFilter parses a CSS backdrop-filter value (same syntax as filter).
func ParseBackdropFilter(s string) ([]FilterEffect, error) {
	return ParseFilter(s)
}

// CounterStyle represents a CSS counter style.
type CounterStyle int

const (
	CounterStyleDecimal CounterStyle = iota
	CounterStyleLowerAlpha
	CounterStyleUpperAlpha
	CounterStyleLowerRoman
	CounterStyleUpperRoman
	CounterStyleDisc
	CounterStyleSquare
	CounterStyleCircle
)

// CursorType represents CSS cursor styles.
type CursorType int

const (
	CursorAuto CursorType = iota
	CursorDefault
	CursorPointer
	CursorText
	CursorWait
	CursorCrosshair
	CursorMove
	CursorNotAllowed
	CursorGrab
	CursorGrabbing
	CursorVerticalText
	CursorHelp
	CursorProgress
	CursorCell
	CursorContextMenu
	CursorAlias
	CursorCopy
	CursorNone
)

// ParseCursor parses a CSS cursor value and returns the CursorType.
// Supports: auto, default, pointer, text, wait, crosshair, move, not-allowed, grab, grabbing,
// vertical-text, help, progress, cell, context-menu, alias, copy, none
func ParseCursor(s string) CursorType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "auto":
		return CursorAuto
	case "default":
		return CursorDefault
	case "pointer":
		return CursorPointer
	case "text":
		return CursorText
	case "wait":
		return CursorWait
	case "crosshair":
		return CursorCrosshair
	case "move":
		return CursorMove
	case "not-allowed", "notallowed":
		return CursorNotAllowed
	case "grab":
		return CursorGrab
	case "grabbing":
		return CursorGrabbing
	case "vertical-text":
		return CursorVerticalText
	case "help":
		return CursorHelp
	case "progress":
		return CursorProgress
	case "cell":
		return CursorCell
	case "context-menu":
		return CursorContextMenu
	case "alias":
		return CursorAlias
	case "copy":
		return CursorCopy
	case "none":
		return CursorNone
	default:
		return CursorAuto
	}
}

// CursorToString returns a human-readable string for the cursor type.
func CursorToString(ct CursorType) string {
	switch ct {
	case CursorAuto:
		return "auto"
	case CursorDefault:
		return "default"
	case CursorPointer:
		return "pointer"
	case CursorText:
		return "text"
	case CursorWait:
		return "wait"
	case CursorCrosshair:
		return "crosshair"
	case CursorMove:
		return "move"
	case CursorNotAllowed:
		return "not-allowed"
	case CursorGrab:
		return "grab"
	case CursorGrabbing:
		return "grabbing"
	case CursorVerticalText:
		return "vertical-text"
	case CursorHelp:
		return "help"
	case CursorProgress:
		return "progress"
	case CursorCell:
		return "cell"
	case CursorContextMenu:
		return "context-menu"
	case CursorAlias:
		return "alias"
	case CursorCopy:
		return "copy"
	case CursorNone:
		return "none"
	default:
		return "auto"
	}
}

// Counter represents a CSS counter() value.
type Counter struct {
	Name  string
	Style CounterStyle
}

// ParseCounter parses a counter() value.
// Format: counter(name) or counter(name, style)
// style is optional and defaults to decimal.
func ParseCounter(s string) Counter {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "counter(") || !strings.HasSuffix(s, ")") {
		return Counter{}
	}
	inner := s[8 : len(s)-1] // Remove "counter(" and ")"
	parts := splitFunctionParts(inner)
	if len(parts) == 0 {
		return Counter{}
	}
	name := strings.TrimSpace(parts[0])
	style := CounterStyleDecimal
	if len(parts) >= 2 {
		style = parseCounterStyle(strings.TrimSpace(parts[1]))
	}
	return Counter{Name: name, Style: style}
}

// Counters represents a CSS counters() value.
type Counters struct {
	Name      string
	Separator string
	Style     CounterStyle
}

// ParseCounters parses a counters() value.
// Format: counters(name, separator) or counters(name, separator, style)
// separator is required; style is optional and defaults to decimal.
func ParseCounters(s string) Counters {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "counters(") || !strings.HasSuffix(s, ")") {
		return Counters{}
	}
	inner := s[9 : len(s)-1] // Remove "counters(" and ")"
	parts := splitFunctionParts(inner)
	if len(parts) < 2 {
		return Counters{}
	}
	name := strings.TrimSpace(parts[0])
	separator := strings.TrimSpace(parts[1])
	// Strip surrounding quotes if present (CSS string literals)
	separator = stripCSSStringQuotes(separator)
	style := CounterStyleDecimal
	if len(parts) >= 3 {
		style = parseCounterStyle(strings.TrimSpace(parts[2]))
	}
	return Counters{Name: name, Separator: separator, Style: style}
}

// Attr represents a CSS attr() value.
type Attr struct {
	Name     string
	Type     string
	Fallback string
}

// ParseAttr parses an attr() value.
// Format: attr(data-xxx) or attr(data-xxx, type) or attr(data-xxx, type, fallback)
// type defaults to "string" if not specified.
func ParseAttr(s string) Attr {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "attr(") || !strings.HasSuffix(s, ")") {
		return Attr{}
	}
	inner := s[5 : len(s)-1] // Remove "attr(" and ")"
	parts := splitFunctionParts(inner)
	if len(parts) == 0 {
		return Attr{}
	}
	name := strings.TrimSpace(parts[0])
	attrType := "string"
	fallback := ""
	if len(parts) == 2 {
		rest := strings.TrimSpace(parts[1])
		// Check if second arg contains a comma (type,fallback format)
		commaIdx := -1
		depth := 0
		for i, c := range rest {
			if c == '(' {
				depth++
			} else if c == ')' {
				depth--
			} else if c == ',' && depth == 0 {
				commaIdx = i
				break
			}
		}
		if commaIdx > 0 {
			// type,fallback
			attrType = strings.TrimSpace(rest[:commaIdx])
			fallback = strings.TrimSpace(rest[commaIdx+1:])
		} else if len(rest) > 0 && (rest[0] == '"' || rest[0] == '\'') {
			// Quoted value is fallback, type defaults to "string"
			fallback = rest
		} else {
			// Unquoted word is type
			attrType = rest
		}
	} else if len(parts) >= 3 {
		attrType = strings.TrimSpace(parts[1])
		fallback = strings.TrimSpace(parts[2])
	}
	return Attr{Name: name, Type: attrType, Fallback: fallback}
}

// parseCounterStyle converts a string to CounterStyle.
func parseCounterStyle(s string) CounterStyle {
	switch strings.ToLower(s) {
	case "decimal":
		return CounterStyleDecimal
	case "lower-alpha", "lower-latin":
		return CounterStyleLowerAlpha
	case "upper-alpha", "upper-latin":
		return CounterStyleUpperAlpha
	case "lower-roman":
		return CounterStyleLowerRoman
	case "upper-roman":
		return CounterStyleUpperRoman
	case "disc":
		return CounterStyleDisc
	case "square":
		return CounterStyleSquare
	case "circle":
		return CounterStyleCircle
	default:
		return CounterStyleDecimal
	}
}

// stripCSSStringQuotes removes surrounding single or double quotes from a CSS string literal.
func stripCSSStringQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// splitFunctionParts splits a function's inner arguments, handling nested parentheses and quoted strings.
// Each part is individually trimmed of surrounding whitespace.
func splitFunctionParts(inner string) []string {
	var parts []string
	var current []byte
	depth := 0
	inQuote := false
	quoteChar := byte(0)
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		// Enter quote
		if !inQuote && (c == '"' || c == '\'') {
			inQuote = true
			quoteChar = c
			current = append(current, c)
		} else if inQuote && c == quoteChar {
			// Exit quote
			inQuote = false
			quoteChar = 0
			current = append(current, c)
		} else if c == '(' && !inQuote {
			depth++
			current = append(current, c)
		} else if c == ')' && !inQuote {
			depth--
			if depth < 0 {
				depth = 0
			}
			current = append(current, c)
		} else if c == ',' && depth == 0 && !inQuote {
			parts = append(parts, strings.TrimSpace(string(current)))
			current = nil
		} else {
			current = append(current, c)
		}
	}
	if len(current) > 0 {
		parts = append(parts, strings.TrimSpace(string(current)))
	}
	return parts
}
