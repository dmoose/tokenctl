package main

import (
	"fmt"
	"os"

	"github.com/dmoose/tokctl/pkg/tokens"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [directory]",
	Short: "Validate the token system integrity",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runValidate,
}

var strictMode bool

func init() {
	validateCmd.Flags().BoolVar(&strictMode, "strict", false, "Fail on warnings")
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
	for name, themeDict := range themes {
		fmt.Printf("Checking Theme '%s'...\n", name)
		
		// Inherit to check validity of the final result
		merged, err := tokens.Inherit(baseDict, themeDict)
		if err != nil {
			fmt.Printf("  [Error] Inheritance failed: %v\n", err)
			hasErrors = true
			continue
		}

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

	if hasErrors {
		os.Exit(1)
	}

	fmt.Println("\nValidation Passed!")
	return nil
}
