package main

import (
	"fmt"
	"io"
	"os"

	"github.com/dmoose/tokenctl/v2/internal/load"
	"github.com/dmoose/tokenctl/v2/pkg/tokens"
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
		_, _ = fmt.Fprintf(w, "%s: dropped input — %s\n", label, f)
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

	srcs := make([]load.Source, len(dirs))
	for i, d := range dirs {
		srcs[i] = load.Dir(d)
	}

	loader := tokens.NewLoader()
	baseDict, themes, err := load.Merge(loader, srcs)
	if err != nil {
		return nil, nil, nil, err
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
