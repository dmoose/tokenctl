// tokenctl/pkg/tokens/expressions_test.go

package tokens

import (
	"strings"
	"testing"
)

func TestIsExpression(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"calc({size.field} * 0.6)", true},
		{"calc(10px + 5px)", true},
		{"contrast({color.primary})", true},
		{"darken({color.primary}, 10%)", true},
		{"lighten({color.primary}, 20%)", true},
		{"scale({size.base}, 1.5)", true},
		{"shade({color.base}, 1)", true},
		{"shade({color.base}, 2)", true},
		{"{color.primary}", false},
		{"#3b82f6", false},
		{"10px", false},
		{"", false},
		{"calculus", false},
		{"contrasty", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsExpression(tt.input)
			if got != tt.want {
				t.Errorf("IsExpression(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// Helper to create a resolver with test tokens
func createTestResolver(tokens map[string]any) *Resolver {
	dict := NewDictionary()

	// Build nested structure from flat tokens
	for path, value := range tokens {
		parts := strings.Split(path, ".")
		current := dict.Root

		for i, part := range parts {
			if i == len(parts)-1 {
				// Last part - create token
				current[part] = map[string]any{
					"$value": value,
				}
			} else {
				// Intermediate part - create/get group
				if _, ok := current[part]; !ok {
					current[part] = make(map[string]any)
				}
				current = current[part].(map[string]any)
			}
		}
	}

	resolver, _ := NewResolver(dict)
	return resolver
}

func TestExpressionEvaluator_Calc(t *testing.T) {
	tests := []struct {
		name    string
		tokens  map[string]any
		expr    string
		want    string
		wantErr bool
	}{
		{
			name:   "multiply dimension by number",
			tokens: map[string]any{"size.base": "2.5rem"},
			expr:   "calc({size.base} * 0.6)",
			want:   "1.5rem",
		},
		{
			name:   "multiply by 1.0 (no change)",
			tokens: map[string]any{"size.base": "2.5rem"},
			expr:   "calc({size.base} * 1)",
			want:   "2.5rem",
		},
		{
			name:   "multiply pixel value",
			tokens: map[string]any{"size.field": "40px"},
			expr:   "calc({size.field} * 0.8)",
			want:   "32px",
		},
		{
			name:   "divide dimension",
			tokens: map[string]any{"size.large": "24px"},
			expr:   "calc({size.large} / 2)",
			want:   "12px",
		},
		{
			name:   "add dimensions",
			tokens: map[string]any{"spacing.sm": "0.5rem", "spacing.md": "1rem"},
			expr:   "calc({spacing.sm} + {spacing.md})",
			want:   "1.5rem",
		},
		{
			name:    "missing token reference",
			tokens:  map[string]any{},
			expr:    "calc({missing.token} * 2)",
			wantErr: true,
		},
		{
			name:    "incompatible units for add",
			tokens:  map[string]any{"a": "10px", "b": "1rem"},
			expr:    "calc({a} + {b})",
			wantErr: true,
		},
		{
			name:    "divide by zero",
			tokens:  map[string]any{"size": "10px"},
			expr:    "calc({size} / 0)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := createTestResolver(tt.tokens)
			eval := NewExpressionEvaluator(resolver)

			result, err := eval.Evaluate(tt.expr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Evaluate(%q) expected error, got nil", tt.expr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Evaluate(%q) unexpected error: %v", tt.expr, err)
			}

			if result != tt.want {
				t.Errorf("Evaluate(%q) = %q, want %q", tt.expr, result, tt.want)
			}
		})
	}
}

func TestExpressionEvaluator_Contrast(t *testing.T) {
	tests := []struct {
		name    string
		tokens  map[string]any
		expr    string
		wantErr bool
	}{
		{
			name:   "contrast for dark color returns light",
			tokens: map[string]any{"color.primary": "#1e3a5f"},
			expr:   "contrast({color.primary})",
		},
		{
			name:   "contrast for light color returns dark",
			tokens: map[string]any{"color.background": "#ffffff"},
			expr:   "contrast({color.background})",
		},
		{
			name:   "contrast for oklch color",
			tokens: map[string]any{"color.brand": "oklch(49.12% 0.309 275.75)"},
			expr:   "contrast({color.brand})",
		},
		{
			name:    "contrast with missing token",
			tokens:  map[string]any{},
			expr:    "contrast({color.missing})",
			wantErr: true,
		},
		{
			name:    "contrast with invalid color",
			tokens:  map[string]any{"color.bad": "not-a-color"},
			expr:    "contrast({color.bad})",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := createTestResolver(tt.tokens)
			eval := NewExpressionEvaluator(resolver)

			result, err := eval.Evaluate(tt.expr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Evaluate(%q) expected error, got nil", tt.expr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Evaluate(%q) unexpected error: %v", tt.expr, err)
			}

			// Result should be a valid color string
			resultStr, ok := result.(string)
			if !ok {
				t.Errorf("Evaluate(%q) result is not a string: %T", tt.expr, result)
				return
			}

			// Should start with # or oklch
			if !strings.HasPrefix(resultStr, "#") && !strings.HasPrefix(resultStr, "oklch") {
				t.Errorf("Evaluate(%q) = %q, expected color format", tt.expr, resultStr)
			}
		})
	}
}

