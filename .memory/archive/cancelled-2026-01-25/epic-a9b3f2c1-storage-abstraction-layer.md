---
id: a9b3f2c1
title: Storage Abstraction Layer with Pluggable Backends
created_at: 2026-01-23T21:20:00+10:30
updated_at: 2026-01-23T21:50:00+10:30
status: approved
spec: spec-b8c4d5e6-storage-abstraction-architecture.md
epic_id: null
tags:
  - infrastructure
  - testing
  - architecture
  - vfs
related_research: research-7f4c2e1a-afero-vfs-integration.md
---

# Epic: Storage Abstraction Layer with Pluggable Backends

## Vision/Goal

Eliminate test pollution by introducing a storage abstraction layer using `afero` that enables:
1. **Production**: Real filesystem via OsFs
2. **Testing**: In-memory filesystem via MemMapFs
3. **Future**: Pluggable backends (cloud storage, S3, etc.)

**Primary Problem Solved**: Tests modifying `~/.config/opennotes/config.json` during test execution.

---

## Success Criteria

- ✅ No test modifies `~/.config/opennotes/config.json`
- ✅ All existing tests pass with zero functionality change
- ✅ Test setup is simpler and faster
- ✅ New tests use MemMapFs by default
- ✅ Architecture allows future backends without major refactoring
- ✅ Test coverage maintained at 85%+
- ✅ Production code unchanged (only constructor signatures)
- ✅ Documentation updated for test writers

---

## Phases

1. **Phase 1: Infrastructure Setup** (~1.5 hours)
   - Add afero dependency
   - Create test helpers and utilities
   - Document patterns for test writers

2. **Phase 2: Core Services Migration** (~1.5 hours)
   - ConfigService: Accept and use afero.Fs
   - NotebookService: Accept and use afero.Fs
   - Update constructors in commands

3. **Phase 3: Test Migration** (~1-2 hours)
   - Convert config-related tests
   - Convert notebook-related tests
   - Convert view system tests
   - Convert integration tests

4. **Phase 4: Validation & Cleanup** (~0.5 hours)
   - Verify no ~/.config access during tests
   - Performance benchmarks
   - Documentation review

---

## Overall Timeline

**Estimated**: 4-5 hours total across 4 phases  
**Team Capacity**: Can be completed in 1-2 development sessions  
**Risk Level**: Low (isolated to services and tests)  
**Blocking**: No (new infrastructure, non-blocking)

---

## Dependencies

- ✅ **afero library**: `github.com/spf13/afero` (public, no conflicts)
- ✅ **Go 1.24.7**: Already required
- ✅ **No breaking changes**: Full backward compatibility
- ✅ **No database changes**: DuckDB unaffected

---

## Key Decisions

### Decision 1: Use afero (Not blang/vfs)
- ✅ Actively maintained by spf13 (Cobra/Viper creator)
- ✅ Production-tested in many projects
- ✅ Better integration with Go ecosystem
- ✅ Larger community and documentation

### Decision 2: Direct afero.Fs Integration (Not StorageProvider wrapper)
- Start simple with direct `afero.Fs` parameter injection
- Evolve to `StorageProvider` interface if cloud storage needed
- Avoids over-engineering at this stage
- Pattern: Constructor parameter pattern (proven, minimal)

### Decision 3: OsFs + MemMapFs Initially (Add Others Later)
- Phase 1: Only OsFs and MemMapFs
- Future: Add UnionFs, BasePathFs, etc. as needed
- Reduces initial complexity
- Proven backends sufficient for current needs

---

## Architecture

### Service Layer Changes

All services follow injection pattern:

```go
type ConfigService struct {
    fs afero.Fs
    configPath string
    // existing fields...
}

func NewConfigService(fs afero.Fs) *ConfigService {
    return &ConfigService{
        fs: fs,
        configPath: "/.config/opennotes/config.json",  // Relative to fs root
        // ...
    }
}

// All file operations use: fs.Open, afero.ReadFile, etc.
```

### Commands Pattern

```go
// In cmd/root.go or individual commands
cfg := services.NewConfigService(afero.NewOsFs())
nb := services.NewNotebookService(afero.NewOsFs(), cfg)
```

### Test Pattern

```go
// In tests
testFs := afero.NewMemMapFs()

// Setup test data
afero.WriteFile(testFs, "/.config/opennotes/config.json", data, 0644)
afero.Mkdir(testFs, "/home/user/Notes/OpenNotes", 0755)
afero.WriteFile(testFs, "/home/user/Notes/OpenNotes/note.md", content, 0644)

// Create service with test filesystem
cfg := services.NewConfigService(testFs)

// Run test - no side effects on real filesystem
result, err := cfg.GetNotebooks()
```

---

## Test Helpers (New File)

