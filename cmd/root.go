package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zip",
	Short: "Explore and manipulate zip files",
	Long:  "This is a toy utility to learn Go and work on something fairly low level.",
}

func init() {
	rootCmd.AddCommand(exploreCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
