package generators

import (
	"strings"
	"testing"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

func TestTailwindGenerator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tokens      map[string]any
		components  map[string]tokens.ComponentDefinition
		expected    []string
		notExpected []string
	}{
		{
			name: "Metadata Filtering",
			components: map[string]tokens.ComponentDefinition{
				"button": {
					Class: "btn",
					Base: map[string]any{
						"display": "block",
						"$desc":   "should be skipped",
					},
				},
			},
			expected: []string{
				".btn {",
				"display: block;",
			},
			notExpected: []string{
				"$desc",
				"should be skipped",
			},
		},
		{
			name: "Complex Value Serialization",
			components: map[string]tokens.ComponentDefinition{
				"card": {
					Class: "card",
					Base: map[string]any{
						"box-shadow": []any{
							"0 1px 2px rgba(0,0,0,0.1)",
							"0 2px 4px rgba(0,0,0,0.1)",
						},
					},
				},
			},
			expected: []string{
				"box-shadow: 0 1px 2px rgba(0,0,0,0.1), 0 2px 4px rgba(0,0,0,0.1);",
			},
		},
		{
			name: "Reference Resolution in Component",
			components: map[string]tokens.ComponentDefinition{
				"alert": {
					Class: "alert",
					Base: map[string]any{
						"color": "{color.primary}",
					},
				},
			},
			expected: []string{
				"color: var(--color-primary);",
			},
		},
		{
			name: "Multiple Token References in Single Property",
			components: map[string]tokens.ComponentDefinition{
				"button": {
					Class: "btn",
					Base: map[string]any{
						"padding": "{spacing.sm} {spacing.md}",
						"margin":  "{spacing.xs} {spacing.sm} {spacing.md}",
					},
				},
			},
			expected: []string{
				"padding: var(--spacing-sm) var(--spacing-md);",
				"margin: var(--spacing-xs) var(--spacing-sm) var(--spacing-md);",
			},
			notExpected: []string{
				"{spacing",
				"} {spacing",
			},
		},
		{
			name: "Token References with Strings",
			components: map[string]tokens.ComponentDefinition{
				"button": {
					Class: "btn",
					Base: map[string]any{
						"border": "1px solid {color.border}",
					},
				},
			},
			expected: []string{
				"border: 1px solid var(--color-border);",
			},
			notExpected: []string{
				"{color",
			},
		},
		{
			name: "No Token References",
			components: map[string]tokens.ComponentDefinition{
				"button": {
					Class: "btn",
					Base: map[string]any{
						"display": "flex",
						"padding": "0.5rem 1rem",
					},
				},
			},
			expected: []string{
				"display: flex;",
				"padding: 0.5rem 1rem;",
			},
		},
	}

	gen := NewTailwindGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := gen.GenerateComponents(tt.components)
			if err != nil {
				t.Fatalf("GenerateComponents failed: %v", err)
			}

			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", exp, output)
				}
			}

			for _, notExp := range tt.notExpected {
				if strings.Contains(output, notExp) {
					t.Errorf("Expected output NOT to contain %q, but it did.\nOutput:\n%s", notExp, output)
				}
			}
		})
	}
}

func TestTailwindGenerator_ArraySerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		components map[string]tokens.ComponentDefinition
		expected   []string
	}{
		{
			name: "Margin with space-separated array",
			components: map[string]tokens.ComponentDefinition{
				"card": {
					Class: "card",
					Base: map[string]any{
						"margin": []any{"10px", "20px", "10px", "20px"},
					},
				},
			},
			expected: []string{
				".card {",
				"margin: 10px 20px 10px 20px;",
			},
		},
		{
			name: "Padding with space-separated array",
			components: map[string]tokens.ComponentDefinition{
				"button": {
					Class: "btn",
					Base: map[string]any{
						"padding": []any{"0.5rem", "1rem"},
					},
				},
			},
			expected: []string{
				".btn {",
				"padding: 0.5rem 1rem;",
			},
		},
		{
			name: "Box-shadow with comma-separated array",
			components: map[string]tokens.ComponentDefinition{
				"card": {
					Class: "card",
					Base: map[string]any{
						"box-shadow": []any{
							"0 1px 2px rgba(0,0,0,0.1)",
							"0 2px 4px rgba(0,0,0,0.2)",
						},
					},
				},
			},
			expected: []string{
				".card {",
				"box-shadow: 0 1px 2px rgba(0,0,0,0.1), 0 2px 4px rgba(0,0,0,0.2);",
			},
		},
		{
			name: "Font-family with comma-separated array",
			components: map[string]tokens.ComponentDefinition{
				"text": {
					Class: "text",
					Base: map[string]any{
						"font-family": []any{"Inter", "Arial", "sans-serif"},
					},
				},
			},
			expected: []string{
				".text {",
				"font-family: Inter, Arial, sans-serif;",
			},
		},
		{
			name: "Border-radius with space-separated array",
			components: map[string]tokens.ComponentDefinition{
				"box": {
					Class: "box",
					Base: map[string]any{
						"border-radius": []any{"4px", "4px", "0", "0"},
					},
				},
			},
			expected: []string{
				".box {",
				"border-radius: 4px 4px 0 0;",
			},
		},
		{
			name: "Transform with comma-separated array",
			components: map[string]tokens.ComponentDefinition{
				"animated": {
					Class: "animated",
					Base: map[string]any{
						"transform": []any{"rotate(45deg)", "scale(1.5)"},
					},
				},
			},
			expected: []string{
				".animated {",
				"transform: rotate(45deg), scale(1.5);",
			},
		},
		{
			name: "Mixed space and comma separated properties",
			components: map[string]tokens.ComponentDefinition{
				"complex": {
					Class: "complex",
					Base: map[string]any{
						"margin":      []any{"1rem", "2rem"},
						"padding":     []any{"0.5rem", "1rem"},
						"box-shadow":  []any{"0 1px 2px black", "0 2px 4px red"},
						"font-family": []any{"Arial", "sans-serif"},
					},
				},
			},
			expected: []string{
				".complex {",
				"margin: 1rem 2rem;",
				"padding: 0.5rem 1rem;",
				"box-shadow: 0 1px 2px black, 0 2px 4px red;",
				"font-family: Arial, sans-serif;",
			},
		},
	}

	gen := NewTailwindGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := gen.GenerateComponents(tt.components)
			if err != nil {
				t.Fatalf("GenerateComponents failed: %v", err)
			}

			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", exp, output)
				}
			}
		})
	}
}

