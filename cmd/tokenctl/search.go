// tokenctl/cmd/tokenctl/search.go
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dmoose/tokenctl/pkg/tokens"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search tokens by name, type, or category",
	Long: `Search tokens in a design system by name pattern, type, or category.

Examples:
  tokenctl search primary              # Find tokens containing "primary"
  tokenctl search --type=color         # List all color tokens
  tokenctl search --category=spacing   # List all spacing tokens
  tokenctl search btn --type=color     # Color tokens containing "btn"

Output includes token path, resolved value, and description (if available).`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSearch,
}

var (
	searchType     string
	searchCategory string
	searchDir      string
)

func init() {
	searchCmd.Flags().StringVarP(&searchType, "type", "t", "", "Filter by token type (color, dimension, number, etc.)")
	searchCmd.Flags().StringVarP(&searchCategory, "category", "c", "", "Filter by category (top-level key)")
	searchCmd.Flags().StringVarP(&searchDir, "dir", "d", ".", "Token directory to search")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := ""
	if len(args) > 0 {
		query = strings.ToLower(args[0])
	}

	// Load tokens
	loader := tokens.NewLoader()
	loader.WarnConflicts = false // Suppress warnings during search

	baseDict, err := loader.LoadBase(searchDir)
	if err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}

	// Extract metadata for rich output
	metadata := tokens.ExtractMetadata(baseDict)

	// Resolve values
	resolver, err := tokens.NewResolver(baseDict)
	if err != nil {
		return fmt.Errorf("failed to create resolver: %w", err)
	}

	resolved, err := resolver.ResolveAll()
	if err != nil {
		return fmt.Errorf("failed to resolve tokens: %w", err)
	}

	// Filter and collect results
	type searchResult struct {
		Path        string
		Value       interface{}
		Type        string
		Description string
	}

	var results []searchResult

	for path, value := range resolved {
		// Skip non-atomic values
		if _, ok := value.(map[string]interface{}); ok {
			continue
		}

		meta := metadata[path]

		// Apply filters
		if !matchesSearch(path, value, meta, query, searchType, searchCategory) {
			continue
		}

		result := searchResult{
			Path:  path,
			Value: value,
		}
		if meta != nil {
			result.Type = meta.Type
			result.Description = meta.Description
		}
		results = append(results, result)
	}

	// Sort by path
	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})

	// Output results
	if len(results) == 0 {
		fmt.Println("No tokens found matching the search criteria.")
		return nil
	}

	fmt.Printf("Found %d token(s):\n\n", len(results))

	for _, r := range results {
		// Format value for display
		valueStr := formatValue(r.Value)

		// Primary line: path and value
		if r.Type != "" {
			fmt.Printf("%s [%s]: %s\n", r.Path, r.Type, valueStr)
		} else {
			fmt.Printf("%s: %s\n", r.Path, valueStr)
		}

		// Description on second line if present
		if r.Description != "" {
			fmt.Printf("  %s\n", r.Description)
		}
		fmt.Println()
	}

	return nil
}

// matchesSearch checks if a token matches the search criteria
func matchesSearch(path string, value interface{}, meta *tokens.TokenMetadata, query, filterType, filterCategory string) bool {
	pathLower := strings.ToLower(path)

	// Query filter (matches path or description)
	if query != "" {
		matchesPath := strings.Contains(pathLower, query)
		matchesDesc := meta != nil && strings.Contains(strings.ToLower(meta.Description), query)
		if !matchesPath && !matchesDesc {
			return false
		}
	}

	// Type filter
	if filterType != "" {
		if meta == nil || !strings.EqualFold(meta.Type, filterType) {
			return false
		}
	}

	// Category filter (first segment of path)
	if filterCategory != "" {
		category := getCategory(path)
		if !matchesCategory(category, filterCategory) {
			return false
		}
	}

	return true
}

// getCategory extracts the first segment of a token path
func getCategory(path string) string {
	if idx := strings.Index(path, "."); idx != -1 {
		return path[:idx]
	}
	return path
}

// matchesCategory checks if a category matches the filter (handles plural/singular)
func matchesCategory(category, filter string) bool {
	category = strings.ToLower(category)
	filter = strings.ToLower(filter)

	if category == filter {
		return true
	}

	// Handle plural/singular
	if strings.HasSuffix(filter, "s") {
		if category == filter[:len(filter)-1] {
			return true
		}
	} else {
		if category == filter+"s" {
			return true
		}
	}

	return false
}

// formatValue converts a value to a display string
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case []interface{}:
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, ", ")
	case float64:
		// Clean up float display
		if v == float64(int(v)) {
			return fmt.Sprintf("%d", int(v))
		}
		return fmt.Sprintf("%g", v)
	default:
		return fmt.Sprintf("%v", value)
	}
}
