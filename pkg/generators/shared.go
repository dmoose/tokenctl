package generators

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

// tokenRefPattern matches {token.path.here} references in CSS values.
// Compiled once at package level to avoid per-call overhead.
var tokenRefPattern = regexp.MustCompile(`\{([^}]+)\}`)

// resolveTokenReferences converts all {token.path} references to var(--token-path).
// Handles multiple references in a single string.
func resolveTokenReferences(value string) string {
	return tokenRefPattern.ReplaceAllStringFunc(value, func(match string) string {
		tokenPath := match[1 : len(match)-1]
		cssVar := strings.ReplaceAll(tokenPath, ".", "-")
		return fmt.Sprintf("var(--%s)", cssVar)
	})
}

// generatePropertyDeclarations creates @property declarations for typed tokens.
func generatePropertyDeclarations(properties []tokens.PropertyToken) string {
	var sb strings.Builder

	sorted := make([]tokens.PropertyToken, len(properties))
	copy(sorted, properties)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	for _, prop := range sorted {
		sb.WriteString(fmt.Sprintf("@property %s {\n", prop.CSSName))
		sb.WriteString(fmt.Sprintf("  syntax: '%s';\n", prop.CSSSyntax))
		if prop.Inherits {
			sb.WriteString("  inherits: true;\n")
		} else {
			sb.WriteString("  inherits: false;\n")
		}
		sb.WriteString(fmt.Sprintf("  initial-value: %s;\n", prop.InitialValue))
		sb.WriteString("}\n\n")
	}

	return sb.String()
}

// buildStateSelector converts a state key to a CSS selector.
//
// A state key may be a selector *list*: "& a, & b" is two selectors, and
// each segment gets its own expansion. Expanding only the first segment
// left a bare "&" in the emitted CSS ("`.prose ul, & ol`") — legal to the
// parser but matching nothing, so the second half of the rule silently
// did not apply. Split on top-level commas first, expand each segment,
// then rejoin.
func buildStateSelector(className, stateKey string) string {
	segments := splitSelectorList(stateKey)
	expanded := make([]string, 0, len(segments))
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		expanded = append(expanded, expandSelectorSegment(className, seg))
	}
	if len(expanded) == 0 {
		return fmt.Sprintf(".%s", className)
	}
	return strings.Join(expanded, ", ")
}

// expandSelectorSegment expands one comma-free selector segment against
// the owning class. Every "&" in the segment is replaced by the parent
// class (CSS nesting semantics); a segment with no "&" is attached as a
// pseudo/attribute suffix when it starts with ":", and as a descendant
// otherwise.
func expandSelectorSegment(className, segment string) string {
	parent := "." + className
	if strings.Contains(segment, "&") {
		return replaceNestingSelector(segment, parent)
	}
	if strings.HasPrefix(segment, ":") {
		return parent + segment
	}
	return parent + " " + segment
}

// replaceNestingSelector replaces every "&" in sel with parent, skipping
// occurrences inside quoted strings so attribute values such as
// [data-x="a&b"] survive intact.
func replaceNestingSelector(sel, parent string) string {
	var sb strings.Builder
	var quote byte
	for i := 0; i < len(sel); i++ {
		c := sel[i]
		switch {
		case quote != 0:
			if c == '\\' && i+1 < len(sel) {
				sb.WriteByte(c)
				i++
				sb.WriteByte(sel[i])
				continue
			}
			if c == quote {
				quote = 0
			}
			sb.WriteByte(c)
		case c == '\'' || c == '"':
			quote = c
			sb.WriteByte(c)
		case c == '&':
			sb.WriteString(parent)
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

// splitSelectorList splits a CSS selector list on top-level commas.
// Commas nested inside (), [], or quotes — :is(a, b), [x="a,b"],
// :nth-child(2n + 1 of .a, .b) — are not separators.
func splitSelectorList(sel string) []string {
	var (
		out   []string
		sb    strings.Builder
		depth int
		quote byte
	)
	flush := func() {
		out = append(out, strings.TrimSpace(sb.String()))
		sb.Reset()
	}
	for i := 0; i < len(sel); i++ {
		c := sel[i]
		switch {
		case quote != 0:
			if c == '\\' && i+1 < len(sel) {
				sb.WriteByte(c)
				i++
				sb.WriteByte(sel[i])
				continue
			}
			if c == quote {
				quote = 0
			}
			sb.WriteByte(c)
		case c == '\'' || c == '"':
			quote = c
			sb.WriteByte(c)
		case c == '(' || c == '[':
			depth++
			sb.WriteByte(c)
		case c == ')' || c == ']':
			if depth > 0 {
				depth--
			}
			sb.WriteByte(c)
		case c == ',' && depth == 0:
			flush()
		default:
			sb.WriteByte(c)
		}
	}
	flush()
	return out
}

// writeProperties writes CSS properties with proper indentation and serialization.
func writeProperties(sb *strings.Builder, props map[string]any, indent int) {
	if len(props) == 0 {
		return
	}

	padding := strings.Repeat(" ", indent)

	propNames := make([]string, 0, len(props))
	for k := range props {
		propNames = append(propNames, k)
	}
	sort.Strings(propNames)

	for _, k := range propNames {
		v := props[k]

		if strings.HasPrefix(k, "$") {
			continue
		}

		valStr := SerializeValueForProperty(k, v)
		if valStr == "" {
			// Empty serialization (e.g. malformed token-shaped value
			// with no $value). Skip rather than emit `prop: ;`.
			continue
		}
		val := resolveTokenReferences(valStr)

		fmt.Fprintf(sb, "%s%s: %s;\n", padding, k, val)
	}
}
