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
	Short: "Run the CI/CD pipeline",
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
