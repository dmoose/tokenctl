// tokenctl/pkg/derive/emit.go
//
// Rendering a derived theme into the shapes tokenctl already speaks:
// a W3C Design Tokens JSON document that `tokenctl build` can consume,
// or a plain CSS custom-property block.
package derive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// CSSVarToTokenPath converts an emitted CSS variable name to the token
// path that would produce it. tokenctl renders a token at path a.b.c as
// --a-b-c, so the mapping has to be chosen deliberately: only the known
// leading groups are split, and whatever remains stays a single leaf.
// Splitting on every dash instead would turn --color-primary-foreground
// into color.primary.foreground, which is a different token.
func CSSVarToTokenPath(cssVar string) string {
	name := strings.TrimPrefix(cssVar, "--")

	// Two-segment groups first so font-family beats font.
	twoPart := []string{"font-family", "font-size", "font-weight"}
	for _, group := range twoPart {
		if rest, ok := strings.CutPrefix(name, group+"-"); ok {
			return strings.ReplaceAll(group, "-", ".") + "." + rest
		}
	}

	onePart := []string{"color", "spacing", "radius", "leading", "tracking"}
	for _, group := range onePart {
		if rest, ok := strings.CutPrefix(name, group+"-"); ok {
			return group + "." + rest
		}
	}

	return name
}

// tokenTypeFor picks the W3C $type for a derived token.
func tokenTypeFor(cssVar, value string) string {
	switch {
	case strings.HasPrefix(cssVar, "--color-"):
		return "color"
	case strings.HasPrefix(cssVar, "--font-family-"):
		return "fontFamily"
	case strings.HasPrefix(cssVar, "--font-weight-"):
		return "fontWeight"
	case strings.HasPrefix(cssVar, "--leading-"):
		// Line height is a unitless multiplier.
		return "number"
	case strings.HasSuffix(value, "rem") || strings.HasSuffix(value, "em"):
		return "dimension"
	default:
		return ""
	}
}

// ToTokenJSON renders the theme as a W3C Design Tokens document.
// layer, when non-empty, is written as the document's $layer so the
// output drops straight into a layered token tree.
func (t *Theme) ToTokenJSON(layer string) ([]byte, error) {
	root := map[string]any{}
	if layer != "" {
		root["$layer"] = layer
	}
	root["$description"] = t.provenance()

	for _, cssVar := range t.Order {
		value := t.Values[cssVar]
		path := CSSVarToTokenPath(cssVar)

		token := map[string]any{"$value": value}
		if typ := tokenTypeFor(cssVar, value); typ != "" {
			token["$type"] = typ
		}

		if err := insertPath(root, strings.Split(path, "."), token); err != nil {
			return nil, fmt.Errorf("%s: %w", cssVar, err)
		}
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(root); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ToCSS renders the theme as a custom-property block under selector.
// Variable names are the engine's, unchanged.
func (t *Theme) ToCSS(selector string) string {
	if selector == "" {
		selector = ":root"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "/* %s */\n", t.provenance())
	fmt.Fprintf(&sb, "%s {\n", selector)
	for _, cssVar := range t.Order {
		fmt.Fprintf(&sb, "  %s: %s;\n", cssVar, t.Values[cssVar])
	}
	sb.WriteString("}\n")
	return sb.String()
}

// provenance records the controls a theme was derived from, so a
// generated file can be regenerated without guessing its inputs.
func (t *Theme) provenance() string {
	mode := "light"
	if t.Params.IsDark {
		mode = "dark"
	}
	return fmt.Sprintf(
		"Derived by tokenctl derive — hue %s, chroma %s, %s, tint %s, saturation %s, type %s, density %s",
		numberToString(t.Params.Hue), numberToString(t.Params.Chroma), mode,
		numberToString(t.Params.Tint), numberToString(t.Params.Saturation),
		t.Params.FontPairing, numberToString(t.Params.Density),
	)
}

// insertPath places token at segments within root, creating groups as
// needed. A collision means two CSS variables mapped onto paths where
// one is a prefix of the other, which would silently drop a token.
func insertPath(root map[string]any, segments []string, token map[string]any) error {
	node := root
	for i, seg := range segments {
		last := i == len(segments)-1
		if last {
			if _, exists := node[seg]; exists {
				return fmt.Errorf("token path %q collides with an existing entry",
					strings.Join(segments, "."))
			}
			node[seg] = token
			return nil
		}
		child, ok := node[seg]
		if !ok {
			next := map[string]any{}
			node[seg] = next
			node = next
			continue
		}
		childMap, ok := child.(map[string]any)
		if !ok {
			return fmt.Errorf("token path %q passes through a non-group at %q",
				strings.Join(segments, "."), seg)
		}
		if _, isToken := childMap["$value"]; isToken {
			return fmt.Errorf("token path %q passes through the token at %q",
				strings.Join(segments, "."), seg)
		}
		node = childMap
	}
	return nil
}
