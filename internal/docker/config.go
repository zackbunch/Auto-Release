package docker

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	User              string
	Password          string
	Registry          string
	RegistryImagePath string
	Project           string
	Ref               string
	Dockerfile        string
	ExtraBuildArgs    []string
	ApplicationName   string
	OpenShiftEnv      string
}

// LoadConfig populates the Config from environment variables and validates required fields
func LoadConfig() (*Config, error) {
	ref := os.Getenv("CI_COMMIT_REF_NAME")
	openShiftEnv := deriveOpenShiftEnv(ref)

	dockerfile := os.Getenv("SYAC_DOCKERFILE")
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	cfg := &Config{
		User:              os.Getenv("CI_REGISTRY_USER"),
		Password:          os.Getenv("CI_REGISTRY_PASSWORD"),
		Registry:          os.Getenv("CI_REGISTRY"),
		RegistryImagePath: os.Getenv("CI_REGISTRY_IMAGE"),
		Project:           os.Getenv("CI_PROJECT_PATH"),
		Ref:               ref,
		OpenShiftEnv:      openShiftEnv,
		Dockerfile:        dockerfile,
		ExtraBuildArgs:    parseArgs(os.Getenv("SYAC_BUILD_EXTRA_ARGS")),
		ApplicationName:   os.Getenv("SYAC_APPLICATION_NAME"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// parseArgs splits space-separated args into a slice
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
	if c.RegistryImagePath == "" {
		missing = append(missing, "CI_REGISTRY_IMAGE")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

// ImageName returns the ApplicationName override if set, else defaults to the project name in CI_REGISTRY_IMAGE
func (c *Config) ImageName() string {
	if c.ApplicationName != "" {
		return c.ApplicationName
	}
	parts := strings.Split(c.RegistryImagePath, "/")
	return parts[len(parts)-1] // fallback to last segment of CI_REGISTRY_IMAGE
}

// deriveOpenShiftEnv maps the Git branch/ref to the corresponding OpenShift environment
func deriveOpenShiftEnv(ref string) string {
	switch ref {
	case "main", "master":
		return "prod"
	case "test":
		return "test"
	case "int":
		return "int"
	default:
		return "dev"
	}
}

// TargetImage returns the full image path with OpenShift environment and tag
func (c *Config) TargetImage(tag string) string {
	path := strings.TrimSuffix(c.RegistryImagePath, "/")
	return fmt.Sprintf("%s/%s/%s:%s",
		path,
		c.OpenShiftEnv,
		c.ImageName(),
		tag,
	)
}
