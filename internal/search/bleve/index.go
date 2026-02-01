package bleve

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	bsearch "github.com/blevesearch/bleve/v2/search"
	bindex "github.com/blevesearch/bleve_index_api"

	"github.com/zenobi-us/opennotes/internal/search"
)

// Index implements search.Index using Bleve full-text search.
type Index struct {
	mu       sync.RWMutex
	index    bleve.Index
	storage  search.Storage
	indexDir string
	status   search.IndexStatus
}

// Options configures the Bleve index.
type Options struct {
	// IndexDir is the directory to store the index (relative to storage root)
	// Defaults to ".opennotes/index"
	IndexDir string

	// InMemory creates an in-memory index (for testing)
	InMemory bool
}

// DefaultOptions returns default index options.
func DefaultOptions() Options {
	return Options{
		IndexDir: IndexDir,
		InMemory: false,
	}
}

// NewIndex creates or opens a Bleve index.
func NewIndex(storage search.Storage, opts Options) (*Index, error) {
	if opts.IndexDir == "" {
		opts.IndexDir = IndexDir
	}

	idx := &Index{
		storage:  storage,
		indexDir: opts.IndexDir,
		status:   search.IndexStatusUnopened,
	}

	var bleveIdx bleve.Index
	var err error

	if opts.InMemory {
		// Create in-memory index
		mapping := BuildDocumentMapping()
		bleveIdx, err = bleve.NewMemOnly(mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create in-memory index: %w", err)
		}
	} else {
		// Create or open on-disk index
		indexPath := filepath.Join(storage.Root(), opts.IndexDir)

		if _, statErr := os.Stat(indexPath); os.IsNotExist(statErr) {
			// Create new index
			mapping := BuildDocumentMapping()
			bleveIdx, err = bleve.New(indexPath, mapping)
			if err != nil {
				return nil, fmt.Errorf("failed to create index at %s: %w", indexPath, err)
			}
		} else {
			// Open existing index
			bleveIdx, err = bleve.Open(indexPath)
			if err != nil {
				return nil, fmt.Errorf("failed to open index at %s: %w", indexPath, err)
			}
		}
	}

	idx.index = bleveIdx
	idx.status = search.IndexStatusReady

	return idx, nil
}

// Add adds or updates a document in the index.
func (idx *Index) Add(ctx context.Context, doc search.Document) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if idx.index == nil {
		return search.ErrIndexClosed
	}

	// Convert to Bleve document
	bleveDoc := BleveDocument{
		Path:     doc.Path,
		Title:    doc.Title,
		Body:     doc.Body,
		Lead:     doc.Lead,
		Tags:     doc.Tags,
		Created:  doc.Created.Format(TimeFormat),
		Modified: doc.Modified.Format(TimeFormat),
		Checksum: doc.Checksum,
		Metadata: doc.Metadata,
	}

	// Use path as document ID
	return idx.index.Index(doc.Path, bleveDoc)
}

// Remove removes a document from the index by path.
func (idx *Index) Remove(ctx context.Context, path string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if idx.index == nil {
		return search.ErrIndexClosed
	}

	return idx.index.Delete(path)
}

// Find executes a search query and returns matching documents.
func (idx *Index) Find(ctx context.Context, opts search.FindOpts) (search.Results, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if idx.index == nil {
		return search.Results{}, search.ErrIndexClosed
	}

	start := time.Now()

	// Translate FindOpts to Bleve query
	query, err := TranslateFindOpts(opts)
	if err != nil {
		return search.Results{}, fmt.Errorf("failed to translate query: %w", err)
	}

	// Create search request
	req := bleve.NewSearchRequest(query)

	// Apply limit and offset
	if opts.Limit > 0 {
		req.Size = opts.Limit
	} else {
		req.Size = 100 // Default limit
	}
	req.From = opts.Offset

	// Apply sorting
	req.SortBy(translateSort(opts.Sort))

	// Request stored fields
	req.Fields = []string{
		FieldPath, FieldTitle, FieldLead, FieldTags,
		FieldCreated, FieldModified, FieldChecksum,
	}

	// Include snippets for body
	req.Highlight = bleve.NewHighlight()

	// Execute search
	result, err := idx.index.Search(req)
	if err != nil {
		return search.Results{}, fmt.Errorf("search failed: %w", err)
	}

	// Convert results
	items := make([]search.Result, 0, len(result.Hits))
	for _, hit := range result.Hits {
		doc := extractDocument(hit)
		snippets := extractSnippets(hit)

		items = append(items, search.Result{
			Document: doc,
			Score:    hit.Score,
			Snippets: snippets,
		})
	}

	return search.Results{
		Items:    items,
		Total:    int64(result.Total),
		Query:    opts,
		Duration: time.Since(start),
	}, nil
}

// FindByPath retrieves a single document by its exact path.
func (idx *Index) FindByPath(ctx context.Context, path string) (search.Document, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if idx.index == nil {
		return search.Document{}, search.ErrIndexClosed
	}

	// Use Document method for direct lookup by ID
	doc, err := idx.index.Document(path)
	if err != nil {
		return search.Document{}, err
	}
	if doc == nil {
		return search.Document{}, search.ErrNotFound
	}

	// Extract fields from internal document
	result := search.Document{Path: path}

	doc.VisitFields(func(field bindex.Field) {
		switch field.Name() {
		case FieldTitle:
			result.Title = string(field.Value())
		case FieldLead:
			result.Lead = string(field.Value())
		case FieldChecksum:
			result.Checksum = string(field.Value())
		}
	})

	return result, nil
}

