package docker

// Build builds a Docker image with the specified tag, Dockerfile, and context directory.
func Build(tag, dockerfile, contextDir string, dryRun bool) error {
	args := []string{"build", "-t", tag, "-f", dockerfile, contextDir}
	if dryRun {
		DryRun("docker", args...)
		return nil
	}
	return RunCMD("docker", args...)
}

// Tag tags an existing Docker image with a new tag.
func Tag(source, target string, dryRun bool) error {
	args := []string{"tag", source, target}
	if dryRun {
		DryRun("docker", args...)
		return nil
	}
	return RunCMD("docker", args...)
}

// Push pushes a Docker image to a remote registry.
func Push(image string, dryRun bool) error {
	args := []string{"push", image}
	if dryRun {
		DryRun("docker", args...)
		return nil
	}
	return RunCMD("docker", args...)
}

// Login logs into a Docker registry with a username and password.
func Login(registry, username, password string, dryRun bool) error {
	args := []string{"login", registry, "-u", username, "-p", password}
	if dryRun {
		DryRun("docker", args...)
		return nil
	}
	return RunCMD("docker", args...)
}
