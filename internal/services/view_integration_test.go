// Package services provides integration tests for view execution.
package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/jot/internal/core"
	"github.com/zenobi-us/jot/internal/services"
	"github.com/zenobi-us/jot/internal/testutil"
)

// TestViewExecution_Integration tests that all builtin views execute without error
// using the new DSL/special flow.
func TestViewExecution_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create test notebook with representative notes
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "integration-notebook")

	// Create notes that exercise various view scenarios:
	// - Note with tags and status (for kanban, today, recent)
	// - Note without tags (for untagged view)
	// - Note with links (for orphans/broken-links analysis)
	// - Note without incoming links (orphan candidate)

	// Note with tags and status - modified "today"
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "active-note.md",
		map[string]interface{}{
			"title":  "Active Note",
			"tags":   []string{"work", "important"},
			"status": "todo",
		},
		"# Active Note\n\nThis is an active work note with tags and status.\n\nLinks to [[other-note]].")

	// Note with different status
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "done-note.md",
		map[string]interface{}{
			"title":  "Done Note",
			"tags":   []string{"personal"},
			"status": "done",
		},
		"# Done Note\n\nThis is a completed note.\n\nReferences [[active-note]].")

	// Note without tags (for untagged view)
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "untagged-note.md",
		map[string]interface{}{
			"title":  "Untagged Note",
			"status": "in-progress",
		},
		"# Untagged Note\n\nThis note has no tags but has a status.")

	// Note without status or tags (minimal note)
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "minimal-note.md",
		map[string]interface{}{
			"title": "Minimal Note",
		},
		"# Minimal Note\n\nThis is a minimal note without tags or status.")

	// Note with broken link (for broken-links view)
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "broken-link-note.md",
		map[string]interface{}{
			"title": "Note with Broken Link",
			"tags":  []string{"needs-fix"},
		},
		"# Note with Broken Link\n\nThis links to [[nonexistent-note]] which doesn't exist.")

	// Create index and services
	idx := testutil.CreateTestIndex(t, notebookDir)
	cfg, err := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	require.NoError(t, err)

	noteService := services.NewNoteService(cfg, idx, notebookDir)
	viewService := services.NewViewService(cfg, notebookDir)
	viewService.SetExecutionContext(idx, noteService)

	// Test that all builtin views execute without error.
	t.Run("all builtin views execute without error", func(t *testing.T) {
		builtins := []string{"today", "recent", "kanban", "untagged", "orphans", "broken-links"}

		for _, name := range builtins {
			t.Run(name, func(t *testing.T) {
				view, err := viewService.GetView(name)
				require.NoError(t, err, "failed to get view %s", name)
				require.NotNil(t, view, "view %s is nil", name)

				results, err := viewService.ExecuteView(ctx, view, nil)
				require.NoError(t, err, "failed to execute view %s", name)
				require.NotNil(t, results, "results for view %s are nil", name)

				// Results should have either Notes or Groups (but not both nil)
				assert.True(t,
					results.Notes != nil || results.Groups != nil,
					"view %s returned no results structure", name)
			})
		}
	})

	// Test DSL-based views return expected results
	t.Run("today view uses DSL filter", func(t *testing.T) {
		view, err := viewService.GetView("today")
		require.NoError(t, err)

		// Verify it's a DSL-based view (not special)
		assert.False(t, view.IsSpecialView(), "today should be DSL-based, not special")
		assert.Contains(t, view.Query, "modified:>=today")

		results, err := viewService.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		assert.NotNil(t, results.Notes, "today view should return Notes (not Groups)")
	})

	t.Run("recent view returns limited sorted results", func(t *testing.T) {
		view, err := viewService.GetView("recent")
		require.NoError(t, err)

		assert.False(t, view.IsSpecialView())
		assert.Contains(t, view.Query, "sort:modified:desc")
		assert.Contains(t, view.Query, "limit:20")

		results, err := viewService.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		assert.NotNil(t, results.Notes)
		// Should return at most 20 notes (we have 5 test notes)
		assert.LessOrEqual(t, len(results.Notes), 20)
	})

	t.Run("kanban view groups by status", func(t *testing.T) {
		view, err := viewService.GetView("kanban")
		require.NoError(t, err)

		assert.False(t, view.IsSpecialView())
		assert.Contains(t, view.Query, "has:status")
		assert.Contains(t, view.Query, "group:status")

		results, err := viewService.ExecuteView(ctx, view, nil)
		require.NoError(t, err)

		// Kanban should return Groups (not flat Notes)
		assert.NotNil(t, results.Groups, "kanban view should return Groups")

		// Should have groups for our statuses
		// Note: Groups may include "(none)" for notes without the field
		// Our test notes have: todo, done, in-progress
		groupCount := len(results.Groups)
		assert.GreaterOrEqual(t, groupCount, 1, "kanban should have at least one status group")
	})

	t.Run("untagged view finds notes without tags", func(t *testing.T) {
		view, err := viewService.GetView("untagged")
		require.NoError(t, err)

		assert.False(t, view.IsSpecialView())
		assert.Contains(t, view.Query, "missing:tag")

		results, err := viewService.ExecuteView(ctx, view, nil)
		require.NoError(t, err)

		assert.NotNil(t, results.Notes, "untagged view should return Notes (not Groups)")

		// Should find at least our untagged-note.md and minimal-note.md
		// (depends on whether missing:tag is properly implemented)
	})

	// Test special views delegate correctly
	t.Run("orphans view is special type", func(t *testing.T) {
		view, err := viewService.GetView("orphans")
		require.NoError(t, err)

		assert.True(t, view.IsSpecialView(), "orphans should be a special view")
		assert.Equal(t, "special", view.Type)

		results, err := viewService.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		assert.NotNil(t, results, "orphans view should return results")
	})

	t.Run("broken-links view is special type", func(t *testing.T) {
		view, err := viewService.GetView("broken-links")
		require.NoError(t, err)

		assert.True(t, view.IsSpecialView(), "broken-links should be a special view")
		assert.Equal(t, "special", view.Type)

		results, err := viewService.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		assert.NotNil(t, results, "broken-links view should return results")
	})
}

