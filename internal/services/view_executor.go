// Package services provides view execution functionality.
package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/zenobi-us/jot/internal/core"
	"github.com/zenobi-us/jot/internal/search"
	"github.com/zenobi-us/jot/internal/search/parser"
)

// ViewResults holds the results of executing a view.
type ViewResults struct {
	// Notes is a flat list of notes (when not grouped)
	Notes []Note

	// Groups contains grouped results (when group directive is used)
	// Key is the group value (e.g., "todo", "done" for group:status)
	Groups map[string][]Note
}

// ViewDirectiveOverrides captures runtime overrides for directives supplied via CLI flags.
type ViewDirectiveOverrides struct {
	SortField     string
	SortDirection string
	Limit         *int
	Offset        *int
	GroupBy       *string
}

// ViewExecutor provides a unified interface for executing views.
// It requires a search.Index to execute queries.
type ViewExecutor struct {
	index       search.Index
	noteService *NoteService
}

// NewViewExecutor creates a new ViewExecutor with the given index.
func NewViewExecutor(index search.Index, noteService *NoteService) *ViewExecutor {
	return &ViewExecutor{
		index:       index,
		noteService: noteService,
	}
}

// ExecuteView executes a view definition and returns results.
// For DSL-based views, parses the query, builds FindOpts, and executes search.
// For special views, delegates to the special view executor.
func (ve *ViewExecutor) ExecuteView(ctx context.Context, view *core.ViewDefinition, params map[string]string, overrides *ViewDirectiveOverrides, viewService *ViewService) (*ViewResults, error) {
	// Handle special views
	if view.IsSpecialView() {
		return ve.executeSpecialView(ctx, view)
	}

	// Resolve template variables ({{today}}, {{param_name}}, etc.)
	resolvedQuery := view.Query
	if viewService != nil {
		resolvedQuery = viewService.ResolveTemplateVariables(view.Query)
	}

	// Apply parameter substitutions
	for name, value := range params {
		placeholder := "{{" + name + "}}"
		resolvedQuery = strings.ReplaceAll(resolvedQuery, placeholder, value)
	}

	// Split query into filter and directives
	filterPart, directivesPart := SplitViewQuery(resolvedQuery)

	// Parse directives
	directives, err := ParseDirectives(directivesPart)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directives: %w", err)
	}

	directives = applyDirectiveOverrides(directives, overrides)

	// Build FindOpts
	opts := search.FindOpts{
		Limit:  directives.Limit,
		Offset: directives.Offset,
	}

	// Parse filter DSL if present
	if filterPart != "" {
		p := parser.New()
		query, err := p.Parse(filterPart)
		if err != nil {
			return nil, fmt.Errorf("failed to parse filter: %w", err)
		}
		opts.Query = query
		opts.RawQuery = filterPart
	}

	// Set sort from directives
	if directives.SortField != "" {
		opts.Sort = ve.directiveToSortSpec(directives.SortField, directives.SortDirection)
	}

	// Execute search via index
	notes, err := ve.executeSearch(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Handle grouping
	if directives.GroupBy != "" {
		groups := ve.groupNotesByField(notes, directives.GroupBy)
		return &ViewResults{Groups: groups}, nil
	}

	return &ViewResults{Notes: notes}, nil
}

func applyDirectiveOverrides(base *ViewDirectives, overrides *ViewDirectiveOverrides) *ViewDirectives {
	if base == nil {
		base = &ViewDirectives{}
	}

	if overrides == nil {
		return base
	}

	if overrides.SortField != "" {
		base.SortField = overrides.SortField
		if overrides.SortDirection != "" {
			base.SortDirection = overrides.SortDirection
		}
	}

	if overrides.Limit != nil {
		base.Limit = *overrides.Limit
	}

	if overrides.Offset != nil {
		base.Offset = *overrides.Offset
	}

	if overrides.GroupBy != nil {
		base.GroupBy = *overrides.GroupBy
	}

	return base
}

// executeSearch executes a search with the given options and returns notes.
func (ve *ViewExecutor) executeSearch(ctx context.Context, opts search.FindOpts) ([]Note, error) {
	if ve.index == nil {
		return nil, fmt.Errorf("index not initialized")
	}

	// Get count if no limit is set (need all results)
	if opts.Limit == 0 {
		count, err := ve.index.Count(ctx, search.FindOpts{})
		if err != nil {
			return nil, fmt.Errorf("failed to count documents: %w", err)
		}
		if count == 0 {
			return []Note{}, nil
		}
		opts.Limit = int(count)
	}

	// Execute search
	results, err := ve.index.Find(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("index search failed: %w", err)
	}

	// Convert results to Notes
	notes := make([]Note, len(results.Items))
	for i, result := range results.Items {
		notes[i] = documentToNote(result.Document)
	}

	return notes, nil
}

