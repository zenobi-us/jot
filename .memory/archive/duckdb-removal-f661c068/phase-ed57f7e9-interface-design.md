---
id: ed57f7e9
title: Phase 2 - Interface Design
created_at: 2026-02-01T15:59:00+10:30
updated_at: 2026-02-01T16:30:00+10:30
status: complete
epic_id: f661c068
start_criteria: Research phase complete, strategic decisions confirmed
end_criteria: All interfaces defined, ready for Query Parser implementation
---

# Phase 2 - Interface Design

## Overview

Design the core interfaces for the pure Go search implementation. This phase defines the contracts between components without implementing the actual search logic.

**Key Principle**: Inspired by zk's excellent interface design, adapted for OpenNotes' needs with afero integration throughout.

## Deliverables

### 1. Search Index Interface (`NoteIndex`)
Define the contract for indexing and querying notes:
- `Add(note)` - Index a note
- `Update(note)` - Update indexed note
- `Remove(path)` - Remove from index
- `Find(opts)` - Query with filters
- `Reindex()` - Full reindex
- `Stats()` - Index statistics

### 2. Query AST Structures
Define the Abstract Syntax Tree for parsed queries:
- `Query` - Root node
- `Term` - Simple text term
- `FieldExpr` - Field:value expression (tag:work)
- `NegationExpr` - Negated expression (-archived)
- `DateExpr` - Date comparison (created:>2024-01-01)
- `RangeExpr` - Date range (created:2024-01..2024-06)

### 3. Query Options (`NoteFindOpts`)
Functional options pattern for query building:
- `WithTags(tags)` - Filter by tags
- `WithPath(prefix)` - Filter by path
- `WithDateRange(field, start, end)` - Date filtering
- `ExcludingPaths(paths)` - Exclusions
- `WithSort(field, direction)` - Ordering
- `WithLimit(n)` - Pagination

### 4. File Storage Interface (`FileStorage`)
Afero-compatible interface for filesystem access:
- `Read(path)` - Read file content
- `Write(path, content)` - Write file
- `Walk(root, fn)` - Traverse directory
- `Stat(path)` - File metadata
- `Exists(path)` - Check existence

### 5. Search Result Types
- `SearchResult` - Single result with ranking
- `SearchResults` - Collection with metadata
- `Snippet` - Context-aware text snippet with highlighting

## Tasks

| Task | Description | Status |
|------|-------------|--------|
| 1. Define `NoteIndex` interface | Core search contract | ✅ |
| 2. Define `Query` AST structures | Parse tree for DSL | ✅ |
| 3. Define `NoteFindOpts` | Query options with functional pattern | ✅ |
| 4. Define `FileStorage` interface | Afero adapter contract | ✅ |
| 5. Define result types | SearchResult, Snippet, etc. | ✅ |
| 6. Create interface file structure | `internal/search/` package | ✅ |
| 7. Write interface documentation | Godoc comments | ✅ |
| 8. Review with human | Phase gate | 🔜 |

## Dependencies

- **Research**: [research-f410e3ba-search-replacement-synthesis.md](research-f410e3ba-search-replacement-synthesis.md)
- **ZK Insights**: [research-parallel/subtopic-1-zk-search/insights.md](research-parallel/subtopic-1-zk-search/insights.md)
- **Query DSL**: [research-parallel/subtopic-3-query-dsl/](research-parallel/subtopic-3-query-dsl/)

## Design Decisions

### Package Structure
```
internal/
├── search/                 # NEW - Search package
│   ├── index.go           # NoteIndex interface
│   ├── query.go           # Query AST types
│   ├── options.go         # NoteFindOpts
│   ├── result.go          # Result types
│   └── storage.go         # FileStorage interface (afero adapter)
├── services/              # EXISTING
│   ├── note.go           # Will use search.NoteIndex
│   └── search.go         # Will be replaced/refactored
```

### Interface Philosophy

1. **Small interfaces** - Single responsibility, composable
2. **Functional options** - Immutable, chainable query building
3. **Context support** - All operations accept `context.Context`
4. **Error handling** - Explicit errors, no panics
5. **Testability** - In-memory implementations for testing

### zk Patterns to Adopt

From research:
- ✅ `NoteIndex` interface pattern
- ✅ `FileStorage` interface (afero-compatible)
- ✅ `NoteFindOpts` functional options
- ✅ BM25 ranking concept (in result type)
- ✅ Channel-based streaming for large results

### Deferred to Later Phases

- ❌ Bleve implementation (Phase 4)
- ❌ Query parser (Phase 3)
- ❌ DuckDB removal (Phase 5)
- ❌ Semantic search (Phase 6)

## Next Steps

After Phase 2 completion:
1. Human review of interfaces
2. Proceed to Phase 3: Query Parser (Participle-based)

## Notes

- Interfaces should be defined in pure Go, no external dependencies
- All interfaces should work with `afero.Fs` for testability
- Query AST should be parser-agnostic (could use Participle or hand-written)
