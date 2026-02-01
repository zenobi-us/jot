package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var notesSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes with text, fuzzy matching, or boolean queries",
	Long: `Search notes using multiple methods: text search, fuzzy matching, or boolean queries.

SEARCH METHODS:

  1. Text Search (default): Exact substring matching
     opennotes notes search "meeting"

  2. Fuzzy Search: Similarity-based, typo-tolerant, ranked results
     opennotes notes search --fuzzy "mtng"

  3. Boolean Queries: Structured AND/OR/NOT filtering (see 'query' subcommand)
     opennotes notes search query --and data.tag=workflow

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

DOCUMENTATION:
  📖 Command Reference: docs/commands/notes-search.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

	// Add --fuzzy flag for fuzzy matching
	notesSearchCmd.Flags().Bool(
		"fuzzy",
		false,
		"Enable fuzzy matching for ranked results. Matches notes by similarity instead of exact text. Title matches weighted higher than body matches.",
	)
}
