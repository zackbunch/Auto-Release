package docker

import (
	"fmt"
	"os"
	"strings"

	"syac/internal/ci"
)

type BuildOptions struct {
	Dockerfile     string
	ContextPath    string
	ExtraBuildArgs []string
	ImageName      string
	TargetTag      string
	FullImages     []string
	Push           bool
	DryRun         bool
}

// BuildOptionsFromContext constructs BuildOptions based on the CI context.
func BuildOptionsFromContext(ctx ci.Context) (*BuildOptions, error) {
	// Determine Dockerfile
	dockerfile := os.Getenv("SYAC_DOCKERFILE")
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	// Determine build context
	buildContext := os.Getenv("SYAC_BUILD_CONTEXT")
	if buildContext == "" {
		buildContext = "."
	}

	// Parse extra build args
	extraArgs := strings.Fields(os.Getenv("SYAC_BUILD_EXTRA_ARGS"))

	// Application name and SHA tag
	appName := ctx.ApplicationName
	sha := ctx.ShortSHA

	// Determine OpenShift environment
	env := deriveOpenShiftEnv(ctx)

	// Generate all image tags (immutable + floating)
	fullImages := generateBuildTags(ctx, env, appName, sha)

	return &BuildOptions{
		Dockerfile:     dockerfile,
		ContextPath:    buildContext,
		ExtraBuildArgs: extraArgs,
		ImageName:      appName,
		TargetTag:      sha,
		FullImages:     fullImages,
		Push:           os.Getenv("SYAC_PUSH") == "true",
		DryRun:         ctx.DryRun,
	}, nil
}

// deriveOpenShiftEnv maps branch context to OpenShift environment
func deriveOpenShiftEnv(ctx ci.Context) string {
	if ctx.IsTag {
		return "prod"
	}
	if ctx.IsFeatureBranch {
		return "development"
	}
	switch ctx.RefName {
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

// generateBuildTags returns both immutable and floating tags
// for the given context, environment, application name, and SHA.
func generateBuildTags(ctx ci.Context, env, appName, sha string) []string {
	base := fmt.Sprintf("%s/%s/%s", ctx.RegistryImage, env, appName)

	switch {
	// Release Candidate build on dev MRs: rc-<sha> + rc-latest
	case ctx.IsMergeRequest && ctx.MergeRequestTargetBranch == "dev":
		rcTag := fmt.Sprintf("rc-%s", sha)
		return []string{
			fmt.Sprintf("%s:%s", base, rcTag), // e.g. myreg/dev/myapp:rc-a1b2c3d
			fmt.Sprintf("%s:rc-latest", base), // e.g. myreg/dev/myapp:rc-latest
		}

	// Feature branch build: feature-specific namespace + SHA
	case ctx.IsFeatureBranch:
		return []string{
			fmt.Sprintf("%s/%s:%s", base, ctx.RefName, sha),
		}

	// Default build: appName + SHA
	default:
		return []string{
			fmt.Sprintf("%s:%s", base, sha),
		}
	}
}
