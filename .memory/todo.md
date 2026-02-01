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
| 4. Bleve Backend | ✅ **COMPLETE** | Full-text indexing (9 files, 36 tests, 6 benchmarks) |
| 5. DuckDB Removal | 🔜 **NEXT** | Remove all DuckDB code |
| 6. Semantic Search | 🔜 | Optional chromem-go integration |

### Session 2026-02-01 Evening - ✅ PHASE 4 COMPLETE

**Phase 4 - Bleve Backend** ✅ **COMPLETED**:

All Tasks Complete:
- [x] Add Bleve dependency: `go get github.com/blevesearch/bleve/v2`
- [x] Add afero dependency: `go get github.com/spf13/afero`
- [x] Create `internal/search/bleve/doc.go`
- [x] Create `internal/search/bleve/mapping.go` - document mapping with field weights
- [x] Create `internal/search/bleve/storage.go` - afero adapter
- [x] Create `internal/search/bleve/query.go` - query AST translation
- [x] Create `internal/search/bleve/index.go` - full Index implementation
- [x] Write query translation tests (14 tests)
- [x] Write index integration tests (8 tests)
- [x] Add benchmarks for performance verification (6 benchmarks)
- [x] Integrate parser with Index for query string support (FindByQueryString method)
- [x] Fix tag matching bug (TermQuery → MatchQuery)

### Files Created This Session

```
internal/search/bleve/
├── doc.go                      # Package documentation
├── mapping.go                  # Document mapping (field weights: path=1000, title=500, etc.)
├── storage.go                  # AferoStorage adapter
├── query.go                    # TranslateQuery, TranslateFindOpts (tag bug fixed)
├── index.go                    # Index implementation + FindByQueryString
├── index_test.go               # 8 integration tests
├── query_test.go               # 14 query translation tests
├── parser_integration_test.go  # 6 parser integration tests
└── index_bench_test.go         # 6 performance benchmarks
```

**Performance**: 36 tests passing, search <1ms, all targets met

---

## 🎯 Next: Phase 5 - DuckDB Removal

**Ready to Start**: All Phase 4 tasks complete

### Phase 5 Tasks (Proposed)

- [ ] Create Phase 5 document
- [ ] Audit codebase for DuckDB references
- [ ] Replace NoteService.SearchNotes with Index.Find
- [ ] Remove DbService entirely
- [ ] Update CLI commands (notes search, notes list, etc.)
- [ ] Migrate SQL views to query DSL
- [ ] Remove DuckDB from dependencies
- [ ] Verify binary size reduction (target: <15MB)
- [ ] Verify startup time improvement (target: <100ms)
- [ ] Update documentation

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
