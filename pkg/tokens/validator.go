package tokens

import (
	"fmt"
	"sort"
	"strings"
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

	return errors, nil
}

func (v *Validator) validateSchema(dict *Dictionary, node map[string]interface{}, currentPath string) []ValidationError {
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
