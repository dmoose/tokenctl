package generators

import (
	"fmt"
	"strings"
)

// ThemeGenerator generates CSS for theme variables
type ThemeGenerator struct {
}

func NewThemeGenerator() *ThemeGenerator {
	return &ThemeGenerator{}
}

// GenerateThemes generates CSS blocks for multiple themes
func (g *ThemeGenerator) GenerateThemes(themes map[string]map[string]interface{}) (string, error) {
	var sb strings.Builder
	sb.WriteString("@layer base {\n")

	for themeName, tokens := range themes {
		selector := fmt.Sprintf(`[data-theme="%s"]`, themeName)
		if themeName == "light" {
			selector = fmt.Sprintf(`:root, [data-theme="%s"]`, themeName)
		}

		sb.WriteString(fmt.Sprintf("  %s {\n", selector))

		// Sort keys if needed, but for now we iterate
		for key, val := range tokens {
			// Flatten and output vars
			// Assumption: 'tokens' here is already a flattened map of atomic tokens
			// e.g. "color.primary": "#3b82f6"
			cssVar := strings.ReplaceAll(key, ".", "-")
			sb.WriteString(fmt.Sprintf("    --%s: %v;\n", cssVar, val))
		}

		sb.WriteString("  }\n")
	}

	sb.WriteString("}\n")
	return sb.String(), nil
}
