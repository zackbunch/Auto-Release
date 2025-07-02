package main

import (
	"log"
	"syac/internal/docker"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	log.Println("Starting SYAC Image Build...")

	cfg, err := docker.LoadConfig()
	if err != nil {
		log.Fatalf("Environment Configuration error: %v", err)
	}

	cfg.PrintSummary()
	log.Printf("Tag resolved to: %s", cfg.Tag)

	if err := docker.BuildImage(cfg); err != nil {
		log.Fatalf("Image build failed: %v", err)
	}

	if cfg.ShouldPush() {
		if err := docker.PushImage(cfg); err != nil {
			log.Fatalf("Image push failed: %v", err)
		}
		log.Printf("Image pushed to: %s", cfg.TargetImage())
	} else {
		log.Println("Skipping image push (dev branch and SYAC_FORCE_PUSH not set)")
	}

	if cfg.EmitMetadata {
		if err := docker.WriteMetadata(cfg); err != nil {
			log.Fatalf("Failed to emit metadata: %v", err)
		}
		log.Println("Metadata written to syac_output.json")
	}
}
