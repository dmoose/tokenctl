package generators

import (
	"strings"
	"testing"
)

func TestGenerateThemes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		themes       map[string]map[string]any
		defaultTheme string // empty = use DefaultThemeName
		expected     []string
		notExpected  []string
	}{
		{
			name: "single theme with tokens",
			themes: map[string]map[string]any{
				"dark": {
					"color.primary":    "#60a5fa",
					"color.background": "#1e1e1e",
				},
			},
			expected: []string{
				"@layer base {",
				`[data-theme="dark"]`,
				"--color-primary: #60a5fa;",
				"--color-background: #1e1e1e;",
				"}",
			},
		},
		{
			name: "light theme gets root selector",
			themes: map[string]map[string]any{
				"light": {
					"color.surface": "#ffffff",
				},
			},
			expected: []string{
				`:root, [data-theme="light"]`,
				"--color-surface: #ffffff;",
			},
		},
		{
			name: "non-light theme gets data-theme selector only",
			themes: map[string]map[string]any{
				"ocean": {
					"color.primary": "#0077b6",
				},
			},
			expected: []string{
				`[data-theme="ocean"]`,
				"--color-primary: #0077b6;",
			},
			notExpected: []string{
				":root",
			},
		},
		{
			name: "multiple themes in deterministic sorted order",
			themes: map[string]map[string]any{
				"zebra": {
					"color.primary": "black",
				},
				"amber": {
					"color.primary": "#ffbf00",
				},
				"light": {
					"color.primary": "#3b82f6",
				},
			},
			expected: []string{
				"@layer base {",
				`[data-theme="amber"]`,
				`:root, [data-theme="light"]`,
				`[data-theme="zebra"]`,
			},
		},
		{
			name:   "empty themes map",
			themes: map[string]map[string]any{},
			expected: []string{
				"@layer base {",
				"}",
			},
			notExpected: []string{
				"data-theme",
				":root",
			},
		},
		{
			name: "token paths converted from dots to dashes",
			themes: map[string]map[string]any{
				"dark": {
					"color.semantic.success.background": "#22c55e",
					"spacing.layout.page.margin":        "2rem",
				},
			},
			expected: []string{
				"--color-semantic-success-background: #22c55e;",
				"--spacing-layout-page-margin: 2rem;",
			},
		},
		{
			name: "different value types",
			themes: map[string]map[string]any{
				"dark": {
					"color.primary": "#3b82f6",
					"opacity.muted": 0.5,
					"z.index.modal": 1000,
				},
			},
			expected: []string{
				"--color-primary: #3b82f6;",
				"--opacity-muted: 0.5;",
				"--z-index-modal: 1000;",
			},
		},
		{
			name:         "custom default theme",
			defaultTheme: "dark",
			themes: map[string]map[string]any{
				"dark": {
					"color.background": "#1e1e1e",
				},
				"light": {
					"color.background": "#ffffff",
				},
			},
			expected: []string{
				`:root, [data-theme="dark"]`,
				`[data-theme="light"]`,
			},
			notExpected: []string{
				`:root, [data-theme="light"]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := GenerateThemes(tt.themes, tt.defaultTheme)
			if err != nil {
				t.Fatalf("GenerateThemes failed: %v", err)
			}

			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("expected output to contain %q, but it didn't.\nOutput:\n%s", exp, output)
				}
			}

			for _, notExp := range tt.notExpected {
				if strings.Contains(output, notExp) {
					t.Errorf("expected output NOT to contain %q, but it did.\nOutput:\n%s", notExp, output)
				}
			}
		})
	}
}

func TestGenerateThemes_DeterministicOutput(t *testing.T) {
	t.Parallel()

	themes := map[string]map[string]any{
		"dark": {
			"z.last":   "3",
			"a.first":  "1",
			"m.middle": "2",
		},
		"light": {
			"color.primary": "#fff",
		},
	}

	output1, _ := GenerateThemes(themes, "")
	output2, _ := GenerateThemes(themes, "")

	if output1 != output2 {
		t.Error("output should be deterministic across multiple calls")
	}

	aIdx := strings.Index(output1, "--a-first")
	mIdx := strings.Index(output1, "--m-middle")
	zIdx := strings.Index(output1, "--z-last")

	if aIdx > mIdx || mIdx > zIdx {
		t.Error("token variables should be sorted alphabetically within a theme")
	}

	lightIdx := strings.Index(output1, `:root, [data-theme="light"]`)
	darkIdx := strings.Index(output1, `[data-theme="dark"]`)

	if lightIdx > darkIdx {
		t.Error("default theme (light) should come first so non-default themes override :root via cascade")
	}
}
