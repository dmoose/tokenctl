package tokens

import (
	"encoding/json"
)

// ComponentDefinition represents a semantic component
type ComponentDefinition struct {
	Name     string                 `json:"-"`
	Class    string                 `json:"$class"`
	Base     map[string]interface{} `json:"base"`
	Variants map[string]VariantDef  `json:"variants"`
	Sizes    map[string]VariantDef  `json:"sizes"`
	States   map[string]interface{} `json:"states"` // Reserved for future state enforcement
}

// VariantDef represents a specific variant (primary, outline) or size (sm, lg)
type VariantDef struct {
	Class      string                 `json:"$class"`
	Properties map[string]interface{} `json:"-"` // CSS properties
	States     map[string]State       `json:"-"` // :hover, :focus, etc
}

// State represents a CSS pseudo-class state
type State struct {
	Properties map[string]interface{}
}

// Helper to unmarshal VariantDef handling generic map properties
func (v *VariantDef) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	v.Properties = make(map[string]interface{})
	v.States = make(map[string]State)

	for key, val := range raw {
		if key == "$class" {
			v.Class = val.(string)
			continue
		}

		// Check if it's a state (starts with & or :)
		if len(key) > 0 && (key[0] == '&' || key[0] == ':') {
			stateProps := make(map[string]interface{})
			if stateMap, ok := val.(map[string]interface{}); ok {
				for pKey, pVal := range stateMap {
					stateProps[pKey] = pVal
				}
			}
			v.States[key] = State{Properties: stateProps}
			continue
		}

		// Otherwise it's a property
		v.Properties[key] = val
	}
	return nil
}

// ExtractComponents finds all tokens with $type: "component"
func (d *Dictionary) ExtractComponents() (map[string]ComponentDefinition, error) {
	components := make(map[string]ComponentDefinition)
	err := walkComponents(d.Root, "", components)
	return components, err
}

func walkComponents(node map[string]interface{}, currentPath string, results map[string]ComponentDefinition) error {
	// Check if this node is a component definition
	if t, ok := node["$type"]; ok && t == "component" {
		// Marshal to JSON and back to struct to use generic unmarshaling
		data, err := json.Marshal(node)
		if err != nil {
			return err
		}
		var comp ComponentDefinition
		if err := json.Unmarshal(data, &comp); err != nil {
			return err
		}
		comp.Name = currentPath
		results[currentPath] = comp
		return nil
	}

	// Traverse deeper
	for key, val := range node {
		if len(key) > 0 && key[0] == '$' {
			continue
		}
		if child, ok := val.(map[string]interface{}); ok {
			childPath := key
			if currentPath != "" {
				childPath = currentPath + "." + key
			}
			if err := walkComponents(child, childPath, results); err != nil {
				return err
			}
		}
	}
	return nil
}
