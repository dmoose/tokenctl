package generators

import (
	"fmt"
	"sort"
	"strings"
)

// DefaultThemeName is the fallback when no theme declares "$default": true.
const DefaultThemeName = "light"

// sortThemeNames sorts theme names with the default theme first,
// then remaining themes alphabetically. This ensures the default theme's
// :root selector is emitted first so non-default themes can override it
// via CSS cascade order (same specificity, later rule wins).
func sortThemeNames(names []string, defaultTheme string) {
	sort.Slice(names, func(i, j int) bool {
		if names[i] == defaultTheme {
			return true
		}
		if names[j] == defaultTheme {
			return false
		}
		return names[i] < names[j]
	})
}

// themeSelector returns the CSS selector for a theme. The default theme
// gets `:root, [data-theme="name"]` so it applies without any attribute;
// all other themes get just `[data-theme="name"]`.
func themeSelector(themeName, defaultTheme string) string {
	if themeName == defaultTheme {
		return fmt.Sprintf(`:root, [data-theme="%s"]`, themeName)
	}
	return fmt.Sprintf(`[data-theme="%s"]`, themeName)
}

// GenerateThemes generates CSS blocks for multiple themes.
// defaultTheme controls which theme maps to :root; pass "" to use DefaultThemeName.
func GenerateThemes(themes map[string]map[string]any, defaultTheme string) (string, error) {
	if defaultTheme == "" {
		defaultTheme = DefaultThemeName
	}

	var sb strings.Builder
	sb.WriteString("@layer base {\n")

	// Sort: default theme first so non-default themes override :root via cascade
	themeNames := make([]string, 0, len(themes))
	for name := range themes {
		themeNames = append(themeNames, name)
	}
	sortThemeNames(themeNames, defaultTheme)

	for _, themeName := range themeNames {
		tokens := themes[themeName]

		sb.WriteString(fmt.Sprintf("  %s {\n", themeSelector(themeName, defaultTheme)))

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
