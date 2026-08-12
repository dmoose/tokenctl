// tokenctl/pkg/generators/catalog_shape_test.go
//
// The catalog is a consumer contract: fastatic-studio's token browser
// renders from it, and anything else pointed at `build --format=catalog`
// reads it as the description of what the stylesheet contains. These
// tests pin the two properties that failure made expensive — that a
// definition carries its styling, and that the same input produces the
// same bytes.
package generators

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

// fixture mirrors the shapes a real component file produces: a base with
// a nested pseudo-selector mixed into its properties, variants with their
// own states, sizes, and a component-level state class.
func shapeFixture() map[string]tokens.ComponentDefinition {
	return map[string]tokens.ComponentDefinition{
		"input": {
			Class: "input",
			Base: map[string]any{
				"display": "block",
				"border":  "1px solid var(--color-border)",
				"&:focus": map[string]any{"outline": "2px solid var(--color-ring)"},
			},
			Variants: map[string]tokens.VariantDef{
				"ghost": {
					Class:      "input-ghost",
					Properties: map[string]any{"border": "none"},
					States: map[string]tokens.State{
						"&:hover": {Properties: map[string]any{"background": "var(--color-muted)"}},
					},
				},
				"outline": {Class: "input-outline", Properties: map[string]any{"border-width": "2px"}},
			},
			Sizes: map[string]tokens.VariantDef{
				"sm": {Class: "input-sm", Properties: map[string]any{"padding": "0.25rem"}},
				"lg": {Class: "input-lg", Properties: map[string]any{"padding": "0.75rem"}},
			},
			States: map[string]tokens.VariantDef{
				"error": {
					Class:      "input-error",
					Properties: map[string]any{"border-color": "var(--color-error)"},
					States: map[string]tokens.State{
						"&:focus": {Properties: map[string]any{"outline-color": "var(--color-error)"}},
					},
				},
			},
		},
	}
}

func generateShapeCatalog(t *testing.T) map[string]any {
	t.Helper()
	out, err := NewCatalogGenerator().Generate(
		map[string]any{"color.primary": "#3b82f6"},
		shapeFixture(),
		nil,
	)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("catalog is not valid JSON: %v", err)
	}
	return doc
}

