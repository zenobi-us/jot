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
  🔍 Advanced SQL Queries: opennotes notes search --sql "SELECT ... FROM read_markdown('**/*.md')"
  🤖 JSON Output for Automation: Results automatically JSON-formatted for jq and tool integration
  📊 Markdown Analysis: Extract word counts, statistics, and structure using SQL functions
  💾 Large Notebook Support: Efficiently query thousands of notes in seconds

DOCUMENTATION:
  📚 SQL Query Guide: https://github.com/zenobi-us/opennotes/blob/main/docs/sql-guide.md
  🚀 Automation & JSON: https://github.com/zenobi-us/opennotes/blob/main/docs/json-sql-guide.md

Examples:
  # List all notes
  opennotes notes list

  # Add a new note with title
  opennotes notes add --title "Meeting Notes"

  # Search notes by content
  opennotes notes search "project deadline"

  # Query with SQL (see sql-guide.md for more examples)
  opennotes notes search --sql "SELECT file_path FROM read_markdown('**/*.md') WHERE content LIKE '%TODO%' ORDER BY file_path"

  # Remove a note
  opennotes notes remove my-note.md`,
}

func init() {
	rootCmd.AddCommand(notesCmd)
}
