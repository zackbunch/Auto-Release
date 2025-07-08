package main

import (
	"fmt"
	"os"

	"syac/internal/ci"
	"syac/internal/docker"
)

func main() {
	if err := ci.LoadEnvFileFromFlag(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load env: %v\n", err)
		os.Exit(1)
	}

	ctx, err := ci.LoadContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load context: %v\n", err)
		os.Exit(1)
	}

	if err := docker.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: execution failed: %v\n", err)
		os.Exit(1)
	}
}