func definitionOf(t *testing.T, doc map[string]any, class string) map[string]any {
	t.Helper()
	comps, _ := doc["components"].(map[string]any)
	input, _ := comps["input"].(map[string]any)
	defs, _ := input["definitions"].(map[string]any)
	def, ok := defs[class].(map[string]any)
	if !ok {
		t.Fatalf("no definition for %q; definitions present: %v", class, keysOf(defs))
	}
	return def
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TestCatalogDefinitionsCarryProperties is the S2 blocker in test form.
// Every definition used to serialize to {"$class": …} because
// VariantDef's Properties and States were json:"-" with no MarshalJSON,
// so a consumer reading the catalog to find out what a class does got
// only its name back.
func TestCatalogDefinitionsCarryProperties(t *testing.T) {
	t.Parallel()
	doc := generateShapeCatalog(t)

	variant := definitionOf(t, doc, "input-ghost")
	props, ok := variant["properties"].(map[string]any)
	if !ok {
		t.Fatalf("input-ghost carries no properties: %v", variant)
	}
	if props["border"] != "none" {
		t.Errorf("input-ghost border = %v, want none", props["border"])
	}
	states, ok := variant["states"].(map[string]any)
	if !ok {
		t.Fatalf("input-ghost carries no states: %v", variant)
	}
	hover, ok := states["&:hover"].(map[string]any)
	if !ok {
		t.Fatalf("input-ghost has no &:hover state: %v", states)
	}
	if hover["background"] != "var(--color-muted)" {
		t.Errorf("&:hover background = %v", hover["background"])
	}
}

// TestCatalogBaseSplitsNestedSelectors: a component's base map is
// authored flat, pseudo-selector blocks mixed in with real properties.
// Handing "&:focus" over as if it were a CSS property described a
// stylesheet that does not exist.
func TestCatalogBaseSplitsNestedSelectors(t *testing.T) {
	t.Parallel()
	base := definitionOf(t, generateShapeCatalog(t), "input")

	props, _ := base["properties"].(map[string]any)
	if _, leaked := props["&:focus"]; leaked {
		t.Error("base properties still carry the &:focus block as a property")
	}
	if props["display"] != "block" {
		t.Errorf("base display = %v, want block", props["display"])
	}
	states, ok := base["states"].(map[string]any)
	if !ok {
		t.Fatalf("base carries no states: %v", base)
	}
	focus, ok := states["&:focus"].(map[string]any)
	if !ok {
		t.Fatalf("base has no &:focus state: %v", states)
	}
	if focus["outline"] != "2px solid var(--color-ring)" {
		t.Errorf("&:focus outline = %v", focus["outline"])
	}
}

// TestCatalogCarriesComponentStates: the CSS generator emits a class for
// every entry in a component's `states` map. The catalog listed none of
// them, so a consumer asking what `.input-error` is was told no such
// class exists while the stylesheet in front of it defined one.
func TestCatalogCarriesComponentStates(t *testing.T) {
	t.Parallel()
	doc := generateShapeCatalog(t)

	comps, _ := doc["components"].(map[string]any)
	input, _ := comps["input"].(map[string]any)
	var classes []string
	for _, c := range input["classes"].([]any) {
		classes = append(classes, c.(string))
	}
	if !slices.Contains(classes, "input-error") {
		t.Errorf("classes omit input-error: %v", classes)
	}

	def := definitionOf(t, doc, "input-error")
	props, _ := def["properties"].(map[string]any)
	if props["border-color"] != "var(--color-error)" {
		t.Errorf("input-error border-color = %v", props["border-color"])
	}
	states, _ := def["states"].(map[string]any)
	if _, ok := states["&:focus"]; !ok {
		t.Errorf("input-error state's own &:focus is missing: %v", def)
	}
}

// TestCatalogOmitsEmptyPropertyBlocks keeps the shape honest: a class
// with no states should not claim an empty set of them.
func TestCatalogOmitsEmptyPropertyBlocks(t *testing.T) {
	t.Parallel()
	def := definitionOf(t, generateShapeCatalog(t), "input-sm")
	if _, ok := def["states"]; ok {
		t.Errorf("input-sm has no states but the catalog wrote a states key: %v", def)
	}
	if _, ok := def["properties"]; !ok {
		t.Errorf("input-sm lost its properties: %v", def)
	}
}

// TestCatalogIsByteIdentical is the determinism gate. Two things used to
// break it: meta.generated_at was wall-clock, and the class list was
// built by ranging Go maps, whose iteration order is randomized.
func TestCatalogIsByteIdentical(t *testing.T) {
	t.Parallel()

	resolved := map[string]any{
		"color.primary":   "#3b82f6",
		"color.secondary": "#8b5cf6",
		"spacing.sm":      "0.5rem",
	}
	themes := map[string]CatalogThemeInput{
		"dark": {
			Description:    "Dark",
			ResolvedTokens: map[string]any{"color.primary": "#60a5fa"},
			DiffTokens:     map[string]any{"color.primary": "#60a5fa"},
		},
	}

	first, err := NewCatalogGenerator().Generate(resolved, shapeFixture(), themes)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Enough runs that a randomized map order would have to be lucky
	// many times over to slip through.
	for i := range 20 {
		again, err := NewCatalogGenerator().Generate(resolved, shapeFixture(), themes)
		if err != nil {
			t.Fatalf("Generate (run %d): %v", i, err)
		}
		if again != first {
			t.Fatalf("run %d differs from run 0:\n%s", i, firstDiffLine(first, again))
		}
	}
	if strings.Contains(first, "generated_at") {
		t.Error("default catalog carries a generated_at stamp, which cannot be reproduced")
	}
}

// TestCatalogGeneratedAtIsInjectable: a caller that genuinely wants a
// stamp can have one, and it is exactly what they passed.
func TestCatalogGeneratedAtIsInjectable(t *testing.T) {
	t.Parallel()
	gen := NewCatalogGeneratorWithOptions(CatalogOptions{GeneratedAt: "2026-08-10T00:00:00Z"})
	out, err := gen.Generate(map[string]any{"color.primary": "#3b82f6"}, nil, nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	var doc CatalogSchema
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if doc.Meta.GeneratedAt != "2026-08-10T00:00:00Z" {
		t.Errorf("generated_at = %q, want the injected value", doc.Meta.GeneratedAt)
	}
}

// TestCatalogVersionTracksTheBinary: the stamped version used to be a
// hand-maintained const that read "1.2.0" whatever the binary was.
func TestCatalogVersionTracksTheBinary(t *testing.T) {
	t.Parallel()
	out, err := NewCatalogGenerator().Generate(map[string]any{"color.primary": "#3b82f6"}, nil, nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	var doc CatalogSchema
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if doc.Meta.TokenctlVersion == "" {
		t.Fatal("tokenctl_version is empty")
	}
	if doc.Meta.TokenctlVersion == "1.2.0" {
		t.Error("tokenctl_version is still the retired hardcoded const")
	}
}

func firstDiffLine(a, b string) string {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := range min(len(al), len(bl)) {
		if al[i] != bl[i] {
			return "line " + strconv.Itoa(i+1) + ":\n  A: " + al[i] + "\n  B: " + bl[i]
		}
	}
	return "lengths differ"
}
