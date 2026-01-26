package tokens

import (
	"reflect"
	"sort"
	"testing"
)

func TestExtractMetadata_BasicToken(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"color": map[string]any{
			"$type": "color",
			"primary": map[string]any{
				"$value": "#3b82f6",
			},
		},
	}

	result := ExtractMetadata(d)

	meta, ok := result["color.primary"]
	if !ok {
		t.Fatal("expected metadata for color.primary")
	}
	if meta.Path != "color.primary" {
		t.Errorf("Path = %q, want %q", meta.Path, "color.primary")
	}
	if meta.Value != "#3b82f6" {
		t.Errorf("Value = %v, want %q", meta.Value, "#3b82f6")
	}
	if meta.Type != "color" {
		t.Errorf("Type = %q, want %q", meta.Type, "color")
	}
}

func TestExtractMetadata_TypeInheritance(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"spacing": map[string]any{
			"$type": "dimension",
			"sm": map[string]any{
				"$value": "4px",
			},
			"md": map[string]any{
				"$value": "8px",
			},
			"nested": map[string]any{
				"lg": map[string]any{
					"$value": "16px",
				},
			},
		},
	}

	result := ExtractMetadata(d)

	for _, path := range []string{"spacing.sm", "spacing.md", "spacing.nested.lg"} {
		meta, ok := result[path]
		if !ok {
			t.Fatalf("expected metadata for %s", path)
		}
		if meta.Type != "dimension" {
			t.Errorf("%s: Type = %q, want %q", path, meta.Type, "dimension")
		}
	}
}

func TestExtractMetadata_TypeOverride(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"group": map[string]any{
			"$type": "color",
			"child": map[string]any{
				"$type": "dimension",
				"token": map[string]any{
					"$value": "10px",
				},
			},
		},
	}

	result := ExtractMetadata(d)

	meta, ok := result["group.child.token"]
	if !ok {
		t.Fatal("expected metadata for group.child.token")
	}
	if meta.Type != "dimension" {
		t.Errorf("Type = %q, want %q", meta.Type, "dimension")
	}
}

func TestExtractMetadata_Description(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"brand": map[string]any{
			"$value":       "#ff0000",
			"$description": "Primary brand color used across the app",
		},
	}

	result := ExtractMetadata(d)

	meta, ok := result["brand"]
	if !ok {
		t.Fatal("expected metadata for brand")
	}
	if meta.Description != "Primary brand color used across the app" {
		t.Errorf("Description = %q, want %q", meta.Description, "Primary brand color used across the app")
	}
}

func TestExtractMetadata_Usage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		usage     any
		wantUsage []string
	}{
		{
			name:      "string form",
			usage:     "backgrounds",
			wantUsage: []string{"backgrounds"},
		},
		{
			name:      "array of any form",
			usage:     []any{"backgrounds", "borders", "text"},
			wantUsage: []string{"backgrounds", "borders", "text"},
		},
		{
			name:      "string slice form",
			usage:     []string{"buttons", "cards"},
			wantUsage: []string{"buttons", "cards"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := NewDictionary()
			d.Root = map[string]any{
				"token": map[string]any{
					"$value": "value",
					"$usage": tt.usage,
				},
			}

			result := ExtractMetadata(d)
			meta, ok := result["token"]
			if !ok {
				t.Fatal("expected metadata for token")
			}
			if !reflect.DeepEqual(meta.Usage, tt.wantUsage) {
				t.Errorf("Usage = %v, want %v", meta.Usage, tt.wantUsage)
			}
		})
	}
}

func TestExtractMetadata_Avoid(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"color": map[string]any{
			"legacy": map[string]any{
				"$value": "#999",
				"$avoid": "Use color.neutral instead",
			},
		},
	}

	result := ExtractMetadata(d)

	meta, ok := result["color.legacy"]
	if !ok {
		t.Fatal("expected metadata for color.legacy")
	}
	if meta.Avoid != "Use color.neutral instead" {
		t.Errorf("Avoid = %q, want %q", meta.Avoid, "Use color.neutral instead")
	}
}

func TestExtractMetadata_Deprecated(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		deprecated     any
		wantDeprecated any
	}{
		{
			name:           "bool form",
			deprecated:     true,
			wantDeprecated: true,
		},
		{
			name:           "string reason",
			deprecated:     "Use spacing.gap instead",
			wantDeprecated: "Use spacing.gap instead",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := NewDictionary()
			d.Root = map[string]any{
				"token": map[string]any{
					"$value":      "old-value",
					"$deprecated": tt.deprecated,
				},
			}

			result := ExtractMetadata(d)
			meta, ok := result["token"]
			if !ok {
				t.Fatal("expected metadata for token")
			}
			if meta.Deprecated != tt.wantDeprecated {
				t.Errorf("Deprecated = %v, want %v", meta.Deprecated, tt.wantDeprecated)
			}
		})
	}
}

