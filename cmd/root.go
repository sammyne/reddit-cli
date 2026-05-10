// Package cmd defines the CLI commands for the reddit binary.
package cmd

import (
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "reddit",
	Short: "reddit — Reddit in your terminal",
}

func init() {
	rootCmd.Version = fullVersion()
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
