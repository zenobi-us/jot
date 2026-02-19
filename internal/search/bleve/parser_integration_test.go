package bleve

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/opennotes/internal/search"
)

func TestIndex_FindByQueryString(t *testing.T) {
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = idx.Close() }()

	ctx := context.Background()

	// Add test documents
	docs := []search.Document{
		{
			Path:     "work/meeting.md",
			Title:    "Team Meeting Notes",
			Body:     "Discussed the new project timeline and deliverables.",
			Tags:     []string{"work", "meeting"},
			Modified: time.Now(),
		},
		{
			Path:     "personal/journal.md",
			Title:    "Personal Journal",
			Body:     "Reflections on the day and personal thoughts.",
			Tags:     []string{"personal", "journal"},
			Modified: time.Now(),
		},
		{
			Path:     "work/archived/old.md",
			Title:    "Old Project",
			Body:     "Archived project documentation.",
			Tags:     []string{"work", "archived"},
			Modified: time.Now(),
		},
	}

	for _, doc := range docs {
		err = idx.Add(ctx, doc)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		query         string
		expectedCount int64
		expectedPaths []string
	}{
		{
			name:          "simple term search",
			query:         "meeting",
			expectedCount: 1,
			expectedPaths: []string{"work/meeting.md"},
		},
		{
			name:          "tag search",
			query:         "tag:work",
			expectedCount: 2,
			expectedPaths: []string{"work/meeting.md", "work/archived/old.md"},
		},
		{
			name:          "exclude tag",
			query:         "tag:work -tag:archived",
			expectedCount: 1,
			expectedPaths: []string{"work/meeting.md"},
		},
		{
			name:          "field search",
			query:         "title:journal",
			expectedCount: 1,
			expectedPaths: []string{"personal/journal.md"},
		},
		{
			name:          "combined query",
			query:         "project tag:work",
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := idx.FindByQueryString(ctx, tt.query, search.FindOpts{})
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, results.Total, "wrong result count")

			if len(tt.expectedPaths) > 0 {
				gotPaths := make([]string, len(results.Items))
				for i, item := range results.Items {
					gotPaths[i] = item.Document.Path
				}
				assert.ElementsMatch(t, tt.expectedPaths, gotPaths, "wrong paths returned")
			}
		})
	}
}

func TestIndex_FindByQueryString_InvalidQuery(t *testing.T) {
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = idx.Close() }()

	ctx := context.Background()

	// Test with invalid query syntax
	_, err = idx.FindByQueryString(ctx, "field::", search.FindOpts{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse query")
}
