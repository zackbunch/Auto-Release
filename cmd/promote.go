package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"syac/internal/ci"
	"syac/internal/docker"
)

var promoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "Promote a Docker image to a new environment",
	Run: func(cmd *cobra.Command, args []string) {
		fromEnv, _ := cmd.Flags().GetString("from")
		toEnv, _ := cmd.Flags().GetString("to")

		ctx, err := ci.LoadContext()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load context: %v\n", err)
			os.Exit(1)
		}

		opts, err := docker.BuildOptionsFromContext(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create build options: %v\n", err)
			os.Exit(1)
		}

		if err := docker.PromoteImage(opts, fromEnv, toEnv); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to promote image: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	promoteCmd.Flags().String("from", "", "Source environment")
	promoteCmd.Flags().String("to", "", "Target environment")
	promoteCmd.MarkFlagRequired("from")
	promoteCmd.MarkFlagRequired("to")
	rootCmd.AddCommand(promoteCmd)
}
