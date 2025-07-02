package docker

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	User                string   // Docker registry username (from CI_REGISTRY_USER)
	Password            string   // Docker registry password (from CI_REGISTRY_PASSWORD)
	Registry            string   // Base registry hostname (e.g. registry.gitlab.com)
	RegistryImagePath   string   // Full image path with group/subgroup (from CI_REGISTRY_IMAGE)
	Project             string   // GitLab project path (e.g. devops/syac/myapp)
	Ref                 string   // Git branch/ref name
	Dockerfile          string   // Dockerfile path, defaulting to "Dockerfile"
	ExtraBuildArgs      []string // Optional space-separated build args
	ApplicationName     string   // Optional image name override
	OpenShiftEnv        string   // Derived OpenShift environment (e.g. dev, test, prod)
	ForcePush           bool     // Force image push in dev
	Sprint              string   // Sprint number for release candidate tags
	Tag                 string   // The final tag to use for image
	InMergeRequest      bool     // Whether the current pipeline is running in a merge request
	FullSHA             string   // Full Git commit SHA
	PipelineID          string   // GitLab CI pipeline ID
	JobID               string   // GitLab CI job ID
	EmitMetadata        bool     // Whether to emit syac_output.json
	RequestedSemverBump string   // Optional semver bump type: patch, minor, major
	BuildContext        string   // Optional build context path

	// Additional fields
	MetadataFilePath string // Optional override for metadata output path
	IsCI             bool   // True if running inside GitLab CI
	SourceBranch     string // MR source branch (if applicable)
	TargetBranch     string // MR target branch (if applicable)
	ServiceName      string // Optional logical name of the service
	BuildID          string // Derived from PipelineID-JobID
	ChangelogPath    string // Optional path to changelog file
}

