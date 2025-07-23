package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// bumpCmd handles version bumping based on merge request metadata.
var bumpCmd = &cobra.Command{
	Use:   "bump [mr-id]",
	Short: "Bump version based on merge request description. If mr-id is not provided, the latest open MR will be used.",
	Long: `Analyzes the given merge request for version bump checkboxes 
(e.g., Patch, Minor, Major), computes the next version based on the 
latest tag, and prints the bump result.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var mrID string
		if len(args) > 0 {
			mrID = args[0]
		} else {
			mr, err := gitlabClient.MergeRequests.GetLatestMergeRequest()
			if err != nil {
				return fmt.Errorf("failed to get latest merge request: %w", err)
			}
			mrID = strconv.Itoa(mr.IID)
		}

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
