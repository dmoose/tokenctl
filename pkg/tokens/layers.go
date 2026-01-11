// tokenctl/pkg/tokens/layers.go
package tokens

import (
	"fmt"
	"strings"
)

// Layer represents a design system layer
type Layer string

const (
	LayerBrand     Layer = "brand"
	LayerSemantic  Layer = "semantic"
	LayerComponent Layer = "component"
)

// LayerOrder defines the reference hierarchy (lower index can't reference higher)
var LayerOrder = map[Layer]int{
	LayerBrand:     0,
	LayerSemantic:  1,
	LayerComponent: 2,
}

// CanReference returns true if fromLayer is allowed to reference toLayer
// Rules:
// - brand: can only use raw values (no references)
// - semantic: can reference brand
// - component: can reference semantic (and transitively brand)
func CanReference(fromLayer, toLayer Layer) bool {
	fromOrder, fromOk := LayerOrder[fromLayer]
	toOrder, toOk := LayerOrder[toLayer]

	// Unknown layers are permissive
	if !fromOk || !toOk {
		return true
	}

	// Can reference same or lower layer
	return fromOrder >= toOrder
}

// LayerViolation represents a layer reference violation
type LayerViolation struct {
	TokenPath   string
	TokenLayer  Layer
	RefPath     string
	RefLayer    Layer
	SourceFile  string
}

func (v LayerViolation) Error() string {
	msg := fmt.Sprintf("%s [%s] cannot reference %s [%s]: layer violation",
		v.TokenPath, v.TokenLayer, v.RefPath, v.RefLayer)
	if v.SourceFile != "" {
		msg = fmt.Sprintf("%s [%s] [%s] cannot reference %s [%s]: layer violation",
			v.TokenPath, v.TokenLayer, v.SourceFile, v.RefPath, v.RefLayer)
	}
	return msg
}

// LayerValidator validates layer reference rules
type LayerValidator struct {
	tokenLayers map[string]Layer // token path -> layer
}

// NewLayerValidator creates a validator from a dictionary
func NewLayerValidator(d *Dictionary) *LayerValidator {
	v := &LayerValidator{
		tokenLayers: make(map[string]Layer),
	}
	v.extractLayers(d.Root, "", "")
	return v
}

// extractLayers walks the dictionary and builds the layer map
func (v *LayerValidator) extractLayers(node map[string]interface{}, currentPath string, inheritedLayer string) {
	// Check for $layer at this level
	currentLayer := inheritedLayer
	if layer, ok := node["$layer"].(string); ok {
		currentLayer = layer
	}

	if IsToken(node) {
		if currentLayer != "" {
			v.tokenLayers[currentPath] = Layer(currentLayer)
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

		v.extractLayers(childMap, childPath, currentLayer)
	}
}

// GetLayer returns the layer for a token path
func (v *LayerValidator) GetLayer(path string) Layer {
	return v.tokenLayers[path]
}

// ValidateReferences checks all token references against layer rules
func (v *LayerValidator) ValidateReferences(d *Dictionary) []LayerViolation {
	var violations []LayerViolation

	// Walk all tokens and check their references
	v.walkAndValidate(d, d.Root, "", &violations)

	return violations
}

// walkAndValidate recursively validates layer references
func (v *LayerValidator) walkAndValidate(d *Dictionary, node map[string]interface{}, currentPath string, violations *[]LayerViolation) {
	if IsToken(node) {
		value, ok := node["$value"].(string)
		if !ok {
			return
		}

		fromLayer := v.tokenLayers[currentPath]
		if fromLayer == "" {
			return // No layer assigned, skip validation
		}

		// Find all references in the value
		refs := refRegex.FindAllStringSubmatch(value, -1)
		for _, ref := range refs {
			refPath := ref[1]
			toLayer := v.tokenLayers[refPath]

			if toLayer == "" {
				continue // Referenced token has no layer, skip
			}

			if !CanReference(fromLayer, toLayer) {
				violation := LayerViolation{
					TokenPath:  currentPath,
					TokenLayer: fromLayer,
					RefPath:    refPath,
					RefLayer:   toLayer,
				}
				if sourceFile, ok := d.SourceFiles[currentPath]; ok {
					violation.SourceFile = sourceFile
				}
				*violations = append(*violations, violation)
			}
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

		v.walkAndValidate(d, childMap, childPath, violations)
	}
}
