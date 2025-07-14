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
	Short: "Promotes a Docker image from one environment to another by re-tagging and pushing.",
	Long: `The promote command facilitates the movement of an immutable Docker image
through the defined environments (e.g., dev -> test, test -> int).

It re-tags an existing image from a source environment's tag to a target
environment's tag and pushes the newly tagged image to the container registry.

This command embodies the "Promote the Artifact, Not the Code" principle.`,
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
