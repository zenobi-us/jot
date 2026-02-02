package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

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

		// Parse user parameters
		userParams, err := vs.ParseViewParameters(viewParams)
		if err != nil {
			return fmt.Errorf("failed to parse parameters: %w", err)
		}

		// SQL views are no longer supported after DuckDB removal
		return fmt.Errorf(`SQL views are no longer supported

OpenNotes has removed DuckDB and SQL query support in favor of a pure Go
full-text search implementation using Bleve.

MIGRATION OPTIONS:

1. Use query DSL for searches:
   opennotes notes search "tag:work status:todo"
   opennotes notes search "project AND (urgent OR high-priority)"

2. Use 'notes list' with filtering:
   opennotes notes list

3. For advanced filtering, use jq with JSON output:
   opennotes notes list --format json | jq '.notes[] | select(.metadata.status == "todo")'

BREAKING CHANGE: This is part of the DuckDB removal in version 0.1.0.
Custom views using SQL queries must be migrated to use the query DSL or
external tooling (jq, grep, etc.).

View name: %s
Parameters: %v`, viewName, userParams)
	},
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
