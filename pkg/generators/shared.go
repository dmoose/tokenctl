package generators

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

// tokenRefPattern matches {token.path.here} references in CSS values.
// Compiled once at package level to avoid per-call overhead.
var tokenRefPattern = regexp.MustCompile(`\{([^}]+)\}`)

// resolveTokenReferences converts all {token.path} references to var(--token-path).
// Handles multiple references in a single string.
func resolveTokenReferences(value string) string {
	return tokenRefPattern.ReplaceAllStringFunc(value, func(match string) string {
		tokenPath := match[1 : len(match)-1]
		cssVar := strings.ReplaceAll(tokenPath, ".", "-")
		return fmt.Sprintf("var(--%s)", cssVar)
	})
}

// generatePropertyDeclarations creates @property declarations for typed tokens.
func generatePropertyDeclarations(properties []tokens.PropertyToken) string {
	var sb strings.Builder

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

// buildStateSelector converts a state key to a CSS selector.
func buildStateSelector(className, stateKey string) string {
	if strings.HasPrefix(stateKey, "&") {
		return fmt.Sprintf(".%s%s", className, stateKey[1:])
	} else if strings.HasPrefix(stateKey, ":") {
		return fmt.Sprintf(".%s%s", className, stateKey)
	}
	return fmt.Sprintf(".%s %s", className, stateKey)
}

// writeProperties writes CSS properties with proper indentation and serialization.
func writeProperties(sb *strings.Builder, props map[string]any, indent int) {
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

		fmt.Fprintf(sb, "%s%s: %s;\n", padding, k, val)
	}
}
