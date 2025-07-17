package cmd

import (
	"fmt"
	"os"

	"syac/internal/ci"
	"syac/internal/docker"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Pushes a Docker image to the registry.",
	Long: `The push command pushes a Docker image to the configured container registry.
It uses parameters derived from the CI environment variables and SYAC's internal configuration.`,
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

		if err := docker.PushImage(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to push image: %v", err)
			os.Exit(1)
		}

		fmt.Println("Image push successful.")
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
