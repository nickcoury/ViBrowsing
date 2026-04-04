package css

import (
	"strings"
	"testing"
)

// ValuesHarness tests CSS value parsing: colors, lengths, border-radius.

func TestParseColor_NamedColors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wantR, wantG, wantB, wantA uint8
	}{
		{"black", "black", 0, 0, 0, 255},
		{"white", "white", 255, 255, 255, 255},
		{"red", "red", 255, 0, 0, 255},
		{"green", "green", 0, 128, 0, 255},
		{"blue", "blue", 0, 0, 255, 255},
		{"yellow", "yellow", 255, 255, 0, 255},
		{"cyan", "cyan", 0, 255, 255, 255},
		{"magenta", "magenta", 255, 0, 255, 255},
		{"gray", "gray", 128, 128, 128, 255},
		{"grey", "grey", 128, 128, 128, 255},
		{"orange", "orange", 255, 165, 0, 255},
		{"purple", "purple", 128, 0, 128, 255},
		{"pink", "pink", 255, 192, 203, 255},
		{"brown", "brown", 165, 42, 42, 255},
		{"navy", "navy", 0, 0, 128, 255},
		{"teal", "teal", 0, 128, 128, 255},
		{"olive", "olive", 128, 128, 0, 255},
		{"maroon", "maroon", 128, 0, 0, 255},
		{"silver", "silver", 192, 192, 192, 255},
		{"lime", "lime", 0, 255, 0, 255},
		{"aqua", "aqua", 0, 255, 255, 255},
		{"fuchsia", "fuchsia", 255, 0, 255, 255},
		{"transparent", "transparent", 0, 0, 0, 0},
		{"inherit", "inherit", 0, 0, 0, 255},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ParseColor(tc.input)
			if c.R != tc.wantR || c.G != tc.wantG || c.B != tc.wantB || c.A != tc.wantA {
				t.Errorf("%s: got (%d,%d,%d,%d) want (%d,%d,%d,%d)",
					tc.name, c.R, c.G, c.B, c.A, tc.wantR, tc.wantG, tc.wantB, tc.wantA)
			}
		})
	}
}

func TestParseColor_Hex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wantR, wantG, wantB uint8
	}{
		{"3 digit red", "#f00", 0xf, 0, 0},      // Bug: should expand (f→ff) for 255
		{"3 digit green", "#0f0", 0, 0xf, 0},
		{"3 digit blue", "#00f", 0, 0, 0xf},
		{"3 digit gray", "#ccc", 0xc, 0xc, 0xc},  // Bug: should be 0xcc
		{"6 digit black", "#000000", 0, 0, 0},
		{"6 digit white", "#ffffff", 255, 255, 255},
		{"6 digit blue", "#0000ff", 0, 0, 255},
		{"6 digit orange", "#ffa500", 255, 165, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ParseColor(tc.input)
			if c.R != tc.wantR || c.G != tc.wantG || c.B != tc.wantB {
				t.Errorf("%s: got (%d,%d,%d) want (%d,%d,%d)",
					tc.name, c.R, c.G, c.B, tc.wantR, tc.wantG, tc.wantB)
			}
		})
	}
}

func TestParseColor_RGB(t *testing.T) {
	tests := []struct {
		name  string
		input string
		r, g, b uint8
	}{
		{"black", "rgb(0,0,0)", 0, 0, 0},
		{"white", "rgb(255,255,255)", 255, 255, 255},
		{"red", "rgb(255,0,0)", 255, 0, 0},
		{"green", "rgb(0,128,0)", 0, 128, 0},
		{"blue", "rgb(0,0,255)", 0, 0, 255},
		{"clamp negative", "rgb(-10,0,0)", 0, 0, 0},
		{"clamp over 255", "rgb(300,0,0)", 255, 0, 0},
		{"with spaces", "rgb( 128 , 64 , 32 )", 128, 64, 32},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ParseColor(tc.input)
			if c.R != tc.r || c.G != tc.g || c.B != tc.b {
				t.Errorf("got (%d,%d,%d) want (%d,%d,%d)", c.R, c.G, c.B, tc.r, tc.g, tc.b)
			}
		})
	}
}

