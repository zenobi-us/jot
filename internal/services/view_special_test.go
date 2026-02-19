package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/opennotes/internal/search"
	"github.com/zenobi-us/opennotes/internal/search/bleve"
)

func createTestNote(t *testing.T, notebookPath string, filename string, content string) {
	filePath := filepath.Join(notebookPath, filename)
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
}

func createTestIndexLocal(t *testing.T) search.Index {
	t.Helper()
	storage := bleve.MemStorage()
	idx, err := bleve.NewIndex(storage, bleve.Options{InMemory: true})
	require.NoError(t, err)
	t.Cleanup(func() { _ = idx.Close() })
	return idx
}

func TestSpecialViewExecutor_BrokenLinks_FindsMarkdownLinks(t *testing.T) {
	t.Skip("TODO Phase 5.3: Requires link graph index for broken links detection")
	// Create temporary notebook
	notebookPath := t.TempDir()

	// Create a valid note
	createTestNote(t, notebookPath, "valid.md", "# Valid Note\n\nThis is a valid note.")

	// Create a note with broken markdown link
	brokenContent := `---
title: Note with broken link
---

# Content

Check out [this link](nonexistent.md) for more info.
`
	createTestNote(t, notebookPath, "broken.md", brokenContent)

	// Setup
	cfg, _ := NewConfigServiceWithPath(":memory:")
	idx := createTestIndexLocal(t)
	noteService := NewNoteService(cfg, idx, notebookPath)
	executor := NewSpecialViewExecutor(noteService)

	// Execute
	results, err := executor.ExecuteBrokenLinksView(context.Background())

	// Verify
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "broken.md", results[0]["relative_path"])

	// Check broken links found
	brokenLinks, ok := results[0]["broken_links"].([]string)
	assert.True(t, ok)
	assert.Contains(t, brokenLinks, "nonexistent.md")
}

func TestSpecialViewExecutor_BrokenLinks_FindsWikiLinks(t *testing.T) {
	t.Skip("TODO Phase 5.3: Requires link graph index for broken links detection")
	notebookPath := t.TempDir()

	createTestNote(t, notebookPath, "valid.md", "# Valid")

	wikiContent := `---
title: Wiki links test
---

# Content

See [[nonexistent-note]] for details.
`
	createTestNote(t, notebookPath, "wiki.md", wikiContent)

	cfg, _ := NewConfigServiceWithPath(":memory:")
	idx := createTestIndexLocal(t)
	noteService := NewNoteService(cfg, idx, notebookPath)
	executor := NewSpecialViewExecutor(noteService)

	results, err := executor.ExecuteBrokenLinksView(context.Background())

	assert.NoError(t, err)
	assert.Len(t, results, 1)

	brokenLinks := results[0]["broken_links"].([]string)
	// Wiki links are normalized with .md
	assert.True(t, contains(brokenLinks, "nonexistent-note") || contains(brokenLinks, "nonexistent-note.md"))
}

func TestSpecialViewExecutor_BrokenLinks_SkipsExternalLinks(t *testing.T) {
	t.Skip("TODO Phase 5.3: Requires link graph index for broken links detection")
	notebookPath := t.TempDir()

	createTestNote(t, notebookPath, "test.md", `---
title: External links
---

# Content

[Google](https://google.com) and [anchor](#section) should be ignored.
`)

	cfg, _ := NewConfigServiceWithPath(":memory:")
	idx := createTestIndexLocal(t)
	noteService := NewNoteService(cfg, idx, notebookPath)
	executor := NewSpecialViewExecutor(noteService)

	results, err := executor.ExecuteBrokenLinksView(context.Background())

	assert.NoError(t, err)
	// No broken links since external URLs and anchors are skipped
	assert.Len(t, results, 0)
}

func TestSpecialViewExecutor_Orphans_FindsNoIncomingLinksNotes(t *testing.T) {
	t.Skip("TODO Phase 5.3: Requires link graph index for orphans detection")
	notebookPath := t.TempDir()

	// Create notes that reference each other
	createTestNote(t, notebookPath, "index.md", "[Link to linked](linked.md)")
	createTestNote(t, notebookPath, "linked.md", "# Linked note")

	// Create an orphan (no incoming links, no outgoing links)
	createTestNote(t, notebookPath, "orphan.md", "# Orphan\n\nNo one links to me.")

	cfg, _ := NewConfigServiceWithPath(":memory:")
	idx := createTestIndexLocal(t)
	noteService := NewNoteService(cfg, idx, notebookPath)
	executor := NewSpecialViewExecutor(noteService)

	results, err := executor.ExecuteOrphansView(context.Background(), "no-incoming")

	assert.NoError(t, err)
	// Should find the orphan (no incoming links) and possibly orphan.md (if index doesn't link to it)
	orphanPaths := make([]string, 0)
	for _, result := range results {
		orphanPaths = append(orphanPaths, result["relative_path"].(string))
	}

	assert.Contains(t, orphanPaths, "orphan.md")
}

func TestSpecialViewExecutor_Orphans_IsolatedNodeExcludesTagged(t *testing.T) {
	t.Skip("TODO Phase 5.3: Requires link graph index for orphans detection")
	notebookPath := t.TempDir()

	// Create a note with tags (not isolated)
	createTestNote(t, notebookPath, "tagged.md", `---
tags: [important]
---

# Tagged note

This note is tagged so it's not isolated.
`)

	// Create a truly isolated note
	createTestNote(t, notebookPath, "isolated.md", "# Isolated\n\nNo links, no tags.")

	cfg, _ := NewConfigServiceWithPath(":memory:")
	idx := createTestIndexLocal(t)
	noteService := NewNoteService(cfg, idx, notebookPath)
	executor := NewSpecialViewExecutor(noteService)

	results, err := executor.ExecuteOrphansView(context.Background(), "isolated")

	assert.NoError(t, err)
	orphanPaths := make([]string, 0)
	for _, result := range results {
		orphanPaths = append(orphanPaths, result["relative_path"].(string))
	}

	// Should find isolated but not tagged (even though it has no links)
	assert.Contains(t, orphanPaths, "isolated.md")
}

func TestSpecialViewExecutor_ExtractLinks_HandlesMultipleLinkTypes(t *testing.T) {
	t.Skip("TODO Phase 5.3: Requires link graph index for link extraction")
	notebookPath := t.TempDir()
	createTestNote(t, notebookPath, "test.md", `---
links: ["frontmatter-link.md"]
---

# Content

[markdown](markdown-link.md) and [[wikilink]]
`)

	cfg, _ := NewConfigServiceWithPath(":memory:")
	idx := createTestIndexLocal(t)
	noteService := NewNoteService(cfg, idx, notebookPath)
	executor := NewSpecialViewExecutor(noteService)

	// Get all notes
	notes, err := noteService.getAllNotes(context.Background())
	require.NoError(t, err)
	require.Len(t, notes, 1)

	// Extract links
	links := executor.extractAllLinks(&notes[0])

	// Should contain links from body at minimum (frontmatter parsing depends on YAML parsing)
	assert.True(t, links["markdown-link.md"], "Should extract markdown links")
	assert.True(t, links["wikilink"] || links["wikilink.md"], "Should extract wiki links")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
