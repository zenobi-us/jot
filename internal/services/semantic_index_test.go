package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/jot/internal/search"
)

type mockSemanticIndex struct {
	available bool
	results   []SemanticResult
	err       error
	called    bool
}

func (m *mockSemanticIndex) FindSimilar(ctx context.Context, query string, opts SemanticFindOpts) ([]SemanticResult, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return m.results, nil
}

func (m *mockSemanticIndex) Close() error {
	return nil
}

func (m *mockSemanticIndex) IsAvailable() bool {
	return m.available
}

func TestNoopSemanticIndex_Unavailable(t *testing.T) {
	idx := NewNoopSemanticIndex()

	assert.False(t, idx.IsAvailable())

	results, err := idx.FindSimilar(context.Background(), "meeting", SemanticFindOpts{TopK: 5})
	assert.Nil(t, results)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrSemanticUnavailable))
}

func TestNoteService_SetSemanticIndex_NilUsesNoopFallback(t *testing.T) {
	svc := NewNoteService(nil, nil, "/tmp/notebook")

	svc.SetSemanticIndex(nil)
	assert.False(t, svc.SemanticAvailable())
}

func TestNoteService_FindSemanticCandidates_NoNotebook(t *testing.T) {
	svc := NewNoteService(nil, nil, "")

	results, err := svc.FindSemanticCandidates(context.Background(), "meeting", 10)
	assert.Nil(t, results)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no notebook selected")
}

func TestNoteService_FindSemanticCandidates_BackendUnavailable(t *testing.T) {
	svc := NewNoteService(nil, nil, "/tmp/notebook")
	svc.SetSemanticIndex(NewNoopSemanticIndex())

	results, err := svc.FindSemanticCandidates(context.Background(), "meeting", 10)
	assert.Nil(t, results)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrSemanticUnavailable))
}

func TestNoteService_FindSemanticCandidates_UsesConfiguredBackend(t *testing.T) {
	svc := NewNoteService(nil, nil, "/tmp/notebook")
	mock := &mockSemanticIndex{
		available: true,
		results: []SemanticResult{
			{Document: search.Document{Path: "notes/a.md"}, Score: 0.91},
		},
	}
	svc.SetSemanticIndex(mock)

	results, err := svc.FindSemanticCandidates(context.Background(), "meeting notes", 5)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.True(t, mock.called)
	assert.Equal(t, "notes/a.md", results[0].Document.Path)
}

func TestNotebookService_Open_UsesNoopSemanticFallback(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "semantic-open")

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)
	require.NotNil(t, notebook)
	require.NotNil(t, notebook.Notes)
	assert.False(t, notebook.Notes.SemanticAvailable())
}
