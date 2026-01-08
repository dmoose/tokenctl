package generators

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dmoose/tokenctl/pkg/tokens"
)

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
}

type CatalogMeta struct {
	Version     string `json:"version"`
	GeneratedAt string `json:"generated_at"`
}

type ComponentSummary struct {
	Classes     []string                     `json:"classes"`
	Definitions map[string]tokens.VariantDef `json:"definitions"`
}

// Generate creates the JSON catalog
func (g *CatalogGenerator) Generate(
	resolvedTokens map[string]interface{},
	components map[string]tokens.ComponentDefinition,
) (string, error) {

	catalog := CatalogSchema{
		Meta: CatalogMeta{
			Version:     "1.0",
			GeneratedAt: time.Now().Format(time.RFC3339),
		},
		Tokens:     make(map[string]interface{}),
		Components: make(map[string]ComponentSummary),
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

	// Marshal to JSON
	bytes, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal catalog: %w", err)
	}

	return string(bytes), nil
}
