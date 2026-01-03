package tokens

import (
	"reflect"
	"testing"
)

func TestInherit(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]interface{}
		theme    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "Simple Override",
			base: map[string]interface{}{
				"color.primary":   "#000",
				"color.secondary": "#111",
			},
			theme: map[string]interface{}{
				"color.primary": "#fff",
			},
			expected: map[string]interface{}{
				"color.primary":   "#fff",
				"color.secondary": "#111",
			},
		},
		{
			name: "Add New Token",
			base: map[string]interface{}{
				"a": 1,
			},
			theme: map[string]interface{}{
				"b": 2,
			},
			expected: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	tests := []struct {
		name      string
		base      map[string]interface{}
		themes    map[string]map[string]interface{}
		expected  map[string]map[string]interface{}
		expectErr bool
		errMsg    string
	}{
		{
			name: "No Inheritance - Simple Themes",
			base: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]interface{}{
				"light": {
					"color": map[string]interface{}{
						"primary": map[string]interface{}{
							"$value": "#fff",
						},
					},
				},
				"dark": {
					"color": map[string]interface{}{
						"primary": map[string]interface{}{
							"$value": "#111",
						},
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"light": {
					"color": map[string]interface{}{
						"primary": map[string]interface{}{
							"$value": "#fff",
						},
					},
				},
				"dark": {
					"color": map[string]interface{}{
						"primary": map[string]interface{}{
							"$value": "#111",
						},
					},
				},
			},
		},
		{
			name: "Single Level Inheritance",
			base: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "#000",
					},
					"secondary": map[string]interface{}{
						"$value": "#666",
					},
				},
			},
			themes: map[string]map[string]interface{}{
				"light": {
					"color": map[string]interface{}{
						"primary": map[string]interface{}{
							"$value": "#fff",
						},
					},
				},
				"dark": {
					"$extends": "light",
					"color": map[string]interface{}{
						"primary": map[string]interface{}{
							"$value": "#111",
						},
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"light": {
					"color": map[string]interface{}{
						"primary": map[string]interface{}{
							"$value": "#fff",
						},
						"secondary": map[string]interface{}{
							"$value": "#666",
						},
					},
				},
				"dark": {
					"color": map[string]interface{}{
						"primary": map[string]interface{}{
							"$value": "#111",
						},
						"secondary": map[string]interface{}{
							"$value": "#666",
						},
					},
				},
			},
		},
		{
			name: "Multi-Level Inheritance Chain",
			base: map[string]interface{}{
				"spacing": map[string]interface{}{
					"base": map[string]interface{}{
						"$value": "1rem",
					},
				},
			},
			themes: map[string]map[string]interface{}{
				"compact": {
					"spacing": map[string]interface{}{
						"base": map[string]interface{}{
							"$value": "0.5rem",
						},
					},
				},
				"comfortable": {
					"$extends": "compact",
					"spacing": map[string]interface{}{
						"base": map[string]interface{}{
							"$value": "1.5rem",
						},
					},
				},
				"spacious": {
					"$extends": "comfortable",
					"spacing": map[string]interface{}{
						"base": map[string]interface{}{
							"$value": "2rem",
						},
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"compact": {
					"spacing": map[string]interface{}{
						"base": map[string]interface{}{
							"$value": "0.5rem",
						},
					},
				},
				"comfortable": {
					"spacing": map[string]interface{}{
						"base": map[string]interface{}{
							"$value": "1.5rem",
						},
					},
				},
				"spacious": {
					"spacing": map[string]interface{}{
						"base": map[string]interface{}{
							"$value": "2rem",
						},
					},
				},
			},
		},
		{
			name: "Direct Circular Dependency",
			base: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]interface{}{
				"theme-a": {
					"$extends": "theme-a",
				},
			},
			expectErr: true,
			errMsg:    "circular theme inheritance",
		},
		{
			name: "Indirect Circular Dependency",
			base: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]interface{}{
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
			base: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]interface{}{
				"child": {
					"$extends": "nonexistent",
				},
			},
			expectErr: true,
			errMsg:    "theme 'nonexistent' not found",
		},
		{
			name: "Invalid $extends Type",
			base: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]interface{}{
				"invalid": {
					"$extends": 123, // Should be string
				},
			},
			expectErr: true,
			errMsg:    "invalid $extends value",
		},
		{
			name: "$extends Metadata is Cleaned",
			base: map[string]interface{}{
				"color": map[string]interface{}{
					"primary": map[string]interface{}{
						"$value": "#000",
					},
				},
			},
			themes: map[string]map[string]interface{}{
				"parent": {
					"color": map[string]interface{}{
						"primary": map[string]interface{}{
							"$value": "#fff",
						},
					},
				},
				"child": {
					"$extends": "parent",
					"$schema":  "should-be-removed",
				},
			},
			expected: map[string]map[string]interface{}{
				"parent": {
					"color": map[string]interface{}{
						"primary": map[string]interface{}{
							"$value": "#fff",
						},
					},
				},
				"child": {
					"color": map[string]interface{}{
						"primary": map[string]interface{}{
							"$value": "#fff",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && len(substr) > 0 && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}

func TestDiff(t *testing.T) {
	tests := []struct {
		name     string
		target   map[string]interface{}
		base     map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "No Changes",
			target: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			base: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			expected: map[string]interface{}{},
		},
		{
			name: "One Change",
			target: map[string]interface{}{
				"a": 1,
				"b": 3, // Changed from 2
			},
			base: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			expected: map[string]interface{}{
				"b": 3,
			},
		},
		{
			name: "New Key",
			target: map[string]interface{}{
				"a": 1,
				"c": 4, // New
			},
			base: map[string]interface{}{
				"a": 1,
			},
			expected: map[string]interface{}{
				"c": 4,
			},
		},
		{
			name: "Missing Key in Target (Should Ignore)",
			// If target doesn't have it, it's not in the diff (inheritance implies target is a superset)
			// But Diff function logic is: keys FROM target that differ.
			// So if base has 'a' but target doesn't, 'a' is not in output.
			target: map[string]interface{}{
				"b": 2,
			},
			base: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Diff(tt.target, tt.base)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
