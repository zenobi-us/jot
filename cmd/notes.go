package cmd

import (
	"github.com/spf13/cobra"
)

var notesCmd = &cobra.Command{
	Use:   "notes",
	Short: "Manage notes - list, search, add, remove",
	Long: `Commands for managing notes - list, search, add, and remove notes.

Notes are markdown files stored in the notebook's notes directory.
The notebook is automatically discovered from the current directory,
or can be specified with the --notebook flag.

POWER USER FEATURES:
  🔍 Advanced Query Filters: opennotes notes search query --and path=projects/*.md
  🤖 JSON Output for Automation: Results automatically JSON-formatted for jq and tool integration
  ✨ Fuzzy Search: opennotes notes search --fuzzy "mtng"
  💾 Large Notebook Support: Efficiently search thousands of notes in seconds

DOCUMENTATION:
  📚 Search Guide: https://github.com/zenobi-us/opennotes/blob/main/docs/commands/notes-search.md

Examples:
  # List all notes
  opennotes notes list

  # Add a new note with title
  opennotes notes add --title "Meeting Notes"

  # Search notes by content
  opennotes notes search "project deadline"

  # Query with boolean filters
  opennotes notes search query --and path=**/*.md --not path=archive/*

  # Remove a note
  opennotes notes remove my-note.md`,
}

func init() {
	rootCmd.AddCommand(notesCmd)
}
