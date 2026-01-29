package tokens

import "sort"

// ContainerOverride represents a single component's CSS overrides within a container query.
type ContainerOverride struct {
	ContainerQuery string         // e.g. "blogpost (max-width: 1024px)"
	ComponentClass string         // e.g. "blogpost-sidebar"
	Properties     map[string]any // e.g. {"display": "none"}
	SourceFile     string
}

// ExtractContainerOverrides collects all $container entries from extracted components.
func ExtractContainerOverrides(components map[string]ComponentDefinition) []ContainerOverride {
	var results []ContainerOverride

	// Sort component names for deterministic output
	names := make([]string, 0, len(components))
	for name := range components {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		comp := components[name]
		if len(comp.ContainerOverrides) == 0 {
			continue
		}

		// Sort query keys for deterministic output
		queries := make([]string, 0, len(comp.ContainerOverrides))
		for q := range comp.ContainerOverrides {
			queries = append(queries, q)
		}
		sort.Strings(queries)

		for _, query := range queries {
			props := comp.ContainerOverrides[query]
			results = append(results, ContainerOverride{
				ContainerQuery: query,
				ComponentClass: comp.Class,
				Properties:     props,
			})
		}
	}

	return results
}
