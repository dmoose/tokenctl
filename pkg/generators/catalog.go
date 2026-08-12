// tokenctl/pkg/generators/catalog.go
package generators

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/dmoose/tokenctl/pkg/tokens"
	"github.com/dmoose/tokenctl/pkg/version"
)

// CatalogSchemaVersion is the current catalog schema version.
//
// 3.0: variant/size/state definitions carry their properties and states
// (they previously serialized to `{"$class": …}` alone); component-level
// states reach classes and definitions; meta.generated_at is opt-in so
// the same input produces the same bytes.
const CatalogSchemaVersion = "3.0"

// CatalogGenerator generates a structured JSON catalog for external tools
type CatalogGenerator struct {
	Category         string // Optional: filter to specific category (e.g., "color", "spacing")
	CustomizableOnly bool   // If true, only include tokens marked $customizable: true

	// GeneratedAt, when non-empty, is written to meta.generated_at
	// verbatim. Empty — the default — omits the field.
	//
	// It used to be time.Now(). That made every export differ from the
	// last one in a field describing nothing about the tokens, so the
	// output could not be content-hashed, diffed, or cached, and a
	// consumer could not tell a real change from a re-run. Callers that
	// genuinely want a stamp pass one; the caller decides its format and
	// owns the reproducibility it gives up.
	GeneratedAt string
}

// CatalogOptions configures catalog generation
type CatalogOptions struct {
	Category         string // Filter to specific category (empty = all)
	CustomizableOnly bool   // If true, only include tokens marked $customizable: true
	GeneratedAt      string // Opt-in meta.generated_at stamp; empty omits it
}

func NewCatalogGenerator() *CatalogGenerator {
	return &CatalogGenerator{}
}

// NewCatalogGeneratorWithOptions creates a generator with specific options
func NewCatalogGeneratorWithOptions(opts CatalogOptions) *CatalogGenerator {
	return &CatalogGenerator{
		Category:         opts.Category,
		CustomizableOnly: opts.CustomizableOnly,
		GeneratedAt:      opts.GeneratedAt,
	}
}

// CatalogSchema represents the output format
type CatalogSchema struct {
	Meta       CatalogMeta                 `json:"meta"`
	Tokens     map[string]any              `json:"tokens"`
	Components map[string]ComponentSummary `json:"components,omitempty"`
	Themes     map[string]ThemeInfo        `json:"themes,omitempty"`
}

// RichTokenInfo contains full token information for LLM consumption
type RichTokenInfo struct {
	Value        any      `json:"value"`
	Type         string   `json:"type,omitempty"`
	Description  string   `json:"description,omitempty"`
	Usage        []string `json:"usage,omitempty"`
	Avoid        string   `json:"avoid,omitempty"`
	Deprecated   any      `json:"deprecated,omitempty"`
	Customizable bool     `json:"customizable,omitempty"`
}

type CatalogMeta struct {
	Version string `json:"version"`
	// GeneratedAt is present only when the caller asked for a stamp.
	GeneratedAt     string `json:"generated_at,omitempty"`
	TokenctlVersion string `json:"tokenctl_version"`
	Category        string `json:"category,omitempty"`
}

type ComponentSummary struct {
	Description string   `json:"description,omitempty"`
	Contains    []string `json:"contains,omitempty"`
	Requires    string   `json:"requires,omitempty"`

	// Classes is every class the CSS generator emits for this
	// component, in the order it emits them: base, then variants,
	// sizes and states by name. Sorted rather than map-ordered — a
	// catalog whose class list reshuffles between identical runs
	// cannot be diffed or hashed.
	Classes     []string                     `json:"classes"`
	Definitions map[string]tokens.VariantDef `json:"definitions"`
}

// ThemeInfo contains resolved theme data for external consumers
type ThemeInfo struct {
	Extends     *string        `json:"extends"`
	Description string         `json:"description,omitempty"`
	Tokens      map[string]any `json:"tokens"`
	Diff        map[string]any `json:"diff,omitempty"`
}

// CatalogThemeInput provides theme data from the build process
type CatalogThemeInput struct {
	Extends        *string        // Parent theme name (nil if extends base)
	Description    string         // From $description field
	ResolvedTokens map[string]any // Fully resolved token values
	DiffTokens     map[string]any // Only tokens that differ from parent/base
}

// Generate creates the JSON catalog
// metadata is optional - if provided, tokens will include rich metadata (description, usage, avoid)
func (g *CatalogGenerator) Generate(
	resolvedTokens map[string]any,
	components map[string]tokens.ComponentDefinition,
	themes map[string]CatalogThemeInput,
) (string, error) {
	return g.GenerateWithMetadata(resolvedTokens, components, themes, nil)
}

