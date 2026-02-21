package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/core"
	"github.com/zenobi-us/jot/internal/services"
)

var (
	viewFormat         string
	viewParams         string
	viewList           bool
	viewSave           string
	viewDelete         string
	viewDescription    string
	viewSortOverride   string
	viewLimitOverride  int
	viewOffsetOverride int
	viewGroupOverride  string
)

type viewDirectiveOverrideState struct {
	Sort      string
	LimitSet  bool
	Limit     int
	OffsetSet bool
	Offset    int
	GroupSet  bool
	Group     string
}

var notesViewCmd = &cobra.Command{
	Use:   "view [name]",
	Short: "Execute a named reusable query preset or list available views",
	Long: `Execute a named query preset (view) with optional parameters, or list all available views.

When called without arguments or with --list, displays all available views.

BUILT-IN VIEWS:

  View queries use DSL pipe syntax: filter | directives

  today            modified:>=today | sort:modified:desc
  recent           | sort:modified:desc limit:20
  kanban           has:status | group:status sort:title:asc
  untagged         missing:tag | sort:created:desc
  orphans          (special) Notes with no incoming links
  broken-links     (special) Notes with broken references

DSL FILTER SYNTAX:

  Filters go before the pipe (|). Multiple filters combine with AND.

  Fields:
    tag:<value>           Notes with a specific tag
    status:<value>        Notes with a status value
    title:<text>          Search within title
    path:<prefix>         Notes under a path prefix
    body:<text>           Search within body content
    created:<date>        Notes created on date (YYYY-MM-DD)
    modified:<date>       Notes modified on date (YYYY-MM-DD)

  Operators (on date fields):
    field:>=value         Greater than or equal
    field:<=value         Less than or equal
    field:>value          Greater than
    field:<value          Less than

  Existence checks:
    has:<field>           Notes where field exists (e.g., has:tag)
    missing:<field>       Notes where field is absent (e.g., missing:status)

  Negation:
    -<term>               Exclude notes matching term
    -field:<value>        Exclude notes where field matches value

  Text:
    <word>                Full-text search term
    "multi word phrase"   Quoted phrase search

DIRECTIVES (after |):

  sort:<field>:<dir>    Sort by field: modified, created, title, path, relevance
                        Direction: asc (default) or desc
  limit:<n>             Return at most n results
  offset:<n>            Skip first n results (pagination)
  group:<field>         Group results by field (e.g., group:status)

CUSTOM VIEWS:

  Define custom views using the same DSL pipe syntax in notebook config:
  - Notebook: <notebook>/.jot.json

  Save/delete from CLI:
    jot notes view --save active-work "tag:work status:todo | sort:modified:desc limit:50" --description "Active work items"
    jot notes view --delete active-work

  Example custom view in .jot.json:
    {
      "views": {
        "active-work": {
          "name": "active-work",
          "description": "Active work items sorted by modification",
          "query": "tag:work status:todo | sort:modified:desc limit:50"
        }
      }
    }

OUTPUT FORMATS:

  list    - Simple list format (default)
  table   - ASCII table format
  json    - JSON array format

EXAMPLES:

  jot notes view                                    # List all views
  jot notes view --list                             # List all views
  jot notes view --list --format json               # List views as JSON
  jot notes view today                              # Notes modified today
  jot notes view recent                             # Last 20 modified notes
  jot notes view kanban                             # Notes grouped by status
  jot notes view untagged                           # Notes without tags
  jot notes view orphans --format json              # Orphaned notes as JSON
  jot notes view my-workflow --param sprint=Q1-S3   # Custom view with params
  jot notes view --save work-inbox "tag:work status:todo | sort:created:desc"
  jot notes view --save work-inbox "tag:work | sort:modified:desc" --description "Work queue"
  jot notes view --delete work-inbox

  DSL examples (use with custom views or 'notes search'):
    tag:work status:todo                                  # Work items that are todo
    tag:meeting modified:>=2026-01-01 | sort:modified:desc  # Recent meetings
    has:status -tag:archived | group:status sort:title:asc  # Kanban without archived
    missing:tag | sort:created:desc limit:10              # 10 newest untagged notes
    "project plan" | sort:relevance:desc                  # Phrase search, best match first`,

	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		overridesState := collectViewOverrideState(cmd)

		if err := validateViewCommandUsage(args, viewSave, viewDelete, viewList, viewDescription, viewParams, viewFormat, overridesState); err != nil {
			return err
		}

		if viewSave != "" {
			return handleViewSave(cmd, viewSave, viewDescription, args[0])
		}

		if viewDelete != "" {
			return handleViewDelete(cmd, viewDelete)
		}

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

		directiveOverrides, err := buildDirectiveOverrides(overridesState)
		if err != nil {
			return err
		}

		// Execute the view
		ctx := context.Background()
		results, err := vs.ExecuteViewWithOverrides(ctx, viewDef, userParams, directiveOverrides)
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

func collectViewOverrideState(cmd *cobra.Command) viewDirectiveOverrideState {
	state := viewDirectiveOverrideState{Sort: viewSortOverride}

	if cmd.Flags().Changed("limit") {
		state.LimitSet = true
		state.Limit = viewLimitOverride
	}

	if cmd.Flags().Changed("offset") {
		state.OffsetSet = true
		state.Offset = viewOffsetOverride
	}

	if cmd.Flags().Changed("group") {
		state.GroupSet = true
		state.Group = viewGroupOverride
	}

	return state
}

func buildDirectiveOverrides(state viewDirectiveOverrideState) (*services.ViewDirectiveOverrides, error) {
	hasOverride := false
	overrides := &services.ViewDirectiveOverrides{}

	if strings.TrimSpace(state.Sort) != "" {
		field, direction, err := parseSortOverride(state.Sort)
		if err != nil {
			return nil, err
		}
		overrides.SortField = field
		overrides.SortDirection = direction
		hasOverride = true
	}

	if state.LimitSet {
		if state.Limit <= 0 {
			return nil, fmt.Errorf("--limit override must be greater than 0")
		}
		limit := state.Limit
		overrides.Limit = &limit
		hasOverride = true
	}

	if state.OffsetSet {
		if state.Offset < 0 {
			return nil, fmt.Errorf("--offset override must be zero or greater")
		}
		offset := state.Offset
		overrides.Offset = &offset
		hasOverride = true
	}

	if state.GroupSet {
		group := strings.TrimSpace(state.Group)
		if group == "" {
			return nil, fmt.Errorf("--group override requires a field name")
		}
		groupLower := strings.ToLower(group)
		overrides.GroupBy = &groupLower
		hasOverride = true
	}

	if !hasOverride {
		return nil, nil
	}

	return overrides, nil
}

func parseSortOverride(value string) (string, string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", "", fmt.Errorf("--sort override cannot be empty")
	}

	parts := strings.Split(trimmed, ":")
	if len(parts) > 2 {
		return "", "", fmt.Errorf("invalid --sort value %q (use field[:asc|desc])", value)
	}

	field := strings.ToLower(strings.TrimSpace(parts[0]))
	if field == "" {
		return "", "", fmt.Errorf("--sort override requires a field name")
	}

	direction := "asc"
	if len(parts) == 2 {
		dir := strings.ToLower(strings.TrimSpace(parts[1]))
		if dir == "" {
			return "", "", fmt.Errorf("--sort override direction cannot be empty")
		}
		if dir != "asc" && dir != "desc" {
			return "", "", fmt.Errorf("invalid sort direction %q (use asc or desc)", parts[1])
		}
		direction = dir
	}

	return field, direction, nil
}

