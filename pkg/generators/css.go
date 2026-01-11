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
		propertyDecls := g.generatePropertyDeclarations(ctx.PropertyTokens)
		sb.WriteString(propertyDecls)
	}

	// 3. Reset layer
	sb.WriteString(generateReset())

	// 4. Root variables (in tokens layer)
	rootVars, err := g.generateRootVariables(ctx.ResolvedTokens)
	if err != nil {
		return "", fmt.Errorf("failed to generate root variables: %w", err)
	}
	sb.WriteString(rootVars)

	// 5. Theme variations
	if len(ctx.Themes) > 0 {
		themeVariations, err := g.generateThemeVariations(ctx.Themes)
		if err != nil {
			return "", fmt.Errorf("failed to generate theme variations: %w", err)
		}
		sb.WriteString(themeVariations)
	}

	// 6. Components
	if len(ctx.Components) > 0 {
		components, err := g.generateComponents(ctx.Components)
		if err != nil {
			return "", fmt.Errorf("failed to generate components: %w", err)
		}
		sb.WriteString(components)
	}

	// 7. Responsive overrides via media queries
	if len(ctx.ResponsiveTokens) > 0 {
		responsiveCSS := tokens.GenerateResponsiveCSS(ctx.Breakpoints, ctx.ResponsiveTokens)
		if responsiveCSS != "" {
			sb.WriteString("\n")
			sb.WriteString(responsiveCSS)
		}
	}

	return sb.String(), nil
}

// generatePropertyDeclarations creates @property declarations
func (g *CSSGenerator) generatePropertyDeclarations(properties []tokens.PropertyToken) string {
	var sb strings.Builder

	// Sort by path for deterministic output
	sorted := make([]tokens.PropertyToken, len(properties))
	copy(sorted, properties)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	for _, prop := range sorted {
		sb.WriteString(fmt.Sprintf("@property %s {\n", prop.CSSName))
		sb.WriteString(fmt.Sprintf("  syntax: '%s';\n", prop.CSSSyntax))
		if prop.Inherits {
			sb.WriteString("  inherits: true;\n")
		} else {
			sb.WriteString("  inherits: false;\n")
		}
		sb.WriteString(fmt.Sprintf("  initial-value: %s;\n", prop.InitialValue))
		sb.WriteString("}\n\n")
	}

	return sb.String()
}

// generateRootVariables creates :root block with base tokens in @layer tokens
func (g *CSSGenerator) generateRootVariables(resolvedTokens map[string]interface{}) (string, error) {
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
		if _, ok := value.(map[string]interface{}); ok {
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
		if themeName == "light" {
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
			sb.WriteString(fmt.Sprintf("  .%s {\n", comp.Class))
			g.writeProperties(&sb, comp.Base, 4)
			sb.WriteString("  }\n\n")
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
				g.writeProperties(&sb, variant.Properties, 4)
				sb.WriteString("  }\n\n")

				// States
				stateKeys := make([]string, 0, len(variant.States))
				for skey := range variant.States {
					stateKeys = append(stateKeys, skey)
				}
				sort.Strings(stateKeys)

				for _, stateKey := range stateKeys {
					state := variant.States[stateKey]
					selector := g.buildStateSelector(variant.Class, stateKey)
					sb.WriteString(fmt.Sprintf("  %s {\n", selector))
					g.writeProperties(&sb, state.Properties, 4)
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
				g.writeProperties(&sb, size.Properties, 4)
				sb.WriteString("  }\n\n")
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

// buildStateSelector converts state key to CSS selector
func (g *CSSGenerator) buildStateSelector(className, stateKey string) string {
	if strings.HasPrefix(stateKey, "&") {
		return fmt.Sprintf(".%s%s", className, stateKey[1:])
	} else if strings.HasPrefix(stateKey, ":") {
		return fmt.Sprintf(".%s%s", className, stateKey)
	}
	return fmt.Sprintf(".%s %s", className, stateKey)
}

// writeProperties writes CSS properties with proper indentation
func (g *CSSGenerator) writeProperties(sb *strings.Builder, props map[string]interface{}, indent int) {
	if len(props) == 0 {
		return
	}

	padding := strings.Repeat(" ", indent)

	propNames := make([]string, 0, len(props))
	for k := range props {
		propNames = append(propNames, k)
	}
	sort.Strings(propNames)

	for _, k := range propNames {
		v := props[k]

		if strings.HasPrefix(k, "$") {
			continue
		}

		valStr := SerializeValueForProperty(k, v)
		val := resolveTokenReferences(valStr)

		sb.WriteString(fmt.Sprintf("%s%s: %s;\n", padding, k, val))
	}
}
