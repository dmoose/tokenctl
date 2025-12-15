package tokens

import (
	"fmt"
	"regexp"
	"strings"
)

// Resolver handles the resolution of token references ({path.to.token})
type Resolver struct {
	flatTokens map[string]interface{}
	cache      map[string]interface{}
	stack      []string // Cycle detection stack
}

// refRegex matches {path.to.token}
var refRegex = regexp.MustCompile(`\{([^}]+)\}`)

// NewResolver creates a new resolver from a dictionary
func NewResolver(d *Dictionary) (*Resolver, error) {
	flat := make(map[string]interface{})
	if err := flatten(d.Root, "", flat); err != nil {
		return nil, err
	}
	return &Resolver{
		flatTokens: flat,
		cache:      make(map[string]interface{}),
		stack:      []string{},
	}, nil
}

// ResolveAll resolves all tokens in the dictionary
func (r *Resolver) ResolveAll() (map[string]interface{}, error) {
	resolved := make(map[string]interface{})
	for path, val := range r.flatTokens {
		r.stack = []string{} // Reset stack for each root token
		res, err := r.ResolveValue(path, val)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", path, err)
		}
		resolved[path] = res
	}
	return resolved, nil
}

// ResolveValue resolves a value that might contain references
func (r *Resolver) ResolveValue(path string, value interface{}) (interface{}, error) {
	// If it's not a string, it can't contain a reference (in this spec version)
	// Unless it's a composite value which we don't fully support resolving *inside* yet
	valStr, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Check for references
	if !strings.Contains(valStr, "{") {
		return value, nil
	}

	// Cycle detection
	for _, p := range r.stack {
		if p == path {
			return nil, fmt.Errorf("circular dependency detected: %s -> %s", strings.Join(r.stack, " -> "), path)
		}
	}
	r.stack = append(r.stack, path)
	defer func() {
		// Pop from stack
		if len(r.stack) > 0 {
			r.stack = r.stack[:len(r.stack)-1]
		}
	}()

	// Replace all occurrences of {path}
	// We handle two cases:
	// 1. The whole value is a reference: "{color.brand}" -> returns the raw value of color.brand (preserves type)
	// 2. String interpolation: "1px solid {color.brand}" -> returns string with replaced value

	// Check exact match first (preserves type)
	if strings.HasPrefix(valStr, "{") && strings.HasSuffix(valStr, "}") && strings.Count(valStr, "{") == 1 {
		refPath := valStr[1 : len(valStr)-1]
		return r.resolveReference(refPath)
	}

	// String interpolation
	resolvedStr := valStr
	matches := refRegex.FindAllStringSubmatch(valStr, -1)
	for _, match := range matches {
		fullMatch := match[0] // {foo.bar}
		refPath := match[1]   // foo.bar

		resolvedVal, err := r.resolveReference(refPath)
		if err != nil {
			return nil, err
		}

		// Convert resolved value to string for interpolation
		resolvedStr = strings.Replace(resolvedStr, fullMatch, fmt.Sprintf("%v", resolvedVal), 1)
	}

	return resolvedStr, nil
}

func (r *Resolver) resolveReference(path string) (interface{}, error) {
	// Check cache
	if val, ok := r.cache[path]; ok {
		return val, nil
	}

	// Lookup token
	val, ok := r.flatTokens[path]
	if !ok {
		return nil, fmt.Errorf("reference not found: %s", path)
	}

	// Recursively resolve
	resolved, err := r.ResolveValue(path, val)
	if err != nil {
		return nil, err
	}

	// Cache result
	r.cache[path] = resolved
	return resolved, nil
}

// flatten walks the dictionary and flattens it into dot-notation paths mapping to $value
func flatten(node map[string]interface{}, currentPath string, result map[string]interface{}) error {
	if IsToken(node) {
		result[currentPath] = node["$value"]
		return nil
	}

	for key, val := range node {
		if strings.HasPrefix(key, "$") {
			continue
		}

		childMap, ok := val.(map[string]interface{})
		if !ok {
			// Skip malformed children
			continue
		}

		childPath := key
		if currentPath != "" {
			childPath = currentPath + "." + key
		}

		if err := flatten(childMap, childPath, result); err != nil {
			return err
		}
	}
	return nil
}