**Location**: `internal/testing/helpers.go`

```go
package testing

import (
    "github.com/spf13/afero"
)

// SetupTestFS creates a fresh MemMapFs with standard directories
func SetupTestFS(t TestingT) afero.Fs {
    fs := afero.NewMemMapFs()
    // Create standard directories
    afero.Mkdir(fs, "/.config", 0755)
    afero.Mkdir(fs, "/.config/opennotes", 0755)
    return fs
}

// CreateTestConfig writes config to MemMapFs
func CreateTestConfig(t TestingT, fs afero.Fs, config map[string]interface{}) {
    // Marshal config to JSON
    // Write to fs using afero.WriteFile
}

// CreateTestNotebook creates notebook structure in MemMapFs
func CreateTestNotebook(t TestingT, fs afero.Fs, name string) string {
    // Create notebook directory structure
    // Return notebook path
}

// CreateTestNotes creates markdown files in notebook
func CreateTestNotes(t TestingT, fs afero.Fs, notebookPath string, notes map[string]string) {
    // Create markdown files in notebook
}
```

---

## Migration Path

### Services to Update

| Service | Status | Effort | Priority |
|---------|--------|--------|----------|
| ConfigService | Ready | 30 min | P1 |
| NotebookService | Ready | 30 min | P1 |
| DbService | Ready | 15 min | P2 |
| DisplayService | No change | 0 min | N/A |
| SearchService | No change* | 0 min | N/A |
| ViewService | Ready | 20 min | P2 |

\* May need fs for future features

### Test Coverage Strategy

**Immediate** (High Priority):
- [ ] Config loading tests
- [ ] Notebook discovery tests
- [ ] Notebook config tests
- [ ] View system tests

**Follow-up** (Medium Priority):
- [ ] Integration tests
- [ ] Command end-to-end tests
- [ ] Search tests

**Optional** (Low Priority):
- [ ] Display/formatting tests (don't touch FS)

---

## Risk Assessment

### Technical Risks
- **Low**: afero is production-tested, widely used
- **Mitigation**: Phase-by-phase rollout with verification

### Integration Risks
- **Low**: Constructor-only changes to services
- **Mitigation**: Full backward compatibility via default parameters (if needed)

### Performance Risks
- **None**: MemMapFs actually faster than disk I/O
- **Benefit**: Faster test execution

---

## Anti-Patterns to Avoid

❌ **Don't**: Create `StorageProvider` abstraction immediately
✅ **Do**: Use simple `afero.Fs` parameter injection

❌ **Don't**: Migrate all tests at once
✅ **Do**: Phase-by-phase conversion starting with config tests

❌ **Don't**: Add cloud storage support in Phase 1
✅ **Do**: Only OsFs + MemMapFs initially

❌ **Don't**: Share MemMapFs instances between tests
✅ **Do**: Create fresh MemMapFs for each test

---

## Success Metrics

After completion, verify:
- ✅ `ls -la ~/.config/opennotes/` shows no recent modifications during test run
- ✅ All tests pass: `mise run test`
- ✅ Test execution time ≤ 5 seconds (potentially faster)
- ✅ New tests use MemMapFs by default
- ✅ Test coverage ≥ 85% maintained
- ✅ No false positives or flaky tests

---

## Documentation Needed

1. **For Developers**:
   - How to write tests with MemMapFs
   - Test helper reference
   - Common patterns and examples

2. **In Code**:
   - Comments explaining afero.Fs parameter
   - Examples in package-level docs

3. **In Memory**:
   - Update `.memory/learning-*.md` with best practices
   - Document patterns discovered during implementation

---

## Remaining Questions

### Q1: Should services have optional fs parameter with default?
```go
// Option A: Required (current plan)
func NewConfigService(fs afero.Fs) *ConfigService

// Option B: Optional with default
func NewConfigService(fs ...afero.Fs) *ConfigService {
    if len(fs) == 0 {
        fs[0] = afero.NewOsFs()
    }
}
```
**Decision**: Use Option A (required) - explicit is better

### Q2: How to handle $HOME and $XDG_CONFIG_HOME?
- For tests: Use absolute paths like `/.config/opennotes`
- For production: Use `os.Expand` or `afero.GetOsFs` detection
- Pattern: Paths are relative to filesystem root, not user home

---

## Approval Checklist

- [ ] Human review: Architecture approved
- [ ] Human review: Risk assessment acceptable
- [ ] Human review: Ready to proceed with Phase 1

---

## Status Tracking

**Created**: 2026-01-23 21:20 GMT+10:30  
**Status**: 🔴 **PROPOSED** (awaiting human review)  
**Reviewer Assignment**: [NEEDS-HUMAN-REVIEW]

Next action: Await human approval to proceed with Phase 1 implementation.
