package docker

import (
	"fmt"
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

	fmt.Printf("Building image: %s\n", opts.FullImage)
	return RunCMD("docker", args...)
}
