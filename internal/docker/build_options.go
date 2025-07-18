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
	// For RC builds, drop the env segment entirely:
	if ctx.IsMergeRequest && ctx.MergeRequestTargetBranch == "dev" {
		base := fmt.Sprintf("%s/%s", ctx.RegistryImage, appName)
		rcTag := fmt.Sprintf("rc-%s", sha)
		return []string{
			fmt.Sprintf("%s:%s", base, rcTag), // e.g. registry/.../app:rc-abc123
			fmt.Sprintf("%s:rc-latest", base), // e.g. registry/.../app:rc-latest
		}
	}

	// Otherwise, include the env folder as before:
	base := fmt.Sprintf("%s/%s/%s", ctx.RegistryImage, env, appName)
	if ctx.IsFeatureBranch {
		// feature branches get their own folder under the env
		return []string{
			fmt.Sprintf("%s/%s:%s", base, ctx.RefName, sha),
		}
	}

	// default: one immutable tag in its env
	return []string{
		fmt.Sprintf("%s:%s", base, sha),
	}
}
