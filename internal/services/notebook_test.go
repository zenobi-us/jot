package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions

// createTestNotebook creates a notebook directory with config for testing.
func createTestNotebook(t *testing.T, dir, name string) string {
	t.Helper()

	notebookDir := filepath.Join(dir, name)
	notesDir := filepath.Join(notebookDir, ".notes")

	require.NoError(t, os.MkdirAll(notesDir, 0755))

	config := StoredNotebookConfig{
		Name:     name,
		Root:     ".notes",
		Contexts: []string{notebookDir},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err)

	configPath := filepath.Join(notebookDir, NotebookConfigFile)
	require.NoError(t, os.WriteFile(configPath, data, 0644))

	return notebookDir
}

// createTestConfigService creates a ConfigService with a test config file.
func createTestConfigService(t *testing.T, tmpDir string, notebooks []string) *ConfigService {
	t.Helper()

	configPath := filepath.Join(tmpDir, "opennotes", "config.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))

	config := Config{
		Notebooks:    notebooks,
		NotebookPath: "",
	}

	data, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, data, 0644))

	svc, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)

	return svc
}

// HasNotebook tests

func TestNotebookService_HasNotebook_ExistsTrue(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "test-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc, NewDbService())

	assert.True(t, svc.HasNotebook(notebookDir))
}

func TestNotebookService_HasNotebook_NotExistsFalse(t *testing.T) {
	tmpDir := t.TempDir()

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc, NewDbService())

	nonExistent := filepath.Join(tmpDir, "non-existent")
	assert.False(t, svc.HasNotebook(nonExistent))
}

func TestNotebookService_HasNotebook_EmptyPath(t *testing.T) {
	tmpDir := t.TempDir()

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc, NewDbService())

	assert.False(t, svc.HasNotebook(""))
}

// LoadConfig tests

func TestNotebookService_LoadConfig_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "test-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc, NewDbService())

	config, err := svc.LoadConfig(notebookDir)
	require.NoError(t, err)

	assert.Equal(t, "test-notebook", config.Name)
	assert.Equal(t, filepath.Join(notebookDir, ".notes"), config.Root)
	assert.Equal(t, []string{notebookDir}, config.Contexts)
}

func TestNotebookService_LoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "invalid")

	require.NoError(t, os.MkdirAll(notebookDir, 0755))
	configPath := filepath.Join(notebookDir, NotebookConfigFile)
	require.NoError(t, os.WriteFile(configPath, []byte("{ invalid json }"), 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc, NewDbService())

	_, err := svc.LoadConfig(notebookDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid notebook config")
}

func TestNotebookService_LoadConfig_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "missing")

	require.NoError(t, os.MkdirAll(notebookDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc, NewDbService())

	_, err := svc.LoadConfig(notebookDir)
	assert.Error(t, err)
}

func TestNotebookService_LoadConfig_CreatesRootIfMissing(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notebook")

	require.NoError(t, os.MkdirAll(notebookDir, 0755))

	// Create config pointing to non-existent root
	config := StoredNotebookConfig{
		Name: "test",
		Root: "notes-missing",
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	configPath := filepath.Join(notebookDir, NotebookConfigFile)
	require.NoError(t, os.WriteFile(configPath, data, 0644))

	configSvc := createTestConfigService(t, tmpDir, nil)
	svc := NewNotebookService(configSvc, NewDbService())

	loadedConfig, err := svc.LoadConfig(notebookDir)
	require.NoError(t, err)

	// Root directory should have been created
	rootPath := filepath.Join(notebookDir, "notes-missing")
	_, err = os.Stat(rootPath)
	assert.NoError(t, err)
	assert.Equal(t, rootPath, loadedConfig.Root)
}

// Open tests

func TestNotebookService_Open_Success(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "test-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	assert.Equal(t, "test-notebook", notebook.Config.Name)
}

func TestNotebookService_Open_LoadsNoteService(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "test-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	assert.NotNil(t, notebook.Notes)
}

// Create tests

func TestNotebookService_Create_CreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "new-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	notebook, err := svc.Create("new-notebook", notebookDir, false)
	require.NoError(t, err)

	// Check notebook dir exists
	_, err = os.Stat(notebookDir)
	assert.NoError(t, err)

	// Check notes dir exists
	notesDir := filepath.Join(notebookDir, ".notes")
	_, err = os.Stat(notesDir)
	assert.NoError(t, err)

	assert.Equal(t, "new-notebook", notebook.Config.Name)
}

