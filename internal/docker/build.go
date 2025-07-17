package docker

import (
	"fmt"
	"syac/internal/executil"
)

func BuildImage(opts *BuildOptions) error {
	args := []string{
		"build",
		"-t", opts.FullImage,
		"-f", opts.Dockerfile,
	}

	for _, arg := range opts.ExtraBuildArgs {
		args = append(args, "--build-arg", arg)
	}

	args = append(args, opts.ContextPath)

	if opts.DryRun {
		executil.DryRunCMD("docker", args...)
		return nil
	}

	fmt.Printf("Building image: %s\n", opts.FullImage)
	return executil.RunCMD("docker", args...)
}
