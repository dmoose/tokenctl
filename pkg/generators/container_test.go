package generators

import (
	"strings"
	"testing"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

func TestGenerateContainerCSS(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		overrides []tokens.ContainerOverride
		wantParts []string
		wantEmpty bool
	}{
		{
			name:      "nil overrides returns empty",
			overrides: nil,
			wantEmpty: true,
		},
		{
			name:      "empty overrides returns empty",
			overrides: []tokens.ContainerOverride{},
			wantEmpty: true,
		},
		{
			name: "single override produces correct @container block",
			overrides: []tokens.ContainerOverride{
				{
					ContainerQuery: "blogpost (max-width: 1024px)",
					ComponentClass: "blogpost-sidebar",
					Properties:     map[string]any{"display": "none"},
				},
			},
			wantParts: []string{
				"@container blogpost (max-width: 1024px) {",
				".blogpost-sidebar {",
				"display: none;",
			},
		},
		{
			name: "multiple overrides same query grouped under one @container",
			overrides: []tokens.ContainerOverride{
				{
					ContainerQuery: "blogpost (max-width: 1024px)",
					ComponentClass: "blogpost-sidebar",
					Properties:     map[string]any{"display": "none"},
				},
				{
					ContainerQuery: "blogpost (max-width: 1024px)",
					ComponentClass: "blogpost-layout",
					Properties:     map[string]any{"display": "block"},
				},
				{
					ContainerQuery: "blogpost (max-width: 1024px)",
					ComponentClass: "blogpost-content",
					Properties:     map[string]any{"width": "100%"},
				},
			},
			wantParts: []string{
				"@container blogpost (max-width: 1024px) {",
				".blogpost-content {",
				"width: 100%;",
				".blogpost-layout {",
				"display: block;",
				".blogpost-sidebar {",
				"display: none;",
			},
		},
		{
			name: "different queries produce separate blocks",
			overrides: []tokens.ContainerOverride{
				{
					ContainerQuery: "card (max-width: 300px)",
					ComponentClass: "card-body",
					Properties:     map[string]any{"padding": "8px"},
				},
				{
					ContainerQuery: "blogpost (max-width: 1024px)",
					ComponentClass: "blogpost-sidebar",
					Properties:     map[string]any{"display": "none"},
				},
			},
			wantParts: []string{
				"@container blogpost (max-width: 1024px) {",
				"@container card (max-width: 300px) {",
			},
		},
		{
			name: "token references are resolved",
			overrides: []tokens.ContainerOverride{
				{
					ContainerQuery: "main (max-width: 800px)",
					ComponentClass: "sidebar",
					Properties:     map[string]any{"padding": "{spacing.lg}"},
				},
			},
			wantParts: []string{
				"padding: var(--spacing-lg);",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GenerateContainerCSS(tt.overrides)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("GenerateContainerCSS() = %q, want empty string", got)
				}
				return
			}

			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("GenerateContainerCSS() missing %q in output:\n%s", part, got)
				}
			}
		})
	}
}

func TestGenerateContainerCSS_DeterministicOrdering(t *testing.T) {
	t.Parallel()

	overrides := []tokens.ContainerOverride{
		{
			ContainerQuery: "z-container (max-width: 500px)",
			ComponentClass: "z-child",
			Properties:     map[string]any{"display": "none"},
		},
		{
			ContainerQuery: "a-container (max-width: 500px)",
			ComponentClass: "a-child",
			Properties:     map[string]any{"display": "block"},
		},
	}

	got := GenerateContainerCSS(overrides)

	aIdx := strings.Index(got, "a-container")
	zIdx := strings.Index(got, "z-container")

	if aIdx == -1 || zIdx == -1 {
		t.Fatalf("Missing container query in output:\n%s", got)
	}

	if aIdx > zIdx {
		t.Errorf("Container queries not in alphabetical order (a=%d, z=%d) in output:\n%s", aIdx, zIdx, got)
	}
}

func TestGenerateContainerCSS_SelectorOrderWithinQuery(t *testing.T) {
	t.Parallel()

	overrides := []tokens.ContainerOverride{
		{
			ContainerQuery: "main (max-width: 800px)",
			ComponentClass: "z-component",
			Properties:     map[string]any{"display": "none"},
		},
		{
			ContainerQuery: "main (max-width: 800px)",
			ComponentClass: "a-component",
			Properties:     map[string]any{"display": "block"},
		},
	}

	got := GenerateContainerCSS(overrides)

	aIdx := strings.Index(got, ".a-component")
	zIdx := strings.Index(got, ".z-component")

	if aIdx == -1 || zIdx == -1 {
		t.Fatalf("Missing selectors in output:\n%s", got)
	}

	if aIdx > zIdx {
		t.Errorf("Selectors not in alphabetical order within query (a=%d, z=%d) in output:\n%s", aIdx, zIdx, got)
	}
}

func TestGenerateContainerCSS_NotInsideLayer(t *testing.T) {
	t.Parallel()

	overrides := []tokens.ContainerOverride{
		{
			ContainerQuery: "main (max-width: 800px)",
			ComponentClass: "sidebar",
			Properties:     map[string]any{"display": "none"},
		},
	}

	got := GenerateContainerCSS(overrides)

	if strings.Contains(got, "@layer") {
		t.Errorf("GenerateContainerCSS() should NOT contain @layer, but got:\n%s", got)
	}
}
