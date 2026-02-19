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

func TestViewService_ExecuteView(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create test notebook and notes
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	// Add test notes with frontmatter
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "note1.md",
		map[string]interface{}{"tags": []string{"work"}, "status": "todo", "title": "Note 1"},
		"# Note 1\n\nWork-related content.")

	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "note2.md",
		map[string]interface{}{"tags": []string{"personal"}, "status": "done", "title": "Note 2"},
		"# Note 2\n\nPersonal content.")

	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "note3.md",
		map[string]interface{}{"status": "todo", "title": "Note 3"},
		"# Note 3\n\nNote without tags.")

	// Create index and services
	idx := testutil.CreateTestIndex(t, notebookDir)
	cfg, err := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	require.NoError(t, err)

	noteService := services.NewNoteService(cfg, idx, notebookDir)
	vs := services.NewViewService(cfg, notebookDir)

	// Set execution context
	vs.SetExecutionContext(idx, noteService)

	t.Run("executes simple filter view", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "work",
			Query: "tag:work",
		}

		results, err := vs.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Len(t, results.Notes, 1)
		assert.Contains(t, results.Notes[0].File.Relative, "note1.md")
	})

	t.Run("executes view with no filter (all notes)", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "all",
			Query: "|",
		}

		results, err := vs.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		require.NotNil(t, results)
		// Should return all 3 notes
		assert.Len(t, results.Notes, 3)
	})

	t.Run("executes view with limit", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "limited",
			Query: "| limit:1",
		}

		results, err := vs.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Len(t, results.Notes, 1)
	})

	t.Run("executes view with grouping", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "by-status",
			Query: "| group:status",
		}

		results, err := vs.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.NotNil(t, results.Groups)

		// Should have "todo" and "done" groups
		assert.Contains(t, results.Groups, "todo")
		assert.Contains(t, results.Groups, "done")

		// "todo" group should have 2 notes (note1 and note3)
		assert.Len(t, results.Groups["todo"], 2)
		// "done" group should have 1 note (note2)
		assert.Len(t, results.Groups["done"], 1)
	})

	t.Run("executes view with sort directive", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "sorted",
			Query: "| sort:title:asc",
		}

		results, err := vs.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		require.NotNil(t, results)
		// Should have all 3 notes
		assert.Len(t, results.Notes, 3)
	})

	t.Run("executes view with multiple directives", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "multi-directive",
			Query: "status:todo | sort:title:asc limit:2",
		}

		results, err := vs.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		require.NotNil(t, results)
		// Should have at most 2 notes with status:todo
		assert.LessOrEqual(t, len(results.Notes), 2)
	})

	t.Run("returns error for invalid filter DSL", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "invalid",
			Query: "::invalid::",
		}

		_, err := vs.ExecuteView(ctx, view, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse filter")
	})

	t.Run("returns error for invalid directive", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "bad-directive",
			Query: "| unknown:value",
		}

		_, err := vs.ExecuteView(ctx, view, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse directives")
	})

	t.Run("returns error without execution context", func(t *testing.T) {
		vs2 := services.NewViewService(cfg, notebookDir)
		// Don't set execution context

		view := &core.ViewDefinition{
			Name:  "test",
			Query: "tag:work",
		}

		_, err := vs2.ExecuteView(ctx, view, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "executor not initialized")
	})
}

func TestViewService_ExecuteView_SpecialViews(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create test notebook
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	// Create notes with links for orphan/broken-link testing
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "linked.md",
		map[string]interface{}{"title": "Linked Note", "links": []string{"orphan.md"}},
		"# Linked Note\n\nThis links to [orphan](orphan.md).")

	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "orphan.md",
		map[string]interface{}{"title": "Orphan Note"},
		"# Orphan Note\n\nNo one links to this.")

	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "broken.md",
		map[string]interface{}{"title": "Broken Note"},
		"# Broken Note\n\nThis links to [missing](nonexistent.md).")

	// Create index and services
	idx := testutil.CreateTestIndex(t, notebookDir)
	cfg, err := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	require.NoError(t, err)

	noteService := services.NewNoteService(cfg, idx, notebookDir)
	vs := services.NewViewService(cfg, notebookDir)
	vs.SetExecutionContext(idx, noteService)

	t.Run("delegates orphans view to special executor", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name: "orphans",
			Type: "special",
		}

		results, err := vs.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		require.NotNil(t, results)
		// Special view should return Notes (not Groups)
		assert.NotNil(t, results.Notes)
	})

	t.Run("delegates broken-links view to special executor", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name: "broken-links",
			Type: "special",
		}

		results, err := vs.ExecuteView(ctx, view, nil)
		require.NoError(t, err)
		require.NotNil(t, results)
		// Special view should return Notes (not Groups)
		assert.NotNil(t, results.Notes)
	})

	t.Run("returns error for unknown special view", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name: "unknown-special",
			Type: "special",
		}

		_, err := vs.ExecuteView(ctx, view, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown special view")
	})
}

func TestViewService_ExecuteView_ParameterSubstitution(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create test notebook
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "note1.md",
		map[string]interface{}{"tags": []string{"work"}, "title": "Work Note"},
		"# Work Note\n\nWork content.")
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "note2.md",
		map[string]interface{}{"tags": []string{"personal"}, "title": "Personal Note"},
		"# Personal Note\n\nPersonal content.")

	idx := testutil.CreateTestIndex(t, notebookDir)
	cfg, err := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	require.NoError(t, err)

	noteService := services.NewNoteService(cfg, idx, notebookDir)
	vs := services.NewViewService(cfg, notebookDir)
	vs.SetExecutionContext(idx, noteService)

	t.Run("substitutes parameters in query", func(t *testing.T) {
		view := &core.ViewDefinition{
			Name:  "parameterized",
			Query: "tag:{{tag_name}}",
			Parameters: []core.ViewParameter{
				{Name: "tag_name", Type: "string", Required: true},
			},
		}

		params := map[string]string{"tag_name": "work"}
		results, err := vs.ExecuteView(ctx, view, params)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Len(t, results.Notes, 1)
	})
}

func TestViewExecutor_GroupNotesByField(t *testing.T) {
	// Grouping logic is tested via ExecuteView with grouping directive
	// in TestViewService_ExecuteView "executes view with grouping" test case
	t.Skip("Grouping logic tested via ExecuteView integration tests")
}

func TestViewService_HasExecutionContext(t *testing.T) {
	cfg, err := services.NewConfigServiceWithPath(":memory:")
	require.NoError(t, err)

	vs := services.NewViewService(cfg, "")

	t.Run("returns false without context", func(t *testing.T) {
		assert.False(t, vs.HasExecutionContext())
	})

	t.Run("returns true after setting context", func(t *testing.T) {
		// Create a minimal index for the test
		tmpDir := t.TempDir()
		notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test")
		idx := testutil.CreateTestIndex(t, notebookDir)
		noteService := services.NewNoteService(cfg, idx, notebookDir)

		vs.SetExecutionContext(idx, noteService)
		assert.True(t, vs.HasExecutionContext())
	})
}
