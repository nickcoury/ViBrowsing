package render

import (
	"testing"

	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/layout"
)

func TestParseTransform(t *testing.T) {
	tests := []struct {
		input    string
		expected css.Transform
	}{
		// None
		{"none", css.Transform{Type: "none"}},
		{"", css.Transform{Type: "none"}},

		// Rotate
		{"rotate(45deg)", css.Transform{Type: "rotate", Rotate: 45}},
		{"rotate(90)", css.Transform{Type: "rotate", Rotate: 90}},
		{"rotate(-45deg)", css.Transform{Type: "rotate", Rotate: -45}},

		// Scale
		{"scale(1.5)", css.Transform{Type: "scale", ScaleX: 1.5, ScaleY: 1.5}},
		{"scale(2, 3)", css.Transform{Type: "scale", ScaleX: 2, ScaleY: 3}},
		{"scale(0.5)", css.Transform{Type: "scale", ScaleX: 0.5, ScaleY: 0.5}},

		// Translate
		{"translate(10px)", css.Transform{Type: "translate", TranslateX: 10}},
		{"translate(10px, 20px)", css.Transform{Type: "translate", TranslateX: 10, TranslateY: 20}},
		{"translate(-5px, 10px)", css.Transform{Type: "translate", TranslateX: -5, TranslateY: 10}},

		// Skew
		{"skew(10deg)", css.Transform{Type: "skew", SkewX: 10}},
		{"skew(10deg, 20deg)", css.Transform{Type: "skew", SkewX: 10, SkewY: 20}},
		{"skew(-5deg)", css.Transform{Type: "skew", SkewX: -5}},

		// Matrix
		{"matrix(1,0,0,1,0,0)", css.Transform{Type: "matrix", Matrix: [6]float64{1, 0, 0, 1, 0, 0}}},
		{"matrix(0.866,0.5,-0.5,0.866,0,0)", css.Transform{Type: "matrix", Matrix: [6]float64{0.866, 0.5, -0.5, 0.866, 0, 0}}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := css.ParseTransform(tt.input)
			if result.Type != tt.expected.Type {
				t.Errorf("ParseTransform(%q).Type = %q, want %q", tt.input, result.Type, tt.expected.Type)
			}
			switch tt.expected.Type {
			case "rotate":
				if result.Rotate != tt.expected.Rotate {
					t.Errorf("ParseTransform(%q).Rotate = %v, want %v", tt.input, result.Rotate, tt.expected.Rotate)
				}
			case "scale":
				if result.ScaleX != tt.expected.ScaleX || result.ScaleY != tt.expected.ScaleY {
					t.Errorf("ParseTransform(%q).Scale = (%v, %v), want (%v, %v)", tt.input, result.ScaleX, result.ScaleY, tt.expected.ScaleX, tt.expected.ScaleY)
				}
			case "translate":
				if result.TranslateX != tt.expected.TranslateX || result.TranslateY != tt.expected.TranslateY {
					t.Errorf("ParseTransform(%q).Translate = (%v, %v), want (%v, %v)", tt.input, result.TranslateX, result.TranslateY, tt.expected.TranslateX, tt.expected.TranslateY)
				}
			case "skew":
				if result.SkewX != tt.expected.SkewX || result.SkewY != tt.expected.SkewY {
					t.Errorf("ParseTransform(%q).Skew = (%v, %v), want (%v, %v)", tt.input, result.SkewX, result.SkewY, tt.expected.SkewX, tt.expected.SkewY)
				}
			case "matrix":
				for i := 0; i < 6; i++ {
					if result.Matrix[i] != tt.expected.Matrix[i] {
						t.Errorf("ParseTransform(%q).Matrix[%d] = %v, want %v", tt.input, i, result.Matrix[i], tt.expected.Matrix[i])
					}
				}
			}
		})
	}
}

