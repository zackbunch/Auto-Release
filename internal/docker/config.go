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
	CommitSHA         string   // CI commit hash, used for tagging dev builds
	Dockerfile        string   // Dockerfile path, defaulting to "Dockerfile"
	ExtraBuildArgs    []string // Optional space-separated build args
	ApplicationName   string   // Optional image name override
	OpenShiftEnv      string   // Derived OpenShift environment (e.g. dev, test, prod)
	ForcePush         bool     // Force image push even for dev builds
	Tag               string   // Final derived tag (commit SHA for dev, semantic version otherwise)
}

// LoadConfig populates the Config from environment variables and validates required fields
func LoadConfig() (*Config, error) {
	ref := os.Getenv("CI_COMMIT_REF_NAME")
	commitSHA := os.Getenv("CI_COMMIT_SHA")
	openShiftEnv := deriveOpenShiftEnv(ref)

	dockerfile := os.Getenv("SYAC_DOCKERFILE")
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	var tag string
	if openShiftEnv == "dev" {
		tag = commitSHA
	} else {
		tag = "0.0.1" // TODO: replace with dynamic versioning logic
	}

	cfg := &Config{
		User:              os.Getenv("CI_REGISTRY_USER"),
		Password:          os.Getenv("CI_REGISTRY_PASSWORD"),
		Registry:          os.Getenv("CI_REGISTRY"),
		RegistryImagePath: os.Getenv("CI_REGISTRY_IMAGE"),
		Project:           os.Getenv("CI_PROJECT_PATH"),
		Ref:               ref,
		CommitSHA:         commitSHA,
		OpenShiftEnv:      openShiftEnv,
		Dockerfile:        dockerfile,
		ExtraBuildArgs:    parseArgs(os.Getenv("SYAC_BUILD_EXTRA_ARGS")),
		ApplicationName:   os.Getenv("SYAC_APPLICATION_NAME"),
		ForcePush:         os.Getenv("SYAC_FORCE_PUSH") == "true",
		Tag:               tag,
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

// parseArgs splits space-separated build args into a slice
func parseArgs(raw string) []string {
	return strings.Fields(raw)
}

// ShouldPush returns whether the image should be pushed to the registry
func (c *Config) ShouldPush() bool {
	return c.OpenShiftEnv != "dev" || c.ForcePush
}

// PrintSummary displays a detailed summary of the loaded configuration.
func (c *Config) PrintSummary() {
	fmt.Println("--------- SYAC Build Configuration ---------")
	fmt.Printf("User:               %s\n", c.User)
	fmt.Printf("Registry:           %s\n", c.Registry)
	fmt.Printf("Registry Image Path:%s\n", c.RegistryImagePath)
	fmt.Printf("Project:            %s\n", c.Project)
	fmt.Printf("Git Ref:            %s\n", c.Ref)
	fmt.Printf("Commit SHA:         %s\n", c.CommitSHA)
	fmt.Printf("Dockerfile:         %s\n", c.Dockerfile)
	fmt.Printf("Application Name:   %s\n", c.ImageName())
	fmt.Printf("OpenShift Env:      %s\n", c.OpenShiftEnv)
	fmt.Printf("Force Push:         %t\n", c.ForcePush)
	fmt.Printf("Final Tag:          %s\n", c.Tag)
	fmt.Printf("Target Image:       %s\n", c.TargetImage(c.Tag))
	fmt.Printf("Extra Build Args:   %v\n", c.ExtraBuildArgs)
	fmt.Println("--------------------------------------------")
}
