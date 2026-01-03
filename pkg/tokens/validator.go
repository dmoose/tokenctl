package tokens

import (
	"fmt"
	"sort"
	"strings"
)

// ValidationError represents a validation issue
type ValidationError struct {
	Path    string
	Message string
}

func (v ValidationError) Error() string {
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
			errors = append(errors, ValidationError{
				Path:    path,
				Message: err.Error(),
			})
		}
	}

	// 2. Schema Validation (Basic)
	// Walk the tree and check required fields
	walkErrors := v.validateSchema(d.Root, "")
	errors = append(errors, walkErrors...)

	return errors, nil
}

func (v *Validator) validateSchema(node map[string]interface{}, currentPath string) []ValidationError {
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

		childMap, ok := val.(map[string]interface{})
		if !ok {
			childPath := key
			if currentPath != "" {
				childPath = currentPath + "." + key
			}
			errors = append(errors, ValidationError{
				Path:    childPath,
				Message: fmt.Sprintf("expected object, got %T", val),
			})
			continue
		}

		childPath := key
		if currentPath != "" {
			childPath = currentPath + "." + key
		}

		childErrors := v.validateSchema(childMap, childPath)
		errors = append(errors, childErrors...)
	}

	return errors
}
