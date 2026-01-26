package tokens

import "maps"

// W3C Design Tokens 2025.10 Specification Types

// Dictionary represents a collection of tokens or nested groups.
// W3C spec allows mixed content (groups and tokens in the same object)
// but defines tokens by the presence of "$value".
// We use map[string]any to represent the raw structure
// and helper methods to traverse it.
type Dictionary struct {
	Root        map[string]any
	SourceFiles map[string]string // Maps token path to source file
}

// NewDictionary creates an empty dictionary
func NewDictionary() *Dictionary {
	return &Dictionary{
		Root:        make(map[string]any),
		SourceFiles: make(map[string]string),
	}
}

// IsToken checks if a map node matches the Token signature
func IsToken(node map[string]any) bool {
	_, ok := node["$value"]
	return ok
}

// DeepCopy creates a deep copy of a Dictionary
func (d *Dictionary) DeepCopy() *Dictionary {
	copiedSourceFiles := make(map[string]string, len(d.SourceFiles))
	maps.Copy(copiedSourceFiles, d.SourceFiles)
	return &Dictionary{
		Root:        deepCopyMap(d.Root),
		SourceFiles: copiedSourceFiles,
	}
}

// deepCopyMap recursively copies a map[string]any
func deepCopyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = deepCopyValue(v)
	}
	return dst
}

// deepCopyValue recursively copies any any value
func deepCopyValue(src any) any {
	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case map[string]any:
		return deepCopyMap(v)
	case []any:
		dst := make([]any, len(v))
		for i, item := range v {
			dst[i] = deepCopyValue(item)
		}
		return dst
	default:
		// Primitives (string, int, float, bool) are copied by value
		return v
	}
}
