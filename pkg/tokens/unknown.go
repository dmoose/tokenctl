// tokenctl/pkg/tokens/unknown.go
//
// Loudness for input tokenctl silently drops.
//
// The build consumes a fixed vocabulary: a component node reads
// base/variants/sizes/states, metadata is read from a known set of
// $-prefixed keys, and a variant/size/state is only emitted when it
// carries $class. Anything outside that vocabulary used to be dropped on
// the floor without a word — a site wrote "ratios" where the schema says
// "variants" and shipped five classes with zero styling, and nothing in
// the build reported a problem. This file names what was dropped.
package tokens

import (
	"fmt"
	"sort"
	"strings"
)

// FindingKind classifies why a key was dropped.
type FindingKind string

const (
	// FindingUnknownComponentKey is a non-$ key inside a $type:component
	// node that is neither a sub-block (base/variants/sizes/states) nor a
	// nested selector. Its contents never reach the generator.
	FindingUnknownComponentKey FindingKind = "unknown-component-key"

	// FindingUnknownMetadataKey is a $-prefixed key tokenctl never reads.
	// Usually a typo ($vairants) or an intent the tool does not implement.
	FindingUnknownMetadataKey FindingKind = "unknown-metadata-key"

	// FindingMissingClass is a variants/sizes/states entry with no $class.
	// The generator emits a rule only when it has a class to hang it on,
	// so the entry's declarations are silently discarded.
	FindingMissingClass FindingKind = "missing-class"
)

// Finding is one piece of input the build will not consume.
type Finding struct {
	Kind       FindingKind
	Path       string // dot path to the offending key
	Key        string // the key itself
	SourceFile string // file the key was read from, when known
	Hint       string // what the author probably meant
}

func (f Finding) String() string {
	loc := f.Path
	if f.SourceFile != "" {
		loc = fmt.Sprintf("%s (%s)", f.Path, f.SourceFile)
	}
	if f.Hint != "" {
		return fmt.Sprintf("%s: %s — %s", f.Kind, loc, f.Hint)
	}
	return fmt.Sprintf("%s: %s", f.Kind, loc)
}

// componentSubBlocks are the only non-$ keys a component node may hold
// besides nested selectors and // comments.
var componentSubBlocks = map[string]bool{
	"base":     true,
	"variants": true,
	"sizes":    true,
	"states":   true,
}

// knownMetadataKeys is every $-prefixed key tokenctl reads somewhere.
// Adding a feature that reads a new $key means adding it here, otherwise
// the audit will call the feature's own input unknown.
var knownMetadataKeys = map[string]bool{
	"$avoid":        true,
	"$breakpoints":  true,
	"$class":        true,
	"$container":    true,
	"$contains":     true,
	"$customizable": true,
	"$default":      true,
	"$deprecated":   true,
	"$desc":         true,
	"$description":  true,
	"$extends":      true,
	"$extensions":   true,
	"$layer":        true,
	"$max":          true,
	"$meta":         true,
	"$min":          true,
	"$property":     true,
	"$requires":     true,
	"$responsive":   true,
	"$scale":        true,
	"$schema":       true,
	"$type":         true,
	"$usage":        true,
	"$value":        true,
	"$version":      true,
}

// IsCommentKey reports whether a key is an in-file comment rather than
// data. The token files carry schema notes as "// why"-style keys and
// that convention is deliberate, so comments are never findings.
func IsCommentKey(key string) bool {
	return strings.HasPrefix(key, "//")
}

// isNestedSelector reports whether a key is a CSS nested selector.
func isNestedSelector(key string) bool {
	return len(key) > 0 && (key[0] == '&' || key[0] == ':')
}

// AuditUnknownKeys walks a dictionary and returns every key whose
// content the build will not consume. Findings are sorted by path so
// output is deterministic.
func AuditUnknownKeys(dict *Dictionary) []Finding {
	if dict == nil {
		return nil
	}
	var findings []Finding
	auditNode(dict.Root, "", dict, &findings)
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		return findings[i].Kind < findings[j].Kind
	})
	return findings
}

func auditNode(node map[string]any, path string, dict *Dictionary, out *[]Finding) {
	isComponent := false
	if t, ok := node["$type"]; ok && t == "component" {
		isComponent = true
	}

	keys := make([]string, 0, len(node))
	for k := range node {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		val := node[key]
		child := path
		if child == "" {
			child = key
		} else {
			child = child + "." + key
		}

		if IsCommentKey(key) {
			continue
		}

		if strings.HasPrefix(key, "$") {
			if !knownMetadataKeys[key] {
				*out = append(*out, Finding{
					Kind:       FindingUnknownMetadataKey,
					Path:       child,
					Key:        key,
					SourceFile: sourceFileFor(dict, path),
					Hint:       "tokenctl reads no such key; its value is discarded",
				})
			}
			continue
		}

		childMap, isMap := val.(map[string]any)

		if isComponent && !isNestedSelector(key) && !componentSubBlocks[key] {
			*out = append(*out, Finding{
				Kind:       FindingUnknownComponentKey,
				Path:       child,
				Key:        key,
				SourceFile: sourceFileFor(dict, path),
				Hint:       "not a component sub-block (base, variants, sizes, states); nothing under it is emitted",
			})
			// Still descend: a misnamed block can hold a nested component.
		}

		if isComponent && componentSubBlocks[key] && key != "base" && isMap {
			auditVariantBlock(childMap, child, key, dict, out)
		}

		if isMap {
			auditNode(childMap, child, dict, out)
		}
	}
}

// auditVariantBlock checks the entries of a variants/sizes/states block.
// Each entry needs $class; without one the generator has no selector to
// emit and drops the whole entry.
func auditVariantBlock(block map[string]any, path, blockName string, dict *Dictionary, out *[]Finding) {
	names := make([]string, 0, len(block))
	for k := range block {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
		if strings.HasPrefix(name, "$") || IsCommentKey(name) {
			continue
		}
		entry, ok := block[name].(map[string]any)
		if !ok {
			continue
		}
		if _, hasClass := entry["$class"]; !hasClass {
			*out = append(*out, Finding{
				Kind:       FindingMissingClass,
				Path:       path + "." + name,
				Key:        name,
				SourceFile: sourceFileFor(dict, path),
				Hint:       fmt.Sprintf("%s entries need $class; without it no rule is emitted", blockName),
			})
		}
	}
}

// sourceFileFor finds the file a path came from. SourceFiles is keyed by
// token path, so walk up until a prefix matches.
func sourceFileFor(dict *Dictionary, path string) string {
	if dict == nil || len(dict.SourceFiles) == 0 {
		return ""
	}
	if f, ok := dict.SourceFiles[path]; ok {
		return f
	}
	// Several token paths can sit under the same prefix; sort so the
	// reported file is stable run to run.
	prefixed := make([]string, 0, len(dict.SourceFiles))
	for p := range dict.SourceFiles {
		if strings.HasPrefix(p, path+".") {
			prefixed = append(prefixed, p)
		}
	}
	if len(prefixed) > 0 {
		sort.Strings(prefixed)
		return dict.SourceFiles[prefixed[0]]
	}
	for path != "" {
		idx := strings.LastIndex(path, ".")
		if idx < 0 {
			break
		}
		path = path[:idx]
		if f, ok := dict.SourceFiles[path]; ok {
			return f
		}
	}
	return ""
}