func TestExpressionEvaluator_Darken(t *testing.T) {
	tests := []struct {
		name    string
		tokens  map[string]any
		expr    string
		wantErr bool
	}{
		{
			name:   "darken hex color",
			tokens: map[string]any{"color.primary": "#3b82f6"},
			expr:   "darken({color.primary}, 20%)",
		},
		{
			name:   "darken oklch color",
			tokens: map[string]any{"color.primary": "oklch(60% 0.2 250)"},
			expr:   "darken({color.primary}, 30%)",
		},
		{
			name:    "darken missing token",
			tokens:  map[string]any{},
			expr:    "darken({color.missing}, 10%)",
			wantErr: true,
		},
		{
			name:    "darken invalid color",
			tokens:  map[string]any{"color.bad": "invalid"},
			expr:    "darken({color.bad}, 10%)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := createTestResolver(tt.tokens)
			eval := NewExpressionEvaluator(resolver)

			result, err := eval.Evaluate(tt.expr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Evaluate(%q) expected error, got nil", tt.expr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Evaluate(%q) unexpected error: %v", tt.expr, err)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Errorf("Evaluate(%q) result is not a string", tt.expr)
				return
			}

			if !strings.HasPrefix(resultStr, "#") && !strings.HasPrefix(resultStr, "oklch") {
				t.Errorf("Evaluate(%q) = %q, expected color format", tt.expr, resultStr)
			}
		})
	}
}

func TestExpressionEvaluator_Lighten(t *testing.T) {
	tests := []struct {
		name    string
		tokens  map[string]any
		expr    string
		wantErr bool
	}{
		{
			name:   "lighten hex color",
			tokens: map[string]any{"color.primary": "#1e3a5f"},
			expr:   "lighten({color.primary}, 20%)",
		},
		{
			name:   "lighten oklch color",
			tokens: map[string]any{"color.primary": "oklch(40% 0.2 250)"},
			expr:   "lighten({color.primary}, 30%)",
		},
		{
			name:    "lighten missing token",
			tokens:  map[string]any{},
			expr:    "lighten({color.missing}, 10%)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := createTestResolver(tt.tokens)
			eval := NewExpressionEvaluator(resolver)

			result, err := eval.Evaluate(tt.expr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Evaluate(%q) expected error, got nil", tt.expr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Evaluate(%q) unexpected error: %v", tt.expr, err)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Errorf("Evaluate(%q) result is not a string", tt.expr)
				return
			}

			if !strings.HasPrefix(resultStr, "#") && !strings.HasPrefix(resultStr, "oklch") {
				t.Errorf("Evaluate(%q) = %q, expected color format", tt.expr, resultStr)
			}
		})
	}
}

