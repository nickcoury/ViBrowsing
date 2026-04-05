package css

import (
	"testing"
)

func TestSupports(t *testing.T) {
	tests := []struct {
		property string
		value    string
		expected bool
	}{
		// Display values
		{"display", "block", true},
		{"display", "flex", true},
		{"display", "grid", true},
		{"display", "none", true},
		{"display", "invalid", false},

		// Position values
		{"position", "absolute", true},
		{"position", "relative", true},
		{"position", "fixed", true},
		{"position", "sticky", true},
		{"position", "invalid", false},

		// Float values
		{"float", "left", true},
		{"float", "right", true},
		{"float", "none", true},
		{"float", "invalid", false},

		// Overflow values
		{"overflow", "hidden", true},
		{"overflow", "scroll", true},
		{"overflow", "auto", true},
		{"overflow", "visible", true},
		{"overflow", "invalid", false},

		// Visibility values
		{"visibility", "visible", true},
		{"visibility", "hidden", true},
		{"visibility", "collapse", true},

		// Font weight values
		{"font-weight", "normal", true},
		{"font-weight", "bold", true},
		{"font-weight", "400", true},
		{"font-weight", "700", true},

		// Transform values
		{"transform", "none", true},
		{"transform", "rotate(45deg)", true},
		{"transform", "scale(1.5)", true},
		{"transform", "translate(10px, 20px)", true},

		// Filter values
		{"filter", "none", true},
		{"filter", "blur(5px)", true},
		{"filter", "brightness(1.2)", true},
		{"filter", "grayscale(100%)", true},

		// Clip-path values
		{"clip-path", "none", true},
		{"clip-path", "circle(50%)", true},
		{"clip-path", "inset(10px)", true},
		{"clip-path", "polygon(50% 0%, 100% 100%, 0% 100%)", true},

		// Background image values
		{"background-image", "none", true},
		{"background-image", "url(image.png)", true},
		{"background-image", "linear-gradient(red, blue)", true},
		{"background-image", "radial-gradient(red, blue)", true},

		// Background repeat values
		{"background-repeat", "repeat", true},
		{"background-repeat", "no-repeat", true},
		{"background-repeat", "space", true},
		{"background-repeat", "round", true},

		// White-space values
		{"white-space", "normal", true},
		{"white-space", "pre", true},
		{"white-space", "pre-wrap", true},
		{"white-space", "nowrap", true},

		// Object-fit values
		{"object-fit", "cover", true},
		{"object-fit", "contain", true},
		{"object-fit", "fill", true},
		{"object-fit", "none", true},
		{"object-fit", "scale-down", true},

		// Font stretch values
		{"font-stretch", "ultra-condensed", true},
		{"font-stretch", "condensed", true},
		{"font-stretch", "normal", true},
		{"font-stretch", "expanded", true},
		{"font-stretch", "ultra-expanded", true},
		{"font-stretch", "50%", true},
		{"font-stretch", "200%", true},

		// Blend mode values
		{"mix-blend-mode", "normal", true},
		{"mix-blend-mode", "multiply", true},
		{"mix-blend-mode", "screen", true},
		{"mix-blend-mode", "overlay", true},
		{"mix-blend-mode", "darken", true},
		{"mix-blend-mode", "lighten", true},

		// Transform-box values
		{"transform-box", "view-box", true},
		{"transform-box", "fill-box", true},
		{"transform-box", "content-box", true},
		{"transform-box", "border-box", true},

		// Direction values
		{"direction", "ltr", true},
		{"direction", "rtl", true},

		// Writing mode values
		{"writing-mode", "horizontal-tb", true},
		{"writing-mode", "vertical-rl", true},
		{"writing-mode", "vertical-lr", true},

		// Caption side values
		{"caption-side", "top", true},
		{"caption-side", "bottom", true},

		// Empty cells values
		{"empty-cells", "show", true},
		{"empty-cells", "hide", true},

		// Unicode-bidi values
		{"unicode-bidi", "normal", true},
		{"unicode-bidi", "embed", true},
		{"unicode-bidi", "isolate", true},
		{"unicode-bidi", "bidi-override", true},
		{"unicode-bidi", "isolate-override", true},

		// Font-synthesis values
		{"font-synthesis", "none", true},
		{"font-synthesis", "weight", true},
		{"font-synthesis", "style", true},
		{"font-synthesis", "auto", true},
		{"font-synthesis", "invalid", false},

		// Hyphens values
		{"hyphens", "none", true},
		{"hyphens", "manual", true},
		{"hyphens", "auto", true},
		{"hyphens", "invalid", false},

		// Text-justify values
		{"text-justify", "auto", true},
		{"text-justify", "none", true},
		{"text-justify", "inter-word", true},
		{"text-justify", "inter-character", true},
		{"text-justify", "invalid", false},

		// Appearance values
		{"appearance", "auto", true},
		{"appearance", "none", true},
		{"appearance", "menulist", true},
	}

	for _, tt := range tests {
		t.Run(tt.property+" "+tt.value, func(t *testing.T) {
			result := Supports(tt.property, tt.value)
			if result != tt.expected {
				t.Errorf("Supports(%q, %q) = %v, want %v", tt.property, tt.value, result, tt.expected)
			}
		})
	}
}

func TestSupportsCondition(t *testing.T) {
	tests := []struct {
		condition string
		expected  bool
	}{
		{"(max-width: 768px)", true},
		{"(min-width: 320px)", true},
		{"(width: 100px)", true},
		{"(orientation: portrait)", true},
		{"(orientation: landscape)", true},
		{"(aspect-ratio: 16/9)", true},
		{"(color)", true},
		{"(hover: hover)", true},
		{"(pointer: fine)", true},
	}

	for _, tt := range tests {
		t.Run(tt.condition, func(t *testing.T) {
			result := Supports(tt.condition, "")
			if result != tt.expected {
				t.Errorf("Supports(%q, ) = %v, want %v", tt.condition, result, tt.expected)
			}
		})
	}
}

func TestMatchMedia(t *testing.T) {
	tests := []struct {
		query         string
		vw            float64
		vh            float64
		expectedMatch bool
	}{
		{"(max-width: 768px)", 600, 800, true},
		{"(max-width: 768px)", 800, 600, false},
		{"(min-width: 320px)", 400, 800, true},
		{"(min-width: 320px)", 300, 800, false},
		{"(orientation: portrait)", 600, 800, true},
		{"(orientation: portrait)", 800, 600, false},
		{"(orientation: landscape)", 800, 600, true},
		{"(orientation: landscape)", 600, 800, false},
		{"(min-width: 320px) and (max-width: 768px)", 400, 800, true},
		{"(min-width: 320px) and (max-width: 768px)", 200, 800, false},
		{"not (max-width: 768px)", 800, 600, true},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			mql := MatchMedia(tt.query, tt.vw, tt.vh)
			if mql.Matches != tt.expectedMatch {
				t.Errorf("MatchMedia(%q, %v, %v).Matches = %v, want %v",
					tt.query, tt.vw, tt.vh, mql.Matches, tt.expectedMatch)
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"10", 10, false},
		{"-10", -10, false},
		{"10.5", 10.5, false},
		{"-10.5", -10.5, false},
		{"0", 0, false},
		{"100%", 100, false},
		{"16px", 16, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseFloat(tt.input, 64)
			if (err != nil) != tt.hasError {
				t.Errorf("ParseFloat(%q) error = %v, hasError = %v", tt.input, err, tt.hasError)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseFloat(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
