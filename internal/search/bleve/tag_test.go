package bleve

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/jot/internal/search"
)

// TestTagArrayIndexing tests that tags stored as arrays are properly indexed and searchable.
func TestTagArrayIndexing(t *testing.T) {
	ctx := context.Background()
	storage := MemStorage()
	index, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = index.Close() }()

	// Create a document with multiple tags
	doc := search.Document{
		Path:  "work-note.md",
		Title: "Project Meeting",
		Body:  "Discussion about the new project features.",
		Tags:  []string{"work", "meeting", "urgent"},
		Metadata: map[string]any{
			"status": "active",
		},
	}

	err = index.Add(ctx, doc)
	require.NoError(t, err)

	// Test 1: Query for single tag using FieldExpr (what data.tag=work does)
	t.Run("FieldExpr single tag", func(t *testing.T) {
		query := &search.Query{
			Expressions: []search.Expr{
				search.FieldExpr{
					Field: "tags",
					Op:    search.OpEquals,
					Value: "work",
				},
			},
		}

		results, err := index.Find(ctx, search.FindOpts{Query: query})
		require.NoError(t, err)
		assert.Len(t, results.Documents(), 1, "Should find document with 'work' tag")
		if len(results.Documents()) > 0 {
			assert.Equal(t, "work-note.md", results.Documents()[0].Path)
		}
	})

	// Test 2: Query for different tag in the array
	t.Run("FieldExpr different tag", func(t *testing.T) {
		query := &search.Query{
			Expressions: []search.Expr{
				search.FieldExpr{
					Field: "tags",
					Op:    search.OpEquals,
					Value: "meeting",
				},
			},
		}

		results, err := index.Find(ctx, search.FindOpts{Query: query})
		require.NoError(t, err)
		assert.Len(t, results.Documents(), 1, "Should find document with 'meeting' tag")
	})

	// Test 3: Query for non-existent tag
	t.Run("FieldExpr non-existent tag", func(t *testing.T) {
		query := &search.Query{
			Expressions: []search.Expr{
				search.FieldExpr{
					Field: "tags",
					Op:    search.OpEquals,
					Value: "personal",
				},
			},
		}

		results, err := index.Find(ctx, search.FindOpts{Query: query})
		require.NoError(t, err)
		assert.Len(t, results.Documents(), 0, "Should not find document without 'personal' tag")
	})

	// Test 4: Compare with FindOpts.Tags convenience filter
	t.Run("FindOpts Tags filter", func(t *testing.T) {
		results, err := index.Find(ctx, search.FindOpts{
			Tags: []string{"work"},
		})
		require.NoError(t, err)
		assert.Len(t, results.Documents(), 1, "FindOpts.Tags should also work")
	})
}

// TestMultipleDocumentsTagFiltering tests tag filtering with multiple documents.
func TestMultipleDocumentsTagFiltering(t *testing.T) {
	ctx := context.Background()
	storage := MemStorage()
	index, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = index.Close() }()

	// Add multiple documents with different tags
	docs := []search.Document{
		{
			Path:  "work1.md",
			Title: "Work Note 1",
			Body:  "Work content",
			Tags:  []string{"work", "urgent"},
		},
		{
			Path:  "work2.md",
			Title: "Work Note 2",
			Body:  "More work content",
			Tags:  []string{"work", "planning"},
		},
		{
			Path:  "personal1.md",
			Title: "Personal Note",
			Body:  "Personal content",
			Tags:  []string{"personal", "urgent"},
		},
	}

	for _, doc := range docs {
		err := index.Add(ctx, doc)
		require.NoError(t, err)
	}

	// Test: Query for "work" tag should return 2 documents
	t.Run("Query work tag", func(t *testing.T) {
		query := &search.Query{
			Expressions: []search.Expr{
				search.FieldExpr{
					Field: "tags",
					Op:    search.OpEquals,
					Value: "work",
				},
			},
		}

		results, err := index.Find(ctx, search.FindOpts{Query: query})
		require.NoError(t, err)
		assert.Len(t, results.Documents(), 2, "Should find 2 documents with 'work' tag")
	})

	// Test: Query for "urgent" tag should return 2 documents (1 work, 1 personal)
	t.Run("Query urgent tag", func(t *testing.T) {
		query := &search.Query{
			Expressions: []search.Expr{
				search.FieldExpr{
					Field: "tags",
					Op:    search.OpEquals,
					Value: "urgent",
				},
			},
		}

		results, err := index.Find(ctx, search.FindOpts{Query: query})
		require.NoError(t, err)
		assert.Len(t, results.Documents(), 2, "Should find 2 documents with 'urgent' tag")
	})

	// Test: Query for "planning" tag should return 1 document
	t.Run("Query planning tag", func(t *testing.T) {
		query := &search.Query{
			Expressions: []search.Expr{
				search.FieldExpr{
					Field: "tags",
					Op:    search.OpEquals,
					Value: "planning",
				},
			},
		}

		results, err := index.Find(ctx, search.FindOpts{Query: query})
		require.NoError(t, err)
		assert.Len(t, results.Documents(), 1, "Should find 1 document with 'planning' tag")
	})
}

// TestTagCaseInsensitivity tests that tag matching is case-insensitive.
func TestTagCaseInsensitivity(t *testing.T) {
	ctx := context.Background()
	storage := MemStorage()
	index, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = index.Close() }()

	// Add document with mixed-case tags
	doc := search.Document{
		Path:  "test.md",
		Title: "Test Note",
		Body:  "Test content",
		Tags:  []string{"Work", "Meeting", "URGENT"},
	}

	err = index.Add(ctx, doc)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		tagQuery string
	}{
		{"lowercase", "work"},
		{"uppercase", "WORK"},
		{"mixed case", "WoRk"},
		{"meeting lowercase", "meeting"},
		{"urgent lowercase", "urgent"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := &search.Query{
				Expressions: []search.Expr{
					search.FieldExpr{
						Field: "tags",
						Op:    search.OpEquals,
						Value: tc.tagQuery,
					},
				},
			}

			results, err := index.Find(ctx, search.FindOpts{Query: query})
			require.NoError(t, err)
			assert.Len(t, results.Documents(), 1, "Tag matching should be case-insensitive for: %s", tc.tagQuery)
		})
	}
}
