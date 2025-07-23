package cmd

import (
	"fmt"
	"syac/internal/ci"
	"syac/internal/version"
	"syac/pkg/gitlab"
)

// ReleaseOptions defines the configuration for performing a release.
// This includes semantic bump type (patch/minor/major), target commit ref,
// and optional release name/description. The GitLab client is required.
type ReleaseOptions struct {
	DryRun      bool           // If true, skip side effects and print what would happen
	Bump        string         // Version bump type: "patch", "minor", or "major"
	Ref         string         // Git SHA or branch name to tag from
	Name        string         // Optional release name (defaults to version string)
	Description string         // Optional GitLab release description
	GitLab      *gitlab.Client // Initialized GitLab API client
}

// RunRelease performs a version bump and creates a GitLab release.
// If DryRun is true, it prints the current and next version but does not tag or release.
//
// This will:
//
//   - Parse the bump type (patch/minor/major)
//   - Look up the latest Git tag
//   - Compute the next version
//   - Use CI commit SHA as the tag ref if not explicitly passed
//   - Create the GitLab release (unless DryRun is true)
//   - Always print the next version to stdout (for CI pipelines to consume)
func RunRelease(opts ReleaseOptions) error {
	if opts.GitLab == nil {
		return fmt.Errorf("GitLab client is not initialized")
	}

	bumpType, err := version.ParseVersionType(opts.Bump)
	if err != nil {
		return err
	}

	// Resolve the ref if not explicitly provided
	if opts.Ref == "" {
		ctx, err := ci.LoadContext(opts.DryRun)
		if err != nil {
			return fmt.Errorf("--ref not provided and failed to load CI context: %w", err)
		}
		if ctx.SHA == "" {
			return fmt.Errorf("unable to determine ref from context (CommitSHA missing)")
		}
		opts.Ref = ctx.SHA
		fmt.Printf("[release] Using commit ref from CI context: %s\n", opts.Ref)
	}

	if len(opts.Ref) < 7 {
		return fmt.Errorf("ref %q is too short to be a valid commit SHA or branch name", opts.Ref)
	}

	// Compute version bump from existing tags
	current, next, err := opts.GitLab.Tags.GetNextVersion(bumpType)
	if err != nil {
		return fmt.Errorf("failed to compute next version: %w", err)
	}

	// Dry-run mode: just print what would happen
	if opts.DryRun {
		fmt.Printf("[release] Dry run:\n  Current: %s\n  Next:    %s\n  Bump:    %s\n", current, next, bumpType)
		return nil
	}

	// Use version as release name if not explicitly provided
	name := opts.Name
	if name == "" {
		name = next.String()
	}

	fmt.Printf("[release] Creating release %q from ref %q...\n", name, opts.Ref)

	if err := opts.GitLab.Releases.CreateRelease(next.String(), opts.Ref, name, opts.Description); err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	fmt.Printf("[release] Successfully created release: %s\n", name)

	// Always print the version last so CI can capture it via $(...)
	fmt.Println(next.String())

	return nil
}