func validateViewCommandUsage(args []string, saveName, deleteName string, list bool, description, params, format string, overrides viewDirectiveOverrideState) error {
	if saveName != "" && deleteName != "" {
		return fmt.Errorf("cannot use --save and --delete together")
	}

	if description != "" && saveName == "" {
		return fmt.Errorf("--description can only be used with --save")
	}

	if err := ensureOverridesDisallowed("save", saveName != "", overrides); err != nil {
		return err
	}
	if err := ensureOverridesDisallowed("delete", deleteName != "", overrides); err != nil {
		return err
	}
	if err := ensureOverridesDisallowed("list", list, overrides); err != nil {
		return err
	}

	if saveName != "" {
		if params != "" {
			return fmt.Errorf("cannot use --param with --save")
		}
		if format != "list" {
			return fmt.Errorf("cannot use --format=%s with --save", format)
		}
	}

	if deleteName != "" {
		if params != "" {
			return fmt.Errorf("cannot use --param with --delete")
		}
		if format != "list" {
			return fmt.Errorf("cannot use --format=%s with --delete", format)
		}
	}

	if saveName != "" {
		if list {
			return fmt.Errorf("cannot combine --save with --list")
		}
		if len(args) != 1 {
			return fmt.Errorf(`--save requires exactly one query argument: jot notes view --save <name> "<query>"`)
		}
	}

	if deleteName != "" {
		if list {
			return fmt.Errorf("cannot combine --delete with --list")
		}
		if len(args) != 0 {
			return fmt.Errorf("--delete does not accept positional arguments")
		}
	}

	return nil
}