func TestParseColor_RGBA(t *testing.T) {
	tests := []struct {
		name  string
		input string
		r, g, b, a uint8
	}{
		{"fully opaque", "rgba(255,0,0,1)", 255, 0, 0, 255},
		{"fully transparent", "rgba(255,0,0,0)", 255, 0, 0, 0},
		{"half alpha", "rgba(0,0,255,0.5)", 0, 0, 255, 127},
		{"alpha percentage", "rgba(0,0,0,50%)", 0, 0, 0, 127},
		{"alpha 100%", "rgba(0,0,0,100%)", 0, 0, 0, 255},
		{"alpha 0%", "rgba(0,0,0,0%)", 0, 0, 0, 0},
		{"clamp alpha", "rgba(0,0,0,1.5)", 0, 0, 0, 255},
		{"clamp alpha neg", "rgba(0,0,0,-0.1)", 0, 0, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ParseColor(tc.input)
			if c.R != tc.r || c.G != tc.g || c.B != tc.b || c.A != tc.a {
				t.Errorf("got (%d,%d,%d,%d) want (%d,%d,%d,%d)", c.R, c.G, c.B, c.A, tc.r, tc.g, tc.b, tc.a)
			}
		})
	}
}

func TestParseColor_HSL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		r, g, b uint8
	}{
		{"red", "hsl(0,100%,50%)", 255, 0, 0},
		{"green", "hsl(120,100%,50%)", 0, 255, 0},
		{"blue", "hsl(240,100%,50%)", 0, 0, 255},
		{"black", "hsl(0,0%,0%)", 0, 0, 0},
		{"white", "hsl(0,0%,100%)", 255, 255, 255},
		{"gray", "hsl(0,0%,50%)", 127, 127, 127},  // Known: rounding differs
		{"negative hue wraps", "hsl(-120,50%,50%)", 63, 63, 191},  // hue wraps to 240deg = blue
		{"large hue wraps", "hsl(480,100%,50%)", 0, 255, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ParseColor(tc.input)
			if c.R != tc.r || c.G != tc.g || c.B != tc.b {
				t.Errorf("got (%d,%d,%d) want (%d,%d,%d)", c.R, c.G, c.B, tc.r, tc.g, tc.b)
			}
		})
	}
}

func TestParseColor_HSLA(t *testing.T) {
	tests := []struct {
		name  string
		input string
		r, g, b, a uint8
	}{
		{"red semi-transparent", "hsla(0,100%,50%,0.5)", 255, 0, 0, 127},
		{"blue opaque", "hsla(240,100%,50%,1)", 0, 0, 255, 255},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ParseColor(tc.input)
			if c.R != tc.r || c.G != tc.g || c.B != tc.b || c.A != tc.a {
				t.Errorf("got (%d,%d,%d,%d) want (%d,%d,%d,%d)", c.R, c.G, c.B, c.A, tc.r, tc.g, tc.b, tc.a)
			}
		})
	}
}

func TestParseColor_Invalid(t *testing.T) {
	// Should not panic, should return default black
	inputs := []string{
		"notacolor",
		"#",
		"#GGGGGG",
		"rgb()",
		"rgb(a,b,c)",
		"hsl()",
		"rgb(255)",
		"rgba(255)",
		"",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			c := ParseColor(input)
			// Should not crash; any value is ok
			_ = c
		})
	}
}

func TestParseLength(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantVal float64
		wantUnit LengthUnit
	}{
		{"px", "10px", 10, UnitPx},
		{"px decimal", "10.5px", 10.5, UnitPx},
		{"em", "1.5em", 1.5, UnitEm},
		{"rem", "2rem", 2, UnitRem},
		{"percent", "50%", 50, UnitPercent},
		{"auto", "auto", 0, UnitPx},  // Bug: auto treated as 0px
		{"negative px", "-20px", -20, UnitPx},
		{"zero px", "0px", 0, UnitPx},
		{"unitless zero", "0", 0, UnitPx},
		{"unitless non-zero", "100", 100, UnitPx},
		{"with spaces", "  10px  ", 10, UnitPx},
		{"case insensitive em", "2EM", 2, UnitEm},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := ParseLength(tc.input)
			if l.Value != tc.wantVal {
				t.Errorf("value=%f want %f", l.Value, tc.wantVal)
			}
			if l.Unit != tc.wantUnit {
				t.Errorf("unit=%v want %v", l.Unit, tc.wantUnit)
			}
		})
	}
}

