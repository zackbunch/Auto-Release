package cmd

import (
	"fmt"
	"syac/internal/version"

	"github.com/spf13/cobra"
)

var (
	dryRun bool
	bump   string
	ref    string
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create a new release.",
	Long:  `This command handles the release process.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !dryRun && ref == "" {
			return fmt.Errorf("--ref is required when not in dry-run mode")
		}

		var bumpType version.VersionType
		switch bump {
		case "major":
			bumpType = version.Major
		case "minor":
			bumpType = version.Minor
		case "patch":
			bumpType = version.Patch
		default:
			return fmt.Errorf("invalid bump type: %s. Please use 'major', 'minor', or 'patch'", bump)
		}

		current, next, err := gitlabClient.Tags.GetNextVersion(bumpType)
		if err != nil {
			return err
		}

		if dryRun {
			fmt.Printf("Dry run enabled. Current version: %s, Next version: %s\n", current, next)
			return nil
		}

		err = gitlabClient.Tags.CreateTag(next.String(), ref, "")
		if err != nil {
			return err
		}

		fmt.Printf("Successfully created release %s\n", next.String())

		return nil
	},
}

func init() {
	releaseCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Enable dry run mode.")
	releaseCmd.Flags().StringVar(&bump, "bump", "patch", "Version bump type: major, minor, or patch.")
	releaseCmd.Flags().StringVar(&ref, "ref", "", "The commit SHA or branch to tag from.")
	rootCmd.AddCommand(releaseCmd)
}
