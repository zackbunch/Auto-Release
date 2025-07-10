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
		return handleTagPush(ctx, opts)
	case ctx.IsProtected:
		return handleProtectedBranch(ctx, opts)
	case ctx.IsFeatureBranch:
		return handleFeatureBranch(ctx, opts)
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
	fmt.Printf("Detected bump type %s", bumpType)
	current, next, err := gitlabClient.Tags.GetNextVersion(bumpType)
	if err != nil {
		return fmt.Errorf("failed to calculate next version: %w", err)
	}
	fmt.Printf("Next version: %s -> %s", current.String(), next.String())

	// In a merge request, we only calculate and display the proposed version.
	// The actual image build and push happens on the target branch after merge.
	return nil
}

func handleTagPush(ctx ci.Context, opts *BuildOptions) error {
	fmt.Println("Release tag detected. Building and pushing image...")
	return buildAndPush(ctx, opts)
}

func handleProtectedBranch(ctx ci.Context, opts *BuildOptions) error {
	fmt.Println("Protected branch push detected. Building and pushing image...")
	return buildAndPush(ctx, opts)
}

func handleFeatureBranch(ctx ci.Context, opts *BuildOptions) error {
	fmt.Println("Feature branch push detected. Building image...")
	return buildAndPush(ctx, opts)
}

func buildAndPush(ctx ci.Context, opts *BuildOptions) error {
	if err := BuildImage(opts); err != nil {
		return err
	}
	if !opts.Push {
		return nil
	}

	if err := PushImage(opts); err != nil {
		return err
	}

	// For pushes to the default branch or release tags, also push a 'latest' tag
	shouldTagLatest := false
	if ctx.IsProtected && ctx.RefName == ctx.DefaultBranch {
		shouldTagLatest = true
	} else if ctx.IsTag && !strings.Contains(ctx.RefName, "-") { // Avoids tagging pre-releases as 'latest'
		shouldTagLatest = true
	}

	if shouldTagLatest {
		imageParts := strings.Split(opts.FullImage, ":")
		baseImage := imageParts[0]
		latestImage := baseImage + ":latest"

		fmt.Printf("Tagging %s as %s", opts.FullImage, latestImage)

		if opts.DryRun {
			DryRun("docker", "tag", opts.FullImage, latestImage)
			DryRun("docker", "push", latestImage)
		} else {
			if err := RunCMD("docker", "tag", opts.FullImage, latestImage); err != nil {
				return fmt.Errorf("failed to tag image as latest: %w", err)
			}

			fmt.Printf("Pushing latest tag: %s", latestImage)
			if err := RunCMD("docker", "push", latestImage); err != nil {
				return fmt.Errorf("failed to push latest image: %w", err)
			}
		}
	}

	return nil
}