func TestExpressionEvaluator_Scale(t *testing.T) {
	tests := []struct {
		name    string
		tokens  map[string]any
		expr    string
		want    string
		wantErr bool
	}{
		{
			name:   "scale up",
			tokens: map[string]any{"size.base": "2.5rem"},
			expr:   "scale({size.base}, 1.2)",
			want:   "3rem",
		},
		{
			name:   "scale down",
			tokens: map[string]any{"size.base": "2.5rem"},
			expr:   "scale({size.base}, 0.6)",
			want:   "1.5rem",
		},
		{
			name:   "scale pixels",
			tokens: map[string]any{"size.field": "40px"},
			expr:   "scale({size.field}, 0.5)",
			want:   "20px",
		},
		{
			name:    "scale missing token",
			tokens:  map[string]any{},
			expr:    "scale({size.missing}, 1.5)",
			wantErr: true,
		},
		{
			name:    "scale invalid dimension",
			tokens:  map[string]any{"size.bad": "not-a-dimension"},
			expr:    "scale({size.bad}, 1.5)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := createTestResolver(tt.tokens)
			eval := NewExpressionEvaluator(resolver)

			result, err := eval.Evaluate(tt.expr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Evaluate(%q) expected error, got nil", tt.expr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Evaluate(%q) unexpected error: %v", tt.expr, err)
			}

			if result != tt.want {
				t.Errorf("Evaluate(%q) = %q, want %q", tt.expr, result, tt.want)
			}
		})
	}
}

func TestExpressionEvaluator_UnrecognizedExpression(t *testing.T) {
	resolver := createTestResolver(map[string]any{})
	eval := NewExpressionEvaluator(resolver)

	_, err := eval.Evaluate("unknown({something})")
	if err == nil {
		t.Error("Expected error for unrecognized expression")
	}

	if !strings.Contains(err.Error(), "unrecognized") {
		t.Errorf("Error should mention 'unrecognized', got: %v", err)
	}
}

func TestResolverWithExpressions(t *testing.T) {
	// Test that expressions are evaluated during resolution
	dict := NewDictionary()
	dict.Root = map[string]any{
		"size": map[string]any{
			"base": map[string]any{
				"$value": "2.5rem",
			},
			"small": map[string]any{
				"$value": "calc({size.base} * 0.8)",
			},
		},
		"color": map[string]any{
			"primary": map[string]any{
				"$value": "#3b82f6",
			},
			"primary-content": map[string]any{
				"$value": "contrast({color.primary})",
			},
		},
	}

	resolver, err := NewResolver(dict)
	if err != nil {
		t.Fatalf("NewResolver failed: %v", err)
	}

	resolved, err := resolver.ResolveAll()
	if err != nil {
		t.Fatalf("ResolveAll failed: %v", err)
	}

	// Check size.small was calculated
	if small, ok := resolved["size.small"]; ok {
		if small != "2rem" {
			t.Errorf("size.small = %q, want \"2rem\"", small)
		}
	} else {
		t.Error("size.small not found in resolved tokens")
	}

	// Check color.primary-content was generated
	if content, ok := resolved["color.primary-content"]; ok {
		contentStr, _ := content.(string)
		if !strings.HasPrefix(contentStr, "#") && !strings.HasPrefix(contentStr, "oklch") {
			t.Errorf("color.primary-content = %q, expected color format", contentStr)
		}
	} else {
		t.Error("color.primary-content not found in resolved tokens")
	}
}

func TestExpressionEvaluator_Shade(t *testing.T) {
	tests := []struct {
		name    string
		tokens  map[string]any
		expr    string
		wantErr bool
	}{
		{
			name:   "shade level 1 from white",
			tokens: map[string]any{"color.base": "oklch(100% 0 0)"},
			expr:   "shade({color.base}, 1)",
		},
		{
			name:   "shade level 2 from white",
			tokens: map[string]any{"color.base": "oklch(100% 0 0)"},
			expr:   "shade({color.base}, 2)",
		},
		{
			name:   "shade hex color",
			tokens: map[string]any{"color.base": "#ffffff"},
			expr:   "shade({color.base}, 1)",
		},
		{
			name:    "shade missing token",
			tokens:  map[string]any{},
			expr:    "shade({color.missing}, 1)",
			wantErr: true,
		},
		{
			name:    "shade invalid color",
			tokens:  map[string]any{"color.bad": "not-a-color"},
			expr:    "shade({color.bad}, 1)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := createTestResolver(tt.tokens)
			eval := NewExpressionEvaluator(resolver)

			result, err := eval.Evaluate(tt.expr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Evaluate(%q) expected error, got nil", tt.expr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Evaluate(%q) unexpected error: %v", tt.expr, err)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Errorf("Evaluate(%q) result is not a string", tt.expr)
				return
			}

			if !strings.HasPrefix(resultStr, "#") && !strings.HasPrefix(resultStr, "oklch") {
				t.Errorf("Evaluate(%q) = %q, expected color format", tt.expr, resultStr)
			}
		})
	}
}
