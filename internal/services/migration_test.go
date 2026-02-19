package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationService_MigrateOpenNotesToJot_DryRun(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))
	t.Setenv("OPENNOTES_API_KEY", "secret")

	oldConfigPath := filepath.Join(tmpHome, ".config", "opennotes", "config.json")
	notebookPath := filepath.Join(tmpHome, "notes")
	require.NoError(t, os.MkdirAll(notebookPath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(notebookPath, ".opennotes.json"), []byte(`{"name":"Test","root":"."}`), 0644))
	writeLegacyConfig(t, oldConfigPath, Config{Notebooks: []string{notebookPath}})
	require.NoError(t, os.WriteFile(filepath.Join(tmpHome, ".bashrc"), []byte("export OPENNOTES_API_KEY=secret\n"), 0644))

	svc := NewMigrationService()
	report, err := svc.MigrateOpenNotesToJot(MigrationOptions{Apply: false})
	require.NoError(t, err)

	assert.False(t, report.Applied)
	assert.Equal(t, "would-migrate", report.GlobalConfig.Status)
	assert.Len(t, report.NotebookConfigs, 1)
	assert.Equal(t, "would-migrate", report.NotebookConfigs[0].Status)
	assert.Contains(t, report.LegacyEnvVars, "OPENNOTES_API_KEY")
	assert.NotEmpty(t, report.ProfileReferences)

	_, err = os.Stat(filepath.Join(tmpHome, ".config", "jot", "config.json"))
	assert.Error(t, err)
	_, err = os.Stat(filepath.Join(notebookPath, ".opennotes.json"))
	assert.NoError(t, err)
}

func TestMigrationService_MigrateOpenNotesToJot_Apply(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	oldConfigPath := filepath.Join(tmpHome, ".config", "opennotes", "config.json")
	notebookPath := filepath.Join(tmpHome, "notes")
	require.NoError(t, os.MkdirAll(notebookPath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(notebookPath, ".opennotes.json"), []byte(`{"name":"Test","root":"."}`), 0644))
	writeLegacyConfig(t, oldConfigPath, Config{Notebooks: []string{notebookPath}})

	svc := NewMigrationService()
	report, err := svc.MigrateOpenNotesToJot(MigrationOptions{Apply: true})
	require.NoError(t, err)

	assert.True(t, report.Applied)
	assert.Equal(t, "migrated", report.GlobalConfig.Status)
	assert.Len(t, report.NotebookConfigs, 1)
	assert.Equal(t, "migrated", report.NotebookConfigs[0].Status)

	_, err = os.Stat(filepath.Join(tmpHome, ".config", "jot", "config.json"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(notebookPath, ".jot.json"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(notebookPath, ".opennotes.json"))
	assert.Error(t, err)
}

func TestMigrationService_MigrateOpenNotesToJot_SkipsWhenTargetExists(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	oldConfigPath := filepath.Join(tmpHome, ".config", "opennotes", "config.json")
	notebookPath := filepath.Join(tmpHome, "notes")
	require.NoError(t, os.MkdirAll(notebookPath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(notebookPath, ".opennotes.json"), []byte(`{"name":"Old","root":"."}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(notebookPath, ".jot.json"), []byte(`{"name":"New","root":"."}`), 0644))
	writeLegacyConfig(t, oldConfigPath, Config{Notebooks: []string{notebookPath}})

	svc := NewMigrationService()
	report, err := svc.MigrateOpenNotesToJot(MigrationOptions{Apply: true})
	require.NoError(t, err)

	require.Len(t, report.NotebookConfigs, 1)
	assert.Equal(t, "skipped-target-exists", report.NotebookConfigs[0].Status)

	_, err = os.Stat(filepath.Join(notebookPath, ".opennotes.json"))
	assert.NoError(t, err)
}

func writeLegacyConfig(t *testing.T, path string, cfg Config) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	data, err := json.MarshalIndent(cfg, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0644))
}
