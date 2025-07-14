package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "syac",
	Short: "SYAC (Sprint-Yielded Artifact Control) automates CI/CD workflows and image promotion.",
	Long: `SYAC (Sprint-Yielded Artifact Control) is a command-line tool designed to automate
CI/CD workflows and Docker image promotion based on a sprint-centric model.

It implements the "Promote the Artifact, Not the Code" principle, ensuring
that a single, immutable Docker image is built once and then promoted
through successive environments (dev, test, int, production).

SYAC provides commands for building, promoting, and releasing Docker images,
integrating with GitLab for versioning and release management.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
