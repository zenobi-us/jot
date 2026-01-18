package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print detailed version information for OpenNotes including build metadata",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("OpenNotes %s\n", Version)
		if BuildDate != "unknown" {
			fmt.Printf("Built: %s\n", BuildDate)
		}
		if GitCommit != "unknown" {
			fmt.Printf("Commit: %s\n", GitCommit)
		}
		if GitBranch != "unknown" {
			fmt.Printf("Branch: %s\n", GitBranch)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}