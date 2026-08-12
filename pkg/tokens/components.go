package tokens

import (
	"encoding/json"
	"maps"
)

// ComponentDefinition represents a semantic component
type ComponentDefinition struct {
	Name               string                    `json:"-"`
	Class              string                    `json:"$class"`
	Description        string                    `json:"$description,omitempty"`
	Contains           []string                  `json:"$contains,omitempty"` // Child components this can contain
	Requires           string                    `json:"$requires,omitempty"` // Parent component this must be inside
	Base               map[string]any            `json:"base"`
	Variants           map[string]VariantDef     `json:"variants"`
	Sizes              map[string]VariantDef     `json:"sizes"`
	States             map[string]VariantDef     `json:"states"` // Component states (error, active, etc.)
	ContainerOverrides map[string]map[string]any `json:"-"`      // $container: query → properties
}

// VariantDef represents a specific variant (primary, outline) or size (sm, lg)
type VariantDef struct {
	Class      string           `json:"$class"`
	Properties map[string]any   `json:"-"` // CSS properties — see MarshalJSON
	States     map[string]State `json:"-"` // :hover, :focus, etc — see MarshalJSON
}

// State represents a CSS pseudo-class state
type State struct {
	Properties map[string]any
}

// MarshalJSON writes a variant in the catalog's explicit shape:
//
//	{"$class": "btn-primary",
//	 "properties": {"background": "…"},
//	 "states": {"&:hover": {"background": "…"}}}
//
// Deliberately not the inverse of UnmarshalJSON, which reads the authored
// shape — properties inline alongside $class, states keyed by a leading
// & or :. Round-tripping that shape out would hand every consumer the
// same sniffing problem tokenctl solved on the way in, and would collide
// with any property literally named "$class". The catalog is an export
// for external tools, never re-read by tokenctl, so it gets the shape
// that is cheapest to consume correctly.
//
// Empty properties and states are omitted rather than written as {}: a
// size class with no states should not claim an empty set of them.
func (v VariantDef) MarshalJSON() ([]byte, error) {
	out := map[string]any{"$class": v.Class}
	if len(v.Properties) > 0 {
		out["properties"] = v.Properties
	}
	if len(v.States) > 0 {
		states := make(map[string]map[string]any, len(v.States))
		for key, st := range v.States {
			props := st.Properties
			if props == nil {
				props = map[string]any{}
			}
			states[key] = props
		}
		out["states"] = states
	}
	return json.Marshal(out)
}

// SplitProperties separates authored properties from nested pseudo-selector
// blocks, the way the CSS generator does when it walks a component's base.
//
// A component's `base` map is authored flat: real CSS properties sit
// beside keys like "&:hover" whose values are property maps of their own.
// Anything consuming base has to make that split, and making it twice in
// two files is how the catalog came to describe a base class as having a
// property called "&:hover".
func SplitProperties(base map[string]any) (props map[string]any, states map[string]State) {
	props = make(map[string]any, len(base))
	states = map[string]State{}
	for k, v := range base {
		if len(k) > 0 && (k[0] == '&' || k[0] == ':') {
			nested := map[string]any{}
			if m, ok := v.(map[string]any); ok {
				maps.Copy(nested, m)
			}
			states[k] = State{Properties: nested}
			continue
		}
		props[k] = v
	}
	return props, states
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
			if s, ok := val.(string); ok {
				v.Class = s
			}
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

		// Extract $container overrides (query → properties)
		if containerRaw, ok := node["$container"].(map[string]any); ok {
			comp.ContainerOverrides = make(map[string]map[string]any)
			for query, propsRaw := range containerRaw {
				if props, ok := propsRaw.(map[string]any); ok {
					comp.ContainerOverrides[query] = props
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
