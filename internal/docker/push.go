package docker

import (
	"fmt"
	"os"
)

func PushImage(opts *BuildOptions) error {
	if opts.DryRun {
		DryRun("docker", "login", "-u", "<redacted>", "-p", "<redacted>", os.Getenv("CI_REGISTRY"))
		DryRun("docker", "push", opts.FullImage)
		return nil
	}

	// real login + push
	registry := os.Getenv("CI_REGISTRY")
	user := os.Getenv("CI_REGISTRY_USER")
	password := os.Getenv("CI_REGISTRY_PASSWORD")

	if registry == "" || user == "" || password == "" {
		return fmt.Errorf("missing CI_REGISTRY, CI_REGISTRY_USER, or CI_REGISTRY_PASSWORD environment variable")
	}

	if err := RunCMD("docker", "login", "-u", user, "-p", password, registry); err != nil {
		return fmt.Errorf("docker login failed: %w", err)
	}

	return RunCMD("docker", "push", opts.FullImage)
}
