package generators

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

// serializeValueForCSS converts any value to a proper CSS string
// Handles arrays by joining with appropriate separator based on value type
func serializeValueForCSS(val any) string {
	switch v := val.(type) {
	case []any:
		// Arrays are comma-separated for CSS custom properties
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, ", ")
	case []string:
		// Handle string slices as well
		return strings.Join(v, ", ")
	default:
		// Check for other slice types via reflection-like approach
		str := fmt.Sprintf("%v", val)
		// If the default formatting produced brackets, it's likely a slice we didn't handle
		// This is a fallback - ideally all slice types should be handled above
		if len(str) > 2 && str[0] == '[' && str[len(str)-1] == ']' {
			// Remove brackets and replace spaces with ", " for comma separation
			inner := str[1 : len(str)-1]
			// This is a simple heuristic - space-separated values inside brackets
			parts := strings.Fields(inner)
			if len(parts) > 1 {
				return strings.Join(parts, ", ")
			}
		}
		return str
	}
}

// GenerationContext provides all necessary data for generation
type GenerationContext struct {
	BaseDict         *tokens.Dictionary                    // Original base dictionary (unresolved)
	ResolvedTokens   map[string]any                // Flattened, resolved atomic tokens
	Components       map[string]tokens.ComponentDefinition // Extracted components
	Themes           map[string]ThemeContext               // Theme-specific contexts
	PropertyTokens   []tokens.PropertyToken                // Tokens with $property for @property declarations
	Breakpoints      map[string]string                     // Breakpoint definitions (name -> min-width)
	ResponsiveTokens []tokens.ResponsiveToken              // Tokens with responsive overrides
}

// ThemeContext provides theme-specific generation data
type ThemeContext struct {
	Dict           *tokens.Dictionary     // Full theme dictionary
	ResolvedTokens map[string]any // Resolved tokens for this theme
	DiffTokens     map[string]any // Only tokens that differ from base
}

// TailwindGenerator generates Tailwind 4 CSS
type TailwindGenerator struct {
}

func NewTailwindGenerator() *TailwindGenerator {
	return &TailwindGenerator{}
}

// Generate creates complete Tailwind CSS from generation context
func (g *TailwindGenerator) Generate(ctx *GenerationContext) (string, error) {
	var sb strings.Builder

	// 1. @property declarations (before @theme for type registration)
	if len(ctx.PropertyTokens) > 0 {
		propertyDecls := g.generatePropertyDeclarations(ctx.PropertyTokens)
		sb.WriteString(propertyDecls)
	}

	// 2. Import and base @theme block
	baseTheme, err := g.generateBaseTheme(ctx.ResolvedTokens)
	if err != nil {
		return "", fmt.Errorf("failed to generate base theme: %w", err)
	}
	sb.WriteString(baseTheme)

	// 3. Theme variations in @layer base
	if len(ctx.Themes) > 0 {
		themeVariations, err := g.generateThemeVariations(ctx.Themes)
		if err != nil {
			return "", fmt.Errorf("failed to generate theme variations: %w", err)
		}
		sb.WriteString("\n")
		sb.WriteString(themeVariations)
	}

	// 4. Components in @layer components (always output for consistency)
	components, err := g.generateComponents(ctx.Components)
	if err != nil {
		return "", fmt.Errorf("failed to generate components: %w", err)
	}
	sb.WriteString("\n")
	sb.WriteString(components)

	// 5. Responsive overrides via media queries
	if len(ctx.ResponsiveTokens) > 0 {
		responsiveCSS := tokens.GenerateResponsiveCSS(ctx.Breakpoints, ctx.ResponsiveTokens)
		if responsiveCSS != "" {
			sb.WriteString("\n")
			sb.WriteString(responsiveCSS)
		}
	}

	return sb.String(), nil
}

