package generators

import (
	"strings"
	"testing"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

// An @media block in a component's base used to be read as a property
// and dropped without a word. It must fail the build instead, naming
// the component and the key, in both generators.
func TestComponentObjectValueRejected(t *testing.T) {
	comps := map[string]tokens.ComponentDefinition{
		"docs-shell": {Class: "docs-shell", Base: map[string]any{
			"display": "grid",
			"@media (min-width: 64rem)": map[string]any{
				"grid-template-columns": "16rem 1fr",
			},
		}},
	}
	_, cssErr := NewCSSGenerator().generateComponents(comps, tokens.DefaultBreakpoints)
	_, twErr := NewTailwindGenerator().generateComponents(comps)
	for name, err := range map[string]error{"css": cssErr, "tailwind": twErr} {
		if err == nil {
			t.Errorf("%s: an @media block in base was accepted", name)
			continue
		}
		for _, want := range []string{"docs-shell", "@media (min-width: 64rem)", "$responsive"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: error does not mention %s: %v", name, want, err)
			}
		}
	}
}

// The same check reaches nested selectors, and leaves the supported
// object shape ({"$value", "$responsive"}) alone.
func TestComponentObjectValueScope(t *testing.T) {
	ok := map[string]tokens.ComponentDefinition{
		"grid": {Class: "grid", Base: map[string]any{
			"grid-template-columns": map[string]any{"$value": "1fr", "$responsive": map[string]any{"lg": "1fr 1fr"}},
			"&:hover":               map[string]any{"color": "red"},
		}},
	}
	if _, err := NewCSSGenerator().generateComponents(ok, tokens.DefaultBreakpoints); err != nil {
		t.Errorf("supported shapes rejected: %v", err)
	}
	nested := map[string]tokens.ComponentDefinition{
		"grid": {Class: "grid", Base: map[string]any{
			"& > .card": map[string]any{"@media (max-width: 640px)": map[string]any{"padding": "0"}},
		}},
	}
	if _, err := NewCSSGenerator().generateComponents(nested, tokens.DefaultBreakpoints); err == nil || !strings.Contains(err.Error(), "& > .card") {
		t.Errorf("an object inside a nested selector was not reported with its selector: %v", err)
	}
}
