package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"syac/internal/ci"
	"syac/internal/docker"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Pushes a Docker image to the configured container registry.",
	Long: `The push command takes the Docker image built by 'syac build' and pushes
it to the container registry specified in the CI environment variables.

This command is typically used after a successful build to make the image
available for subsequent promotion or deployment steps.`,Run: func(cmd *cobra.Command, args []string) {
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

		if err := docker.PushImage(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to push image: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}