package docker

import (
	"fmt"
	"os"

	"syac/internal/executil"
)

// PushImage authenticates to the GitLab container registry and pushes all image tags in opts.FullImages.
// Supports dry-run mode for safe testing.
func PushImage(opts *BuildOptions) error {
	registry := os.Getenv("CI_REGISTRY")
	user := os.Getenv("CI_REGISTRY_USER")
	password := os.Getenv("CI_REGISTRY_PASSWORD")
	if password == "" {
		password = os.Getenv("CI_JOB_TOKEN")
	}

	if registry == "" || user == "" {
		return fmt.Errorf("missing CI_REGISTRY or CI_REGISTRY_USER environment variable")
	}
	if password == "" {
		return fmt.Errorf("missing CI_REGISTRY_PASSWORD or CI_JOB_TOKEN environment variable")
	}

	// Perform login
	if err := LoginToRegistry(registry, user, password, opts.DryRun); err != nil {
		return fmt.Errorf("docker login failed: %w", err)
	}
	if !opts.DryRun {
		defer LogoutFromRegistry(registry)
	}

	// Push each image
	for _, image := range opts.FullImages {
		if opts.DryRun {
			fmt.Printf("DRY RUN: docker push %s\n", image)
			executil.DryRunCMD("docker", "push", image)
			continue
		}

		fmt.Printf("Pushing image: %s\n", image)
		if err := executil.RunCMD("docker", "push", image); err != nil {
			return fmt.Errorf("docker push failed for %s: %w", image, err)
		}
	}

	return nil
}

// LoginToRegistry performs a docker login using the provided credentials.
// In dry-run mode, it prints the command without executing it.
func LoginToRegistry(registry, user, password string, dryRun bool) error {
	if dryRun {
		fmt.Printf("DRY RUN: docker login -u %s %s\n", user, registry)
		return nil
	}
	return executil.RunCMD("docker", "login", "-u", user, "-p", password, registry)
}

// LogoutFromRegistry performs a docker logout from the given registry.
// This is a best-effort cleanup and does not error if logout fails.
func LogoutFromRegistry(registry string) {
	if err := executil.RunCMD("docker", "logout", registry); err != nil {
		fmt.Fprintf(os.Stderr, "warning: docker logout failed: %v\n", err)
	}
}
