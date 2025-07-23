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

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release-related commands like creating releases, tagging, or bumping versions.",
}

var releaseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new GitLab release.",
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

var releaseInferCmd = &cobra.Command{
	Use:   "infer-bump",
	Short: "Print the inferred version bump from the MR context.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := ci.LoadContext(dryRun)
		if err != nil {
			return fmt.Errorf("failed to load CI context: %w", err)
		}

		if ctx.MRID == "" {
			return fmt.Errorf("no merge request ID found in CI context")
		}

		bump, err := gitlabClient.MergeRequests.GetVersionBump(ctx.MRID)
		if err != nil {
			return fmt.Errorf("failed to infer bump: %w", err)
		}

		fmt.Println(bump)
		return nil
	},
}

func init() {
	// Common flags for `release create`
	releaseCreateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run without making any changes.")
	releaseCreateCmd.Flags().StringVar(&bump, "bump", "", "Version bump type: patch, minor, or major.")
	releaseCreateCmd.Flags().StringVar(&ref, "ref", "", "Git commit SHA or branch to create the release from.")
	releaseCreateCmd.Flags().StringVar(&releaseName, "name", "", "Optional release name. Defaults to the version.")
	releaseCreateCmd.Flags().StringVar(&releaseDescription, "description", "", "Optional release description.")

	// Optional dry-run for `infer-bump`
	releaseInferCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run without making any changes.")

	// Wire it up
	releaseCmd.AddCommand(releaseCreateCmd)
	releaseCmd.AddCommand(releaseInferCmd)
	rootCmd.AddCommand(releaseCmd)
}
