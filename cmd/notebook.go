package cmd

import (
	"github.com/spf13/cobra"
)

var notebookCmd = &cobra.Command{
	Use:     "notebook",
	Aliases: []string{"nb"},
	Short:   "Manage notebooks - create, list, auto-discovery",
	Long: `Commands for managing notebooks - create, list, register, and configure notebooks.

A notebook is a directory containing markdown notes with a .opennotes.json config file.
When run without a subcommand, displays info about the current notebook.

QUICK START WITH EXISTING MARKDOWN:
  1. Import: opennotes notebook create "My Notes" --path ~/my-notes
  2. Verify: opennotes notes list
  3. Search: opennotes notes search "meeting" --fuzzy

AUTO-DISCOVERY:
  - Notebooks are discovered by looking for .opennotes.json in current directory or ancestors
  - This allows context-aware workflows: cd ~/work/notes && opennotes notes list (uses work notebook)
  - Perfect for multi-project setups where each project has its own notes

DOCUMENTATION:
  📋 Notebook Discovery & Management: https://github.com/zenobi-us/opennotes/blob/main/docs/notebook-discovery.md

Examples:
  # Show current notebook info
  opennotes notebook

  # List all notebooks
  opennotes notebook list

  # Create a notebook with existing markdown files
  opennotes notebook create "My Notes" --path ~/my-notes

  # Create new empty notebook
  opennotes notebook create --name "Work Notes"

  # Register existing notebook globally
  opennotes notebook register /path/to/notebook`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default: show current notebook info
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		return displayNotebookInfo(nb)
	},
}

func init() {
	rootCmd.AddCommand(notebookCmd)
}
