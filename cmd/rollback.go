package cmd

import (
	"fmt"
	"os"

	"syac/internal/ci"
	"syac/internal/docker"

	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rolls back a deployed Docker image to a previous version in a specified environment.",
	Long: `The rollback command allows you to revert an environment's deployed image
to a previously known good version.

You must specify the target environment (e.g., 'test', 'int', 'prod') and the
exact image tag (e.g., 'rc-abcdef1', 'test-1234567', '1.2.3') to roll back to.

This command re-tags the specified historical image with the current environment's
expected tag and pushes it to the registry, making it available for re-deployment.`, Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Flags().GetString("environment")
		tag, _ := cmd.Flags().GetString("target-tag")

		if env == "" || tag == "" {
			fmt.Fprintf(os.Stderr, "Error: --environment and --target-tag are required.\n")
			os.Exit(1)
		}

		ctx, err := ci.LoadContext()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load context: %v\n", err)
			os.Exit(1)
		}

		if err := docker.RollbackImage(ctx, env, tag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: rollback failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rollbackCmd.Flags().StringP("environment", "e", "", "The target environment to rollback (e.g., test, int, prod)")
	rollbackCmd.Flags().StringP("target-tag", "t", "", "The specific image tag to rollback to (e.g., rc-abcdef1, 1.2.3)")
	rollbackCmd.MarkFlagRequired("environment")
	rollbackCmd.MarkFlagRequired("target-tag")
	rootCmd.AddCommand(rollbackCmd)
}
