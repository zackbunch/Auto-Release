package docker

import (
	"fmt"
	"syac/internal/executil"
)

// BuildImage builds one or more Docker images using the provided build options.
// It supports tagging multiple outputs, injecting build args, using a custom Dockerfile,
// and executing in dry-run mode (prints the docker command instead of running it).
func BuildImage(opts *BuildOptions) error {
	// Start constructing the docker build command
	args := []string{"build"}

	// Tag each image (can be multiple for same build output)
	for _, img := range opts.FullImages {
		args = append(args, "-t", img)
	}

	// Add custom Dockerfile if specified
	args = append(args, "-f", opts.Dockerfile)

	// Include all provided --build-arg key=value flags
	for _, arg := range opts.ExtraBuildArgs {
		args = append(args, "--build-arg", arg)
	}

	// Add build context (usually ".")
	args = append(args, opts.ContextPath)

	// Support dry-run execution (no actual build)
	if opts.DryRun {
		executil.DryRunCMD("docker", args...)
		return nil
	}

	// Print image tags being built for visibility
	fmt.Printf("Building images: %v\n", opts.FullImages)
	return executil.RunCMD("docker", args...)
}
