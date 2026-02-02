---
id: a7b8c9d0
title: Phase 5.2.6 - Service Method Cleanup
created_at: 2026-02-02T14:03:00+10:30
updated_at: 2026-02-02T14:03:00+10:30
status: in-progress
epic_id: f661c068
phase_id: 02df510c
assigned_to: 2026-02-02-afternoon
---

# Task: Phase 5.2.6 - Service Method Cleanup

## Objective

Complete removal of DbService from the codebase by:
1. Removing DbService references from NoteService and NotebookService
2. Removing DbService initialization from cmd/root.go
3. Deleting internal/services/db.go
4. Removing any remaining DuckDB-related helper code
5. Cleaning up view.go SQL conversion helpers

## Related Story

Part of Epic [epic-f661c068](epic-f661c068-remove-duckdb-alternative-search.md) - Remove DuckDB

## Steps

### Step 1: Remove DbService from NoteService

**Files**: `internal/services/note.go`

- [ ] Remove `dbService *DbService` field from struct
- [ ] Remove `db` parameter from `NewNoteService()` constructor
- [ ] Remove TODO comments about DbService removal
- [ ] Update all callers (in NotebookService)

### Step 2: Remove DbService from NotebookService

**Files**: `internal/services/notebook.go`

- [ ] Remove `dbService *DbService` field from struct
- [ ] Remove `db` parameter from `NewNotebookService()` constructor
- [ ] Update constructor calls to `NewNoteService()` (remove db parameter)
- [ ] Update all callers (in cmd/root.go)

### Step 3: Remove DbService from cmd/root.go

**Files**: `cmd/root.go`

- [ ] Remove `dbService *services.DbService` global variable
- [ ] Remove `dbService = services.NewDbService()` initialization
- [ ] Remove DbService cleanup in PersistentPostRun
- [ ] Update `NewNotebookService()` call (remove dbService parameter)

### Step 4: Handle cmd/notes_view.go

**Files**: `cmd/notes_view.go`

This file uses `dbService.GetDB(ctx)` for SQL views. Options:
1. Remove SQL view support entirely (breaking change)
2. Keep minimal SQL support for views only
3. Migrate views to query DSL

**Decision**: Remove SQL view support as part of DuckDB removal. SQL views are an advanced feature that conflicts with the pure Go goal.

- [ ] Remove SQL view execution code
- [ ] Update command to show error message for SQL views
- [ ] Document breaking change in CHANGELOG

### Step 5: Clean up internal/services/view.go

**Files**: `internal/services/view.go`

- [ ] Remove `convertToJSONSafe()` function (DuckDB-specific type handling)
- [ ] Review other view-related functions for SQL dependencies
- [ ] Keep non-SQL view helpers if any exist

### Step 6: Delete internal/services/db.go

**Files**: `internal/services/db.go`

- [ ] Verify no remaining references
- [ ] Delete entire file (373 lines)
- [ ] Remove from imports across codebase

### Step 7: Testing

- [ ] Run `mise run test` - verify all tests pass
- [ ] Check for compilation errors
- [ ] Verify services initialize correctly
- [ ] Test notebook operations work without DbService

### Step 8: Documentation

- [ ] Update CHANGELOG.md with SQL view removal
- [ ] Update AGENTS.md architecture section
- [ ] Update any remaining docs mentioning DbService

## Expected Outcome

- Zero DbService references in codebase
- internal/services/db.go deleted
- All services using only Bleve Index
- SQL views no longer supported (breaking change)
- All tests passing
- Clean compilation with no unused imports

## Actual Outcome

*To be filled upon completion*

## Lessons Learned

*To be filled upon completion*

## Notes

**Breaking Changes**:
- SQL views no longer supported
- Users must migrate views to query DSL or custom Go code
- This completes the transition away from SQL entirely

**Files to Modify**:
1. `internal/services/note.go` - Remove DbService field
2. `internal/services/notebook.go` - Remove DbService field
3. `cmd/root.go` - Remove DbService initialization
4. `cmd/notes_view.go` - Remove SQL view support
5. `internal/services/view.go` - Clean up SQL helpers
6. `internal/services/db.go` - Delete entire file

**Estimated Effort**: 1-2 hours
- Service refactoring: 30 mins
- View migration/removal: 30 mins
- Testing: 30 mins
- Documentation: 30 mins
