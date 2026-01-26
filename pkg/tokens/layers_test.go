package tokens

import (
	"strings"
	"testing"
)

func TestCanReference(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		fromLayer Layer
		toLayer   Layer
		expected  bool
	}{
		{
			name:      "Brand to Brand",
			fromLayer: LayerBrand,
			toLayer:   LayerBrand,
			expected:  true,
		},
		{
			name:      "Semantic to Brand",
			fromLayer: LayerSemantic,
			toLayer:   LayerBrand,
			expected:  true,
		},
		{
			name:      "Semantic to Semantic",
			fromLayer: LayerSemantic,
			toLayer:   LayerSemantic,
			expected:  true,
		},
		{
			name:      "Component to Brand",
			fromLayer: LayerComponent,
			toLayer:   LayerBrand,
			expected:  true,
		},
		{
			name:      "Component to Semantic",
			fromLayer: LayerComponent,
			toLayer:   LayerSemantic,
			expected:  true,
		},
		{
			name:      "Component to Component",
			fromLayer: LayerComponent,
			toLayer:   LayerComponent,
			expected:  true,
		},
		{
			name:      "Brand to Semantic",
			fromLayer: LayerBrand,
			toLayer:   LayerSemantic,
			expected:  false,
		},
		{
			name:      "Brand to Component",
			fromLayer: LayerBrand,
			toLayer:   LayerComponent,
			expected:  false,
		},
		{
			name:      "Semantic to Component",
			fromLayer: LayerSemantic,
			toLayer:   LayerComponent,
			expected:  false,
		},
		{
			name:      "Unknown From Layer",
			fromLayer: Layer("unknown"),
			toLayer:   LayerBrand,
			expected:  true,
		},
		{
			name:      "Unknown To Layer",
			fromLayer: LayerBrand,
			toLayer:   Layer("unknown"),
			expected:  true,
		},
		{
			name:      "Both Unknown Layers",
			fromLayer: Layer("foo"),
			toLayer:   Layer("bar"),
			expected:  true,
		},
		{
			name:      "Empty From Layer",
			fromLayer: Layer(""),
			toLayer:   LayerSemantic,
			expected:  true,
		},
		{
			name:      "Empty To Layer",
			fromLayer: LayerComponent,
			toLayer:   Layer(""),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CanReference(tt.fromLayer, tt.toLayer)
			if got != tt.expected {
				t.Errorf("CanReference(%q, %q) = %v, want %v",
					tt.fromLayer, tt.toLayer, got, tt.expected)
			}
		})
	}
}

func TestLayerViolation_Error(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		violation  LayerViolation
		wantParts  []string
		wantAbsent []string
	}{
		{
			name: "Without SourceFile",
			violation: LayerViolation{
				TokenPath:  "button.color",
				TokenLayer: LayerBrand,
				RefPath:    "semantic.primary",
				RefLayer:   LayerSemantic,
			},
			wantParts: []string{
				"button.color",
				"brand",
				"semantic.primary",
				"semantic",
				"layer violation",
			},
		},
		{
			name: "With SourceFile",
			violation: LayerViolation{
				TokenPath:  "button.color",
				TokenLayer: LayerBrand,
				RefPath:    "semantic.primary",
				RefLayer:   LayerSemantic,
				SourceFile: "tokens/button.json",
			},
			wantParts: []string{
				"button.color",
				"brand",
				"tokens/button.json",
				"semantic.primary",
				"semantic",
				"layer violation",
			},
		},
		{
			name: "Empty SourceFile Uses Short Format",
			violation: LayerViolation{
				TokenPath:  "a.b",
				TokenLayer: LayerBrand,
				RefPath:    "c.d",
				RefLayer:   LayerComponent,
				SourceFile: "",
			},
			wantParts:  []string{"a.b [brand] cannot reference c.d [component]"},
			wantAbsent: []string{"[]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.violation.Error()
			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("Error() = %q, want it to contain %q", got, part)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("Error() = %q, should not contain %q", got, absent)
				}
			}
		})
	}
}

