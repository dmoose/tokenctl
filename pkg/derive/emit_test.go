package derive

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCSSVarToTokenPath(t *testing.T) {
	t.Parallel()

	tests := []struct{ cssVar, want string }{
		{"--color-primary", "color.primary"},
		// A compound leaf stays one segment: color.primary.foreground
		// would be a different token from the one the engine names.
		{"--color-primary-foreground", "color.primary-foreground"},
		{"--color-success-subtle", "color.success-subtle"},
		{"--spacing-md", "spacing.md"},
		{"--radius-lg", "radius.lg"},
		{"--font-family-sans", "font.family.sans"},
		{"--font-size-2xl", "font.size.2xl"},
		{"--font-size-h1", "font.size.h1"},
		{"--font-weight-semibold", "font.weight.semibold"},
		{"--leading-relaxed", "leading.relaxed"},
		{"--tracking-wider", "tracking.wider"},
	}

	for _, tt := range tests {
		if got := CSSVarToTokenPath(tt.cssVar); got != tt.want {
			t.Errorf("CSSVarToTokenPath(%q) = %q, want %q", tt.cssVar, got, tt.want)
		}
	}
}

// Every derived token must survive the trip into the token document and
// back out as the same CSS variable name. A mapping that collapsed two
// variables onto one path would drop a token silently.
func TestToTokenJSON_RoundTripsEveryVariable(t *testing.T) {
	t.Parallel()

	for _, dark := range []bool{false, true} {
		p := DefaultParams
		p.IsDark = dark
		theme := Generate(p)

		data, err := theme.ToTokenJSON("semantic")
		if err != nil {
			t.Fatalf("ToTokenJSON: %v", err)
		}

		var doc map[string]any
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("emitted JSON does not parse: %v", err)
		}

		got := map[string]string{}
		collectTokens(t, doc, "", got)

		if len(got) != theme.Len() {
			t.Errorf("dark=%v: document holds %d tokens, theme has %d", dark, len(got), theme.Len())
		}
		for _, cssVar := range theme.Order {
			path := CSSVarToTokenPath(cssVar)
			value, ok := got[path]
			if !ok {
				t.Errorf("dark=%v: %s (path %s) missing from document", dark, cssVar, path)
				continue
			}
			if want := theme.Values[cssVar]; value != want {
				t.Errorf("dark=%v: %s = %q, want %q", dark, path, value, want)
			}
			// tokenctl renders path a.b.c as --a-b-c; that has to give
			// the variable name back.
			if rendered := "--" + strings.ReplaceAll(path, ".", "-"); rendered != cssVar {
				t.Errorf("path %s renders as %s, want %s", path, rendered, cssVar)
			}
		}
	}
}

func collectTokens(t *testing.T, node map[string]any, path string, out map[string]string) {
	t.Helper()
	if v, ok := node["$value"]; ok {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("token at %s has a non-string $value %T", path, v)
		}
		out[path] = s
		return
	}
	for k, v := range node {
		if strings.HasPrefix(k, "$") {
			continue
		}
		child, ok := v.(map[string]any)
		if !ok {
			continue
		}
		next := k
		if path != "" {
			next = path + "." + k
		}
		collectTokens(t, child, next, out)
	}
}

func TestToTokenJSON_TypesAndLayer(t *testing.T) {
	t.Parallel()

	data, err := Generate(DefaultParams).ToTokenJSON("semantic")
	if err != nil {
		t.Fatalf("ToTokenJSON: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse: %v", err)
	}

	if doc["$layer"] != "semantic" {
		t.Errorf("$layer = %v, want semantic", doc["$layer"])
	}
	if desc, _ := doc["$description"].(string); !strings.Contains(desc, "tokenctl derive") {
		t.Errorf("$description should record provenance, got %q", desc)
	}

	typeAt := func(group, leaf string) string {
		g, _ := doc[group].(map[string]any)
		l, _ := g[leaf].(map[string]any)
		s, _ := l["$type"].(string)
		return s
	}
	if got := typeAt("color", "primary"); got != "color" {
		t.Errorf("color.primary $type = %q, want color", got)
	}
	if got := typeAt("spacing", "md"); got != "dimension" {
		t.Errorf("spacing.md $type = %q, want dimension", got)
	}
	if got := typeAt("leading", "tight"); got != "number" {
		t.Errorf("leading.tight $type = %q, want number", got)
	}

	font, _ := doc["font"].(map[string]any)
	family, _ := font["family"].(map[string]any)
	sans, _ := family["sans"].(map[string]any)
	if sans["$type"] != "fontFamily" {
		t.Errorf("font.family.sans $type = %v, want fontFamily", sans["$type"])
	}

	// Omitting the layer must leave the key out entirely rather than
	// writing an empty one, which tokenctl would read as a real layer.
	data, err = Generate(DefaultParams).ToTokenJSON("")
	if err != nil {
		t.Fatalf("ToTokenJSON: %v", err)
	}
	// A fresh map: unmarshalling into a populated one merges, which
	// would carry the previous document's $layer over.
	var unlayered map[string]any
	if err := json.Unmarshal(data, &unlayered); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, present := unlayered["$layer"]; present {
		t.Error("empty layer should omit $layer, not write it empty")
	}
}

