---
id: f661c068
title: Remove DuckDB - Alternative Search Implementation
created_at: 2026-02-01T14:39:00+10:30
updated_at: 2026-02-01T14:39:00+10:30
status: proposed
---

# Remove DuckDB - Alternative Search Implementation

## Vision/Goal

Replace OpenNotes' current DuckDB-based search backend with a native Go search implementation inspired by **zk-org/zk**, enabling:

1. **Filesystem abstraction** with `spf13/afero` for mockable file system access
2. **Removal of DuckDB dependency** to eliminate compilation complexity and enable full VFS control
3. **Expressive search DSL** with capabilities comparable or superior to current SQL-based search
4. **Preserved user experience** for views and templates - no breaking changes to user-facing features

This epic represents a fundamental architectural shift from database-backed search to native Go file system indexing and query execution.

## Success Criteria

### Prime Concepts (Non-Negotiable)

- [x] **Concept 1**: Filesystem operations abstracted via `spf13/afero` interface
  - All file I/O goes through `afero.Fs` interface
  - Tests use `afero.MemMapFs` for in-memory filesystem
  - Production uses `afero.OsFs` for real filesystem
  
- [ ] **Concept 2**: Complete removal of DuckDB dependency
  - No DuckDB imports in codebase
  - Markdown extension no longer required
  - Simplified binary build (no C++ dependencies)
  - Smaller binary size
  
- [ ] **Concept 3**: Expressive search DSL implementation
  - Parse and execute complex search queries
  - Support for: text search, field filters, boolean logic, sorting, pagination
  - Performance comparable to current DuckDB implementation
  - Extensible for future query capabilities
  
- [ ] **Concept 4**: Zero user-facing functional regression
  - Views work identically from user perspective
  - Templates render the same output
  - All existing commands produce same results
  - Query syntax may differ, but capabilities must match or exceed

### Functional Requirements

- [ ] Search supports: full-text, frontmatter fields, tags, dates, paths
- [ ] Boolean operators: AND, OR, NOT
- [ ] Sorting by: modified, created, title, path
- [ ] Pagination with offset/limit
- [ ] View system unchanged (same view definitions work)
- [ ] Template system unchanged (same templates work)
- [ ] Performance: sub-100ms for typical queries on <10,000 notes

### Quality Requirements

- [ ] 100% afero abstraction - no direct `os.` or `ioutil.` calls
- [ ] Comprehensive test coverage with in-memory filesystem
- [ ] Zero flaky tests (deterministic file system access)
- [ ] Migration guide for users (query syntax changes documented)
- [ ] Benchmark suite proving comparable performance

### Technical Requirements

- [ ] Parser for new query DSL
- [ ] Indexer for note metadata and content
- [ ] Query executor with filtering and sorting
- [ ] Afero-based file operations throughout codebase
- [ ] Backward compatibility layer for existing views (if needed)

## Phases

| Phase | Title | Status | File |
|-------|-------|--------|------|
| 1 | Research & Analysis | 🔜 `proposed` | TBD |
| 2 | Query DSL Design | 🔜 `proposed` | TBD |
| 3 | Indexer Implementation | 🔜 `proposed` | TBD |
| 4 | Query Executor | 🔜 `proposed` | TBD |
| 5 | Afero Migration | 🔜 `proposed` | TBD |
| 6 | DuckDB Removal | 🔜 `proposed` | TBD |
| 7 | Testing & Validation | 🔜 `proposed` | TBD |

## Dependencies

### Technical Dependencies

- **spf13/afero** (v1.11.0+) - Filesystem abstraction layer
- **zk-org/zk** (reference implementation) - Search architecture inspiration
- **Existing OpenNotes codebase** - Current DuckDB integration points

### Knowledge Dependencies

- [research-dbb5cdc8-zk-search-analysis.md](research-dbb5cdc8-zk-search-analysis.md) - zk search implementation analysis
- Current DuckDB integration points (to be documented)
- View and template system architecture (to be documented)

### Blocking Dependencies

