package main

import (
	"fmt"
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
	fmt.Println(cfg.Dockerfile)

}
