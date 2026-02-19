package search

import "time"

// Results represents search results from a query.
type Results struct {
	// Items is the list of matching documents with scores
	Items []Result

	// Total is the total number of matches (before limit/offset)
	Total int64

	// Query is the query that produced these results
	Query FindOpts

	// Duration is how long the search took
	Duration time.Duration
}

// Result represents a single search result.
type Result struct {
	// Document is the matched document
	Document Document

	// Score is the relevance score (higher is better)
	// For BM25, this is typically in the range 0-25+
	Score float64

	// Snippets are context-aware excerpts showing matches
	Snippets []Snippet
}

// Snippet represents a text excerpt with highlighted matches.
type Snippet struct {
	// Field is the field this snippet comes from (body, title, etc.)
	Field string

	// Text is the excerpt text
	Text string

	// Ranges indicates the byte ranges of matched terms
	// Used for highlighting
	Ranges []MatchRange
}

// MatchRange represents a matched span in a snippet.
type MatchRange struct {
	// Start is the byte offset where the match begins
	Start int

	// End is the byte offset where the match ends
	End int
}

// Empty returns true if there are no results.
func (r Results) Empty() bool {
	return len(r.Items) == 0
}

// Paths returns just the paths of the matched documents.
func (r Results) Paths() []string {
	paths := make([]string, len(r.Items))
	for i, item := range r.Items {
		paths[i] = item.Document.Path
	}
	return paths
}

// Documents returns just the documents without scores.
func (r Results) Documents() []Document {
	docs := make([]Document, len(r.Items))
	for i, item := range r.Items {
		docs[i] = item.Document
	}
	return docs
}
