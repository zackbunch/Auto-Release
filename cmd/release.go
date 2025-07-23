package cmd

import (
	"fmt"
	"os"
	"strconv"
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
			mrID, err := resolveMRID(dryRun)
			if err != nil {
				fmt.Printf("[release] Warning: %v. Defaulting to Patch bump.\n", err)
				bumpType = version.Patch
			} else if mrID == "" {
				fmt.Println("[release] Dry run: no MR context available. Defaulting to Patch bump.")
				bumpType = version.Patch
			} else {
				fmt.Printf("[release] Using MR #%s to determine bump.\n", mrID)
				bumpType, err = gitlabClient.MergeRequests.GetVersionBump(mrID)
				if err != nil {
					fmt.Printf("[release] Warning: Failed to determine bump from MR %s: %v. Defaulting to Patch.\n", mrID, err)
					bumpType = version.Patch
				}
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
		mrID, err := resolveMRID(dryRun)
		if err != nil {
			return fmt.Errorf("failed to resolve MR ID: %w", err)
		}
		if mrID == "" {
			fmt.Println("Dry run mode with no MR ID. Defaulting to patch.")
			fmt.Println("patch")
			return nil
		}

		bump, err := gitlabClient.MergeRequests.GetVersionBump(mrID)
		if err != nil {
			return fmt.Errorf("failed to infer bump from MR %s: %w", mrID, err)
		}

		current, next, err := gitlabClient.Tags.GetNextVersion(bump)
		if err != nil {
			return fmt.Errorf("failed to calculate next version: %w", err)
		}

		fmt.Printf("[release] Current: %s\n", current)
		fmt.Printf("[release] Bump:    %s\n", bump)
		fmt.Printf("[release] Next:    %s\n", next)
		fmt.Println(next.String())

		return nil
	},
}

func resolveMRID(dryRun bool) (string, error) {
	if mrID := os.Getenv("CI_MERGE_REQUEST_IID"); mrID != "" {
		return mrID, nil
	}

	ctx, err := ci.LoadContext(dryRun)
	if err == nil && ctx.MRID != "" {
		return ctx.MRID, nil
	}

	if dryRun {
		return "", nil
	}

	mr, err := gitlabClient.MergeRequests.GetLatestMergeRequest()
	if err != nil {
		return "", fmt.Errorf("failed to get latest open merge request: %w", err)
	}
	return strconv.Itoa(mr.IID), nil
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
