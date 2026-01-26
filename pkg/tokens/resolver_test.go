package tokens

import (
	"testing"
)

func TestResolveValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		tokens    map[string]any
		input     string
		expected  any
		expectErr bool
	}{
		{
			name: "Simple Resolution",
			tokens: map[string]any{
				"color.red": "#f00",
			},
			input:    "{color.red}",
			expected: "#f00",
		},
		{
			name: "Interpolation",
			tokens: map[string]any{
				"color.red": "#f00",
			},
			input:    "1px solid {color.red}",
			expected: "1px solid #f00",
		},
		{
			name: "Deep Resolution",
			tokens: map[string]any{
				"a": "{b}",
				"b": "{c}",
				"c": "val",
			},
			input:    "{a}",
			expected: "val",
		},
		{
			name: "Missing Reference",
			tokens: map[string]any{
				"a": "val",
			},
			input:     "{missing}",
			expectErr: true,
		},
		{
			name: "Direct Cycle",
			tokens: map[string]any{
				"a": "{b}",
				"b": "{a}",
			},
			input:     "{a}",
			expectErr: true,
		},
		{
			name: "Indirect Cycle",
			tokens: map[string]any{
				"a": "{b}",
				"b": "{c}",
				"c": "{a}",
			},
			input:     "{a}",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Manually construct flat tokens for resolver since we are testing core logic
			r := &Resolver{
				flatTokens: tt.tokens,
				cache:      make(map[string]any),
				stack:      []string{},
			}

			val, err := r.ResolveValue("root", tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if val != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, val)
			}
		})
	}
}