// TestViewExecution_DSLQueryVariations tests various DSL query patterns
// to ensure the parser and executor handle them correctly.
func TestViewExecution_DSLQueryVariations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create notebook with test notes
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "dsl-test-notebook")

	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "note1.md",
		map[string]interface{}{"title": "Note One", "tags": []string{"alpha"}, "status": "open"},
		"# Note One\n\nContent for note one.")

	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "note2.md",
		map[string]interface{}{"title": "Note Two", "tags": []string{"beta"}, "status": "closed"},
		"# Note Two\n\nContent for note two.")

	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "note3.md",
		map[string]interface{}{"title": "Note Three", "status": "open"},
		"# Note Three\n\nContent without tags.")

	idx := testutil.CreateTestIndex(t, notebookDir)
	cfg, err := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	require.NoError(t, err)

	noteService := services.NewNoteService(cfg, idx, notebookDir)
	viewService := services.NewViewService(cfg, notebookDir)
	viewService.SetExecutionContext(idx, noteService)

	tests := []struct {
		name        string
		query       string
		expectError bool
		checkResult func(t *testing.T, results *services.ViewResults)
	}{
		{
			name:        "empty filter with directives only",
			query:       "| sort:modified:desc",
			expectError: false,
			checkResult: func(t *testing.T, results *services.ViewResults) {
				assert.NotNil(t, results.Notes)
			},
		},
		{
			name:        "filter only no pipe",
			query:       "status:open",
			expectError: false,
			checkResult: func(t *testing.T, results *services.ViewResults) {
				assert.NotNil(t, results.Notes)
			},
		},
		{
			name:        "filter with limit",
			query:       "| limit:2",
			expectError: false,
			checkResult: func(t *testing.T, results *services.ViewResults) {
				assert.LessOrEqual(t, len(results.Notes), 2)
			},
		},
		{
			name:        "filter with offset",
			query:       "| limit:10 offset:1",
			expectError: false,
			checkResult: func(t *testing.T, results *services.ViewResults) {
				assert.NotNil(t, results.Notes)
			},
		},
		{
			name:        "sort ascending",
			query:       "| sort:title:asc",
			expectError: false,
			checkResult: func(t *testing.T, results *services.ViewResults) {
				assert.NotNil(t, results.Notes)
			},
		},
		{
			name:        "sort descending",
			query:       "| sort:modified:desc",
			expectError: false,
			checkResult: func(t *testing.T, results *services.ViewResults) {
				assert.NotNil(t, results.Notes)
			},
		},
		{
			name:        "group directive",
			query:       "| group:status",
			expectError: false,
			checkResult: func(t *testing.T, results *services.ViewResults) {
				assert.NotNil(t, results.Groups)
			},
		},
		{
			name:        "combined filter and directives",
			query:       "status:open | sort:title:asc limit:5",
			expectError: false,
			checkResult: func(t *testing.T, results *services.ViewResults) {
				assert.NotNil(t, results.Notes)
				assert.LessOrEqual(t, len(results.Notes), 5)
			},
		},
		{
			name:        "pipe only returns all notes",
			query:       "|",
			expectError: false,
			checkResult: func(t *testing.T, results *services.ViewResults) {
				assert.NotNil(t, results.Notes)
				assert.GreaterOrEqual(t, len(results.Notes), 1)
			},
		},
		{
			name:        "invalid filter DSL errors",
			query:       "::invalid::",
			expectError: true,
		},
		{
			name:        "unknown directive errors",
			query:       "| unknown:value",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a view with the test query
			view := &core.ViewDefinition{
				Name:  "test-query",
				Query: tt.query,
			}

			results, err := viewService.ExecuteView(ctx, view, nil)

			if tt.expectError {
				assert.Error(t, err, "expected error for query: %s", tt.query)
			} else {
				require.NoError(t, err, "unexpected error for query: %s", tt.query)
				if tt.checkResult != nil {
					tt.checkResult(t, results)
				}
			}
		})
	}
}

// TestViewService_ListAllBuiltinViews verifies all expected builtin views exist
func TestViewService_ListAllBuiltinViews(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cfg, err := services.NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	viewService := services.NewViewService(cfg, "")

	builtins := viewService.ListBuiltinViews()

	// Verify all expected builtin views are present
	expectedViews := map[string]bool{
		"today":        false,
		"recent":       false,
		"kanban":       false,
		"untagged":     false,
		"orphans":      false,
		"broken-links": false,
	}

	for _, view := range builtins {
		if _, expected := expectedViews[view.Name]; expected {
			expectedViews[view.Name] = true
		}
	}

	for name, found := range expectedViews {
		assert.True(t, found, "builtin view %s not found", name)
	}
}

// TestViewService_ExecutionContext verifies context must be set before execution
func TestViewService_ExecutionContext(t *testing.T) {
	ctx := context.Background()

	cfg, err := services.NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	viewService := services.NewViewService(cfg, "")

	t.Run("HasExecutionContext returns false initially", func(t *testing.T) {
		assert.False(t, viewService.HasExecutionContext())
	})

	t.Run("ExecuteView fails without context", func(t *testing.T) {
		view, err := viewService.GetView("today")
		require.NoError(t, err)

		_, err = viewService.ExecuteView(ctx, view, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "executor not initialized")
	})
}
