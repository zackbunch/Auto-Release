package cmd

import (
	"fmt"
	"syac/internal/ci"
	"syac/internal/version"

	"github.com/spf13/cobra"
)

var (
	dryRun             bool
	bump               string
	ref                string
	releaseName        string
	releaseDescription string
)

// releaseCmd triggers the release process using the specified options.
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create a new release.",
	Long: `Creates a new GitLab release by determining the next version 
(either explicitly via --bump or inferred from the MR context), tagging the commit, 
and publishing the release with optional name and description.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var bumpType version.VersionType
		var err error

		if bump != "" {
			bumpType, err = version.ParseVersionType(bump)
			if err != nil {
				return err
			}
		} else {
			ctx, err := ci.LoadContext(dryRun)
			if err != nil {
				return fmt.Errorf("failed to load CI context: %w", err)
			}

			if ctx.MRID != "" {
				bumpType, err = gitlabClient.MergeRequests.GetVersionBump(ctx.MRID)
				if err != nil {
					fmt.Printf("[release] Warning: Failed to determine bump from MR %s: %v. Defaulting to Patch.\n", ctx.MRID, err)
					bumpType = version.Patch
				}
			} else {
				fmt.Println("[release] Warning: No MR ID found. Defaulting to Patch bump.")
				bumpType = version.Patch
			}
		}

		opts := ReleaseOptions{
			DryRun:      dryRun,
			Bump:        bumpType.String(),
			Ref:         ref,
			Name:        releaseName,
			Description: releaseDescription,
			GitLab:      gitlabClient,
		}
		return RunRelease(opts)
	},
}

func init() {
	releaseCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run without making any changes.")
	releaseCmd.Flags().StringVar(&bump, "bump", "", "Version bump type: patch, minor, or major. If not provided, it will be determined from the merge request.")
	releaseCmd.Flags().StringVar(&ref, "ref", "", "Git commit SHA or branch to create the release from.")
	releaseCmd.Flags().StringVar(&releaseName, "name", "", "Optional release name. Defaults to the version.")
	releaseCmd.Flags().StringVar(&releaseDescription, "description", "", "Optional release description.")

	rootCmd.AddCommand(releaseCmd)
}
