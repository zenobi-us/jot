package search

import (
	"time"
)

// FindOpts specifies options for finding documents.
//
// FindOpts uses the functional options pattern - methods return modified
// copies rather than mutating in place. This makes options thread-safe
// and chainable.
//
// Example:
//
//	opts := FindOpts{}.
//	    WithTags("work", "urgent").
//	    WithPath("projects/").
//	    ExcludingPaths("projects/archive/").
//	    WithLimit(10)
type FindOpts struct {
	// Query is parsed query AST (from Parser.Parse)
	Query *Query

	// RawQuery is the original query string (for error messages)
	RawQuery string

	// Tags filters by tag (all must match - AND)
	Tags []string

	// ExcludeTags excludes notes with these tags
	ExcludeTags []string

	// PathPrefix filters by path prefix (e.g., "projects/")
	PathPrefix string

	// ExcludePaths excludes notes with these path prefixes
	ExcludePaths []string

	// CreatedAfter filters notes created after this time
	CreatedAfter time.Time

	// CreatedBefore filters notes created before this time
	CreatedBefore time.Time

	// ModifiedAfter filters notes modified after this time
	ModifiedAfter time.Time

	// ModifiedBefore filters notes modified before this time
	ModifiedBefore time.Time

	// Metadata filters by arbitrary frontmatter fields
	Metadata map[string]any

	// Sort specifies the sort order
	Sort SortSpec

	// Limit is the maximum number of results (0 = no limit)
	Limit int

	// Offset is the number of results to skip (for pagination)
	Offset int
}

// SortSpec specifies how results should be sorted.
type SortSpec struct {
	// Field is the field to sort by
	Field SortField

	// Direction is the sort direction
	Direction SortDirection
}

// SortField specifies which field to sort by.
type SortField string

const (
	SortByRelevance SortField = "relevance" // Default: BM25 score
	SortByCreated   SortField = "created"
	SortByModified  SortField = "modified"
	SortByTitle     SortField = "title"
	SortByPath      SortField = "path"
)

// SortDirection specifies the sort order.
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// WithQuery returns a copy with the parsed query set.
func (o FindOpts) WithQuery(q *Query) FindOpts {
	o.Query = q
	return o
}

// WithRawQuery returns a copy with the raw query string set.
func (o FindOpts) WithRawQuery(raw string) FindOpts {
	o.RawQuery = raw
	return o
}

// WithTags returns a copy that filters by the given tags (AND).
func (o FindOpts) WithTags(tags ...string) FindOpts {
	o.Tags = append(o.Tags, tags...)
	return o
}

// ExcludingTags returns a copy that excludes notes with the given tags.
func (o FindOpts) ExcludingTags(tags ...string) FindOpts {
	o.ExcludeTags = append(o.ExcludeTags, tags...)
	return o
}

// WithPath returns a copy that filters by path prefix.
func (o FindOpts) WithPath(prefix string) FindOpts {
	o.PathPrefix = prefix
	return o
}

// ExcludingPaths returns a copy that excludes notes with the given path prefixes.
func (o FindOpts) ExcludingPaths(paths ...string) FindOpts {
	o.ExcludePaths = append(o.ExcludePaths, paths...)
	return o
}

// WithCreatedAfter returns a copy that filters by creation date.
func (o FindOpts) WithCreatedAfter(t time.Time) FindOpts {
	o.CreatedAfter = t
	return o
}

// WithCreatedBefore returns a copy that filters by creation date.
func (o FindOpts) WithCreatedBefore(t time.Time) FindOpts {
	o.CreatedBefore = t
	return o
}

// WithCreatedBetween returns a copy that filters by creation date range.
func (o FindOpts) WithCreatedBetween(after, before time.Time) FindOpts {
	o.CreatedAfter = after
	o.CreatedBefore = before
	return o
}

// WithModifiedAfter returns a copy that filters by modification date.
func (o FindOpts) WithModifiedAfter(t time.Time) FindOpts {
	o.ModifiedAfter = t
	return o
}

// WithModifiedBefore returns a copy that filters by modification date.
func (o FindOpts) WithModifiedBefore(t time.Time) FindOpts {
	o.ModifiedBefore = t
	return o
}

// WithMetadata returns a copy that filters by frontmatter field.
func (o FindOpts) WithMetadata(key string, value any) FindOpts {
	if o.Metadata == nil {
		o.Metadata = make(map[string]any)
	}
	o.Metadata[key] = value
	return o
}

// WithSort returns a copy with the specified sort order.
func (o FindOpts) WithSort(field SortField, direction SortDirection) FindOpts {
	o.Sort = SortSpec{Field: field, Direction: direction}
	return o
}

// WithLimit returns a copy with the specified result limit.
func (o FindOpts) WithLimit(limit int) FindOpts {
	o.Limit = limit
	return o
}

// WithOffset returns a copy with the specified offset (for pagination).
func (o FindOpts) WithOffset(offset int) FindOpts {
	o.Offset = offset
	return o
}

// IsEmpty returns true if no filters are set.
func (o FindOpts) IsEmpty() bool {
	return o.Query == nil &&
		o.RawQuery == "" &&
		len(o.Tags) == 0 &&
		len(o.ExcludeTags) == 0 &&
		o.PathPrefix == "" &&
		len(o.ExcludePaths) == 0 &&
		o.CreatedAfter.IsZero() &&
		o.CreatedBefore.IsZero() &&
		o.ModifiedAfter.IsZero() &&
		o.ModifiedBefore.IsZero() &&
		len(o.Metadata) == 0
}
