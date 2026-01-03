package tokens

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ParseJSON parses JSON data into a Dictionary
func ParseJSON(r io.Reader) (*Dictionary, error) {
	dec := json.NewDecoder(r)
	var root map[string]interface{}
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}
	return &Dictionary{Root: root}, nil
}

// WriteJSON writes the dictionary to an io.Writer
func (d *Dictionary) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(d.Root)
}

// Loader handles loading tokens from files and directories
type Loader struct {
	Extensions []string
}

// NewLoader creates a default loader
func NewLoader() *Loader {
	return &Loader{
		Extensions: []string{".json", ".tokens.json"},
	}
}

// LoadBase loads all token files EXCEPT those in the themes directory
func (l *Loader) LoadBase(path string) (*Dictionary, error) {
	master := NewDictionary()
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Skip themes directory
		if strings.Contains(filePath, "/themes/") || strings.Contains(filePath, "\\themes\\") {
			return nil
		}

		if l.isTokenFile(filePath) {
			dict, err := l.loadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", filePath, err)
			}
			if err := master.Merge(dict); err != nil {
				return fmt.Errorf("failed to merge %s: %w", filePath, err)
			}
		}
		return nil
	})
	return master, err
}

// LoadThemes scans the themes directory and returns a map of ThemeName -> Dictionary
func (l *Loader) LoadThemes(rootPath string) (map[string]*Dictionary, error) {
	themes := make(map[string]*Dictionary)
	themesPath := filepath.Join(rootPath, "tokens", "themes")

	// Check if themes directory exists
	if _, err := os.Stat(themesPath); os.IsNotExist(err) {
		return themes, nil // Return empty map if no themes dir
	}

	err := filepath.Walk(themesPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !l.isTokenFile(filePath) {
			return nil
		}

		// Theme name is the filename without extension
		filename := filepath.Base(filePath)
		ext := filepath.Ext(filename)
		// Handle .tokens.json double extension
		if strings.HasSuffix(filename, ".tokens.json") {
			ext = ".tokens.json"
		}
		themeName := strings.TrimSuffix(filename, ext)

		dict, err := l.loadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load theme %s: %w", filePath, err)
		}

		// Unwrap root key if it matches theme name
		// Example: dark.json contains { "dark": { ... } }
		if root, ok := dict.Root[themeName]; ok {
			if rootMap, ok := root.(map[string]interface{}); ok {
				// Replace dict root with the unwrapped content
				dict.Root = rootMap
			}
		}

		themes[themeName] = dict
		return nil
	})

	return themes, err
}

func (l *Loader) isTokenFile(path string) bool {
	for _, ext := range l.Extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

func (l *Loader) loadFile(path string) (*Dictionary, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseJSON(f)
}

// Merge merges another dictionary into this one (deep merge)
func (d *Dictionary) Merge(other *Dictionary) error {
	return deepMerge(d.Root, other.Root)
}

func deepMerge(dst, src map[string]interface{}) error {
	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			// Collision handling
			dstMap, dstOk := dstVal.(map[string]interface{})
			srcMap, srcOk := srcVal.(map[string]interface{})

			if dstOk && srcOk {
				// Both are maps, recursive merge
				if err := deepMerge(dstMap, srcMap); err != nil {
					return err
				}
			} else {
				// Type mismatch or value overwrite.
				// For now, we allow overwrite but could warn.
				// In a "strict" mode we might return error.
				dst[key] = srcVal
			}
		} else {
			// No collision, just add
			dst[key] = srcVal
		}
	}
	return nil
}
