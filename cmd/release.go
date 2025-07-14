package cmd

import (
	"fmt"
	"os"

	"syac/internal/version"
	"syac/pkg/gitlab"

	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create a new release",
	Run: func(cmd *cobra.Command, args []string) {
		bump, _ := cmd.Flags().GetString("bump")

		gitlabClient, err := gitlab.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create GitLab client: %v\n", err)
			os.Exit(1)
		}

		currentVersion, nextVersion, err := gitlabClient.Tags.GetNextVersion(version.VersionType(bump))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get next version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Current version: %s, Next version: %s\n", currentVersion, nextVersion)

		if err := gitlabClient.Tags.CreateTag(nextVersion.String(), "master", ""); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create tag: %v\n", err)
			os.Exit(1)
		}

		if err := gitlabClient.Releases.CreateRelease(nextVersion.String(), "master", "", ""); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create release: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	releaseCmd.Flags().String("bump", "patch", "Version bump type (major, minor, patch)")
	rootCmd.AddCommand(releaseCmd)
}
