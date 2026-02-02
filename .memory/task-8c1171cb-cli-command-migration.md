---
id: 8c1171cb
title: Phase 5.2.5 - CLI Command Migration
created_at: 2026-02-02T13:32:00+10:30
updated_at: 2026-02-02T13:32:00+10:30
status: in-progress
epic_id: f661c068
phase_id: 02df510c
assigned_to: 2026-02-02-afternoon
---

# Task: Phase 5.2.5 - CLI Command Migration

## Objective

Remove all SQL interface methods from NoteService and update CLI commands to use the new Bleve-based search exclusively. This completes the CLI layer migration by removing the last DuckDB dependencies.

## Related Story

Part of Epic [epic-f661c068](epic-f661c068-remove-duckdb-alternative-search.md) - Remove DuckDB

## Steps

### Step 1: Audit CLI Commands for DuckDB Usage ✅

- [x] Check `cmd/notes_search.go` for SQL usage
  - Uses `ExecuteSQLSafe()` for `--sql` flag
  - Uses `SearchNotes()` for regular search (already migrated)
- [x] Check `cmd/notes_list.go` for SQL usage
  - Uses `SearchNotes()` only - no DuckDB
  - No changes needed
- [x] Identify methods to remove from NoteService:
  - `ExecuteSQLSafe(sql string) error`
  - `Query(sql string) (*sql.Rows, error)`

### Step 2: Remove SQL Methods from NoteService

- [ ] Remove `ExecuteSQLSafe()` method from `internal/services/note.go`
- [ ] Remove `Query()` method from `internal/services/note.go`
- [ ] Remove any SQL-related imports (database/sql, etc.)
- [ ] Update comments/documentation

### Step 3: Update CLI Commands

#### notes_search.go

- [ ] Remove `--sql` flag completely (breaking change)
- [ ] Update help text:
  - Remove SQL examples
  - Add query DSL examples
  - Add migration guide
- [ ] Update error messages to guide users to new syntax
- [ ] Add deprecation notice in help text

#### notes_list.go

- [ ] Verify no changes needed (already uses SearchNotes())
- [ ] No action required

### Step 4: Verify requireNotebook() Helper

- [ ] Check that `requireNotebook()` in `cmd/root.go` calls `NotebookService.createIndex()`
- [ ] Already implemented in Phase 5.2.3 - verify still working

### Step 5: Testing

- [ ] Run `mise run test` - all tests should pass
- [ ] Manual CLI testing:
  - `opennotes notes search "tag:work"` - should work
  - `opennotes notes search "content AND tag:urgent"` - should work
  - `opennotes notes list` - should work
  - `opennotes notes search --sql "SELECT ..."` - should fail with helpful error

### Step 6: Documentation

- [ ] Update CHANGELOG.md with breaking change notice
- [ ] Add migration guide for users with custom SQL queries
- [ ] Update README.md examples to use new query syntax

## Expected Outcome

- Zero SQL interface methods in NoteService
- `notes search --sql` flag removed
- All CLI commands using Bleve-based search
- Clear migration guide for users
- All tests passing
- Documentation updated

## Actual Outcome

*To be filled upon completion*

## Lessons Learned

*To be filled upon completion*

## Notes

**Breaking Changes**:
- `--sql` flag removed from `notes search`
- Custom SQL queries no longer supported
- Users must migrate to new query DSL

**Migration Path for Users**:
1. Simple queries: `tag:work` instead of `SELECT * FROM notes WHERE tags LIKE '%work%'`
2. Complex queries: Use query DSL operators (AND, OR, NOT, parentheses)
3. Advanced features: Wait for Phase 5.3 (link graph) or use external tools

**Dependencies**:
- Phase 5.2.3 complete (SearchWithConditions migrated)
- Phase 5.2.4 complete (Count migrated)