func TestNotebookService_Create_WritesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "new-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	_, err := svc.Create("new-notebook", notebookDir, false)
	require.NoError(t, err)

	// Check config file exists
	configPath := filepath.Join(notebookDir, NotebookConfigFile)
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Verify config content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var stored StoredNotebookConfig
	require.NoError(t, json.Unmarshal(data, &stored))

	assert.Equal(t, "new-notebook", stored.Name)
	assert.Equal(t, ".notes", stored.Root) // Should be relative
}

func TestNotebookService_Create_RegistersGlobally(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "new-notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	_, err := svc.Create("new-notebook", notebookDir, true)
	require.NoError(t, err)

	// Verify notebook was registered
	assert.Contains(t, configSvc.Store.Notebooks, notebookDir)
}

func TestNotebookService_Create_WithoutRegister(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "new-notebook")

	initialNotebooks := []string{"/existing/notebook"}
	configSvc := createTestConfigService(t, tmpDir, initialNotebooks)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	_, err := svc.Create("new-notebook", notebookDir, false)
	require.NoError(t, err)

	// Verify notebook was NOT registered
	assert.NotContains(t, configSvc.Store.Notebooks, notebookDir)
	assert.Equal(t, initialNotebooks, configSvc.Store.Notebooks)
}

// Infer tests

func TestNotebookService_Infer_DeclaredPathPriority(t *testing.T) {
	tmpDir := t.TempDir()
	declaredNotebook := createTestNotebook(t, tmpDir, "declared")
	ancestorNotebook := createTestNotebook(t, tmpDir, "ancestor")

	// Config points to declared notebook
	configPath := filepath.Join(tmpDir, "opennotes", "config.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))
	config := Config{
		Notebooks:    []string{},
		NotebookPath: declaredNotebook,
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	require.NoError(t, os.WriteFile(configPath, data, 0644))

	configSvc, err := NewConfigServiceWithPath(configPath)
	require.NoError(t, err)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	// Infer from ancestor notebook directory should still return declared
	notebook, err := svc.Infer(ancestorNotebook)
	require.NoError(t, err)
	require.NotNil(t, notebook)

	assert.Equal(t, "declared", notebook.Config.Name)
}

func TestNotebookService_Infer_ContextMatchPriority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a notebook with context matching workDir
	workDir := filepath.Join(tmpDir, "work", "project")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	notebookDir := filepath.Join(tmpDir, "notebooks", "work-notebook")
	notesDir := filepath.Join(notebookDir, ".notes")
	require.NoError(t, os.MkdirAll(notesDir, 0755))

	config := StoredNotebookConfig{
		Name:     "work-notebook",
		Root:     ".notes",
		Contexts: []string{filepath.Join(tmpDir, "work")}, // Parent of workDir
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	configPath := filepath.Join(notebookDir, NotebookConfigFile)
	require.NoError(t, os.WriteFile(configPath, data, 0644))

	// Register the notebook
	configSvc := createTestConfigService(t, tmpDir, []string{notebookDir})
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	// Infer should find via context match
	notebook, err := svc.Infer(workDir)
	require.NoError(t, err)
	require.NotNil(t, notebook)

	assert.Equal(t, "work-notebook", notebook.Config.Name)
}

func TestNotebookService_Infer_AncestorSearchPriority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notebook in ancestor
	ancestorNotebook := createTestNotebook(t, tmpDir, "project")

	// Work from a subdirectory
	subDir := filepath.Join(ancestorNotebook, "src", "deep")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	// Infer from subdirectory should find ancestor notebook
	notebook, err := svc.Infer(subDir)
	require.NoError(t, err)
	require.NotNil(t, notebook)

	assert.Equal(t, "project", notebook.Config.Name)
}

func TestNotebookService_Infer_NoneFound(t *testing.T) {
	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	notebook, err := svc.Infer(workDir)
	require.NoError(t, err)
	assert.Nil(t, notebook)
}

// List tests

func TestNotebookService_List_FromRegistered(t *testing.T) {
	tmpDir := t.TempDir()

	nb1 := createTestNotebook(t, tmpDir, "notebook1")
	nb2 := createTestNotebook(t, tmpDir, "notebook2")

	configSvc := createTestConfigService(t, tmpDir, []string{nb1, nb2})
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	workDir := filepath.Join(tmpDir, "work")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	notebooks, err := svc.List(workDir)
	require.NoError(t, err)

	assert.Len(t, notebooks, 2)
}

func TestNotebookService_List_FromAncestors(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notebook in ancestor directory
	ancestorNb := createTestNotebook(t, tmpDir, "ancestor-notebook")

	// Work from subdirectory
	subDir := filepath.Join(ancestorNb, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	notebooks, err := svc.List(subDir)
	require.NoError(t, err)

	assert.Len(t, notebooks, 1)
	assert.Equal(t, "ancestor-notebook", notebooks[0].Config.Name)
}

func TestNotebookService_List_Deduplicated(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notebook
	nbDir := createTestNotebook(t, tmpDir, "notebook")

	// Register and also be an ancestor
	subDir := filepath.Join(nbDir, "src")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, []string{nbDir})
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	// List from subdir - should find via registered AND ancestor, but dedup
	notebooks, err := svc.List(subDir)
	require.NoError(t, err)

	assert.Len(t, notebooks, 1)
}

func TestNotebookService_List_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, "work")
	require.NoError(t, os.MkdirAll(workDir, 0755))

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	notebooks, err := svc.List(workDir)
	require.NoError(t, err)

	assert.Empty(t, notebooks)
}

