package cmd

import (
	"fmt"
	"strings"
	"syac/internal/promote"

	"github.com/spf13/cobra"
)

// promoteCmd represents the promote command in the syac CLI.
var promoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "Promote an image between environments by retagging and pushing",
	Long:  `Retags an existing image (e.g., dev-latest) to a target environment tag (e.g., test-<sha>) and optionally updates the floating latest pointer.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read flags
		from, err := cmd.Flags().GetString("from")
		if err != nil {
			return fmt.Errorf("failed to read 'from' flag: %w", err)
		}

		to, err := cmd.Flags().GetString("to")
		if err != nil {
			return fmt.Errorf("failed to read 'to' flag: %w", err)
		}

		pushLatest, err := cmd.Flags().GetBool("push-latest")
		if err != nil {
			return fmt.Errorf("failed to read 'push-latest' flag: %w", err)
		}

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return fmt.Errorf("failed to read 'dry-run' flag: %w", err)
		}

		strategy, err := cmd.Flags().GetString("strategy")
		if err != nil {
			return fmt.Errorf("failed to read 'strategy' flag: %w", err)
		}

		opts := promote.Options{
			PushLatest: pushLatest,
			DryRun:     dryRun,
		}

		// Normalize strategy
		strategy = strings.ToLower(strategy)

		switch strategy {
		case "standard":
			if from == "" {
				return fmt.Errorf("--from is required for standard strategy")
			}
			return promote.Standard(from, to, opts)

		case "bluegreen", "canary", "rollback":
			return fmt.Errorf("strategy '%s' is not yet implemented", strategy)

		default:
			return fmt.Errorf("unknown strategy: '%s' (must be: standard)", strategy)
		}
	},
}

func init() {
	rootCmd.AddCommand(promoteCmd)

	// Define command-line flags
	promoteCmd.Flags().String("from", "", "Image tag to promote from (e.g. dev-abc123)")
	promoteCmd.Flags().String("to", "", "Target image tag (e.g. test-abc123)")
	promoteCmd.Flags().Bool("push-latest", false, "Also tag and push 'latest' for the target environment")
	promoteCmd.Flags().String("strategy", "standard", "Promotion strategy (standard only supported)")
	promoteCmd.Flags().Bool("dry-run", false, "Simulate promotion without executing Docker commands")

	// 'to' flag is mandatory
	_ = promoteCmd.MarkFlagRequired("to")
}
