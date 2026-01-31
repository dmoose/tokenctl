package main

import (
	"fmt"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

// loadTokens loads the base dictionary and theme dictionaries from dir.
func loadTokens(dir string) (*tokens.Dictionary, map[string]*tokens.Dictionary, error) {
	loader := tokens.NewLoader()

	baseDict, err := loader.LoadBase(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load base tokens: %w", err)
	}

	themes, err := loader.LoadThemes(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load themes: %w", err)
	}

	return baseDict, themes, nil
}

// resolveTokens creates a resolver and resolves all tokens in a dictionary.
func resolveTokens(d *tokens.Dictionary) (map[string]any, error) {
	resolver, err := tokens.NewResolver(d)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize resolver: %w", err)
	}

	resolved, err := resolver.ResolveAll()
	if err != nil {
		return nil, fmt.Errorf("resolution failed: %w", err)
	}

	return resolved, nil
}
