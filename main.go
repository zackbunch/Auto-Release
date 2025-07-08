package main

import (
	"log"
	"os"
	"syac/internal/ci"
	"syac/internal/docker"
)

func main() {
	if err := ci.LoadEnvFileFromFlag(os.Args); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}
	ctx := ci.LoadContext()

	if err := docker.Execute(ctx); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}
