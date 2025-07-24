package docker

import (
	"fmt"
	"os"
	"strings"

	"syac/internal/ci"
)

// BuildOptions defines all configuration required to build a Docker image.
// This struct is typically constructed from CI/CD environment variables and pipeline context.
type BuildOptions struct {
	Dockerfile     string   // Path to the Dockerfile to use (default: "Dockerfile")
	ContextPath    string   // Build context directory (default: ".")
	ExtraBuildArgs []string // Additional --build-arg entries passed to docker build
	ImageName      string   // Application name, used for tagging
	TargetTag      string   // Git SHA or release tag used as the version tag
	FullImages     []string // Full docker image names to tag the built image with
	Push           bool     // Whether to push the image after build (controlled via SYAC_PUSH)
	DryRun         bool     // If true, log commands instead of executing (useful for debugging)
}

// BuildOptionsFromContext extracts and constructs a BuildOptions instance from CI context and env vars.
// This enforces consistent tagging and context conventions across all repositories.
func BuildOptionsFromContext(ctx ci.Context) (*BuildOptions, error) {
	// Determine Dockerfile path (default: "Dockerfile")
	dockerfile := os.Getenv("SYAC_DOCKERFILE")
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	// Determine Docker build context (default: ".")
	buildContext := os.Getenv("SYAC_BUILD_CONTEXT")
	if buildContext == "" {
		buildContext = "."
	}

	// Optional: Extra build arguments passed via --build-arg
	extraArgs := strings.Fields(os.Getenv("SYAC_BUILD_EXTRA_ARGS"))

	// Determine image name and SHA tag
	appName := ctx.ApplicationName
	sha := ctx.ShortSHA

	// Resolve OpenShift target environment
	env := deriveOpenShiftEnv(ctx)

	// Generate image tags based on context
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

// deriveOpenShiftEnv returns the appropriate OpenShift namespace/environment based on branch context.
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

func generateBuildTags(ctx ci.Context, env, appName, sha string) []string {
	base := fmt.Sprintf("%s/%s/%s", ctx.RegistryImage, env, appName)

	// Feature branches
	if ctx.IsFeatureBranch {
		return []string{
			fmt.Sprintf("%s/%s:%s", base, ctx.RefName, sha),
		}
	}

	tags := []string{
		fmt.Sprintf("%s:%s", base, sha),
	}

	if ctx.ApplicationVersion != "" {
		tags = append(tags, fmt.Sprintf("%s:%s", base, ctx.ApplicationVersion))
	}

	if env == "prod" {
		tags = append(tags, fmt.Sprintf("%s:latest", base))
	}

	return tags
}
