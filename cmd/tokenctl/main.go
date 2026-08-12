package main

import (
	"fmt"
	"os"

	"github.com/dmoose/tokenctl/pkg/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tokenctl",
	Short: "tokenctl: Semantic Design Tokens Manager",
	Long: `A W3C Design Tokens 2025.10 compliant tool for creating, maintaining,
and generating semantic design systems in Go applications.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		c := version.Commit
		if len(c) > 7 {
			c = c[:7]
		}
		fmt.Printf("tokenctl version %s (%s) built %s\n", version.String(), c, version.BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
