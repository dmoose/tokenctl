package tokens

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	return &Dictionary{
		Root:        root,
		SourceFiles: make(map[string]string),
	}, nil
}

// WriteJSON writes the dictionary to an io.Writer
func (d *Dictionary) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(d.Root)
}

// Loader handles loading tokens from files and directories
type Loader struct {
	Extensions    []string
	WarnConflicts bool // Emit warnings when merge conflicts occur
}

// NewLoader creates a default loader with conflict warnings enabled
func NewLoader() *Loader {
	return &Loader{
		Extensions:    []string{".json", ".tokens.json"},
		WarnConflicts: true,
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
			if err := master.MergeWithPath(dict, l.WarnConflicts, filePath); err != nil {
				return fmt.Errorf("failed to merge %s: %w", filePath, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Expand any $scale definitions into individual tokens
	if err := ExpandScales(master); err != nil {
		return nil, fmt.Errorf("failed to expand scales: %w", err)
	}

	return master, nil
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

		// Track source file for all tokens in this theme
		l.annotateSourceFile(dict, "", filePath)

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
	if err != nil {
		return nil, err
	}

	// Expand scales in each theme
	for name, dict := range themes {
		if err := ExpandScales(dict); err != nil {
			return nil, fmt.Errorf("failed to expand scales in theme %s: %w", name, err)
		}
	}

	return themes, nil
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
	dict, err := ParseJSON(f)
	if err != nil {
		return nil, err
	}

	// Annotate all tokens with source file
	l.annotateSourceFile(dict, "", path)
	return dict, nil
}

// annotateSourceFile recursively marks all tokens in the dictionary with their source file
func (l *Loader) annotateSourceFile(dict *Dictionary, currentPath, sourceFile string) {
	annotateSourceFileRecursive(dict, dict.Root, currentPath, sourceFile)
}

func annotateSourceFileRecursive(dict *Dictionary, node map[string]interface{}, currentPath, sourceFile string) {
	if IsToken(node) {
		if currentPath != "" {
			dict.SourceFiles[currentPath] = sourceFile
		}
		return
	}

	for key, val := range node {
		if strings.HasPrefix(key, "$") {
			continue
		}

		childMap, ok := val.(map[string]interface{})
		if !ok {
			continue
		}

		childPath := key
		if currentPath != "" {
			childPath = currentPath + "." + key
		}

		annotateSourceFileRecursive(dict, childMap, childPath, sourceFile)
	}
}

// Merge merges another dictionary into this one (deep merge)
func (d *Dictionary) Merge(other *Dictionary) error {
	if err := deepMerge(d.Root, other.Root, ""); err != nil {
		return err
	}
	// Merge source file mappings
	for path, file := range other.SourceFiles {
		d.SourceFiles[path] = file
	}
	return nil
}

// MergeWithPath is like Merge but tracks the current path for better error messages
func (d *Dictionary) MergeWithPath(other *Dictionary, warnConflicts bool, sourceFile string) error {
	if err := deepMergeWithWarnings(d.Root, other.Root, "", warnConflicts); err != nil {
		return err
	}
	// Merge source file mappings, preferring the new source file for conflicts
	for path, file := range other.SourceFiles {
		d.SourceFiles[path] = file
	}
	return nil
}

func deepMerge(dst, src map[string]interface{}, currentPath string) error {
	return deepMergeWithWarnings(dst, src, currentPath, false)
}

func deepMergeWithWarnings(dst, src map[string]interface{}, currentPath string, warnConflicts bool) error {
	for key, srcVal := range src {
		// Build path for error messages
		path := key
		if currentPath != "" {
			path = currentPath + "." + key
		}

		if dstVal, ok := dst[key]; ok {
			// Collision handling
			dstMap, dstOk := dstVal.(map[string]interface{})
			srcMap, srcOk := srcVal.(map[string]interface{})

			if dstOk && srcOk {
				// Both are maps, check if either is a token before recursing
				isDstToken := IsToken(dstMap)
				isSrcToken := IsToken(srcMap)

				if isDstToken || isSrcToken {
					// One or both are tokens - this is an overwrite
					if warnConflicts {
						log.Printf("Warning: Token '%s' redefined (overwriting)\n", path)
					}
					dst[key] = srcVal
				} else {
					// Both are groups, recursive merge
					if err := deepMergeWithWarnings(dstMap, srcMap, path, warnConflicts); err != nil {
						return err
					}
				}
			} else {
				// Type mismatch or value overwrite
				if warnConflicts {
					log.Printf("Warning: Token '%s' redefined (overwriting %T with %T)\n", path, dstVal, srcVal)
				}
				dst[key] = srcVal
			}
		} else {
			// No collision, just add
			dst[key] = srcVal
		}
	}
	return nil
}
