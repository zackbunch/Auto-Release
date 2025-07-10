package docker

import (
	"fmt"
	"strings"
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

	bumpType, err := gitlabClient.MergeRequests.GetVersionBump(ctx.MRID)
	if err != nil {
		return fmt.Errorf("failed to get version bump from MR notes: %w", err)
	}
	fmt.Printf("Detected bump type %s\n", bumpType)
	current, next, err := gitlabClient.Tags.GetNextVersion(bumpType)
	if err != nil {
		return fmt.Errorf("failed to calculate next version: %w", err)
	}
	fmt.Printf("Next version: %s -> %s", current.String(), next.String())

	// Replace the old tag (which is likely a SHA) with the new semantic version tag.
	imageParts := strings.Split(opts.FullImage, ":")
	imageParts[len(imageParts)-1] = next.String()
	opts.FullImage = strings.Join(imageParts, ":")
	opts.TargetTag = next.String()

	return buildAndPush(opts)
}

func handleTagPush(opts *BuildOptions) error {
	fmt.Println("Tag push detected. Building and pushing image...")
	return buildAndPush(opts)
}

func handleProtectedBranch(opts *BuildOptions) error {
	fmt.Println("Protected branch push detected. Building and pushing image...")
	return buildAndPush(opts)
}

func handleFeatureBranch(opts *BuildOptions) error {
	fmt.Println("Feature branch push detected. Building image...")
	return buildAndPush(opts)
}

func buildAndPush(opts *BuildOptions) error {
	if err := BuildImage(opts); err != nil {
		return err
	}
	if opts.Push {
		return PushImage(opts)
	}
	return nil
}
