package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// bumpCmd handles version bumping based on merge request metadata.
var bumpCmd = &cobra.Command{
	Use:   "bump [mr-id]",
	Short: "Bump version based on merge request description.",
	Long: `Analyzes the given merge request for version bump checkboxes 
(e.g., Patch, Minor, Major), computes the next version based on the 
latest tag, and prints the bump result.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mrID := args[0]

		bumpType, err := gitlabClient.MergeRequests.GetVersionBump(mrID)
		if err != nil {
			return fmt.Errorf("unable to determine bump type from MR %s: %w", mrID, err)
		}

		fmt.Printf("Version bump: %s\n", bumpType)
		return nil
	},
}

func init() {
	bumpCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run without applying changes.")
	rootCmd.AddCommand(bumpCmd)
}
