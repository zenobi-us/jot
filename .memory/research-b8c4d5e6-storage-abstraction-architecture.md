---
id: b8c4d5e6
title: Storage Abstraction - Revised Architecture with Per-Notebook Backends
created_at: 2026-01-23T21:45:00+10:30
updated_at: 2026-01-23T21:45:00+10:30
status: planning
epic_id: a9b3f2c1
---

# Storage Abstraction - Revised Architecture

## Design Decisions

1. ✅ **Per-notebook storage** - Each notebook can have its own backend
2. ✅ **Config in .opennotes.json** - Backend specified per-notebook, absent = OsFs
3. ✅ **Separate ConfigStorage vs NotebookStorage** - Different interfaces for different concerns
4. ✅ **OsFs + MemMapFs now** - Future backends added without breaking changes

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                           OpenNotes                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────┐         ┌─────────────────────────────┐   │
│  │   ConfigService     │         │     NotebookService         │   │
│  │                     │         │                             │   │
│  │  storage: Storage   │         │  configStorage: Storage     │   │
│  │  (for config.json)  │         │  (to find notebooks)        │   │
│  └─────────────────────┘         │                             │   │
│                                  │  notebooks: []Notebook      │   │
│                                  │    └─ each has own Storage  │   │
│                                  └─────────────────────────────┘   │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                        Storage Interface                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │    OsFs      │  │  MemMapFs    │  │   Future:    │              │
│  │  (default)   │  │  (testing)   │  │  S3, Git...  │              │
│  └──────────────┘  └──────────────┘  └──────────────┘              │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Core Interfaces

### Storage Interface

```go
// internal/storage/storage.go

package storage

import "github.com/spf13/afero"

// Storage is the core filesystem abstraction
// Embeds afero.Fs for full filesystem operations
type Storage interface {
    afero.Fs
    
    // Type returns the storage backend type (e.g., "os", "memory", "s3")
    Type() string
    
    // Config returns backend-specific configuration (for debugging/display)
    Config() map[string]string
}
```

### Built-in Implementations

```go
// internal/storage/os.go

package storage

import "github.com/spf13/afero"

// OsStorage wraps afero.OsFs for real filesystem access
type OsStorage struct {
    afero.Fs
}

func NewOsStorage() *OsStorage {
    return &OsStorage{
        Fs: afero.NewOsFs(),
    }
}

func (s *OsStorage) Type() string {
    return "os"
}

func (s *OsStorage) Config() map[string]string {
    return map[string]string{"type": "os"}
}
```

```go
// internal/storage/memory.go

package storage

import "github.com/spf13/afero"

// MemoryStorage wraps afero.MemMapFs for in-memory filesystem (testing)
type MemoryStorage struct {
    afero.Fs
}

func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        Fs: afero.NewMemMapFs(),
    }
}

func (s *MemoryStorage) Type() string {
    return "memory"
}

func (s *MemoryStorage) Config() map[string]string {
    return map[string]string{"type": "memory"}
}
```

### Storage Factory

```go
// internal/storage/factory.go

package storage

import "fmt"

// StorageConfig represents storage configuration from .opennotes.json
type StorageConfig struct {
    Type   string            `json:"type"`   // "os", "memory", "s3", etc.
    Config map[string]string `json:"config"` // Backend-specific config
}

// NewStorage creates a Storage instance from configuration
// If config is nil or empty, returns OsStorage (default)
func NewStorage(cfg *StorageConfig) (Storage, error) {
    if cfg == nil || cfg.Type == "" {
        return NewOsStorage(), nil  // Default: OsFs
    }
    
    switch cfg.Type {
    case "os":
        return NewOsStorage(), nil
    case "memory":
        return NewMemoryStorage(), nil
    // Future:
    // case "s3":
    //     return NewS3Storage(cfg.Config)
    // case "git":
    //     return NewGitStorage(cfg.Config)
    default:
        return nil, fmt.Errorf("unknown storage type: %s", cfg.Type)
    }
}

// Default returns the default storage (OsFs)
func Default() Storage {
    return NewOsStorage()
}
```

---

## Notebook Configuration

### .opennotes.json Schema (Extended)

```json
{
  "name": "My Notes",
  "description": "Personal notebook",
  
  "storage": {
    "type": "os"
  },
  
  "views": {
    "today": { ... }
  }
}
```

### Storage Configuration Examples

**Default (OsFs) - field absent or empty**:
```json
{
  "name": "Local Notes"
}
```

**Explicit OsFs**:
```json
{
  "name": "Local Notes",
  "storage": {
    "type": "os"
  }
}
```

**Future: S3 Backend**:
```json
{
  "name": "Cloud Notes",
  "storage": {
    "type": "s3",
    "config": {
      "bucket": "my-notes-bucket",
      "region": "us-east-1",
      "prefix": "notebooks/work/"
    }
  }
}
```

