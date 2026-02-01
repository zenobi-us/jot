---
id: 02df510c
title: Phase 5 - DuckDB Removal & Cleanup
created_at: 2026-02-01T21:17:00+10:30
updated_at: 2026-02-02T07:39:00+10:30
status: in-progress
epic_id: f661c068
start_criteria: Phase 4 (Bleve Backend) complete with all tests passing
end_criteria: All DuckDB code removed, CLI commands migrated, binary size <15MB, tests passing
---

# Phase 5 - DuckDB Removal & Cleanup

## Overview

Complete removal of DuckDB from the OpenNotes codebase. This phase replaces all DuckDB-dependent code with the new Bleve-based search implementation and verifies that performance targets are met.

**Prerequisites**: Phase 4 complete with Bleve backend fully implemented and tested.

## Deliverables

1. **DbService Removal**: Complete removal of `internal/services/db.go`
2. **NoteService Migration**: Replace DuckDB queries with Bleve Index calls
3. **CLI Command Updates**: Migrate all commands using DuckDB to new search
4. **SQL Views Migration**: Convert SQL views to query DSL or Go helpers
5. **Dependency Cleanup**: Remove DuckDB from `go.mod`
6. **Performance Validation**: Verify binary size (<15MB) and startup time (<100ms)
7. **Documentation Updates**: Update docs to reflect new query syntax

## Tasks

### 1. Codebase Audit ✅ COMPLETE

- [x] Scan codebase for all DuckDB references
  - `grep -r "duckdb" --include="*.go" .`
  - `grep -r "DbService" --include="*.go" .`
  - `grep -r "markdown_scan" --include="*.go" .`
- [x] List all files that need modification
- [x] Create comprehensive task checklist

**Result**: [task-9b9e6fb4-phase5-codebase-audit.md](.memory/task-9b9e6fb4-phase5-codebase-audit.md)
- 14 production files + 8 dependencies identified
- Migration order established
- NoteService usage patterns documented

### 2. Service Layer Migration

- [ ] **Audit NoteService** (`internal/services/note.go`)
  - Identify all methods using DbService
  - Plan replacement with Index interface
- [ ] **Migrate NoteService.SearchNotes**
  - Replace DuckDB query with `Index.Find()`
  - Update return type handling
  - Add tests for migration
- [ ] **Migrate other NoteService methods**
  - Handle any remaining DuckDB dependencies
  - Ensure all methods use new search backend
- [ ] **Remove DbService** (`internal/services/db.go`)
  - Delete entire file
  - Remove from service initialization in `cmd/root.go`

### 3. CLI Command Migration

- [ ] **Audit all CLI commands** for DuckDB usage
  - `cmd/notes_search.go`
  - `cmd/notes_list.go`
  - Any others using SQL queries
- [ ] **Migrate `notes search` command**
  - Update to use new query DSL
  - Update help text with new syntax examples
  - Test with various query patterns
- [ ] **Migrate `notes list` command**
  - Replace SQL with Index.Find + filters
  - Maintain existing output format
- [ ] **Migrate SQL views** (if any exist)
  - Convert to query DSL equivalents
  - Or implement as Go helper functions
  - Document new approach

### 4. Dependency Cleanup

- [ ] **Remove DuckDB from go.mod**
  - `go get -u` to clean dependencies
  - `go mod tidy`
- [ ] **Verify no CGO dependencies remain** for search
  - Check `go.mod` for CGO-requiring packages
  - Document remaining CGO uses (if any)
- [ ] **Check for unused imports**
  - Run linter to catch orphaned imports
  - Clean up any DuckDB-related test utilities

### 5. Integration & Testing

- [ ] **Run full test suite**
  - `mise run test`
  - Verify all tests pass
- [ ] **Add integration tests for migrated commands**
  - Test notes search with various query patterns
  - Test notes list with filters
  - Verify output format consistency
- [ ] **Manual CLI testing**
  - Search with tags: `opennotes notes search "tag:work"`
  - Search with exclusions: `opennotes notes search "-archived"`
  - Search with date ranges
  - List all notes
  - Verify performance feels fast

### 6. Performance Validation

- [ ] **Measure binary size**
  - `ls -lh dist/opennotes`
  - Target: <15MB (down from 64MB)
  - Document actual size
- [ ] **Measure startup time**
  - `hyperfine "opennotes --version"`
  - Target: <100ms (down from 500ms)
  - Document actual time
- [ ] **Measure search performance**
  - Benchmark typical searches
  - Verify <25ms latency target
  - Compare with Phase 4 benchmarks

### 7. Documentation Updates

- [ ] **Update README.md**
  - Remove any DuckDB references
  - Document new query syntax
  - Add query examples
- [ ] **Update AGENTS.md**
  - Remove DuckDB from architecture
  - Document new search architecture
  - Update technical decisions section
- [ ] **Update internal docs**
  - Services documentation
  - Architecture diagrams
  - Any SQL query examples

## Dependencies

**Requires**:
- Phase 4 (Bleve Backend) complete
- All Bleve tests passing
- Performance benchmarks meeting targets

**Blocks**:
- Phase 6 (Semantic Search) - needs clean codebase

## Next Steps

After phase completion:
1. Archive phase document
2. Create learning document with migration insights
3. Update summary.md with Phase 5 completion
4. Decide whether to proceed to Phase 6 (Semantic Search) or complete epic

## Expected Outcome

- Zero DuckDB references in codebase
- All CLI commands using new Bleve search
- Binary size <15MB
- Startup time <100ms
- All tests passing
- Documentation updated

## Actual Outcome

*To be filled upon completion*

## Lessons Learned

*To be filled upon completion*

## Notes

- This is a breaking change for users with custom SQL queries
- Views using SQL will need migration or removal
- Query syntax is fundamentally different (Gmail-style vs SQL)
- Performance should improve significantly (smaller binary, faster startup)
- No rollback plan - this is a one-way migration
