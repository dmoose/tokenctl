package generators

import (
	"fmt"
	"strings"
)

// SerializeValue converts any interface{} (string, array, etc) to a valid CSS value string
func SerializeValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case []interface{}:
		// Join arrays with spaces (common for short-hand props like margin/padding)
		// Or commas? Context matters. But for design tokens, space is safer default for shadows/etc unless it's font-family.
		// W3C spec usually defines shadow arrays. CSS requires comma for multiple shadows.
		// Let's check if it looks like a shadow definition.
		// For MVP, space separation is risky for multi-layer shadows.
		// Let's default to comma separation for arrays, as that is standard for multi-value props (font-family, box-shadow, transition).
		// Space separation is usually intra-value (10px 20px).
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprintf("%v", v)
	}
}
