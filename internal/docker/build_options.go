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

	appName := ctx.ApplicationName

	tag := ctx.ShortSHA
	env := deriveOpenShiftEnv(ctx)

	var fullImages []string
	if ctx.RefName == "dev" {
		baseImage := fmt.Sprintf("%s/%s/%s", ctx.RegistryImage, env, appName)
		fullImages = []string{baseImage + ":dev-" + tag, baseImage + ":latest"}
	} else if ctx.IsFeatureBranch {
		fullImages = []string{fmt.Sprintf("%s/%s/%s:%s", ctx.RegistryImage, env, ctx.RefName, tag)}
	} else {
		fullImages = []string{fmt.Sprintf("%s/%s/%s:%s", ctx.RegistryImage, env, appName, tag)}
	}

	return &BuildOptions{
		Dockerfile:     dockerfile,
		ContextPath:    buildContext,
		ExtraBuildArgs: extraArgs,
		ImageName:      appName,
		TargetTag:      tag,
		FullImages:     fullImages,
		Push:           os.Getenv("SYAC_PUSH") == "true",
		DryRun:         ctx.DryRun,
	}, nil
}

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
