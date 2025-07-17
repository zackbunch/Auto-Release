package promote

import (
	"fmt"
	"syac/internal/executil"
)

type Options struct {
	PushLatest bool
	DryRun     bool
	// Future: ChangelogPath, Verbose, etc.
}

// Standard promotes one image tag to another and optionally also pushes <to>-latest
func Standard(from, to string, opts Options) error {
	if opts.DryRun {
		fmt.Printf("[DRY RUN] Promote %s -> %s\n", from, to)
		return nil
	}

	fmt.Printf("[STANDARD] Tagging %s -> %s\n", from, to)
	if err := tagAndPush(from, to); err != nil {
		return fmt.Errorf("failed to promote tag: %w", err)
	}

	if opts.PushLatest {
		latest := fmt.Sprintf("%s-latest", to)
		fmt.Printf("[STANDARD] Tagging %s -> %s\n", from, latest)
		if err := tagAndPush(from, latest); err != nil {
			return fmt.Errorf("failed to push latest tag: %w", err)
		}
	}

	return nil
}

// tagAndPush is currently stubbed — replace with docker exec logic later
func tagAndPush(from, to string) error {
	if err := executil.RunCMD("docker", "tag", from, to); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}
	if err := executil.RunCMD("docker", "push", to); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}
	return nil
}

func BlueGreen(from, to string, opts Options) error {
	fmt.Println("[STUB] BlueGreen promotion is not implemented.")
	return nil
}

func Canary(from, to string, opts Options) error {
	fmt.Println("[STUB] Canary promotion is not implemented.")
	return nil
}

func Rollback(to string, opts Options) error {
	fmt.Println("[STUB] Rollback is not implemented.")
	return nil
}
