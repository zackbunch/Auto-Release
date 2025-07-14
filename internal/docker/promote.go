package docker

import (
	"fmt"
	"strings"
)

// PromoteImage re-tags a Docker image and pushes it to the registry.
func PromoteImage(opts *BuildOptions, fromEnv, toEnv string) error {
	// Determine the source and target image names
	sourceImage := strings.Replace(opts.FullImage, toEnv, fromEnv, 1)
	targetImage := opts.FullImage

	// Re-tag the image
	if err := RunCMD("docker", "tag", sourceImage, targetImage); err != nil {
		return fmt.Errorf("failed to re-tag image: %w", err)
	}

	// Push the new tag
	return PushImage(opts)
}
