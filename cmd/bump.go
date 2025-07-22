package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var bumpTypeCmd = &cobra.Command{
	Use:   "bump-type [mr-id]",
	Short: "Print the version bump type selected in the MR description (Patch, Minor, Major)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mrID := args[0]

		bump, err := gitlabClient.MergeRequests.GetVersionBump(mrID)
		if err != nil {
			return fmt.Errorf("failed to get version bump from MR %s: %w", mrID, err)
		}

		fmt.Println(bump)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(bumpTypeCmd)
}