func TestResolveTokenReferences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single token reference",
			input:    "{color.primary}",
			expected: "var(--color-primary)",
		},
		{
			name:     "Multiple token references",
			input:    "{spacing.sm} {spacing.md}",
			expected: "var(--spacing-sm) var(--spacing-md)",
		},
		{
			name:     "Token reference with surrounding text",
			input:    "1px solid {color.border}",
			expected: "1px solid var(--color-border)",
		},
		{
			name:     "Multiple token references with text",
			input:    "{spacing.xs} {spacing.sm} {spacing.md} {spacing.lg}",
			expected: "var(--spacing-xs) var(--spacing-sm) var(--spacing-md) var(--spacing-lg)",
		},
		{
			name:     "No token references",
			input:    "0.5rem 1rem",
			expected: "0.5rem 1rem",
		},
		{
			name:     "Nested path reference",
			input:    "{color.semantic.success.background}",
			expected: "var(--color-semantic-success-background)",
		},
		{
			name:     "Complex CSS value",
			input:    "0 1px 2px {color.shadow}",
			expected: "0 1px 2px var(--color-shadow)",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Mixed tokens and values",
			input:    "{spacing.md} auto {spacing.lg}",
			expected: "var(--spacing-md) auto var(--spacing-lg)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := resolveTokenReferences(tt.input)
			if result != tt.expected {
				t.Errorf("resolveTokenReferences(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTailwindGenerator_PropertyDeclarations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		properties []tokens.PropertyToken
		expected   []string
	}{
		{
			name: "Color property",
			properties: []tokens.PropertyToken{
				{
					Path:         "color.primary",
					CSSName:      "--color-primary",
					CSSSyntax:    "<color>",
					Inherits:     true,
					InitialValue: "oklch(50% 0.2 250)",
				},
			},
			expected: []string{
				"@property --color-primary {",
				"syntax: '<color>';",
				"inherits: true;",
				"initial-value: oklch(50% 0.2 250);",
			},
		},
		{
			name: "Dimension with inherits false",
			properties: []tokens.PropertyToken{
				{
					Path:         "timing.fast",
					CSSName:      "--timing-fast",
					CSSSyntax:    "<time>",
					Inherits:     false,
					InitialValue: "150ms",
				},
			},
			expected: []string{
				"@property --timing-fast {",
				"syntax: '<time>';",
				"inherits: false;",
				"initial-value: 150ms;",
			},
		},
		{
			name: "Multiple properties sorted by path",
			properties: []tokens.PropertyToken{
				{
					Path:         "spacing.md",
					CSSName:      "--spacing-md",
					CSSSyntax:    "<length>",
					Inherits:     true,
					InitialValue: "1rem",
				},
				{
					Path:         "color.primary",
					CSSName:      "--color-primary",
					CSSSyntax:    "<color>",
					Inherits:     true,
					InitialValue: "#3b82f6",
				},
			},
			expected: []string{
				"@property --color-primary {",
				"@property --spacing-md {",
			},
		},
	}

	gen := NewTailwindGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := &GenerationContext{
				ResolvedTokens: map[string]any{},
				PropertyTokens: tt.properties,
			}

			output, err := gen.Generate(ctx)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", exp, output)
				}
			}
		})
	}
}

func TestTailwindGenerator_EffectTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tokens   map[string]any
		expected []string
	}{
		{
			name: "Effect tokens with value 0 and 1",
			tokens: map[string]any{
				"effect.depth": 1,
				"effect.noise": 0,
			},
			expected: []string{
				"--effect-depth: 1;",
				"--effect-noise: 0;",
			},
		},
		{
			name: "Effect tokens mixed with other tokens",
			tokens: map[string]any{
				"color.primary": "#3b82f6",
				"effect.depth":  1,
				"size.field":    "2.5rem",
				"effect.noise":  0,
			},
			expected: []string{
				"--color-primary: #3b82f6;",
				"--effect-depth: 1;",
				"--effect-noise: 0;",
				"--size-field: 2.5rem;",
			},
		},
	}

	gen := NewTailwindGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := gen.GenerateFromResolved(tt.tokens)
			if err != nil {
				t.Fatalf("GenerateFromResolved failed: %v", err)
			}

			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", exp, output)
				}
			}
		})
	}
}

