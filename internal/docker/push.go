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

	if opts.DryRun {
		executil.DryRunCMD("docker", "login", "-u", "<redacted>", "-p", "<redacted>", os.Getenv("CI_REGISTRY"))
		executil.DryRunCMD("docker", "push", opts.FullImage)
		return nil
	}

	// real login + push
	registry := os.Getenv("CI_REGISTRY")
	user := os.Getenv("CI_REGISTRY_USER")
	password := os.Getenv("CI_REGISTRY_PASSWORD")

	if registry == "" || user == "" || password == "" {
		return fmt.Errorf("missing CI_REGISTRY, CI_REGISTRY_USER, or CI_REGISTRY_PASSWORD environment variable")
	}

	if err := executil.RunCMD("docker", "login", "-u", user, "-p", password, registry); err != nil {
		return fmt.Errorf("docker login failed: %w", err)
	}

	return executil.RunCMD("docker", "push", opts.FullImage)
}
