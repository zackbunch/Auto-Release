package cmd

import (
	"fmt"
	"os"
	"strings"

	"syac/internal/assets"
	"syac/internal/ci"

	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:   "update-mr [mr-id]",
	Short: "Add the SYAC release checklist to a GitLab Merge Request description. If no MR ID is provided, it attempts to get it from the CI context.",
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

		// Read the embedded markdown checklist content
		contentBytes, err := assets.MrCommentContent.ReadFile("mr_comment.md")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading mr_comment.md: %v\n", err)
			os.Exit(1)
		}
		newBlock := string(contentBytes)

		// Fetch the existing description
		description, err := gitlabClient.MergeRequests.GetMergeRequestDescription(mrID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching MR description: %v\n", err)
			os.Exit(1)
		}

		if strings.Contains(description, "[SYAC]") || strings.Contains(description, "<!-- syac:release-type -->") {
			fmt.Printf("MR %s already contains the SYAC checklist. Skipping update.\n", mrID)
			return
		}

		// Append the checklist to the existing description
		updated := strings.TrimSpace(description) + "\n\n" + newBlock

		if err := gitlabClient.MergeRequests.UpdateMergeRequestDescription(mrID, updated); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating MR description: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("SYAC checklist successfully injected into MR %s description.\n", mrID)
	},
}

func init() {
	rootCmd.AddCommand(commentCmd)
}
