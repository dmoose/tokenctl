package tokens

import (
	"fmt"
)

// Inherit creates a new dictionary by merging base and theme
// Supports $extends for theme inheritance with circular dependency detection
func Inherit(base *Dictionary, theme *Dictionary) (*Dictionary, error) {
	// 1. Create a deep copy of base
	result := base.DeepCopy()

	// 2. Merge theme overrides
	if err := result.Merge(theme); err != nil {
		return nil, fmt.Errorf("failed to merge theme dictionary: %w", err)
	}

	// Clean up metadata
	delete(result.Root, "$extends")
	delete(result.Root, "$schema")

	return result, nil
}

// ResolveThemeInheritance resolves the full inheritance chain for all themes
// Returns a map of theme names to their fully resolved dictionaries
func ResolveThemeInheritance(base *Dictionary, themes map[string]*Dictionary) (map[string]*Dictionary, error) {
	resolved := make(map[string]*Dictionary)
	resolving := make(map[string]bool) // Track currently resolving themes for cycle detection

	// Helper function to resolve a single theme recursively
	var resolveTheme func(name string) (*Dictionary, error)
	resolveTheme = func(name string) (*Dictionary, error) {
		// Check if already resolved
		if result, ok := resolved[name]; ok {
			return result, nil
		}

		// Check for circular dependency
		if resolving[name] {
			return nil, fmt.Errorf("circular theme inheritance detected: theme '%s' extends itself", name)
		}

		// Get theme dictionary
		themeDict, ok := themes[name]
		if !ok {
			return nil, fmt.Errorf("theme '%s' not found", name)
		}

		// Mark as currently resolving
		resolving[name] = true
		defer delete(resolving, name)

		// Check if this theme extends another
		if extendsVal, ok := themeDict.Root["$extends"]; ok {
			parentName, ok := extendsVal.(string)
			if !ok {
				return nil, fmt.Errorf("theme '%s' has invalid $extends value (expected string, got %T)", name, extendsVal)
			}

			// Recursively resolve parent theme
			parentResolved, err := resolveTheme(parentName)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve parent theme '%s' for '%s': %w", parentName, name, err)
			}

			// Inherit from parent instead of base
			result, err := Inherit(parentResolved, themeDict)
			if err != nil {
				return nil, fmt.Errorf("failed to inherit theme '%s' from '%s': %w", name, parentName, err)
			}

			resolved[name] = result
			return result, nil
		}

		// No parent, inherit from base
		result, err := Inherit(base, themeDict)
		if err != nil {
			return nil, fmt.Errorf("failed to inherit theme '%s' from base: %w", name, err)
		}

		resolved[name] = result
		return result, nil
	}

	// Resolve all themes
	for name := range themes {
		if _, err := resolveTheme(name); err != nil {
			return nil, err
		}
	}

	return resolved, nil
}
