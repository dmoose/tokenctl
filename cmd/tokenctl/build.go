// tokenctl/cmd/tokenctl/build.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dmoose/tokenctl/v2/pkg/build"
	"github.com/dmoose/tokenctl/v2/pkg/generators"
	"github.com/dmoose/tokenctl/v2/pkg/tokens"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [directory...]",
	Short: "Build token artifacts",
	Long: `Build token artifacts from JSON token definitions.

Multiple directories can be provided and are merged left-to-right.
Later directories extend or override tokens from earlier ones.
See MERGE.md for details on multi-directory merge behavior.

Output formats:
  tailwind          Tailwind CSS 4 with @theme and @layer (default)
  css               Pure CSS without Tailwind import
  catalog           Full JSON catalog for external tools
  manifest:CATEGORY Category-scoped JSON manifest for LLM context
                    Categories: color, spacing, font, size, components, etc.

Flags:
  --customizable-only   Only include tokens marked with $customizable: true
                        Useful for generating LLM manifests of override points
  --strict-unknown-keys Fail instead of warning when the input contains keys
                        tokenctl does not consume (misnamed component
                        sub-blocks, unread $metadata, variants/sizes/states
                        entries with no $class). These warn by default.

Examples:
  tokenctl build ./my-tokens --format=tailwind
  tokenctl build ./base-tokens ./dashboard-tokens
  tokenctl build ./my-tokens --format=manifest:color --customizable-only`,
	Args: cobra.ArbitraryArgs,
	RunE: runBuild,
}

var (
	format            string
	outputDir         string
	customizableOnly  bool
	strictUnknownKeys bool
	generatedAt       string
)

func init() {
	buildCmd.Flags().StringVarP(&format, "format", "f", "tailwind", "Output format (tailwind, css, catalog, manifest:CATEGORY)")
	buildCmd.Flags().StringVarP(&outputDir, "output", "o", "dist", "Output directory")
	buildCmd.Flags().BoolVar(&customizableOnly, "customizable-only", false, "Only include tokens marked $customizable: true (manifest/catalog only)")
	buildCmd.Flags().BoolVar(&strictUnknownKeys, "strict-unknown-keys", false, "Fail the build on input tokenctl does not consume (default: warn)")
	buildCmd.Flags().StringVar(&generatedAt, "generated-at", "", "Stamp meta.generated_at in catalog/manifest output: `now` for the current UTC time, or a literal string. Off by default so the same tokens produce the same bytes.")
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
	dirs := args
	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	formatType, category, err := parseFormat(format)
	if err != nil {
		return err
	}

	fmt.Printf("Building tokens from %s...\n", strings.Join(dirs, ", "))

	var content string
	switch formatType {
	case "tailwind", "css":
		content, err = buildCSS(dirs, formatType)
	case "catalog", "manifest":
		content, err = buildCatalog(dirs, category)
	default:
		return fmt.Errorf("unknown format: %s (valid: tailwind, css, catalog, manifest:CATEGORY)", format)
	}
	if err != nil {
		return err
	}

	return writeOutput(formatType, category, content)
}

func buildCSS(dirs []string, formatType string) (string, error) {
	opts := build.Options{Format: build.FormatCSS, StrictUnknownKeys: strictUnknownKeys}
	if formatType == "tailwind" {
		opts.Format = build.FormatTailwind
	}

	res, err := build.BuildCSSDirs(dirs, opts)
	for _, c := range res.Conflicts {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", c)
	}
	// Name anything the generators will not read, so a misnamed block is
	// visible in the build log rather than as missing styling on a shipped page.
	if auditErr := auditLoaded(strictUnknownKeys, res.Findings); auditErr != nil {
		return "", auditErr
	}
	return res.CSS, err
}

func buildCatalog(dirs []string, category string) (string, error) {
	baseDict, themes, findings, err := loadTokens(dirs...)
	if err != nil {
		return "", err
	}
	if err := auditLoaded(strictUnknownKeys, findings); err != nil {
		return "", err
	}

	resolvedBase, err := resolveTokens(baseDict)
	if err != nil {
		return "", err
	}

	return buildCatalogOutput(category, baseDict, resolvedBase, themes)
}

// buildCatalogOutput generates a JSON catalog or category-scoped manifest.
func buildCatalogOutput(category string, baseDict *tokens.Dictionary, resolvedBase map[string]any, themes map[string]*tokens.Dictionary) (string, error) {
	stamp := generatedAt
	if stamp == "now" {
		stamp = time.Now().UTC().Format(time.RFC3339)
	}
	gen := generators.NewCatalogGeneratorWithOptions(generators.CatalogOptions{
		Category:         category,
		CustomizableOnly: customizableOnly,
		GeneratedAt:      stamp,
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
