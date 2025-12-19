package tokens

// W3C Design Tokens 2025.10 Specification Types

// Token represents a single design token
type Token struct {
	Value       interface{}            `json:"$value"`
	Type        string                 `json:"$type,omitempty"`
	Description string                 `json:"$description,omitempty"`
	Extensions  map[string]interface{} `json:"$extensions,omitempty"`
	// Additional metadata
	Deprecated interface{} `json:"$deprecated,omitempty"` // bool or string reason
}

// TokenGroup represents a collection of tokens or nested groups
// W3C spec allows mixed content (groups and tokens in the same object)
// but defines tokens by the presence of "$value".
// We use map[string]interface{} to represent the raw structure
// and helper methods to traverse it.
type Dictionary struct {
	Root        map[string]interface{}
	SourceFiles map[string]string // Maps token path to source file
}

// NewDictionary creates an empty dictionary
func NewDictionary() *Dictionary {
	return &Dictionary{
		Root:        make(map[string]interface{}),
		SourceFiles: make(map[string]string),
	}
}

// IsToken checks if a map node matches the Token signature
func IsToken(node map[string]interface{}) bool {
	_, ok := node["$value"]
	return ok
}

// DeepCopy creates a deep copy of a Dictionary
func (d *Dictionary) DeepCopy() *Dictionary {
	copiedSourceFiles := make(map[string]string, len(d.SourceFiles))
	for k, v := range d.SourceFiles {
		copiedSourceFiles[k] = v
	}
	return &Dictionary{
		Root:        deepCopyMap(d.Root),
		SourceFiles: copiedSourceFiles,
	}
}

// deepCopyMap recursively copies a map[string]interface{}
func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}

	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = deepCopyValue(v)
	}
	return dst
}

// deepCopyValue recursively copies any interface{} value
func deepCopyValue(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case map[string]interface{}:
		return deepCopyMap(v)
	case []interface{}:
		dst := make([]interface{}, len(v))
		for i, item := range v {
			dst[i] = deepCopyValue(item)
		}
		return dst
	default:
		// Primitives (string, int, float, bool) are copied by value
		return v
	}
}