// Count returns the number of documents matching the options.
func (idx *Index) Count(ctx context.Context, opts search.FindOpts) (int64, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if idx.index == nil {
		return 0, search.ErrIndexClosed
	}

	if opts.IsEmpty() {
		// Fast path: total document count
		count, err := idx.index.DocCount()
		return int64(count), err
	}

	// Translate and execute query
	query, err := TranslateFindOpts(opts)
	if err != nil {
		return 0, err
	}

	req := bleve.NewSearchRequest(query)
	req.Size = 0 // We only need the count

	result, err := idx.index.Search(req)
	if err != nil {
		return 0, err
	}

	return int64(result.Total), nil
}

// Reindex rebuilds the entire index from source files.
func (idx *Index) Reindex(ctx context.Context) error {
	idx.mu.Lock()
	idx.status = search.IndexStatusIndexing
	idx.mu.Unlock()

	defer func() {
		idx.mu.Lock()
		idx.status = search.IndexStatusReady
		idx.mu.Unlock()
	}()

	// Walk all markdown files and index them
	return idx.storage.Walk(".", func(path string, info search.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir {
			// Skip hidden directories
			if len(info.Name) > 0 && info.Name[0] == '.' {
				return search.SkipDir
			}
			return nil
		}

		if !isMarkdown(path) {
			return nil
		}

		// Read and parse the file
		content, readErr := idx.storage.Read(path)
		if readErr != nil {
			return readErr
		}

		// TODO: Parse frontmatter and extract document fields
		// For now, create a basic document
		doc := search.Document{
			Path:     path,
			Title:    filepath.Base(path),
			Body:     string(content),
			Modified: info.ModTime,
		}

		return idx.Add(ctx, doc)
	})
}

// Stats returns statistics about the index.
func (idx *Index) Stats(ctx context.Context) (search.IndexStats, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if idx.index == nil {
		return search.IndexStats{Status: search.IndexStatusUnopened}, nil
	}

	count, err := idx.index.DocCount()
	if err != nil {
		return search.IndexStats{}, err
	}

	return search.IndexStats{
		DocumentCount: int64(count),
		IndexPath:     filepath.Join(idx.storage.Root(), idx.indexDir),
		Status:        idx.status,
	}, nil
}

// Close releases resources held by the index.
func (idx *Index) Close() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if idx.index == nil {
		return nil
	}

	err := idx.index.Close()
	idx.index = nil
	idx.status = search.IndexStatusUnopened
	return err
}

// translateSort converts search.SortSpec to Bleve sort fields.
func translateSort(spec search.SortSpec) []string {
	if spec.Field == "" {
		spec.Field = search.SortByRelevance
	}

	var field string
	switch spec.Field {
	case search.SortByRelevance:
		return []string{"-_score"} // Descending by score
	case search.SortByCreated:
		field = FieldCreated
	case search.SortByModified:
		field = FieldModified
	case search.SortByTitle:
		field = FieldTitle
	case search.SortByPath:
		field = FieldPath
	default:
		field = string(spec.Field)
	}

	if spec.Direction == search.SortDesc {
		return []string{"-" + field}
	}
	return []string{field}
}

// extractDocument converts a Bleve search hit to a search.Document.
func extractDocument(hit *bsearch.DocumentMatch) search.Document {
	doc := search.Document{
		Path: hit.ID,
	}

	if v, ok := hit.Fields[FieldTitle].(string); ok {
		doc.Title = v
	}
	if v, ok := hit.Fields[FieldLead].(string); ok {
		doc.Lead = v
	}
	if v, ok := hit.Fields[FieldChecksum].(string); ok {
		doc.Checksum = v
	}

	// Parse tags
	if v, ok := hit.Fields[FieldTags]; ok {
		switch tags := v.(type) {
		case []interface{}:
			doc.Tags = make([]string, 0, len(tags))
			for _, t := range tags {
				if s, ok := t.(string); ok {
					doc.Tags = append(doc.Tags, s)
				}
			}
		case string:
			doc.Tags = []string{tags}
		}
	}

	// Parse dates
	if v, ok := hit.Fields[FieldCreated].(string); ok {
		if t, err := time.Parse(TimeFormat, v); err == nil {
			doc.Created = t
		}
	}
	if v, ok := hit.Fields[FieldModified].(string); ok {
		if t, err := time.Parse(TimeFormat, v); err == nil {
			doc.Modified = t
		}
	}

	return doc
}

// extractSnippets converts Bleve fragments to search.Snippets.
func extractSnippets(hit *bsearch.DocumentMatch) []search.Snippet {
	if hit.Fragments == nil {
		return nil
	}

	var snippets []search.Snippet
	for field, frags := range hit.Fragments {
		for _, frag := range frags {
			snippets = append(snippets, search.Snippet{
				Field: field,
				Text:  frag,
			})
		}
	}
	return snippets
}

// isMarkdown returns true if the path is a markdown file.
func isMarkdown(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".md" || ext == ".markdown"
}

// Field interface for visiting document fields.
type Field interface {
	Name() string
	Value() interface{}
}

// Ensure Index implements search.Index.
var _ search.Index = (*Index)(nil)
