package tokens

import (
	"strings"
	"testing"
)

func TestExtractBreakpoints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		root map[string]any
		want map[string]string
	}{
		{
			name: "custom breakpoints",
			root: map[string]any{
				"$breakpoints": map[string]any{
					"tablet": "600px",
					"desktop": "900px",
				},
			},
			want: map[string]string{
				"tablet":  "600px",
				"desktop": "900px",
			},
		},
		{
			name: "falls back to defaults when no breakpoints defined",
			root: map[string]any{},
			want: map[string]string{
				"sm": "640px",
				"md": "768px",
				"lg": "1024px",
				"xl": "1280px",
			},
		},
		{
			name: "falls back to defaults when breakpoints key is wrong type",
			root: map[string]any{
				"$breakpoints": "not a map",
			},
			want: map[string]string{
				"sm": "640px",
				"md": "768px",
				"lg": "1024px",
				"xl": "1280px",
			},
		},
		{
			name: "skips non-string values in breakpoints",
			root: map[string]any{
				"$breakpoints": map[string]any{
					"sm":      "640px",
					"invalid": 768,
					"also":    true,
				},
			},
			want: map[string]string{
				"sm": "640px",
			},
		},
		{
			name: "falls back to defaults when all values are non-string",
			root: map[string]any{
				"$breakpoints": map[string]any{
					"sm": 640,
					"md": 768,
				},
			},
			want: map[string]string{
				"sm": "640px",
				"md": "768px",
				"lg": "1024px",
				"xl": "1280px",
			},
		},
		{
			name: "single breakpoint",
			root: map[string]any{
				"$breakpoints": map[string]any{
					"wide": "1400px",
				},
			},
			want: map[string]string{
				"wide": "1400px",
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
			got := ExtractBreakpoints(d)

			if len(got) != len(tt.want) {
				t.Fatalf("ExtractBreakpoints() returned %d entries, want %d", len(got), len(tt.want))
			}

			for k, wantVal := range tt.want {
				gotVal, ok := got[k]
				if !ok {
					t.Errorf("ExtractBreakpoints() missing key %q", k)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("ExtractBreakpoints()[%q] = %q, want %q", k, gotVal, wantVal)
				}
			}
		})
	}
}

