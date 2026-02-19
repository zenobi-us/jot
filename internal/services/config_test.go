package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigService_Defaults(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "opennotes", "config.json")

	// No config file exists, should use defaults
	svc, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	// Default notebooks should be relative to config path
	expectedNotebooks := filepath.Join(tmpDir, "opennotes", "notebooks")
	assert.Equal(t, []string{expectedNotebooks}, svc.Store.Notebooks)
	assert.Equal(t, "", svc.Store.NotebookPath)
}

func TestNewConfigService_LoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "opennotes", "config.json")

	// Create config file
	config := Config{
		Notebooks:    []string{"/path/to/notebooks", "/another/path"},
		NotebookPath: "/current/notebook",
	}
	createTestConfigFile(t, configPath, config)

	svc, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	assert.Equal(t, []string{"/path/to/notebooks", "/another/path"}, svc.Store.Notebooks)
	assert.Equal(t, "/current/notebook", svc.Store.NotebookPath)
}

func TestNewConfigService_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "opennotes", "config.json")

	// Create invalid JSON file
	err := os.MkdirAll(filepath.Dir(configPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(configPath, []byte("{ invalid json }"), 0644)
	require.NoError(t, err)

	// Should still succeed (falls back to defaults) but log a warning
	svc, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	// Should have defaults since file was invalid
	expectedNotebooks := filepath.Join(tmpDir, "opennotes", "notebooks")
	assert.Equal(t, []string{expectedNotebooks}, svc.Store.Notebooks)
}

func TestNewConfigService_EnvVarOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "opennotes", "config.json")

	// Set environment variable
	t.Setenv("OPENNOTES_NOTEBOOKPATH", "/env/notebook")

	svc, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	assert.Equal(t, "/env/notebook", svc.Store.NotebookPath)
}

func TestNewConfigService_EnvVarPriorityOverFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "opennotes", "config.json")

	// Create config file with notebookPath
	config := Config{
		Notebooks:    []string{"/path/to/notebooks"},
		NotebookPath: "/file/notebook",
	}
	createTestConfigFile(t, configPath, config)

	// Set environment variable (should override file)
	t.Setenv("OPENNOTES_NOTEBOOKPATH", "/env/notebook")

	svc, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	// Env var should take priority
	assert.Equal(t, "/env/notebook", svc.Store.NotebookPath)
	// File value for notebooks should still be loaded
	assert.Equal(t, []string{"/path/to/notebooks"}, svc.Store.Notebooks)
}

func TestConfigService_Write_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "opennotes", "config.json")

	// Create service (directory doesn't exist yet)
	svc, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	// Write config (should create directory)
	newConfig := Config{
		Notebooks:    []string{"/new/path"},
		NotebookPath: "/new/notebook",
	}
	err = svc.Write(newConfig)
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(filepath.Dir(configPath))
	assert.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}

func TestConfigService_Write_PersistsConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "opennotes", "config.json")

	svc, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	// Write new config
	newConfig := Config{
		Notebooks:    []string{"/path/a", "/path/b"},
		NotebookPath: "/current/notebook",
	}
	err = svc.Write(newConfig)
	require.NoError(t, err)

	// Verify internal store is updated
	assert.Equal(t, newConfig, svc.Store)

	// Verify file contains correct JSON
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var readConfig Config
	err = json.Unmarshal(data, &readConfig)
	require.NoError(t, err)

	assert.Equal(t, []string{"/path/a", "/path/b"}, readConfig.Notebooks)
	assert.Equal(t, "/current/notebook", readConfig.NotebookPath)
}

func TestConfigService_Write_Roundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "opennotes", "config.json")

	// Create and write config
	svc1, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	config := Config{
		Notebooks:    []string{"/notebooks/work", "/notebooks/personal"},
		NotebookPath: "/notebooks/work",
	}
	err = svc1.Write(config)
	require.NoError(t, err)

	// Create new service that reads the written config
	svc2, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	assert.Equal(t, config.Notebooks, svc2.Store.Notebooks)
	assert.Equal(t, config.NotebookPath, svc2.Store.NotebookPath)
}

func TestGlobalConfigFile(t *testing.T) {
	path := GlobalConfigFile()

	// Should end with expected suffix
	assert.Contains(t, path, "opennotes")
	assert.True(t, filepath.IsAbs(path), "config path should be absolute")
	assert.True(t,
		filepath.Base(path) == "config.json",
		"config file should be named config.json",
	)
}

func TestGlobalConfigFile_EnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	overridePath := filepath.Join(tmpDir, "custom", "config.json")

	t.Setenv("OPENNOTES_CONFIG", overridePath)

	path := GlobalConfigFile()

	assert.Equal(t, overridePath, path)
}

func TestConfigService_Path(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "opennotes", "config.json")

	svc, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	assert.Equal(t, configPath, svc.Path())
}

// Helper function to create test config files
func createTestConfigFile(t *testing.T, path string, config Config) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(t, err)

	data, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(path, data, 0644)
	require.NoError(t, err)
}
