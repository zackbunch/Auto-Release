package cmd

import (
	"fmt"
	"syac/internal/version"
	"syac/pkg/gitlab"
)

// ReleaseOptions encapsulates input to the release command.
type ReleaseOptions struct {
	DryRun      bool
	Bump        string
	Ref         string
	Name        string
	Description string
	GitLab      *gitlab.Client
}

// RunRelease performs a version bump and creates a GitLab release.
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

	bumpType, err := version.ParseVersionType(opts.Bump)
	if err != nil {
		return err
	}

	// Compute next version based on current tags.
	current, next, err := opts.GitLab.Tags.GetNextVersion(bumpType)
	if err != nil {
		return fmt.Errorf("failed to compute next version: %w", err)
	}

	if opts.DryRun {
		fmt.Printf("[release] Dry run:\n  Current: %s\n  Next:    %s\n  Bump:    %s\n", current, next, bumpType)
		return nil
	}

	// Use the computed version if no release name is given.
	name := opts.Name
	if name == "" {
		name = next.String()
	}

	fmt.Printf("[release] Creating release %q from ref %q...\n", name, opts.Ref)
	if err := opts.GitLab.Releases.CreateRelease(next.String(), opts.Ref, name, opts.Description); err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	fmt.Printf("[release] Successfully created release: %s\n", name)
	return nil
}

