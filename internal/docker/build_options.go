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
	FullImage      string
	Push           bool
	DryRun         bool
}

func BuildOptionsFromContext(ctx ci.Context) (*BuildOptions, error) {
	dockerfile := os.Getenv("SYAC_DOCKERFILE")
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	buildContext := os.Getenv("SYAC_BUILD_CONTEXT")
	if buildContext == "" {
		buildContext = "."
	}

	extraArgs := strings.Fields(os.Getenv("SYAC_BUILD_EXTRA_ARGS"))
	appName := os.Getenv("SYAC_APPLICATION_NAME")
	if appName == "" {
		parts := strings.Split(ctx.RegistryImage, "/")
		appName = parts[len(parts)-1]
	}

	env := deriveOpenShiftEnv(ctx.RefName)
	tag := generateTag(ctx)
	fullImage := fmt.Sprintf("%s/%s/%s:%s", ctx.RegistryImage, env, appName, tag)
	push := shouldPush(env, os.Getenv("SYAC_FORCE_PUSH") == "true")

	return &BuildOptions{
		Dockerfile:     dockerfile,
		ContextPath:    buildContext,
		ExtraBuildArgs: extraArgs,
		ImageName:      appName,
		TargetTag:      tag,
		FullImage:      fullImage,
		Push:           push,
		DryRun:         os.Getenv("SYAC_DRY_RUN") == "true",
	}, nil
}

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

func generateTag(ctx ci.Context) string {
	// If it's a tag push (a release), use the tag name itself.
	if ctx.IsTag {
		return ctx.RefName
	}

	// Check for merge into dev for RC tag
	if ctx.RefName == "dev" && (ctx.Source == "merge_request_event" || ctx.Source == "push") {
		rcNumber := os.Getenv("SYAC_RC_NUMBER")
		if rcNumber != "" {
			return fmt.Sprintf("rc.%s", rcNumber)
		}
	}
	// Default to short SHA
	return os.Getenv("CI_COMMIT_SHORT_SHA")
}

func shouldPush(env string, force bool) bool {
	return env != "dev" || force
}
