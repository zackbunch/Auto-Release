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
	log.Printf("Target image: %s", cfg.TargetImage("0.0.1"))

}
