package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tokctl",
	Short: "tokctl: Semantic Design Tokens Manager",
	Long: `A W3C Design Tokens 2025.10 compliant tool for creating, maintaining,
and generating semantic design systems in Go applications.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
