package docker

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	User            string
	Password        string
	Registry        string
	Image           string
	Project         string
	Ref             string
	Dockerfile      string
	ExtraBuildArgs  []string
	ApplicationName string
}

// LoadConfig populates the Config from environment variables and validates required fields
func LoadConfig() (*Config, error) {
	cfg := &Config{
		User:            os.Getenv("CI_REGISTRY_USER"),
		Password:        os.Getenv("CI_REGISTRY_PASSWORD"),
		Registry:        os.Getenv("CI_REGISTRY"),
		Image:           os.Getenv("CI_REGISTRY_IMAGE"),
		Project:         os.Getenv("CI_PROJECT_PATH"),
		Ref:             os.Getenv("CI_COMMIT_REF_NAME"),
		Dockerfile:      os.Getenv("SYAC_DOCKERFILE"),
		ExtraBuildArgs:  parseArgs(os.Getenv("SYAC_BUILD_EXTRA_ARGS")),
		ApplicationName: os.Getenv("SYAC_APPLICATION_NAME"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// internal parseArgs splits strings
func parseArgs(raw string) []string {
	return strings.Fields(raw)
}

// Validate checks for missing required fields
func (c *Config) Validate() error {
	var missing []string

	if c.User == "" {
		missing = append(missing, "CI_REGISTRY_USER")
	}
	if c.Password == "" {
		missing = append(missing, "CI_REGISTRY_PASSWORD")
	}
	if c.Registry == "" {
		missing = append(missing, "CI_REGISTRY")
	}
	if c.Image == "" {
		missing = append(missing, "CI_REGISTRY_IMAGE")
	}
	if c.Dockerfile == "" {
		missing = append(missing, "SYAC_DOCKERFILE")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

// ImageName returns the ApplicationName override if set else defaults to CI_PROJECT_NAME
func (c *Config) ImageName() string {
	if c.ApplicationName != "" {
		return c.ApplicationName
	}
	parts := strings.Split(c.Image, "/")
	return parts[len(parts)-1] // fallback to last part of CI_REGISTRY_IMAGE (usually the project name)
}
