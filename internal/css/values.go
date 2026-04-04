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