// LoadConfig populates the Config from environment variables and validates required fields
func LoadConfig() (*Config, error) {
	ref := os.Getenv("CI_COMMIT_REF_NAME")
	openShiftEnv := deriveOpenShiftEnv(ref)

	dockerfile := os.Getenv("SYAC_DOCKERFILE")
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	shortSHA := os.Getenv("CI_COMMIT_SHORT_SHA")
	sprint := os.Getenv("SYAC_SPRINT")
	mrTarget := os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME")
	inMR := os.Getenv("CI_MERGE_REQUEST_ID") != ""

	tag := determineTag(ref, mrTarget, sprint, shortSHA, inMR)

	fullSHA := os.Getenv("CI_COMMIT_SHA")
	pipelineID := os.Getenv("CI_PIPELINE_ID")
	jobID := os.Getenv("CI_JOB_ID")
	emitMetadata := os.Getenv("SYAC_EMIT_METADATA") != "false" // default true
	requestedBump := os.Getenv("SYAC_SEMVER_BUMP")
	buildContext := os.Getenv("SYAC_BUILD_CONTEXT")
	if buildContext == "" {
		buildContext = "." // default to current directory
	}

	// New fields
	metadataFilePath := os.Getenv("SYAC_METADATA_PATH")
	isCI := os.Getenv("GITLAB_CI") == "true"
	sourceBranch := os.Getenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME")
	targetBranch := mrTarget
	serviceName := os.Getenv("SYAC_SERVICE_NAME")
	changelogPath := os.Getenv("SYAC_CHANGELOG_PATH")
	buildID := fmt.Sprintf("%s-%s", pipelineID, jobID)

	cfg := &Config{
		User:                os.Getenv("CI_REGISTRY_USER"),
		Password:            os.Getenv("CI_REGISTRY_PASSWORD"),
		Registry:            os.Getenv("CI_REGISTRY"),
		RegistryImagePath:   os.Getenv("CI_REGISTRY_IMAGE"),
		Project:             os.Getenv("CI_PROJECT_PATH"),
		Ref:                 ref,
		OpenShiftEnv:        openShiftEnv,
		Dockerfile:          dockerfile,
		ExtraBuildArgs:      parseArgs(os.Getenv("SYAC_BUILD_EXTRA_ARGS")),
		ApplicationName:     os.Getenv("SYAC_APPLICATION_NAME"),
		ForcePush:           os.Getenv("SYAC_FORCE_PUSH") == "true",
		Sprint:              sprint,
		Tag:                 tag,
		InMergeRequest:      inMR,
		FullSHA:             fullSHA,
		PipelineID:          pipelineID,
		JobID:               jobID,
		EmitMetadata:        emitMetadata,
		RequestedSemverBump: requestedBump,
		BuildContext:        buildContext,
		MetadataFilePath:    metadataFilePath,
		IsCI:                isCI,
		SourceBranch:        sourceBranch,
		TargetBranch:        targetBranch,
		ServiceName:         serviceName,
		BuildID:             buildID,
		ChangelogPath:       changelogPath,
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
func (c *Config) TargetImage() string {
	path := strings.TrimSuffix(c.RegistryImagePath, "/")
	return fmt.Sprintf("%s/%s/%s:%s",
		path,
		c.OpenShiftEnv,
		c.ImageName(),
		c.Tag,
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

// determineTag decides whether to use SHA or RC tag
func determineTag(ref, mrTarget, sprint, shortSHA string, inMR bool) string {
	if ref == "dev" && mrTarget == "dev" && inMR && sprint != "" {
		return fmt.Sprintf("rc.%s", sprint)
	}
	return shortSHA
}

// IsMergeToDev returns true if this is a merge request into the dev branch
func (c *Config) IsMergeToDev() bool {
	return c.InMergeRequest && c.TargetBranch == "dev"
}

// parseArgs splits space-separated build args into a slice
func parseArgs(raw string) []string {
	return strings.Fields(raw)
}

// ShouldPush determines if image should be pushed
func (c *Config) ShouldPush() bool {
	return c.OpenShiftEnv != "dev" || c.ForcePush
}

// PrintSummary logs all relevant config for debugging
func (c *Config) PrintSummary() {
	fmt.Println("Resolved Configuration:")
	fmt.Printf("  Registry:             %s\n", c.Registry)
	fmt.Printf("  RegistryImagePath:    %s\n", c.RegistryImagePath)
	fmt.Printf("  ApplicationName:      %s\n", c.ImageName())
	fmt.Printf("  OpenShiftEnv:         %s\n", c.OpenShiftEnv)
	fmt.Printf("  Dockerfile:           %s\n", c.Dockerfile)
	fmt.Printf("  Tag:                  %s\n", c.Tag)
	fmt.Printf("  Target Image:         %s\n", c.TargetImage())
	fmt.Printf("  ForcePush:            %v\n", c.ForcePush)
	fmt.Printf("  ExtraBuildArgs:       %v\n", c.ExtraBuildArgs)
	fmt.Printf("  InMergeRequest:       %v\n", c.InMergeRequest)
	fmt.Printf("  Full SHA:             %s\n", c.FullSHA)
	fmt.Printf("  Pipeline ID:          %s\n", c.PipelineID)
	fmt.Printf("  Job ID:               %s\n", c.JobID)
	fmt.Printf("  Emit Metadata:        %v\n", c.EmitMetadata)
	fmt.Printf("  Requested Semver Bump:%s\n", c.RequestedSemverBump)
	fmt.Printf("  Build Context:        %s\n", c.BuildContext)
	fmt.Printf("  Build ID:             %s\n", c.BuildID)
	fmt.Printf("  Metadata File Path:   %s\n", c.MetadataFilePath)
	fmt.Printf("  Service Name:         %s\n", c.ServiceName)
	fmt.Printf("  Changelog Path:       %s\n", c.ChangelogPath)
	fmt.Printf("  IsCI:                 %v\n", c.IsCI)
	fmt.Printf("  Source Branch:        %s\n", c.SourceBranch)
	fmt.Printf("  Target Branch:        %s\n", c.TargetBranch)
	fmt.Println()
}
