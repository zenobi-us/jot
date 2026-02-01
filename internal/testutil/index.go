package testutil

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/opennotes/internal/search"
	"github.com/zenobi-us/opennotes/internal/search/bleve"
	"gopkg.in/yaml.v3"
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

		// Parse frontmatter and content
		metadata, body := parseFrontmatter(content)

		// Create a document from the markdown file
		doc := search.Document{
			Path:     relPath,
			Title:    getTitle(metadata, filepath.Base(relPath)),
			Body:     body,
			Lead:     extractLead(body),
			Metadata: metadata,
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

// parseFrontmatter extracts YAML frontmatter from markdown content
func parseFrontmatter(content []byte) (map[string]any, string) {
	// Check for frontmatter delimiter
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return make(map[string]any), string(content)
	}

	// Find the end of frontmatter
	rest := content[4:] // Skip first "---\n"
	endIdx := bytes.Index(rest, []byte("\n---\n"))
	if endIdx == -1 {
		// No closing delimiter, treat as no frontmatter
		return make(map[string]any), string(content)
	}

	// Extract frontmatter and body
	frontmatterBytes := rest[:endIdx]
	bodyBytes := rest[endIdx+5:] // Skip "\n---\n"

	// Parse YAML frontmatter
	var metadata map[string]any
	if err := yaml.Unmarshal(frontmatterBytes, &metadata); err != nil {
		// Failed to parse, return empty metadata
		return make(map[string]any), string(content)
	}

	return metadata, string(bodyBytes)
}

// getTitle extracts title from metadata or uses filename
func getTitle(metadata map[string]any, defaultTitle string) string {
	if title, ok := metadata["title"].(string); ok && title != "" {
		return title
	}
	// Remove .md extension for default title
	return strings.TrimSuffix(defaultTitle, ".md")
}

// extractLead extracts the first paragraph from markdown content
func extractLead(body string) string {
	lines := strings.Split(body, "\n")
	var lead strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines at the start
		if lead.Len() == 0 && line == "" {
			continue
		}

		// Skip headings
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Stop at first empty line after content
		if lead.Len() > 0 && line == "" {
			break
		}

		// Add line to lead
		if line != "" {
			if lead.Len() > 0 {
				lead.WriteString(" ")
			}
			lead.WriteString(line)
		}
	}

	result := lead.String()
	if len(result) > 200 {
		return result[:200] + "..."
	}
	return result
}
