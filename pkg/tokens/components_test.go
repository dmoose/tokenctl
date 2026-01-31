package tokens

import (
	"encoding/json"
	"testing"
)

func TestVariantDef_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		input          string
		wantClass      string
		wantProps      map[string]any
		wantStateKeys  []string
		wantErr        bool
	}{
		{
			name:      "class and properties",
			input:     `{"$class":"btn-primary","background":"blue","color":"white"}`,
			wantClass: "btn-primary",
			wantProps: map[string]any{
				"background": "blue",
				"color":      "white",
			},
			wantStateKeys: nil,
		},
		{
			name:  "state with ampersand prefix",
			input: `{"background":"blue","&:hover":{"background":"darkblue"}}`,
			wantProps: map[string]any{
				"background": "blue",
			},
			wantStateKeys: []string{"&:hover"},
		},
		{
			name:  "state with colon prefix",
			input: `{"color":"red",":focus":{"outline":"2px solid blue"},":active":{"color":"darkred"}}`,
			wantProps: map[string]any{
				"color": "red",
			},
			wantStateKeys: []string{":focus", ":active"},
		},
		{
			name:  "mixed class properties and states",
			input: `{"$class":"btn","padding":"8px","&:hover":{"opacity":"0.8"},":focus":{"outline":"none"}}`,
			wantClass: "btn",
			wantProps: map[string]any{
				"padding": "8px",
			},
			wantStateKeys: []string{"&:hover", ":focus"},
		},
		{
			name:          "empty object",
			input:         `{}`,
			wantClass:     "",
			wantProps:     map[string]any{},
			wantStateKeys: nil,
		},
		{
			name:          "only class",
			input:         `{"$class":"solo"}`,
			wantClass:     "solo",
			wantProps:     map[string]any{},
			wantStateKeys: nil,
		},
		{
			name:    "invalid json",
			input:   `{invalid`,
			wantErr: true,
		},
		{
			name:  "state with non-map value",
			input: `{"&:hover":"not-a-map"}`,
			wantProps: map[string]any{},
			wantStateKeys: []string{"&:hover"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var v VariantDef
			err := json.Unmarshal([]byte(tt.input), &v)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if v.Class != tt.wantClass {
				t.Errorf("Class = %q, want %q", v.Class, tt.wantClass)
			}

			if len(v.Properties) != len(tt.wantProps) {
				t.Errorf("Properties count = %d, want %d", len(v.Properties), len(tt.wantProps))
			}
			for k, want := range tt.wantProps {
				got, ok := v.Properties[k]
				if !ok {
					t.Errorf("missing property %q", k)
					continue
				}
				if got != want {
					t.Errorf("Properties[%q] = %v, want %v", k, got, want)
				}
			}

			for _, key := range tt.wantStateKeys {
				if _, ok := v.States[key]; !ok {
					t.Errorf("missing state %q", key)
				}
			}

			expectedStateCount := len(tt.wantStateKeys)
			if len(v.States) != expectedStateCount {
				t.Errorf("States count = %d, want %d", len(v.States), expectedStateCount)
			}
		})
	}
}

func TestVariantDef_UnmarshalJSON_StateProperties(t *testing.T) {
	t.Parallel()
	input := `{"&:hover":{"background":"darkblue","color":"white"}}`

	var v VariantDef
	if err := json.Unmarshal([]byte(input), &v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state, ok := v.States["&:hover"]
	if !ok {
		t.Fatal("missing &:hover state")
	}

	if len(state.Properties) != 2 {
		t.Fatalf("expected 2 state properties, got %d", len(state.Properties))
	}

	if state.Properties["background"] != "darkblue" {
		t.Errorf("state background = %v, want darkblue", state.Properties["background"])
	}
	if state.Properties["color"] != "white" {
		t.Errorf("state color = %v, want white", state.Properties["color"])
	}
}

func TestVariantDef_UnmarshalJSON_NonMapStateHasEmptyProperties(t *testing.T) {
	t.Parallel()
	input := `{"&:hover":"string-value"}`

	var v VariantDef
	if err := json.Unmarshal([]byte(input), &v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state, ok := v.States["&:hover"]
	if !ok {
		t.Fatal("missing &:hover state")
	}

	if len(state.Properties) != 0 {
		t.Errorf("expected 0 state properties for non-map value, got %d", len(state.Properties))
	}
}

func TestExtractComponents_SingleComponent(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"button": map[string]any{
				"$type": "component",
				"base": map[string]any{
					"padding": "8px",
				},
				"variants": map[string]any{
					"primary": map[string]any{
						"$class":     "btn-primary",
						"background": "blue",
					},
				},
				"sizes": map[string]any{
					"sm": map[string]any{
						"padding": "4px",
					},
				},
				"states": map[string]any{
					"disabled": map[string]any{
						"opacity": "0.5",
					},
				},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(components))
	}

	comp, ok := components["button"]
	if !ok {
		t.Fatal("missing 'button' component")
	}

	if comp.Name != "button" {
		t.Errorf("Name = %q, want %q", comp.Name, "button")
	}

	if comp.Base["padding"] != "8px" {
		t.Errorf("Base padding = %v, want '8px'", comp.Base["padding"])
	}
}

func TestExtractComponents_NoComponents(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"color": map[string]any{
				"$type": "color",
				"primary": map[string]any{
					"$value": "#3b82f6",
				},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(components) != 0 {
		t.Errorf("expected 0 components, got %d", len(components))
	}
}

func TestExtractComponents_EmptyDictionary(t *testing.T) {
	t.Parallel()
	dict := NewDictionary()

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(components) != 0 {
		t.Errorf("expected 0 components, got %d", len(components))
	}
}

func TestExtractComponents_NestedComponent(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"ui": map[string]any{
				"form": map[string]any{
					"input": map[string]any{
						"$type": "component",
						"base": map[string]any{
							"border": "1px solid gray",
						},
					},
				},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(components))
	}

	comp, ok := components["ui.form.input"]
	if !ok {
		t.Fatal("missing 'ui.form.input' component")
	}

	if comp.Name != "ui.form.input" {
		t.Errorf("Name = %q, want %q", comp.Name, "ui.form.input")
	}
}

func TestExtractComponents_MultipleComponents(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"button": map[string]any{
				"$type": "component",
				"base":  map[string]any{"padding": "8px"},
			},
			"card": map[string]any{
				"$type": "component",
				"base":  map[string]any{"border-radius": "4px"},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(components))
	}

	if _, ok := components["button"]; !ok {
		t.Error("missing 'button' component")
	}
	if _, ok := components["card"]; !ok {
		t.Error("missing 'card' component")
	}
}