// generatePropertyDeclarations creates @property declarations for typed tokens
func (g *TailwindGenerator) generatePropertyDeclarations(properties []tokens.PropertyToken) string {
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

// generateBaseTheme creates the root @theme block with base tokens
func (g *TailwindGenerator) generateBaseTheme(resolvedTokens map[string]any) (string, error) {
	var sb strings.Builder
	sb.WriteString("@import \"tailwindcss\";\n\n")
	sb.WriteString("@theme {\n")

	// Sort keys for deterministic output
	keys := make([]string, 0, len(resolvedTokens))
	for k := range resolvedTokens {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, path := range keys {
		value := resolvedTokens[path]
		// Skip non-primitive values (shouldn't happen in resolved tokens, but defensive)
		if _, ok := value.(map[string]any); ok {
			continue
		}

		cssVar := strings.ReplaceAll(path, ".", "-")
		cssValue := serializeValueForCSS(value)
		sb.WriteString(fmt.Sprintf("  --%s: %s;\n", cssVar, cssValue))
	}

	sb.WriteString("}\n\n")
	return sb.String(), nil
}

// generateThemeVariations creates @layer base with theme-specific overrides
func (g *TailwindGenerator) generateThemeVariations(themes map[string]ThemeContext) (string, error) {
	var sb strings.Builder
	sb.WriteString("@layer base {\n")

	// Sort theme names for deterministic output
	themeNames := make([]string, 0, len(themes))
	for name := range themes {
		themeNames = append(themeNames, name)
	}
	sort.Strings(themeNames)

	for _, themeName := range themeNames {
		themeCtx := themes[themeName]

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

	sb.WriteString("}\n")
	return sb.String(), nil
}

// generateComponents creates @layer components with component styles
func (g *TailwindGenerator) generateComponents(components map[string]tokens.ComponentDefinition) (string, error) {
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
			g.writeProperties(&sb, baseProps, 4)
			sb.WriteString("  }\n")

			// Write nested pseudo-selectors as separate rules
			nestedKeys := make([]string, 0, len(nestedSelectors))
			for k := range nestedSelectors {
				nestedKeys = append(nestedKeys, k)
			}
			sort.Strings(nestedKeys)

			for _, selectorKey := range nestedKeys {
				props := nestedSelectors[selectorKey]
				selector := g.buildStateSelector(comp.Class, selectorKey)
				sb.WriteString(fmt.Sprintf("  %s {\n", selector))
				g.writeProperties(&sb, props, 4)
				sb.WriteString("  }\n")
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
				g.writeProperties(&sb, variant.Properties, 4)
				sb.WriteString("  }\n")

				// States inside variant (hover, focus, etc)
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
					sb.WriteString("  }\n")
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
				sb.WriteString("  }\n")
			}
		}
	}

	sb.WriteString("}\n")
	return sb.String(), nil
}

// buildStateSelector converts state key to CSS selector
func (g *TailwindGenerator) buildStateSelector(className, stateKey string) string {
	// Handle state syntax like "&:hover" or ":hover"
	if strings.HasPrefix(stateKey, "&") {
		return fmt.Sprintf(".%s%s", className, stateKey[1:])
	} else if strings.HasPrefix(stateKey, ":") {
		return fmt.Sprintf(".%s%s", className, stateKey)
	}
	// Fallback for complex selectors
	return fmt.Sprintf(".%s %s", className, stateKey)
}

// writeProperties writes CSS properties with proper indentation and serialization
func (g *TailwindGenerator) writeProperties(sb *strings.Builder, props map[string]any, indent int) {
	if len(props) == 0 {
		return
	}

	padding := strings.Repeat(" ", indent)

	// Sort property names for deterministic output
	propNames := make([]string, 0, len(props))
	for k := range props {
		propNames = append(propNames, k)
	}
	sort.Strings(propNames)

	for _, k := range propNames {
		v := props[k]

		// Skip metadata keys (start with $)
		if strings.HasPrefix(k, "$") {
			continue
		}

		// Serialize complex types (arrays, etc) with context-aware handling
		valStr := SerializeValueForProperty(k, v)

		// Resolve all token references to var(--token)
		val := resolveTokenReferences(valStr)

		fmt.Fprintf(sb, "%s%s: %s;\n", padding, k, val)
	}
}

// Legacy methods for backwards compatibility during migration

// GenerateFromResolved is deprecated - use Generate with GenerationContext
// Kept for backwards compatibility with existing tests
func (g *TailwindGenerator) GenerateFromResolved(tokens map[string]any) (string, error) {
	return g.generateBaseTheme(tokens)
}

// GenerateComponents is deprecated - use Generate with GenerationContext
// Kept for backwards compatibility with existing tests
func (g *TailwindGenerator) GenerateComponents(components map[string]tokens.ComponentDefinition) (string, error) {
	return g.generateComponents(components)
}

// resolveTokenReferences converts all {token.path} references to var(--token-path)
// Handles multiple references in a single string: "{spacing.sm} {spacing.md}" -> "var(--spacing-sm) var(--spacing-md)"
func resolveTokenReferences(value string) string {
	// Pattern matches {token.path.here}
	refPattern := regexp.MustCompile(`\{([^}]+)\}`)

	// Replace all matches
	result := refPattern.ReplaceAllStringFunc(value, func(match string) string {
		// Extract token path (remove { and })
		tokenPath := match[1 : len(match)-1]
		// Convert to CSS variable format
		cssVar := strings.ReplaceAll(tokenPath, ".", "-")
		return fmt.Sprintf("var(--%s)", cssVar)
	})

	return result
}
