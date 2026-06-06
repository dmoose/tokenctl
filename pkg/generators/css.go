// tokenctl/pkg/generators/css.go
package generators

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

// CSSGenerator generates pure CSS without Tailwind dependencies
type CSSGenerator struct {
}

func NewCSSGenerator() *CSSGenerator {
	return &CSSGenerator{}
}

// Generate creates pure CSS from generation context
func (g *CSSGenerator) Generate(ctx *GenerationContext) (string, error) {
	var sb strings.Builder

	// 1. Layer order declaration
	sb.WriteString("@layer reset, tokens, themes, components;\n\n")

	// 2. @property declarations (if any)
	if len(ctx.PropertyTokens) > 0 {
		sb.WriteString(generatePropertyDeclarations(ctx.PropertyTokens))
	}

	// 3. @keyframes declarations (global animations)
	if len(ctx.Keyframes) > 0 {
		keyframesCSS := tokens.GenerateKeyframesCSS(ctx.Keyframes)
		sb.WriteString(keyframesCSS)
	}

	// 4. Reset layer
	sb.WriteString(generateReset())

	// 5. Root variables (in tokens layer)
	rootVars, err := g.generateRootVariables(ctx.ResolvedTokens)
	if err != nil {
		return "", fmt.Errorf("failed to generate root variables: %w", err)
	}
	sb.WriteString(rootVars)

	// 6. Theme variations
	if len(ctx.Themes) > 0 {
		defaultTheme := ctx.DefaultTheme
		if defaultTheme == "" {
			defaultTheme = DefaultThemeName
		}
		themeVariations, err := g.generateThemeVariations(ctx.Themes, defaultTheme)
		if err != nil {
			return "", fmt.Errorf("failed to generate theme variations: %w", err)
		}
		sb.WriteString(themeVariations)
	}

	// 7. Components
	if len(ctx.Components) > 0 {
		components, err := g.generateComponents(ctx.Components, ctx.Breakpoints)
		if err != nil {
			return "", fmt.Errorf("failed to generate components: %w", err)
		}
		sb.WriteString(components)
	}

	// 8. Responsive overrides via media queries
	if len(ctx.ResponsiveTokens) > 0 {
		responsiveCSS := tokens.GenerateResponsiveCSS(ctx.Breakpoints, ctx.ResponsiveTokens)
		if responsiveCSS != "" {
			sb.WriteString("\n")
			sb.WriteString(responsiveCSS)
		}
	}

	// 9. Container query overrides
	if len(ctx.ContainerOverrides) > 0 {
		containerCSS := GenerateContainerCSS(ctx.ContainerOverrides)
		if containerCSS != "" {
			sb.WriteString("\n")
			sb.WriteString(containerCSS)
		}
	}

	return sb.String(), nil
}