func ensureOverridesDisallowed(mode string, active bool, overrides viewDirectiveOverrideState) error {
	if !active {
		return nil
	}

	if overrides.Sort != "" {
		return overrideUsageError(mode, "sort")
	}
	if overrides.LimitSet {
		return overrideUsageError(mode, "limit")
	}
	if overrides.OffsetSet {
		return overrideUsageError(mode, "offset")
	}
	if overrides.GroupSet {
		return overrideUsageError(mode, "group")
	}

	return nil
}

func overrideUsageError(mode, flag string) error {
	if mode == "list" {
		return fmt.Errorf("cannot combine --%s with --list", flag)
	}
	return fmt.Errorf("cannot use --%s with --%s", flag, mode)
}

func handleViewSave(cmd *cobra.Command, name, description, query string) error {
	nb, err := requireNotebook(cmd)
	if err != nil {
		return err
	}

	notebookDir := filepath.Dir(nb.Config.Path)
	vs := services.NewViewService(cfgService, notebookDir)

	overwritten, err := vs.SaveNotebookView(&core.ViewDefinition{
		Name:        name,
		Description: description,
		Query:       query,
	})
	if err != nil {
		return fmt.Errorf("failed to save view '%s': %w", name, err)
	}

	if overwritten {
		fmt.Printf("Updated notebook view '%s' in %s\n", name, filepath.Join(notebookDir, services.NotebookConfigFile))
		return nil
	}

	fmt.Printf("Saved notebook view '%s' in %s\n", name, filepath.Join(notebookDir, services.NotebookConfigFile))
	return nil
}

func handleViewDelete(cmd *cobra.Command, name string) error {
	nb, err := requireNotebook(cmd)
	if err != nil {
		return err
	}

	notebookDir := filepath.Dir(nb.Config.Path)
	vs := services.NewViewService(cfgService, notebookDir)

	deleted, err := vs.DeleteNotebookView(name)
	if err != nil {
		return fmt.Errorf("failed to delete view '%s': %w", name, err)
	}

	if !deleted {
		return fmt.Errorf("view '%s' does not exist in notebook config", name)
	}

	fmt.Printf("Deleted notebook view '%s' from %s\n", name, filepath.Join(notebookDir, services.NotebookConfigFile))
	return nil
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
				fmt.Println("\nGlobal Views (from ~/.config/jot/config.json):")
			case "notebook":
				fmt.Println("\nNotebook Views (from <notebook>/.jot.json):")
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
	notesViewCmd.Flags().StringVar(&viewSave, "save", "", "Save a notebook view name")
	notesViewCmd.Flags().StringVar(&viewDelete, "delete", "", "Delete a notebook view name")
	notesViewCmd.Flags().StringVar(&viewDescription, "description", "", "Optional description when saving a view")
	notesViewCmd.Flags().StringVar(&viewSortOverride, "sort", "", "Override sort directive (field[:direction])")
	notesViewCmd.Flags().IntVar(&viewLimitOverride, "limit", 0, "Override result limit")
	notesViewCmd.Flags().IntVar(&viewOffsetOverride, "offset", 0, "Override result offset")
	notesViewCmd.Flags().StringVar(&viewGroupOverride, "group", "", "Override group directive (e.g., status)")

	notesCmd.AddCommand(notesViewCmd)
}
