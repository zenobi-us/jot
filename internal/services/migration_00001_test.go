package services_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrations "github.com/zenobi-us/jot/internal/migrations"
	"github.com/zenobi-us/jot/internal/services"
)

func TestMigration00001_Up_RenamesLegacyConfig(t *testing.T) {
	notebookPath := t.TempDir()
	legacyPath := filepath.Join(notebookPath, ".opennotes.json")
	currentPath := filepath.Join(notebookPath, services.NotebookConfigFile)

	require.NoError(t, os.WriteFile(legacyPath, []byte(`{"name":"legacy"}`), 0644))

	m := migrations.Migration00001RenameOpenNotesConfig{}
	err := m.Up(context.Background(), services.Context{NotebookPath: notebookPath})
	require.NoError(t, err)

	_, err = os.Stat(legacyPath)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(currentPath)
	assert.NoError(t, err)
}

func TestMigration00001_Up_DryRun_NoChanges(t *testing.T) {
	notebookPath := t.TempDir()
	legacyPath := filepath.Join(notebookPath, ".opennotes.json")
	currentPath := filepath.Join(notebookPath, services.NotebookConfigFile)

	require.NoError(t, os.WriteFile(legacyPath, []byte(`{"name":"legacy"}`), 0644))

	m := migrations.Migration00001RenameOpenNotesConfig{}
	err := m.Up(context.Background(), services.Context{NotebookPath: notebookPath, DryRun: true})
	require.NoError(t, err)

	_, err = os.Stat(legacyPath)
	assert.NoError(t, err)
	_, err = os.Stat(currentPath)
	assert.True(t, os.IsNotExist(err))
}

func TestMigration00001_Down_RenamesBack(t *testing.T) {
	notebookPath := t.TempDir()
	legacyPath := filepath.Join(notebookPath, ".opennotes.json")
	currentPath := filepath.Join(notebookPath, services.NotebookConfigFile)

	require.NoError(t, os.WriteFile(currentPath, []byte(`{"name":"jot"}`), 0644))

	m := migrations.Migration00001RenameOpenNotesConfig{}
	err := m.Down(context.Background(), services.Context{NotebookPath: notebookPath})
	require.NoError(t, err)

	_, err = os.Stat(currentPath)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(legacyPath)
	assert.NoError(t, err)
}
