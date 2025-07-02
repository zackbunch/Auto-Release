package docker

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	User              string   // Docker registry username (from CI_REGISTRY_USER)
	Password          string   // Docker registry password (from CI_REGISTRY_PASSWORD)
	Registry          string   // Base registry hostname (e.g. registry.gitlab.com)
	RegistryImagePath string   // Full image path with group/subgroup (from CI_REGISTRY_IMAGE)
	Project           string   // GitLab project path (e.g. devops/syac/myapp)
	Ref               string   // Git branch/ref name
	Dockerfile        string   // Dockerfile path, defaulting to "Dockerfile"
	ExtraBuildArgs    []string // Optional space-separated build args
	ApplicationName   string   // Optional image name override
	OpenShiftEnv      string   // Derived OpenShift environment (e.g. dev, test, prod)
	ForcePush         bool     // Whether to force push even in dev
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
		ForcePush:         os.Getenv("SYAC_FORCE_PUSH") == "true",
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
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
	return parts[len(parts)-1]
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

// ShouldPush returns true if the image should be pushed to the registry
func (c *Config) ShouldPush() bool {
	return c.OpenShiftEnv != "dev" || c.ForcePush
}

// deriveOpenShiftEnv maps the Git branch/ref to the corresponding OpenShift environment
func deriveOpenShiftEnv(ref string) string {
	switch {
	case ref == "main" || ref == "master":
		return "prod"
	case ref == "test":
		return "test"
	case ref == "int":
		return "int"
	case strings.HasPrefix(ref, "release/"):
		return "stage"
	default:
		return "dev"
	}
}

// parseArgs splits space-separated build args into a slice
func parseArgs(raw string) []string {
	return strings.Fields(raw)
}

// PrintSummary outputs the current config state for debugging
func (c *Config) PrintSummary() {
	fmt.Println("---- SYAC CONFIG ----")
	fmt.Printf("Registry:           %s\n", c.Registry)
	fmt.Printf("Registry Path:      %s\n", c.RegistryImagePath)
	fmt.Printf("Ref:                %s\n", c.Ref)
	fmt.Printf("OpenShift Env:      %s\n", c.OpenShiftEnv)
	fmt.Printf("App Name:           %s\n", c.ImageName())
	fmt.Printf("Dockerfile:         %s\n", c.Dockerfile)
	fmt.Printf("Should Push:        %v\n", c.ShouldPush())
	fmt.Printf("Extra Build Args:   %v\n", c.ExtraBuildArgs)
	fmt.Println("---------------------")
}