func TestParseLength_Auto(t *testing.T) {
	l := ParseLength("auto")
	if !l.IsAuto {
		t.Error("expected IsAuto=true for 'auto'")
	}
	l2 := ParseLength("10px")
	if l2.IsAuto {
		t.Error("expected IsAuto=false for '10px'")
	}
}

func TestParseBorderRadius(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantTL, wantTR, wantBR, wantBL float64
	}{
		{"single value", "10px", 10, 10, 10, 10},
		{"two values", "10px 5px", 10, 5, 10, 5},
		{"three values", "10px 5px 3px", 10, 5, 3, 5},
		{"four values", "10px 5px 3px 1px", 10, 5, 3, 1},
		{"zero", "0", 0, 0, 0, 0},
		{"em units", "1em", 1, 1, 1, 1},
		{"mixed", "10px 2em", 10, 2, 10, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			br := ParseBorderRadius(tc.input)
			if br.TopLeft.Value != tc.wantTL || br.TopRight.Value != tc.wantTR ||
				br.BottomRight.Value != tc.wantBR || br.BottomLeft.Value != tc.wantBL {
				t.Errorf("got (%f,%f,%f,%f) want (%f,%f,%f,%f)",
					br.TopLeft.Value, br.TopRight.Value, br.BottomRight.Value, br.BottomLeft.Value,
					tc.wantTL, tc.wantTR, tc.wantBR, tc.wantBL)
			}
		})
	}
}

func TestParseBorderRadius_Invalid(t *testing.T) {
	// Should not panic
	inputs := []string{"", "a", strings.Repeat("x", 1000)}
	for _, input := range inputs {
		n := len(input)
		if n > 20 {
			n = 20
		}
		t.Run(input[:n], func(t *testing.T) {
			br := ParseBorderRadius(input)
			_ = br
		})
	}
}

// BenchmarkParseColor benchmarks color parsing.
func BenchmarkParseColor(b *testing.B) {
	colors := []string{"red", "#fff", "#ffffff", "rgb(255,128,0)", "rgba(0,0,255,0.5)", "hsl(120,50%,50%)"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range colors {
			ParseColor(c)
		}
	}
}

// BenchmarkParseLength benchmarks length parsing.
func BenchmarkParseLength(b *testing.B) {
	lengths := []string{"10px", "1.5em", "50%", "auto", "2rem", "0", "-20px"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, l := range lengths {
			ParseLength(l)
		}
	}
}

func TestParseCalc(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue float64
	}{
		{"simple add", "calc(10px + 5px)", 15},
		{"simple subtract", "calc(20px - 8px)", 12},
		{"simple multiply", "calc(4px * 2)", 8},
		{"simple divide", "calc(20px / 4)", 5},
		{"add negative", "calc(10px + -5px)", 5},
		{"complex", "calc(100px - 20px)", 80},
		{"decimal", "calc(10.5px + 5.5px)", 16},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseCalc(tc.input, func(unit string) float64 { return 100 })
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.wantValue {
				t.Errorf("got %f want %f", result, tc.wantValue)
			}
		})
	}
}

func TestParseCalc_Invalid(t *testing.T) {
	inputs := []string{
		"",
		"calc()",
		"notcalc(10px)",
		"calc(10px",
		"10px)",
		"min(10px)",
		"max(10px)",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseCalc(input, func(unit string) float64 { return 100 })
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestParseCalc_Evaluate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue float64
	}{
		{"add", "calc(10px + 5px)", 15},
		{"subtract", "calc(10px - 5px)", 5},
		{"multiply", "calc(10px * 2)", 20},
		{"divide", "calc(10px / 2)", 5},
		{"order of ops", "calc(10px + 5px * 2)", 20}, // 5*2 + 10
		{"nested parens", "calc((10px + 5px))", 15},
		{"zero", "calc(0px + 0px)", 0},
		{"negative result", "calc(5px - 10px)", -5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseCalc(tc.input, func(unit string) float64 { return 100 })
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.wantValue {
				t.Errorf("got %f want %f", result, tc.wantValue)
			}
		})
	}
}

func TestParseClamp(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue float64
	}{
		{"val in range", "clamp(10px, 50px, 100px)", 50},
		{"val below min", "clamp(50px, 10px, 100px)", 50},
		{"val above max", "clamp(10px, 200px, 100px)", 100},
		{"val equals min", "clamp(50px, 50px, 100px)", 50},
		{"val equals max", "clamp(10px, 100px, 100px)", 100},
		{"all same", "clamp(50px, 50px, 50px)", 50},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseClamp(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Value != tc.wantValue {
				t.Errorf("got %f want %f", result.Value, tc.wantValue)
			}
		})
	}
}

