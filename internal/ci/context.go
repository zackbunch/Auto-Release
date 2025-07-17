package ci

import (
	"fmt"
	"os"
	"strings"
	"syac/pkg/gitlab"
)

type Context struct {
	Source          string
	RefName         string
	SHA             string
	ShortSHA        string
	MRID            string
	Tag             string
	ProjectPath     string
	RegistryImage   string
	DefaultBranch   string
	Sprint          string
	ForcePush       bool
	ApplicationName string

	IsMergeRequest  bool
	IsTag           bool
	IsFeatureBranch bool
	IsDefaultBranch bool
}

func LoadContext() (Context, error) {
	ref := os.Getenv("CI_COMMIT_REF_NAME")
	tag := os.Getenv("CI_COMMIT_TAG")
	defaultBranch := os.Getenv("CI_DEFAULT_BRANCH")

	return Context{
		Source:          os.Getenv("CI_PIPELINE_SOURCE"),
		RefName:         ref,
		SHA:             os.Getenv("CI_COMMIT_SHA"),
		ShortSHA:        os.Getenv("CI_COMMIT_SHORT_SHA"),
		MRID:            os.Getenv("CI_MERGE_REQUEST_IID"),
		Tag:             tag,
		ProjectPath:     os.Getenv("CI_PROJECT_PATH"),
		RegistryImage:   os.Getenv("CI_REGISTRY_IMAGE"),
		DefaultBranch:   defaultBranch,
		Sprint:          os.Getenv("SYAC_SPRINT"),
		IsMergeRequest:  os.Getenv("CI_PIPELINE_SOURCE") == "merge_request_event",
		IsTag:           tag != "",
		IsFeatureBranch: strings.HasPrefix(ref, "gm-"),
		IsDefaultBranch: ref == defaultBranch,
		ForcePush:       os.Getenv("SYAC_FORCE_PUSH") == "true",
		ApplicationName: func() string {
			appName := os.Getenv("SYAC_APPLICATION_NAME")
			if appName == "" {
				parts := strings.Split(os.Getenv("CI_REGISTRY_IMAGE"), "/")
				appName = parts[len(parts)-1]
			}
			return appName
		}(),
	}, nil
}

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
	fmt.Printf("  Project Path          : %s\n", c.ProjectPath)
	fmt.Printf("  Registry Image        : %s\n", c.RegistryImage)
	fmt.Printf("  Default Branch        : %s\n", c.DefaultBranch)
	fmt.Printf("  Is Default Branch     : %t\n", c.IsDefaultBranch)
	fmt.Printf("  Sprint                : %s\n", c.Sprint)
	fmt.Printf("  Feature Branch        : %t\n", c.IsFeatureBranch)
	fmt.Printf("  Force Push            : %t\n", c.ForcePush)
	fmt.Printf("  Application Name      : %s\n", c.ApplicationName)
}

func (c Context) describeContext() string {
	switch {
	case c.IsMergeRequest:
		return "Merge Request"
	case c.IsTag:
		return fmt.Sprintf("Tag push (%s)", c.Tag)
	case c.IsFeatureBranch:
		return fmt.Sprintf("Developer Push to feature branch (%s)", c.RefName)
	case c.IsDefaultBranch:
		return fmt.Sprintf("Push to default branch (%s)", c.RefName)
	default:
		return "Unknown or unsupported CI context"
	}
}