- Must complete research phase before designing new DSL
- Cannot remove DuckDB until replacement is feature-complete
- Afero migration can proceed in parallel with search work

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    OpenNotes (Current)                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐       ┌──────────┐       ┌────────────┐      │
│  │  Search  │──────▶│  DuckDB  │──────▶│  Markdown  │      │
│  │ Service  │       │  + Ext   │       │   Files    │      │
│  └──────────┘       └──────────┘       └────────────┘      │
│                           │                                 │
│                           ▼                                 │
│                    ┌─────────────┐                          │
│                    │   Direct    │                          │
│                    │  os.File    │                          │
│                    └─────────────┘                          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    OpenNotes (Target)                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐       ┌──────────┐       ┌────────────┐      │
│  │  Search  │──────▶│  Query   │──────▶│   Afero    │      │
│  │ Service  │       │ Executor │       │    Fs      │      │
│  └──────────┘       └──────────┘       └─────┬──────┘      │
│       │                   ▲                   │             │
│       │                   │                   ▼             │
│       │            ┌──────────────┐    ┌────────────┐      │
│       └───────────▶│   Indexer    │    │  Markdown  │      │
│                    │  (in-memory) │    │   Files    │      │
│                    └──────────────┘    └────────────┘      │
│                                                             │
│  Features:                                                  │
│  • Mockable filesystem (afero.MemMapFs)                    │
│  • No C++ dependencies (pure Go)                           │
│  • Fast in-memory indexing                                 │
│  • Flexible query DSL                                      │
└─────────────────────────────────────────────────────────────┘
```

## Migration Strategy

### Phase Approach

1. **Research**: Analyze zk-org/zk search architecture (current task)
2. **Design**: Define new query DSL and indexer architecture
3. **Parallel Implementation**: Build new search alongside DuckDB
4. **Feature Flag**: Allow toggling between implementations
5. **Validation**: Comprehensive testing with both systems
6. **Cutover**: Make new search default, deprecate DuckDB
7. **Cleanup**: Remove DuckDB code and dependencies

### Backward Compatibility

- Views: Keep same view definition format, translate queries if needed
- Templates: No changes required (data structure stays same)
- CLI flags: Maintain existing flags, add new DSL syntax options
- Migration script: Convert existing SQL queries to new DSL (if applicable)

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Query DSL not expressive enough | High | Medium | Research zk + other tools thoroughly; design with extensibility |
| Performance regression | High | Medium | Benchmark suite; optimize indexer; consider caching |
| Breaking user workflows | High | Low | Feature flag; extensive testing; migration guide |
| Afero abstraction leaks | Medium | Medium | Strict code review; 100% test coverage with MemMapFs |
| Implementation complexity | Medium | High | Start with MVP DSL; iterate based on real usage |
| Timeline longer than expected | Low | High | Phased approach; can ship partial improvements |

## Success Metrics

### Performance Benchmarks

- **Search latency**: <100ms for 10k notes (match or beat DuckDB)
- **Index build time**: <500ms for 10k notes
- **Memory usage**: <100MB index size for 10k notes
- **Binary size**: Reduce by >10MB (removing DuckDB)

### Quality Metrics

- **Test coverage**: >90% for new search code
- **Zero test flakes**: All tests deterministic with afero.MemMapFs
- **Migration success**: 100% of existing queries translatable
- **User satisfaction**: No regression in usability or capability

## Notes

- This epic is independent of the pi-opennotes extension (epic-1f41631e)
- Research phase must complete before implementation design begins
- Consider creating a feature flag for gradual rollout
- Document lessons learned from DuckDB experience (both pros and cons)
- Explore zk's indexing strategy, query parser, and result ranking

## Related Work

- **Epic**: [epic-1f41631e-pi-opennotes-extension.md](epic-1f41631e-pi-opennotes-extension.md) - Pi extension (uses current OpenNotes)
- **Research**: [research-4e873bd0-vfs-summary.md](research-4e873bd0-vfs-summary.md) - VFS integration research
- **Research**: [research-7f4c2e1a-afero-vfs-integration.md](research-7f4c2e1a-afero-vfs-integration.md) - Afero exploration
- **Research**: [research-8a9b0c1d-duckdb-filesystem-findings.md](research-8a9b0c1d-duckdb-filesystem-findings.md) - DuckDB limitations
