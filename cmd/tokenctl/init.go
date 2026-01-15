package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize a new token system",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	fmt.Printf("Initializing new semantic token system in %s...\n", dir)

	// Create directory structure
	dirs := []string{
		"tokens/brand",
		"tokens/surface",
		"tokens/semantic",
		"tokens/typography",
		"tokens/spacing",
		"tokens/themes",
	}

	for _, d := range dirs {
		path := filepath.Join(dir, d)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}

	// Create default token files
	defaults := map[string]any{
		"tokens/brand/colors.json": map[string]any{
			"color": map[string]any{
				"brand": map[string]any{
					"$type":        "color",
					"$description": "Core brand identity colors",
					"primary": map[string]any{
						"$value": "#3b82f6",
					},
					"secondary": map[string]any{
						"$value": "#8b5cf6",
					},
				},
			},
		},
		"tokens/semantic/status.json": map[string]any{
			"color": map[string]any{
				"status": map[string]any{
					"$type": "color",
					"success": map[string]any{
						"$value": "#10b981",
					},
					"error": map[string]any{
						"$value": "#ef4444",
					},
					"warning": map[string]any{
						"$value": "#f59e0b",
					},
				},
			},
		},
		"tokens/spacing/scale.json": map[string]any{
			"spacing": map[string]any{
				"$type": "dimension",
				"sm": map[string]any{
					"$value": "0.5rem",
				},
				"md": map[string]any{
					"$value": "1rem",
				},
				"lg": map[string]any{
					"$value": "1.5rem",
				},
			},
		},
	}

	for path, content := range defaults {
		fullPath := filepath.Join(dir, path)
		f, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
		}
		defer func() { _ = f.Close() }()

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(content); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}
		fmt.Printf("Created %s\n", fullPath)
	}

	fmt.Println("Done! You can now run 'tokenctl validate' to check your tokens.")
	return nil
}