// GenerateWithMetadata creates the JSON catalog with optional rich metadata
func (g *CatalogGenerator) GenerateWithMetadata(
	resolvedTokens map[string]any,
	components map[string]tokens.ComponentDefinition,
	themes map[string]CatalogThemeInput,
	metadata map[string]*tokens.TokenMetadata,
) (string, error) {

	catalog := CatalogSchema{
		Meta: CatalogMeta{
			Version:         CatalogSchemaVersion,
			GeneratedAt:     g.GeneratedAt,
			TokenctlVersion: version.String(),
			Category:        g.Category,
		},
		Tokens:     make(map[string]any),
		Components: make(map[string]ComponentSummary),
		Themes:     make(map[string]ThemeInfo),
	}

	// 1. Filter Atomic Tokens (exclude components/maps)
	for k, v := range resolvedTokens {
		if _, ok := v.(map[string]any); !ok {
			// Apply category filter if specified
			if g.Category != "" && !g.matchesCategory(k) {
				continue
			}

			// Get metadata for this token
			var meta *tokens.TokenMetadata
			if metadata != nil {
				meta = metadata[k]
			}

			// Apply customizable filter if specified
			if g.CustomizableOnly {
				if meta == nil || !meta.Customizable {
					continue
				}
			}

			// If metadata is provided and has rich info, use RichTokenInfo
			if meta != nil && hasRichMetadata(meta) {
				catalog.Tokens[k] = RichTokenInfo{
					Value:        v,
					Type:         meta.Type,
					Description:  meta.Description,
					Usage:        meta.Usage,
					Avoid:        meta.Avoid,
					Deprecated:   meta.Deprecated,
					Customizable: meta.Customizable,
				}
				continue
			}
			// Otherwise just use the value
			catalog.Tokens[k] = v
		}
	}

	// 2. Process Components (skip if filtering to non-component category)
	//
	// The class list has to match what the CSS generator actually emits,
	// which is base + variants + sizes + component-level states. States
	// were missing here: a consumer reading the catalog to find out what
	// `.input-error` is got told no such class exists, while the
	// stylesheet in front of it defined one.
	if g.Category == "" || g.Category == "components" {
		for name, comp := range components {
			summary := ComponentSummary{
				Description: comp.Description,
				Contains:    comp.Contains,
				Requires:    comp.Requires,
				Classes:     []string{},
				Definitions: make(map[string]tokens.VariantDef),
			}

			if comp.Class != "" {
				// The base map is authored flat, pseudo-selector blocks
				// mixed in with properties. Split it the same way the
				// CSS generator does.
				props, states := tokens.SplitProperties(comp.Base)
				summary.Classes = append(summary.Classes, comp.Class)
				summary.Definitions[comp.Class] = tokens.VariantDef{
					Class:      comp.Class,
					Properties: props,
					States:     states,
				}
			}

			// Sorted by the name the component authored, so two runs over
			// the same input emit the same list. Go map iteration order
			// is randomized; the previous code took it as given.
			for _, group := range []map[string]tokens.VariantDef{comp.Variants, comp.Sizes, comp.States} {
				for _, key := range sortedKeys(group) {
					def := group[key]
					if def.Class == "" {
						continue
					}
					summary.Classes = append(summary.Classes, def.Class)
					summary.Definitions[def.Class] = def
				}
			}

			catalog.Components[name] = summary
		}
	}

	// 3. Process Themes (include category-filtered tokens if filtering)
	if g.Category == "" {
		for name, themeInput := range themes {
			themeInfo := ThemeInfo{
				Extends:     themeInput.Extends,
				Description: themeInput.Description,
				Tokens:      filterAtomicTokens(themeInput.ResolvedTokens),
				Diff:        filterAtomicTokens(themeInput.DiffTokens),
			}
			catalog.Themes[name] = themeInfo
		}
	} else {
		// When filtering by category, only include relevant theme tokens
		for name, themeInput := range themes {
			filteredTokens := g.filterByCategory(filterAtomicTokens(themeInput.ResolvedTokens))
			filteredDiff := g.filterByCategory(filterAtomicTokens(themeInput.DiffTokens))

			// Only include theme if it has tokens in this category
			if len(filteredTokens) > 0 || len(filteredDiff) > 0 {
				themeInfo := ThemeInfo{
					Extends:     themeInput.Extends,
					Description: themeInput.Description,
					Tokens:      filteredTokens,
					Diff:        filteredDiff,
				}
				catalog.Themes[name] = themeInfo
			}
		}
	}

	// Omit themes section if empty
	if len(catalog.Themes) == 0 {
		catalog.Themes = nil
	}

	// Omit components section if empty
	if len(catalog.Components) == 0 {
		catalog.Components = nil
	}

	// Marshal to JSON
	bytes, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal catalog: %w", err)
	}

	return string(bytes), nil
}

// sortedKeys returns a map's keys in a stable order.
func sortedKeys(m map[string]tokens.VariantDef) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// filterAtomicTokens filters out nested maps, keeping only atomic token values
func filterAtomicTokens(tokens map[string]any) map[string]any {
	if tokens == nil {
		return nil
	}
	result := make(map[string]any)
	for k, v := range tokens {
		if _, ok := v.(map[string]any); !ok {
			result[k] = v
		}
	}
	return result
}

// matchesCategory checks if a token path belongs to the specified category
// Category matching is based on the first segment of the token path
// e.g., "color.primary" matches category "color"
// Also supports "colors" matching "color" (plural/singular flexibility)
func (g *CatalogGenerator) matchesCategory(tokenPath string) bool {
	if g.Category == "" {
		return true
	}

	// Get first segment of the token path
	topLevel, _, found := strings.Cut(tokenPath, ".")
	if !found {
		topLevel = tokenPath
	}

	// Direct match
	if topLevel == g.Category {
		return true
	}

	// Handle plural/singular variations
	// "colors" matches "color", "spacing" matches "spacings", etc.
	category := g.Category
	if strings.HasSuffix(category, "s") {
		if topLevel == category[:len(category)-1] {
			return true
		}
	} else {
		if topLevel == category+"s" {
			return true
		}
	}

	return false
}

// filterByCategory filters a token map to only include tokens matching the category
func (g *CatalogGenerator) filterByCategory(tokens map[string]any) map[string]any {
	if g.Category == "" || tokens == nil {
		return tokens
	}

	result := make(map[string]any)
	for k, v := range tokens {
		if g.matchesCategory(k) {
			result[k] = v
		}
	}
	return result
}

// hasRichMetadata checks if a TokenMetadata has any rich fields worth including
func hasRichMetadata(meta *tokens.TokenMetadata) bool {
	if meta == nil {
		return false
	}
	return meta.Description != "" || len(meta.Usage) > 0 || meta.Avoid != "" || meta.Deprecated != nil || meta.Customizable
}
