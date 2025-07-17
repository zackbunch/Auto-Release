package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var protectedBranchesCmd = &cobra.Command{
	Use:   "protectedbranches",
	Short: "List protected branches for the project",
	Long:  `This command fetches and displays a list of protected branches for the configured GitLab project.`, 
	Run: func(cmd *cobra.Command, args []string) {
		branches, err := gitlabClient.Branches.ListProtectedBranches()
		if err != nil {
			fmt.Printf("Error listing protected branches: %v\n", err)
			return
		}

		if len(branches) == 0 {
			fmt.Println("No protected branches found.")
			return
		}

		fmt.Println("Protected Branches:")
		for _, branch := range branches {
			fmt.Printf("- %s\n", branch.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(protectedBranchesCmd)
}
