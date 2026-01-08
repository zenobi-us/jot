package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var notebookCmd = &cobra.Command{
	Use:   "notebook",
	Short: "Manage notebooks",
	Long:  `Commands for managing notebooks - create, list, register, and configure notebooks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default: show current notebook info
		nb, err := notebookService.Infer("")
		if err != nil {
			return err
		}

		if nb == nil {
			fmt.Println("No notebook found.")
			fmt.Println("")
			fmt.Println("Create one with:")
			fmt.Println("  opennotes notebook create --name \"My Notebook\"")
			return nil
		}

		return displayNotebookInfo(nb)
	},
}

func init() {
	rootCmd.AddCommand(notebookCmd)
}
