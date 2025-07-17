package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev" // overridden at build time

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the SYAC CLI version information.",
	Long:  `The version command prints the current version of the SYAC command-line tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("syac version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
