---
id: 5d2e1f4c
title: Storage Abstraction Layer - Rough Implementation Plan
created_at: 2026-01-23T21:25:00+10:30
updated_at: 2026-01-23T21:25:00+10:30
status: planning
epic_id: a9b3f2c1
assigned_to: human-review-needed
---

# Storage Abstraction Layer - Implementation Plan

## Quick Overview

Transform OpenNotes from using real filesystem (`os.Open`, `ioutil.ReadFile`) to an abstracted filesystem using `github.com/spf13/afero`.

**Result**: 
- 🎯 Tests no longer pollute `~/.config/opennotes/config.json`
- 🎯 Fresh in-memory filesystem for each test
- 🎯 Future-proof for cloud storage backends
- 🎯 Faster test execution (no disk I/O)

---

## The High-Level Plan

```
┌──────────────────────────────────┐
│   Storage Abstraction Layer      │
│  (Replace direct os.* calls)     │
└───────────┬──────────────────────┘
            │
    ┌───────┴────────┐
    │                │
   OsFs          MemMapFs
(Production)    (Testing)
    │                │
    └─────────────────┘
         Selectable per context
```

### Core Idea

1. **Services accept `afero.Fs` parameter** instead of using `os.Open` directly
2. **Production**: Pass `afero.NewOsFs()` → uses real filesystem
3. **Tests**: Pass `afero.NewMemMapFs()` → uses in-memory filesystem

---

## Four Phases

### Phase 1: Infrastructure (1.5 hours)
- Add `afero` to `go.mod`
- Create test helper functions in `internal/testing/helpers.go`
- Document patterns for future test writers

### Phase 2: Core Services (1.5 hours)
- Update `ConfigService` constructor to accept `afero.Fs`
- Update `NotebookService` constructor to accept `afero.Fs`
- Update commands to pass `afero.NewOsFs()` when creating services
- Replace all `os.Open()`, `ioutil.ReadFile()` with `fs.Open()`, `afero.ReadFile()`

### Phase 3: Test Migration (1-2 hours)
- Convert existing tests incrementally
- Use test helpers to create MemMapFs with sample data
- Verify no test touches real `~/.config/opennotes/`

### Phase 4: Validation (0.5 hours)
- Run full test suite
- Check test speed improvements
- Verify coverage maintained
- Documentation cleanup

---

## File Changes Summary

### New Files
```
internal/testing/helpers.go
  - SetupTestFS()
  - CreateTestConfig()
  - CreateTestNotebook()
  - CreateTestNotes()
```

### Modified Files
```
internal/services/config.go
  - Add: fs afero.Fs field
  - Add: fs parameter to constructor
  - Replace: os.* calls with afero.*

internal/services/notebook.go
  - Same pattern as ConfigService

internal/services/db.go
  - Minor: Path resolution may change

cmd/root.go (and other commands)
  - Pass: afero.NewOsFs() to service constructors
```

### Test Files (*_test.go)
```
All test files:
  - Setup: testFs := afero.NewMemMapFs()
  - Use: helpers to populate test data
  - Pass: testFs to service constructors
```

---

## Before/After Code Example

### Before (Current)
```go
// ConfigService
type ConfigService struct {
    configPath string
}

func NewConfigService() *ConfigService {
    configPath := filepath.Join(os.Getenv("HOME"), ".config/opennotes/config.json")
    // Hard-coded filesystem access!
}

// In test
func TestLoadConfig(t *testing.T) {
    // Directly modifies ~/.config/opennotes/config.json ⚠️
    cfg := NewConfigService()
    // ...
}
```

### After (With Afero)
```go
// ConfigService
type ConfigService struct {
    fs afero.Fs
    configPath string
}

func NewConfigService(fs afero.Fs) *ConfigService {
    return &ConfigService{
        fs: fs,
        configPath: "/.config/opennotes/config.json",  // Relative to fs root
    }
}

// File reads use: afero.ReadFile(cs.fs, cs.configPath)

// Production
cfg := services.NewConfigService(afero.NewOsFs())

// In test - Clean! No side effects!
func TestLoadConfig(t *testing.T) {
    testFs := afero.NewMemMapFs()
    afero.WriteFile(testFs, "/.config/opennotes/config.json", testData, 0644)
    cfg := services.NewConfigService(testFs)
    // ... no pollution of ~/config!
}
```

---

## Test Pattern Examples

