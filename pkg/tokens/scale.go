// tokctl/pkg/tokens/scale.go

package tokens

import (
	"fmt"
	"strings"
)

// ExpandScales walks the dictionary and expands any tokens with $scale definitions
// A token with $scale like:
//
//	{
//	  "size": {
//	    "field": {
//	      "$value": "2.5rem",
//	      "$scale": { "xs": 0.6, "sm": 0.8, "md": 1.0, "lg": 1.2, "xl": 1.4 }
//	    }
//	  }
//	}
//
// Expands to create additional tokens:
//
//	size.field-xs: calc({size.field} * 0.6)
//	size.field-sm: calc({size.field} * 0.8)
//	size.field-md: {size.field}  (1.0 = no change, just reference)
//	size.field-lg: calc({size.field} * 1.2)
//	size.field-xl: calc({size.field} * 1.4)
func ExpandScales(d *Dictionary) error {
	return expandScalesRecursive(d, d.Root, "")
}

func expandScalesRecursive(d *Dictionary, node map[string]interface{}, currentPath string) error {
	// Collect keys to avoid modifying map during iteration
	keys := make([]string, 0, len(node))
	for k := range node {
		keys = append(keys, k)
	}

	for _, key := range keys {
		if strings.HasPrefix(key, "$") {
			continue
		}

		val, ok := node[key].(map[string]interface{})
		if !ok {
			continue
		}

		childPath := key
		if currentPath != "" {
			childPath = currentPath + "." + key
		}

		// Check if this is a token with $scale
		if IsToken(val) {
			if scaleVal, hasScale := val["$scale"]; hasScale {
				scaleMap, ok := scaleVal.(map[string]interface{})
				if !ok {
					return fmt.Errorf("%s: $scale must be an object", childPath)
				}

				// Expand the scale
				if err := expandScaleToken(d, node, key, childPath, val, scaleMap); err != nil {
					return err
				}
			}
		} else {
			// Recurse into nested groups
			if err := expandScalesRecursive(d, val, childPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// expandScaleToken creates new tokens for each scale factor
func expandScaleToken(d *Dictionary, parent map[string]interface{}, baseKey, basePath string, baseToken map[string]interface{}, scale map[string]interface{}) error {
	// Get the base token's type and description for inheritance
	baseType, _ := baseToken["$type"].(string)
	baseDesc, _ := baseToken["$description"].(string)

	for scaleName, factorVal := range scale {
		factor, ok := toFloat64(factorVal)
		if !ok {
			return fmt.Errorf("%s.$scale.%s: factor must be a number", basePath, scaleName)
		}

		// Create the new token name
		newKey := baseKey + "-" + scaleName
		newPath := basePath + "-" + scaleName

		// Create the scaled token
		var newValue string
		if factor == 1.0 {
			// For factor 1.0, just reference the base token
			newValue = "{" + basePath + "}"
		} else {
			// Use calc() expression for other factors
			newValue = fmt.Sprintf("calc({%s} * %g)", basePath, factor)
		}

		newToken := map[string]interface{}{
			"$value": newValue,
		}

		// Inherit type if present
		if baseType != "" {
			newToken["$type"] = baseType
		}

		// Add description
		if baseDesc != "" {
			newToken["$description"] = fmt.Sprintf("%s (%s scale)", baseDesc, scaleName)
		} else {
			newToken["$description"] = fmt.Sprintf("%s scale variant", scaleName)
		}

		// Add to parent
		parent[newKey] = newToken

		// Track source file if the base token has one
		if sourceFile, ok := d.SourceFiles[basePath]; ok {
			d.SourceFiles[newPath] = sourceFile
		}
	}

	// Remove $scale from the base token (it's been processed)
	delete(baseToken, "$scale")

	return nil
}

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	default:
		return 0, false
	}
}

// StandardScale returns the DaisyUI-style size scale factors
func StandardScale() map[string]interface{} {
	return map[string]interface{}{
		"xs": 0.6,
		"sm": 0.8,
		"md": 1.0,
		"lg": 1.2,
		"xl": 1.4,
	}
}

// TypographyScale returns a typographic scale (based on major third)
func TypographyScale() map[string]interface{} {
	return map[string]interface{}{
		"xs":  0.64,  // 1 / 1.25^2
		"sm":  0.8,   // 1 / 1.25
		"md":  1.0,   // base
		"lg":  1.25,  // 1.25
		"xl":  1.563, // 1.25^2
		"2xl": 1.953, // 1.25^3
		"3xl": 2.441, // 1.25^4
	}
}
