package generators

import (
	"fmt"
	"strings"
)

// SerializeValue converts any interface{} to a CSS value string
// For arrays, uses comma separation (safe default for most CSS properties)
func SerializeValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case []interface{}:
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// SerializeValueForProperty converts a value to CSS with context-aware array handling
// Different CSS properties require different separators for array values:
// - Space-separated: margin, padding, border-width, border-radius, etc.
// - Comma-separated: font-family, box-shadow, text-shadow, transform, transition
func SerializeValueForProperty(property string, val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case []interface{}:
		separator := getArraySeparator(property)
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, separator)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// getArraySeparator returns the appropriate separator for CSS property arrays
func getArraySeparator(property string) string {
	// Normalize property name (remove vendor prefixes, convert to lowercase)
	prop := strings.ToLower(property)
	prop = strings.TrimPrefix(prop, "-webkit-")
	prop = strings.TrimPrefix(prop, "-moz-")
	prop = strings.TrimPrefix(prop, "-ms-")
	prop = strings.TrimPrefix(prop, "-o-")

	// Properties that use space separation
	spaceSeperatedProps := map[string]bool{
		// Box model
		"margin":        true,
		"padding":       true,
		"border-width":  true,
		"border-style":  true,
		"border-color":  true,
		"border-radius": true,
		"inset":         true,

		// Backgrounds
		"background-position": true,
		"background-size":     true,

		// Flexbox/Grid
		"grid-template-columns": true,
		"grid-template-rows":    true,
		"grid-template-areas":   true,
		"grid-gap":              true,
		"gap":                   true,
		"flex":                  true,

		// Text
		"text-decoration": true,
		"font":            true, // shorthand

		// Borders shorthand
		"border":        true,
		"border-top":    true,
		"border-right":  true,
		"border-bottom": true,
		"border-left":   true,

		// Outline
		"outline": true,

		// Others
		"clip-path": true,
		"offset":    true,
	}

	if spaceSeperatedProps[prop] {
		return " "
	}

	// Comma-separated is the default for:
	// - font-family
	// - box-shadow, text-shadow, filter, backdrop-filter
	// - transform, transform-origin
	// - transition, animation
	// - background (multi-layer), background-image
	// - And most other multi-value properties
	return ", "
}
