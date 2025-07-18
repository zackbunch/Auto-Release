package promote

import (
	"fmt"
	"syac/internal/ci"
	"syac/internal/executil"
)

// Options configure behavior for promotion operations.
type Options struct {
	PushLatest bool // If true, also tag and push the ">-latest" floating tag
	DryRun     bool // If true, only simulate the operations
}

// Standard promotes one image tag to another and optionally also pushes <to>-latest.
func Standard(from, to string, opts Options) error {
	if opts.DryRun {
		fmt.Printf("[DRY RUN] Promote %s -> %s\n", from, to)
		return nil
	}

	fmt.Printf("[STANDARD] Tagging %s -> %s\n", from, to)
	if err := tagAndPush(from, to); err != nil {
		return fmt.Errorf("failed to promote tag %s->%s: %w", from, to, err)
	}

	if opts.PushLatest {
		latest := fmt.Sprintf("%s-latest", to)
		fmt.Printf("[STANDARD] Tagging %s -> %s\n", from, latest)
		if err := tagAndPush(from, latest); err != nil {
			return fmt.Errorf("failed to push latest tag %s->%s: %w", from, latest, err)
		}
	}

	return nil
}

// tagAndPush is a helper that tags a Docker image and pushes it to the registry.
func tagAndPush(from, to string) error {
	if err := executil.RunCMD("docker", "tag", from, to); err != nil {
		return fmt.Errorf("docker tag failed for %s -> %s: %w", from, to, err)
	}
	if err := executil.RunCMD("docker", "push", to); err != nil {
		return fmt.Errorf("docker push failed for %s: %w", to, err)
	}
	return nil
}

// BlueGreen is a stub for future blue/green promotion logic.
func BlueGreen(from, to string, opts Options) error {
	fmt.Println("[STUB] BlueGreen promotion is not implemented.")
	return nil
}

// Canary is a stub for future canary deployment logic.
func Canary(from, to string, opts Options) error {
	fmt.Println("[STUB] Canary promotion is not implemented.")
	return nil
}

// Rollback is a stub for future rollback logic.
func Rollback(to string, opts Options) error {
	fmt.Println("[STUB] Rollback is not implemented.")
	return nil
}

// SprintPromotion runs the end-of-sprint promotion from dev->test and test->int.
func SprintPromotion(ctx ci.Context, opts Options) error {
	reg := ctx.RegistryImage // e.g. "registry.gitlab.com/group/project"
	sha := ctx.ShortSHA      // short Git SHA

	// 1) dev -> test
	fromDev := fmt.Sprintf("%s/dev-latest", reg)
	toTest := fmt.Sprintf("%s/test-%s", reg, sha)
	if err := Standard(fromDev, toTest, opts); err != nil {
		return fmt.Errorf("dev->test promotion failed: %w", err)
	}

	// 2) test -> int
	fromTest := fmt.Sprintf("%s/test-latest", reg)
	toInt := fmt.Sprintf("%s/int-%s", reg, sha)
	if err := Standard(fromTest, toInt, opts); err != nil {
		return fmt.Errorf("test->int promotion failed: %w", err)
	}

	return nil
}
