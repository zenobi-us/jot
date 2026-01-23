package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/services"
)

var (
	viewFormat string
	viewParams string
)

var notesViewCmd = &cobra.Command{
	Use:   "view <name>",
	Short: "Execute a named reusable query preset",
	Long: `Execute a named query preset (view) with optional parameters.

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

  opennotes notes view today
  opennotes notes view recent --format table
  opennotes notes view kanban --param status=todo,in-progress,done
  opennotes notes view orphans --format json
  opennotes notes view my-workflow --param sprint=Q1-S3`,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		viewName := args[0]

		// Get notebook
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		// Initialize ViewService
		vs := services.NewViewService(cfgService, nb.Config.Path)

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
		glob := fmt.Sprintf("%s/**/*.md", nb.Config.Path)
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
		defer rows.Close()

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

func init() {
	notesViewCmd.Flags().StringVar(&viewFormat, "format", "list", "Output format: list, table, or json")
	notesViewCmd.Flags().StringVar(&viewParams, "param", "", "View parameters (key=value,key2=value2)")

	notesCmd.AddCommand(notesViewCmd)
}
