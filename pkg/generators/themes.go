package generators

import (
	"fmt"
	"sort"
	"strings"
)

// ThemeGenerator generates CSS for theme variables
type ThemeGenerator struct {
}

func NewThemeGenerator() *ThemeGenerator {
	return &ThemeGenerator{}
}

// GenerateThemes generates CSS blocks for multiple themes
func (g *ThemeGenerator) GenerateThemes(themes map[string]map[string]any) (string, error) {
	var sb strings.Builder
	sb.WriteString("@layer base {\n")

	// Sort theme names for deterministic output
	themeNames := make([]string, 0, len(themes))
	for name := range themes {
		themeNames = append(themeNames, name)
	}
	sort.Strings(themeNames)

	for _, themeName := range themeNames {
		tokens := themes[themeName]

		selector := fmt.Sprintf(`[data-theme="%s"]`, themeName)
		if themeName == "light" {
			selector = fmt.Sprintf(`:root, [data-theme="%s"]`, themeName)
		}

		sb.WriteString(fmt.Sprintf("  %s {\n", selector))

		// Sort token keys for deterministic output
		tokenKeys := make([]string, 0, len(tokens))
		for key := range tokens {
			tokenKeys = append(tokenKeys, key)
		}
		sort.Strings(tokenKeys)

		for _, key := range tokenKeys {
			val := tokens[key]
			cssVar := strings.ReplaceAll(key, ".", "-")
			sb.WriteString(fmt.Sprintf("    --%s: %v;\n", cssVar, val))
		}

		sb.WriteString("  }\n")
	}

	sb.WriteString("}\n")
	return sb.String(), nil
}
