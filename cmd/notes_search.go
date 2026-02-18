package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/search"
	"github.com/zenobi-us/opennotes/internal/search/parser"
	"github.com/zenobi-us/opennotes/internal/services"
)

var notesSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes with text, fuzzy matching, boolean queries, or DSL pipe syntax",
	Long: `Search notes using multiple methods: text search, fuzzy matching, boolean queries, or DSL with pipe syntax.

SEARCH METHODS:

  1. Text Search (default): Exact substring matching
     opennotes notes search "meeting"

  2. Fuzzy Search: Similarity-based, typo-tolerant, ranked results
     opennotes notes search --fuzzy "mtng"

  3. Boolean Queries: Structured AND/OR/NOT filtering (see 'query' subcommand)
     opennotes notes search query --and data.tag=workflow

  4. DSL Pipe Syntax: Filter with directives for sorting and limits
     opennotes notes search "tag:work | sort:modified:desc limit:10"

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

DSL PIPE SYNTAX EXAMPLES:
  opennotes notes search "tag:work | sort:modified:desc"
  opennotes notes search "status:todo | sort:created:asc limit:20"
  opennotes notes search "| sort:title:asc"     # All notes, sorted by title

  DSL Filter:
  - tag:<value>      Notes with specific tag
  - status:<value>   Notes with status field
  - title:<text>     Search in title
  - path:<prefix>    Notes in path prefix
  - created:>date    Created after date
  - modified:<date   Modified before date

  Directives (after |):
  - sort:<field>:<dir>  Sort by field (modified, created, title, path)
                        Direction: asc or desc (default: asc)
  - limit:<n>           Return at most n results
  - offset:<n>          Skip first n results (for pagination)

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

		// Check if query contains pipe syntax (and not fuzzy mode)
		if !fuzzyFlag && strings.Contains(searchTerm, "|") {
			return runSearchWithPipeSyntax(cmd.Context(), nb, searchTerm)
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

// runSearchWithPipeSyntax executes a search using pipe syntax (filter | directives).
// This allows DSL-based search with sort, limit, and other options.
// Example: "tag:work | sort:modified:desc limit:10"
func runSearchWithPipeSyntax(ctx context.Context, nb *services.Notebook, query string) error {
	// Split query into filter and directives
	filterPart, directivesPart := services.SplitViewQuery(query)

	// Parse directives
	directives, err := services.ParseDirectives(directivesPart)
	if err != nil {
		return fmt.Errorf("failed to parse directives: %w", err)
	}

	// Build FindOpts from directives
	opts := search.FindOpts{
		Limit:  directives.Limit,
		Offset: directives.Offset,
	}

	// Parse filter DSL if present
	if filterPart != "" {
		p := parser.New()
		parsedQuery, err := p.Parse(filterPart)
		if err != nil {
			return fmt.Errorf("failed to parse filter: %w", err)
		}
		opts.Query = parsedQuery
		opts.RawQuery = filterPart
	}

	// Set sort from directives
	if directives.SortField != "" {
		opts.Sort = directiveToSortSpec(directives.SortField, directives.SortDirection)
	}

	// Execute search using the new method
	notes, err := nb.Notes.SearchWithFindOpts(ctx, opts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(notes) == 0 {
		fmt.Printf("No notes found matching query\n")
		return nil
	}

	fmt.Printf("Found %d note(s):\n\n", len(notes))
	return displayNoteList(notes)
}

// directiveToSortSpec converts directive sort parameters to search.SortSpec
func directiveToSortSpec(field, direction string) search.SortSpec {
	var sortDirection search.SortDirection
	if direction == "desc" {
		sortDirection = search.SortDesc
	} else {
		sortDirection = search.SortAsc
	}

	var sortField search.SortField
	switch field {
	case "modified":
		sortField = search.SortByModified
	case "created":
		sortField = search.SortByCreated
	case "title":
		sortField = search.SortByTitle
	case "path":
		sortField = search.SortByPath
	default:
		sortField = search.SortByRelevance
	}

	return search.SortSpec{Field: sortField, Direction: sortDirection}
}