// generateRootVariables creates :root block with base tokens in @layer tokens
func (g *CSSGenerator) generateRootVariables(resolvedTokens map[string]any) (string, error) {
	var sb strings.Builder
	sb.WriteString("@layer tokens {\n")
	sb.WriteString("  :root {\n")

	// Sort keys for deterministic output
	keys := make([]string, 0, len(resolvedTokens))
	for k := range resolvedTokens {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, path := range keys {
		value := resolvedTokens[path]
		// Skip non-primitive values
		if _, ok := value.(map[string]any); ok {
			continue
		}

		cssVar := strings.ReplaceAll(path, ".", "-")
		cssValue := serializeValueForCSS(value)
		sb.WriteString(fmt.Sprintf("    --%s: %s;\n", cssVar, cssValue))
	}

	sb.WriteString("  }\n")
	sb.WriteString("}\n\n")
	return sb.String(), nil
}

// generateThemeVariations creates theme-specific CSS with data-theme selectors
func (g *CSSGenerator) generateThemeVariations(themes map[string]ThemeContext, defaultTheme string) (string, error) {
	var sb strings.Builder
	sb.WriteString("@layer themes {\n")

	// Sort: default theme first so non-default themes override :root via cascade
	themeNames := make([]string, 0, len(themes))
	for name := range themes {
		themeNames = append(themeNames, name)
	}
	sortThemeNames(themeNames, defaultTheme)

	for _, themeName := range themeNames {
		themeCtx := themes[themeName]

		// Skip if no diff tokens
		if len(themeCtx.DiffTokens) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("  %s {\n", themeSelector(themeName, defaultTheme)))

		// Sort token keys for deterministic output
		tokenKeys := make([]string, 0, len(themeCtx.DiffTokens))
		for key := range themeCtx.DiffTokens {
			tokenKeys = append(tokenKeys, key)
		}
		sort.Strings(tokenKeys)

		for _, key := range tokenKeys {
			val := themeCtx.DiffTokens[key]
			cssVar := strings.ReplaceAll(key, ".", "-")
			cssValue := serializeValueForCSS(val)
			sb.WriteString(fmt.Sprintf("    --%s: %s;\n", cssVar, cssValue))
		}

		sb.WriteString("  }\n")
	}

	sb.WriteString("}\n\n")
	return sb.String(), nil
}