func TestApplyTransform(t *testing.T) {
	// Test identity transform
	identity := css.Transform{Type: "none"}
	x, y := applyTransform(10, 20, identity, 100, 100)
	if x != 10 || y != 20 {
		t.Errorf("Identity transform: expected (10, 20), got (%v, %v)", x, y)
	}

	// Test rotate 90 degrees around center of 100x100 box
	// Center is (50, 50). Point (60, 50) should rotate to approximately (50, 60)
	rotate := css.Transform{Type: "rotate", Rotate: 90}
	x, y = applyTransform(60, 50, rotate, 100, 100)
	// Allow small floating point error
	if abs(x-50) > 0.01 || abs(y-60) > 0.01 {
		t.Errorf("Rotate 90 degrees: expected approximately (50, 60), got (%v, %v)", x, y)
	}

	// Test scale 2x around center
	scale := css.Transform{Type: "scale", ScaleX: 2, ScaleY: 2}
	// Point at center (50, 50) should stay at center
	x, y = applyTransform(50, 50, scale, 100, 100)
	if abs(x-50) > 0.01 || abs(y-50) > 0.01 {
		t.Errorf("Scale 2x at center: expected (50, 50), got (%v, %v)", x, y)
	}

	// Test translate
	translate := css.Transform{Type: "translate", TranslateX: 10, TranslateY: 20}
	x, y = applyTransform(5, 5, translate, 100, 100)
	// Translate is relative to center, then we add it
	if abs(x-15) > 0.01 || abs(y-25) > 0.01 {
		t.Errorf("Translate (10, 20): expected (15, 25), got (%v, %v)", x, y)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestFillRectTransformed(t *testing.T) {
	canvas := NewCanvas(200, 200)

	// Create a simple red rectangle
	red := css.Color{R: 255, G: 0, B: 0, A: 255}

	// Test with no transform
	noTransform := css.Transform{Type: "none"}
	canvas.FillRectTransformed(10, 10, 20, 20, red, noTransform)

	// Verify a pixel that should be red (10, 10)
	r, g, b, a := canvas.Pixels.At(10, 10).RGBA()
	if r == 0 && g == 0 && b == 0 && a == 0 {
		t.Error("Pixel (10, 10) should be set but is transparent/black")
	}

	// Test with scale transform
	canvas2 := NewCanvas(200, 200)
	scale := css.Transform{Type: "scale", ScaleX: 2, ScaleY: 2}
	canvas2.FillRectTransformed(10, 10, 10, 10, red, scale)
	// With scale 2x, pixel at (10,10) in 10x10 box around center should map to different location
	// Original box 10-20, 10-20 with center at 15,15 scaled 2x around center of 10x10 box
	// The scale should spread the drawing

	// Test with rotate transform
	canvas3 := NewCanvas(200, 200)
	rotate := css.Transform{Type: "rotate", Rotate: 45}
	canvas3.FillRectTransformed(80, 80, 40, 40, red, rotate)
	// This should apply a 45 degree rotation around the box center
}

func TestTransformRendering(t *testing.T) {
	// Test that a rotated rectangle renders correctly
	canvas := NewCanvas(300, 300)
	blue := css.Color{R: 0, G: 0, B: 255, A: 255}

	// Rotate a 50x50 box at (100, 100) by 45 degrees
	rotate := css.Transform{Type: "rotate", Rotate: 45}
	canvas.FillRectTransformed(100, 100, 50, 50, blue, rotate)

	// With a 45 degree rotation, pixels should be spread around the original position
	// Check that something was actually drawn outside the original axis-aligned box
	found := false
	for y := 80; y < 170; y++ {
		for x := 80; x < 170; x++ {
			if x >= 100 && x < 150 && y >= 100 && y < 150 {
				continue // Skip original bounding box
			}
			_, _, _, a := canvas.Pixels.At(x, y).RGBA()
			if a > 0 {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("Rotated rectangle should have pixels outside original bounding box")
	}
}

func TestOutlineRendering(t *testing.T) {
	canvas := NewCanvas(200, 200)

	// Create a box with outline styling
	box := &layout.Box{
		ContentX: 50,
		ContentY: 50,
		ContentW: 100,
		ContentH: 100,
		Style: map[string]string{
			"outline-width": "5px",
			"outline-style": "solid",
			"outline-color": "blue",
			"outline-offset": "3px",
		},
	}

	// Draw the box (this should draw outline as well)
	canvas.DrawBox(box)

	// Verify outline is drawn outside the content area (at x=42, y=42 to x=161, y=161)
	// outline offset 3px + outline width 5px = 8px outside content
	// Content at 50-150, outline starts at 50-5-3=42, ends at 150+5+3=158

	// Check that outline pixels are present at the expected outer positions
	// Top-left corner of outline (at x=42, y=42 area)
	foundOutline := false
	for y := 40; y < 50; y++ {
		for x := 40; x < 50; x++ {
			_, _, _, a := canvas.Pixels.At(x, y).RGBA()
			if a > 0 {
				foundOutline = true
				break
			}
		}
		if foundOutline {
			break
		}
	}
	if !foundOutline {
		t.Error("Outline should be drawn outside content area at top-left")
	}

	// Check that outline is NOT in the content area (50-150, 50-150) but is at edges
	// The outline should be visible at x=45 which is outside content but inside outline
	edgePixelFound := false
	for x := 42; x < 50; x++ {
		_, g, _, a := canvas.Pixels.At(x, 45).RGBA()
		if a > 0 && g == 0 {
			edgePixelFound = true
			break
		}
	}
	if !edgePixelFound {
		t.Error("Blue outline pixel should be visible near content edge")
	}
}

func TestOutlineNoDrawWhenDisabled(t *testing.T) {
	canvas := NewCanvas(200, 200)

	// Create a box with outline disabled and transparent background
	box := &layout.Box{
		ContentX: 50,
		ContentY: 50,
		ContentW: 100,
		ContentH: 100,
		Style: map[string]string{
			"outline-style": "none",
			"background":    "transparent",
		},
	}

	canvas.DrawBox(box)

	// No outline should be drawn - all pixels should be transparent
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			_, _, _, a := canvas.Pixels.At(x, y).RGBA()
			if a > 0 {
				t.Error("No pixels should be drawn when outline-style is none")
				return
			}
		}
	}
}
