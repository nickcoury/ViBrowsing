package css

import (
	"testing"
)

func TestParseProperty(t *testing.T) {
	// Clear registered properties before each test
	original := registeredProperties
	registeredProperties = make(map[string]RegisteredProperty)
	defer func() { registeredProperties = original }()

	tests := []struct {
		name      string
		input     string
		wantName  string
		wantSyntax string
		wantInherits bool
		wantInitial string
	}{
		{
			name:       "color property",
			input:     "@property --color { syntax: '<color>'; inherits: false; initial-value: 'red'; }",
			wantName:   "--color",
			wantSyntax: "<color>",
			wantInherits: false,
			wantInitial: "red",
		},
		{
			name:       "length property with inherits true",
			input:     "@property --size { syntax: '<length>'; inherits: true; initial-value: '10px'; }",
			wantName:   "--size",
			wantSyntax: "<length>",
			wantInherits: true,
			wantInitial: "10px",
		},
		{
			name:       "percentage property",
			input:     "@property --opacity { syntax: '<percentage>'; inherits: false; initial-value: '50%'; }",
			wantName:   "--opacity",
			wantSyntax: "<percentage>",
			wantInherits: false,
			wantInitial: "50%",
		},
		{
			name:       "number property",
			input:     "@property --count { syntax: '<number>'; inherits: false; initial-value: '0'; }",
			wantName:   "--count",
			wantSyntax: "<number>",
			wantInherits: false,
			wantInitial: "0",
		},
		{
			name:       "integer property",
			input:     "@property --index { syntax: '<integer>'; inherits: false; initial-value: '1'; }",
			wantName:   "--index",
			wantSyntax: "<integer>",
			wantInherits: false,
			wantInitial: "1",
		},
		{
			name:       "with single quotes",
			input:     "@property --my-color { syntax: '<color>'; inherits: false; initial-value: '#ff0000'; }",
			wantName:   "--my-color",
			wantSyntax: "<color>",
			wantInherits: false,
			wantInitial: "#ff0000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Parse(tc.input)
			prop, ok := registeredProperties[tc.wantName]
			if !ok {
				t.Errorf("property %q not registered", tc.wantName)
				return
			}
			if prop.Syntax != tc.wantSyntax {
				t.Errorf("syntax=%q want %q", prop.Syntax, tc.wantSyntax)
			}
			if prop.Inherits != tc.wantInherits {
				t.Errorf("inherits=%v want %v", prop.Inherits, tc.wantInherits)
			}
			if prop.InitialValue != tc.wantInitial {
				t.Errorf("initial-value=%q want %q", prop.InitialValue, tc.wantInitial)
			}
		})
	}
}

func TestParseProperty_Invalid(t *testing.T) {
	// Clear registered properties before each test
	original := registeredProperties
	registeredProperties = make(map[string]RegisteredProperty)
	defer func() { registeredProperties = original }()

	invalidInputs := []string{
		"",                               // empty
		"@property {}",                    // no name
		"@property -- { syntax: '<color>'; inherits: false; initial-value: 'red'; }", // no name
		"@property not-starting-with-dash { syntax: '<color>'; inherits: false; initial-value: 'red'; }", // doesn't start with --
		"@property --color { syntax: '<color>'; initial-value: 'red'; }", // missing inherits
		"@property --color { inherits: false; initial-value: 'red'; }", // missing syntax
		"@property --color { syntax: '<color>'; inherits: false; }", // missing initial-value
	}

	for _, input := range invalidInputs {
		t.Run(input[:min(len(input), 30)], func(t *testing.T) {
			registeredProperties = make(map[string]RegisteredProperty)
			Parse(input)
			if len(registeredProperties) > 0 {
				t.Errorf("expected no properties for invalid input %q, got %v", input, registeredProperties)
			}
		})
	}
}

func TestParseProperty_InCSSStylesheet(t *testing.T) {
	// Clear registered properties before each test
	original := registeredProperties
	registeredProperties = make(map[string]RegisteredProperty)
	defer func() { registeredProperties = original }()

	css := `
		@property --brand-color {
			syntax: '<color>';
			inherits: false;
			initial-value: '#007bff';
		}
		@property --spacing {
			syntax: '<length>';
			inherits: true;
			initial-value: '8px';
		}
		div { color: var(--brand-color); }
		p { padding: var(--spacing); }
	`

	rules := Parse(css)

	// Check properties are registered
	if len(registeredProperties) != 2 {
		t.Errorf("expected 2 registered properties, got %d", len(registeredProperties))
	}

	brand, ok := registeredProperties["--brand-color"]
	if !ok {
		t.Error("--brand-color not registered")
	} else {
		if brand.Syntax != "<color>" {
			t.Errorf("--brand-color syntax=%q want %q", brand.Syntax, "<color>")
		}
		if brand.Inherits != false {
			t.Error("--brand-color inherits should be false")
		}
		if brand.InitialValue != "#007bff" {
			t.Errorf("--brand-color initial=%q want %q", brand.InitialValue, "#007bff")
		}
	}

	spacing, ok := registeredProperties["--spacing"]
	if !ok {
		t.Error("--spacing not registered")
	} else {
		if spacing.Syntax != "<length>" {
			t.Errorf("--spacing syntax=%q want %q", spacing.Syntax, "<length>")
		}
		if spacing.Inherits != true {
			t.Error("--spacing inherits should be true")
		}
		if spacing.InitialValue != "8px" {
			t.Errorf("--spacing initial=%q want %q", spacing.InitialValue, "8px")
		}
	}

	// Check rules are still parsed correctly
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}

func TestGetPropertyValue_RegisteredProperty(t *testing.T) {
	// Clear registered properties before each test
	original := registeredProperties
	registeredProperties = make(map[string]RegisteredProperty)
	defer func() { registeredProperties = original }()

	// Register a custom property
	registeredProperties["--brand-color"] = RegisteredProperty{
		Name:         "--brand-color",
		Syntax:       "<color>",
		Inherits:     false,
		InitialValue: "#007bff",
	}

	props := map[string]string{
		"--other": "purple",
	}

	// GetPropertyValue should return the initial-value for registered properties
	val := GetPropertyValue(props, "--brand-color")
	if val != "#007bff" {
		t.Errorf("got %q want %q", val, "#007bff")
	}

	// GetPropertyValue should return the value from props for unregistered properties
	val = GetPropertyValue(props, "--other")
	if val != "purple" {
		t.Errorf("got %q want %q", val, "purple")
	}

	// GetPropertyValue should return empty string for unknown properties
	val = GetPropertyValue(props, "--unknown")
	if val != "" {
		t.Errorf("got %q want empty string", val)
	}
}

func TestGetPropertyValue_EmptyProps(t *testing.T) {
	// Clear registered properties before each test
	original := registeredProperties
	registeredProperties = make(map[string]RegisteredProperty)
	defer func() { registeredProperties = original }()

	registeredProperties["--test"] = RegisteredProperty{
		Name:         "--test",
		Syntax:       "<number>",
		Inherits:     false,
		InitialValue: "42",
	}

	props := map[string]string{}

	// GetPropertyValue should return the initial-value even with empty props map
	val := GetPropertyValue(props, "--test")
	if val != "42" {
		t.Errorf("got %q want %q", val, "42")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}