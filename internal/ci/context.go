package ci

import (
	"fmt"
	"os"
	"strings"
	"syac/pkg/gitlab"
)

// Context captures the relevant CI/CD environment state for SYAC's execution.
// It infers execution context from GitLab-provided CI variables (e.g., MR, tag, feature branch).
type Context struct {
	// GitLab CI/CD metadata
	Source                   string // CI_PIPELINE_SOURCE
	RefName                  string // CI_COMMIT_REF_NAME
	SHA                      string // CI_COMMIT_SHA
	ShortSHA                 string // CI_COMMIT_SHORT_SHA
	MRID                     string // CI_MERGE_REQUEST_IID
	Tag                      string // CI_COMMIT_TAG
	ProjectPath              string // CI_PROJECT_PATH
	ApplicationVersion       string // APP_VERSION (or fallback to CI_COMMIT_TAG)
	RegistryImage            string // CI_REGISTRY_IMAGE
	DefaultBranch            string // CI_DEFAULT_BRANCH
	Sprint                   string // SYAC_SPRINT
	ForcePush                bool   // SYAC_FORCE_PUSH
	ApplicationName          string // SYAC_APPLICATION_NAME or derived from RegistryImage
	DryRun                   bool   // CLI --dry-run flag
	MergeRequestTargetBranch string // CI_MERGE_REQUEST_TARGET_BRANCH_NAME

	// Derived booleans
	IsMergeRequest  bool // true if CI_PIPELINE_SOURCE == "merge_request_event"
	IsTag           bool // true if CI_COMMIT_TAG is non-empty
	IsFeatureBranch     bool   // true if RefName starts with 
	FeatureBranchPrefix string // SYAC_FEATURE_BRANCH_PREFIX 
	IsDefaultBranch     bool   // true if RefName equals CI_DEFAULT_BRANCH
}

// LoadContext constructs a CI Context by reading GitLab CI/CD environment variables.
// It infers flags like IsMergeRequest, IsTag, etc., and safely derives ApplicationName and Version.
func LoadContext(dryRun bool) (Context, error) {
	ref := os.Getenv("CI_COMMIT_REF_NAME")
	tag := os.Getenv("CI_COMMIT_TAG")
	defaultBranch := os.Getenv("CI_DEFAULT_BRANCH")
	featureBranchPrefix := os.Getenv("SYAC_FEATURE_BRANCH_PREFIX")
	if featureBranchPrefix == "" {
		featureBranchPrefix = "gmarm-"
	}

	// Prefer APP_VERSION, fallback to CI_COMMIT_TAG
	appVersion := os.Getenv("APP_VERSION")
	if appVersion == "" && tag != "" {
		appVersion = tag
	}

	return Context{
		Source:                   os.Getenv("CI_PIPELINE_SOURCE"),
		RefName:                  ref,
		SHA:                      os.Getenv("CI_COMMIT_SHA"),
		ShortSHA:                 os.Getenv("CI_COMMIT_SHORT_SHA"),
		MRID:                     os.Getenv("CI_MERGE_REQUEST_IID"),
		MergeRequestTargetBranch: os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME"),
		Tag:                      tag,
		ProjectPath:              os.Getenv("CI_PROJECT_PATH"),
		ApplicationVersion:       appVersion,
		RegistryImage:            os.Getenv("CI_REGISTRY_IMAGE"),
		DefaultBranch:            defaultBranch,
		Sprint:                   os.Getenv("SYAC_SPRINT"),
		IsMergeRequest:           os.Getenv("CI_PIPELINE_SOURCE") == "merge_request_event",
		IsTag:                    tag != "",
		IsFeatureBranch:          strings.HasPrefix(ref, featureBranchPrefix),
		FeatureBranchPrefix:      featureBranchPrefix,
		IsDefaultBranch:          ref == defaultBranch,
		ForcePush:                os.Getenv("SYAC_FORCE_PUSH") == "true",
		ApplicationName: func() string {
			appName := os.Getenv("SYAC_APPLICATION_NAME")
			if appName == "" {
				parts := strings.Split(os.Getenv("CI_REGISTRY_IMAGE"), "/")
				appName = parts[len(parts)-1]
			}
			return appName
		}(),
		DryRun: dryRun,
	}, nil
}

// PrintSummary emits a full CI/CD context report for debugging or transparency during the pipeline.
// It queries GitLab for the latest release metadata if available.
func (c Context) PrintSummary(client *gitlab.Client) {
	fmt.Println("CI/CD Environment Summary")
	fmt.Println("--------------------------")
	fmt.Printf("  Context               : %s\n", c.describeContext())
	fmt.Printf("  Pipeline Source       : %s\n", c.Source)
	fmt.Printf("  Branch or Tag         : %s\n", c.RefName)
	if c.IsTag {
		fmt.Printf("  Tag                   : %s\n", c.Tag)
	}
	fmt.Printf("  Commit SHA            : %s\n", c.SHA)
	fmt.Printf("  Commit Short SHA      : %s\n", c.ShortSHA)
	fmt.Printf("  Merge Request IID     : %s\n", c.MRID)
	fmt.Printf("  Merge Request Target  : %s\n", c.MergeRequestTargetBranch)
	fmt.Printf("  Project Path          : %s\n", c.ProjectPath)
	fmt.Printf("  Registry Image        : %s\n", c.RegistryImage)
	fmt.Printf("  Default Branch        : %s\n", c.DefaultBranch)
	fmt.Printf("  Is Default Branch     : %t\n", c.IsDefaultBranch)
	fmt.Printf("  Sprint                : %s\n", c.Sprint)
	fmt.Printf("  Feature Branch        : %t\n", c.IsFeatureBranch)
	fmt.Printf("  Force Push            : %t\n", c.ForcePush)
	fmt.Printf("  Application Name      : %s\n", c.ApplicationName)
	fmt.Printf("  App Version           : %s\n", c.ApplicationVersion)

	latestRelease, err := client.Releases.GetLatestRelease()
	if err != nil {
		fmt.Printf("  Latest Release        : No Release Yet (%v)\n", err)
	} else {
		fmt.Printf("  Latest Release        : %s (Tag: %s)\n", latestRelease.Name, latestRelease.TagName)
		fmt.Printf("    Description         : %s\n", latestRelease.Description)
		fmt.Printf("    Created At          : %s\n", latestRelease.CreatedAt)
	}

	fmt.Println()
}

// describeContext returns a human-readable string summarizing the current CI/CD pipeline context.
func (c Context) describeContext() string {
	switch {
	case c.IsMergeRequest:
		return "Merge Request"
	case c.IsTag:
		return fmt.Sprintf("Tag push (%s)", c.Tag)
	case c.IsFeatureBranch:
		return fmt.Sprintf("Development Branch (%s)", c.RefName)
	case c.IsDefaultBranch:
		return fmt.Sprintf("Push to default branch (%s)", c.RefName)
	default:
		return "Unknown or unsupported CI context"
	}
}
