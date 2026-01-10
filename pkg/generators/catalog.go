// tokenctl/pkg/generators/catalog.go
package generators

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

// CatalogSchemaVersion is the current catalog schema version
const CatalogSchemaVersion = "2.0"

// TokenctlVersion is the tokenctl version (updated on releases)
const TokenctlVersion = "1.1.0"

// CatalogGenerator generates a structured JSON catalog for external tools
type CatalogGenerator struct {
}

func NewCatalogGenerator() *CatalogGenerator {
	return &CatalogGenerator{}
}

// CatalogSchema represents the output format
type CatalogSchema struct {
	Meta       CatalogMeta                 `json:"meta"`
	Tokens     map[string]interface{}      `json:"tokens"`
	Components map[string]ComponentSummary `json:"components"`
	Themes     map[string]ThemeInfo        `json:"themes,omitempty"`
}

type CatalogMeta struct {
	Version         string `json:"version"`
	GeneratedAt     string `json:"generated_at"`
	TokenctlVersion string `json:"tokenctl_version"`
}

type ComponentSummary struct {
	Classes     []string                     `json:"classes"`
	Definitions map[string]tokens.VariantDef `json:"definitions"`
}

// ThemeInfo contains resolved theme data for external consumers
type ThemeInfo struct {
	Extends     *string                `json:"extends"`
	Description string                 `json:"description,omitempty"`
	Tokens      map[string]interface{} `json:"tokens"`
	Diff        map[string]interface{} `json:"diff,omitempty"`
}

// CatalogThemeInput provides theme data from the build process
type CatalogThemeInput struct {
	Extends        *string                // Parent theme name (nil if extends base)
	Description    string                 // From $description field
	ResolvedTokens map[string]interface{} // Fully resolved token values
	DiffTokens     map[string]interface{} // Only tokens that differ from parent/base
}

// Generate creates the JSON catalog
func (g *CatalogGenerator) Generate(
	resolvedTokens map[string]interface{},
	components map[string]tokens.ComponentDefinition,
	themes map[string]CatalogThemeInput,
) (string, error) {

	catalog := CatalogSchema{
		Meta: CatalogMeta{
			Version:         CatalogSchemaVersion,
			GeneratedAt:     time.Now().Format(time.RFC3339),
			TokenctlVersion: TokenctlVersion,
		},
		Tokens:     make(map[string]interface{}),
		Components: make(map[string]ComponentSummary),
		Themes:     make(map[string]ThemeInfo),
	}

	// 1. Filter Atomic Tokens (exclude components/maps)
	for k, v := range resolvedTokens {
		if _, ok := v.(map[string]interface{}); !ok {
			catalog.Tokens[k] = v
		}
	}

	// 2. Process Components
	for name, comp := range components {
		summary := ComponentSummary{
			Classes:     []string{},
			Definitions: make(map[string]tokens.VariantDef),
		}

		// Collect all generated classes
		if comp.Class != "" {
			summary.Classes = append(summary.Classes, comp.Class)
			// Add base definition
			summary.Definitions[comp.Class] = tokens.VariantDef{
				Class:      comp.Class,
				Properties: comp.Base,
			}
		}

		for _, variant := range comp.Variants {
			if variant.Class != "" {
				summary.Classes = append(summary.Classes, variant.Class)
				summary.Definitions[variant.Class] = variant
			}
		}

		for _, size := range comp.Sizes {
			if size.Class != "" {
				summary.Classes = append(summary.Classes, size.Class)
				summary.Definitions[size.Class] = size
			}
		}

		catalog.Components[name] = summary
	}

	// 3. Process Themes
	for name, themeInput := range themes {
		themeInfo := ThemeInfo{
			Extends:     themeInput.Extends,
			Description: themeInput.Description,
			Tokens:      filterAtomicTokens(themeInput.ResolvedTokens),
			Diff:        filterAtomicTokens(themeInput.DiffTokens),
		}
		catalog.Themes[name] = themeInfo
	}

	// Omit themes section if empty
	if len(catalog.Themes) == 0 {
		catalog.Themes = nil
	}

	// Marshal to JSON
	bytes, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal catalog: %w", err)
	}

	return string(bytes), nil
}

// filterAtomicTokens filters out nested maps, keeping only atomic token values
func filterAtomicTokens(tokens map[string]interface{}) map[string]interface{} {
	if tokens == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range tokens {
		if _, ok := v.(map[string]interface{}); !ok {
			result[k] = v
		}
	}
	return result
}
