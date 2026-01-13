// tokenctl/pkg/tokens/responsive.go
package tokens

import (
	"fmt"
	"maps"
	"sort"
	"strings"
)

// DefaultBreakpoints provides sensible defaults if not specified
var DefaultBreakpoints = map[string]string{
	"sm": "640px",
	"md": "768px",
	"lg": "1024px",
	"xl": "1280px",
}

// ResponsiveToken represents a token with breakpoint-specific overrides
type ResponsiveToken struct {
	Path       string
	BaseValue  any
	Overrides  map[string]any // breakpoint name -> value
	Type       string
	SourceFile string
}

// ExtractBreakpoints retrieves the $breakpoints configuration from a dictionary
// Falls back to DefaultBreakpoints if not defined
func ExtractBreakpoints(d *Dictionary) map[string]string {
	breakpoints := make(map[string]string)

	if bp, ok := d.Root["$breakpoints"].(map[string]any); ok {
		for name, value := range bp {
			if strVal, ok := value.(string); ok {
				breakpoints[name] = strVal
			}
		}
	}

	// If no breakpoints defined, use defaults
	if len(breakpoints) == 0 {
		maps.Copy(breakpoints, DefaultBreakpoints)
	}

	return breakpoints
}

// ExtractResponsiveTokens finds all tokens with $responsive overrides
func ExtractResponsiveTokens(d *Dictionary) []ResponsiveToken {
	var results []ResponsiveToken
	extractResponsiveRecursive(d, d.Root, "", "", &results)
	return results
}

// extractResponsiveRecursive walks the tree looking for $responsive fields
func extractResponsiveRecursive(d *Dictionary, node map[string]any, currentPath string, inheritedType string, results *[]ResponsiveToken) {
	// Check for $type at this level
	currentType := inheritedType
	if t, ok := node["$type"].(string); ok {
		currentType = t
	}

	if IsToken(node) {
		// Check for $responsive field
		responsiveRaw, hasResponsive := node["$responsive"]
		if !hasResponsive {
			return
		}

		responsive, ok := responsiveRaw.(map[string]any)
		if !ok {
			return
		}

		rt := ResponsiveToken{
			Path:      currentPath,
			BaseValue: node["$value"],
			Type:      currentType,
			Overrides: make(map[string]any),
		}

		maps.Copy(rt.Overrides, responsive)

		if sourceFile, ok := d.SourceFiles[currentPath]; ok {
			rt.SourceFile = sourceFile
		}

		*results = append(*results, rt)
		return
	}

	// Recurse
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

		extractResponsiveRecursive(d, childMap, childPath, currentType, results)
	}
}

// GenerateResponsiveCSS creates media query blocks for responsive tokens
// Returns a string with all the media query CSS
func GenerateResponsiveCSS(breakpoints map[string]string, responsiveTokens []ResponsiveToken) string {
	if len(responsiveTokens) == 0 {
		return ""
	}

	// Group tokens by breakpoint
	byBreakpoint := make(map[string][]ResponsiveToken)
	for _, rt := range responsiveTokens {
		for bp := range rt.Overrides {
			byBreakpoint[bp] = append(byBreakpoint[bp], rt)
		}
	}

	// Sort breakpoints by their pixel value (ascending)
	sortedBreakpoints := sortBreakpointsBySize(breakpoints)

	var sb strings.Builder

	for _, bp := range sortedBreakpoints {
		tokens := byBreakpoint[bp]
		if len(tokens) == 0 {
			continue
		}

		minWidth, ok := breakpoints[bp]
		if !ok {
			continue
		}

		sb.WriteString(fmt.Sprintf("@media (min-width: %s) {\n", minWidth))
		sb.WriteString("  :root {\n")

		// Sort tokens by path for deterministic output
		sort.Slice(tokens, func(i, j int) bool {
			return tokens[i].Path < tokens[j].Path
		})

		for _, rt := range tokens {
			if value, ok := rt.Overrides[bp]; ok {
				cssVar := strings.ReplaceAll(rt.Path, ".", "-")
				sb.WriteString(fmt.Sprintf("    --%s: %v;\n", cssVar, value))
			}
		}

		sb.WriteString("  }\n")
		sb.WriteString("}\n\n")
	}

	return sb.String()
}

// sortBreakpointsBySize returns breakpoint names sorted by their pixel value
func sortBreakpointsBySize(breakpoints map[string]string) []string {
	type bpEntry struct {
		name  string
		value int
	}

	entries := make([]bpEntry, 0, len(breakpoints))
	for name, value := range breakpoints {
		// Parse pixel value (assumes format like "640px")
		var px int
		_, _ = fmt.Sscanf(value, "%dpx", &px)
		entries = append(entries, bpEntry{name: name, value: px})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].value < entries[j].value
	})

	result := make([]string, len(entries))
	for i, e := range entries {
		result[i] = e.name
	}
	return result
}
