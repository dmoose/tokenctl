package main

import (
	"fmt"
	"io"
	"os"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

// reportUnknownKeys writes one line per piece of input the build will not
// consume. Silence here is what let a site ship five classes with zero
// styling: the tokens parsed, the build succeeded, and the misnamed block
// went straight to the floor. Returns an error when strict is set so the
// build can fail instead of merely complaining.
func reportUnknownKeys(w io.Writer, strict bool, findings []tokens.Finding) error {
	if len(findings) == 0 {
		return nil
	}

	label := "Warning"
	if strict {
		label = "Error"
	}
	for _, f := range findings {
		fmt.Fprintf(w, "%s: dropped input — %s\n", label, f)
	}
	if strict {
		return fmt.Errorf("%d unconsumed key(s) in token input (--strict-unknown-keys)", len(findings))
	}
	return nil
}

// auditLoaded reports the findings the loader collected, to stderr.
func auditLoaded(strict bool, findings []tokens.Finding) error {
	return reportUnknownKeys(os.Stderr, strict, findings)
}

// loadTokens loads and merges base dictionaries and theme dictionaries from one
// or more directories. When multiple directories are provided, they are merged
// left-to-right: later directories extend or override earlier ones.
//
// The returned findings list every key the loader read but the build will
// not consume; callers decide whether that warns or fails.
func loadTokens(dirs ...string) (*tokens.Dictionary, map[string]*tokens.Dictionary, []tokens.Finding, error) {
	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	loader := tokens.NewLoader()

	// Load first directory as the master base + themes
	baseDict, err := loader.LoadBase(dirs[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load base tokens from %s: %w", dirs[0], err)
	}

	themes, err := loader.LoadThemes(dirs[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load themes from %s: %w", dirs[0], err)
	}

	// Merge subsequent directories
	for _, dir := range dirs[1:] {
		extBase, err := loader.LoadBase(dir)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to load base tokens from %s: %w", dir, err)
		}
		if err := baseDict.MergeWithPath(extBase, loader.WarnConflicts); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to merge base tokens from %s: %w", dir, err)
		}

		extThemes, err := loader.LoadThemes(dir)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to load themes from %s: %w", dir, err)
		}
		for name, extTheme := range extThemes {
			if existing, ok := themes[name]; ok {
				if err := existing.MergeWithPath(extTheme, loader.WarnConflicts); err != nil {
					return nil, nil, nil, fmt.Errorf("failed to merge theme %s from %s: %w", name, dir, err)
				}
			} else {
				themes[name] = extTheme
			}
		}
	}

	return baseDict, themes, loader.Findings, nil
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
