// tokenctl/pkg/tokens/metadata.go
package tokens

import (
	"strings"
)

// TokenMetadata holds rich metadata for a token
type TokenMetadata struct {
	Path         string      `json:"path"`
	Value        any `json:"value"`
	Type         string      `json:"type,omitempty"`
	Description  string      `json:"description,omitempty"`
	Usage        []string    `json:"usage,omitempty"`
	Avoid        string      `json:"avoid,omitempty"`
	Deprecated   any `json:"deprecated,omitempty"`
	Customizable bool        `json:"customizable,omitempty"`
	SourceFile   string      `json:"source_file,omitempty"`
}

// ExtractMetadata walks the dictionary and extracts rich metadata for all tokens
// Returns a map of token path -> TokenMetadata
func ExtractMetadata(d *Dictionary) map[string]*TokenMetadata {
	result := make(map[string]*TokenMetadata)
	extractMetadataRecursive(d, d.Root, "", "", result)
	return result
}

// extractMetadataRecursive walks the tree and extracts metadata
// inheritedType is the $type from parent groups
func extractMetadataRecursive(d *Dictionary, node map[string]any, currentPath string, inheritedType string, result map[string]*TokenMetadata) {
	// Check for $type at this level
	currentType := inheritedType
	if t, ok := node["$type"].(string); ok {
		currentType = t
	}

	if IsToken(node) {
		meta := &TokenMetadata{
			Path:  currentPath,
			Value: node["$value"],
			Type:  currentType,
		}

		// Extract description
		if desc, ok := node["$description"].(string); ok {
			meta.Description = desc
		}

		// Extract usage (can be string or array)
		if usage, ok := node["$usage"]; ok {
			switch u := usage.(type) {
			case string:
				meta.Usage = []string{u}
			case []any:
				for _, item := range u {
					if s, ok := item.(string); ok {
						meta.Usage = append(meta.Usage, s)
					}
				}
			case []string:
				meta.Usage = u
			}
		}

		// Extract avoid
		if avoid, ok := node["$avoid"].(string); ok {
			meta.Avoid = avoid
		}

		// Extract deprecated
		if deprecated, ok := node["$deprecated"]; ok {
			meta.Deprecated = deprecated
		}

		// Extract customizable flag
		if customizable, ok := node["$customizable"].(bool); ok {
			meta.Customizable = customizable
		}

		// Extract source file if tracked
		if sourceFile, ok := d.SourceFiles[currentPath]; ok {
			meta.SourceFile = sourceFile
		}

		result[currentPath] = meta
		return
	}

	// Recurse into children
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

		extractMetadataRecursive(d, childMap, childPath, currentType, result)
	}
}
