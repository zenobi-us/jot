package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/core"
	"github.com/zenobi-us/opennotes/internal/services"
)

var (
	viewFormat string
	viewParams string
	viewList   bool
)

var notesViewCmd = &cobra.Command{
	Use:   "view [name]",
	Short: "Execute a named reusable query preset or list available views",
	Long: `Execute a named query preset (view) with optional parameters, or list all available views.

When called without arguments or with --list, displays all available views.

BUILT-IN VIEWS:

  today            - Notes created or updated today
  recent           - Recently modified notes (last 20)
  kanban           - Notes grouped by status column
  untagged         - Notes without any tags
  orphans          - Notes with no incoming links
  broken-links     - Notes with broken references

CUSTOM VIEWS:

  Define custom views in:
  - Global: ~/.config/opennotes/config.json
  - Notebook: <notebook>/.opennotes.json

OUTPUT FORMATS:

  list    - Simple list format (default)
  table   - ASCII table format
  json    - JSON array format

EXAMPLES:

  opennotes notes view                                    # List all views
  opennotes notes view --list                             # List all views explicitly
  opennotes notes view --list --format json               # List views as JSON
  opennotes notes view today                              # Execute a view
  opennotes notes view recent --format table              # Execute with table format
  opennotes notes view kanban --param status=todo,done    # Execute with parameters
  opennotes notes view orphans --format json              # Execute with JSON format
  opennotes notes view my-workflow --param sprint=Q1-S3   # Execute custom view`,

	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle list mode
		if viewList || len(args) == 0 {
			return handleViewList(cmd, viewFormat)
		}

		viewName := args[0]

		// Get notebook
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		// Extract notebook directory from config path
		notebookDir := filepath.Dir(nb.Config.Path)

		// Initialize ViewService
		vs := services.NewViewService(cfgService, notebookDir)

		// Set execution context for view execution
		// This requires the search index and note service from the notebook
		vs.SetExecutionContext(nb.Notes.GetIndex(), nb.Notes)

		// Get the view definition
		viewDef, err := vs.GetView(viewName)
		if err != nil {
			return fmt.Errorf("failed to get view '%s': %w", viewName, err)
		}

		// Parse user parameters
		userParams, err := vs.ParseViewParameters(viewParams)
		if err != nil {
			return fmt.Errorf("failed to parse parameters: %w", err)
		}

		// Validate parameters against view definition
		if err := vs.ValidateParameters(viewDef, userParams); err != nil {
			return fmt.Errorf("invalid parameters: %w", err)
		}

		// Apply parameter defaults
		userParams = vs.ApplyParameterDefaults(viewDef, userParams)

		// Execute the view
		ctx := context.Background()
		results, err := vs.ExecuteView(ctx, viewDef, userParams)
		if err != nil {
			return fmt.Errorf("failed to execute view '%s': %w", viewName, err)
		}

		// Render results based on whether they are grouped or flat
		if len(results.Groups) > 0 {
			return displayGroupedViewResults(viewName, results.Groups, viewFormat)
		}

		return displayViewResults(viewName, results.Notes, viewFormat)
	},
}

// displayViewResults displays flat view results (non-grouped)
func displayViewResults(viewName string, notes []services.Note, format string) error {
	if len(notes) == 0 {
		fmt.Printf("View '%s': No notes found\n", viewName)
		return nil
	}

	switch format {
	case "json":
		return displayViewResultsJSON(notes)
	case "table":
		fmt.Printf("View '%s' (%d notes):\n\n", viewName, len(notes))
		return displayNoteList(notes)
	case "list":
		fallthrough
	default:
		fmt.Printf("View '%s' (%d notes):\n\n", viewName, len(notes))
		return displayNoteList(notes)
	}
}

