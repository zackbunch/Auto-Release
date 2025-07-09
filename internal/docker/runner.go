package docker

import (
	"fmt"
	"syac/internal/ci"
	"syac/pkg/gitlab"
)

func Execute(ctx ci.Context, gitlabClient *gitlab.Client) error { // Modified signature
	ctx.PrintSummary()

	opts, err := BuildOptionsFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare docker build options: %w", err)
	}

	switch {
	case ctx.IsMergeRequest: // Re-add merge request handling
		return handleMergeRequest(ctx, opts, gitlabClient)
	case ctx.IsTag:
		return handleTagPush(opts)
	case ctx.IsProtected:
		return handleProtectedBranch(opts)
	case ctx.IsFeatureBranch:
		return handleFeatureBranch(opts)
	default:
		fmt.Println("Unknown context — skipping execution.")
		return nil
	}
}

func handleMergeRequest(ctx ci.Context, opts *BuildOptions, gitlabClient *gitlab.Client) error {
	fmt.Println("Merge request detected. Checking for version bump hint...")

	description, err := gitlabClient.MergeRequests.GetMergeRequestDescription(ctx.MRID)
	if err != nil {
		return fmt.Errorf("failed to get merge request description: %w", err)
	}

	// This part needs to be re-evaluated based on how version bumping will work without gitlab.ParseVersionBumpHint
	// For now, I'll just print the description and proceed with a default tag.
	fmt.Printf("MR Description: %s\n", description)
	fmt.Println("Version bumping logic needs to be re-implemented. Building with default tag...")

	// Fallback to default build for now, as version bumping logic is not yet re-implemented
	return handleFeatureBranch(opts)
}

func handleTagPush(opts *BuildOptions) error {
	fmt.Println("Tag push detected. Building and pushing image...")
	if err := BuildImage(opts); err != nil {
		return err
	}
	if opts.Push {
		return PushImage(opts)
	}
	return nil
}

func handleProtectedBranch(opts *BuildOptions) error {
	fmt.Println("Protected branch push detected. Building and pushing image...")
	if err := BuildImage(opts); err != nil {
		return err
	}
	if opts.Push {
		return PushImage(opts)
	}
	return nil
}

func handleFeatureBranch(opts *BuildOptions) error {
	fmt.Println("Feature branch push detected. Building image...")
	if err := BuildImage(opts); err != nil {
		return err
	}
	if opts.Push {
		return PushImage(opts)
	}
	return nil
}
