package tokens

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ParseJSON parses JSON data into a Dictionary
func ParseJSON(r io.Reader) (*Dictionary, error) {
	dec := json.NewDecoder(r)
	var root map[string]any
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

	// OnConflict receives conflict warnings when set; otherwise they go to stderr.
	OnConflict func(msg string)

	// Findings accumulates keys the build will not consume, one entry per
	// offending key, in load order. Collected per file rather than after
	// the merge so each finding can name the file it came from — the
	// merged dictionary has no source record for component nodes, and a
	// warning without a filename is not actionable across 170-odd token
	// files.
	Findings []Finding
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
	if _, err := os.Lstat(path); err != nil {
		return nil, err
	}
	fsys, root, prefix := dirFS(path)
	return l.loadBase(fsys, root, prefix)
}

// LoadBaseFS is LoadBase over fsys, walking from root.
func (l *Loader) LoadBaseFS(fsys fs.FS, root string) (*Dictionary, error) {
	return l.loadBase(fsys, root, "")
}

func (l *Loader) loadBase(fsys fs.FS, root, prefix string) (*Dictionary, error) {
	master := NewDictionary()
	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Skip themes directory
		filePath := displayPath(prefix, p)
		if strings.Contains(filePath, "/themes/") || strings.Contains(filePath, "\\themes\\") ||
			strings.Contains("/"+p, "/themes/") {
			return nil
		}

		if l.isTokenFile(p) {
			dict, err := l.loadFile(fsys, p, filePath)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", filePath, err)
			}
			if err := l.Merge(master, dict); err != nil {
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
	return l.loadThemes(os.DirFS(rootPath), ".", rootPath)
}

// LoadThemesFS is LoadThemes over fsys, looking for themes under root.
func (l *Loader) LoadThemesFS(fsys fs.FS, root string) (map[string]*Dictionary, error) {
	return l.loadThemes(fsys, root, "")
}

func (l *Loader) loadThemes(fsys fs.FS, root, prefix string) (map[string]*Dictionary, error) {
	themes := make(map[string]*Dictionary)

	// Support both layouts:
	// 1. root/themes/ - when user passes tokens directory directly (tokenctl build tokens)
	// 2. root/tokens/themes/ - when user passes project root (tokenctl build .)
	themesPath := path.Join(root, "themes")
	if _, err := fs.Stat(fsys, themesPath); errors.Is(err, fs.ErrNotExist) {
		themesPath = path.Join(root, "tokens", "themes")
	}

	// Check if themes directory exists
	if _, err := fs.Stat(fsys, themesPath); errors.Is(err, fs.ErrNotExist) {
		return themes, nil // Return empty map if no themes dir
	}

	err := fs.WalkDir(fsys, themesPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !l.isTokenFile(p) {
			return nil
		}
		filePath := displayPath(prefix, p)

		// Theme name is the filename without extension
		filename := path.Base(p)
		ext := path.Ext(filename)
		// Handle .tokens.json double extension
		if strings.HasSuffix(filename, ".tokens.json") {
			ext = ".tokens.json"
		}
		themeName := strings.TrimSuffix(filename, ext)

		dict, err := l.loadFile(fsys, p, filePath)
		if err != nil {
			return fmt.Errorf("failed to load theme %s: %w", filePath, err)
		}

		// Track source file for all tokens in this theme
		l.annotateSourceFile(dict, "", filePath)

		// Unwrap root key if it matches theme name
		// Example: dark.json contains { "dark": { ... } }
		if root, ok := dict.Root[themeName]; ok {
			if rootMap, ok := root.(map[string]any); ok {
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

// Merge merges src into dst, reporting conflicts as the loader is configured to.
func (l *Loader) Merge(dst, src *Dictionary) error {
	return dst.merge(src, l.warn())
}

func (l *Loader) warn() func(string) {
	if !l.WarnConflicts {
		return nil
	}
	if l.OnConflict != nil {
		return l.OnConflict
	}
	return warnStderr
}

func warnStderr(msg string) {
	fmt.Fprintf(os.Stderr, "Warning: %s\n", msg)
}

// dirFS returns an fs.FS, walk root and display prefix for a directory or file path.
func dirFS(p string) (fs.FS, string, string) {
	if info, err := os.Stat(p); err == nil && !info.IsDir() {
		return os.DirFS(filepath.Dir(p)), filepath.Base(p), filepath.Dir(p)
	}
	return os.DirFS(p), ".", p
}

func displayPath(prefix, p string) string {
	if prefix == "" {
		return p
	}
	return filepath.Join(prefix, filepath.FromSlash(p))
}

func (l *Loader) isTokenFile(path string) bool {
	for _, ext := range l.Extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

func (l *Loader) loadFile(fsys fs.FS, name, path string) (*Dictionary, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	dict, err := ParseJSON(f)
	if err != nil {
		return nil, err
	}

	// Audit before merging: after the merge, paths from different files
	// are indistinguishable and component nodes carry no source record.
	for _, finding := range AuditUnknownKeys(dict) {
		finding.SourceFile = path
		l.Findings = append(l.Findings, finding)
	}

	// Annotate all tokens with source file
	l.annotateSourceFile(dict, "", path)
	return dict, nil
}

// annotateSourceFile recursively marks all tokens in the dictionary with their source file
func (l *Loader) annotateSourceFile(dict *Dictionary, currentPath, sourceFile string) {
	annotateSourceFileRecursive(dict, dict.Root, currentPath, sourceFile)
}

func annotateSourceFileRecursive(dict *Dictionary, node map[string]any, currentPath, sourceFile string) {
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

		childMap, ok := val.(map[string]any)
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
	maps.Copy(d.SourceFiles, other.SourceFiles)
	return nil
}

// MergeWithPath is like Merge but allows controlling conflict warnings
func (d *Dictionary) MergeWithPath(other *Dictionary, warnConflicts bool) error {
	var warn func(string)
	if warnConflicts {
		warn = warnStderr
	}
	return d.merge(other, warn)
}

func (d *Dictionary) merge(other *Dictionary, warn func(string)) error {
	if err := deepMergeWithWarnings(d.Root, other.Root, "", warn); err != nil {
		return err
	}
	// Merge source file mappings, preferring the new source file for conflicts
	maps.Copy(d.SourceFiles, other.SourceFiles)
	return nil
}

// DetectDefaultTheme scans theme dictionaries for "$default": true metadata.
// Returns the name of the first theme that declares itself as default,
// or "light" as a fallback if no theme declares $default.
func DetectDefaultTheme(themes map[string]*Dictionary) string {
	for name, dict := range themes {
		if def, ok := dict.Root["$default"]; ok {
			if b, ok := def.(bool); ok && b {
				return name
			}
		}
	}
	return "light"
}

func deepMerge(dst, src map[string]any, currentPath string) error {
	return deepMergeWithWarnings(dst, src, currentPath, nil)
}

func deepMergeWithWarnings(dst, src map[string]any, currentPath string, warn func(string)) error {
	for key, srcVal := range src {
		// Build path for error messages
		path := key
		if currentPath != "" {
			path = currentPath + "." + key
		}

		// Skip warning for metadata keys ($ prefix) - these are expected to be redefined across files
		isMetadataKey := strings.HasPrefix(key, "$")

		if dstVal, ok := dst[key]; ok {
			// Collision handling
			dstMap, dstOk := dstVal.(map[string]any)
			srcMap, srcOk := srcVal.(map[string]any)

			if dstOk && srcOk {
				// Both are maps, check if either is a token before recursing
				isDstToken := IsToken(dstMap)
				isSrcToken := IsToken(srcMap)

				if isDstToken || isSrcToken {
					// One or both are tokens - this is an overwrite
					if warn != nil && !isMetadataKey {
						warn(fmt.Sprintf("Token '%s' redefined (overwriting)", path))
					}
					dst[key] = srcVal
				} else {
					// Both are groups, recursive merge
					if err := deepMergeWithWarnings(dstMap, srcMap, path, warn); err != nil {
						return err
					}
				}
			} else {
				// Type mismatch or value overwrite
				if warn != nil && !isMetadataKey {
					warn(fmt.Sprintf("Token '%s' redefined (overwriting %T with %T)", path, dstVal, srcVal))
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
