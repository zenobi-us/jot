package bleve

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zenobi-us/jot/internal/search"
)

func TestNewIndex_InMemory(t *testing.T) {
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = idx.Close() }()

	stats, err := idx.Stats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, search.IndexStatusReady, stats.Status)
	assert.Equal(t, int64(0), stats.DocumentCount)
}

func TestIndex_Add_Find(t *testing.T) {
	ctx := context.Background()
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = idx.Close() }()

	// Add a document
	doc := search.Document{
		Path:     "projects/todo.md",
		Title:    "Project Todo List",
		Body:     "This is a list of things to do for the project.",
		Lead:     "This is a list of things to do",
		Tags:     []string{"work", "urgent"},
		Created:  time.Now().Add(-24 * time.Hour),
		Modified: time.Now(),
		Checksum: "abc123",
	}

	err = idx.Add(ctx, doc)
	require.NoError(t, err)

	// Count should be 1
	count, err := idx.Count(ctx, search.FindOpts{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Find all documents
	results, err := idx.Find(ctx, search.FindOpts{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), results.Total)
	assert.Len(t, results.Items, 1)
	assert.Equal(t, "projects/todo.md", results.Items[0].Document.Path)
}

func TestIndex_Find_WithTags(t *testing.T) {
	ctx := context.Background()
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = idx.Close() }()

	// Add documents with different tags
	docs := []search.Document{
		{
			Path:     "work/meeting.md",
			Title:    "Meeting Notes",
			Body:     "Notes from the meeting.",
			Tags:     []string{"work", "meeting"},
			Modified: time.Now(),
		},
		{
			Path:     "personal/diary.md",
			Title:    "Personal Diary",
			Body:     "My personal thoughts.",
			Tags:     []string{"personal", "diary"},
			Modified: time.Now(),
		},
		{
			Path:     "work/project.md",
			Title:    "Project Plan",
			Body:     "The project plan document.",
			Tags:     []string{"work", "project"},
			Modified: time.Now(),
		},
	}

	for _, doc := range docs {
		err = idx.Add(ctx, doc)
		require.NoError(t, err)
	}

	// Find documents with "work" tag
	results, err := idx.Find(ctx, search.FindOpts{}.WithTags("work"))
	require.NoError(t, err)
	assert.Equal(t, int64(2), results.Total)

	// Find documents with "personal" tag
	results, err = idx.Find(ctx, search.FindOpts{}.WithTags("personal"))
	require.NoError(t, err)
	assert.Equal(t, int64(1), results.Total)
	assert.Equal(t, "personal/diary.md", results.Items[0].Document.Path)
}

func TestIndex_Find_ExcludeTags(t *testing.T) {
	ctx := context.Background()
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = idx.Close() }()

	// Add documents
	docs := []search.Document{
		{Path: "active.md", Tags: []string{"active"}, Modified: time.Now()},
		{Path: "archived.md", Tags: []string{"archived"}, Modified: time.Now()},
	}

	for _, doc := range docs {
		err = idx.Add(ctx, doc)
		require.NoError(t, err)
	}

	// Find excluding "archived"
	results, err := idx.Find(ctx, search.FindOpts{}.ExcludingTags("archived"))
	require.NoError(t, err)
	assert.Equal(t, int64(1), results.Total)
	assert.Equal(t, "active.md", results.Items[0].Document.Path)
}

func TestIndex_Find_PathPrefix(t *testing.T) {
	ctx := context.Background()
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = idx.Close() }()

	// Add documents in different directories
	docs := []search.Document{
		{Path: "projects/alpha/readme.md", Title: "Alpha", Modified: time.Now()},
		{Path: "projects/beta/readme.md", Title: "Beta", Modified: time.Now()},
		{Path: "notes/random.md", Title: "Random", Modified: time.Now()},
	}

	for _, doc := range docs {
		err = idx.Add(ctx, doc)
		require.NoError(t, err)
	}

	// Find documents in projects/
	results, err := idx.Find(ctx, search.FindOpts{}.WithPath("projects/"))
	require.NoError(t, err)
	assert.Equal(t, int64(2), results.Total)
}

func TestIndex_Remove(t *testing.T) {
	ctx := context.Background()
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = idx.Close() }()

	// Add and then remove a document
	doc := search.Document{
		Path:     "temp.md",
		Title:    "Temporary",
		Modified: time.Now(),
	}

	err = idx.Add(ctx, doc)
	require.NoError(t, err)

	count, err := idx.Count(ctx, search.FindOpts{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	err = idx.Remove(ctx, "temp.md")
	require.NoError(t, err)

	count, err = idx.Count(ctx, search.FindOpts{})
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestIndex_FindByPath(t *testing.T) {
	ctx := context.Background()
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)
	defer func() { _ = idx.Close() }()

	// Add a document
	doc := search.Document{
		Path:     "specific.md",
		Title:    "Specific Document",
		Lead:     "This is the lead",
		Checksum: "xyz789",
		Modified: time.Now(),
	}

	err = idx.Add(ctx, doc)
	require.NoError(t, err)

	// Find by path
	found, err := idx.FindByPath(ctx, "specific.md")
	require.NoError(t, err)
	assert.Equal(t, "specific.md", found.Path)
	assert.Equal(t, "Specific Document", found.Title)

	// Not found case
	_, err = idx.FindByPath(ctx, "nonexistent.md")
	assert.ErrorIs(t, err, search.ErrNotFound)
}

func TestIndex_Close(t *testing.T) {
	ctx := context.Background()
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	require.NoError(t, err)

	err = idx.Close()
	require.NoError(t, err)

	// Operations after close should fail
	_, err = idx.Find(ctx, search.FindOpts{})
	assert.ErrorIs(t, err, search.ErrIndexClosed)
}
