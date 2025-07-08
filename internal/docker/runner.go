package docker

import (
	"fmt"
	"syac/internal/ci"
	"syac/internal/gitlab"
)

func Execute(ctx ci.Context) error {
	ctx.PrintSummary()

	opts, err := BuildOptionsFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare docker build options: %w", err)
	}

	switch {
	case ctx.IsMergeRequest:
		return handleMergeRequest(ctx, opts)
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

func handleMergeRequest(ctx ci.Context, opts *BuildOptions) error {
	fmt.Println("Merge request detected. Checking for version bump hint...")

	client, err := gitlab.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create gitlab client: %w", err)
	}

	description, err := client.GetMergeRequestDescription(ctx.MRID)
	if err != nil {
		return fmt.Errorf("failed to get merge request description: %w", err)
	}

	bumpHint := gitlab.ParseVersionBumpHint(description)
	if bumpHint == "" {
		fmt.Println("No version bump hint found. Building with default tag...")
		return handleFeatureBranch(opts) // Fallback to default build
	}

	fmt.Printf("Version bump hint found: %s. Calculating new version...\n", bumpHint)

	latestTag, err := client.GetLatestTag()
	if err != nil {
		return fmt.Errorf("failed to get latest tag: %w", err)
	}

	newVersion := latestTag.Inc(bumpHint)
	opts.TargetTag = newVersion.String()
	opts.FullImage = fmt.Sprintf("%s/%s/%s:%s", opts.FullImage, deriveOpenShiftEnv(ctx.RefName), opts.ImageName, opts.TargetTag)

	fmt.Printf("New version: %s. Building and pushing image...\n", newVersion)

	if err := BuildImage(opts); err != nil {
		return err
	}
	if opts.Push {
		return PushImage(opts)
	}
	return nil
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
