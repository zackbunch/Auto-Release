package cmd

import (
	"fmt"
	"syac/internal/promote"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(promoteCmd)
}

var promoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "Promote an image between environments",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		strategy, err := cmd.Flags().GetString("strategy")
		if err != nil {
			return fmt.Errorf("failed to read 'strategy' flag: %w", err)
		}

		opts := promote.Options{
			PushLatest: pushLatest,
		}

		switch strategy {
		case "standard", "bluegreen", "canary":
			if from == "" {
				return fmt.Errorf("'from' is required for the '%s' strategy", strategy)
			}
			return promote.Standard(from, to, opts)
		case "rollback":
			return promote.Rollback(to, opts)
		default:
			return fmt.Errorf("unknown strategy: %s", strategy)
		}
	},
}

func init() {
	promoteCmd.Flags().String("from", "", "Image tag to promote from (e.g. dev-abc123)")
	promoteCmd.Flags().String("to", "", "Target environment name (e.g. test)")
	promoteCmd.Flags().Bool("push-latest", false, "Also tag and push 'latest'")
	promoteCmd.Flags().String("strategy", "standard", "Promotion strategy (standard|bluegreen|canary|rollback)")
	
	_ = promoteCmd.MarkFlagRequired("to")
}
