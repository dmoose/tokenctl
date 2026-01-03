package tokens

import (
	"fmt"
)

// Inherit creates a new dictionary by merging base and theme
// Note: This logic assumes simple overwriting for now.
// Real W3C inheritance might involve deep merging specific paths.
func Inherit(base *Dictionary, theme *Dictionary) (*Dictionary, error) {
	// 1. Create a deep copy of base
	result := NewDictionary()
	if err := result.Merge(base); err != nil {
		return nil, fmt.Errorf("failed to copy base dictionary: %w", err)
	}

	// 2. Resolve parent theme if $extends is present
	if extends, ok := theme.Root["$extends"].(string); ok {
		// TODO: Load the parent theme dynamically.
		// This requires access to the Loader or a map of all themes.
		// For now, we assume simple inheritance or manual composition by the caller.
		// Ideally, the caller passes the resolved parent as 'base'.
		_ = extends
	}

	// 3. Merge theme overrides
	if err := result.Merge(theme); err != nil {
		return nil, fmt.Errorf("failed to merge theme dictionary: %w", err)
	}

	// Clean up metadata
	delete(result.Root, "$extends")
	delete(result.Root, "$schema")

	return result, nil
}
