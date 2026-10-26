package build

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

// A semantic token that references another semantic token is written
// as a var() chain, so overriding the root at runtime moves it; a
// reference to the brand layer stays literal.
func TestSemanticReferencesEmitVarChains(t *testing.T) {
	tree := fstest.MapFS{
		"brand/colors.json": {Data: []byte(`{"$layer": "brand", "brand": {"blue": {"$value": "#00f"}, "white": {"$value": "#fff"}}}`)},
		"semantic/colors.json": {Data: []byte(`{"$layer": "semantic", "color": {
			"primary": {"$value": "{brand.blue}"},
			"link": {"$value": "{color.primary}"},
			"primary-hover": {"$value": "color-mix(in oklch, {color.primary} 85%, black)"},
			"on-primary": {"$value": "{brand.white}"}
		}}`)},
		"themes/dark.json": {Data: []byte(`{"color": {"link": {"$value": "{brand.white}"}}}`)},
	}
	res, err := BuildCSS([]fs.FS{tree}, Options{Format: FormatCSS})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"--color-primary: #00f;",
		"--color-link: var(--color-primary);",
		"--color-primary-hover: color-mix(in oklch, var(--color-primary) 85%, black);",
		"--color-on-primary: #fff;",
	} {
		if !strings.Contains(res.CSS, want) {
			t.Errorf("missing %q in:\n%s", want, res.CSS)
		}
	}
	// The dark theme repoints link to a brand literal; its block must
	// say so even though nothing else about link differs.
	dark := res.CSS[strings.Index(res.CSS, `[data-theme="dark"]`):]
	if !strings.Contains(dark, "--color-link: #fff;") {
		t.Errorf("dark theme lost its link override:\n%s", dark)
	}
}
