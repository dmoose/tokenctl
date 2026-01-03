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
	Root map[string]interface{}
}

// NewDictionary creates an empty dictionary
func NewDictionary() *Dictionary {
	return &Dictionary{
		Root: make(map[string]interface{}),
	}
}

// IsToken checks if a map node matches the Token signature
func IsToken(node map[string]interface{}) bool {
	_, ok := node["$value"]
	return ok
}