// directiveToSortSpec converts directive strings to search.SortSpec
func (ve *ViewExecutor) directiveToSortSpec(field, direction string) search.SortSpec {
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

// groupNotesByField groups notes by a field value
func (ve *ViewExecutor) groupNotesByField(notes []Note, field string) map[string][]Note {
	groups := make(map[string][]Note)

	for _, note := range notes {
		key := ve.getNoteFieldValue(note, field)
		if key == "" {
			key = "(none)"
		}
		groups[key] = append(groups[key], note)
	}

	return groups
}

// getNoteFieldValue extracts a field value from a note for grouping
func (ve *ViewExecutor) getNoteFieldValue(note Note, field string) string {
	switch field {
	case "status":
		if status, ok := note.Metadata["status"]; ok {
			if str, ok := status.(string); ok {
				return str
			}
		}
	case "tag", "tags":
		// For tags, use first tag as group key
		if tags, ok := note.Metadata["tags"]; ok {
			switch t := tags.(type) {
			case []interface{}:
				if len(t) > 0 {
					if str, ok := t[0].(string); ok {
						return str
					}
				}
			case []string:
				if len(t) > 0 {
					return t[0]
				}
			case string:
				return t
			}
		}
	case "category":
		if cat, ok := note.Metadata["category"]; ok {
			if str, ok := cat.(string); ok {
				return str
			}
		}
	case "type":
		if typ, ok := note.Metadata["type"]; ok {
			if str, ok := typ.(string); ok {
				return str
			}
		}
	default:
		// Try to get arbitrary metadata field
		if val, ok := note.Metadata[field]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}

	return ""
}

// executeSpecialView dispatches to special view executor
func (ve *ViewExecutor) executeSpecialView(ctx context.Context, view *core.ViewDefinition) (*ViewResults, error) {
	if ve.noteService == nil {
		return nil, fmt.Errorf("note service not available for special view execution")
	}

	executor := NewSpecialViewExecutor(ve.noteService)

	switch view.Name {
	case "orphans":
		results, err := executor.ExecuteOrphansView(ctx, "no-incoming")
		if err != nil {
			return nil, err
		}
		return &ViewResults{Notes: convertMapSliceToNotes(results)}, nil

	case "broken-links":
		results, err := executor.ExecuteBrokenLinksView(ctx)
		if err != nil {
			return nil, err
		}
		return &ViewResults{Notes: convertMapSliceToNotes(results)}, nil

	default:
		return nil, fmt.Errorf("unknown special view: %s", view.Name)
	}
}

// convertMapSliceToNotes converts special view results to Note slice
func convertMapSliceToNotes(results []map[string]interface{}) []Note {
	notes := make([]Note, len(results))

	for i, result := range results {
		note := Note{
			Metadata: make(map[string]any),
		}

		// Extract file path
		if fp, ok := result["file_path"].(string); ok {
			note.File.Filepath = fp
		}
		if rp, ok := result["relative_path"].(string); ok {
			note.File.Relative = rp
		}

		// Extract content
		if body, ok := result["body"].(string); ok {
			note.Content = body
		}

		// Extract title
		if title, ok := result["title"].(string); ok {
			note.Metadata["title"] = title
		}

		// Copy other metadata
		for k, v := range result {
			if k != "file_path" && k != "relative_path" && k != "body" && k != "title" {
				note.Metadata[k] = v
			}
		}

		notes[i] = note
	}

	return notes
}

// ExecuteView is a convenience method on ViewService that creates a ViewExecutor and executes the view.
// Requires index and noteService to be set via SetExecutionContext.
func (vs *ViewService) ExecuteView(ctx context.Context, view *core.ViewDefinition, params map[string]string) (*ViewResults, error) {
	return vs.ExecuteViewWithOverrides(ctx, view, params, nil)
}

// ExecuteViewWithOverrides executes a view definition with directive overrides.
func (vs *ViewService) ExecuteViewWithOverrides(ctx context.Context, view *core.ViewDefinition, params map[string]string, overrides *ViewDirectiveOverrides) (*ViewResults, error) {
	if vs.executor == nil {
		return nil, fmt.Errorf("view executor not initialized - call SetExecutionContext first")
	}

	return vs.executor.ExecuteView(ctx, view, params, overrides, vs)
}

// SetExecutionContext sets the index and note service for view execution.
// Must be called before ExecuteView.
func (vs *ViewService) SetExecutionContext(index search.Index, noteService *NoteService) {
	vs.executor = NewViewExecutor(index, noteService)
}

// HasExecutionContext returns true if the execution context is set.
func (vs *ViewService) HasExecutionContext() bool {
	return vs.executor != nil
}
