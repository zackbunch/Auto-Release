package cmd

import (
	"fmt"
	"os"

	"syac/pkg/gitlab"

	"github.com/spf13/cobra"
)

var (
	gitlabClient *gitlab.Client

	rootCmd = &cobra.Command{
		Use:   "syac",
		Short: "SYAC (Sprint-Yielded Artifact Control) automates CI/CD workflows and image promotion.",
		Long: `SYAC (Sprint-Yielded Artifact Control) is a command-line tool designed to automate
CI/CD workflows and Docker image promotion based on a sprint-centric model.

It follows the principle of "Promote the Artifact, Not the Code", ensuring that a
single, immutable Docker image is built once and promoted across dev, test, int, and prod
environments.

SYAC provides commands for building, promoting, and releasing Docker images,
and integrates with GitLab for versioning and release management.`,
	}
)

func init() {
	rootCmd.PersistentFlags().Bool("dry-run", false, "Simulate execution without making any changes")
}

// Execute is the entrypoint for the CLI
func Execute() {
	var err error
	gitlabClient, err = gitlab.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing GitLab client: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Command failed: %v\n", err)
		os.Exit(1)
	}
}
