package css

import (
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
