---
id: f661c068
title: Remove DuckDB - Pure Go Search Implementation
created_at: 2026-02-01T14:39:00+10:30
updated_at: 2026-02-01T15:59:00+10:30
status: planning
---

# Remove DuckDB - Pure Go Search Implementation

## Vision/Goal

**Complete removal of DuckDB** from OpenNotes, replacing it with a pure Go search implementation. This is a clean break, not a migration:

1. **No DuckDB at all** - complete removal of the dependency
2. **Pure Go search** using Bleve for full-text indexing
3. **Gmail-style DSL** for intuitive query syntax
4. **Filesystem abstraction** with `spf13/afero` throughout
5. **Optional semantic search** with chromem-go (future enhancement)

> **Note**: This is NOT a migration. We are completely replacing DuckDB. There is no dual-support period, no feature flags to toggle between implementations. DuckDB is being removed entirely.

## Success Criteria

### Prime Concepts (Non-Negotiable)

- [x] **Concept 1**: Filesystem operations abstracted via `spf13/afero` interface
  - All file I/O goes through `afero.Fs` interface
  - Tests use `afero.MemMapFs` for in-memory filesystem
  - Production uses `afero.OsFs` for real filesystem
  
- [ ] **Concept 2**: Complete removal of DuckDB dependency
  - No DuckDB imports in codebase
  - No markdown extension
  - No CGO dependencies for search
  - Smaller binary size (target: <15MB from 64MB)
  - Faster startup (target: <100ms from 500ms)
  
- [ ] **Concept 3**: Pure Go search implementation
  - Bleve for full-text indexing with BM25 ranking
  - Gmail-style DSL: `tag:work`, `title:meeting`, `-archived`
  - Participle parser for query parsing
  - afero-compatible persistence
  
- [ ] **Concept 4**: Feature parity with current search
  - Full-text search across note content
  - Frontmatter field filtering
  - Tag filtering
  - Date range queries
  - Path prefix filtering
  - Sorting and pagination

### Performance Targets

| Metric | Current (DuckDB) | Target (Pure Go) | Improvement |
|--------|------------------|------------------|-------------|
| Binary size | 64 MB | <15 MB | -78% |
| Startup time | 500ms | <100ms | -80% |
| Search latency | 29.9ms | <25ms | -16% |
| Index build 10k | N/A | <500ms | New capability |

## Phases

| Phase | Title | Status | File |
|-------|-------|--------|------|
| 1 | Research & Analysis | ✅ `complete` | [research-f410e3ba-search-replacement-synthesis.md](research-f410e3ba-search-replacement-synthesis.md) |
| 2 | Interface Design | ✅ `complete` | [phase-ed57f7e9-interface-design.md](phase-ed57f7e9-interface-design.md) |
| 3 | Query Parser | ✅ `complete` | [phase-f29cef1b-query-parser.md](phase-f29cef1b-query-parser.md) |
| 4 | Bleve Search Backend | 🔄 `in-progress` | [phase-3a5e0381-bleve-backend.md](phase-3a5e0381-bleve-backend.md) |
| 5 | DuckDB Removal & Cleanup | 🔜 `proposed` | TBD |
| 6 | Semantic Search (Optional) | 🔜 `proposed` | TBD |

## Research Findings Summary

**Completed Research**: [research-f410e3ba-search-replacement-synthesis.md](research-f410e3ba-search-replacement-synthesis.md)

### Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Full-text search | **Bleve** | Pure Go, BM25 ranking, 9+ years mature |
| Query syntax | **Gmail-style DSL** | `tag:work -archived` - familiar, safe, concise |
| Parser | **Participle** | Go-idiomatic, type-safe AST |
| Semantic search | **chromem-go** | Optional, progressive enhancement |
| Filesystem | **afero** | Testable, mockable, VFS ready |

### What We're NOT Doing

- ❌ No "migration period" with dual support
- ❌ No feature flags to toggle between DuckDB and new search
- ❌ No SQL query compatibility layer
- ❌ No keeping DuckDB "just in case"

### What We ARE Doing

- ✅ Complete DuckDB removal
- ✅ New Gmail-style query syntax (different from SQL)
- ✅ Pure Go implementation (no CGO)
- ✅ Clean, fresh search architecture

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    OpenNotes (Target)                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐       ┌──────────┐       ┌────────────┐      │
│  │   CLI    │──────▶│  Query   │──────▶│   Afero    │      │
│  │ (Cobra)  │       │  Parser  │       │    Fs      │      │
│  └──────────┘       │(Participle│      └─────┬──────┘      │
│                     └──────────┘             │              │
│                          │                   ▼              │
│                          ▼            ┌────────────┐       │
│                   ┌──────────────┐    │  Markdown  │       │
│                   │   Bleve      │◄───│   Files    │       │
│                   │ (full-text)  │    └────────────┘       │
│                   └──────────────┘                          │
│                          │                                  │
│                          ▼                                  │
│                   ┌──────────────┐                          │
│                   │  chromem-go  │ (optional, Phase 6)      │
│                   │  (vectors)   │                          │
│                   └──────────────┘                          │
└─────────────────────────────────────────────────────────────┘

REMOVED:
  ❌ DuckDB
  ❌ Markdown extension
  ❌ CGO dependencies
  ❌ SQL query interface
```

## Implementation Approach

### No Migration - Clean Replacement

This is a clean break:

1. **Design new interfaces** - Based on zk patterns, adapted for our needs
2. **Implement Bleve backend** - Complete search functionality
3. **Build query parser** - Gmail-style DSL
4. **Remove DuckDB entirely** - Delete all DuckDB code
5. **Update commands** - Use new search throughout

### Query Syntax Change

Users will use new syntax. Examples:

| Old (SQL) | New (DSL) |
|-----------|-----------|
| `SELECT * FROM notes WHERE tag='work'` | `tag:work` |
| `SELECT * FROM notes WHERE title LIKE '%meeting%'` | `title:meeting` |
| `SELECT * FROM notes WHERE NOT archived` | `-archived` |
| `SELECT * FROM notes WHERE created > '2024-01-01'` | `created:>2024-01-01` |
| Complex SQL joins | Not needed - simpler model |

## Dependencies

### New Dependencies

- **blevesearch/bleve** - Full-text search engine
- **alecthomas/participle** - Parser combinator
- **spf13/afero** - Filesystem abstraction
- **philippgille/chromem-go** - Vector search (optional, Phase 6)

### Removed Dependencies

- ~~marcboeker/go-duckdb~~ - Removed
- ~~DuckDB markdown extension~~ - Removed
- ~~CGO for DuckDB~~ - Removed

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Missing features in new search | Medium | Low | Feature parity checklist before removal |
| Performance regression | Medium | Low | Benchmark gates before completion |
| Learning curve for new syntax | Low | Medium | Clear documentation, intuitive DSL |
| Binary size not reaching target | Low | Low | Profile and optimize |

## Notes

- This epic completely removes DuckDB - no legacy code remains
- SQL queries will no longer work after implementation
- Views need updating to use new query syntax
- Documentation will need full rewrite for query syntax

## Related Work

- **Research**: [research-f410e3ba-search-replacement-synthesis.md](research-f410e3ba-search-replacement-synthesis.md)
- **Research Details**: `.memory/research-parallel/subtopic-*/`