### Example 1: Simple Config Test
```go
func TestConfigService_LoadConfig(t *testing.T) {
    // Setup
    testFs := afero.NewMemMapFs()
    testConfig := `{"notebooks": [{"name": "Notes", "path": "/home/user/Notes"}]}`
    afero.WriteFile(testFs, "/.config/opennotes/config.json", []byte(testConfig), 0644)

    // Act
    cfg := services.NewConfigService(testFs)
    notebooks, err := cfg.GetNotebooks()

    // Assert
    assert.NoError(t, err)
    assert.Len(t, notebooks, 1)
    assert.Equal(t, "Notes", notebooks[0].Name)
    // testFs is garbage collected, no cleanup needed
}
```

### Example 2: Notebook Discovery Test
```go
func TestNotebookService_DiscoverNotebooks(t *testing.T) {
    // Setup
    testFs := afero.NewMemMapFs()
    helpers.CreateTestNotebook(t, testFs, "/home/user/Notes")
    helpers.CreateTestNotes(t, testFs, "/home/user/Notes", map[string]string{
        "note1.md": "# Note 1",
        "note2.md": "# Note 2",
    })

    // Act
    ns := services.NewNotebookService(testFs, configService)
    notes, err := ns.ListNotes()

    // Assert
    assert.NoError(t, err)
    assert.Len(t, notes, 2)
}
```

---

## Pluggable Backends (Future-Proof)

Current:
```
OsFs (production)
MemMapFs (testing)
```

Future (without code changes):
```
OsFs (production)
MemMapFs (testing)
UnionFs (layered filesystems)
BasePathFs (chroot-like)
S3Backend (cloud storage) ← would need new implementation
CloudStorageBackend ← would need new implementation
```

The abstraction makes it trivial to add these later.

---

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| afero API changes | Widely-used, stable; spf13 maintains | 
| Missed file operations | Grep for `os.Open`, `ioutil.` to find all calls |
| Test instability | Each test gets fresh MemMapFs |
| Performance regression | MemMapFs is actually faster than disk I/O |
| Breaking changes | Only constructor signatures change, backward compatible |

---

## Success Criteria (Verification Steps)

After each phase:

**Phase 1**: ✅ afero added to go.mod, helpers compile
**Phase 2**: ✅ Services accept fs parameter, commands compile
**Phase 3**: ✅ Tests run without touching ~/.config (verify: `lsof` or `strace`)
**Phase 4**: ✅ All tests pass, coverage ≥85%, execution time ≤5s

Final verification:
```bash
# Before test run
stat ~/.config/opennotes/config.json  # Record mtime

# Run tests
mise run test

# After test run
stat ~/.config/opennotes/config.json  # Verify mtime unchanged ✅
```

---

## Implementation Notes

### Path Handling
- **Production**: Use system paths with os.Expand for $HOME
- **Testing**: Use absolute paths like `/.config/opennotes`
- **Pattern**: Paths relative to filesystem root, not user home

### Concurrency
- afero.MemMapFs is thread-safe ✅
- Safe to run tests in parallel ✅
- Each test needs own MemMapFs instance ✅

### Performance
- Memory usage: Negligible (test data is small)
- Test speed: Expected to be faster (no disk I/O)
- Benchmark after Phase 4 to measure gains

---

## Next Steps

**For Approval**:
1. Review epic: `.memory/epic-a9b3f2c1-storage-abstraction-layer.md`
2. Review research: `.memory/research-7f4c2e1a-afero-vfs-integration.md`
3. Approve to proceed with Phase 1

**Upon Approval**:
1. Create Phase 1 task: Add afero dependency + test helpers
2. Execute Phase 1
3. Create Phase 2 task: Update services
4. Execute Phases 2-4 sequentially

---

## Questions for Human Review

1. **Architecture**: Does direct `afero.Fs` injection look good, or prefer StorageProvider wrapper?
2. **Scope**: Should we only do OsFs + MemMapFs, or add UnionFs support in Phase 1?
3. **Timeline**: Can this be a side project, or should it block Views System Feature 3?
4. **Testing**: Should existing tests be converted incrementally or all at once?

---

## Summary

**Problem**: Tests pollute `~/.config/opennotes/config.json`

**Solution**: Use afero for filesystem abstraction with pluggable backends

**Implementation**: 4 phases, ~4-5 hours total

**Benefit**: Clean tests, future-proof architecture, no side effects

**Risk Level**: Low (proven library, isolated changes)

**Status**: Ready for human review and approval

---

**Created**: 2026-01-23 21:25 GMT+10:30  
**Type**: Implementation Overview  
**Related Files**:
- Epic: `.memory/epic-a9b3f2c1-storage-abstraction-layer.md`
- Research: `.memory/research-7f4c2e1a-afero-vfs-integration.md`
