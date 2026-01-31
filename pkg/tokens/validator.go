// tokenctl/pkg/tokens/validator.go

package tokens

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dmoose/tokenctl/pkg/colors"
)

// ValidationError represents a validation issue
type ValidationError struct {
	Path       string
	Message    string
	SourceFile string // Optional: file where the token was defined
}

func (v ValidationError) Error() string {
	if v.SourceFile != "" {
		return fmt.Sprintf("%s [%s]: %s", v.Path, v.SourceFile, v.Message)
	}
	return fmt.Sprintf("%s: %s", v.Path, v.Message)
}

// Validate checks the dictionary for:
// 1. Broken references (using Resolver)
// 2. Schema compliance (basic checks)
// 3. Type-specific validation (color, dimension, number, effect)
// 4. Constraint validation ($min/$max)
func Validate(d *Dictionary) ([]ValidationError, error) {
	var errs []ValidationError

	// 1. Check References & Cycles by trying to resolve everything
	r, err := NewResolver(d)
	if err != nil {
		return nil, err
	}

	// Iterate over all tokens to find broken refs
	// The Resolver.ResolveAll() stops at first error, but we want to collect all.
	// So we manually iterate flatTokens.
	paths := make([]string, 0, len(r.flatTokens))
	for k := range r.flatTokens {
		paths = append(paths, k)
	}
	sort.Strings(paths)

	for _, path := range paths {
		val := r.flatTokens[path]
		_, err := r.ResolveValue(path, val)
		if err != nil {
			verr := ValidationError{
				Path:    path,
				Message: err.Error(),
			}
			if sourceFile, ok := d.SourceFiles[path]; ok {
				verr.SourceFile = sourceFile
			}
			errs = append(errs, verr)
		}
	}

	// 2. Schema Validation (Basic)
	errs = append(errs, validateSchema(d, d.Root, "")...)

	// 3. Type-specific and constraint validation
	errs = append(errs, validateTypes(d, d.Root, "")...)

	return errs, nil
}

func validateSchema(dict *Dictionary, node map[string]any, currentPath string) []ValidationError {
	var errs []ValidationError

	if IsToken(node) {
		return errs
	}

	for key, val := range node {
		if strings.HasPrefix(key, "$") {
			continue
		}

		childMap, ok := val.(map[string]any)
		if !ok {
			childPath := key
			if currentPath != "" {
				childPath = currentPath + "." + key
			}
			verr := ValidationError{
				Path:    childPath,
				Message: fmt.Sprintf("expected object, got %T", val),
			}
			if sourceFile, ok := dict.SourceFiles[childPath]; ok {
				verr.SourceFile = sourceFile
			}
			errs = append(errs, verr)
			continue
		}

		childPath := key
		if currentPath != "" {
			childPath = currentPath + "." + key
		}

		errs = append(errs, validateSchema(dict, childMap, childPath)...)
	}

	return errs
}

// validateTypes performs type-specific validation including constraints
// inheritedType is the $type inherited from parent groups
func validateTypes(dict *Dictionary, node map[string]any, currentPath string) []ValidationError {
	return validateTypesWithInheritance(dict, node, currentPath, "")
}

// validateTypesWithInheritance performs type validation with $type inheritance from parent groups
func validateTypesWithInheritance(dict *Dictionary, node map[string]any, currentPath string, inheritedType string) []ValidationError {
	var errs []ValidationError

	// Check for $type at this level to pass to children
	currentType := inheritedType
	if t, ok := node["$type"].(string); ok {
		currentType = t
	}

	if IsToken(node) {
		// Get token type (from token itself or inherited)
		tokenType := getTokenTypeWithInheritance(node, currentPath, currentType)
		value := node["$value"]

		// Type-specific validation
		switch tokenType {
		case "color":
			if err := validateColorFormat(value); err != nil {
				verr := ValidationError{
					Path:    currentPath,
					Message: fmt.Sprintf("invalid color: %s", err.Error()),
				}
				if sourceFile, ok := dict.SourceFiles[currentPath]; ok {
					verr.SourceFile = sourceFile
				}
				errs = append(errs, verr)
			}

		case "dimension":
			if err := validateDimension(value); err != nil {
				verr := ValidationError{
					Path:    currentPath,
					Message: fmt.Sprintf("invalid dimension: %s", err.Error()),
				}
				if sourceFile, ok := dict.SourceFiles[currentPath]; ok {
					verr.SourceFile = sourceFile
				}
				errs = append(errs, verr)
			}

		case "number":
			if err := validateNumber(value); err != nil {
				verr := ValidationError{
					Path:    currentPath,
					Message: fmt.Sprintf("invalid number: %s", err.Error()),
				}
				if sourceFile, ok := dict.SourceFiles[currentPath]; ok {
					verr.SourceFile = sourceFile
				}
				errs = append(errs, verr)
			}

		case "fontFamily":
			if err := validateFontFamily(value); err != nil {
				verr := ValidationError{
					Path:    currentPath,
					Message: fmt.Sprintf("invalid fontFamily: %s", err.Error()),
				}
				if sourceFile, ok := dict.SourceFiles[currentPath]; ok {
					verr.SourceFile = sourceFile
				}
				errs = append(errs, verr)
			}

		case "effect":
			if err := validateEffect(value); err != nil {
				verr := ValidationError{
					Path:    currentPath,
					Message: fmt.Sprintf("invalid effect: %s", err.Error()),
				}
				if sourceFile, ok := dict.SourceFiles[currentPath]; ok {
					verr.SourceFile = sourceFile
				}
				errs = append(errs, verr)
			}
		}

		// Constraint validation ($min/$max)
		errs = append(errs, validateConstraints(dict, currentPath, node, value)...)

		return errs
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

		errs = append(errs, validateTypesWithInheritance(dict, childMap, childPath, currentType)...)
	}

	return errs
}