func TestExtractComponents_WithDescription(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"button": map[string]any{
				"$type":        "component",
				"$description": "A clickable button element",
				"base":         map[string]any{},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := components["button"]
	if comp.Description != "A clickable button element" {
		t.Errorf("Description = %q, want %q", comp.Description, "A clickable button element")
	}
}

func TestExtractComponents_WithRequires(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"listItem": map[string]any{
				"$type":     "component",
				"$requires": "list",
				"base":      map[string]any{},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := components["listItem"]
	if comp.Requires != "list" {
		t.Errorf("Requires = %q, want %q", comp.Requires, "list")
	}
}

func TestExtractComponents_WithContains(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"card": map[string]any{
				"$type":     "component",
				"$contains": []any{"cardHeader", "cardBody", "cardFooter"},
				"base":      map[string]any{},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := components["card"]
	if len(comp.Contains) != 3 {
		t.Fatalf("Contains count = %d, want 3", len(comp.Contains))
	}

	expected := []string{"cardHeader", "cardBody", "cardFooter"}
	for i, want := range expected {
		if comp.Contains[i] != want {
			t.Errorf("Contains[%d] = %q, want %q", i, comp.Contains[i], want)
		}
	}
}

func TestExtractComponents_WithAllMetadata(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"dropdown": map[string]any{
				"$type":        "component",
				"$description": "A dropdown menu",
				"$requires":    "nav",
				"$contains":    []any{"dropdownItem"},
				"base": map[string]any{
					"position": "relative",
				},
				"variants": map[string]any{
					"dark": map[string]any{
						"$class":     "dropdown-dark",
						"background": "#333",
					},
				},
				"sizes": map[string]any{
					"lg": map[string]any{
						"min-width": "300px",
					},
				},
				"states": map[string]any{
					"open": map[string]any{
						"display": "block",
					},
				},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp, ok := components["dropdown"]
	if !ok {
		t.Fatal("missing 'dropdown' component")
	}

	if comp.Description != "A dropdown menu" {
		t.Errorf("Description = %q, want %q", comp.Description, "A dropdown menu")
	}
	if comp.Requires != "nav" {
		t.Errorf("Requires = %q, want %q", comp.Requires, "nav")
	}
	if len(comp.Contains) != 1 || comp.Contains[0] != "dropdownItem" {
		t.Errorf("Contains = %v, want [dropdownItem]", comp.Contains)
	}
	if comp.Base["position"] != "relative" {
		t.Errorf("Base position = %v, want 'relative'", comp.Base["position"])
	}
}

func TestExtractComponents_SkipsDollarPrefixedKeys(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"$meta": map[string]any{
				"version": "1.0",
			},
			"button": map[string]any{
				"$type": "component",
				"base":  map[string]any{},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(components))
	}

	if _, ok := components["button"]; !ok {
		t.Error("missing 'button' component")
	}
}

func TestExtractComponents_NonComponentType(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"spacing": map[string]any{
				"$type": "dimension",
				"sm": map[string]any{
					"$value": "4px",
				},
			},
			"button": map[string]any{
				"$type": "component",
				"base":  map[string]any{"padding": "8px"},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(components))
	}

	if _, ok := components["button"]; !ok {
		t.Error("missing 'button' component")
	}
}

func TestExtractComponents_ComponentWithVariantStates(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"button": map[string]any{
				"$type": "component",
				"base":  map[string]any{"cursor": "pointer"},
				"variants": map[string]any{
					"primary": map[string]any{
						"$class":     "btn-primary",
						"background": "blue",
						"&:hover": map[string]any{
							"background": "darkblue",
						},
						":focus": map[string]any{
							"outline": "2px solid blue",
						},
					},
				},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp, ok := components["button"]
	if !ok {
		t.Fatal("missing 'button' component")
	}

	primary, ok := comp.Variants["primary"]
	if !ok {
		t.Fatal("missing 'primary' variant")
	}

	if primary.Class != "btn-primary" {
		t.Errorf("variant Class = %q, want %q", primary.Class, "btn-primary")
	}

	if primary.Properties["background"] != "blue" {
		t.Errorf("variant background = %v, want 'blue'", primary.Properties["background"])
	}

	if _, ok := primary.States["&:hover"]; !ok {
		t.Error("missing &:hover state on variant")
	}
	if _, ok := primary.States[":focus"]; !ok {
		t.Error("missing :focus state on variant")
	}
}

func TestExtractComponents_ComponentClass(t *testing.T) {
	t.Parallel()
	dict := &Dictionary{
		Root: map[string]any{
			"button": map[string]any{
				"$type":  "component",
				"$class": "btn",
				"base":   map[string]any{},
			},
		},
	}

	components, err := dict.ExtractComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := components["button"]
	if comp.Class != "btn" {
		t.Errorf("Class = %q, want %q", comp.Class, "btn")
	}
}
