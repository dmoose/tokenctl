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
		themeVariations, err := g.generateThemeVariations(ctx.Themes)
		if err != nil {
			return "", fmt.Errorf("failed to generate theme variations: %w", err)
		}
		sb.WriteString(themeVariations)
	}

	// 7. Components
	if len(ctx.Components) > 0 {
		components, err := g.generateComponents(ctx.Components)
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
func (g *CSSGenerator) generateThemeVariations(themes map[string]ThemeContext) (string, error) {
	var sb strings.Builder
	sb.WriteString("@layer themes {\n")

	// Sort theme names for deterministic output
	themeNames := make([]string, 0, len(themes))
	for name := range themes {
		themeNames = append(themeNames, name)
	}
	sort.Strings(themeNames)

	for _, themeName := range themeNames {
		themeCtx := themes[themeName]

		// Skip if no diff tokens
		if len(themeCtx.DiffTokens) == 0 {
			continue
		}

		// Determine selector
		selector := fmt.Sprintf(`[data-theme="%s"]`, themeName)
		if themeName == DefaultThemeName {
			selector = fmt.Sprintf(`:root, [data-theme="%s"]`, themeName)
		}

		sb.WriteString(fmt.Sprintf("  %s {\n", selector))

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

// generateComponents creates @layer components with component styles
func (g *CSSGenerator) generateComponents(components map[string]tokens.ComponentDefinition) (string, error) {
	var sb strings.Builder
	sb.WriteString("@layer components {\n")

	// Sort component names for deterministic output
	compNames := make([]string, 0, len(components))
	for name := range components {
		compNames = append(compNames, name)
	}
	sort.Strings(compNames)

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
	return sb.String(), nil
}

// generateReset creates a minimal modern CSS reset in @layer reset
func generateReset() string {
	return `@layer reset {
  *, *::before, *::after { box-sizing: border-box; }
  * { margin: 0; }
  html { line-height: 1.5; -webkit-text-size-adjust: 100%; }
  body { font-family: var(--font-family-sans, system-ui, sans-serif); }
  img, picture, video, canvas, svg { display: block; max-width: 100%; }
  input, button, textarea, select { font: inherit; }
  p, h1, h2, h3, h4, h5, h6 { overflow-wrap: break-word; }
}

`
}

