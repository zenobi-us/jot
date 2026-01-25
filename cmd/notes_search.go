package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/services"
)

var notesSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes with text, fuzzy matching, boolean queries, or SQL",
	Long: `Search notes using multiple methods: text search, fuzzy matching, boolean queries, or SQL.

SEARCH METHODS:

  1. Text Search (default): Exact substring matching
     opennotes notes search "meeting"

  2. Fuzzy Search: Similarity-based, typo-tolerant, ranked results
     opennotes notes search --fuzzy "mtng"

  3. Boolean Queries: Structured AND/OR/NOT filtering (see 'query' subcommand)
     opennotes notes search query --and data.tag=workflow

  4. SQL Queries: Full DuckDB SQL power
     opennotes notes search --sql "SELECT * FROM read_markdown('**/*.md') LIMIT 10"

TEXT SEARCH EXAMPLES:
  opennotes notes search "meeting"              # Search for "meeting"
  opennotes notes search "todo" --notebook ~/n  # Search in specific notebook
  opennotes notes search                        # List all notes

FUZZY SEARCH EXAMPLES:
  opennotes notes search --fuzzy "mtng"         # Matches "meeting", "meetings"
  opennotes notes search "project" --fuzzy      # Ranked by similarity
  opennotes notes search --fuzzy                # All notes, ranked

  Fuzzy matching:
  - Uses character sequence matching (like VS Code's Ctrl+P)
  - Title matches weighted 2x higher than body matches
  - Results sorted by match score (best first)
  - Searches first 500 chars of body for performance

BOOLEAN QUERY SUBCOMMAND:
  Use 'opennotes notes search query' for structured filtering:
  
  opennotes notes search query --and data.tag=workflow
  opennotes notes search query --and data.tag=epic --not data.status=archived
  opennotes notes search query --or data.priority=high --or data.priority=critical
  opennotes notes search query --and links-to=tasks/**/*.md

  Supported fields:
  - data.tag, data.status, data.priority, data.assignee, data.author
  - data.type, data.category, data.project, data.sprint
  - path, title
  - links-to (find notes linking TO target)
  - linked-by (find notes linked FROM source)

SQL QUERY EXAMPLES:
  opennotes notes search --sql "SELECT * FROM read_markdown('**/*.md') LIMIT 10"
  opennotes notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%todo%'"

SQL SECURITY:
  - Only SELECT and WITH queries allowed (read-only)
  - 30-second timeout per query
  - Path traversal (../) blocked
  - File access restricted to notebook directory

DOCUMENTATION:
  📖 Command Reference: docs/commands/notes-search.md
  🔍 SQL Guide: docs/sql-guide.md
  📚 Functions: docs/sql-functions-reference.md`,
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

		// Get --fuzzy flag
		fuzzyFlag, _ := cmd.Flags().GetBool("fuzzy")

		// Get search term (optional for fuzzy mode)
		var searchTerm string
		if len(args) > 0 {
			searchTerm = args[0]
		}

		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		notes, err := nb.Notes.SearchNotes(context.Background(), searchTerm, fuzzyFlag)
		if err != nil {
			return fmt.Errorf("failed to search notes: %w", err)
		}

		if len(notes) == 0 {
			if searchTerm != "" {
				fmt.Printf("No notes found matching '%s'\n", searchTerm)
			} else {
				fmt.Println("No notes found")
			}
			return nil
		}

		if searchTerm != "" {
			searchMode := "matching"
			if fuzzyFlag {
				searchMode = "fuzzy matching"
			}
			fmt.Printf("Found %d note(s) %s '%s':\n\n", len(notes), searchMode, searchTerm)
		} else {
			fmt.Printf("Found %d note(s):\n\n", len(notes))
		}

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

	// Add --fuzzy flag for fuzzy matching
	notesSearchCmd.Flags().Bool(
		"fuzzy",
		false,
		"Enable fuzzy matching for ranked results. Matches notes by similarity instead of exact text. Title matches weighted higher than body matches.",
	)
}
