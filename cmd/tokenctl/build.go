// tokenctl/cmd/tokenctl/build.go
package main

import (
	"fmt"
	"os"

	"github.com/dmoose/tokenctl/pkg/generators"
	"github.com/dmoose/tokenctl/pkg/tokens"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [directory]",
	Short: "Build token artifacts",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runBuild,
}

var (
	format    string
	outputDir string
)

func init() {
	buildCmd.Flags().StringVarP(&format, "format", "f", "tailwind", "Output format (tailwind, catalog)")
	buildCmd.Flags().StringVarP(&outputDir, "output", "o", "dist", "Output directory")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	fmt.Printf("Building tokens from %s...\n", dir)

	// 1. Load Dictionary
	loader := tokens.NewLoader()

	// Load Base
	baseDict, err := loader.LoadBase(dir)
	if err != nil {
		return fmt.Errorf("failed to load base tokens: %w", err)
	}

	// Load Themes
	themes, err := loader.LoadThemes(dir)
	if err != nil {
		return fmt.Errorf("failed to load themes: %w", err)
	}

	// 2. Resolve Base (Root)
	resolver, err := tokens.NewResolver(baseDict)
	if err != nil {
		return fmt.Errorf("failed to initialize resolver for base: %w", err)
	}
	resolvedBase, err := resolver.ResolveAll()
	if err != nil {
		return fmt.Errorf("resolution failed for base: %w", err)
	}

	// 3. Prepare Generation Context
	var content string
	switch format {
	case "tailwind":
		gen := generators.NewTailwindGenerator()

		// Resolve theme inheritance chains (handles $extends)
		inheritedThemes, err := tokens.ResolveThemeInheritance(baseDict, themes)
		if err != nil {
			return fmt.Errorf("failed to resolve theme inheritance: %w", err)
		}

		// Build theme contexts
		themeContexts := make(map[string]generators.ThemeContext)
		for name, mergedDict := range inheritedThemes {
			// Resolve theme tokens
			themeResolver, err := tokens.NewResolver(mergedDict)
			if err != nil {
				return fmt.Errorf("failed to resolve theme %s: %w", name, err)
			}
			resolvedTheme, err := themeResolver.ResolveAll()
			if err != nil {
				return fmt.Errorf("resolution failed for theme %s: %w", name, err)
			}

			// Calculate diff from base (only output differences)
			themeDiff := tokens.Diff(resolvedTheme, resolvedBase)

			themeContexts[name] = generators.ThemeContext{
				Dict:           mergedDict,
				ResolvedTokens: resolvedTheme,
				DiffTokens:     themeDiff,
			}
		}

		// Extract components from base dictionary
		components, err := baseDict.ExtractComponents()
		if err != nil {
			return fmt.Errorf("failed to extract components: %w", err)
		}

		// Extract tokens with $property for @property declarations
		propertyTokens := tokens.ExtractPropertyTokens(baseDict, resolvedBase)

		// Build generation context
		ctx := &generators.GenerationContext{
			BaseDict:       baseDict,
			ResolvedTokens: resolvedBase,
			Components:     components,
			Themes:         themeContexts,
			PropertyTokens: propertyTokens,
		}

		// Generate complete CSS
		content, err = gen.Generate(ctx)
		if err != nil {
			return fmt.Errorf("tailwind generation failed: %w", err)
		}

	case "catalog":
		gen := generators.NewCatalogGenerator()
		components, err := baseDict.ExtractComponents()
		if err != nil {
			return fmt.Errorf("failed to extract components: %w", err)
		}

		// Build theme inputs for catalog
		var catalogThemes map[string]generators.CatalogThemeInput
		if len(themes) > 0 {
			// Resolve theme inheritance chains (handles $extends)
			inheritedThemes, err := tokens.ResolveThemeInheritance(baseDict, themes)
			if err != nil {
				return fmt.Errorf("failed to resolve theme inheritance: %w", err)
			}

			catalogThemes = make(map[string]generators.CatalogThemeInput)
			for name, mergedDict := range inheritedThemes {
				// Resolve theme tokens
				themeResolver, err := tokens.NewResolver(mergedDict)
				if err != nil {
					return fmt.Errorf("failed to resolve theme %s: %w", name, err)
				}
				resolvedTheme, err := themeResolver.ResolveAll()
				if err != nil {
					return fmt.Errorf("resolution failed for theme %s: %w", name, err)
				}

				// Calculate diff from base
				themeDiff := tokens.Diff(resolvedTheme, resolvedBase)

				// Extract extends and description from original theme dict
				var extends *string
				var description string
				if originalTheme, ok := themes[name]; ok {
					if extendsVal, ok := originalTheme.Root["$extends"].(string); ok {
						extends = &extendsVal
					}
					if descVal, ok := originalTheme.Root["$description"].(string); ok {
						description = descVal
					}
				}

				catalogThemes[name] = generators.CatalogThemeInput{
					Extends:        extends,
					Description:    description,
					ResolvedTokens: resolvedTheme,
					DiffTokens:     themeDiff,
				}
			}
		}

		content, err = gen.Generate(resolvedBase, components, catalogThemes)
		if err != nil {
			return fmt.Errorf("catalog generation failed: %w", err)
		}

	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	// 4. Write
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	outfile := fmt.Sprintf("%s/tokens.css", outputDir)
	if format == "catalog" {
		outfile = fmt.Sprintf("%s/catalog.json", outputDir)
	}

	if err := os.WriteFile(outfile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("Generated %s\n", outfile)
	return nil
}
