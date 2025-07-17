package cmd

import (
	"fmt"
	"os"

	"syac/internal/ci"
	"syac/internal/docker"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds a Docker image based on CI context and SYAC configuration.",
	Long: `The build command constructs a Docker image using parameters derived from
the CI environment variables and SYAC's internal configuration.

It automatically determines the image name, tag (e.g., rc-<sha> for dev branch),
Dockerfile path, and build context.

This command is typically used early in the CI pipeline to create the initial
immutable artifact.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, err := ci.LoadContext(dryRunFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load context: %v", err)
			os.Exit(1)
		}

	

		opts, err := docker.BuildOptionsFromContext(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create build options: %v", err)
			os.Exit(1)
		}

		if err := docker.BuildImage(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to build image: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
