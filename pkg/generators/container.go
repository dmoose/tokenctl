package generators

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

// GenerateContainerCSS creates @container query blocks from container overrides.
// Overrides sharing the same query are grouped into a single @container block.
// Output is NOT inside any @layer — container rules override component layer
// styles via natural cascade (the @container condition provides specificity).
func GenerateContainerCSS(overrides []tokens.ContainerOverride) string {
	if len(overrides) == 0 {
		return ""
	}

	// Group overrides by container query
	byQuery := make(map[string][]tokens.ContainerOverride)
	for _, o := range overrides {
		byQuery[o.ContainerQuery] = append(byQuery[o.ContainerQuery], o)
	}

	// Sort queries alphabetically for deterministic output
	queries := make([]string, 0, len(byQuery))
	for q := range byQuery {
		queries = append(queries, q)
	}
	sort.Strings(queries)

	var sb strings.Builder

	for _, query := range queries {
		group := byQuery[query]

		// Sort selectors within each query by class name
		sort.Slice(group, func(i, j int) bool {
			return group[i].ComponentClass < group[j].ComponentClass
		})

		sb.WriteString(fmt.Sprintf("@container %s {\n", query))

		for _, o := range group {
			sb.WriteString(fmt.Sprintf("  .%s {\n", o.ComponentClass))
			writeProperties(&sb, o.Properties, 4)
			sb.WriteString("  }\n")
		}

		sb.WriteString("}\n\n")
	}

	return sb.String()
}
