// tokctl/pkg/tokens/scale_test.go

package tokens

import (
	"strings"
	"testing"
)

func TestExpandScales(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]interface{}
		wantTokens    []string // Token paths that should exist after expansion
		wantNotTokens []string // Token paths that should NOT exist
		wantValues    map[string]string
		wantErr       bool
	}{
		{
			name: "basic scale expansion",
			input: map[string]interface{}{
				"size": map[string]interface{}{
					"field": map[string]interface{}{
						"$value": "2.5rem",
						"$scale": map[string]interface{}{
							"xs": 0.6,
							"sm": 0.8,
							"md": 1.0,
							"lg": 1.2,
							"xl": 1.4,
						},
					},
				},
			},
			wantTokens: []string{
				"size.field",
				"size.field-xs",
				"size.field-sm",
				"size.field-md",
				"size.field-lg",
				"size.field-xl",
			},
			wantValues: map[string]string{
				"size.field-xs": "calc({size.field} * 0.6)",
				"size.field-sm": "calc({size.field} * 0.8)",
				"size.field-md": "{size.field}",
				"size.field-lg": "calc({size.field} * 1.2)",
				"size.field-xl": "calc({size.field} * 1.4)",
			},
		},
		{
			name: "scale with type inheritance",
			input: map[string]interface{}{
				"spacing": map[string]interface{}{
					"base": map[string]interface{}{
						"$value": "1rem",
						"$type":  "dimension",
						"$scale": map[string]interface{}{
							"sm": 0.5,
							"lg": 2.0,
						},
					},
				},
			},
			wantTokens: []string{
				"spacing.base",
				"spacing.base-sm",
				"spacing.base-lg",
			},
		},
		{
			name: "nested tokens with scales",
			input: map[string]interface{}{
				"components": map[string]interface{}{
					"button": map[string]interface{}{
						"height": map[string]interface{}{
							"$value": "40px",
							"$scale": map[string]interface{}{
								"sm": 0.8,
								"lg": 1.2,
							},
						},
					},
				},
			},
			wantTokens: []string{
				"components.button.height",
				"components.button.height-sm",
				"components.button.height-lg",
			},
		},
		{
			name: "token without scale unchanged",
			input: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "#3b82f6",
					},
				},
			},
			wantTokens: []string{
				"color.primary",
			},
			wantNotTokens: []string{
				"color.primary-sm",
				"color.primary-lg",
			},
		},
		{
			name: "scale removed after expansion",
			input: map[string]interface{}{
				"size": map[string]interface{}{
					"base": map[string]interface{}{
						"$value": "1rem",
						"$scale": map[string]interface{}{
							"sm": 0.5,
						},
					},
				},
			},
			wantTokens: []string{
				"size.base",
				"size.base-sm",
			},
		},
		{
			name: "invalid scale type",
			input: map[string]interface{}{
				"size": map[string]interface{}{
					"base": map[string]interface{}{
						"$value": "1rem",
						"$scale": "not-an-object",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid scale factor",
			input: map[string]interface{}{
				"size": map[string]interface{}{
					"base": map[string]interface{}{
						"$value": "1rem",
						"$scale": map[string]interface{}{
							"sm": "not-a-number",
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict := &Dictionary{
				Root:        tt.input,
				SourceFiles: make(map[string]string),
			}

			err := ExpandScales(dict)

			if tt.wantErr {
				if err == nil {
					t.Error("ExpandScales() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ExpandScales() unexpected error: %v", err)
			}

			// Check expected tokens exist
			for _, path := range tt.wantTokens {
				if !tokenExists(dict.Root, path) {
					t.Errorf("Expected token %q to exist after expansion", path)
				}
			}

			// Check unexpected tokens don't exist
			for _, path := range tt.wantNotTokens {
				if tokenExists(dict.Root, path) {
					t.Errorf("Token %q should not exist after expansion", path)
				}
			}

			// Check expected values
			for path, wantValue := range tt.wantValues {
				value := getTokenValue(dict.Root, path)
				if value != wantValue {
					t.Errorf("Token %q value = %q, want %q", path, value, wantValue)
				}
			}
		})
	}
}

func TestExpandScales_ScaleRemovedFromBaseToken(t *testing.T) {
	dict := &Dictionary{
		Root: map[string]interface{}{
			"size": map[string]interface{}{
				"base": map[string]interface{}{
					"$value": "1rem",
					"$scale": map[string]interface{}{
						"sm": 0.5,
					},
				},
			},
		},
		SourceFiles: make(map[string]string),
	}

	err := ExpandScales(dict)
	if err != nil {
		t.Fatalf("ExpandScales() error: %v", err)
	}

	// Check that $scale is removed from the base token
	baseToken := dict.Root["size"].(map[string]interface{})["base"].(map[string]interface{})
	if _, hasScale := baseToken["$scale"]; hasScale {
		t.Error("$scale should be removed from base token after expansion")
	}

	// But $value should still exist
	if _, hasValue := baseToken["$value"]; !hasValue {
		t.Error("$value should still exist on base token")
	}
}

func TestExpandScales_SourceFileTracking(t *testing.T) {
	dict := &Dictionary{
		Root: map[string]interface{}{
			"size": map[string]interface{}{
				"field": map[string]interface{}{
					"$value": "2.5rem",
					"$scale": map[string]interface{}{
						"sm": 0.8,
						"lg": 1.2,
					},
				},
			},
		},
		SourceFiles: map[string]string{
			"size.field": "tokens/sizing.json",
		},
	}

	err := ExpandScales(dict)
	if err != nil {
		t.Fatalf("ExpandScales() error: %v", err)
	}

	// Check that expanded tokens inherit the source file
	if dict.SourceFiles["size.field-sm"] != "tokens/sizing.json" {
		t.Errorf("size.field-sm source file = %q, want %q",
			dict.SourceFiles["size.field-sm"], "tokens/sizing.json")
	}

	if dict.SourceFiles["size.field-lg"] != "tokens/sizing.json" {
		t.Errorf("size.field-lg source file = %q, want %q",
			dict.SourceFiles["size.field-lg"], "tokens/sizing.json")
	}
}

func TestExpandScales_DescriptionInheritance(t *testing.T) {
	dict := &Dictionary{
		Root: map[string]interface{}{
			"size": map[string]interface{}{
				"field": map[string]interface{}{
					"$value":       "40px",
					"$description": "Base field size",
					"$scale": map[string]interface{}{
						"sm": 0.8,
					},
				},
			},
		},
		SourceFiles: make(map[string]string),
	}

	err := ExpandScales(dict)
	if err != nil {
		t.Fatalf("ExpandScales() error: %v", err)
	}

	// Check that expanded token has a description
	smToken := dict.Root["size"].(map[string]interface{})["field-sm"].(map[string]interface{})
	desc, ok := smToken["$description"].(string)
	if !ok {
		t.Error("Expanded token should have a description")
		return
	}

	if !strings.Contains(desc, "sm") && !strings.Contains(desc, "scale") {
		t.Errorf("Expanded token description should mention scale, got: %q", desc)
	}
}

func TestStandardScale(t *testing.T) {
	scale := StandardScale()

	expectedKeys := []string{"xs", "sm", "md", "lg", "xl"}
	for _, key := range expectedKeys {
		if _, ok := scale[key]; !ok {
			t.Errorf("StandardScale() missing key %q", key)
		}
	}

	// Check md is 1.0 (no scaling)
	if md, ok := scale["md"].(float64); !ok || md != 1.0 {
		t.Errorf("StandardScale()[\"md\"] = %v, want 1.0", scale["md"])
	}

	// Check ordering: xs < sm < md < lg < xl
	xs := scale["xs"].(float64)
	sm := scale["sm"].(float64)
	md := scale["md"].(float64)
	lg := scale["lg"].(float64)
	xl := scale["xl"].(float64)

	if !(xs < sm && sm < md && md < lg && lg < xl) {
		t.Error("StandardScale() values should be in ascending order")
	}
}

func TestTypographyScale(t *testing.T) {
	scale := TypographyScale()

	expectedKeys := []string{"xs", "sm", "md", "lg", "xl", "2xl", "3xl"}
	for _, key := range expectedKeys {
		if _, ok := scale[key]; !ok {
			t.Errorf("TypographyScale() missing key %q", key)
		}
	}

	// Check md is 1.0 (base)
	if md, ok := scale["md"].(float64); !ok || md != 1.0 {
		t.Errorf("TypographyScale()[\"md\"] = %v, want 1.0", scale["md"])
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		want   float64
		wantOk bool
	}{
		{"float64", float64(1.5), 1.5, true},
		{"float32", float32(1.5), 1.5, true},
		{"int", int(42), 42.0, true},
		{"int64", int64(42), 42.0, true},
		{"int32", int32(42), 42.0, true},
		{"string", "1.5", 0, false},
		{"nil", nil, 0, false},
		{"bool", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toFloat64(tt.input)
			if ok != tt.wantOk {
				t.Errorf("toFloat64(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if ok && got != tt.want {
				t.Errorf("toFloat64(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// Helper function to check if a token exists at a path
func tokenExists(root map[string]interface{}, path string) bool {
	parts := strings.Split(path, ".")
	current := root

	for i, part := range parts {
		val, ok := current[part]
		if !ok {
			return false
		}

		if i == len(parts)-1 {
			// Last part - check if it's a token
			if m, ok := val.(map[string]interface{}); ok {
				_, hasValue := m["$value"]
				return hasValue
			}
			return false
		}

		// Navigate deeper
		if m, ok := val.(map[string]interface{}); ok {
			current = m
		} else {
			return false
		}
	}

	return false
}

// Helper function to get a token's value
func getTokenValue(root map[string]interface{}, path string) string {
	parts := strings.Split(path, ".")
	current := root

	for i, part := range parts {
		val, ok := current[part]
		if !ok {
			return ""
		}

		if i == len(parts)-1 {
			if m, ok := val.(map[string]interface{}); ok {
				if v, ok := m["$value"].(string); ok {
					return v
				}
			}
			return ""
		}

		if m, ok := val.(map[string]interface{}); ok {
			current = m
		} else {
			return ""
		}
	}

	return ""
}
