// Package cmd defines the CLI commands for the reddit binary.
package cmd

import (
	"github.com/spf13/cobra"
)

const version = "0.4.1"

var verbose bool

var rootCmd = &cobra.Command{
	Use:     "reddit",
	Short:   "reddit — Reddit in your terminal",
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