func TestToCSS(t *testing.T) {
	t.Parallel()

	theme := Generate(DefaultParams)
	css := theme.ToCSS("")
	if !strings.Contains(css, ":root {") {
		t.Error("empty selector should default to :root")
	}
	for _, cssVar := range theme.Order {
		if !strings.Contains(css, "  "+cssVar+": "+theme.Values[cssVar]+";\n") {
			t.Errorf("CSS is missing a declaration for %s", cssVar)
		}
	}

	dark := Generate(Params{
		Hue: 250, Chroma: 0.2, IsDark: true, Tint: 30,
		Saturation: 100, FontPairing: "system", Density: 100,
	})
	if css := dark.ToCSS(`[data-theme="dark"]`); !strings.Contains(css, `[data-theme="dark"] {`) {
		t.Error("explicit selector was not used")
	}
}

func TestGenerate_IsDeterministic(t *testing.T) {
	t.Parallel()

	first := Generate(DefaultParams)
	for range 20 {
		next := Generate(DefaultParams)
		if len(next.Order) != len(first.Order) {
			t.Fatalf("token count varies: %d vs %d", len(next.Order), len(first.Order))
		}
		for i := range first.Order {
			if next.Order[i] != first.Order[i] {
				t.Fatalf("emission order varies at %d: %q vs %q", i, next.Order[i], first.Order[i])
			}
		}
	}
}

func TestParamsValidate(t *testing.T) {
	t.Parallel()

	// The engine clamped out-of-range controls silently. tokenctl
	// refuses, so a caller never learns a range that does not exist.
	bad := []struct {
		name  string
		mutex func(*Params)
	}{
		{"tint above range", func(p *Params) { p.Tint = 101 }},
		{"tint below range", func(p *Params) { p.Tint = -1 }},
		{"saturation above range", func(p *Params) { p.Saturation = 151 }},
		{"density below range", func(p *Params) { p.Density = 74 }},
		{"density above range", func(p *Params) { p.Density = 131 }},
		{"hue above range", func(p *Params) { p.Hue = 361 }},
		{"negative chroma", func(p *Params) { p.Chroma = -0.1 }},
		{"unknown pairing", func(p *Params) { p.FontPairing = "nope" }},
	}
	for _, tt := range bad {
		p := DefaultParams
		tt.mutex(&p)
		if err := p.Validate(); err == nil {
			t.Errorf("%s: expected an error", tt.name)
		}
	}

	if err := DefaultParams.Validate(); err != nil {
		t.Errorf("defaults must validate: %v", err)
	}
	// Bounds themselves are inclusive.
	for _, p := range []Params{
		{Hue: 0, Chroma: 0, Tint: 0, Saturation: 0, Density: 75, FontPairing: "system"},
		{Hue: 360, Chroma: 0.4, Tint: 100, Saturation: 150, Density: 130, FontPairing: "retro"},
	} {
		if err := p.Validate(); err != nil {
			t.Errorf("bound value rejected: %v", err)
		}
	}
}

func TestPresetsAndSystems(t *testing.T) {
	t.Parallel()

	if len(Presets) != 8 {
		t.Errorf("want the engine's 8 presets, got %d", len(Presets))
	}
	if len(TypographySystems) != 10 {
		t.Errorf("want the engine's 10 typography systems, got %d", len(TypographySystems))
	}
	if _, ok := PresetByName("BLUE"); !ok {
		t.Error("preset lookup should be case-insensitive")
	}
	if _, ok := PresetByName("chartreuse"); ok {
		t.Error("unknown preset should not resolve")
	}
	// An unrecognised pairing falls back to the first system, matching
	// the engine, rather than emitting no font tokens at all.
	p := DefaultParams
	p.FontPairing = "does-not-exist"
	theme := Generate(p)
	if got, _ := theme.Get("--font-family-sans"); got != TypographySystems[0].Sans {
		t.Errorf("unknown pairing gave %q, want the first system's sans", got)
	}
}
