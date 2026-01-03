package generators

import (
	"testing"

	"github.com/dmoose/tokctl/pkg/tokens"
)

func TestTailwindGenerator(t *testing.T) {
	tests := []struct {
		name       string
		tokens     map[string]interface{}
		components map[string]tokens.ComponentDefinition
		expected   []string
		notExpected []string
	}{
		{
			name: "Metadata Filtering",
			components: map[string]tokens.ComponentDefinition{
				"button": {
					Class: "btn",
					Base: map[string]interface{}{
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
					Base: map[string]interface{}{
						"box-shadow": []interface{}{
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
					Base: map[string]interface{}{
						"color": "{color.primary}",
					},
				},
			},
			expected: []string{
				"color: var(--color-primary);",
			},
		},
	}

	gen := NewTailwindGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := gen.GenerateComponents(tt.components)
			if err != nil {
				t.Fatalf("GenerateComponents failed: %v", err)
			}

			for _, exp := range tt.expected {
				if !containsString(output, exp) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", exp, output)
				}
			}

			for _, notExp := range tt.notExpected {
				if containsString(output, notExp) {
					t.Errorf("Expected output NOT to contain %q, but it did.\nOutput:\n%s", notExp, output)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && len(s)-len(substr) >= 0 && (s == substr || (len(s) > len(substr) && (s[0:len(substr)] == substr || s[len(s)-len(substr):] == substr || func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())))
}
