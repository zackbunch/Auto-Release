package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCMD executes the given command with inherited stdout/stderr
func RunCMD(name string, args ...string) error {
	return run("", false, name, args...)
}

// RunCMDWithDir executes the command in a specific directory
func RunCMDWithDir(dir, name string, args ...string) error {
	return run(dir, false, name, args...)
}

// DryRun logs the command that would be run without executing
func DryRun(name string, args ...string) {
	run("", true, name, args...)
}

// internal runner to consolidate logic for stdout/stderr
func run(dir string, dry bool, name string, args ...string) error {
	fullCmd := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	if dry {
		fmt.Printf("[DRY RUN] %s\n", fullCmd)
		return nil
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running: %s\n", fullCmd)
	return cmd.Run()
}
