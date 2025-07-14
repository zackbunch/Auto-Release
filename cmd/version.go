package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the SYAC CLI version information.",
	Long: `The version command prints the current version of the SYAC command-line tool.

This is useful for verifying the installed version and for debugging purposes.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("syac version 2.0.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
