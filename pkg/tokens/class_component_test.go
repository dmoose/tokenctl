package tokens

import "testing"

// A $responsive property inside a class component is CSS for that class.
// It used to also flatten into a :root custom property, with its token
// references left unresolved, and into the token-level @media block.
func TestClassComponentPropertiesAreNotTokens(t *testing.T) {
	d := &Dictionary{Root: map[string]any{
		"spacing": map[string]any{"xl": map[string]any{"$value": "2rem"}},
		"components": map[string]any{
			"docs-nav": map[string]any{
				"$type":  "component",
				"$class": "docs-nav",
				"base": map[string]any{
					"top": map[string]any{"$value": "auto", "$responsive": map[string]any{"lg": "{spacing.xl}"}},
				},
			},
			// No $class: a group of component tokens, which are published.
			"button": map[string]any{
				"$type":   "component",
				"primary": map[string]any{"color": map[string]any{"$value": "#fff", "$responsive": map[string]any{"md": "#000"}}},
			},
		},
	}}

	flat := map[string]any{}
	if err := flatten(d.Root, "", flat); err != nil {
		t.Fatal(err)
	}
	if _, leaked := flat["components.docs-nav.base.top"]; leaked {
		t.Error("a class component's property flattened into a token")
	}
	if _, ok := flat["components.button.primary.color"]; !ok {
		t.Error("a component token group stopped flattening")
	}
	if _, ok := flat["spacing.xl"]; !ok {
		t.Error("an ordinary token stopped flattening")
	}

	var paths []string
	for _, rt := range ExtractResponsiveTokens(d) {
		paths = append(paths, rt.Path)
	}
	if len(paths) != 1 || paths[0] != "components.button.primary.color" {
		t.Errorf("responsive tokens = %q, want only the component token group's", paths)
	}
}
