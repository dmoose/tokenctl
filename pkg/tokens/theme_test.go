package tokens

import (
	"reflect"
	"strings"
	"testing"
)

func TestInherit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		base     map[string]any
		theme    map[string]any
		expected map[string]any
	}{
		{
			name: "Simple Override",
			base: map[string]any{
				"color.primary":   "#000",
				"color.secondary": "#111",
			},
			theme: map[string]any{
				"color.primary": "#fff",
			},
			expected: map[string]any{
				"color.primary":   "#fff",
				"color.secondary": "#111",
			},
		},
		{
			name: "Add New Token",
			base: map[string]any{
				"a": 1,
			},
			theme: map[string]any{
				"b": 2,
			},
			expected: map[string]any{
				"a": 1,
				"b": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			baseDict := &Dictionary{Root: tt.base}
			themeDict := &Dictionary{Root: tt.theme}

			result, err := Inherit(baseDict, themeDict)
			if err != nil {
				t.Fatalf("Inherit failed: %v", err)
			}

			if !reflect.DeepEqual(result.Root, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result.Root)
			}
		})
	}
}

func TestResolveThemeInheritance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		base      map[string]any
		themes    map[string]map[string]any
		expected  map[string]map[string]any
		expectErr bool
		errMsg    string
	}{
		{
			name: "No Inheritance - Simple Themes",
			base: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]any{
				"light": {
					"color": map[string]any{
						"primary": map[string]any{
							"$value": "#fff",
						},
					},
				},
				"dark": {
					"color": map[string]any{
						"primary": map[string]any{
							"$value": "#111",
						},
					},
				},
			},
			expected: map[string]map[string]any{
				"light": {
					"color": map[string]any{
						"primary": map[string]any{
							"$value": "#fff",
						},
					},
				},
				"dark": {
					"color": map[string]any{
						"primary": map[string]any{
							"$value": "#111",
						},
					},
				},
			},
		},
		{
			name: "Single Level Inheritance",
			base: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#000",
					},
					"secondary": map[string]any{
						"$value": "#666",
					},
				},
			},
			themes: map[string]map[string]any{
				"light": {
					"color": map[string]any{
						"primary": map[string]any{
							"$value": "#fff",
						},
					},
				},
				"dark": {
					"$extends": "light",
					"color": map[string]any{
						"primary": map[string]any{
							"$value": "#111",
						},
					},
				},
			},
			expected: map[string]map[string]any{
				"light": {
					"color": map[string]any{
						"primary": map[string]any{
							"$value": "#fff",
						},
						"secondary": map[string]any{
							"$value": "#666",
						},
					},
				},
				"dark": {
					"color": map[string]any{
						"primary": map[string]any{
							"$value": "#111",
						},
						"secondary": map[string]any{
							"$value": "#666",
						},
					},
				},
			},
		},
		{
			name: "Multi-Level Inheritance Chain",
			base: map[string]any{
				"spacing": map[string]any{
					"base": map[string]any{
						"$value": "1rem",
					},
				},
			},
			themes: map[string]map[string]any{
				"compact": {
					"spacing": map[string]any{
						"base": map[string]any{
							"$value": "0.5rem",
						},
					},
				},
				"comfortable": {
					"$extends": "compact",
					"spacing": map[string]any{
						"base": map[string]any{
							"$value": "1.5rem",
						},
					},
				},
				"spacious": {
					"$extends": "comfortable",
					"spacing": map[string]any{
						"base": map[string]any{
							"$value": "2rem",
						},
					},
				},
			},
			expected: map[string]map[string]any{
				"compact": {
					"spacing": map[string]any{
						"base": map[string]any{
							"$value": "0.5rem",
						},
					},
				},
				"comfortable": {
					"spacing": map[string]any{
						"base": map[string]any{
							"$value": "1.5rem",
						},
					},
				},
				"spacious": {
					"spacing": map[string]any{
						"base": map[string]any{
							"$value": "2rem",
						},
					},
				},
			},
		},
		{
			name: "Direct Circular Dependency",
			base: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]any{
				"theme-a": {
					"$extends": "theme-a",
				},
			},
			expectErr: true,
			errMsg:    "circular theme inheritance",
		},
		{
			name: "Indirect Circular Dependency",
			base: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]any{
				"theme-a": {
					"$extends": "theme-b",
				},
				"theme-b": {
					"$extends": "theme-c",
				},
				"theme-c": {
					"$extends": "theme-a",
				},
			},
			expectErr: true,
			errMsg:    "circular theme inheritance",
		},
		{
			name: "Missing Parent Theme",
			base: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]any{
				"child": {
					"$extends": "nonexistent",
				},
			},
			expectErr: true,
			errMsg:    "theme 'nonexistent' not found",
		},
		{
			name: "Invalid $extends Type",
			base: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]any{
				"invalid": {
					"$extends": 123, // Should be string
				},
			},
			expectErr: true,
			errMsg:    "invalid $extends value",
		},
		{
			name: "$extends Metadata is Cleaned",
			base: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]any{
				"parent": {
					"color": map[string]any{
						"primary": map[string]any{
							"$value": "#fff",
						},
					},
				},
				"child": {
					"$extends": "parent",
					"$schema":  "should-be-removed",
				},
			},
			expected: map[string]map[string]any{
				"parent": {
					"color": map[string]any{
						"primary": map[string]any{
							"$value": "#fff",
						},
					},
				},
				"child": {
					"color": map[string]any{
						"primary": map[string]any{
							"$value": "#fff",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			baseDict := &Dictionary{Root: tt.base}
			themeDicts := make(map[string]*Dictionary)
			for name, root := range tt.themes {
				themeDicts[name] = &Dictionary{Root: root}
			}

			result, err := ResolveThemeInheritance(baseDict, themeDicts)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tt.errMsg)
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d themes, got %d", len(tt.expected), len(result))
			}

			for name, expectedRoot := range tt.expected {
				resultDict, ok := result[name]
				if !ok {
					t.Errorf("Expected theme '%s' not found in result", name)
					continue
				}

				if !reflect.DeepEqual(resultDict.Root, expectedRoot) {
					t.Errorf("Theme '%s' mismatch:\nExpected: %+v\nGot: %+v", name, expectedRoot, resultDict.Root)
				}

				// Verify metadata is cleaned
				if _, hasExtends := resultDict.Root["$extends"]; hasExtends {
					t.Errorf("Theme '%s' still contains $extends metadata", name)
				}
				if _, hasSchema := resultDict.Root["$schema"]; hasSchema {
					t.Errorf("Theme '%s' still contains $schema metadata", name)
				}
			}
		})
	}
}

func TestDiff(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		target   map[string]any
		base     map[string]any
		expected map[string]any
	}{
		{
			name: "No Changes",
			target: map[string]any{
				"a": 1,
				"b": 2,
			},
			base: map[string]any{
				"a": 1,
				"b": 2,
			},
			expected: map[string]any{},
		},
		{
			name: "One Change",
			target: map[string]any{
				"a": 1,
				"b": 3, // Changed from 2
			},
			base: map[string]any{
				"a": 1,
				"b": 2,
			},
			expected: map[string]any{
				"b": 3,
			},
		},
		{
			name: "New Key",
			target: map[string]any{
				"a": 1,
				"c": 4, // New
			},
			base: map[string]any{
				"a": 1,
			},
			expected: map[string]any{
				"c": 4,
			},
		},
		{
			name: "Missing Key in Target (Should Ignore)",
			// If target doesn't have it, it's not in the diff (inheritance implies target is a superset)
			// But Diff function logic is: keys FROM target that differ.
			// So if base has 'a' but target doesn't, 'a' is not in output.
			target: map[string]any{
				"b": 2,
			},
			base: map[string]any{
				"a": 1,
				"b": 2,
			},
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Diff(tt.target, tt.base)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
