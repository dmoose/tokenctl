package main

import (
	"fmt"
	"os"

	"github.com/dmoose/tokctl/pkg/generators"
	"github.com/dmoose/tokctl/pkg/tokens"
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

	// 3. Generate
	var content string
	switch format {
	case "tailwind":
		gen := generators.NewTailwindGenerator()
		themeGen := generators.NewThemeGenerator()

		// Generate Root Variables
		content, err = gen.Generate(resolvedBase)
		if err != nil {
			return fmt.Errorf("tailwind generation failed: %w", err)
		}

		// Generate Themes
		// We need to resolve each theme merged with base
		resolvedThemes := make(map[string]map[string]interface{})
		for name, themeDict := range themes {
			// Inherit
			merged, err := tokens.Inherit(baseDict, themeDict)
			if err != nil {
				return fmt.Errorf("failed to inherit theme %s: %w", name, err)
			}
			
			// Resolve
			themeResolver, err := tokens.NewResolver(merged)
			if err != nil {
				return fmt.Errorf("failed to resolve theme %s: %w", name, err)
			}
			resolvedTheme, err := themeResolver.ResolveAll()
			if err != nil {
				return fmt.Errorf("resolution failed for theme %s: %w", name, err)
			}
			
			// Calculate Diff: Theme vs Base
			// We only want to output keys that are DIFFERENT or NEW in the theme.
			themeDiff := tokens.Diff(resolvedTheme, resolvedBase)
			
			resolvedThemes[name] = themeDiff
		}

		themeContent, err := themeGen.GenerateThemes(resolvedThemes)
		if err != nil {
			return fmt.Errorf("theme generation failed: %w", err)
		}
		content += "\n" + themeContent

		// Extract Components from Base (Components are shared across themes usually)
		// If components change per theme, that's a more complex scenario.
		// For now, assume components are defined in base.
		components, err := baseDict.ExtractComponents()
		if err != nil {
			return fmt.Errorf("failed to extract components: %w", err)
		}

		// Append components
		compContent, err := gen.GenerateComponents(components)
		if err != nil {
			return fmt.Errorf("component generation failed: %w", err)
		}
		content += "\n" + compContent

	case "catalog":
		gen := generators.NewCatalogGenerator()
		components, err := baseDict.ExtractComponents()
		if err != nil {
			return fmt.Errorf("failed to extract components: %w", err)
		}
		content, err = gen.Generate(resolvedBase, components)
		if err != nil {
			return fmt.Errorf("catalog generation failed: %w", err)
		}

	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
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