// getTokenTypeWithInheritance extracts the $type from a token, using inherited type if not set
func getTokenTypeWithInheritance(token map[string]any, _ string, inheritedType string) string {
	if t, ok := token["$type"].(string); ok {
		return t
	}
	return inheritedType
}

// validateConstraints checks $min/$max constraints on a token value
func validateConstraints(dict *Dictionary, path string, token map[string]any, value any) []ValidationError {
	var errs []ValidationError

	constraint, err := ParseConstraints(token)
	if err != nil {
		verr := ValidationError{
			Path:    path,
			Message: fmt.Sprintf("constraint error: %s", err.Error()),
		}
		if sourceFile, ok := dict.SourceFiles[path]; ok {
			verr.SourceFile = sourceFile
		}
		errs = append(errs, verr)
		return errs
	}

	if constraint == nil {
		return errs
	}

	// Skip constraint checking for reference values (they'll be checked after resolution)
	if strVal, ok := value.(string); ok {
		if strings.Contains(strVal, "{") && strings.Contains(strVal, "}") {
			return errs
		}
	}

	if err := constraint.CheckValue(value); err != nil {
		verr := ValidationError{
			Path:    path,
			Message: fmt.Sprintf("constraint violation: %s", err.Error()),
		}
		if sourceFile, ok := dict.SourceFiles[path]; ok {
			verr.SourceFile = sourceFile
		}
		errs = append(errs, verr)
	}

	return errs
}

// validateColorFormat ensures a color value is a valid CSS color
func validateColorFormat(value any) error {
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	// Skip validation for reference values
	if strings.Contains(strVal, "{") && strings.Contains(strVal, "}") {
		return nil
	}

	// Skip validation for expression values (evaluated during resolution)
	if strings.HasPrefix(strVal, "contrast(") ||
		strings.HasPrefix(strVal, "darken(") ||
		strings.HasPrefix(strVal, "lighten(") ||
		strings.HasPrefix(strVal, "shade(") {
		return nil
	}

	_, err := colors.Parse(strVal)
	return err
}

// validateDimension ensures a dimension value has valid format and units
func validateDimension(value any) error {
	strVal, ok := value.(string)
	if !ok {
		// Allow numeric zero
		if num, ok := value.(float64); ok && num == 0 {
			return nil
		}
		if num, ok := value.(int); ok && num == 0 {
			return nil
		}
		return fmt.Errorf("expected string, got %T", value)
	}

	// Skip validation for reference values
	if strings.Contains(strVal, "{") && strings.Contains(strVal, "}") {
		return nil
	}

	// Skip validation for calc() expressions
	if strings.HasPrefix(strVal, "calc(") || strings.HasPrefix(strVal, "scale(") {
		return nil
	}

	_, err := ParseDimension(strVal)
	return err
}

// validateNumber ensures a value is numeric
func validateNumber(value any) error {
	switch val := value.(type) {
	case float64:
		return nil
	case int:
		return nil
	case string:
		// Skip validation for reference values
		if strings.Contains(val, "{") && strings.Contains(val, "}") {
			return nil
		}
		// Try to parse as number
		if _, err := ParseDimension(val); err == nil {
			return nil
		}
		return fmt.Errorf("expected number, got string: %s", val)
	default:
		return fmt.Errorf("expected number, got %T", value)
	}
}

// validateFontFamily ensures a font family value is valid
func validateFontFamily(value any) error {
	switch val := value.(type) {
	case string:
		if strings.TrimSpace(val) == "" {
			return fmt.Errorf("fontFamily cannot be empty")
		}
		return nil
	case []any:
		if len(val) == 0 {
			return fmt.Errorf("fontFamily array cannot be empty")
		}
		for i, item := range val {
			str, ok := item.(string)
			if !ok {
				return fmt.Errorf("fontFamily array item %d is not a string", i)
			}
			if strings.TrimSpace(str) == "" {
				return fmt.Errorf("fontFamily array item %d is empty", i)
			}
		}
		return nil
	default:
		return fmt.Errorf("expected string or array, got %T", value)
	}
}

// validateEffect ensures an effect value is 0 or 1
func validateEffect(value any) error {
	switch val := value.(type) {
	case float64:
		if val != 0 && val != 1 {
			return fmt.Errorf("effect must be 0 or 1, got %v", val)
		}
		return nil
	case int:
		if val != 0 && val != 1 {
			return fmt.Errorf("effect must be 0 or 1, got %v", val)
		}
		return nil
	case string:
		// Skip validation for reference values
		if strings.Contains(val, "{") && strings.Contains(val, "}") {
			return nil
		}
		if val != "0" && val != "1" {
			return fmt.Errorf("effect must be 0 or 1, got %s", val)
		}
		return nil
	default:
		return fmt.Errorf("expected 0 or 1, got %T", value)
	}
}
