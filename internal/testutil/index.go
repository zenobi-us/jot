package testutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/opennotes/internal/search"
	"github.com/zenobi-us/opennotes/internal/search/bleve"
)

// CreateTestIndex creates an in-memory Bleve index for testing.
// Optionally populates it with test documents from the given notebook directory.
func CreateTestIndex(t *testing.T, notebookDir string) search.Index {
	t.Helper()

	storage := bleve.MemStorage()
	idx, err := bleve.NewIndex(storage, bleve.Options{InMemory: true})
	require.NoError(t, err, "failed to create test index")

	t.Cleanup(func() {
		_ = idx.Close()
	})

	// If a notebook directory is provided, populate the index with markdown files
	if notebookDir != "" {
		populateIndexFromNotebook(t, idx, notebookDir)
	}

	return idx
}

// populateIndexFromNotebook walks the notebook directory and indexes all .md files
func populateIndexFromNotebook(t *testing.T, idx search.Index, notebookDir string) {
	t.Helper()

	docsAdded := 0

	// Walk the directory and index all markdown files
	err := filepath.Walk(notebookDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process markdown files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Get relative path from notebook root
		relPath, err := filepath.Rel(notebookDir, path)
		if err != nil {
			t.Logf("failed to get relative path: %v", err)
			return nil
		}

		// Read the file content
		content, err := os.ReadFile(path)
		if err != nil {
			t.Logf("failed to read file %s: %v", path, err)
			return nil
		}

		// Create a document from the markdown file
		// For now, use simple defaults - in real use, would parse frontmatter
		doc := search.Document{
			Path:     relPath,
			Title:    filepath.Base(relPath),
			Body:     string(content),
			Lead:     string(content),
			Created:  time.Now(),
			Modified: time.Now(),
			Checksum: "",
		}

		// Add to index
		ctx := context.Background()
		err = idx.Add(ctx, doc)
		if err != nil {
			t.Logf("failed to add document %s to index: %v", relPath, err)
			return nil
		}

		docsAdded++
		return nil
	})

	if err != nil {
		t.Logf("failed to walk notebook directory: %v", err)
	}
}
