// tokenctl/cmd/tokenctl/build.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
func parseFormat(format string) (formatType string, category string, err error) {
	if strings.HasPrefix(format, "manifest:") {
		parts := strings.SplitN(format, ":", 2)
		if len(parts) == 2 {
			cat := parts[1]
			// Sanitize category to prevent path traversal
			if strings.ContainsAny(cat, "/\\") || strings.Contains(cat, "..") {
				return "", "", fmt.Errorf("invalid category %q: must not contain path separators or '..'", cat)
			}
			return "manifest", cat, nil
		}
		return "manifest", "", nil
	}
	return format, "", nil
}

func runBuild(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	fmt.Printf("Building tokens from %s...\n", dir)

	baseDict, themes, err := loadTokens(dir)
	if err != nil {
		return err
	}

	resolvedBase, err := resolveTokens(baseDict)
	if err != nil {
		return err
	}

	formatType, category, err := parseFormat(format)
	if err != nil {
		return err
	}

	var content string
	switch formatType {
	case "tailwind", "css":
		content, err = buildCSSOutput(formatType, baseDict, resolvedBase, themes)
	case "catalog", "manifest":
		content, err = buildCatalogOutput(category, baseDict, resolvedBase, themes)
	default:
		return fmt.Errorf("unknown format: %s (valid: tailwind, css, catalog, manifest:CATEGORY)", format)
	}
	if err != nil {
		return err
	}

	return writeOutput(formatType, category, content)
}

// buildCSSOutput generates Tailwind or pure CSS from resolved tokens and themes.
func buildCSSOutput(formatType string, baseDict *tokens.Dictionary, resolvedBase map[string]any, themes map[string]*tokens.Dictionary) (string, error) {
	inheritedThemes, err := tokens.ResolveThemeInheritance(baseDict, themes)
	if err != nil {
		return "", fmt.Errorf("failed to resolve theme inheritance: %w", err)
	}

	// Build theme contexts (sorted for deterministic error reporting)
	themeContexts := make(map[string]generators.ThemeContext)
	sortedNames := make([]string, 0, len(inheritedThemes))
	for name := range inheritedThemes {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		mergedDict := inheritedThemes[name]
		themeResolver, err := tokens.NewResolver(mergedDict)
		if err != nil {
			return "", fmt.Errorf("failed to resolve theme %s: %w", name, err)
		}
		resolvedTheme, err := themeResolver.ResolveAll()
		if err != nil {
			return "", fmt.Errorf("resolution failed for theme %s: %w", name, err)
		}

		themeContexts[name] = generators.ThemeContext{
			Dict:           mergedDict,
			ResolvedTokens: resolvedTheme,
			DiffTokens:     tokens.Diff(resolvedTheme, resolvedBase),
		}
	}

	components, err := baseDict.ExtractComponents()
	if err != nil {
		return "", fmt.Errorf("failed to extract components: %w", err)
	}

	ctx := &generators.GenerationContext{
		BaseDict:         baseDict,
		ResolvedTokens:   resolvedBase,
		Components:       components,
		Themes:           themeContexts,
		PropertyTokens:   tokens.ExtractPropertyTokens(baseDict, resolvedBase),
		Keyframes:        tokens.ExtractKeyframes(baseDict),
		Breakpoints:      tokens.ExtractBreakpoints(baseDict),
		ResponsiveTokens: tokens.ExtractResponsiveTokens(baseDict),
	}

	if formatType == "css" {
		gen := generators.NewCSSGenerator()
		return gen.Generate(ctx)
	}
	gen := generators.NewTailwindGenerator()
	return gen.Generate(ctx)
}

// buildCatalogOutput generates a JSON catalog or category-scoped manifest.
func buildCatalogOutput(category string, baseDict *tokens.Dictionary, resolvedBase map[string]any, themes map[string]*tokens.Dictionary) (string, error) {
	gen := generators.NewCatalogGeneratorWithOptions(generators.CatalogOptions{
		Category:         category,
		CustomizableOnly: customizableOnly,
	})

	components, err := baseDict.ExtractComponents()
	if err != nil {
		return "", fmt.Errorf("failed to extract components: %w", err)
	}

	metadata := tokens.ExtractMetadata(baseDict)

	var catalogThemes map[string]generators.CatalogThemeInput
	if len(themes) > 0 {
		inheritedThemes, err := tokens.ResolveThemeInheritance(baseDict, themes)
		if err != nil {
			return "", fmt.Errorf("failed to resolve theme inheritance: %w", err)
		}

		catalogThemes = make(map[string]generators.CatalogThemeInput)
		sortedNames := make([]string, 0, len(inheritedThemes))
		for name := range inheritedThemes {
			sortedNames = append(sortedNames, name)
		}
		sort.Strings(sortedNames)

		for _, name := range sortedNames {
			mergedDict := inheritedThemes[name]
			themeResolver, err := tokens.NewResolver(mergedDict)
			if err != nil {
				return "", fmt.Errorf("failed to resolve theme %s: %w", name, err)
			}
			resolvedTheme, err := themeResolver.ResolveAll()
			if err != nil {
				return "", fmt.Errorf("resolution failed for theme %s: %w", name, err)
			}

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
				DiffTokens:     tokens.Diff(resolvedTheme, resolvedBase),
			}
		}
	}

	return gen.GenerateWithMetadata(resolvedBase, components, catalogThemes, metadata)
}

// writeOutput writes generated content to the appropriate output file.
func writeOutput(formatType, category, content string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	var outfile string
	switch formatType {
	case "tailwind", "css":
		outfile = filepath.Join(outputDir, "tokens.css")
	case "catalog":
		outfile = filepath.Join(outputDir, "catalog.json")
	case "manifest":
		if category != "" {
			outfile = filepath.Join(outputDir, fmt.Sprintf("manifest-%s.json", category))
		} else {
			outfile = filepath.Join(outputDir, "manifest.json")
		}
	default:
		outfile = filepath.Join(outputDir, "tokens.css")
	}

	if err := os.WriteFile(outfile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("Generated %s\n", outfile)
	return nil
}
