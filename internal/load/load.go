// tokenctl/internal/load/load.go
package load

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/dmoose/tokenctl/v2/pkg/tokens"
)

// Source is one token tree.
type Source struct {
	Name   string
	base   func(*tokens.Loader) (*tokens.Dictionary, error)
	themes func(*tokens.Loader) (map[string]*tokens.Dictionary, error)
}

// Dir is a token tree on disk.
func Dir(dir string) Source {
	return Source{
		Name:   dir,
		base:   func(l *tokens.Loader) (*tokens.Dictionary, error) { return l.LoadBase(dir) },
		themes: func(l *tokens.Loader) (map[string]*tokens.Dictionary, error) { return l.LoadThemes(dir) },
	}
}

// FS is a token tree rooted at the top of fsys.
func FS(name string, fsys fs.FS) Source {
	return Source{
		Name:   name,
		base:   func(l *tokens.Loader) (*tokens.Dictionary, error) { return l.LoadBaseFS(fsys, ".") },
		themes: func(l *tokens.Loader) (map[string]*tokens.Dictionary, error) { return l.LoadThemesFS(fsys, ".") },
	}
}

// Merge loads the sources and merges them left to right: the first is the
// base, later ones extend or override it.
func Merge(l *tokens.Loader, srcs []Source) (*tokens.Dictionary, map[string]*tokens.Dictionary, error) {
	if len(srcs) == 0 {
		return nil, nil, errors.New("no token sources")
	}

	baseDict, err := srcs[0].base(l)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load base tokens from %s: %w", srcs[0].Name, err)
	}

	themes, err := srcs[0].themes(l)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load themes from %s: %w", srcs[0].Name, err)
	}

	for _, src := range srcs[1:] {
		extBase, err := src.base(l)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load base tokens from %s: %w", src.Name, err)
		}
		if err := l.Merge(baseDict, extBase); err != nil {
			return nil, nil, fmt.Errorf("failed to merge base tokens from %s: %w", src.Name, err)
		}

		extThemes, err := src.themes(l)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load themes from %s: %w", src.Name, err)
		}
		for name, extTheme := range extThemes {
			if existing, ok := themes[name]; ok {
				if err := l.Merge(existing, extTheme); err != nil {
					return nil, nil, fmt.Errorf("failed to merge theme %s from %s: %w", name, src.Name, err)
				}
			} else {
				themes[name] = extTheme
			}
		}
	}

	return baseDict, themes, nil
}
