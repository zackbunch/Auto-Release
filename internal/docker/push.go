package docker

import (
	"fmt"
	"os"
	"syac/internal/executil"
)

// PushImage authenticates to the GitLab container registry and pushes all image tags in opts.FullImages.
// Uses CI-provided credentials from CI_REGISTRY, CI_REGISTRY_USER, and CI_REGISTRY_PASSWORD or CI_JOB_TOKEN.
// If DryRun is true, commands are printed instead of executed.
func PushImage(opts *BuildOptions) error {
	// --- Registry Auth Setup ---
	registry := os.Getenv("CI_REGISTRY")
	user := os.Getenv("CI_REGISTRY_USER")
	password := os.Getenv("CI_REGISTRY_PASSWORD")

	// Allow fallback to CI_JOB_TOKEN if password isn't explicitly set
	if password == "" {
		password = os.Getenv("CI_JOB_TOKEN")
	}

	// Validate required env vars
	if registry == "" || user == "" {
		return fmt.Errorf("missing CI_REGISTRY or CI_REGISTRY_USER environment variable")
	}
	if password == "" {
		return fmt.Errorf("missing CI_REGISTRY_PASSWORD or CI_JOB_TOKEN environment variable")
	}

	// --- Docker Login ---
	if err := executil.RunCMD("docker", "login", "-u", user, "-p", password, registry); err != nil {
		return fmt.Errorf("docker login failed: %w", err)
	}
	// Always logout after push (non-blocking)
	defer executil.RunCMD("docker", "logout", registry)

	// --- Push Image Tags ---
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
