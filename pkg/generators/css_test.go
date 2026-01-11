package generators

import (
	"strings"
	"testing"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

func TestNewCSSGenerator(t *testing.T) {
	g := NewCSSGenerator()
	if g == nil {
		t.Error("NewCSSGenerator returned nil")
	}
}

func TestCSSGenerator_Generate_Basic(t *testing.T) {
	g := NewCSSGenerator()
	ctx := &GenerationContext{
		ResolvedTokens: map[string]interface{}{
			"color.primary": "#3b82f6",
			"spacing.md":    "1rem",
		},
	}

	output, err := g.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check layer declaration
	if !strings.Contains(output, "@layer reset, tokens, themes, components;") {
		t.Error("Missing layer declaration")
	}

	// Check reset layer
	if !strings.Contains(output, "@layer reset {") {
		t.Error("Missing reset layer")
	}

	// Check tokens layer
	if !strings.Contains(output, "@layer tokens {") {
		t.Error("Missing tokens layer")
	}

	// Check CSS variables
	if !strings.Contains(output, "--color-primary: #3b82f6;") {
		t.Error("Missing color-primary variable")
	}
	if !strings.Contains(output, "--spacing-md: 1rem;") {
		t.Error("Missing spacing-md variable")
	}
}

func TestCSSGenerator_Generate_WithThemes(t *testing.T) {
	g := NewCSSGenerator()
	ctx := &GenerationContext{
		ResolvedTokens: map[string]interface{}{
			"color.primary": "#3b82f6",
		},
		Themes: map[string]ThemeContext{
			"dark": {
				DiffTokens: map[string]interface{}{
					"color.primary": "#60a5fa",
				},
			},
		},
	}

	output, err := g.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check themes layer
	if !strings.Contains(output, "@layer themes {") {
		t.Error("Missing themes layer")
	}

	// Check dark theme selector
	if !strings.Contains(output, `[data-theme="dark"]`) {
		t.Error("Missing dark theme selector")
	}
}

func TestCSSGenerator_Generate_WithLightTheme(t *testing.T) {
	g := NewCSSGenerator()
	ctx := &GenerationContext{
		ResolvedTokens: map[string]interface{}{},
		Themes: map[string]ThemeContext{
			"light": {
				DiffTokens: map[string]interface{}{
					"color.surface": "#ffffff",
				},
			},
		},
	}

	output, err := g.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Light theme should have :root selector
	if !strings.Contains(output, `:root, [data-theme="light"]`) {
		t.Error("Light theme should have :root selector")
	}
}

func TestCSSGenerator_Generate_WithComponents(t *testing.T) {
	g := NewCSSGenerator()
	ctx := &GenerationContext{
		ResolvedTokens: map[string]interface{}{},
		Components: map[string]tokens.ComponentDefinition{
			"btn": {
				Class: "btn",
				Base: map[string]interface{}{
					"padding":    "0.5rem 1rem",
					"background": "var(--color-primary)",
				},
				Variants: map[string]tokens.VariantDef{
					"primary": {
						Class: "btn-primary",
						Properties: map[string]interface{}{
							"background-color": "var(--color-primary)",
						},
					},
				},
				Sizes: map[string]tokens.VariantDef{
					"sm": {
						Class: "btn-sm",
						Properties: map[string]interface{}{
							"padding": "0.25rem 0.5rem",
						},
					},
				},
			},
		},
	}

	output, err := g.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check components layer
	if !strings.Contains(output, "@layer components {") {
		t.Error("Missing components layer")
	}

	// Check base class
	if !strings.Contains(output, ".btn {") {
		t.Error("Missing .btn class")
	}

	// Check variant
	if !strings.Contains(output, ".btn-primary {") {
		t.Error("Missing .btn-primary class")
	}

	// Check size
	if !strings.Contains(output, ".btn-sm {") {
		t.Error("Missing .btn-sm class")
	}
}

func TestCSSGenerator_Generate_WithNestedPseudoSelectors(t *testing.T) {
	g := NewCSSGenerator()
	ctx := &GenerationContext{
		ResolvedTokens: map[string]interface{}{},
		Components: map[string]tokens.ComponentDefinition{
			"link": {
				Class: "link",
				Base: map[string]interface{}{
					"color": "var(--color-link)",
					"&:hover": map[string]interface{}{
						"color": "var(--color-link-hover)",
					},
					":visited": map[string]interface{}{
						"color": "var(--color-link-visited)",
					},
				},
			},
		},
	}

	output, err := g.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check base class
	if !strings.Contains(output, ".link {") {
		t.Error("Missing .link class")
	}

	// Check hover pseudo-selector
	if !strings.Contains(output, ".link:hover {") {
		t.Error("Missing .link:hover selector")
	}

	// Check visited pseudo-selector
	if !strings.Contains(output, ".link:visited {") {
		t.Error("Missing .link:visited selector")
	}
}

func TestCSSGenerator_Generate_WithPropertyDeclarations(t *testing.T) {
	g := NewCSSGenerator()
	ctx := &GenerationContext{
		ResolvedTokens: map[string]interface{}{
			"color.primary": "#3b82f6",
		},
		PropertyTokens: []tokens.PropertyToken{
			{
				Path:         "color.primary",
				CSSName:      "--color-primary",
				CSSSyntax:    "<color>",
				InitialValue: "#3b82f6",
				Inherits:     true,
			},
		},
	}

	output, err := g.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check @property declaration
	if !strings.Contains(output, "@property --color-primary {") {
		t.Error("Missing @property declaration")
	}
	if !strings.Contains(output, "syntax: '<color>';") {
		t.Error("Missing syntax in @property")
	}
	if !strings.Contains(output, "inherits: true;") {
		t.Error("Missing inherits in @property")
	}
}

func TestGenerateReset(t *testing.T) {
	reset := generateReset()

	if !strings.Contains(reset, "@layer reset {") {
		t.Error("Reset should be in @layer reset")
	}
	if !strings.Contains(reset, "box-sizing: border-box") {
		t.Error("Reset should include box-sizing")
	}
	if !strings.Contains(reset, "margin: 0") {
		t.Error("Reset should include margin reset")
	}
}

func TestBuildStateSelector(t *testing.T) {
	g := NewCSSGenerator()

	tests := []struct {
		className string
		stateKey  string
		expected  string
	}{
		{"btn", "&:hover", ".btn:hover"},
		{"btn", ":hover", ".btn:hover"},
		{"btn", "&:active", ".btn:active"},
		{"card", ".card-body", ".card .card-body"},
	}

	for _, tt := range tests {
		result := g.buildStateSelector(tt.className, tt.stateKey)
		if result != tt.expected {
			t.Errorf("buildStateSelector(%q, %q) = %q, want %q",
				tt.className, tt.stateKey, result, tt.expected)
		}
	}
}

func TestCSSGenerator_SkipsMapValues(t *testing.T) {
	g := NewCSSGenerator()
	ctx := &GenerationContext{
		ResolvedTokens: map[string]interface{}{
			"color.primary": "#3b82f6",
			"color.nested":  map[string]interface{}{"should": "skip"},
		},
	}

	output, err := g.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(output, "--color-primary:") {
		t.Error("Should include primitive value")
	}
	if strings.Contains(output, "--color-nested:") {
		t.Error("Should skip map values")
	}
}

func TestCSSGenerator_DeterministicOutput(t *testing.T) {
	g := NewCSSGenerator()
	ctx := &GenerationContext{
		ResolvedTokens: map[string]interface{}{
			"z.last":   "3",
			"a.first":  "1",
			"m.middle": "2",
		},
	}

	// Generate multiple times and ensure consistent order
	output1, _ := g.Generate(ctx)
	output2, _ := g.Generate(ctx)

	if output1 != output2 {
		t.Error("Output should be deterministic")
	}

	// Check order (a before m before z)
	aIdx := strings.Index(output1, "--a-first")
	mIdx := strings.Index(output1, "--m-middle")
	zIdx := strings.Index(output1, "--z-last")

	if aIdx > mIdx || mIdx > zIdx {
		t.Error("Variables should be sorted alphabetically")
	}
}
