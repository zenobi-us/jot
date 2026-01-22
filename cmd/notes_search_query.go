package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/services"
)

var notesSearchQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Search notes with boolean operators",
	Long: `Search notes using AND/OR/NOT boolean operators for complex filtering.

Conditions are specified as field=value pairs. Supported fields:
  - data.tag, data.tags   - Note tags
  - data.status           - Note status (e.g., active, archived)
  - data.priority         - Priority level (e.g., high, low, critical)
  - data.assignee         - Assigned person
  - data.author           - Note author
  - data.type             - Note type
  - data.category         - Category classification
  - data.project          - Project name
  - data.sprint           - Sprint identifier
  - path                  - File path (supports globs)
  - title                 - Note title
  - links-to              - Documents this note links to
  - linked-by             - Documents that link to this note

Boolean Operators:
  --and     All AND conditions must match (intersection)
  --or      Any OR condition must match (union)
  --not     Excludes notes matching this condition

Operator precedence: AND conditions are evaluated first, then combined
with OR conditions, then NOT conditions are applied as exclusions.

Examples:
  # Single condition - find notes tagged with "workflow"
  opennotes notes search query --and data.tag=workflow

  # Multiple AND conditions - find active workflow notes
  opennotes notes search query --and data.tag=workflow --and data.status=active

  # OR conditions - find high or critical priority notes
  opennotes notes search query --or data.priority=high --or data.priority=critical

  # Combined - find epic notes that are not archived
  opennotes notes search query --and data.tag=epic --not data.status=archived

  # Path filtering with glob patterns
  opennotes notes search query --and path=projects/*

Security:
  All queries use parameterized SQL to prevent injection attacks.
  Only whitelisted fields can be queried. Values are length-limited.`,
	RunE: notesSearchQueryRunE,
}

func init() {
	notesSearchCmd.AddCommand(notesSearchQueryCmd)

	notesSearchQueryCmd.Flags().StringArray("and", []string{}, "AND condition (field=value) - all must match")
	notesSearchQueryCmd.Flags().StringArray("or", []string{}, "OR condition (field=value) - any must match")
	notesSearchQueryCmd.Flags().StringArray("not", []string{}, "NOT condition (field=value) - excludes matches")
}

func notesSearchQueryRunE(cmd *cobra.Command, args []string) error {
	// Parse flags
	andFlags, err := cmd.Flags().GetStringArray("and")
	if err != nil {
		return fmt.Errorf("failed to parse --and flags: %w", err)
	}

	orFlags, err := cmd.Flags().GetStringArray("or")
	if err != nil {
		return fmt.Errorf("failed to parse --or flags: %w", err)
	}

	notFlags, err := cmd.Flags().GetStringArray("not")
	if err != nil {
		return fmt.Errorf("failed to parse --not flags: %w", err)
	}

	// Check that at least one condition is provided
	if len(andFlags) == 0 && len(orFlags) == 0 && len(notFlags) == 0 {
		return fmt.Errorf("at least one condition is required (use --and, --or, or --not)")
	}

	// Get search service to parse conditions
	searchService := services.NewSearchService()

	// Parse and validate conditions
	conditions, err := searchService.ParseConditions(andFlags, orFlags, notFlags)
	if err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	// Get notebook
	nb, err := requireNotebook(cmd)
	if err != nil {
		return err
	}

	// Execute query
	notes, err := nb.Notes.SearchWithConditions(context.Background(), conditions)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Display results
	if len(notes) == 0 {
		fmt.Println("No notes found matching the specified conditions.")
		return nil
	}

	fmt.Printf("Found %d note(s) matching conditions:\n\n", len(notes))
	return displayNoteList(notes)
}
