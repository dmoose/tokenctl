package main

import (
	"fmt"
	"os"

	"github.com/dmoose/tokenctl/pkg/tokens"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [directory]",
	Short: "Validate the token system integrity",
	Long: `Validate tokens for type correctness, reference integrity, and optionally layer rules.

Layer validation (--strict-layers) enforces design system architecture:
  - brand layer: Can only use raw values (no references)
  - semantic layer: Can reference brand tokens
  - component layer: Can only reference semantic tokens

Example token structure with layers:
  {
    "color": {
      "$layer": "brand",
      "blue-500": { "$value": "#3b82f6" }
    },
    "semantic": {
      "$layer": "semantic",
      "primary": { "$value": "{color.blue-500}" }
    }
  }`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

var (
	strictMode   bool
	strictLayers bool
)

func init() {
	validateCmd.Flags().BoolVar(&strictMode, "strict", false, "Fail on warnings")
	validateCmd.Flags().BoolVar(&strictLayers, "strict-layers", false, "Enforce layer reference rules (brand -> semantic -> component)")
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	fmt.Printf("Validating token system in %s...\n", dir)

	// 1. Load Dictionary (Base + Themes)
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

	validator := tokens.NewValidator()
	hasErrors := false

	// 2. Validate Base
	fmt.Println("Checking Base Dictionary...")
	errs, err := validator.Validate(baseDict)
	if err != nil {
		return fmt.Errorf("base validation failed to run: %w", err)
	}
	if len(errs) > 0 {
		hasErrors = true
		for _, e := range errs {
			fmt.Printf("  [Error] %s\n", e)
		}
	} else {
		fmt.Println("  OK")
	}

	// 3. Validate Themes (Inheritance + Resolution)
	// Resolve theme inheritance chains (handles $extends)
	inheritedThemes, err := tokens.ResolveThemeInheritance(baseDict, themes)
	if err != nil {
		return fmt.Errorf("theme inheritance failed: %w", err)
	}

	for name, merged := range inheritedThemes {
		fmt.Printf("Checking Theme '%s'...\n", name)

		errs, err := validator.Validate(merged)
		if err != nil {
			return fmt.Errorf("theme validation failed to run: %w", err)
		}
		if len(errs) > 0 {
			hasErrors = true
			for _, e := range errs {
				fmt.Printf("  [Error] %s\n", e)
			}
		} else {
			fmt.Println("  OK")
		}
	}

	// 4. Layer Validation (if --strict-layers)
	if strictLayers {
		fmt.Println("Checking Layer Rules...")
		layerValidator := tokens.NewLayerValidator(baseDict)
		violations := layerValidator.ValidateReferences(baseDict)

		if len(violations) > 0 {
			hasErrors = true
			for _, v := range violations {
				fmt.Printf("  [Error] %s\n", v)
			}
		} else {
			fmt.Println("  OK")
		}
	}

	if hasErrors {
		os.Exit(1)
	}

	fmt.Println("\nValidation Passed!")
	return nil
}
