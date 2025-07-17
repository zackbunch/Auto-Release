package cmd

import (
	"fmt"
	"os"
	"syac/internal/ci"

	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Print a summary of the CI/CD context",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, err := ci.LoadContext(dryRunFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading context: %v\n", err)
			os.Exit(1)
		}
		ctx.PrintSummary(gitlabClient)
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
}
