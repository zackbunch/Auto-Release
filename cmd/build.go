package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"syac/internal/ci"
	"syac/internal/docker"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a Docker image",
	Run: func(cmd *cobra.Command, args []string) {
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

		if err := docker.BuildImage(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to build image: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
