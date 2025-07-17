package docker

import (
	"fmt"
	"os"
	"syac/internal/executil"
)

func PushImage(opts *BuildOptions) error {
	if !opts.Push && !opts.DryRun {
		// If not set to push and not a dry run, do nothing.
		return nil
	}

	// real login (if not dry run)
	if !opts.DryRun {
		registry := os.Getenv("CI_REGISTRY")
		user := os.Getenv("CI_REGISTRY_USER")
		password := os.Getenv("CI_REGISTRY_PASSWORD")

		if registry == "" || user == "" || password == "" {
			return fmt.Errorf("missing CI_REGISTRY, CI_REGISTRY_USER, or CI_REGISTRY_PASSWORD environment variable")
		}

		if err := executil.RunCMD("docker", "login", "-u", user, "-p", password, registry); err != nil {
			return fmt.Errorf("docker login failed: %w", err)
		}
	}

	for _, fullImage := range opts.FullImages {
		if opts.DryRun {
			executil.DryRunCMD("docker", "push", fullImage)
		} else {
			if err := executil.RunCMD("docker", "push", fullImage); err != nil {
				return fmt.Errorf("docker push failed for %s: %w", fullImage, err)
			}
		}
	}

	return nil
}
