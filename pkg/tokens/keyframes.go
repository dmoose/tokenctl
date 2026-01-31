// tokenctl/pkg/tokens/keyframes.go

package tokens

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// KeyframeDefinition represents a CSS @keyframes animation
type KeyframeDefinition struct {
	Name   string                       // Animation name (e.g., "skeleton-pulse")
	Frames map[string]map[string]string // Frame selector -> properties (e.g., "0%, 100%" -> {"opacity": "1"})
}

// ExtractKeyframes scans the dictionary for keyframes definitions
// and returns KeyframeDefinition entries for each one found.
// Keyframes are defined at the root level under a "keyframes" key.
func ExtractKeyframes(dict *Dictionary) []KeyframeDefinition {
	var keyframes []KeyframeDefinition

	// Look for "keyframes" at the root level
	keyframesNode, ok := dict.Root["keyframes"]
	if !ok {
		return keyframes
	}

	keyframesMap, ok := keyframesNode.(map[string]any)
	if !ok {
		return keyframes
	}

	// Iterate over each keyframe definition
	for name, framesNode := range keyframesMap {
		framesMap, ok := framesNode.(map[string]any)
		if !ok {
			continue
		}

		kf := KeyframeDefinition{
			Name:   name,
			Frames: make(map[string]map[string]string),
		}

		// Process each frame selector (e.g., "0%, 100%", "50%", "from", "to")
		for selector, propsNode := range framesMap {
			propsMap, ok := propsNode.(map[string]any)
			if !ok {
				continue
			}

			props := make(map[string]string)
			for propName, propValue := range propsMap {
				// Convert value to string
				props[propName] = fmt.Sprintf("%v", propValue)
			}

			kf.Frames[selector] = props
		}

		keyframes = append(keyframes, kf)
	}

	// Sort by name for deterministic output
	sort.Slice(keyframes, func(i, j int) bool {
		return keyframes[i].Name < keyframes[j].Name
	})

	return keyframes
}

// GenerateKeyframesCSS generates CSS @keyframes blocks from keyframe definitions
func GenerateKeyframesCSS(keyframes []KeyframeDefinition) string {
	if len(keyframes) == 0 {
		return ""
	}

	var sb strings.Builder

	for _, kf := range keyframes {
		sb.WriteString(fmt.Sprintf("@keyframes %s {\n", kf.Name))

		// Sort frame selectors for deterministic output
		selectors := make([]string, 0, len(kf.Frames))
		for selector := range kf.Frames {
			selectors = append(selectors, selector)
		}
		sort.Slice(selectors, func(i, j int) bool {
			// Sort by percentage value for natural ordering
			// "from" < "0%" < "50%" < "100%" < "to"
			return keyframeSelectorOrder(selectors[i]) < keyframeSelectorOrder(selectors[j])
		})

		for _, selector := range selectors {
			props := kf.Frames[selector]
			sb.WriteString(fmt.Sprintf("  %s {\n", selector))

			// Sort properties for deterministic output
			propNames := make([]string, 0, len(props))
			for name := range props {
				propNames = append(propNames, name)
			}
			sort.Strings(propNames)

			for _, propName := range propNames {
				sb.WriteString(fmt.Sprintf("    %s: %s;\n", propName, props[propName]))
			}

			sb.WriteString("  }\n")
		}

		sb.WriteString("}\n\n")
	}

	return sb.String()
}

// keyframeSelectorOrder returns a sortable value for keyframe selectors
func keyframeSelectorOrder(selector string) int {
	// Handle special keywords
	switch selector {
	case "from":
		return 0
	case "to":
		return 100
	}

	// Extract first percentage value for sorting
	// Handle "0%, 100%" by taking the first value
	selector = strings.TrimSpace(selector)
	if idx := strings.Index(selector, ","); idx > 0 {
		selector = strings.TrimSpace(selector[:idx])
	}

	// Parse percentage; unparseable selectors default to 0
	selector = strings.TrimSuffix(selector, "%")
	pct, _ := strconv.Atoi(selector)
	return pct
}
