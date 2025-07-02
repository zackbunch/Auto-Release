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
	log.Printf("Loading environment variables...")

	cfg, err := docker.LoadConfig()
	if err != nil {
		log.Fatalf("Environment Configuration error: %v", err)
	}
	log.Println("Environment variables loaded successfully.")

	cfg.PrintSummary()

}
