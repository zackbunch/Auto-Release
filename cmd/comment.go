package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:   "comment <mr-id>",
	Short: "Create a comment on a GitLab Merge Request",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mrID := args[0]
		if err := gitlabClient.MergeRequests.CreateMergeRequestComment(mrID); err != nil {
			if strings.Contains(err.Error(), "comment already exists") {
				fmt.Printf("Comment already exists on MR %s. Skipping.\n", mrID)
			} else {
				fmt.Fprintf(os.Stderr, "Error creating comment: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Comment created successfully on MR %s\n", mrID)
		}
	},
}

func init() {
	rootCmd.AddCommand(commentCmd)
}
