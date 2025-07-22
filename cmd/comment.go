package cmd

import (
	"fmt"
	"os"
	"strings"

	"syac/internal/ci"

	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:   "comment [mr-id]",
	Short: "Create a comment on a GitLab Merge Request. If no MR ID is provided, it attempts to get it from the CI context.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var mrID string
		if len(args) == 1 {
			mrID = args[0]
		} else {
			ctx, err := ci.LoadContext(dryRunFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading CI context: %v\n", err)
				os.Exit(1)
			}
			if ctx.MRID == "" {
				fmt.Fprintf(os.Stderr, "Error: No Merge Request ID provided and could not be determined from CI context.\n")
				os.Exit(1)
			}
			mrID = ctx.MRID
		}

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
