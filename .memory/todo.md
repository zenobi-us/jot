# OpenNotes TODO

## Current Status

**Active Epics**: 2
1. **Remove DuckDB** - Phase 4 (Bleve Backend) In Progress
2. **Pi-OpenNotes Extension** - Phase 3 Complete, Ready for Phase 4

---

## 🔍 Remove DuckDB - Pure Go Search

### Epic f661c068 - Progress Summary

| Phase | Status | Description |
|-------|--------|-------------|
| 1. Research | ✅ Complete | Research synthesis, strategic decisions |
| 2. Interface Design | ✅ Complete | `internal/search/` package (8 files) |
| 3. Query Parser | ✅ Complete | Participle-based parser (5 files, 10 tests) |
| 4. Bleve Backend | ✅ Complete | Full-text indexing (9 files, 36 tests, 6 benchmarks) |
| 5. DuckDB Removal | 🔄 **IN PROGRESS** | Remove all DuckDB code - [phase-02df510c](phase-02df510c-duckdb-removal.md) |
| 6. Semantic Search | 🔜 | Optional chromem-go integration |

### Session 2026-02-01 Evening - 🔄 PHASE 5 IN PROGRESS

**Phase 5 - DuckDB Removal** 🔄 **IN PROGRESS**:

See detailed task checklist in [phase-02df510c-duckdb-removal.md](phase-02df510c-duckdb-removal.md)

### High-Level Tasks

- [ ] Codebase audit for DuckDB references
- [ ] Service layer migration (NoteService, DbService removal)
- [ ] CLI command migration (notes search, notes list)
- [ ] Dependency cleanup (remove from go.mod)
- [ ] Integration & testing (full test suite)
- [ ] Performance validation (binary size, startup time)
- [ ] Documentation updates (README, AGENTS.md)

---

## ✅ Previous Session - Phase 4 Complete

**Phase 4 - Bleve Backend** ✅ **COMPLETED** (21:35 2026-02-01):

All Tasks Complete:
- [x] Add Bleve dependency
- [x] Add afero dependency
- [x] Create 9 files in `internal/search/bleve/`
- [x] Write 36 tests (all passing)
- [x] Implement FindByQueryString
- [x] Fix tag matching bug

**Performance**: 0.754ms search (97% under target), all benchmarks passing

---

## 📦 Pi-OpenNotes Extension

**Epic**: [epic-1f41631e-pi-opennotes-extension.md](epic-1f41631e-pi-opennotes-extension.md)

| Phase | Status |
|-------|--------|
| Phase 1: Research & Design | ✅ Complete |
| Phase 2: Implementation | ✅ Complete (72 tests) |
| Phase 3: Testing & Documentation | ✅ Complete |
| Phase 4: Distribution | 🔜 Next |

---

## Notes

- **Current Work**: Phase 4 (Bleve Backend) implementation
- **Tests**: All passing (22 new bleve tests + existing tests)
- **Lint**: Clean, no issues
- **No Push**: Changes not pushed (awaiting human review)
