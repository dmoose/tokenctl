// tokenctl/pkg/tokens/property_test.go

package tokens

import (
	"testing"
)

func TestCSSPropertySyntax(t *testing.T) {
	tests := []struct {
		tokenType string
		want      string
	}{
		{"color", "<color>"},
		{"dimension", "<length>"},
		{"number", "<number>"},
		{"duration", "<time>"},
		{"effect", "<integer>"},
		{"fontFamily", ""},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.tokenType, func(t *testing.T) {
			got := CSSPropertySyntax(tt.tokenType)
			if got != tt.want {
				t.Errorf("CSSPropertySyntax(%q) = %q, want %q", tt.tokenType, got, tt.want)
			}
		})
	}
}

func TestExtractPropertyTokens_Basic(t *testing.T) {
	dict := &Dictionary{
		Root: map[string]any{
			"color": map[string]any{
				"$type": "color",
				"primary": map[string]any{
					"$value":    "oklch(50% 0.2 250)",
					"$property": true,
				},
				"secondary": map[string]any{
					"$value": "#8b5cf6",
					// No $property - should not be included
				},
			},
		},
	}

	resolved := map[string]any{
		"color.primary":   "oklch(50% 0.2 250)",
		"color.secondary": "#8b5cf6",
	}

	properties := ExtractPropertyTokens(dict, resolved)

	if len(properties) != 1 {
		t.Fatalf("expected 1 property token, got %d", len(properties))
	}

	prop := properties[0]
	if prop.Path != "color.primary" {
		t.Errorf("expected path 'color.primary', got %q", prop.Path)
	}
	if prop.CSSName != "--color-primary" {
		t.Errorf("expected CSSName '--color-primary', got %q", prop.CSSName)
	}
	if prop.CSSSyntax != "<color>" {
		t.Errorf("expected CSSSyntax '<color>', got %q", prop.CSSSyntax)
	}
	expectedInitial := "oklch(50% 0.2 250)"
	if prop.InitialValue != expectedInitial {
		t.Errorf("expected InitialValue %q, got %q", expectedInitial, prop.InitialValue)
	}
	if !prop.Inherits {
		t.Error("expected Inherits to be true")
	}
}

func TestExtractPropertyTokens_InheritsFromParentType(t *testing.T) {
	dict := &Dictionary{
		Root: map[string]any{
			"spacing": map[string]any{
				"$type": "dimension",
				"sm": map[string]any{
					"$value":    "0.5rem",
					"$property": true,
				},
				"md": map[string]any{
					"$value":    "1rem",
					"$property": true,
				},
			},
		},
	}

	resolved := map[string]any{
		"spacing.sm": "0.5rem",
		"spacing.md": "1rem",
	}

	properties := ExtractPropertyTokens(dict, resolved)

	if len(properties) != 2 {
		t.Fatalf("expected 2 property tokens, got %d", len(properties))
	}

	for _, prop := range properties {
		if prop.CSSSyntax != "<length>" {
			t.Errorf("expected CSSSyntax '<length>' for %s, got %s", prop.Path, prop.CSSSyntax)
		}
	}
}

func TestExtractPropertyTokens_CustomInherits(t *testing.T) {
	dict := &Dictionary{
		Root: map[string]any{
			"timing": map[string]any{
				"fast": map[string]any{
					"$value": "150ms",
					"$type":  "duration",
					"$property": map[string]any{
						"inherits": false,
					},
				},
			},
		},
	}

	resolved := map[string]any{
		"timing.fast": "150ms",
	}

	properties := ExtractPropertyTokens(dict, resolved)

	if len(properties) != 1 {
		t.Fatalf("expected 1 property token, got %d", len(properties))
	}

	if properties[0].Inherits {
		t.Error("expected Inherits to be false")
	}
}

