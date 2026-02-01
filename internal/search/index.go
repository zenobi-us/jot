package search

import (
	"context"
	"time"
)

// Index defines the contract for a note search index.
//
// An Index is responsible for indexing notes and executing queries.
// Implementations must be safe for concurrent use.
//
// Example usage:
//
//	idx := bleve.NewIndex(storage, opts)
//	err := idx.Add(ctx, doc)
//	results, err := idx.Find(ctx, FindOpts{}.WithTags("work"))
type Index interface {
	// Add adds or updates a document in the index.
	// If a document with the same path already exists, it is replaced.
	Add(ctx context.Context, doc Document) error

	// Remove removes a document from the index by path.
	// Returns nil if the document doesn't exist.
	Remove(ctx context.Context, path string) error

	// Find executes a search query and returns matching documents.
	// The query is specified through FindOpts.
	Find(ctx context.Context, opts FindOpts) (Results, error)

	// FindByPath retrieves a single document by its exact path.
	// Returns ErrNotFound if the document doesn't exist.
	FindByPath(ctx context.Context, path string) (Document, error)

	// Count returns the number of documents matching the options.
	// If opts is zero value, returns total document count.
	Count(ctx context.Context, opts FindOpts) (int64, error)

	// Reindex rebuilds the entire index from source files.
	// This is an expensive operation and should be used sparingly.
	Reindex(ctx context.Context) error

	// Stats returns statistics about the index.
	Stats(ctx context.Context) (IndexStats, error)

	// Close releases resources held by the index.
	Close() error
}

// Document represents an indexed note.
//
// This is the canonical representation of a note for indexing purposes.
// It contains all searchable fields extracted from the markdown file.
type Document struct {
	// Path is the relative path from notebook root (e.g., "projects/todo.md")
	Path string

	// Title is extracted from frontmatter or first heading
	Title string

	// Body is the full text content (markdown stripped of frontmatter)
	Body string

	// Lead is the first paragraph, used for snippets
	Lead string

	// Tags from frontmatter (normalized, lowercase)
	Tags []string

	// Metadata contains arbitrary frontmatter fields
	Metadata map[string]any

	// Created is the note creation time (from frontmatter or file stat)
	Created time.Time

	// Modified is the last modification time
	Modified time.Time

	// Checksum for change detection (e.g., xxhash of content)
	Checksum string
}

// IndexStats contains statistics about the index.
type IndexStats struct {
	// DocumentCount is the total number of indexed documents
	DocumentCount int64

	// IndexSize is the size of the index in bytes
	IndexSize int64

	// LastIndexed is when the last document was indexed
	LastIndexed time.Time

	// IndexPath is the filesystem path to the index (if persisted)
	IndexPath string

	// Status is the current index status
	Status IndexStatus
}

// IndexStatus represents the current state of the index.
type IndexStatus string

const (
	IndexStatusReady    IndexStatus = "ready"
	IndexStatusIndexing IndexStatus = "indexing"
	IndexStatusError    IndexStatus = "error"
	IndexStatusUnopened IndexStatus = "unopened"
)
