package css

import (
	"testing"
)

func TestParseLayer(t *testing.T) {
	tests := []struct {
		name     string
		css      string
		wantLen  int
		wantRule []string // selectors that should be in the CSS
	}{
		{
			name:     "named layer",
			css:      "@layer utilities { .btn { color: red; } }",
			wantLen:  1,
			wantRule: []string{".btn"},
		},
		{
			name:     "anonymous layer",
			css:      "@layer { .btn { color: red; } }",
			wantLen:  1,
			wantRule: []string{".btn"},
		},
		{
			name:     "multiple layers",
			css:      "@layer reset { * { margin: 0; } } @layer utilities { .btn { color: red; } }",
			wantLen:  2,
			wantRule: []string{"*", ".btn"},
		},
		{
			name:     "nested rules in layer",
			css:      "@layer components { .card { padding: 10px; } .card .title { font-size: 16px; } }",
			wantLen:  2,
			wantRule: []string{".card", ".card .title"},
		},
		{
			name:     "mixed layered and unlayered",
			css:      ".global { color: blue; } @layer utilities { .btn { color: red; } }",
			wantLen:  2,
			wantRule: []string{".global", ".btn"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset anonymous layer counter for each test
			anonymousLayerCounter = 0

			rules := Parse(tt.css)
			if len(rules) != tt.wantLen {
				t.Errorf("Parse() got %d rules, want %d", len(rules), tt.wantLen)
			}

			for i, want := range tt.wantRule {
				if i >= len(rules) {
					t.Errorf("Parse() rule %d: want selector %q, but no more rules", i, want)
					continue
				}
				if rules[i].Selector != want {
					t.Errorf("Parse() rule %d: got selector %q, want %q", i, rules[i].Selector, want)
				}
			}
		})
	}
}

func TestParseLayerAssignsLayerName(t *testing.T) {
	anonymousLayerCounter = 0

	css := "@layer utilities { .btn { color: red; } }"
	rules := Parse(css)

	if len(rules) != 1 {
		t.Fatalf("Parse() got %d rules, want 1", len(rules))
	}

	if rules[0].Layer != "utilities" {
		t.Errorf("Parse() got layer %q, want %q", rules[0].Layer, "utilities")
	}
}

func TestParseAnonymousLayerAssignsUniqueName(t *testing.T) {
	anonymousLayerCounter = 0

	css := "@layer { .btn { color: red; } } @layer { .link { color: blue; } }"
	rules := Parse(css)

	if len(rules) != 2 {
		t.Fatalf("Parse() got %d rules, want 2", len(rules))
	}

	// Anonymous layers should have unique names
	if rules[0].Layer == "" || rules[1].Layer == "" {
		t.Errorf("Anonymous layers should have non-empty layer names")
	}

	if rules[0].Layer == rules[1].Layer {
		t.Errorf("Two anonymous @layer blocks should have different names, got %q for both", rules[0].Layer)
	}
}

func TestLayerOrder(t *testing.T) {
	tests := []struct {
		name     string
		layers   []string
		query    string
		expected int
	}{
		{"first layer", []string{"reset", "utilities"}, "reset", 0},
		{"second layer", []string{"reset", "utilities"}, "utilities", 1},
		{"unlayered", []string{"reset", "utilities"}, "", 2}, // unlayered gets highest priority
		{"unknown layer", []string{"reset", "utilities"}, "unknown", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lo := &LayerOrder{}
			for _, l := range tt.layers {
				lo.RegisterLayer(l)
			}

			got := lo.GetLayerPriority(tt.query)
			if got != tt.expected {
				t.Errorf("GetLayerPriority(%q) = %d, want %d", tt.query, got, tt.expected)
			}
		})
	}
}

func TestBuildLayerOrder(t *testing.T) {
	anonymousLayerCounter = 0

	css := "@layer reset { * { margin: 0; } } @layer utilities { .btn { color: red; } } @layer components { .card { padding: 10px; } }"
	rules := Parse(css)

	lo := BuildLayerOrder(rules)

	// Should track all three layers
	expected := []string{"reset", "utilities", "components"}
	for i, name := range expected {
		if lo.layers[i] != name {
			t.Errorf("BuildLayerOrder() layers[%d] = %q, want %q", i, lo.layers[i], name)
		}
	}
}

func TestSortRulesByLayer(t *testing.T) {
	anonymousLayerCounter = 0

	// CSS where reset is declared first, then utilities - but reset should have lower priority
	css := "@layer reset { .a { color: red; } } @layer utilities { .b { color: blue; } } .c { color: green; }"
	rules := Parse(css)

	if len(rules) != 3 {
		t.Fatalf("Parse() got %d rules, want 3", len(rules))
	}

	lo := BuildLayerOrder(rules)
	sorted := SortRulesByLayer(rules, lo)

	// Unlayered (.c) should be last
	if sorted[len(sorted)-1].Selector != ".c" {
		t.Errorf("SortRulesByLayer() last rule should be unlayered .c, got %q", sorted[len(sorted)-1].Selector)
	}

	// Within layered rules, order should be: reset (lower) before utilities (higher)
	var resetIdx, utilIdx int
	for i, r := range sorted {
		if r.Selector == ".a" {
			resetIdx = i
		}
		if r.Selector == ".b" {
			utilIdx = i
		}
	}

	if resetIdx >= utilIdx {
		t.Errorf("SortRulesByLayer() reset (.a) should come before utilities (.b), got .a at %d, .b at %d", resetIdx, utilIdx)
	}
}

func TestComputeStyleWithLayers(t *testing.T) {
	anonymousLayerCounter = 0

	// CSS: later layer should override earlier layer for same property
	css := `
		@layer base {
			.test {
				color: red;
				font-size: 14px;
			}
		}
		@layer overrides {
			.test {
				color: blue;
			}
		}
	`

	rules := Parse(css)
	if len(rules) != 2 {
		t.Fatalf("Parse() got %d rules, want 2", len(rules))
	}

	// Simulate applying rules
	props := map[string]string{}
	for _, rule := range rules {
		if rule.Selector == ".test" {
			for _, decl := range rule.Declarations {
				props[decl.Property] = decl.Value
			}
		}
	}

	// Later layer (overrides) should win for color
	if props["color"] != "blue" {
		t.Errorf("Later layer should override: got color=%q, want blue", props["color"])
	}

	// font-size should still be from base layer
	if props["font-size"] != "14px" {
		t.Errorf("Earlier layer value should be preserved: got font-size=%q, want 14px", props["font-size"])
	}
}

func TestSupportsLayer(t *testing.T) {
	// Test CSS.Supports("CSS", "@layer")
	if !Supports("CSS", "@layer") {
		t.Error("Supports(\"CSS\", \"@layer\") should return true")
	}

	// Case insensitive test
	if !Supports("css", "@layer") {
		t.Error("Supports(\"css\", \"@layer\") should return true (case insensitive)")
	}
}