**Future: Git Backend**:
```json
{
  "name": "Versioned Notes",
  "storage": {
    "type": "git",
    "config": {
      "remote": "git@github.com:user/notes.git",
      "branch": "main"
    }
  }
}
```

---

## Service Changes

### ConfigService

```go
// internal/services/config.go

type ConfigService struct {
    storage storage.Storage  // For ~/.config/opennotes/
    // ... existing fields
}

func NewConfigService(storage storage.Storage) *ConfigService {
    if storage == nil {
        storage = storage.Default()  // OsFs
    }
    return &ConfigService{
        storage: storage,
        // ...
    }
}

// All file operations use:
// - afero.ReadFile(cs.storage, path)
// - cs.storage.Open(path)
// - etc.
```

### NotebookService

```go
// internal/services/notebook.go

type NotebookService struct {
    configStorage storage.Storage  // For finding notebooks
    // ... existing fields
}

func NewNotebookService(configStorage storage.Storage, cfg *ConfigService) *NotebookService {
    if configStorage == nil {
        configStorage = storage.Default()
    }
    return &NotebookService{
        configStorage: configStorage,
        // ...
    }
}
```

### Notebook (with per-notebook storage)

```go
// internal/core/notebook.go

type Notebook struct {
    Name        string
    Path        string
    Config      *NotebookConfig
    Storage     storage.Storage  // This notebook's storage backend
}

// When loading a notebook:
func (ns *NotebookService) LoadNotebook(path string) (*Notebook, error) {
    // 1. Read .opennotes.json
    configData, err := afero.ReadFile(ns.configStorage, filepath.Join(path, ".opennotes.json"))
    
    // 2. Parse storage config
    var nbConfig NotebookConfig
    json.Unmarshal(configData, &nbConfig)
    
    // 3. Create storage backend for this notebook
    nbStorage, err := storage.NewStorage(nbConfig.Storage)
    if err != nil {
        return nil, err
    }
    
    // 4. Return notebook with its own storage
    return &Notebook{
        Name:    nbConfig.Name,
        Path:    path,
        Config:  &nbConfig,
        Storage: nbStorage,
    }, nil
}
```

### NoteService (uses notebook's storage)

```go
// internal/services/note.go

func (ns *NoteService) ListNotes(notebook *Notebook) ([]Note, error) {
    // Use the notebook's storage backend
    files, err := afero.Glob(notebook.Storage, filepath.Join(notebook.Path, "**/*.md"))
    // ...
}
```

---

## Data Flow

### Production Flow

```
1. App starts
2. ConfigService created with OsStorage (default)
3. ConfigService reads ~/.config/opennotes/config.json
4. NotebookService discovers registered notebooks
5. For each notebook:
   a. Read .opennotes.json
   b. Check storage.type (default: "os")
   c. Create Storage instance for that notebook
   d. Notebook now has its own Storage backend
6. Operations on notebook use its Storage
```

### Test Flow

```
1. Test starts
2. Create MemoryStorage for config
3. Create MemoryStorage for notebooks (or same instance)
4. Populate with test data
5. Create services with MemoryStorage
6. Run test - no real filesystem touched
7. Test ends - MemoryStorage garbage collected
```

---

## Directory Structure

```
internal/
├── storage/                    # NEW: Storage abstraction
│   ├── storage.go              # Storage interface
│   ├── factory.go              # NewStorage factory
│   ├── os.go                   # OsStorage implementation
│   ├── memory.go               # MemoryStorage implementation
│   └── storage_test.go         # Unit tests
│
├── core/
│   ├── notebook.go             # Add Storage field
│   └── ...
│
├── services/
│   ├── config.go               # Accept Storage parameter
│   ├── notebook.go             # Accept Storage, per-notebook Storage
│   ├── note.go                 # Use notebook's Storage
│   └── ...
│
└── testing/                    # NEW: Test helpers
    ├── helpers.go              # SetupTestStorage, CreateTestNotebook
    └── fixtures.go             # Test data generators
```

---

## Revised Phase Plan

### Phase 1: Storage Package (1.5 hours)
- [ ] Add afero to go.mod
- [ ] Create `internal/storage/` package
- [ ] Implement Storage interface
- [ ] Implement OsStorage
- [ ] Implement MemoryStorage
- [ ] Implement factory with defaults
- [ ] Unit tests for storage package

### Phase 2: Core Integration (2 hours)
- [ ] Update Notebook struct with Storage field
- [ ] Update NotebookConfig with storage schema
- [ ] Update ConfigService to accept Storage
- [ ] Update NotebookService to accept Storage
- [ ] Per-notebook Storage loading from .opennotes.json
- [ ] Update commands to use storage.Default()

### Phase 3: Service Migration (1.5 hours)
- [ ] NoteService: Use notebook.Storage
- [ ] ViewService: Use notebook.Storage  
- [ ] DbService: Consider storage implications
- [ ] Update all afero.ReadFile, fs.Open calls