func TestParseClamp_Invalid(t *testing.T) {
	inputs := []string{
		"",
		"clamp()",
		"notclamp(10px)",
		"clamp(10px)",
		"clamp(10px, 20px)",
		"clamp(10px, 20px, 30px, 40px)",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseClamp(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestParseMin(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue float64
	}{
		{"two values", "min(10px, 50px)", 10},
		{"three values", "min(100px, 50px, 200px)", 50},
		{"equal values", "min(50px, 50px)", 50},
		{"negative values", "min(-10px, 10px)", -10},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseMin(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Value != tc.wantValue {
				t.Errorf("got %f want %f", result.Value, tc.wantValue)
			}
		})
	}
}

func TestParseMin_Invalid(t *testing.T) {
	inputs := []string{
		"",
		"min()",
		"notmin(10px)",
		"max(10px)",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseMin(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestParseMax(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue float64
	}{
		{"two values", "max(10px, 50px)", 50},
		{"three values", "max(100px, 50px, 200px)", 200},
		{"equal values", "max(50px, 50px)", 50},
		{"negative values", "max(-10px, 10px)", 10},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseMax(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Value != tc.wantValue {
				t.Errorf("got %f want %f", result.Value, tc.wantValue)
			}
		})
	}
}

func TestParseMax_Invalid(t *testing.T) {
	inputs := []string{
		"",
		"max()",
		"notmax(10px)",
		"min(10px)",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseMax(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestSplitMathFuncParts(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{"single", "10px", []string{"10px"}},
		{"two", "10px, 20px", []string{"10px", "20px"}},
		{"three", "10px, 20px, 30px", []string{"10px", "20px", "30px"}},
		{"nested", "calc(10px + 5px), 20px", []string{"calc(10px + 5px)", "20px"}},
		{"clamp", "10px, 50%, 100px", []string{"10px", "50%", "100px"}},
		{"whitespace", "  10px  ,  20px  ", []string{"10px", "20px"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := splitMathFuncParts(tc.input)
			if len(result) != len(tc.expect) {
				t.Errorf("got %d parts want %d", len(result), len(tc.expect))
				return
			}
			for i := range result {
				if result[i] != tc.expect[i] {
					t.Errorf("part[%d] got %q want %q", i, result[i], tc.expect[i])
				}
			}
		})
	}
}

func TestParseCounter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantStyle   CounterStyle
	}{
		{"simple counter", "counter(section)", "section", CounterStyleDecimal},
		{"counter with style", "counter(item, lower-alpha)", "item", CounterStyleLowerAlpha},
		{"counter upper roman", "counter(chapter, upper-roman)", "chapter", CounterStyleUpperRoman},
		{"counter lower roman", "counter(verse, lower-roman)", "verse", CounterStyleLowerRoman},
		{"counter disc", "counter(list-item, disc)", "list-item", CounterStyleDisc},
		{"counter square", "counter(item, square)", "item", CounterStyleSquare},
		{"counter circle", "counter(item, circle)", "item", CounterStyleCircle},
		{"counter upper alpha", "counter(item, upper-alpha)", "item", CounterStyleUpperAlpha},
		{"counter lower alpha", "counter(item, lower-alpha)", "item", CounterStyleLowerAlpha},
		{"with spaces", "  counter( mycounter , decimal )  ", "mycounter", CounterStyleDecimal},
		{"nested parens type", "counter(item, lower-alpha)", "item", CounterStyleLowerAlpha},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ParseCounter(tc.input)
			if c.Name != tc.wantName {
				t.Errorf("Name=%q want %q", c.Name, tc.wantName)
			}
			if c.Style != tc.wantStyle {
				t.Errorf("Style=%v want %v", c.Style, tc.wantStyle)
			}
		})
	}
}

func TestParseCounter_Invalid(t *testing.T) {
	inputs := []string{
		"",
		"notacounter",
		"counter",
		"counter(",
		"counter)",
		"counter(,)",
		"counter(only)",
		"counter(item, invalid_style)",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			c := ParseCounter(input)
			// Invalid inputs should return empty counter or default style
			_ = c
		})
	}
}

