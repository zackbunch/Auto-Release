package docker

import (
	"fmt"
	"os"
	"os/exec"
)

// BuildImage builds a Docker image with the specified tag
func BuildImage(cfg *Config, tag string) error {
	image := cfg.TargetImage(tag)

	args := []string{
		"build",
		"-t", image,
		"-f", cfg.Dockerfile,
		".", // build context is root of project
	}

	if len(cfg.ExtraBuildArgs) > 0 {
		args = append(args[:len(args)-1], append(cfg.ExtraBuildArgs, args[len(args)-1])...) // insert before context
	}

	fmt.Printf("\n[SYAC] Building image: %s\n", image)
	fmt.Printf("[SYAC] Dockerfile: %s\n", cfg.Dockerfile)
	fmt.Printf("[SYAC] Build args: %v\n", cfg.ExtraBuildArgs)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// PushImage pushes the built Docker image to the registry
func PushImage(cfg *Config, tag string) error {
	image := cfg.TargetImage(tag)

	if !cfg.ShouldPush() {
		fmt.Printf("\n[SYAC] Skipping push for image %s (env = dev, no force push)\n", image)
		return nil
	}

	fmt.Printf("\n[SYAC] Pushing image: %s\n", image)
	cmd := exec.Command("docker", "push", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
