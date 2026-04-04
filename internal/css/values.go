package css

import (
	"image/color"
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
	return uint32(c.R)<<16 | uint32(c.R)<<8 | uint32(c.R),
		uint32(c.G)<<16 | uint32(c.G)<<8 | uint32(c.G),
		uint32(c.B)<<16 | uint32(c.B)<<8 | uint32(c.B),
		uint32(c.A)<<16 | uint32(c.A)<<8 | uint32(c.A)
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

	// rgba(r, g, b, a)
	if strings.HasPrefix(s, "rgba(") && strings.HasSuffix(s, ")") {
		inner := s[5 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 4 {
			r := parseUint8(parts[0])
			g := parseUint8(parts[1])
			b := parseUint8(parts[2])
			a := parseFloat255(parts[3])
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
	if v > 1 {
		v = 1
	}
	return uint8(v * 255)
}
