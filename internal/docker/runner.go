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
	case ctx.IsMergeRequest:
		return handleMergeRequest(ctx, opts, gitlabClient)
	case ctx.IsTag:
		return handleTagPush(ctx, opts)
	case ctx.RefName == "dev": // Explicitly handle 'dev' branch pushes
		fmt.Println("Push to 'dev' branch detected. Building and pushing image...")
		opts.Push = true // Force push for 'dev' branch
		return buildAndPush(ctx, opts)
	case ctx.RefName == "test" || ctx.RefName == "int": // Handle promotion for test and int
		return handlePromotion(ctx, opts, gitlabClient)
	case ctx.RefName == "master":
		return handleMasterMerge(ctx, gitlabClient)
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

func handleMasterMerge(ctx ci.Context, gitlabClient *gitlab.Client) error {
	fmt.Println("Merge to master detected. Creating release...")

	// Get the merge request associated with the commit
	mr, err := gitlabClient.MergeRequests.GetMergeRequestForCommit(ctx.SHA)
	if err != nil {
		return fmt.Errorf("failed to get merge request for commit %s: %w", ctx.SHA, err)
	}

	// Get the version bump type from the merge request description
	bumpType, err := gitlabClient.MergeRequests.GetVersionBump(fmt.Sprintf("%d", mr.IID))
	if err != nil {
		return fmt.Errorf("failed to get version bump from MR %d: %w", mr.IID, err)
	}

	// Get the next version
	_, nextVersion, err := gitlabClient.Tags.GetNextVersion(bumpType)
	if err != nil {
		return fmt.Errorf("failed to get next version: %w", err)
	}

	tagName := fmt.Sprintf("v%s", nextVersion.String())
	releaseName := fmt.Sprintf("Release %s", tagName)
	releaseDescription := fmt.Sprintf("Automated release for merge request !%d: %s", mr.IID, mr.Title)

	// Create the release
	if err := gitlabClient.Releases.CreateRelease(tagName, ctx.SHA, releaseName, releaseDescription); err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	fmt.Printf("Successfully created release %s\n", tagName)
	return nil
}

func handlePromotion(ctx ci.Context, opts *BuildOptions, gitlabClient *gitlab.Client) error {
	fmt.Printf("Promotion to %s branch detected. Promoting image...\n", ctx.RefName)

	// Get the current commit (the merge commit)
	currentCommit, err := gitlabClient.Commits.GetCommit(ctx.SHA)
	if err != nil {
		return fmt.Errorf("failed to get current commit %s: %w", ctx.SHA, err)
	}

	if len(currentCommit.ParentIDs) < 2 {
		return fmt.Errorf("commit %s is not a merge commit, cannot determine source image for promotion", ctx.SHA)
	}

	// The second parent of a merge commit is the HEAD of the branch that was merged in.
	// This is the SHA we need to find the source image.
	sourceSHA := currentCommit.ParentIDs[1]

	// Determine the source environment and image tag prefix based on the target branch
	var sourceEnv, sourceTagPrefix string
	switch ctx.RefName {
	case "test":
		// Promotion from dev -> test
		sourceEnv = "dev"
		sourceTagPrefix = "rc-"
	case "int":
		// Promotion from test -> int
		sourceEnv = "test"
		sourceTagPrefix = "test-"
	default:
		return fmt.Errorf("unsupported promotion target branch: %s", ctx.RefName)
	}

	// Construct the full source image path using the source environment
	sourceImageBasePath := fmt.Sprintf("%s/%s/%s", ctx.RegistryImage, sourceEnv, opts.ImageName)
	sourceImage := fmt.Sprintf("%s:%s%s", sourceImageBasePath, sourceTagPrefix, sourceSHA[:8])

	// The target image tag should be based on the target branch name and the current (merge) commit SHA
	targetImage := fmt.Sprintf("%s:%s-%s", opts.FullImage[:strings.LastIndex(opts.FullImage, ":")], ctx.RefName, ctx.SHA[:8])

	fmt.Printf("Attempting to promote image from %s to %s\n", sourceImage, targetImage)

	if opts.DryRun {
		DryRun("docker", "pull", sourceImage)
		DryRun("docker", "tag", sourceImage, targetImage)
		DryRun("docker", "push", targetImage)
		return nil
	}

	// Pull the source image
	if err := RunCMD("docker", "pull", sourceImage); err != nil {
		return fmt.Errorf("failed to pull source image %s: %w", sourceImage, err)
	}

	// Tag the image for the new environment
	if err := RunCMD("docker", "tag", sourceImage, targetImage); err != nil {
		return fmt.Errorf("failed to tag image %s as %s: %w", sourceImage, targetImage, err)
	}

	// Push the new image tag
	if err := PushImage(&BuildOptions{
		FullImage: targetImage,
		DryRun:    opts.DryRun,
		Push:      true, // Always push promoted images
	}); err != nil {
		return fmt.Errorf("failed to push promoted image %s: %w", targetImage, err)
	}

	fmt.Printf("Successfully promoted image %s to %s\n", sourceImage, targetImage)

	return nil
}