func TestExtractPropertyTokens_PropertyFalse(t *testing.T) {
	dict := &Dictionary{
		Root: map[string]any{
			"color": map[string]any{
				"primary": map[string]any{
					"$value":    "#3b82f6",
					"$type":     "color",
					"$property": false,
				},
			},
		},
	}

	resolved := map[string]any{
		"color.primary": "#3b82f6",
	}

	properties := ExtractPropertyTokens(dict, resolved)

	if len(properties) != 0 {
		t.Errorf("expected 0 property tokens for $property: false, got %d", len(properties))
	}
}

func TestExtractPropertyTokens_SkipsUnmappableTypes(t *testing.T) {
	dict := &Dictionary{
		Root: map[string]any{
			"font": map[string]any{
				"family": map[string]any{
					"$value":    []any{"Inter", "sans-serif"},
					"$type":     "fontFamily",
					"$property": true,
				},
			},
		},
	}

	resolved := map[string]any{
		"font.family": []any{"Inter", "sans-serif"},
	}

	properties := ExtractPropertyTokens(dict, resolved)

	if len(properties) != 0 {
		t.Errorf("expected 0 property tokens for fontFamily (no CSS syntax), got %d", len(properties))
	}
}

func TestExtractPropertyTokens_NoTypeSkipped(t *testing.T) {
	dict := &Dictionary{
		Root: map[string]any{
			"custom": map[string]any{
				"value": map[string]any{
					"$value":    "something",
					"$property": true,
					// No $type - should be skipped
				},
			},
		},
	}

	resolved := map[string]any{
		"custom.value": "something",
	}

	properties := ExtractPropertyTokens(dict, resolved)

	if len(properties) != 0 {
		t.Errorf("expected 0 property tokens for token without type, got %d", len(properties))
	}
}

func TestExtractPropertyTokens_NumericValues(t *testing.T) {
	dict := &Dictionary{
		Root: map[string]any{
			"opacity": map[string]any{
				"$type": "number",
				"half": map[string]any{
					"$value":    0.5,
					"$property": true,
				},
				"full": map[string]any{
					"$value":    1.0,
					"$property": true,
				},
			},
			"effect": map[string]any{
				"$type": "effect",
				"depth": map[string]any{
					"$value":    1,
					"$property": true,
				},
			},
		},
	}

	resolved := map[string]any{
		"opacity.half": 0.5,
		"opacity.full": 1.0,
		"effect.depth": 1,
	}

	properties := ExtractPropertyTokens(dict, resolved)

	if len(properties) != 3 {
		t.Fatalf("expected 3 property tokens, got %d", len(properties))
	}

	// Check numeric formatting
	for _, prop := range properties {
		switch prop.Path {
		case "opacity.half":
			if prop.InitialValue != "0.5" {
				t.Errorf("expected InitialValue '0.5', got %q", prop.InitialValue)
			}
			if prop.CSSSyntax != "<number>" {
				t.Errorf("expected CSSSyntax '<number>', got %q", prop.CSSSyntax)
			}
		case "opacity.full":
			if prop.InitialValue != "1" {
				t.Errorf("expected InitialValue '1', got %q", prop.InitialValue)
			}
		case "effect.depth":
			if prop.InitialValue != "1" {
				t.Errorf("expected InitialValue '1', got %q", prop.InitialValue)
			}
			if prop.CSSSyntax != "<integer>" {
				t.Errorf("expected CSSSyntax '<integer>', got %q", prop.CSSSyntax)
			}
		}
	}
}

func TestFormatInitialValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"string", "oklch(50% 0.2 250)", "oklch(50% 0.2 250)"},
		{"integer", 42, "42"},
		{"float whole", 1.0, "1"},
		{"float decimal", 0.5, "0.5"},
		{"array string", []any{"a", "b", "c"}, "a, b, c"},
		{"array mixed", []any{"Inter", "sans-serif"}, "Inter, sans-serif"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatInitialValue(tt.value)
			if got != tt.want {
				t.Errorf("formatInitialValue(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
