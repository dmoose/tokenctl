// tokenctl/pkg/tokens/property.go

package tokens

import (
	"fmt"
	"strings"
)

// PropertyToken represents a token that should generate a CSS @property declaration
type PropertyToken struct {
	Path         string      // Token path (e.g., "color.primary")
	Value        any // Resolved value
	Type         string      // Token type (color, dimension, number, etc.)
	Inherits     bool        // CSS @property inherits value
	CSSName      string      // CSS variable name (e.g., "--color-primary")
	CSSSyntax    string      // CSS @property syntax (e.g., "<color>")
	InitialValue string      // CSS @property initial-value
}

// CSSPropertySyntax maps token $type to CSS @property syntax
func CSSPropertySyntax(tokenType string) string {
	switch tokenType {
	case "color":
		return "<color>"
	case "dimension":
		return "<length>"
	case "number":
		return "<number>"
	case "duration":
		return "<time>"
	case "effect":
		return "<integer>"
	default:
		// Types like fontFamily don't have a direct CSS syntax
		return ""
	}
}

// ExtractPropertyTokens scans the dictionary for tokens with $property field
// and returns PropertyToken entries for each one found
func ExtractPropertyTokens(dict *Dictionary, resolvedTokens map[string]any) []PropertyToken {
	var properties []PropertyToken
	extractPropertyTokensRecursive(dict.Root, "", resolvedTokens, &properties, "")
	return properties
}

func extractPropertyTokensRecursive(node map[string]any, currentPath string, resolvedTokens map[string]any, properties *[]PropertyToken, inheritedType string) {
	// Check for $type at this level to pass to children
	currentType := inheritedType
	if t, ok := node["$type"].(string); ok {
		currentType = t
	}

	if IsToken(node) {
		// Check for $property field
		propField, hasProperty := node["$property"]
		if !hasProperty {
			return
		}

		// Get token type (from token or inherited)
		tokenType := currentType
		if t, ok := node["$type"].(string); ok {
			tokenType = t
		}
		if tokenType == "" {
			return
		}

		// Get CSS syntax for this type
		syntax := CSSPropertySyntax(tokenType)
		if syntax == "" {
			// Skip types without CSS syntax mapping
			return
		}

		// Determine inherits value (default true)
		inherits := true
		switch v := propField.(type) {
		case bool:
			// $property: true uses defaults
			if !v {
				return // $property: false means skip
			}
		case map[string]any:
			// $property: { inherits: false }
			if inh, ok := v["inherits"].(bool); ok {
				inherits = inh
			}
		}

		// Get resolved value
		resolvedValue, ok := resolvedTokens[currentPath]
		if !ok {
			return
		}

		// Format value as string for initial-value
		initialValue := formatInitialValue(resolvedValue)

		// Build CSS variable name
		cssName := "--" + strings.ReplaceAll(currentPath, ".", "-")

		*properties = append(*properties, PropertyToken{
			Path:         currentPath,
			Value:        resolvedValue,
			Type:         tokenType,
			Inherits:     inherits,
			CSSName:      cssName,
			CSSSyntax:    syntax,
			InitialValue: initialValue,
		})
		return
	}

	// Recurse into children
	for key, val := range node {
		if strings.HasPrefix(key, "$") {
			continue
		}

		childMap, ok := val.(map[string]any)
		if !ok {
			continue
		}

		childPath := key
		if currentPath != "" {
			childPath = currentPath + "." + key
		}

		extractPropertyTokensRecursive(childMap, childPath, resolvedTokens, properties, currentType)
	}
}

// formatInitialValue converts a resolved value to a CSS initial-value string
func formatInitialValue(val any) string {
	switch v := val.(type) {
	case []any:
		// Arrays are comma-separated
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = formatInitialValue(item)
		}
		return strings.Join(parts, ", ")
	case []string:
		return strings.Join(v, ", ")
	case string:
		return v
	case float64:
		if v == float64(int(v)) {
			return fmt.Sprintf("%d", int(v))
		}
		return fmt.Sprintf("%g", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", val)
	}
}