func TestExtractResponsiveTokens(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		root        map[string]any
		sourceFiles map[string]string
		wantCount   int
		validate    func(t *testing.T, results []ResponsiveToken)
	}{
		{
			name: "token with responsive overrides",
			root: map[string]any{
				"spacing": map[string]any{
					"gap": map[string]any{
						"$value": "16px",
						"$type":  "dimension",
						"$responsive": map[string]any{
							"md": "24px",
							"lg": "32px",
						},
					},
				},
			},
			sourceFiles: map[string]string{
				"spacing.gap": "spacing.json",
			},
			wantCount: 1,
			validate: func(t *testing.T, results []ResponsiveToken) {
				rt := results[0]
				if rt.Path != "spacing.gap" {
					t.Errorf("Path = %q, want %q", rt.Path, "spacing.gap")
				}
				if rt.BaseValue != "16px" {
					t.Errorf("BaseValue = %v, want %q", rt.BaseValue, "16px")
				}
				if rt.Type != "dimension" {
					t.Errorf("Type = %q, want %q", rt.Type, "dimension")
				}
				if rt.SourceFile != "spacing.json" {
					t.Errorf("SourceFile = %q, want %q", rt.SourceFile, "spacing.json")
				}
				if len(rt.Overrides) != 2 {
					t.Fatalf("Overrides count = %d, want 2", len(rt.Overrides))
				}
				if rt.Overrides["md"] != "24px" {
					t.Errorf("Overrides[md] = %v, want %q", rt.Overrides["md"], "24px")
				}
				if rt.Overrides["lg"] != "32px" {
					t.Errorf("Overrides[lg] = %v, want %q", rt.Overrides["lg"], "32px")
				}
			},
		},
		{
			name: "token without responsive field is excluded",
			root: map[string]any{
				"color": map[string]any{
					"primary": map[string]any{
						"$value": "#ff0000",
						"$type":  "color",
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "responsive field with wrong type is excluded",
			root: map[string]any{
				"font": map[string]any{
					"size": map[string]any{
						"$value":      "14px",
						"$responsive": "not a map",
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "nested groups with inherited type",
			root: map[string]any{
				"typography": map[string]any{
					"$type": "dimension",
					"heading": map[string]any{
						"size": map[string]any{
							"$value": "24px",
							"$responsive": map[string]any{
								"lg": "32px",
							},
						},
					},
				},
			},
			wantCount: 1,
			validate: func(t *testing.T, results []ResponsiveToken) {
				rt := results[0]
				if rt.Path != "typography.heading.size" {
					t.Errorf("Path = %q, want %q", rt.Path, "typography.heading.size")
				}
				if rt.Type != "dimension" {
					t.Errorf("Type = %q, want %q (inherited from parent)", rt.Type, "dimension")
				}
			},
		},
		{
			name: "child type overrides inherited type",
			root: map[string]any{
				"group": map[string]any{
					"$type": "color",
					"item": map[string]any{
						"$value": "2rem",
						"$type":  "dimension",
						"$responsive": map[string]any{
							"sm": "3rem",
						},
					},
				},
			},
			wantCount: 1,
			validate: func(t *testing.T, results []ResponsiveToken) {
				rt := results[0]
				if rt.Type != "dimension" {
					t.Errorf("Type = %q, want %q (child overrides parent)", rt.Type, "dimension")
				}
			},
		},
		{
			name: "multiple responsive tokens across groups",
			root: map[string]any{
				"spacing": map[string]any{
					"sm": map[string]any{
						"$value": "8px",
						"$responsive": map[string]any{
							"md": "12px",
						},
					},
					"lg": map[string]any{
						"$value": "32px",
						"$responsive": map[string]any{
							"md": "48px",
						},
					},
				},
			},
			wantCount: 2,
		},
		{
			name:      "empty dictionary",
			root:      map[string]any{},
			wantCount: 0,
		},
		{
			name: "dollar-prefixed keys at group level are skipped",
			root: map[string]any{
				"$meta": map[string]any{
					"$value": "should-not-appear",
					"$responsive": map[string]any{
						"sm": "nope",
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "token without source file tracking",
			root: map[string]any{
				"size": map[string]any{
					"$value": "10px",
					"$responsive": map[string]any{
						"xl": "20px",
					},
				},
			},
			wantCount: 1,
			validate: func(t *testing.T, results []ResponsiveToken) {
				if results[0].SourceFile != "" {
					t.Errorf("SourceFile = %q, want empty string", results[0].SourceFile)
				}
			},
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

			got := ExtractResponsiveTokens(d)

			if len(got) != tt.wantCount {
				t.Fatalf("ExtractResponsiveTokens() returned %d tokens, want %d", len(got), tt.wantCount)
			}

			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestGenerateResponsiveCSS(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		breakpoints map[string]string
		tokens      []ResponsiveToken
		wantParts   []string
		wantEmpty   bool
	}{
		{
			name:        "empty tokens returns empty string",
			breakpoints: map[string]string{"sm": "640px"},
			tokens:      nil,
			wantEmpty:   true,
		},
		{
			name:        "empty token slice returns empty string",
			breakpoints: map[string]string{"sm": "640px"},
			tokens:      []ResponsiveToken{},
			wantEmpty:   true,
		},
		{
			name: "single breakpoint single token",
			breakpoints: map[string]string{
				"md": "768px",
			},
			tokens: []ResponsiveToken{
				{
					Path:      "spacing.gap",
					BaseValue: "16px",
					Overrides: map[string]any{
						"md": "24px",
					},
				},
			},
			wantParts: []string{
				"@media (min-width: 768px)",
				":root",
				"--spacing-gap: 24px;",
			},
		},
		{
			name: "multiple breakpoints sorted by size",
			breakpoints: map[string]string{
				"lg": "1024px",
				"sm": "640px",
				"md": "768px",
			},
			tokens: []ResponsiveToken{
				{
					Path:      "font.size",
					BaseValue: "14px",
					Overrides: map[string]any{
						"sm": "16px",
						"md": "18px",
						"lg": "20px",
					},
				},
			},
			wantParts: []string{
				"@media (min-width: 640px)",
				"--font-size: 16px;",
				"@media (min-width: 768px)",
				"--font-size: 18px;",
				"@media (min-width: 1024px)",
				"--font-size: 20px;",
			},
		},
		{
			name: "breakpoint order is ascending by pixel value",
			breakpoints: map[string]string{
				"xl": "1280px",
				"sm": "640px",
			},
			tokens: []ResponsiveToken{
				{
					Path:      "a",
					BaseValue: "1",
					Overrides: map[string]any{
						"sm": "2",
						"xl": "3",
					},
				},
			},
		},
		{
			name: "multiple tokens in same breakpoint sorted by path",
			breakpoints: map[string]string{
				"md": "768px",
			},
			tokens: []ResponsiveToken{
				{
					Path:      "z.token",
					BaseValue: "1",
					Overrides: map[string]any{"md": "2"},
				},
				{
					Path:      "a.token",
					BaseValue: "3",
					Overrides: map[string]any{"md": "4"},
				},
			},
			wantParts: []string{
				"--a-token: 4;",
				"--z-token: 2;",
			},
		},
		{
			name: "token with override for undefined breakpoint is skipped",
			breakpoints: map[string]string{
				"md": "768px",
			},
			tokens: []ResponsiveToken{
				{
					Path:      "spacing.pad",
					BaseValue: "8px",
					Overrides: map[string]any{
						"xxl": "64px",
					},
				},
			},
			wantEmpty: true,
		},
		{
			name: "dots in path become dashes in css variable",
			breakpoints: map[string]string{
				"sm": "640px",
			},
			tokens: []ResponsiveToken{
				{
					Path:      "a.b.c.d",
					BaseValue: "1",
					Overrides: map[string]any{"sm": "2"},
				},
			},
			wantParts: []string{
				"--a-b-c-d: 2;",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GenerateResponsiveCSS(tt.breakpoints, tt.tokens)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("GenerateResponsiveCSS() = %q, want empty string", got)
				}
				return
			}

			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("GenerateResponsiveCSS() missing %q in output:\n%s", part, got)
				}
			}
		})
	}
}

func TestGenerateResponsiveCSS_BreakpointOrdering(t *testing.T) {
	t.Parallel()
	breakpoints := map[string]string{
		"sm": "640px",
		"md": "768px",
		"lg": "1024px",
		"xl": "1280px",
	}

	tokens := []ResponsiveToken{
		{
			Path:      "size",
			BaseValue: "10px",
			Overrides: map[string]any{
				"sm": "12px",
				"md": "14px",
				"lg": "16px",
				"xl": "18px",
			},
		},
	}

	got := GenerateResponsiveCSS(breakpoints, tokens)

	smIdx := strings.Index(got, "640px")
	mdIdx := strings.Index(got, "768px")
	lgIdx := strings.Index(got, "1024px")
	xlIdx := strings.Index(got, "1280px")

	if smIdx == -1 || mdIdx == -1 || lgIdx == -1 || xlIdx == -1 {
		t.Fatalf("Missing breakpoint in output:\n%s", got)
	}

	if smIdx >= mdIdx || mdIdx >= lgIdx || lgIdx >= xlIdx {
		t.Errorf("Breakpoints not in ascending order (sm=%d, md=%d, lg=%d, xl=%d) in output:\n%s",
			smIdx, mdIdx, lgIdx, xlIdx, got)
	}
}

func TestSortBreakpointsBySize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		breakpoints map[string]string
		want        []string
	}{
		{
			name: "default breakpoints",
			breakpoints: map[string]string{
				"xl": "1280px",
				"sm": "640px",
				"lg": "1024px",
				"md": "768px",
			},
			want: []string{"sm", "md", "lg", "xl"},
		},
		{
			name: "single breakpoint",
			breakpoints: map[string]string{
				"only": "500px",
			},
			want: []string{"only"},
		},
		{
			name: "custom sizes",
			breakpoints: map[string]string{
				"large":  "1200px",
				"small":  "320px",
				"medium": "800px",
			},
			want: []string{"small", "medium", "large"},
		},
		{
			name:        "empty map",
			breakpoints: map[string]string{},
			want:        []string{},
		},
		{
			name: "non-parseable value treated as zero",
			breakpoints: map[string]string{
				"a": "notpx",
				"b": "100px",
			},
			want: []string{"a", "b"},
		},
		{
			name: "close values preserve correct order",
			breakpoints: map[string]string{
				"a": "1px",
				"b": "2px",
				"c": "3px",
			},
			want: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := sortBreakpointsBySize(tt.breakpoints)

			if len(got) != len(tt.want) {
				t.Fatalf("sortBreakpointsBySize() returned %d entries, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("sortBreakpointsBySize()[%d] = %q, want %q (full result: %v)", i, got[i], tt.want[i], got)
					break
				}
			}
		})
	}
}
