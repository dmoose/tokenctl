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
				"color.primary": "#000",
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
