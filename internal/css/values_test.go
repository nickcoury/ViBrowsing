package css

import (
	"testing"
)

func TestParseBlendMode(t *testing.T) {
	tests := []struct {
		input    string
		expected BlendModeType
	}{
		{"normal", BlendModeNormal},
		{"multiply", BlendModeMultiply},
		{"screen", BlendModeScreen},
		{"overlay", BlendModeOverlay},
		{"darken", BlendModeDarken},
		{"lighten", BlendModeLighten},
		{"color-dodge", BlendModeColorDodge},
		{"color-burn", BlendModeColorBurn},
		{"hard-light", BlendModeHardLight},
		{"soft-light", BlendModeSoftLight},
		{"difference", BlendModeDifference},
		{"exclusion", BlendModeExclusion},
		{"hue", BlendModeHue},
		{"saturation", BlendModeSaturation},
		{"color", BlendModeColor},
		{"luminosity", BlendModeLuminosity},
		// Case insensitivity
		{"MULTIPLY", BlendModeMultiply},
		{"Screen", BlendModeScreen},
		{"OVERLAY", BlendModeOverlay},
		// Invalid values fall back to normal
		{"invalid", BlendModeNormal},
		{"", BlendModeNormal},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseBlendMode(tt.input)
			if result != tt.expected {
				t.Errorf("ParseBlendMode(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidBlendMode(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"normal", true},
		{"multiply", true},
		{"screen", true},
		{"overlay", true},
		{"darken", true},
		{"lighten", true},
		{"color-dodge", true},
		{"color-burn", true},
		{"hard-light", true},
		{"soft-light", true},
		{"difference", true},
		{"exclusion", true},
		{"hue", true},
		{"saturation", true},
		{"color", true},
		{"luminosity", true},
		// Case insensitivity
		{"MULTIPLY", true},
		{"SCREEN", true},
		// Invalid
		{"invalid", false},
		{"", false},
		{"rainbow", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsValidBlendMode(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidBlendMode(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseColorScheme(t *testing.T) {
	tests := []struct {
		input       string
		wantLight   bool
		wantDark    bool
		wantOnly    bool
	}{
		{"normal", false, false, false},
		{"light", true, false, false},
		{"dark", false, true, false},
		{"only", false, false, true},
		{"light dark", true, true, false},
		{"dark light", true, true, false},
		{"", false, false, false},
		{"light light", true, false, false},  // duplicates handled
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseColorScheme(tt.input)
			if result.Light != tt.wantLight {
				t.Errorf("ParseColorScheme(%q).Light = %v, want %v", tt.input, result.Light, tt.wantLight)
			}
			if result.Dark != tt.wantDark {
				t.Errorf("ParseColorScheme(%q).Dark = %v, want %v", tt.input, result.Dark, tt.wantDark)
			}
			if result.Only != tt.wantOnly {
				t.Errorf("ParseColorScheme(%q).Only = %v, want %v", tt.input, result.Only, tt.wantOnly)
			}
		})
	}
}

func TestParseScrollbarWidth(t *testing.T) {
	tests := []struct {
		input    string
		expected ScrollbarWidthType
	}{
		{"auto", ScrollbarWidthAuto},
		{"thin", ScrollbarWidthThin},
		{"none", ScrollbarWidthNone},
		// Case insensitivity
		{"AUTO", ScrollbarWidthAuto},
		{"Thin", ScrollbarWidthThin},
		{"NONE", ScrollbarWidthNone},
		// Invalid falls back to auto
		{"invalid", ScrollbarWidthAuto},
		{"", ScrollbarWidthAuto},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseScrollbarWidth(tt.input)
			if result != tt.expected {
				t.Errorf("ParseScrollbarWidth(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseScrollbarColor(t *testing.T) {
	tests := []struct {
		input            string
		wantThumbR, wantThumbG, wantThumbB uint8
		wantTrackR, wantTrackG, wantTrackB uint8
		wantHasTrack bool
	}{
		{"auto", 0, 0, 0, 0, 0, 0, false},
		{"", 0, 0, 0, 0, 0, 0, false},
		{"red", 255, 0, 0, 0, 0, 0, false},
		{"#333333", 51, 51, 51, 0, 0, 0, false},
		{"red blue", 255, 0, 0, 0, 0, 255, true},
		// Note: 3-digit hex (#123) is NOT expanded by ParseColor - that's a separate bug
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseScrollbarColor(tt.input)
			if result.Thumb.R != tt.wantThumbR || result.Thumb.G != tt.wantThumbG || result.Thumb.B != tt.wantThumbB {
				t.Errorf("ParseScrollbarColor(%q).Thumb = (%d,%d,%d), want (%d,%d,%d)",
					tt.input, result.Thumb.R, result.Thumb.G, result.Thumb.B, tt.wantThumbR, tt.wantThumbG, tt.wantThumbB)
			}
			if tt.wantHasTrack {
				if result.Track.R != tt.wantTrackR || result.Track.G != tt.wantTrackG || result.Track.B != tt.wantTrackB {
					t.Errorf("ParseScrollbarColor(%q).Track = (%d,%d,%d), want (%d,%d,%d)",
						tt.input, result.Track.R, result.Track.G, result.Track.B, tt.wantTrackR, tt.wantTrackG, tt.wantTrackB)
				}
			}
		})
	}
}

func TestParseAccentColor(t *testing.T) {
	tests := []struct {
		input           string
		wantIsAuto      bool
		wantR, wantG, wantB uint8
	}{
		{"auto", true, 0, 0, 0},
		{"", true, 0, 0, 0},
		{"red", false, 255, 0, 0},
		{"#0000FF", false, 0, 0, 255},
		{"rgb(0, 255, 0)", false, 0, 255, 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseAccentColor(tt.input)
			if result.IsAuto != tt.wantIsAuto {
				t.Errorf("ParseAccentColor(%q).IsAuto = %v, want %v", tt.input, result.IsAuto, tt.wantIsAuto)
			}
			if !tt.wantIsAuto {
				if result.Color.R != tt.wantR || result.Color.G != tt.wantG || result.Color.B != tt.wantB {
					t.Errorf("ParseAccentColor(%q).Color = (%d,%d,%d), want (%d,%d,%d)",
						tt.input, result.Color.R, result.Color.G, result.Color.B, tt.wantR, tt.wantG, tt.wantB)
				}
			}
		})
	}
}

func TestColorSchemeTypeDefaults(t *testing.T) {
	// Test that default ColorSchemeType has all false values
	cs := ColorSchemeType{}
	if cs.Light || cs.Dark || cs.Only {
		t.Error("Default ColorSchemeType should have all false values")
	}

	// Test that empty string returns defaults
	cs2 := ParseColorScheme("")
	if cs2.Light || cs2.Dark || cs2.Only {
		t.Error("ParseColorScheme(\"\") should return defaults")
	}
}

func TestScrollbarColorAutoDefaults(t *testing.T) {
	// Test that auto returns zero values
	sc := ParseScrollbarColor("auto")
	if sc.Thumb.A != 0 || sc.Track.A != 0 {
		t.Error("Auto scrollbar color should have zero alpha values")
	}
}