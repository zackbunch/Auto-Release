package docker

import (
	"fmt"
	"syac/internal/ci"
)

func Execute(ctx ci.Context) error {
	ctx.PrintSummary()

	opts, err := BuildOptionsFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare docker build options: %w", err)
	}

	switch {
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