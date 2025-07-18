package docker

import (
	"fmt"
	"os"
	"syac/internal/executil"
)

// PushImage pushes all images in opts.FullImages. Always pushes unless DryRun is true.
func PushImage(opts *BuildOptions) error {
	// Authenticate to the registry (always)
	registry := os.Getenv("CI_REGISTRY")
	user := os.Getenv("CI_REGISTRY_USER")
	password := os.Getenv("CI_REGISTRY_PASSWORD")
	if registry == "" || user == "" {
		return fmt.Errorf("missing CI_REGISTRY or CI_REGISTRY_USER environment variable")
	}
	if password == "" {
		password = os.Getenv("CI_JOB_TOKEN")
	}
	if password == "" {
		return fmt.Errorf("missing CI_REGISTRY_PASSWORD or CI_JOB_TOKEN environment variable")
	}

	if err := executil.RunCMD("docker", "login", "-u", user, "-p", password, registry); err != nil {
		return fmt.Errorf("docker login failed: %w", err)
	}
	defer executil.RunCMD("docker", "logout", registry)

	// Push each image tag
	for _, image := range opts.FullImages {
		if opts.DryRun {
			executil.DryRunCMD("docker", "push", image)
			continue
		}
		fmt.Printf("⏫ Pushing image: %s\n", image)
		if err := executil.RunCMD("docker", "push", image); err != nil {
			return fmt.Errorf("docker push failed for %s: %w", image, err)
		}
	}

	return nil
}
