---
id: 3a5e0381
title: Bleve Search Backend Implementation
epic_id: f661c068
created_at: 2026-02-01T16:00:00+10:30
updated_at: 2026-02-01T16:00:00+10:30
status: in-progress
start_criteria: Interfaces and parser complete (phases 2-3)
end_criteria: Full Index implementation with all tests passing
---

# Phase 4: Bleve Search Backend Implementation

## Overview

Implement the `search.Index` interface using Bleve for full-text search with BM25 ranking. This is the core search engine that replaces DuckDB.

## Deliverables

1. **`internal/search/bleve/` package** with:
   - `index.go` - Index implementation
   - `mapping.go` - Document mapping with field weights
   - `query.go` - Query AST to Bleve query translation
   - `storage.go` - afero-based storage adapter
   - `doc.go` - Package documentation

2. **Comprehensive tests** in `*_test.go` files

3. **Working integration** with existing interfaces

## Tasks

- [x] Add Bleve dependency
- [x] Create `internal/search/bleve/doc.go`
- [x] Create `internal/search/bleve/mapping.go` - document mapping
- [x] Create `internal/search/bleve/storage.go` - afero adapter for Bleve
- [x] Create `internal/search/bleve/query.go` - query translation
- [x] Create `internal/search/bleve/index.go` - main implementation
- [x] Write unit tests for query translation
- [x] Write integration tests for full search flow
- [ ] Benchmark performance against targets
- [ ] Add integration with parser to convert query strings

## Design Decisions

### Document Mapping

From research synthesis - field weights for BM25:
- `path`: 1000 (strongest signal - exact path matches)
- `title`: 500 (strong signal)
- `tags`: 300 (medium signal)
- `lead`: 50 (first paragraph, higher than body)
- `body`: 1 (baseline)

### Index Location

Index stored in `.opennotes/index/` within notebook root.

### Change Detection

Use xxhash checksum for efficient change detection during incremental updates.

## Dependencies

- `github.com/blevesearch/bleve/v2` - Full-text search engine
- `internal/search/` interfaces - Already complete
- `internal/search/parser/` - Query parser (Phase 3)

## Next Steps

After this phase:
1. Phase 5: DuckDB Removal - Remove all DuckDB code
2. Phase 6: Semantic Search (optional) - chromem-go integration
