package main

import (
	"log"
	"syac/internal/docker"

	"github.com/joho/godotenv"
)

func main() {
	// Load local .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, continuing with CI/CD environment")
	}

	log.Println("Starting SYAC Image Build Pipeline")

	// Load configuration from environment variables
	cfg, err := docker.LoadConfig()
	if err != nil {
		log.Fatalf("Environment Configuration error: %v", err)
	}

	// Print detailed config for visibility
	cfg.PrintSummary()

	// Build the Docker image using the resolved tag
	if err := docker.BuildImage(cfg, cfg.Tag); err != nil {
		log.Fatalf("Failed to build image: %v", err)
	}
	log.Println("Docker image built successfully")

	// Push if applicable
	if cfg.ShouldPush() {
		if err := docker.PushImage(cfg, cfg.Tag); err != nil {
			log.Fatalf("Failed to push image: %v", err)
		}
		log.Println("Image pushed successfully")
	} else {
		log.Println("Skipping image push (dev branch and SYAC_FORCE_PUSH not set)")
	}
}
