package tokens

import (
	"fmt"
	"strings"
)

// EmitAll returns the value each token is written with as a custom
// property, alongside the resolved values ResolveAll computes.
//
// A semantic token that references another semantic token keeps the
// reference as var(--that-token), so the chain survives into the CSS:
// a site or theme that sets --color-primary moves --color-link,
// --color-primary-hover and everything else derived from it, instead
// of each having been flattened to the library's primary at build
// time. References to anything else (brand scales, unlayered tokens)
// stay literal, as do expressions, which tokenctl evaluates itself and
// so need literal operands.
//
// layers comes from Dictionary.Layers; resolved from ResolveAll over the
// same dictionary.
func (r *Resolver) EmitAll(layers map[string]Layer, resolved map[string]any) map[string]any {
	emitted := make(map[string]any, len(resolved))
	for path, val := range resolved {
		emitted[path] = val
		if layers[path] != LayerSemantic {
			continue
		}
		raw, ok := r.flatTokens[path].(string)
		if !ok || IsExpression(raw) || !strings.Contains(raw, "{") {
			continue
		}
		chained := false
		out := refRegex.ReplaceAllStringFunc(raw, func(match string) string {
			ref := match[1 : len(match)-1]
			if layers[ref] == LayerSemantic {
				if _, defined := resolved[ref]; defined {
					chained = true
					return "var(--" + strings.ReplaceAll(ref, ".", "-") + ")"
				}
			}
			return fmt.Sprintf("%v", resolved[ref])
		})
		if chained {
			emitted[path] = out
		}
	}
	return emitted
}

// DiffEmitted is Diff extended to emitted values: a theme token is
// written when its resolved value or its emitted form differs from
// base. The second case is a theme that repoints a reference to a
// token that happens to resolve the same; without it the theme would
// keep following the base reference.
func DiffEmitted(targetResolved, baseResolved, targetEmitted, baseEmitted map[string]any) map[string]any {
	diff := Diff(targetResolved, baseResolved)
	for key, val := range targetEmitted {
		if _, ok := diff[key]; ok {
			continue
		}
		if fmt.Sprintf("%v", val) != fmt.Sprintf("%v", baseEmitted[key]) {
			diff[key] = targetResolved[key]
		}
	}
	return diff
}
