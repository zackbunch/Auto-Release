package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the application version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("syac version 2.0.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
