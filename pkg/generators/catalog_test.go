// tokenctl/pkg/generators/catalog_test.go
package generators

import (
	"encoding/json"
	"testing"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

func TestCatalogGenerator_Generate_BasicTokens(t *testing.T) {
	gen := NewCatalogGenerator()

	resolvedTokens := map[string]any{
		"color.primary":   "#3b82f6",
		"color.secondary": "#8b5cf6",
		"spacing.sm":      "0.5rem",
	}

	components := map[string]tokens.ComponentDefinition{}

	result, err := gen.Generate(resolvedTokens, components, nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var catalog CatalogSchema
	if err := json.Unmarshal([]byte(result), &catalog); err != nil {
		t.Fatalf("Failed to parse catalog JSON: %v", err)
	}

	// Verify meta
	if catalog.Meta.Version != CatalogSchemaVersion {
		t.Errorf("Expected version %s, got %s", CatalogSchemaVersion, catalog.Meta.Version)
	}
	if catalog.Meta.TokenctlVersion != TokenctlVersion {
		t.Errorf("Expected tokenctl_version %s, got %s", TokenctlVersion, catalog.Meta.TokenctlVersion)
	}
	if catalog.Meta.GeneratedAt == "" {
		t.Error("Expected generated_at to be set")
	}

	// Verify tokens
	if len(catalog.Tokens) != 3 {
		t.Errorf("Expected 3 tokens, got %d", len(catalog.Tokens))
	}
	if catalog.Tokens["color.primary"] != "#3b82f6" {
		t.Errorf("Expected color.primary to be #3b82f6, got %v", catalog.Tokens["color.primary"])
	}

	// Verify no themes section when none provided
	if catalog.Themes != nil {
		t.Error("Expected themes to be nil when no themes provided")
	}
}

func TestCatalogGenerator_Generate_WithComponents(t *testing.T) {
	gen := NewCatalogGenerator()

	resolvedTokens := map[string]any{
		"color.primary": "#3b82f6",
	}

	components := map[string]tokens.ComponentDefinition{
		"button": {
			Class: "btn",
			Base: map[string]any{
				"display": "inline-flex",
			},
			Variants: map[string]tokens.VariantDef{
				"primary":   {Class: "btn-primary", Properties: map[string]any{"background": "var(--color-primary)"}},
				"secondary": {Class: "btn-secondary", Properties: map[string]any{"background": "var(--color-secondary)"}},
			},
			Sizes: map[string]tokens.VariantDef{
				"sm": {Class: "btn-sm", Properties: map[string]any{"padding": "0.25rem 0.5rem"}},
				"lg": {Class: "btn-lg", Properties: map[string]any{"padding": "0.75rem 1.5rem"}},
			},
		},
	}

	result, err := gen.Generate(resolvedTokens, components, nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var catalog CatalogSchema
	if err := json.Unmarshal([]byte(result), &catalog); err != nil {
		t.Fatalf("Failed to parse catalog JSON: %v", err)
	}

	// Verify components
	buttonComp, ok := catalog.Components["button"]
	if !ok {
		t.Fatal("Expected button component in catalog")
	}

	// Should have 5 classes: btn, btn-primary, btn-secondary, btn-sm, btn-lg
	if len(buttonComp.Classes) != 5 {
		t.Errorf("Expected 5 classes, got %d: %v", len(buttonComp.Classes), buttonComp.Classes)
	}

	// Check definitions exist
	if _, ok := buttonComp.Definitions["btn"]; !ok {
		t.Error("Expected btn definition")
	}
	if _, ok := buttonComp.Definitions["btn-primary"]; !ok {
		t.Error("Expected btn-primary definition")
	}
}

func TestCatalogGenerator_Generate_WithThemes(t *testing.T) {
	gen := NewCatalogGenerator()

	resolvedTokens := map[string]any{
		"color.primary": "#3b82f6",
	}

	components := map[string]tokens.ComponentDefinition{}

	lightExtends := "" // empty string means no parent, will be converted to nil pointer
	darkExtends := "light"

	themes := map[string]CatalogThemeInput{
		"light": {
			Extends:     nil, // extends base
			Description: "Default light theme",
			ResolvedTokens: map[string]any{
				"color.primary": "#60a5fa",
			},
			DiffTokens: map[string]any{
				"color.primary": "#60a5fa",
			},
		},
		"dark": {
			Extends:     &darkExtends,
			Description: "Dark theme extends light theme",
			ResolvedTokens: map[string]any{
				"color.primary": "#1e40af",
			},
			DiffTokens: map[string]any{
				"color.primary": "#1e40af",
			},
		},
	}

	// Avoid unused variable warning
	_ = lightExtends

	result, err := gen.Generate(resolvedTokens, components, themes)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var catalog CatalogSchema
	if err := json.Unmarshal([]byte(result), &catalog); err != nil {
		t.Fatalf("Failed to parse catalog JSON: %v", err)
	}

	// Verify themes section exists
	if catalog.Themes == nil {
		t.Fatal("Expected themes section in catalog")
	}

	if len(catalog.Themes) != 2 {
		t.Errorf("Expected 2 themes, got %d", len(catalog.Themes))
	}

	// Verify light theme
	lightTheme, ok := catalog.Themes["light"]
	if !ok {
		t.Fatal("Expected light theme in catalog")
	}
	if lightTheme.Extends != nil {
		t.Errorf("Expected light theme extends to be nil, got %v", *lightTheme.Extends)
	}
	if lightTheme.Description != "Default light theme" {
		t.Errorf("Expected light theme description, got %s", lightTheme.Description)
	}
	if lightTheme.Tokens["color.primary"] != "#60a5fa" {
		t.Errorf("Expected light theme color.primary to be #60a5fa, got %v", lightTheme.Tokens["color.primary"])
	}

	// Verify dark theme
	darkTheme, ok := catalog.Themes["dark"]
	if !ok {
		t.Fatal("Expected dark theme in catalog")
	}
	if darkTheme.Extends == nil {
		t.Fatal("Expected dark theme extends to be set")
	}
	if *darkTheme.Extends != "light" {
		t.Errorf("Expected dark theme to extend light, got %s", *darkTheme.Extends)
	}
	if darkTheme.Description != "Dark theme extends light theme" {
		t.Errorf("Expected dark theme description, got %s", darkTheme.Description)
	}
	if darkTheme.Tokens["color.primary"] != "#1e40af" {
		t.Errorf("Expected dark theme color.primary to be #1e40af, got %v", darkTheme.Tokens["color.primary"])
	}
	if darkTheme.Diff["color.primary"] != "#1e40af" {
		t.Errorf("Expected dark theme diff to contain color.primary")
	}
}

func TestCatalogGenerator_Generate_FiltersNestedMaps(t *testing.T) {
	gen := NewCatalogGenerator()

	// Include a nested map that should be filtered out
	resolvedTokens := map[string]any{
		"color.primary": "#3b82f6",
		"nested.group": map[string]any{
			"should": "be filtered",
		},
	}

	components := map[string]tokens.ComponentDefinition{}

	result, err := gen.Generate(resolvedTokens, components, nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var catalog CatalogSchema
	if err := json.Unmarshal([]byte(result), &catalog); err != nil {
		t.Fatalf("Failed to parse catalog JSON: %v", err)
	}

	// Should only have the atomic token, not the nested map
	if len(catalog.Tokens) != 1 {
		t.Errorf("Expected 1 token (nested map filtered), got %d", len(catalog.Tokens))
	}
	if _, ok := catalog.Tokens["nested.group"]; ok {
		t.Error("Expected nested.group to be filtered out")
	}
}

func TestCatalogGenerator_Generate_ThemeFiltersNestedMaps(t *testing.T) {
	gen := NewCatalogGenerator()

	resolvedTokens := map[string]any{
		"color.primary": "#3b82f6",
	}

	components := map[string]tokens.ComponentDefinition{}

	themes := map[string]CatalogThemeInput{
		"light": {
			Extends: nil,
			ResolvedTokens: map[string]any{
				"color.primary": "#60a5fa",
				"nested.group": map[string]any{
					"should": "be filtered",
				},
			},
			DiffTokens: map[string]any{
				"color.primary": "#60a5fa",
			},
		},
	}

	result, err := gen.Generate(resolvedTokens, components, themes)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var catalog CatalogSchema
	if err := json.Unmarshal([]byte(result), &catalog); err != nil {
		t.Fatalf("Failed to parse catalog JSON: %v", err)
	}

	lightTheme := catalog.Themes["light"]
	if len(lightTheme.Tokens) != 1 {
		t.Errorf("Expected 1 token in theme (nested map filtered), got %d", len(lightTheme.Tokens))
	}
}

func TestCatalogGenerator_Generate_EmptyThemesOmitted(t *testing.T) {
	gen := NewCatalogGenerator()

	resolvedTokens := map[string]any{
		"color.primary": "#3b82f6",
	}

	components := map[string]tokens.ComponentDefinition{}

	// Empty themes map
	themes := map[string]CatalogThemeInput{}

	result, err := gen.Generate(resolvedTokens, components, themes)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var catalog CatalogSchema
	if err := json.Unmarshal([]byte(result), &catalog); err != nil {
		t.Fatalf("Failed to parse catalog JSON: %v", err)
	}

	// Themes should be nil/omitted when empty
	if catalog.Themes != nil {
		t.Error("Expected themes to be nil when empty map provided")
	}

	// Verify JSON doesn't contain "themes" key
	var rawCatalog map[string]any
	if err := json.Unmarshal([]byte(result), &rawCatalog); err != nil {
		t.Fatalf("Failed to parse raw catalog JSON: %v", err)
	}
	if _, ok := rawCatalog["themes"]; ok {
		t.Error("Expected themes key to be omitted from JSON when empty")
	}
}

func TestCatalogGenerator_CategoryFilter_Colors(t *testing.T) {
	opts := CatalogOptions{Category: "color"}
	gen := NewCatalogGeneratorWithOptions(opts)

	resolvedTokens := map[string]any{
		"color.primary":   "#3b82f6",
		"color.secondary": "#8b5cf6",
		"spacing.sm":      "0.5rem",
		"spacing.md":      "1rem",
		"font.family":     "Inter",
	}

	components := map[string]tokens.ComponentDefinition{
		"button": {
			Class: "btn",
			Base:  map[string]any{"display": "inline-flex"},
		},
	}

	result, err := gen.Generate(resolvedTokens, components, nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var catalog CatalogSchema
	if err := json.Unmarshal([]byte(result), &catalog); err != nil {
		t.Fatalf("Failed to parse catalog JSON: %v", err)
	}

	// Should only have color tokens
	if len(catalog.Tokens) != 2 {
		t.Errorf("Expected 2 color tokens, got %d: %v", len(catalog.Tokens), catalog.Tokens)
	}
	if _, ok := catalog.Tokens["color.primary"]; !ok {
		t.Error("Expected color.primary in filtered catalog")
	}
	if _, ok := catalog.Tokens["spacing.sm"]; ok {
		t.Error("Expected spacing.sm to be filtered out")
	}

	// Components should be omitted when filtering to non-component category
	if catalog.Components != nil {
		t.Error("Expected components to be omitted when filtering to color category")
	}

	// Meta should include category
	if catalog.Meta.Category != "color" {
		t.Errorf("Expected meta.category to be 'color', got %s", catalog.Meta.Category)
	}
}

func TestCatalogGenerator_CategoryFilter_Components(t *testing.T) {
	opts := CatalogOptions{Category: "components"}
	gen := NewCatalogGeneratorWithOptions(opts)

	resolvedTokens := map[string]any{
		"color.primary": "#3b82f6",
		"spacing.sm":    "0.5rem",
	}

	components := map[string]tokens.ComponentDefinition{
		"button": {
			Class: "btn",
			Base:  map[string]any{"display": "inline-flex"},
			Variants: map[string]tokens.VariantDef{
				"primary": {Class: "btn-primary", Properties: map[string]any{}},
			},
		},
	}

	result, err := gen.Generate(resolvedTokens, components, nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var catalog CatalogSchema
	if err := json.Unmarshal([]byte(result), &catalog); err != nil {
		t.Fatalf("Failed to parse catalog JSON: %v", err)
	}

	// Should have no tokens (filtering to components)
	if len(catalog.Tokens) != 0 {
		t.Errorf("Expected 0 tokens when filtering to components, got %d", len(catalog.Tokens))
	}

	// Components should be present
	if catalog.Components == nil || len(catalog.Components) != 1 {
		t.Errorf("Expected 1 component, got %v", catalog.Components)
	}
}

func TestCatalogGenerator_CategoryFilter_WithThemes(t *testing.T) {
	opts := CatalogOptions{Category: "color"}
	gen := NewCatalogGeneratorWithOptions(opts)

	resolvedTokens := map[string]any{
		"color.primary": "#3b82f6",
		"spacing.sm":    "0.5rem",
	}

	components := map[string]tokens.ComponentDefinition{}

	themes := map[string]CatalogThemeInput{
		"dark": {
			Extends: nil,
			ResolvedTokens: map[string]any{
				"color.primary": "#1e40af",
				"spacing.sm":    "0.75rem",
			},
			DiffTokens: map[string]any{
				"color.primary": "#1e40af",
				"spacing.sm":    "0.75rem",
			},
		},
	}

	result, err := gen.Generate(resolvedTokens, components, themes)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var catalog CatalogSchema
	if err := json.Unmarshal([]byte(result), &catalog); err != nil {
		t.Fatalf("Failed to parse catalog JSON: %v", err)
	}

	// Should only have color tokens in themes too
	darkTheme := catalog.Themes["dark"]
	if len(darkTheme.Tokens) != 1 {
		t.Errorf("Expected 1 token in theme, got %d: %v", len(darkTheme.Tokens), darkTheme.Tokens)
	}
	if _, ok := darkTheme.Tokens["color.primary"]; !ok {
		t.Error("Expected color.primary in theme tokens")
	}
	if _, ok := darkTheme.Tokens["spacing.sm"]; ok {
		t.Error("Expected spacing.sm to be filtered from theme tokens")
	}
}

func TestCatalogGenerator_CategoryFilter_PluralSingular(t *testing.T) {
	// Test that "colors" matches "color" category
	tests := []struct {
		category string
		tokenKey string
		expected bool
	}{
		{"color", "color.primary", true},
		{"colors", "color.primary", true},
		{"spacing", "spacing.md", true},
		{"spacings", "spacing.md", true},
		{"font", "font.family", true},
		{"fonts", "font.family", true},
		{"color", "spacing.md", false},
		{"spacing", "color.primary", false},
	}

	for _, tt := range tests {
		t.Run(tt.category+"_"+tt.tokenKey, func(t *testing.T) {
			gen := NewCatalogGeneratorWithOptions(CatalogOptions{Category: tt.category})
			result := gen.matchesCategory(tt.tokenKey)
			if result != tt.expected {
				t.Errorf("matchesCategory(%s, %s) = %v, want %v", tt.category, tt.tokenKey, result, tt.expected)
			}
		})
	}
}

func TestFilterAtomicTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: map[string]any{},
		},
		{
			name: "only atomic values",
			input: map[string]any{
				"color.primary": "#3b82f6",
				"spacing.sm":    "0.5rem",
				"opacity.half":  0.5,
			},
			expected: map[string]any{
				"color.primary": "#3b82f6",
				"spacing.sm":    "0.5rem",
				"opacity.half":  0.5,
			},
		},
		{
			name: "mixed atomic and nested",
			input: map[string]any{
				"color.primary": "#3b82f6",
				"nested.group": map[string]any{
					"child": "value",
				},
				"spacing.sm": "0.5rem",
			},
			expected: map[string]any{
				"color.primary": "#3b82f6",
				"spacing.sm":    "0.5rem",
			},
		},
		{
			name: "only nested maps",
			input: map[string]any{
				"group1": map[string]any{"a": "b"},
				"group2": map[string]any{"c": "d"},
			},
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterAtomicTokens(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tokens, got %d", len(tt.expected), len(result))
			}

			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("Expected %s=%v, got %v", k, v, result[k])
				}
			}
		})
	}
}
