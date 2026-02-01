# OpenNotes TODO

## Current Status

**Active Epics**: 2
1. **Remove DuckDB** - Phases 1-3 Complete, Phase 4 Next Session
2. **Pi-OpenNotes Extension** - Phase 3 Complete, Ready for Phase 4

---

## 🔍 Remove DuckDB - Pure Go Search

### Epic f661c068 - Progress Summary

| Phase | Status | Description |
|-------|--------|-------------|
| 1. Research | ✅ Complete | Research synthesis, strategic decisions |
| 2. Interface Design | ✅ Complete | `internal/search/` package (8 files) |
| 3. Query Parser | ✅ Complete | Participle-based parser (5 files, 10 tests) |
| 4. Bleve Backend | 🔜 **NEXT SESSION** | Full-text indexing implementation |
| 5. DuckDB Removal | 🔜 | Remove all DuckDB code |
| 6. Semantic Search | 🔜 | Optional chromem-go integration |

### Session 2026-02-01 Completed

**Phase 2 - Interface Design** ✅:
- `internal/search/` package with 8 files
- Index, Query AST, FindOpts, Storage, Parser interfaces
- All files compile, lint passes

**Phase 3 - Query Parser** ✅:
- `internal/search/parser/` with 5 files
- Participle-based Gmail-style syntax
- 10 test cases, all passing

**Commits**:
- `5e1205d` - feat(search): add core interfaces for pure Go search implementation
- `d888253` - feat(search): implement Gmail-style query parser with Participle

---

## 🔜 Next Session: Phase 4 - Bleve Backend

### Tasks for Phase 4

- [ ] Add Bleve dependency: `go get github.com/blevesearch/bleve/v2`
- [ ] Create `internal/search/bleve/` package
- [ ] Implement `Index` interface with Bleve
- [ ] Define document mapping (field weights for BM25)
- [ ] Implement incremental indexing
- [ ] Add afero-based persistence
- [ ] Write comprehensive tests
- [ ] Benchmark performance

### Key Design Decisions

From research:
- Use Bleve's BM25 ranking with field weights (path=1000, title=500, body=1)
- Store index in `.opennotes/index/` directory
- Support incremental updates (checksum-based change detection)
- Use afero for testability (in-memory filesystem for tests)

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

- **Next Priority**: Phase 4 (Bleve Backend) in next session
- **Timeline**: ~2-3 weeks remaining for full DuckDB removal
- **Tests**: All tests passing, 10 new parser tests added
- **No Push**: Changes committed but not pushed (awaiting human review)
