package cmd

import (
	"github.com/spf13/cobra"
)

var notesCmd = &cobra.Command{
	Use:   "notes",
	Short: "Manage notes",
	Long:  `Commands for managing notes - list, search, add, and remove notes.`,
}

func init() {
	rootCmd.AddCommand(notesCmd)
}
