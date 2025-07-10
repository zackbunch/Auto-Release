func handlePromotion(ctx ci.Context, opts *BuildOptions, gitlabClient *gitlab.Client) error{
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

	// Determine the source image tag based on the previous environment
	// For dev -> test, source will be rc-<sha>
	// For test -> int, source will be test-<sha>
	var sourceTagPrefix string
	switch ctx.RefName {
	case "test":
		sourceTagPrefix = "rc-"
	case "int":
		sourceTagPrefix = "test-"
	default:
		return fmt.Errorf("unsupported promotion target branch: %s", ctx.RefName)
	}

	sourceImage := fmt.Sprintf("%s:%s%s", opts.FullImage[:strings.LastIndex(opts.FullImage, ":")], sourceTagPrefix, sourceSHA[:8]) // Use short SHA for source tag
	targetImage := fmt.Sprintf("%s:%s-%s", opts.FullImage[:strings.LastIndex(opts.FullImage, ":")], ctx.RefName, ctx.SHA[:8]) // Use short SHA for target tag

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