// generateComponents creates @layer components with component styles.
// breakpoints are used to emit per-component @media rules for any
// component property declared with {"$value": ..., "$responsive": {bp: value}}.
func (g *CSSGenerator) generateComponents(components map[string]tokens.ComponentDefinition, breakpoints map[string]string) (string, error) {
	var sb strings.Builder
	sb.WriteString("@layer components {\n")

	// Sort component names for deterministic output
	compNames := make([]string, 0, len(components))
	for name := range components {
		compNames = append(compNames, name)
	}
	sort.Strings(compNames)

	// Collected across all components: per-component responsive overrides
	// emitted as @media (min-width: <bp>) { .<class> { <prop>: <val>; } }
	// after the main components block. Outer key: breakpoint name (sorted
	// for emission). Inner: class name → property map.
	componentResponsive := map[string]map[string]map[string]any{}

	for _, name := range compNames {
		comp := components[name]

		// Base class
		if comp.Class != "" {
			// Separate base properties from nested pseudo-selectors
			baseProps := make(map[string]any)
			nestedSelectors := make(map[string]map[string]any)

			for k, v := range comp.Base {
				if strings.HasPrefix(k, "&") || strings.HasPrefix(k, ":") {
					// This is a nested pseudo-selector
					if nested, ok := v.(map[string]any); ok {
						nestedSelectors[k] = nested
					}
				} else {
					baseProps[k] = v
				}
			}

			// Collect $responsive overrides for this class. Walks
			// baseProps for values shaped {"$value": ..., "$responsive": {bp: val}},
			// hoists each breakpoint's val into componentResponsive.
			// $value continues to flow through writeProperties via
			// SerializeValueForProperty so the base property renders normally.
			collectComponentResponsive(comp.Class, baseProps, componentResponsive)
			for selKey, nested := range nestedSelectors {
				stateClass := buildStateSelector(comp.Class, selKey)
				collectComponentResponsive(stateClass, nested, componentResponsive)
			}

			sb.WriteString(fmt.Sprintf("  .%s {\n", comp.Class))
			writeProperties(&sb, baseProps, 4)
			sb.WriteString("  }\n\n")

			// Write nested pseudo-selectors
			nestedKeys := make([]string, 0, len(nestedSelectors))
			for k := range nestedSelectors {
				nestedKeys = append(nestedKeys, k)
			}
			sort.Strings(nestedKeys)

			for _, selectorKey := range nestedKeys {
				props := nestedSelectors[selectorKey]
				selector := buildStateSelector(comp.Class, selectorKey)
				sb.WriteString(fmt.Sprintf("  %s {\n", selector))
				writeProperties(&sb, props, 4)
				sb.WriteString("  }\n\n")
			}
		}

		// Variants
		variantNames := make([]string, 0, len(comp.Variants))
		for vname := range comp.Variants {
			variantNames = append(variantNames, vname)
		}
		sort.Strings(variantNames)

		for _, vname := range variantNames {
			variant := comp.Variants[vname]
			if variant.Class != "" {
				sb.WriteString(fmt.Sprintf("  .%s {\n", variant.Class))
				writeProperties(&sb, variant.Properties, 4)
				sb.WriteString("  }\n\n")

				collectComponentResponsive(variant.Class, variant.Properties, componentResponsive)

				// States
				stateKeys := make([]string, 0, len(variant.States))
				for skey := range variant.States {
					stateKeys = append(stateKeys, skey)
				}
				sort.Strings(stateKeys)

				for _, stateKey := range stateKeys {
					state := variant.States[stateKey]
					selector := buildStateSelector(variant.Class, stateKey)
					sb.WriteString(fmt.Sprintf("  %s {\n", selector))
					writeProperties(&sb, state.Properties, 4)
					sb.WriteString("  }\n\n")

					collectComponentResponsive(selector, state.Properties, componentResponsive)
				}
			}
		}

		// Sizes
		sizeNames := make([]string, 0, len(comp.Sizes))
		for sname := range comp.Sizes {
			sizeNames = append(sizeNames, sname)
		}
		sort.Strings(sizeNames)

		for _, sname := range sizeNames {
			size := comp.Sizes[sname]
			if size.Class != "" {
				sb.WriteString(fmt.Sprintf("  .%s {\n", size.Class))
				writeProperties(&sb, size.Properties, 4)
				sb.WriteString("  }\n\n")

				collectComponentResponsive(size.Class, size.Properties, componentResponsive)
			}
		}

		// States (error, active, disabled, etc.)
		stateNames := make([]string, 0, len(comp.States))
		for sname := range comp.States {
			stateNames = append(stateNames, sname)
		}
		sort.Strings(stateNames)

		for _, sname := range stateNames {
			state := comp.States[sname]
			if state.Class != "" {
				sb.WriteString(fmt.Sprintf("  .%s {\n", state.Class))
				writeProperties(&sb, state.Properties, 4)
				sb.WriteString("  }\n\n")

				// States can also have pseudo-selectors
				stateKeys := make([]string, 0, len(state.States))
				for skey := range state.States {
					stateKeys = append(stateKeys, skey)
				}
				sort.Strings(stateKeys)

				for _, stateKey := range stateKeys {
					pseudoState := state.States[stateKey]
					selector := buildStateSelector(state.Class, stateKey)
					sb.WriteString(fmt.Sprintf("  %s {\n", selector))
					writeProperties(&sb, pseudoState.Properties, 4)
					sb.WriteString("  }\n\n")
				}
			}
		}
	}

	sb.WriteString("}\n")

	// Per-component responsive overrides. For any property declared as
	// {"$value": ..., "$responsive": {bp: val}}, emit one @media block
	// per breakpoint with the matching class+property rules. The base
	// $value is already rendered via writeProperties; these blocks
	// override it at the breakpoint. Wrapped in @layer components so
	// the cascade order matches the base rules.
	if len(componentResponsive) > 0 && len(breakpoints) > 0 {
		mediaCSS := generateComponentResponsiveCSS(breakpoints, componentResponsive)
		if mediaCSS != "" {
			sb.WriteString("\n")
			sb.WriteString(mediaCSS)
		}
	}

	return sb.String(), nil
}

