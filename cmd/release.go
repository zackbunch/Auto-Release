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
	Short: "Creates a new Git tag and GitLab release based on semantic versioning.",
	Long: `The release command automates the creation of a new semantic version tag
(e.g., v1.2.3) and a corresponding GitLab release.

It determines the next version based on the specified bump type (major, minor, or patch)
and interacts with the GitLab API to create the tag and release entry.

This command is typically used as the final step in the promotion pipeline
to mark a production-ready artifact.`,
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
