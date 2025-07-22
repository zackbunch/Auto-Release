package cmd

import (
	"fmt"
	"syac/internal/ci"
	"syac/internal/version"
	"syac/pkg/gitlab"
)

type ReleaseOptions struct {
	DryRun      bool
	Bump        string
	Ref         string
	Name        string
	Description string
	GitLab      *gitlab.Client
}

func RunRelease(opts ReleaseOptions) error {
	if opts.GitLab == nil {
		return fmt.Errorf("GitLab client is not initialized")
	}

	if !opts.DryRun {
		if opts.Ref == "" {
			return fmt.Errorf("--ref is required when not in dry-run mode")
		}
		if len(opts.Ref) < 7 {
			return fmt.Errorf("ref %q is too short to be a valid commit SHA or branch name", opts.Ref)
		}
	}

	var bumpType version.VersionType
	if opts.Bump != "" {
		parsedBumpType, err := version.ParseVersionType(opts.Bump)
		if err != nil {
			return err
		}
		bumpType = parsedBumpType
	} else {
		// Try to get bump type from MR description if not explicitly provided
		ctx, err := ci.LoadContext(opts.DryRun)
		if err != nil {
			return fmt.Errorf("failed to load CI context: %w", err)
		}

		if ctx.MRID != "" {
			mrBumpType, err := opts.GitLab.MergeRequests.GetVersionBump(ctx.MRID)
			if err != nil {
				fmt.Printf("Warning: Failed to get version bump from MR %s description: %v. Defaulting to Patch.\n", ctx.MRID, err)
				bumpType = version.Patch
			} else {
				bumpType = mrBumpType
			}
		} else {
			fmt.Println("Warning: No explicit bump type provided and no MR ID found in CI context. Defaulting to Patch.")
			bumpType = version.Patch
		}
	}

	current, next, err := opts.GitLab.Tags.GetNextVersion(bumpType)
	if err != nil {
		return fmt.Errorf("failed to determine next version: %w", err)
	}

	if opts.DryRun {
		fmt.Printf("[release] Dry run: current=%s next=%s bump=%s\n", current, next, bumpType)
		return nil
	}

	name := opts.Name
	if name == "" {
		name = next.String()
	}

	fmt.Printf("[release] Creating release %s from ref %s\n", name, opts.Ref)

	err = opts.GitLab.Releases.CreateRelease(next.String(), opts.Ref, name, opts.Description)
	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	fmt.Printf("[release] Successfully created release %s\n", name)
	return nil
}
