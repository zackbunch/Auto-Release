package cmd

import (
	"fmt"
	"os"

	"syac/internal/ci"
	"syac/internal/docker"
	"syac/pkg/gitlab"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Executes the full SYAC CI/CD pipeline (legacy command).",
	Long: `The run command executes the entire SYAC CI/CD pipeline as a single, monolithic operation.

This command is primarily for backward compatibility. For more granular control and
modern pipeline integration, it is recommended to use the 'build', 'promote', and 'release'
commands individually within your CI/CD configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := ci.LoadEnvFileFromFlag(os.Args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load env: %v\n", err)
			os.Exit(1)
		}

		ctx, err := ci.LoadContext()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load context: %v\n", err)
			os.Exit(1)
		}

		// Create GitLab client
		gitlabClient, err := gitlab.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create GitLab client: %v\n", err)
			os.Exit(1)
		}

		if err := docker.Execute(ctx, gitlabClient); err != nil {
			fmt.Fprintf(os.Stderr, "Error: execution failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
