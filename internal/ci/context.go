package ci

import (
	"fmt"
	"os"
	"strings"

	"syac/internal/config"
)

type Context struct {
	Source        string
	RefName       string
	SHA           string
	MRID          string
	ProjectPath   string
	RegistryImage string
	DefaultBranch string
	Sprint        string

	IsProtected     bool
	IsMergeRequest  bool
	IsTag           bool
	IsFeatureBranch bool

	Config *config.Config
}

func LoadContext() (Context, error) {
	cfg, err := config.LoadConfig(".syac.yml")
	if err != nil {
		// if the config file doesn't exist, we can use default values
		if os.IsNotExist(err) {
			cfg = &config.Config{
				ProtectedBranches: []string{"main", "dev", "release", "staging"},
			}
		} else {
			return Context{}, fmt.Errorf("failed to load config: %w", err)
		}
	}

	ref := os.Getenv("CI_COMMIT_REF_NAME")
	return Context{
		Source:          os.Getenv("CI_PIPELINE_SOURCE"),
		RefName:         ref,
		SHA:             os.Getenv("CI_COMMIT_SHA"),
		MRID:            os.Getenv("CI_MERGE_REQUEST_IID"),
		ProjectPath:     os.Getenv("CI_PROJECT_PATH"),
		RegistryImage:   os.Getenv("CI_REGISTRY_IMAGE"),
		DefaultBranch:   os.Getenv("CI_DEFAULT_BRANCH"),
		Sprint:          os.Getenv("SYAC_SPRINT"),
		IsProtected:     isProtectedBranch(ref, cfg.ProtectedBranches),
		IsMergeRequest:  os.Getenv("CI_PIPELINE_SOURCE") == "merge_request_event",
		IsTag:           strings.HasPrefix(ref, "refs/tags/"),
		IsFeatureBranch: !isProtectedBranch(ref, cfg.ProtectedBranches) && !strings.HasPrefix(ref, "refs/tags/"),
		Config:          cfg,
	}, nil
}

func isProtectedBranch(ref string, protectedBranches []string) bool {
	for _, b := range protectedBranches {
		if ref == b || strings.HasPrefix(ref, b+"/") {
			return true
		}
	}
	return false
}

func (c Context) PrintSummary() {
	fmt.Println("CI/CD Environment Summary")
	fmt.Println("--------------------------")
	fmt.Printf("  Context               : %s\n", c.describeContext())
	fmt.Printf("  Pipeline Source       : %s\n", c.Source)
	fmt.Printf("  Branch or Tag         : %s\n", c.RefName)
	fmt.Printf("  Commit SHA            : %s\n", c.SHA)
	fmt.Printf("  Merge Request IID     : %s\n", c.MRID)
	fmt.Printf("  Project Path          : %s\n", c.ProjectPath)
	fmt.Printf("  Registry Image        : %s\n", c.RegistryImage)
	fmt.Printf("  Default Branch        : %s\n", c.DefaultBranch)
	fmt.Printf("  Sprint                : %s\n", c.Sprint)
	fmt.Println()
}

func (c Context) describeContext() string {
	switch {
	case c.IsMergeRequest:
		return "Merge Request"
	case c.IsTag:
		return fmt.Sprintf("Tag push (%s)", c.RefName)
	case c.IsProtected:
		return fmt.Sprintf("Push to protected branch (%s)", c.RefName)
	case c.IsFeatureBranch:
		return fmt.Sprintf("Push to feature branch (%s)", c.RefName)
	default:
		return "Unknown or unsupported CI context"
	}
}