func TestNewLayerValidator_GetLayer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		root     map[string]any
		expected map[string]Layer
	}{
		{
			name: "Single Token With Layer",
			root: map[string]any{
				"$layer": "brand",
				"color": map[string]any{
					"$value": "#ff0000",
				},
			},
			expected: map[string]Layer{
				"color": LayerBrand,
			},
		},
		{
			name: "Nested Group Inherits Layer",
			root: map[string]any{
				"$layer": "brand",
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#3b82f6",
					},
					"secondary": map[string]any{
						"$value": "#10b981",
					},
				},
			},
			expected: map[string]Layer{
				"color.primary":   LayerBrand,
				"color.secondary": LayerBrand,
			},
		},
		{
			name: "Override Inherited Layer",
			root: map[string]any{
				"$layer": "brand",
				"primitives": map[string]any{
					"red": map[string]any{
						"$value": "#ff0000",
					},
				},
				"semantic": map[string]any{
					"$layer": "semantic",
					"primary": map[string]any{
						"$value": "{primitives.red}",
					},
				},
			},
			expected: map[string]Layer{
				"primitives.red":  LayerBrand,
				"semantic.primary": LayerSemantic,
			},
		},
		{
			name: "Multiple Layers At Different Levels",
			root: map[string]any{
				"brand": map[string]any{
					"$layer": "brand",
					"red": map[string]any{
						"$value": "#ff0000",
					},
				},
				"semantic": map[string]any{
					"$layer": "semantic",
					"danger": map[string]any{
						"$value": "{brand.red}",
					},
				},
				"component": map[string]any{
					"$layer": "component",
					"button": map[string]any{
						"error": map[string]any{
							"$value": "{semantic.danger}",
						},
					},
				},
			},
			expected: map[string]Layer{
				"brand.red":              LayerBrand,
				"semantic.danger":        LayerSemantic,
				"component.button.error": LayerComponent,
			},
		},
		{
			name: "Token Without Layer",
			root: map[string]any{
				"color": map[string]any{
					"$value": "#000",
				},
			},
			expected: map[string]Layer{},
		},
		{
			name: "Empty Dictionary",
			root: map[string]any{},
			expected: map[string]Layer{},
		},
		{
			name: "Dollar Prefix Keys Skipped",
			root: map[string]any{
				"$layer":      "brand",
				"$extensions": map[string]any{"foo": "bar"},
				"color": map[string]any{
					"$value": "#fff",
				},
			},
			expected: map[string]Layer{
				"color": LayerBrand,
			},
		},
		{
			name: "Deeply Nested Inheritance",
			root: map[string]any{
				"$layer": "component",
				"a": map[string]any{
					"b": map[string]any{
						"c": map[string]any{
							"$value": "deep",
						},
					},
				},
			},
			expected: map[string]Layer{
				"a.b.c": LayerComponent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := &Dictionary{
				Root:        tt.root,
				SourceFiles: make(map[string]string),
			}
			v := NewLayerValidator(d)

			for path, wantLayer := range tt.expected {
				gotLayer := v.GetLayer(path)
				if gotLayer != wantLayer {
					t.Errorf("GetLayer(%q) = %q, want %q", path, gotLayer, wantLayer)
				}
			}
		})
	}
}

func TestGetLayer_UnknownPath(t *testing.T) {
	t.Parallel()
	d := &Dictionary{
		Root: map[string]any{
			"$layer": "brand",
			"color": map[string]any{
				"$value": "#000",
			},
		},
		SourceFiles: make(map[string]string),
	}
	v := NewLayerValidator(d)

	got := v.GetLayer("nonexistent.path")
	if got != "" {
		t.Errorf("GetLayer for unknown path = %q, want empty string", got)
	}
}

