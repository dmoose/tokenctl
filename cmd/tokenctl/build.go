// tokenctl/cmd/tokenctl/build.go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dmoose/tokenctl/pkg/generators"
	"github.com/dmoose/tokenctl/pkg/tokens"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [directory]",
	Short: "Build token artifacts",
	Long: `Build token artifacts from JSON token definitions.

Output formats:
  tailwind          Tailwind CSS 4 with @theme and @layer (default)
  css               Pure CSS without Tailwind import
  catalog           Full JSON catalog for external tools
  manifest:CATEGORY Category-scoped JSON manifest for LLM context
                    Categories: color, spacing, font, size, components, etc.

Flags:
  --customizable-only   Only include tokens marked with $customizable: true
                        Useful for generating LLM manifests of override points

Examples:
  tokenctl build ./my-tokens --format=tailwind
  tokenctl build ./my-tokens --format=manifest:color
  tokenctl build ./my-tokens --format=manifest:color --customizable-only
  tokenctl build ./my-tokens --format=manifest:components`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBuild,
}

var (
	format           string
	outputDir        string
	customizableOnly bool
)

func init() {
	buildCmd.Flags().StringVarP(&format, "format", "f", "tailwind", "Output format (tailwind, css, catalog, manifest:CATEGORY)")
	buildCmd.Flags().StringVarP(&outputDir, "output", "o", "dist", "Output directory")
	buildCmd.Flags().BoolVar(&customizableOnly, "customizable-only", false, "Only include tokens marked $customizable: true (manifest/catalog only)")
	rootCmd.AddCommand(buildCmd)
}

// parseFormat extracts format type and optional category from format string
// e.g., "manifest:color" returns ("manifest", "color")
func parseFormat(format string) (formatType string, category string) {
	if strings.HasPrefix(format, "manifest:") {
		parts := strings.SplitN(format, ":", 2)
		if len(parts) == 2 {
			return "manifest", parts[1]
		}
		return "manifest", ""
	}
	return format, ""
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
	formatType, category := parseFormat(format)

	switch formatType {
	case "tailwind", "css":
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

		// Extract @keyframes definitions
		keyframes := tokens.ExtractKeyframes(baseDict)

		// Extract breakpoints and responsive tokens
		breakpoints := tokens.ExtractBreakpoints(baseDict)
		responsiveTokens := tokens.ExtractResponsiveTokens(baseDict)

		// Build generation context
		ctx := &generators.GenerationContext{
			BaseDict:         baseDict,
			ResolvedTokens:   resolvedBase,
			Components:       components,
			Themes:           themeContexts,
			PropertyTokens:   propertyTokens,
			Keyframes:        keyframes,
			Breakpoints:      breakpoints,
			ResponsiveTokens: responsiveTokens,
		}

		// Generate CSS using appropriate generator
		if formatType == "css" {
			gen := generators.NewCSSGenerator()
			content, err = gen.Generate(ctx)
			if err != nil {
				return fmt.Errorf("css generation failed: %w", err)
			}
		} else {
			gen := generators.NewTailwindGenerator()
			content, err = gen.Generate(ctx)
			if err != nil {
				return fmt.Errorf("tailwind generation failed: %w", err)
			}
		}

	case "catalog", "manifest":
		// Create generator with optional category and customizable filters
		opts := generators.CatalogOptions{
			Category:         category,
			CustomizableOnly: customizableOnly,
		}
		gen := generators.NewCatalogGeneratorWithOptions(opts)

		components, err := baseDict.ExtractComponents()
		if err != nil {
			return fmt.Errorf("failed to extract components: %w", err)
		}

		// Extract rich metadata from base dictionary
		metadata := tokens.ExtractMetadata(baseDict)

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

		content, err = gen.GenerateWithMetadata(resolvedBase, components, catalogThemes, metadata)
		if err != nil {
			return fmt.Errorf("catalog generation failed: %w", err)
		}

	default:
		return fmt.Errorf("unknown format: %s (valid: tailwind, css, catalog, manifest:CATEGORY)", format)
	}

	// 4. Write
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	// Determine output filename based on format
	var outfile string
	switch formatType {
	case "tailwind", "css":
		outfile = fmt.Sprintf("%s/tokens.css", outputDir)
	case "catalog":
		outfile = fmt.Sprintf("%s/catalog.json", outputDir)
	case "manifest":
		if category != "" {
			outfile = fmt.Sprintf("%s/manifest-%s.json", outputDir, category)
		} else {
			outfile = fmt.Sprintf("%s/manifest.json", outputDir)
		}
	default:
		outfile = fmt.Sprintf("%s/tokens.css", outputDir)
	}

	if err := os.WriteFile(outfile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("Generated %s\n", outfile)
	return nil
}
