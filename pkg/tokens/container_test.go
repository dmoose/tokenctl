package tokens

import (
	"testing"
)

func TestExtractContainerOverrides(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		components map[string]ComponentDefinition
		wantCount  int
		validate   func(t *testing.T, results []ContainerOverride)
	}{
		{
			name:       "no components returns empty",
			components: map[string]ComponentDefinition{},
			wantCount:  0,
		},
		{
			name: "component without container overrides returns empty",
			components: map[string]ComponentDefinition{
				"button": {
					Name:  "button",
					Class: "btn",
					Base:  map[string]any{"display": "inline-flex"},
				},
			},
			wantCount: 0,
		},
		{
			name: "single component single override",
			components: map[string]ComponentDefinition{
				"blogpost-sidebar": {
					Name:  "blogpost-sidebar",
					Class: "blogpost-sidebar",
					ContainerOverrides: map[string]map[string]any{
						"blogpost (max-width: 1024px)": {
							"display": "none",
						},
					},
				},
			},
			wantCount: 1,
			validate: func(t *testing.T, results []ContainerOverride) {
				r := results[0]
				if r.ContainerQuery != "blogpost (max-width: 1024px)" {
					t.Errorf("ContainerQuery = %q, want %q", r.ContainerQuery, "blogpost (max-width: 1024px)")
				}
				if r.ComponentClass != "blogpost-sidebar" {
					t.Errorf("ComponentClass = %q, want %q", r.ComponentClass, "blogpost-sidebar")
				}
				if r.Properties["display"] != "none" {
					t.Errorf("Properties[display] = %v, want %q", r.Properties["display"], "none")
				}
			},
		},
		{
			name: "single component multiple queries",
			components: map[string]ComponentDefinition{
				"sidebar": {
					Name:  "sidebar",
					Class: "sidebar",
					ContainerOverrides: map[string]map[string]any{
						"main (max-width: 800px)": {
							"display": "none",
						},
						"main (max-width: 400px)": {
							"font-size": "12px",
						},
					},
				},
			},
			wantCount: 2,
			validate: func(t *testing.T, results []ContainerOverride) {
				// Sorted alphabetically by query
				if results[0].ContainerQuery != "main (max-width: 400px)" {
					t.Errorf("results[0].ContainerQuery = %q, want %q", results[0].ContainerQuery, "main (max-width: 400px)")
				}
				if results[1].ContainerQuery != "main (max-width: 800px)" {
					t.Errorf("results[1].ContainerQuery = %q, want %q", results[1].ContainerQuery, "main (max-width: 800px)")
				}
			},
		},
		{
			name: "multiple components same query are grouped correctly",
			components: map[string]ComponentDefinition{
				"blogpost-layout": {
					Name:  "blogpost-layout",
					Class: "blogpost-layout",
					ContainerOverrides: map[string]map[string]any{
						"blogpost (max-width: 1024px)": {
							"display": "block",
						},
					},
				},
				"blogpost-sidebar": {
					Name:  "blogpost-sidebar",
					Class: "blogpost-sidebar",
					ContainerOverrides: map[string]map[string]any{
						"blogpost (max-width: 1024px)": {
							"display": "none",
						},
					},
				},
			},
			wantCount: 2,
			validate: func(t *testing.T, results []ContainerOverride) {
				// Both share the same query, sorted by component name
				if results[0].ComponentClass != "blogpost-layout" {
					t.Errorf("results[0].ComponentClass = %q, want %q", results[0].ComponentClass, "blogpost-layout")
				}
				if results[1].ComponentClass != "blogpost-sidebar" {
					t.Errorf("results[1].ComponentClass = %q, want %q", results[1].ComponentClass, "blogpost-sidebar")
				}
				if results[0].ContainerQuery != results[1].ContainerQuery {
					t.Error("expected same ContainerQuery for both results")
				}
			},
		},
		{
			name: "different queries produce distinct overrides",
			components: map[string]ComponentDefinition{
				"card-body": {
					Name:  "card-body",
					Class: "card-body",
					ContainerOverrides: map[string]map[string]any{
						"card (max-width: 300px)": {
							"padding": "8px",
						},
					},
				},
				"sidebar-nav": {
					Name:  "sidebar-nav",
					Class: "sidebar-nav",
					ContainerOverrides: map[string]map[string]any{
						"sidebar (max-width: 200px)": {
							"flex-direction": "column",
						},
					},
				},
			},
			wantCount: 2,
			validate: func(t *testing.T, results []ContainerOverride) {
				queries := map[string]bool{}
				for _, r := range results {
					queries[r.ContainerQuery] = true
				}
				if !queries["card (max-width: 300px)"] {
					t.Error("missing query 'card (max-width: 300px)'")
				}
				if !queries["sidebar (max-width: 200px)"] {
					t.Error("missing query 'sidebar (max-width: 200px)'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ExtractContainerOverrides(tt.components)

			if len(got) != tt.wantCount {
				t.Fatalf("ExtractContainerOverrides() returned %d overrides, want %d", len(got), tt.wantCount)
			}

			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}
