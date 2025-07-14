package docker

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"syac/internal/ci"
)

// RollbackImage re-tags a specific image version to the target environment's expected tag
// and pushes it to the registry.
func RollbackImage(ctx ci.Context, targetEnv, targetTag string) error {
	logrus.Infof("Attempting to rollback %s environment to image tag: %s", targetEnv, targetTag)

	// Derive the full image name for the target environment based on the targetTag
	// This assumes the targetTag is a full image name (e.g., myapp:1.2.3) or a simple tag (e.g., 1.2.3)
	// We need to construct the full image path for the target environment.

	// First, get the base image name from the context (e.g., registry.example.com/group/project/app-name)
	appName := os.Getenv("SYAC_APPLICATION_NAME")
	if appName == "" {
		parts := strings.Split(ctx.RegistryImage, "/")
		appName = parts[len(parts)-1]
	}

	// Construct the full image path for the source (the image we are rolling back to)
	sourceFullImage := fmt.Sprintf("%s/%s/%s:%s", ctx.RegistryImage, deriveOpenShiftEnv(ctx), appName, targetTag)

	// Construct the full image path for the destination (the current environment's expected tag)
	// This will be the same as what BuildOptionsFromContext would generate for the targetEnv
	destinationFullImage := fmt.Sprintf("%s/%s/%s:%s", ctx.RegistryImage, targetEnv, appName, targetTag)

	logrus.Infof("Source image for rollback: %s", sourceFullImage)
	logrus.Infof("Destination image for rollback: %s", destinationFullImage)

	// Pull the source image to ensure it exists locally before re-tagging
	logrus.Infof("Pulling source image: %s", sourceFullImage)
	if err := RunCMD("docker", "pull", sourceFullImage); err != nil {
		return fmt.Errorf("failed to pull source image %s: %w", sourceFullImage, err)
	}

	// Re-tag the image
	logrus.Infof("Re-tagging %s to %s", sourceFullImage, destinationFullImage)
	if err := RunCMD("docker", "tag", sourceFullImage, destinationFullImage); err != nil {
		return fmt.Errorf("failed to re-tag image %s to %s: %w", sourceFullImage, destinationFullImage, err)
	}

	// Push the new tag
	logrus.Infof("Pushing re-tagged image: %s", destinationFullImage)
	// We need to create a dummy BuildOptions for PushImage, as it expects it.
	// Only FullImage and DryRun are strictly needed for PushImage.
	dummyOpts := &BuildOptions{
		FullImage: destinationFullImage,
		DryRun:    os.Getenv("SYAC_DRY_RUN") == "true",
	}
	if err := PushImage(dummyOpts); err != nil {
		return fmt.Errorf("failed to push re-tagged image %s: %w", destinationFullImage, err)
	}

	logrus.Infof("Rollback successful for environment %s to tag %s", targetEnv, targetTag)
	return nil
}
