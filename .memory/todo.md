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
| 4. Bleve Backend | 🔄 **IN PROGRESS** | Full-text indexing implementation |
| 5. DuckDB Removal | 🔜 | Remove all DuckDB code |
| 6. Semantic Search | 🔜 | Optional chromem-go integration |

### Session 2026-02-01 Evening - In Progress

**Phase 4 - Bleve Backend** 🔄:

Completed:
- [x] Add Bleve dependency: `go get github.com/blevesearch/bleve/v2`
- [x] Add afero dependency: `go get github.com/spf13/afero`
- [x] Create `internal/search/bleve/doc.go`
- [x] Create `internal/search/bleve/mapping.go` - document mapping with field weights
- [x] Create `internal/search/bleve/storage.go` - afero adapter
- [x] Create `internal/search/bleve/query.go` - query AST translation
- [x] Create `internal/search/bleve/index.go` - full Index implementation
- [x] Write query translation tests (14 tests)
- [x] Write index integration tests (8 tests)
- [x] Lint passes, all 22 tests pass

Remaining:
- [ ] Add benchmarks for performance verification
- [ ] Integrate parser with Index for query string support
- [ ] Add frontmatter parsing in Reindex method

### Files Created This Session

```
internal/search/bleve/
├── doc.go           # Package documentation
├── mapping.go       # Document mapping (field weights: path=1000, title=500, etc.)
├── storage.go       # AferoStorage adapter
├── query.go         # TranslateQuery, TranslateFindOpts
├── index.go         # Index implementation
├── index_test.go    # 8 integration tests
└── query_test.go    # 14 query translation tests
```

---

## 🔜 Next: Complete Phase 4

### Remaining Tasks

- [ ] Benchmark Index.Find vs DuckDB (target: <25ms)
- [ ] Benchmark Index.Add for bulk indexing (target: 10k docs in <500ms)
- [ ] Add parser integration method to Index
- [ ] Complete Reindex with frontmatter parsing

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