// collectComponentResponsive walks `props` for token-shaped values
// ({"$value": ..., "$responsive": {bp: val}}) and records each
// breakpoint's override into `out[bp][class][prop] = val`. The base
// $value continues to render in the main class rule via
// SerializeValueForProperty; this only collects the breakpoint
// overrides for hoisting into per-component @media blocks.
func collectComponentResponsive(class string, props map[string]any, out map[string]map[string]map[string]any) {
	for k, v := range props {
		if strings.HasPrefix(k, "$") {
			continue
		}
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		respRaw, ok := m["$responsive"]
		if !ok {
			continue
		}
		resp, ok := respRaw.(map[string]any)
		if !ok {
			continue
		}
		for bp, val := range resp {
			if out[bp] == nil {
				out[bp] = make(map[string]map[string]any)
			}
			if out[bp][class] == nil {
				out[bp][class] = make(map[string]any)
			}
			out[bp][class][k] = val
		}
	}
}

// generateComponentResponsiveCSS emits @layer components { @media (...) { .class { ... } } }
// blocks for each breakpoint's collected overrides. Wrapping in
// @layer components keeps cascade order intact relative to the base
// rules — site-local overrides in @layer site still win.
func generateComponentResponsiveCSS(breakpoints map[string]string, overrides map[string]map[string]map[string]any) string {
	// Sort breakpoints by pixel size for deterministic mobile-first order.
	bpNames := sortBreakpointsBySize(breakpoints)

	var sb strings.Builder
	sb.WriteString("@layer components {\n")
	for _, bp := range bpNames {
		classes, ok := overrides[bp]
		if !ok {
			continue
		}
		minWidth, ok := breakpoints[bp]
		if !ok {
			continue
		}
		fmt.Fprintf(&sb, "  @media (min-width: %s) {\n", minWidth)

		// Sort class names for deterministic output.
		classNames := make([]string, 0, len(classes))
		for c := range classes {
			classNames = append(classNames, c)
		}
		sort.Strings(classNames)

		for _, class := range classNames {
			props := classes[class]
			// Selector: bare classes already start with "."? collectComponentResponsive
			// records "selector" form: state-selectors include the leading "."
			// already (via buildStateSelector); base classes are the bare class
			// name without ".", so prepend.
			selector := class
			if !strings.HasPrefix(selector, ".") {
				selector = "." + selector
			}
			fmt.Fprintf(&sb, "    %s {\n", selector)
			writeProperties(&sb, props, 6)
			sb.WriteString("    }\n")
		}
		sb.WriteString("  }\n")
	}
	sb.WriteString("}\n")
	return sb.String()
}

// sortBreakpointsBySize returns breakpoint names sorted ascending by px.
// Local helper to avoid coupling generators/css.go to the tokens package
// for a sort helper; mirrors tokens.sortBreakpointsBySize.
func sortBreakpointsBySize(breakpoints map[string]string) []string {
	type entry struct {
		name string
		px   int
	}
	entries := make([]entry, 0, len(breakpoints))
	for name, val := range breakpoints {
		px := 0
		fmt.Sscanf(val, "%dpx", &px)
		entries = append(entries, entry{name: name, px: px})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].px < entries[j].px })
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.name
	}
	return out
}

// generateReset creates a minimal modern CSS reset in @layer reset
func generateReset() string {
	return `@layer reset {
  *, *::before, *::after { box-sizing: border-box; }
  * { margin: 0; }
  html { line-height: var(--leading-normal, 1.5); -webkit-text-size-adjust: 100%; }
  body { font-family: var(--font-family-sans, system-ui, sans-serif); background-color: var(--color-background); color: var(--color-foreground); }
  a { color: var(--color-link, inherit); }
  a:visited { color: var(--color-link-visited, var(--color-link, inherit)); }
  img, picture, video, canvas, svg { display: block; max-width: 100%; }
  input, button, textarea, select { font: inherit; }
  p, h1, h2, h3, h4, h5, h6 { overflow-wrap: break-word; }
  code, kbd, pre { font-family: var(--font-family-mono, ui-monospace, monospace); }
  hr { border-color: var(--color-border, currentColor); }
  table { border-collapse: collapse; }
}

`
}

