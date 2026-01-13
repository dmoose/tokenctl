package tokens

import (
	"encoding/json"
	"maps"
)

// ComponentDefinition represents a semantic component
type ComponentDefinition struct {
	Name        string                 `json:"-"`
	Class       string                 `json:"$class"`
	Description string                 `json:"$description,omitempty"`
	Contains    []string               `json:"$contains,omitempty"` // Child components this can contain
	Requires    string                 `json:"$requires,omitempty"` // Parent component this must be inside
	Base        map[string]any `json:"base"`
	Variants    map[string]VariantDef  `json:"variants"`
	Sizes       map[string]VariantDef  `json:"sizes"`
	States      map[string]any `json:"states"` // Reserved for future state enforcement
}

// VariantDef represents a specific variant (primary, outline) or size (sm, lg)
type VariantDef struct {
	Class      string                 `json:"$class"`
	Properties map[string]any `json:"-"` // CSS properties
	States     map[string]State       `json:"-"` // :hover, :focus, etc
}

// State represents a CSS pseudo-class state
type State struct {
	Properties map[string]any
}

// Helper to unmarshal VariantDef handling generic map properties
func (v *VariantDef) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	v.Properties = make(map[string]any)
	v.States = make(map[string]State)

	for key, val := range raw {
		if key == "$class" {
			v.Class = val.(string)
			continue
		}

		// Check if it's a state (starts with & or :)
		if len(key) > 0 && (key[0] == '&' || key[0] == ':') {
			stateProps := make(map[string]any)
			if stateMap, ok := val.(map[string]any); ok {
				maps.Copy(stateProps, stateMap)
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

func walkComponents(node map[string]any, currentPath string, results map[string]ComponentDefinition) error {
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

		// Extract composition metadata (may not be handled by JSON unmarshal)
		if desc, ok := node["$description"].(string); ok {
			comp.Description = desc
		}
		if requires, ok := node["$requires"].(string); ok {
			comp.Requires = requires
		}
		if contains, ok := node["$contains"].([]any); ok {
			comp.Contains = make([]string, 0, len(contains))
			for _, item := range contains {
				if s, ok := item.(string); ok {
					comp.Contains = append(comp.Contains, s)
				}
			}
		}

		results[currentPath] = comp
		return nil
	}

	// Traverse deeper
	for key, val := range node {
		if len(key) > 0 && key[0] == '$' {
			continue
		}
		if child, ok := val.(map[string]any); ok {
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
