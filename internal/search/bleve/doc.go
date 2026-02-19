// Package bleve implements the search.Index interface using Bleve full-text search.
//
// Bleve provides BM25 ranking and is a pure Go implementation with no CGO dependencies.
// This package translates the search.Query AST into Bleve queries and manages the
// index lifecycle.
//
// # Index Location
//
// The index is stored in `.opennotes/index/` within the notebook root directory.
// The directory is created automatically when the index is first opened.
//
// # Document Mapping
//
// Documents are indexed with field weights for BM25 ranking:
//   - path: 1000 (strongest signal for exact path matches)
//   - title: 500 (strong signal)
//   - tags: 300 (medium signal)
//   - lead: 50 (first paragraph)
//   - body: 1 (baseline)
//
// # Usage
//
//	storage := bleve.NewAferoStorage(afero.NewOsFs(), notebookRoot)
//	idx, err := bleve.NewIndex(storage, bleve.DefaultOptions())
//	if err != nil {
//	    return err
//	}
//	defer idx.Close()
//
//	// Add documents
//	err = idx.Add(ctx, search.Document{...})
//
//	// Search
//	results, err := idx.Find(ctx, search.FindOpts{}.WithTags("work"))
package bleve
