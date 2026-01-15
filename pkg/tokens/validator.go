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

// Validator checks token dictionaries for compliance
type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks the dictionary for:
// 1. Broken references (using Resolver)
// 2. Schema compliance (basic checks)
// 3. Type-specific validation (color, dimension, number, effect)
// 4. Constraint validation ($min/$max)
func (v *Validator) Validate(d *Dictionary) ([]ValidationError, error) {
	var errors []ValidationError

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
			// Is it a cycle or missing ref?
			verr := ValidationError{
				Path:    path,
				Message: err.Error(),
			}
			// Add source file if available
			if sourceFile, ok := d.SourceFiles[path]; ok {
				verr.SourceFile = sourceFile
			}
			errors = append(errors, verr)
		}
	}

	// 2. Schema Validation (Basic)
	// Walk the tree and check required fields
	walkErrors := v.validateSchema(d, d.Root, "")
	errors = append(errors, walkErrors...)

	// 3. Type-specific and constraint validation
	typeErrors := v.validateTypes(d, d.Root, "")
	errors = append(errors, typeErrors...)

	return errors, nil
}

func (v *Validator) validateSchema(dict *Dictionary, node map[string]any, currentPath string) []ValidationError {
	var errors []ValidationError

	if IsToken(node) {
		// Check for valid $value (cannot be nil or empty string unless explicitly allowed?)
		// W3C spec says $value is required (implied by IsToken check)
		// We could check $type valid values here if we wanted strictly typed validation
		return errors
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
			// Add source file if available
			if sourceFile, ok := dict.SourceFiles[childPath]; ok {
				verr.SourceFile = sourceFile
			}
			errors = append(errors, verr)
			continue
		}

		childPath := key
		if currentPath != "" {
			childPath = currentPath + "." + key
		}

		childErrors := v.validateSchema(dict, childMap, childPath)
		errors = append(errors, childErrors...)
	}

	return errors
}

// validateTypes performs type-specific validation including constraints
// inheritedType is the $type inherited from parent groups
func (v *Validator) validateTypes(dict *Dictionary, node map[string]any, currentPath string) []ValidationError {
	return v.validateTypesWithInheritance(dict, node, currentPath, "")
}

// validateTypesWithInheritance performs type validation with $type inheritance from parent groups
func (v *Validator) validateTypesWithInheritance(dict *Dictionary, node map[string]any, currentPath string, inheritedType string) []ValidationError {
	var errors []ValidationError

	// Check for $type at this level to pass to children
	currentType := inheritedType
	if t, ok := node["$type"].(string); ok {
		currentType = t
	}

	if IsToken(node) {
		// Get token type (from token itself or inherited)
		tokenType := v.getTokenTypeWithInheritance(node, currentPath, currentType)
		value := node["$value"]

		// Type-specific validation
		switch tokenType {
		case "color":
			if err := v.validateColorFormat(value); err != nil {
				verr := ValidationError{
					Path:    currentPath,
					Message: fmt.Sprintf("invalid color: %s", err.Error()),
				}
				if sourceFile, ok := dict.SourceFiles[currentPath]; ok {
					verr.SourceFile = sourceFile
				}
				errors = append(errors, verr)
			}

		case "dimension":
			if err := v.validateDimension(value); err != nil {
				verr := ValidationError{
					Path:    currentPath,
					Message: fmt.Sprintf("invalid dimension: %s", err.Error()),
				}
				if sourceFile, ok := dict.SourceFiles[currentPath]; ok {
					verr.SourceFile = sourceFile
				}
				errors = append(errors, verr)
			}

		case "number":
			if err := v.validateNumber(value); err != nil {
				verr := ValidationError{
					Path:    currentPath,
					Message: fmt.Sprintf("invalid number: %s", err.Error()),
				}
				if sourceFile, ok := dict.SourceFiles[currentPath]; ok {
					verr.SourceFile = sourceFile
				}
				errors = append(errors, verr)
			}

		case "fontFamily":
			if err := v.validateFontFamily(value); err != nil {
				verr := ValidationError{
					Path:    currentPath,
					Message: fmt.Sprintf("invalid fontFamily: %s", err.Error()),
				}
				if sourceFile, ok := dict.SourceFiles[currentPath]; ok {
					verr.SourceFile = sourceFile
				}
				errors = append(errors, verr)
			}

		case "effect":
			if err := v.validateEffect(value); err != nil {
				verr := ValidationError{
					Path:    currentPath,
					Message: fmt.Sprintf("invalid effect: %s", err.Error()),
				}
				if sourceFile, ok := dict.SourceFiles[currentPath]; ok {
					verr.SourceFile = sourceFile
				}
				errors = append(errors, verr)
			}
		}

		// Constraint validation ($min/$max)
		constraintErrors := v.validateConstraints(dict, currentPath, node, value)
		errors = append(errors, constraintErrors...)

		return errors
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

		childErrors := v.validateTypesWithInheritance(dict, childMap, childPath, currentType)
		errors = append(errors, childErrors...)
	}

	return errors
}

// getTokenTypeWithInheritance extracts the $type from a token, using inherited type if not set
func (v *Validator) getTokenTypeWithInheritance(token map[string]any, _ string, inheritedType string) string {
	if t, ok := token["$type"].(string); ok {
		return t
	}
	return inheritedType
}

// validateConstraints checks $min/$max constraints on a token value
func (v *Validator) validateConstraints(dict *Dictionary, path string, token map[string]any, value any) []ValidationError {
	var errors []ValidationError

	constraint, err := ParseConstraints(token)
	if err != nil {
		verr := ValidationError{
			Path:    path,
			Message: fmt.Sprintf("constraint error: %s", err.Error()),
		}
		if sourceFile, ok := dict.SourceFiles[path]; ok {
			verr.SourceFile = sourceFile
		}
		errors = append(errors, verr)
		return errors
	}

	if constraint == nil {
		return errors
	}

	// Skip constraint checking for reference values (they'll be checked after resolution)
	if strVal, ok := value.(string); ok {
		if strings.Contains(strVal, "{") && strings.Contains(strVal, "}") {
			return errors
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
		errors = append(errors, verr)
	}

	return errors
}

// validateColorFormat ensures a color value is a valid CSS color
func (v *Validator) validateColorFormat(value any) error {
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	// Skip validation for reference values
	if strings.Contains(strVal, "{") && strings.Contains(strVal, "}") {
		return nil
	}

	// Skip validation for expression values (contrast, darken, lighten, etc.)
	if strings.HasPrefix(strVal, "contrast(") ||
		strings.HasPrefix(strVal, "darken(") ||
		strings.HasPrefix(strVal, "lighten(") {
		return nil
	}

	_, err := colors.Parse(strVal)
	return err
}

// validateDimension ensures a dimension value has valid format and units
func (v *Validator) validateDimension(value any) error {
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
func (v *Validator) validateNumber(value any) error {
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
func (v *Validator) validateFontFamily(value any) error {
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
func (v *Validator) validateEffect(value any) error {
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
