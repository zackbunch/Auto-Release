package cmd

import (
	"github.com/spf13/cobra"
)

var (
	dryRun     bool
	bump       string
	ref        string
	tagMessage string

	releaseCmd = &cobra.Command{
		Use:   "release",
		Short: "Create a new release.",
		Long:  `This command handles the release process.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ReleaseOptions{
				DryRun:  dryRun,
				Bump:    bump,
				Ref:     ref,
				Message: tagMessage,
				GitLab:  gitlabClient,
			}
			return RunRelease(opts)
		},
	}
)

func init() {
	releaseCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Enable dry run mode.")
	releaseCmd.Flags().StringVar(&bump, "bump", "patch", "Version bump type: major, minor, or patch.")
	releaseCmd.Flags().StringVar(&ref, "ref", "", "The commit SHA or branch to tag from.")
	releaseCmd.Flags().StringVar(&tagMessage, "message", "", "Optional message to annotate the tag.")
	rootCmd.AddCommand(releaseCmd)
}