### Phase 4: Test Migration (1.5 hours)
- [ ] Create test helpers in `internal/testing/`
- [ ] Convert config tests to MemoryStorage
- [ ] Convert notebook tests to MemoryStorage
- [ ] Convert view tests to MemoryStorage
- [ ] Verify no ~/.config pollution

### Phase 5: Validation (0.5 hours)
- [ ] Full test suite passes
- [ ] Verify no real FS access during tests
- [ ] Documentation update
- [ ] Performance check

**Total: ~7 hours** (increased from 4-5 due to per-notebook storage)

---

## Test Helper Examples

```go
// internal/testing/helpers.go

package testing

import (
    "testing"
    "github.com/zenobi-us/opennotes/internal/storage"
)

// SetupTestStorage creates isolated MemoryStorage for tests
func SetupTestStorage(t *testing.T) storage.Storage {
    t.Helper()
    return storage.NewMemoryStorage()
}

// CreateTestConfig writes config.json to storage
func CreateTestConfig(t *testing.T, s storage.Storage, notebooks []NotebookRef) {
    t.Helper()
    // Create directory structure
    afero.MkdirAll(s, "/.config/opennotes", 0755)
    // Write config
    data, _ := json.Marshal(map[string]interface{}{
        "notebooks": notebooks,
    })
    afero.WriteFile(s, "/.config/opennotes/config.json", data, 0644)
}

// CreateTestNotebook creates a notebook in storage
func CreateTestNotebook(t *testing.T, s storage.Storage, path string, cfg *NotebookConfig) {
    t.Helper()
    afero.MkdirAll(s, path, 0755)
    if cfg == nil {
        cfg = &NotebookConfig{Name: filepath.Base(path)}
    }
    data, _ := json.Marshal(cfg)
    afero.WriteFile(s, filepath.Join(path, ".opennotes.json"), data, 0644)
}

// CreateTestNote creates a markdown file in notebook
func CreateTestNote(t *testing.T, s storage.Storage, notebookPath, name, content string) {
    t.Helper()
    afero.WriteFile(s, filepath.Join(notebookPath, name), []byte(content), 0644)
}
```

---

## Example Test

```go
func TestNotebookService_LoadNotebook_WithDefaultStorage(t *testing.T) {
    // Setup
    s := testing.SetupTestStorage(t)
    testing.CreateTestConfig(t, s, []NotebookRef{
        {Name: "Notes", Path: "/home/user/Notes"},
    })
    testing.CreateTestNotebook(t, s, "/home/user/Notes", nil)  // No storage = OsFs default
    testing.CreateTestNote(t, s, "/home/user/Notes", "note1.md", "# Hello")
    
    // Act
    cfg := services.NewConfigService(s)
    ns := services.NewNotebookService(s, cfg)
    notebook, err := ns.LoadNotebook("/home/user/Notes")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Notes", notebook.Name)
    assert.Equal(t, "memory", notebook.Storage.Type())  // Inherited from test
}

func TestNotebookService_LoadNotebook_WithExplicitStorage(t *testing.T) {
    // Setup
    s := testing.SetupTestStorage(t)
    testing.CreateTestNotebook(t, s, "/home/user/Notes", &NotebookConfig{
        Name: "Notes",
        Storage: &storage.StorageConfig{
            Type: "os",  // Explicit OsFs
        },
    })
    
    // Act
    cfg := services.NewConfigService(s)
    ns := services.NewNotebookService(s, cfg)
    notebook, err := ns.LoadNotebook("/home/user/Notes")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "os", notebook.Storage.Type())
}
```

---

## Future Backend Stubs

For documentation purposes, here's what future backends might look like:

```go
// internal/storage/s3.go (FUTURE)

type S3Storage struct {
    bucket string
    region string
    prefix string
    client *s3.Client
}

func NewS3Storage(cfg map[string]string) (*S3Storage, error) {
    // Initialize S3 client
    // Return storage that implements afero.Fs via s3fs adapter
}

func (s *S3Storage) Type() string { return "s3" }
```

```go
// internal/storage/git.go (FUTURE)

type GitStorage struct {
    remote string
    branch string
    repo   *git.Repository
}

func NewGitStorage(cfg map[string]string) (*GitStorage, error) {
    // Clone/open git repository
    // Return storage backed by git worktree
}

func (s *GitStorage) Type() string { return "git" }
```

---

## Summary

| Component | Storage Source | Default |
|-----------|---------------|---------|
| ConfigService | Constructor parameter | OsFs |
| NotebookService | Constructor parameter | OsFs |
| Notebook | .opennotes.json `storage` field | OsFs (if absent) |
| Tests | MemoryStorage | N/A |

**Key Principle**: If storage config is absent or nil → default to OsFs

---

## Next Steps

1. ✅ Review this architecture
2. ✅ Approve to proceed
3. 🔜 Create Phase 1 task (storage package)
4. 🔜 Execute phases sequentially

**Status**: Ready for human review
