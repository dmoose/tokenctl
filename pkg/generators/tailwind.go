package generators

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dmoose/tokctl/pkg/tokens"
)

// TailwindGenerator generates Tailwind 4 CSS
type TailwindGenerator struct {
}

func NewTailwindGenerator() *TailwindGenerator {
	return &TailwindGenerator{}
}

// Generate creates the CSS content from a resolved token map
func (g *TailwindGenerator) Generate(tokens map[string]interface{}) (string, error) {
	var sb strings.Builder
	sb.WriteString("@import \"tailwindcss\";\n\n")
	sb.WriteString("@theme {\n")

	// Filter and sort keys for deterministic output
	keys := make([]string, 0, len(tokens))
	for k := range tokens {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, path := range keys {
		value := tokens[path]
		// Skip components - they belong in @layer components
		if _, ok := value.(map[string]interface{}); ok {
			continue // Skip maps (like components or groups) that weren't flattened to a primitive
		}

		cssVar := strings.ReplaceAll(path, ".", "-")
		sb.WriteString(fmt.Sprintf("  --%s: %v;\n", cssVar, value))
	}
	sb.WriteString("}\n\n")

	// 2. Generate Components (@layer components)
	// We need to re-scan for components specifically
	// Ideally we would pass the Dictionary here, but we only have flattened map.
	// We need to pass the dictionary or use a hack.
	// Let's modify Generate to accept *tokens.Dictionary or do extraction before.
	// For now, let's assume we can't easily get them from the flattened map unless we put them there.
	// Update: The current signature is Generate(map[string]interface{}).
	// We need to update the signature to support components.
	return sb.String(), nil
}

// GenerateComponents generates CSS for semantic components
func (g *TailwindGenerator) GenerateComponents(components map[string]tokens.ComponentDefinition) (string, error) {
	var sb strings.Builder
	sb.WriteString("@layer components {\n")

	for _, comp := range components {
		// Base Class
		if comp.Class != "" {
			sb.WriteString(fmt.Sprintf("  .%s {\n", comp.Class))
			writeProps(&sb, comp.Base, 4)
			sb.WriteString("  }\n")
		}

		// Variants
		for _, variant := range comp.Variants {
			if variant.Class != "" {
				sb.WriteString(fmt.Sprintf("  .%s {\n", variant.Class))
				writeProps(&sb, variant.Properties, 4)
				sb.WriteString("  }\n")

				// States inside variant
				for stateKey, state := range variant.States {
					// Handle state syntax like "&:hover" or ":hover"
					selector := stateKey
					if strings.HasPrefix(stateKey, "&") {
						selector = fmt.Sprintf(".%s%s", variant.Class, stateKey[1:])
					} else if strings.HasPrefix(stateKey, ":") {
						selector = fmt.Sprintf(".%s%s", variant.Class, stateKey)
					} else {
						// Fallback or complex selector
						selector = fmt.Sprintf(".%s %s", variant.Class, stateKey)
					}

					sb.WriteString(fmt.Sprintf("  %s {\n", selector))
					writeProps(&sb, state.Properties, 4)
					sb.WriteString("  }\n")
				}
			}
		}

		// Sizes
		for _, size := range comp.Sizes {
			if size.Class != "" {
				sb.WriteString(fmt.Sprintf("  .%s {\n", size.Class))
				writeProps(&sb, size.Properties, 4)
				sb.WriteString("  }\n")
			}
		}
	}

	sb.WriteString("}\n")
	return sb.String(), nil
}

func writeProps(sb *strings.Builder, props map[string]interface{}, indent int) {
	padding := strings.Repeat(" ", indent)
	for k, v := range props {
		// Skip metadata keys (start with $)
		if strings.HasPrefix(k, "$") {
			continue
		}

		// Serialize complex types (arrays, etc) with context-aware handling
		valStr := SerializeValueForProperty(k, v)

		// Resolve simple tokens to var(--token) if they look like {token}
		val := valStr
		if strings.HasPrefix(valStr, "{") && strings.HasSuffix(valStr, "}") {
			tokenName := valStr[1 : len(valStr)-1]
			val = fmt.Sprintf("var(--%s)", strings.ReplaceAll(tokenName, ".", "-"))
		}

		sb.WriteString(fmt.Sprintf("%s%s: %s;\n", padding, k, val))
	}
}