func TestValidateReferences(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		root            map[string]any
		sourceFiles     map[string]string
		wantViolations  int
		wantTokenPath   string
		wantRefPath     string
		wantSourceFile  string
	}{
		{
			name: "Valid Semantic References Brand",
			root: map[string]any{
				"brand": map[string]any{
					"$layer": "brand",
					"red": map[string]any{
						"$value": "#ff0000",
					},
				},
				"semantic": map[string]any{
					"$layer": "semantic",
					"danger": map[string]any{
						"$value": "{brand.red}",
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Valid Component References Semantic",
			root: map[string]any{
				"semantic": map[string]any{
					"$layer": "semantic",
					"primary": map[string]any{
						"$value": "#3b82f6",
					},
				},
				"component": map[string]any{
					"$layer": "component",
					"button": map[string]any{
						"bg": map[string]any{
							"$value": "{semantic.primary}",
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Valid Same Layer Reference",
			root: map[string]any{
				"brand": map[string]any{
					"$layer": "brand",
					"red": map[string]any{
						"$value": "#ff0000",
					},
					"alias": map[string]any{
						"$value": "{brand.red}",
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Invalid Brand References Semantic",
			root: map[string]any{
				"brand": map[string]any{
					"$layer": "brand",
					"color": map[string]any{
						"$value": "{semantic.primary}",
					},
				},
				"semantic": map[string]any{
					"$layer": "semantic",
					"primary": map[string]any{
						"$value": "#3b82f6",
					},
				},
			},
			wantViolations: 1,
			wantTokenPath:  "brand.color",
			wantRefPath:    "semantic.primary",
		},
		{
			name: "Invalid Brand References Component",
			root: map[string]any{
				"brand": map[string]any{
					"$layer": "brand",
					"color": map[string]any{
						"$value": "{component.btn.bg}",
					},
				},
				"component": map[string]any{
					"$layer": "component",
					"btn": map[string]any{
						"bg": map[string]any{
							"$value": "#000",
						},
					},
				},
			},
			wantViolations: 1,
			wantTokenPath:  "brand.color",
			wantRefPath:    "component.btn.bg",
		},
		{
			name: "Invalid Semantic References Component",
			root: map[string]any{
				"semantic": map[string]any{
					"$layer": "semantic",
					"color": map[string]any{
						"$value": "{component.btn.bg}",
					},
				},
				"component": map[string]any{
					"$layer": "component",
					"btn": map[string]any{
						"bg": map[string]any{
							"$value": "#000",
						},
					},
				},
			},
			wantViolations: 1,
			wantTokenPath:  "semantic.color",
			wantRefPath:    "component.btn.bg",
		},
		{
			name: "Violation With SourceFile Tracking",
			root: map[string]any{
				"brand": map[string]any{
					"$layer": "brand",
					"color": map[string]any{
						"$value": "{semantic.primary}",
					},
				},
				"semantic": map[string]any{
					"$layer": "semantic",
					"primary": map[string]any{
						"$value": "#3b82f6",
					},
				},
			},
			sourceFiles: map[string]string{
				"brand.color": "tokens/brand.json",
			},
			wantViolations: 1,
			wantTokenPath:  "brand.color",
			wantRefPath:    "semantic.primary",
			wantSourceFile: "tokens/brand.json",
		},
		{
			name: "No Violations When No Layers Assigned",
			root: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#3b82f6",
					},
					"alias": map[string]any{
						"$value": "{color.primary}",
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Non-String Value Skipped",
			root: map[string]any{
				"brand": map[string]any{
					"$layer": "brand",
					"size": map[string]any{
						"$value": 42,
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Reference To Token Without Layer Skipped",
			root: map[string]any{
				"brand": map[string]any{
					"$layer": "brand",
					"color": map[string]any{
						"$value": "{unlayered.token}",
					},
				},
				"unlayered": map[string]any{
					"token": map[string]any{
						"$value": "#000",
					},
				},
			},
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sf := tt.sourceFiles
			if sf == nil {
				sf = make(map[string]string)
			}
			d := &Dictionary{
				Root:        tt.root,
				SourceFiles: sf,
			}
			v := NewLayerValidator(d)
			violations := v.ValidateReferences(d)

			if len(violations) != tt.wantViolations {
				t.Fatalf("got %d violations, want %d: %v",
					len(violations), tt.wantViolations, violations)
			}

			if tt.wantViolations > 0 && len(violations) > 0 {
				viol := violations[0]
				if tt.wantTokenPath != "" && viol.TokenPath != tt.wantTokenPath {
					t.Errorf("TokenPath = %q, want %q", viol.TokenPath, tt.wantTokenPath)
				}
				if tt.wantRefPath != "" && viol.RefPath != tt.wantRefPath {
					t.Errorf("RefPath = %q, want %q", viol.RefPath, tt.wantRefPath)
				}
				if tt.wantSourceFile != "" && viol.SourceFile != tt.wantSourceFile {
					t.Errorf("SourceFile = %q, want %q", viol.SourceFile, tt.wantSourceFile)
				}
			}
		})
	}
}