// displayViewResultsJSON displays view results in JSON format
func displayViewResultsJSON(notes []services.Note) error {
	type ViewResultsResponse struct {
		Notes []services.Note `json:"notes"`
		Count int             `json:"count"`
	}

	response := ViewResultsResponse{
		Notes: notes,
		Count: len(notes),
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// displayGroupedViewResults displays grouped view results (e.g., kanban)
func displayGroupedViewResults(viewName string, groups map[string][]services.Note, format string) error {
	// Count total notes
	totalNotes := 0
	for _, notes := range groups {
		totalNotes += len(notes)
	}

	if totalNotes == 0 {
		fmt.Printf("View '%s': No notes found\n", viewName)
		return nil
	}

	switch format {
	case "json":
		return displayGroupedResultsJSON(groups)
	case "table":
		fallthrough
	case "list":
		fallthrough
	default:
		return displayGroupedResultsList(viewName, groups, totalNotes)
	}
}

// displayGroupedResultsJSON displays grouped results in JSON format
func displayGroupedResultsJSON(groups map[string][]services.Note) error {
	type GroupedResultsResponse struct {
		Groups map[string][]services.Note `json:"groups"`
		Count  int                        `json:"count"`
	}

	totalNotes := 0
	for _, notes := range groups {
		totalNotes += len(notes)
	}

	response := GroupedResultsResponse{
		Groups: groups,
		Count:  totalNotes,
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// displayGroupedResultsList displays grouped results in list format
func displayGroupedResultsList(viewName string, groups map[string][]services.Note, totalNotes int) error {
	fmt.Printf("View '%s' (%d notes in %d groups):\n\n", viewName, totalNotes, len(groups))

	// Sort group keys for consistent output
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		notes := groups[key]
		fmt.Printf("## %s (%d)\n\n", key, len(notes))

		for _, note := range notes {
			// Get title from metadata or use filename
			title := note.File.Relative
			if t, ok := note.Metadata["title"]; ok {
				if str, ok := t.(string); ok && str != "" {
					title = str
				}
			}
			fmt.Printf("  - %s\n", title)
			fmt.Printf("    Path: %s\n", note.File.Relative)
		}
		fmt.Println()
	}

	return nil
}

// handleViewList lists all available views
func handleViewList(cmd *cobra.Command, format string) error {
	nb, err := requireNotebook(cmd)
	if err != nil {
		return err
	}

	// Extract notebook directory from config path
	notebookDir := filepath.Dir(nb.Config.Path)
	vs := services.NewViewService(cfgService, notebookDir)

	// Get all available views
	views, err := vs.ListAllViews()
	if err != nil {
		return fmt.Errorf("failed to list views: %w", err)
	}

	switch format {
	case "json":
		return displayViewsListJSON(views)
	case "table":
		fallthrough
	case "list":
		fallthrough
	default:
		return displayViewsList(views)
	}
}

// displayViewsList displays views in plain text format
func displayViewsList(views []core.ViewInfo) error {
	if len(views) == 0 {
		fmt.Println("No views available")
		return nil
	}

	// Group views by origin
	grouped := make(map[string][]core.ViewInfo)
	originOrder := []string{"built-in", "global", "notebook"}

	for _, v := range views {
		grouped[v.Origin] = append(grouped[v.Origin], v)
	}

	fmt.Println("\nAVAILABLE VIEWS")
	fmt.Println()

	for _, origin := range originOrder {
		if viewList, ok := grouped[origin]; ok && len(viewList) > 0 {
			// Format origin header
			switch origin {
			case "built-in":
				fmt.Println("Built-in Views:")
			case "global":
				fmt.Println("\nGlobal Views (from ~/.config/opennotes/config.json):")
			case "notebook":
				fmt.Println("\nNotebook Views (from <notebook>/.opennotes.json):")
			}

			for _, viewInfo := range viewList {
				fmt.Printf("  %-20s %s\n", viewInfo.Name, viewInfo.Description)

				// Display parameters if present
				if len(viewInfo.Parameters) > 0 {
					for _, param := range viewInfo.Parameters {
						required := "optional"
						if param.Required {
							required = "required"
						}
						defaultStr := ""
						if param.Default != "" {
							defaultStr = fmt.Sprintf(", default: %s", param.Default)
						}
						fmt.Printf("    - %-16s [%s, %s%s]\n", param.Name, param.Type, required, defaultStr)
					}
				}
			}
		}
	}

	fmt.Println()
	return nil
}

// displayViewsListJSON displays views in JSON format
func displayViewsListJSON(views []core.ViewInfo) error {
	type ViewsResponse struct {
		Views []core.ViewInfo `json:"views"`
	}

	response := ViewsResponse{Views: views}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

func init() {
	notesViewCmd.Flags().StringVar(&viewFormat, "format", "list", "Output format: list, table, or json")
	notesViewCmd.Flags().StringVar(&viewParams, "param", "", "View parameters (key=value,key2=value2)")
	notesViewCmd.Flags().BoolVar(&viewList, "list", false, "List all available views")

	notesCmd.AddCommand(notesViewCmd)
}
