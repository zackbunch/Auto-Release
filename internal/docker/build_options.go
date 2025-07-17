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

	tag := ctx.ShortSHA
	env := deriveOpenShiftEnv(ctx)

	fullImage := fmt.Sprintf("%s/%s/%s:%s", ctx.RegistryImage, env, appName, tag)

	return &BuildOptions{
		Dockerfile:     dockerfile,
		ContextPath:    buildContext,
		ExtraBuildArgs: extraArgs,
		ImageName:      appName,
		TargetTag:      tag,
		FullImage:      fullImage,
		Push:           os.Getenv("SYAC_PUSH") == "true",
		DryRun:         os.Getenv("SYAC_DRY_RUN") == "true",
	}, nil
}

func deriveOpenShiftEnv(ctx ci.Context) string {
	if ctx.IsTag {
		return "prod"
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