func TestExtractMetadata_Customizable(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"theme": map[string]any{
			"accent": map[string]any{
				"$value":        "#0066ff",
				"$customizable": true,
			},
			"fixed": map[string]any{
				"$value":        "#000000",
				"$customizable": false,
			},
			"unset": map[string]any{
				"$value": "#ffffff",
			},
		},
	}

	result := ExtractMetadata(d)

	if meta := result["theme.accent"]; meta == nil {
		t.Fatal("expected metadata for theme.accent")
	} else if !meta.Customizable {
		t.Error("theme.accent: Customizable = false, want true")
	}

	if meta := result["theme.fixed"]; meta == nil {
		t.Fatal("expected metadata for theme.fixed")
	} else if meta.Customizable {
		t.Error("theme.fixed: Customizable = true, want false")
	}

	if meta := result["theme.unset"]; meta == nil {
		t.Fatal("expected metadata for theme.unset")
	} else if meta.Customizable {
		t.Error("theme.unset: Customizable = true, want false (default)")
	}
}

func TestExtractMetadata_SourceFile(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"color": map[string]any{
			"primary": map[string]any{
				"$value": "#3b82f6",
			},
		},
	}
	d.SourceFiles["color.primary"] = "tokens/color.json"

	result := ExtractMetadata(d)

	meta, ok := result["color.primary"]
	if !ok {
		t.Fatal("expected metadata for color.primary")
	}
	if meta.SourceFile != "tokens/color.json" {
		t.Errorf("SourceFile = %q, want %q", meta.SourceFile, "tokens/color.json")
	}
}

func TestExtractMetadata_SourceFileNotPresent(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"token": map[string]any{
			"$value": "val",
		},
	}

	result := ExtractMetadata(d)

	meta, ok := result["token"]
	if !ok {
		t.Fatal("expected metadata for token")
	}
	if meta.SourceFile != "" {
		t.Errorf("SourceFile = %q, want empty", meta.SourceFile)
	}
}

func TestExtractMetadata_NestedGroups(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"$type": "color",
		"brand": map[string]any{
			"primary": map[string]any{
				"$value": "#0066ff",
			},
			"secondary": map[string]any{
				"$value": "#8b5cf6",
			},
		},
		"spacing": map[string]any{
			"$type": "dimension",
			"layout": map[string]any{
				"gutter": map[string]any{
					"$value": "24px",
				},
				"margin": map[string]any{
					"$value": "16px",
				},
			},
		},
	}
	d.SourceFiles["brand.primary"] = "color.json"
	d.SourceFiles["spacing.layout.gutter"] = "spacing.json"

	result := ExtractMetadata(d)

	if len(result) != 4 {
		t.Fatalf("got %d tokens, want 4", len(result))
	}

	expected := map[string]struct {
		value      any
		tokenType  string
		sourceFile string
	}{
		"brand.primary":        {value: "#0066ff", tokenType: "color", sourceFile: "color.json"},
		"brand.secondary":      {value: "#8b5cf6", tokenType: "color", sourceFile: ""},
		"spacing.layout.gutter": {value: "24px", tokenType: "dimension", sourceFile: "spacing.json"},
		"spacing.layout.margin": {value: "16px", tokenType: "dimension", sourceFile: ""},
	}

	for path, want := range expected {
		meta, ok := result[path]
		if !ok {
			t.Errorf("missing metadata for %s", path)
			continue
		}
		if meta.Value != want.value {
			t.Errorf("%s: Value = %v, want %v", path, meta.Value, want.value)
		}
		if meta.Type != want.tokenType {
			t.Errorf("%s: Type = %q, want %q", path, meta.Type, want.tokenType)
		}
		if meta.SourceFile != want.sourceFile {
			t.Errorf("%s: SourceFile = %q, want %q", path, meta.SourceFile, want.sourceFile)
		}
	}
}

func TestExtractMetadata_EmptyDictionary(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	result := ExtractMetadata(d)

	if len(result) != 0 {
		t.Errorf("got %d tokens, want 0", len(result))
	}
}

func TestExtractMetadata_EmptyGroups(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"color": map[string]any{
			"$type": "color",
		},
	}

	result := ExtractMetadata(d)

	if len(result) != 0 {
		t.Errorf("got %d tokens, want 0", len(result))
	}
}

