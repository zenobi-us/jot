package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

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

		// Get the view definition
		view, err := vs.GetView(viewName)
		if err != nil {
			return fmt.Errorf("failed to load view: %w", err)
		}

		// Parse user parameters
		userParams, err := vs.ParseViewParameters(viewParams)
		if err != nil {
			return fmt.Errorf("failed to parse parameters: %w", err)
		}

		// Generate SQL query (note: returns read_markdown() query with placeholder for glob)
		sqlQuery, sqlArgs, err := vs.GenerateSQL(view, userParams)
		if err != nil {
			return fmt.Errorf("failed to generate query: %w", err)
		}

		// Prepend glob pattern to args (read_markdown() requires it as first parameter)
		glob := fmt.Sprintf("%s/**/*.md", notebookDir)
		finalArgs := append([]interface{}{glob}, sqlArgs...)

		// Execute the query using raw database access
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		db, err := dbService.GetDB(ctx)
		if err != nil {
			return fmt.Errorf("database connection failed: %w", err)
		}

		rows, err := db.QueryContext(ctx, sqlQuery, finalArgs...)
		if err != nil {
			return fmt.Errorf("query execution failed: %w", err)
		}
		defer func() {
			_ = rows.Close()
		}()

		// Convert rows to map format for display
		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}

		var results []map[string]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			for i := range columns {
				values[i] = new(interface{})
			}
			if err := rows.Scan(values...); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}
			row := make(map[string]interface{})
			for i, col := range columns {
				row[col] = *(values[i].(*interface{}))
			}
			results = append(results, row)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("row iteration error: %w", err)
		}

		// Render results based on format
		display, err := services.NewDisplay()
		if err != nil {
			return fmt.Errorf("failed to create display: %w", err)
		}

		switch viewFormat {
		case "json":
			return display.RenderSQLResults(results)
		case "table":
			return display.RenderSQLResults(results)
		case "list":
			fallthrough
		default:
			return display.RenderSQLResults(results)
		}
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
