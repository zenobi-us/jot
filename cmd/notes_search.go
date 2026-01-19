package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/services"
)

var notesSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes (text or SQL)",
	Long: `Searches notes by content or filename using DuckDB SQL.

The query searches both file names and content of markdown files.

DOCUMENTATION:
  🔍 Complete SQL Guide: https://github.com/zenobi-us/opennotes/blob/main/docs/sql-guide.md
  📚 Function Reference: https://github.com/zenobi-us/opennotes/blob/main/docs/sql-functions-reference.md
  🤖 Automation & JSON: https://github.com/zenobi-us/opennotes/blob/main/docs/json-sql-guide.md

Examples:
  # Search for notes containing "meeting"
  opennotes notes search "meeting"

  # Search with specific notebook
  opennotes notes search "todo" --notebook ~/notes

  # Execute custom SQL query to find all notes
  opennotes notes search --sql "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 10"

  # Find notes with Python code blocks
  opennotes notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%python%'"

SQL Query Examples:
  Basic pattern search:
    opennotes notes search --sql "SELECT * FROM read_markdown('*.md') LIMIT 5"
  
  Content search across all notes:
    opennotes notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%todo%'"
  
  Specific folder query:
    opennotes notes search --sql "SELECT title FROM read_markdown('projects/*.md') ORDER BY title"
  
  Complex filtering with statistics:
    opennotes notes search --sql "SELECT file_path, (md_stats(content)).word_count as words FROM read_markdown('**/*.md', include_filepath:=true) WHERE (md_stats(content)).word_count > 1000"

File Pattern Behavior:
  - All file patterns resolve from notebook root directory
  - Queries work consistently regardless of current directory
  - Security restrictions prevent access to files outside notebook
  - Use forward slashes in patterns (cross-platform compatibility)
  - Supported patterns: *.md (root files), **/*.md (all files), subfolder/*.md (specific folder)

SQL Security:
  Only SELECT and WITH queries allowed. Read-only access enforced.
  30-second timeout per query. No data modification possible.
  Path traversal protection: attempts to access files outside notebook (../) are blocked.
  All file access restricted to notebook directory tree for security.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get --sql flag if provided
		sqlQuery, _ := cmd.Flags().GetString("sql")

		// If --sql flag is provided, run SQL mode
		if sqlQuery != "" {
			nb, err := requireNotebook(cmd)
			if err != nil {
				return err
			}

			// Execute the SQL query using NoteService
			results, err := nb.Notes.ExecuteSQLSafe(context.Background(), sqlQuery)
			if err != nil {
				return fmt.Errorf("SQL query failed: %w", err)
			}

			// Create display service and render results
			display, err := services.NewDisplay()
			if err != nil {
				return fmt.Errorf("failed to create display: %w", err)
			}

			return display.RenderSQLResults(results)
		}

		// Normal search mode - require a query argument
		if len(args) == 0 {
			return fmt.Errorf("query argument required (or use --sql flag)")
		}

		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		notes, err := nb.Notes.SearchNotes(context.Background(), args[0])
		if err != nil {
			return fmt.Errorf("failed to search notes: %w", err)
		}

		if len(notes) == 0 {
			fmt.Printf("No notes found matching '%s'\n", args[0])
			return nil
		}

		fmt.Printf("Found %d note(s) matching '%s':\n\n", len(notes), args[0])
		return displayNoteList(notes)
	},
}

func init() {
	notesCmd.AddCommand(notesSearchCmd)

	// Add --sql flag for custom SQL queries
	notesSearchCmd.Flags().String(
		"sql",
		"",
		"Execute custom SQL query against notes (read-only, 30s timeout, SELECT/WITH only). File patterns (*.md, **/*.md) are resolved relative to notebook root directory for consistent behavior. Path traversal (../) is blocked for security. Examples: --sql \"SELECT * FROM read_markdown('**/*.md') LIMIT 5\"",
	)
}
