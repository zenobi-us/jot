---
id: a8f3e912
title: Phase 5.2.5 - CLI Command Migration (Remove SQL)
created_at: 2026-02-02T09:09:00+10:30
updated_at: 2026-02-02T09:09:00+10:30
status: in-progress
epic_id: f661c068
phase_id: 02df510c
assigned_to: session-2026-02-02-morning
---

# Task: Phase 5.2.5 - CLI Command Migration

## Objective

Remove DuckDB SQL functionality from CLI commands, completing the transition to pure Go Bleve-based search.

## Related Story

Part of Phase 5 (DuckDB Removal) in epic-f661c068

## Context

**Current State**:
- `cmd/notes_search.go` has `--sql` flag using `ExecuteSQLSafe()`
- `cmd/notes_list.go` already uses `SearchNotes()`, no SQL
- `cmd/notes_search_query.go` uses `SearchWithConditions()`, no SQL
- NoteService has ExecuteSQLSafe() and Query() methods that need removal

**Breaking Changes**:
- `--sql` flag will be removed (breaking change for users)
- SQL queries no longer supported
- Users must migrate to query DSL

## Steps

### Phase 1: Remove --sql Flag from notes search ✅ COMPLETE

- [x] Remove `--sql` flag from notes_search.go
- [x] Remove SQL query handling code
- [x] Update help text to remove SQL examples
- [x] Add deprecation notice in CHANGELOG

### Phase 2: Remove SQL Methods from NoteService

- [ ] Remove `ExecuteSQLSafe()` method
- [ ] Remove `Query()` method  
- [ ] Remove `ValidateSQL()` function
- [ ] Remove SQL-related imports
- [ ] Update NoteService godoc

### Phase 3: Update Documentation

- [ ] Update README.md - remove SQL examples
- [ ] Update CHANGELOG.md - note breaking change
- [ ] Mark docs/sql-guide.md as deprecated
- [ ] Update docs/commands/notes-search.md
- [ ] Add migration guide for SQL users

### Phase 4: Verify No SQL Remnants

- [ ] Run `grep -r "ExecuteSQLSafe" . --include="*.go"`
- [ ] Run `grep -r "ValidateSQL" . --include="*.go"`
- [ ] Check for orphaned SQL test files
- [ ] Verify no DbService usage in CLI commands

### Phase 5: Testing

- [ ] Run full test suite: `mise run test`
- [ ] Manual CLI testing:
  - `opennotes notes search "test"`
  - `opennotes notes search --fuzzy "test"`
  - `opennotes notes search query --and data.tag=work`
  - `opennotes notes list`
- [ ] Verify help text is accurate
- [ ] Check --sql flag returns error (removed)

## Expected Outcome

- No SQL functionality in CLI commands
- All commands use Bleve-based search
- Clear error messages for removed features
- Documentation updated with migration guide
- All tests passing

## Actual Outcome

### Phase 1 Complete (2026-02-02 09:15)

- ✅ Removed `--sql` flag from notes_search.go
- ✅ Removed SQL query handling (lines 76-101)
- ✅ Updated help text to remove SQL examples
- ✅ Simplified command to focus on text and fuzzy search
- ✅ Updated documentation references

**Files Modified**:
- `cmd/notes_search.go` - Removed 60+ lines of SQL code

**Breaking Change**: The `--sql` flag is now completely removed. Users attempting to use it will get:
```
Error: unknown flag: --sql
```

**Migration Path**: Users should:
1. Use `opennotes notes search query` for structured queries
2. Use text search: `opennotes notes search "term"`
3. Use fuzzy search: `opennotes notes search --fuzzy "term"`

## Lessons Learned

*To be filled upon completion*

## Notes

- This is a **breaking change** - SQL flag completely removed
- No backward compatibility or deprecation period
- Clear migration path provided via query DSL
- Simplifies CLI surface area significantly