func TestExtractMetadata_AllFieldsPopulated(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"ns": map[string]any{
			"$type": "color",
			"token": map[string]any{
				"$value":        "#abcdef",
				"$description":  "A fully populated token",
				"$usage":        []any{"backgrounds", "borders"},
				"$avoid":        "Do not use in headers",
				"$deprecated":   "Replaced by ns.token-v2",
				"$customizable": true,
			},
		},
	}
	d.SourceFiles["ns.token"] = "ns.json"

	result := ExtractMetadata(d)

	meta, ok := result["ns.token"]
	if !ok {
		t.Fatal("expected metadata for ns.token")
	}
	if meta.Path != "ns.token" {
		t.Errorf("Path = %q, want %q", meta.Path, "ns.token")
	}
	if meta.Value != "#abcdef" {
		t.Errorf("Value = %v, want %q", meta.Value, "#abcdef")
	}
	if meta.Type != "color" {
		t.Errorf("Type = %q, want %q", meta.Type, "color")
	}
	if meta.Description != "A fully populated token" {
		t.Errorf("Description = %q, want %q", meta.Description, "A fully populated token")
	}
	wantUsage := []string{"backgrounds", "borders"}
	if !reflect.DeepEqual(meta.Usage, wantUsage) {
		t.Errorf("Usage = %v, want %v", meta.Usage, wantUsage)
	}
	if meta.Avoid != "Do not use in headers" {
		t.Errorf("Avoid = %q, want %q", meta.Avoid, "Do not use in headers")
	}
	if meta.Deprecated != "Replaced by ns.token-v2" {
		t.Errorf("Deprecated = %v, want %q", meta.Deprecated, "Replaced by ns.token-v2")
	}
	if !meta.Customizable {
		t.Error("Customizable = false, want true")
	}
	if meta.SourceFile != "ns.json" {
		t.Errorf("SourceFile = %q, want %q", meta.SourceFile, "ns.json")
	}
}

func TestExtractMetadata_TopLevelToken(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"standalone": map[string]any{
			"$value": "42",
			"$type":  "number",
		},
	}

	result := ExtractMetadata(d)

	meta, ok := result["standalone"]
	if !ok {
		t.Fatal("expected metadata for standalone")
	}
	if meta.Path != "standalone" {
		t.Errorf("Path = %q, want %q", meta.Path, "standalone")
	}
	if meta.Type != "number" {
		t.Errorf("Type = %q, want %q", meta.Type, "number")
	}
}

func TestExtractMetadata_DollarKeysSkipped(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"$type":        "color",
		"$description": "Root level description",
		"valid": map[string]any{
			"$value": "#fff",
		},
	}

	result := ExtractMetadata(d)

	if len(result) != 1 {
		t.Errorf("got %d tokens, want 1", len(result))
	}
	if _, ok := result["valid"]; !ok {
		t.Error("expected metadata for valid")
	}
}

func TestExtractMetadata_MultipleTopLevelGroups(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"color": map[string]any{
			"$type": "color",
			"red": map[string]any{
				"$value": "#f00",
			},
		},
		"size": map[string]any{
			"$type": "dimension",
			"sm": map[string]any{
				"$value": "8px",
			},
		},
		"opacity": map[string]any{
			"$type": "number",
			"half": map[string]any{
				"$value": 0.5,
			},
		},
	}

	result := ExtractMetadata(d)

	if len(result) != 3 {
		t.Fatalf("got %d tokens, want 3", len(result))
	}

	paths := make([]string, 0, len(result))
	for p := range result {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	wantPaths := []string{"color.red", "opacity.half", "size.sm"}
	if !reflect.DeepEqual(paths, wantPaths) {
		t.Errorf("paths = %v, want %v", paths, wantPaths)
	}
}

func TestExtractMetadata_UsageWithMixedArrayTypes(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"token": map[string]any{
			"$value": "val",
			"$usage": []any{"valid-string", 42, "another-string"},
		},
	}

	result := ExtractMetadata(d)

	meta, ok := result["token"]
	if !ok {
		t.Fatal("expected metadata for token")
	}
	wantUsage := []string{"valid-string", "another-string"}
	if !reflect.DeepEqual(meta.Usage, wantUsage) {
		t.Errorf("Usage = %v, want %v", meta.Usage, wantUsage)
	}
}

func TestExtractMetadata_NoTypeInherited(t *testing.T) {
	t.Parallel()
	d := NewDictionary()
	d.Root = map[string]any{
		"misc": map[string]any{
			"token": map[string]any{
				"$value": "something",
			},
		},
	}

	result := ExtractMetadata(d)

	meta, ok := result["misc.token"]
	if !ok {
		t.Fatal("expected metadata for misc.token")
	}
	if meta.Type != "" {
		t.Errorf("Type = %q, want empty string", meta.Type)
	}
}
