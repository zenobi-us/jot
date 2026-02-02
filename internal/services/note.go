package services

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/zenobi-us/opennotes/internal/core"
	"github.com/zenobi-us/opennotes/internal/search"
)

// Note represents a markdown note.
type Note struct {
	File struct {
		Filepath string `json:"filepath"`
		Relative string `json:"relative"`
	} `json:"file"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
}

// DisplayName returns the display name for the note.
// Priority:
// 1. metadata["title"] if available
// 2. Slugified filename (without extension)
func (n *Note) DisplayName() string {
	// Check for title in metadata
	if title, ok := n.Metadata["title"]; ok {
		if titleStr, ok := title.(string); ok && titleStr != "" {
			return titleStr
		}
	}

	// Fallback to slugified filename
	filename := path.Base(n.File.Relative)
	// Remove .md extension
	filename = strings.TrimSuffix(filename, ".md")
	return core.Slugify(filename)
}

// documentToNote converts a search.Document to a Note.
// Preserves the Note struct format for backward compatibility.
func documentToNote(doc search.Document) Note {
	note := Note{
		Content:  doc.Body,
		Metadata: make(map[string]any),
	}

	note.File.Relative = doc.Path
	note.File.Filepath = doc.Path // Note: In index, Path is already relative

	// Map Document metadata back to Note metadata
	if doc.Title != "" {
		note.Metadata["title"] = doc.Title
	}
	if len(doc.Tags) > 0 {
		note.Metadata["tags"] = doc.Tags
	}
	if !doc.Created.IsZero() {
		note.Metadata["created"] = doc.Created
	}
	if !doc.Modified.IsZero() {
		note.Metadata["modified"] = doc.Modified
	}

	// Preserve any custom metadata
	if doc.Metadata != nil {
		for k, v := range doc.Metadata {
			note.Metadata[k] = v
		}
	}

	return note
}

// NoteService provides note query operations.
type NoteService struct {
	configService *ConfigService
	index         search.Index
	searchService *SearchService
	notebookPath  string
	log           zerolog.Logger
}

// NewNoteService creates a note service for a notebook.
func NewNoteService(cfg *ConfigService, index search.Index, notebookPath string) *NoteService {
	return &NoteService{
		configService: cfg,
		index:         index,
		searchService: NewSearchService(),
		notebookPath:  notebookPath,
		log:           Log("NoteService"),
	}
}

// SearchNotes returns all notes in the notebook matching the query.
// If fuzzy is true, uses fuzzy matching; otherwise uses exact text search.
func (s *NoteService) SearchNotes(ctx context.Context, query string, fuzzy bool) ([]Note, error) {
	if s.notebookPath == "" {
		return nil, fmt.Errorf("no notebook selected")
	}

	// Get all notes first
	notes, err := s.getAllNotes(ctx)
	if err != nil {
		return nil, err
	}

	// Apply search filtering
	if fuzzy {
		return s.searchService.FuzzySearch(query, notes), nil
	}

	return s.searchService.TextSearch(query, notes), nil
}

// getAllNotes retrieves all notes from the notebook without filtering.
// Uses the Bleve Index to retrieve all indexed documents and converts them to Notes.
func (s *NoteService) getAllNotes(ctx context.Context) ([]Note, error) {
	if s.index == nil {
		return nil, fmt.Errorf("index not initialized")
	}

	s.log.Debug().Msg("loading notes from index")

	// Query index for all documents (empty query matches all)
	results, err := s.index.Find(ctx, search.FindOpts{})
	if err != nil {
		return nil, fmt.Errorf("index query failed: %w", err)
	}

	// Convert search.Document results to Note objects
	notes := make([]Note, len(results.Items))
	for i, result := range results.Items {
		notes[i] = documentToNote(result.Document)
	}

	s.log.Debug().Int("count", len(notes)).Msg("notes loaded from index")
	return notes, nil
}

// Count returns the number of notes in the notebook.
func (s *NoteService) Count(ctx context.Context) (int, error) {
	if s.notebookPath == "" {
		return 0, nil
	}

	if s.index == nil {
		return 0, fmt.Errorf("index not initialized")
	}

	count, err := s.index.Count(ctx, search.FindOpts{})
	if err != nil {
		return 0, fmt.Errorf("index count failed: %w", err)
	}

	return int(count), nil
}

// SearchWithConditions executes a boolean query with the given conditions.
// Uses Bleve Index for querying instead of DuckDB SQL.
func (s *NoteService) SearchWithConditions(ctx context.Context, conditions []QueryCondition) ([]Note, error) {
	if s.notebookPath == "" {
		return nil, fmt.Errorf("no notebook selected")
	}

	// Build search.Query from conditions
	query, err := s.searchService.BuildQuery(ctx, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	s.log.Info().
		Int("conditionCount", len(conditions)).
		Bool("emptyQuery", query.IsEmpty()).
		Msg("executing boolean query")

	// Execute search using Index
	results, err := s.index.Find(ctx, search.FindOpts{
		Query: query,
		Sort: search.SortSpec{
			Field:     search.SortByPath,
			Direction: search.SortAsc,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results to Notes
	notes := make([]Note, len(results.Items))
	for i, result := range results.Items {
		notes[i] = documentToNote(result.Document)
	}

	s.log.Debug().Int("count", len(notes)).Msg("boolean query completed")
	return notes, nil
}

// ParseDataFlags parses --data flags in "field=value" format (exported for cmd package)
func ParseDataFlags(dataFlags []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, dataFlag := range dataFlags {
		parts := strings.SplitN(dataFlag, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --data format: %s (expected field=value)", dataFlag)
		}

		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Validate field name is not empty
		if field == "" {
			return nil, fmt.Errorf("field name cannot be empty in --data flag")
		}

		// Support multiple values for same field (convert to array)
		if existing, ok := result[field]; ok {
			switch v := existing.(type) {
			case []interface{}:
				result[field] = append(v, value)
			default:
				result[field] = []interface{}{v, value}
			}
		} else {
			result[field] = value
		}
	}

	return result, nil
}

// ResolvePath resolves the final note path based on input path and slugified title (exported for cmd package)
func ResolvePath(notebookRoot, inputPath, slugifiedTitle string) string {
	// Case 1: No path specified - use root + slugified title
	if inputPath == "" {
		return filepath.Join(notebookRoot, slugifiedTitle+".md")
	}

	// Case 2: Ends with "/" - explicit folder
	if strings.HasSuffix(inputPath, "/") {
		return filepath.Join(notebookRoot, inputPath, slugifiedTitle+".md")
	}

	// Case 3: Full filepath with .md extension
	if strings.HasSuffix(inputPath, ".md") {
		return filepath.Join(notebookRoot, inputPath)
	}

	// Case 4: Filepath without extension - auto-add .md
	return filepath.Join(notebookRoot, inputPath+".md")
}
