package docker

import (
	"fmt"
	"syac/internal/executil"
)

func BuildImage(opts *BuildOptions) error {
	args := []string{
		"build",
	}

	for _, img := range opts.FullImages {
		args = append(args, "-t", img)
	}

	args = append(args, "-f", opts.Dockerfile)

	for _, arg := range opts.ExtraBuildArgs {
		args = append(args, "--build-arg", arg)
	}

	args = append(args, opts.ContextPath)

	if opts.DryRun {
		executil.DryRunCMD("docker", args...)
		return nil
	}

	fmt.Printf("Building images: %v\n", opts.FullImages)
	return executil.RunCMD("docker", args...)
}
