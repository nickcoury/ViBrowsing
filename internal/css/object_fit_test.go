package css

import (
	"testing"
)

// TestObjectFitValues tests that object-fit parsing accepts valid values.
func TestObjectFitValues(t *testing.T) {
	// These are tested via applyDecl in style.go
	// This test verifies the values are handled correctly
	validValues := []string{"fill", "contain", "cover", "none", "scale-down"}
	for _, v := range validValues {
		props := map[string]string{}
		applyDecl(props, Declaration{Property: "object-fit", Value: v})
		if props["object-fit"] != v {
			t.Errorf("object-fit: got %q want %q", props["object-fit"], v)
		}
	}
}

// TestObjectPositionValues tests that object-position parsing accepts valid values.
func TestObjectPositionValues(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"left top"},
		{"left center"},
		{"left bottom"},
		{"right top"},
		{"right center"},
		{"right bottom"},
		{"center top"},
		{"center center"},
		{"center bottom"},
		{"50% 50%"},
		{"0% 0%"},
		{"100% 100%"},
		{"10px 20px"},
		{"left 10px"},
		{"10px bottom"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			props := map[string]string{}
			applyDecl(props, Declaration{Property: "object-position", Value: tc.input})
			if props["object-position"] != tc.input {
				t.Errorf("object-position: got %q want %q", props["object-position"], tc.input)
			}
		})
	}
}

// TestAspectRatioViaStyle tests that aspect-ratio is correctly applied via style.
func TestAspectRatioViaStyle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{"16/9", "16/9", 16.0 / 9.0},
		{"4/3", "4/3", 4.0 / 3.0},
		{"1/1", "1/1", 1.0},
		{"auto", "auto", 0},
		{"empty", "", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseAspectRatio(tc.input)
			if got != tc.want {
				t.Errorf("ParseAspectRatio(%q): got %f want %f", tc.input, got, tc.want)
			}
		})
	}
}