// Notebook method tests

func TestNotebook_MatchContext_Match(t *testing.T) {
	notebook := &Notebook{
		Config: NotebookConfig{
			StoredNotebookConfig: StoredNotebookConfig{
				Contexts: []string{"/home/user/projects", "/home/user/work"},
			},
		},
	}

	result := notebook.MatchContext("/home/user/projects/myapp/src")
	assert.Equal(t, "/home/user/projects", result)
}

func TestNotebook_MatchContext_NoMatch(t *testing.T) {
	notebook := &Notebook{
		Config: NotebookConfig{
			StoredNotebookConfig: StoredNotebookConfig{
				Contexts: []string{"/home/user/projects"},
			},
		},
	}

	result := notebook.MatchContext("/home/user/documents")
	assert.Equal(t, "", result)
}

func TestNotebook_AddContext_NewContext(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	newContext := "/new/context/path"
	err = notebook.AddContext(newContext, configSvc)
	require.NoError(t, err)

	assert.Contains(t, notebook.Config.Contexts, newContext)
}

func TestNotebook_AddContext_DuplicateIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	// Add same context twice
	existingContext := notebook.Config.Contexts[0]
	originalLen := len(notebook.Config.Contexts)

	err = notebook.AddContext(existingContext, configSvc)
	require.NoError(t, err)

	// Should not have been added again
	assert.Equal(t, originalLen, len(notebook.Config.Contexts))
}

func TestNotebook_SaveConfig_LocalOnly(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	// Modify and save without registering
	notebook.Config.Name = "renamed-notebook"
	err = notebook.SaveConfig(false, configSvc)
	require.NoError(t, err)

	// Verify local config was updated
	data, err := os.ReadFile(notebook.Config.Path)
	require.NoError(t, err)

	var stored StoredNotebookConfig
	require.NoError(t, json.Unmarshal(data, &stored))
	assert.Equal(t, "renamed-notebook", stored.Name)

	// Verify not registered globally
	assert.NotContains(t, configSvc.Store.Notebooks, notebookDir)
}

func TestNotebook_SaveConfig_WithRegistration(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "notebook")

	configSvc := createTestConfigService(t, tmpDir, nil)
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	err = notebook.SaveConfig(true, configSvc)
	require.NoError(t, err)

	// Verify was registered globally
	assert.Contains(t, configSvc.Store.Notebooks, notebookDir)
}

func TestNotebook_SaveConfig_AvoidsDuplicateRegistration(t *testing.T) {
	tmpDir := t.TempDir()
	notebookDir := createTestNotebook(t, tmpDir, "notebook")

	// Already registered
	configSvc := createTestConfigService(t, tmpDir, []string{notebookDir})
	dbSvc := NewDbService()
	t.Cleanup(func() {
		if err := dbSvc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})
	svc := NewNotebookService(configSvc, dbSvc)

	notebook, err := svc.Open(notebookDir)
	require.NoError(t, err)

	// Save with register flag
	err = notebook.SaveConfig(true, configSvc)
	require.NoError(t, err)

	// Should still only have one entry
	count := 0
	for _, p := range configSvc.Store.Notebooks {
		if p == notebookDir {
			count++
		}
	}
	assert.Equal(t, 1, count)
}