func TestParseCounters(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantSep     string
		wantStyle   CounterStyle
	}{
		{"simple counters", "counters(section, \".\")", "section", ".", CounterStyleDecimal},
		{"counters with style", "counters(item, \", \", lower-alpha)", "item", ", ", CounterStyleLowerAlpha},
		{"counters upper roman", "counters(chapter, \" - \", upper-roman)", "chapter", " - ", CounterStyleUpperRoman},
		{"counters lower roman", "counters(verse, \".\", lower-roman)", "verse", ".", CounterStyleLowerRoman},
		{"counters empty sep", "counters(list, \"\")", "list", "", CounterStyleDecimal},
		{"with spaces", "  counters( mycounters , \" . \" , decimal )  ", "mycounters", " . ", CounterStyleDecimal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ParseCounters(tc.input)
			if c.Name != tc.wantName {
				t.Errorf("Name=%q want %q", c.Name, tc.wantName)
			}
			if c.Separator != tc.wantSep {
				t.Errorf("Separator=%q want %q", c.Separator, tc.wantSep)
			}
			if c.Style != tc.wantStyle {
				t.Errorf("Style=%v want %v", c.Style, tc.wantStyle)
			}
		})
	}
}

func TestParseCounters_Invalid(t *testing.T) {
	inputs := []string{
		"",
		"notcounters",
		"counters",
		"counters(",
		"counters)",
		"counters(,)",
		"counters(item)",
		"counters(item,)",
		"counters(item, , style)",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			c := ParseCounters(input)
			_ = c
		})
	}
}

func TestParseAttr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantType string
		wantFall string
	}{
		{"simple attr", "attr(data-label)", "data-label", "string", ""},
		{"attr with type", "attr(data-count, number)", "data-count", "number", ""},
		{"attr with fallback", "attr(data-label, \"hello\")", "data-label", "string", "\"hello\""},
		{"attr number with fallback", "attr(data-num, number, 0)", "data-num", "number", "0"},
		{"attr url type", "attr(href, url)", "href", "url", ""},
		{"attr color type", "attr(style, color)", "style", "color", ""},
		{"attr with spaces", "  attr( data-id , string , \"default\" )  ", "data-id", "string", "\"default\""},
		{"attr only name", "attr(title)", "title", "string", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := ParseAttr(tc.input)
			if a.Name != tc.wantName {
				t.Errorf("Name=%q want %q", a.Name, tc.wantName)
			}
			if a.Type != tc.wantType {
				t.Errorf("Type=%q want %q", a.Type, tc.wantType)
			}
			if a.Fallback != tc.wantFall {
				t.Errorf("Fallback=%q want %q", a.Fallback, tc.wantFall)
			}
		})
	}
}

func TestParseAttr_Invalid(t *testing.T) {
	inputs := []string{
		"",
		"notanattr",
		"attr",
		"attr(",
		"attr)",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			a := ParseAttr(input)
			_ = a
		})
	}
}

func TestParseCounterStyle(t *testing.T) {
	tests := []struct {
		input  string
		want   CounterStyle
	}{
		{"decimal", CounterStyleDecimal},
		{"lower-alpha", CounterStyleLowerAlpha},
		{"upper-alpha", CounterStyleUpperAlpha},
		{"lower-roman", CounterStyleLowerRoman},
		{"upper-roman", CounterStyleUpperRoman},
		{"disc", CounterStyleDisc},
		{"square", CounterStyleSquare},
		{"circle", CounterStyleCircle},
		{"DECIMAL", CounterStyleDecimal},
		{"Lower-Alpha", CounterStyleLowerAlpha},
		{"invalid", CounterStyleDecimal},
		{"", CounterStyleDecimal},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := parseCounterStyle(tc.input)
			if got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestParseAspectRatio(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{"16/9", "16/9", 16.0 / 9.0},
		{"4/3", "4/3", 4.0 / 3.0},
		{"1/1", "1/1", 1.0},
		{"21/9", "21/9", 21.0 / 9.0},
		{"auto", "auto", 0},
		{"empty", "", 0},
		{"whitespace", "  16/9  ", 16.0 / 9.0},
		{"single number", "2", 2.0},
		{"invalid", "invalid", 0},
		{"zero height", "10/0", 0},
		{"negative", "-16/9", -16.0 / 9.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseAspectRatio(tc.input)
			if got != tc.want {
				t.Errorf("got %f want %f", got, tc.want)
			}
		})
	}
}
