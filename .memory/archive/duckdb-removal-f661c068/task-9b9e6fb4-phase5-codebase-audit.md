---
id: 9b9e6fb4
title: Phase 5 Task 1 - Codebase Audit for DuckDB Removal
created_at: 2026-02-01T21:25:00+10:30
updated_at: 2026-02-01T21:25:00+10:30
status: completed
epic_id: f661c068
phase_id: 02df510c
assigned_to: 2026-02-01-evening
---

# Phase 5 Task 1 - Codebase Audit for DuckDB Removal

## Objective

Systematically identify all DuckDB references in the codebase to create a comprehensive migration plan.

## Steps

- [x] Scan for "duckdb" references
- [x] Scan for "DbService" references
- [x] Scan for "markdown_scan" references
- [x] Check go.mod dependencies
- [x] Identify affected CLI commands
- [x] Document findings

## Expected Outcome

Complete list of files requiring modification, organized by category.

## Actual Outcome

### Summary

**Total Files Affected**: 14 production files + dependencies
- 3 service implementation files
- 3 service test files  
- 1 CLI root file
- 6 CLI command files
- 4 e2e test files
- 8 go.mod DuckDB dependencies

### Category 1: Core Service Files (DELETE ENTIRELY)

#### `internal/services/db.go` - **DELETE**
- 2x `sql.Open("duckdb", "")`
- 1x import `_ "github.com/duckdb/duckdb-go/v2"`
- Contains: DbService struct, GetDB, GetReadOnlyDB, Query, preprocessSQL
- **Action**: Delete entire file (will be unused after NoteService migration)

#### `internal/services/db_test.go` - **DELETE**
- 41 test functions for DbService
- Tests: GetDB, GetReadOnlyDB, Query, preprocessSQL, markdown extension
- **Action**: Delete entire file (DbService removed)

### Category 2: Service Files Requiring Migration

#### `internal/services/note.go` - **MIGRATE**
- Has `dbService *DbService` field
- Constructor: `NewNoteService(cfg *ConfigService, db *DbService, notebookPath string)`
- Methods likely using DbService.Query()
- 1x comment about `duckdb.Map` type
- **Action**: 
  - Remove `dbService` field
  - Add `index search.Index` field
  - Update constructor signature
  - Migrate all Query() calls to Index.Find()

#### `internal/services/note_test.go` - **UPDATE**
- 61 instances of `services.NewDbService()`
- Tests using DbService to test NoteService
- **Action**: 
  - Replace DbService with mock/test Index
  - Update test setup to use in-memory Bleve index
  - Verify tests still pass

#### `internal/services/notebook.go` - **MIGRATE**
- Has `dbService *DbService` field
- Constructor: `NewNotebookService(cfg *ConfigService, db *DbService)`
- **Action**:
  - Check if DbService is actually used
  - If used, migrate to Index
  - If not used, just remove field

#### `internal/services/notebook_test.go` - **UPDATE**
- 23 instances of `services.NewDbService()`
- **Action**:
  - Remove DbService from test setup
  - Update constructor calls

#### `internal/services/view.go` - **UPDATE**
- 2x comments about `duckdb.Map` and `duckdb` types
- Function: `convertToJSONSafe` handles duckdb types
- **Action**:
  - Remove duckdb type handling
  - Simplify type conversion (no more duckdb.Map)

#### `internal/services/view_special_test.go` - **UPDATE**
- 6 instances of `NewDbService()`
- **Action**: Remove DbService from test setup

### Category 3: CLI Root Initialization

#### `cmd/root.go` - **MIGRATE**
- Has `dbService *services.DbService` global variable
- Init: `dbService = services.NewDbService()`
- **Action**:
  - Remove dbService global
  - Add index initialization (create Bleve index per notebook)
  - Update service constructors

### Category 4: CLI Commands (Check for Direct Usage)

Files to audit for noteService/notebookService usage:
- `cmd/notebook_addcontext.go` ✓
- `cmd/notebook_create.go` ✓
- `cmd/notebook_list.go` ✓
- `cmd/notebook_register.go` ✓
- `cmd/notes_list.go` ✓
- `cmd/root.go` ✓ (already listed above)

**Action for each**: Verify they don't call DbService directly

### Category 5: E2E Tests

#### `tests/e2e/stress_test.go` - **UPDATE**
- 4 instances of `services.NewDbService()`
- **Action**: Migrate to Index-based tests

#### `tests/e2e/concurrency_test.go` - **UPDATE**
- 9 instances of `services.NewDbService()`
- 2 test functions: `TestDbService_ConnectionPoolStress`, `TestDbService_ConcurrentInitialization`
- **Action**: 
  - Delete DbService-specific tests
  - Add Index concurrency tests if needed

#### `tests/e2e/filesystem_errors_test.go` - **UPDATE**
- 1 instance of `services.NewDbService()`
- **Action**: Update to use Index

### Category 6: Dependencies (go.mod)

DuckDB dependencies to remove:
```
github.com/duckdb/duckdb-go/v2 v2.5.4
github.com/duckdb/duckdb-go-bindings v0.1.24 // indirect
github.com/duckdb/duckdb-go-bindings/darwin-amd64 v0.1.24 // indirect
github.com/duckdb/duckdb-go-bindings/darwin-arm64 v0.1.24 // indirect
github.com/duckdb/duckdb-go-bindings/linux-amd64 v0.1.24 // indirect
github.com/duckdb/duckdb-go-bindings/linux-arm64 v0.1.24 // indirect
github.com/duckdb/duckdb-go-bindings/windows-amd64 v0.1.24 // indirect
github.com/duckdb/duckdb-go/arrowmapping v0.0.27 // indirect
github.com/duckdb/duckdb-go/mapping v0.0.27 // indirect
```

**Action**: Run `go mod tidy` after removing imports

## Migration Order (Recommended)

1. **Phase 5.1**: Migrate `internal/services/note.go` (core functionality)
2. **Phase 5.2**: Update `cmd/root.go` (service initialization)
3. **Phase 5.3**: Check/update CLI commands (if needed)
4. **Phase 5.4**: Migrate `internal/services/notebook.go` (check if used)
5. **Phase 5.5**: Update `internal/services/view.go` (remove duckdb types)
6. **Phase 5.6**: Delete `internal/services/db.go` and `internal/services/db_test.go`
7. **Phase 5.7**: Update all test files (note_test.go, notebook_test.go, view_special_test.go)
8. **Phase 5.8**: Update e2e tests
9. **Phase 5.9**: Remove DuckDB from go.mod via `go mod tidy`
10. **Phase 5.10**: Run full test suite
11. **Phase 5.11**: Verify binary size and startup time

## Key Insights

### What Uses DbService?

1. **NoteService** - Definitely uses it (has field, constructor param)
2. **NotebookService** - Has field, but usage unclear (needs verification)
3. **cmd/root.go** - Initializes and passes to constructors
4. **All tests** - Use `NewDbService()` for test setup

### No Direct SQL in CLI Commands

Good news: CLI commands don't appear to execute SQL directly. They use NoteService/NotebookService methods, which internally use DbService.

This means CLI commands won't need changes if we migrate the service layer cleanly.

### Type Conversion Issues

The `view.go` file handles `duckdb.Map` types for JSON serialization. After removal, we need to ensure new search results serialize correctly.

## NoteService DbService Usage Analysis

After examining `internal/services/note.go`:

**Methods using DbService**:
1. `getAllNotes()` - Uses `read_markdown()` DuckDB function with glob pattern
2. `Count()` - Uses `COUNT(*)` on `read_markdown()`
3. `ExecuteSQLSafe()` - Executes user SQL queries (validation + preprocessing)
4. `Query()` - Direct passthrough to `DbService.Query()`
5. `SearchWithConditions()` - Uses `read_markdown()` with WHERE clauses

**Key Observations**:
- All methods use DuckDB's `read_markdown()` function to scan filesystem
- SearchNotes() calls `getAllNotes()`, then filters in-memory (already pure Go!)
- Most complex is `SearchWithConditions()` - builds SQL WHERE clauses
- `ExecuteSQLSafe()` provides user SQL query interface (may need rethink)

**Migration Strategy**:
1. Replace `getAllNotes()` with filesystem walk + markdown parsing
2. Use Bleve Index for `SearchWithConditions()` 
3. Decide: Keep `ExecuteSQLSafe()` with query DSL or remove entirely
4. `Count()` becomes simple len(notes) or Index.Count()

## Next Steps

1. ✅ Codebase audit complete
2. Create detailed NoteService migration task
3. Examine NotebookService to see if it uses DbService
4. Begin NoteService migration implementation
5. Proceed through migration order

## Lessons Learned

- DuckDB is well-isolated in service layer (good architecture)
- CLI commands depend on services, not directly on DbService
- Test coverage is comprehensive (41 DbService tests to migrate/replace)
- Clear separation of concerns will make migration cleaner

## Notes

- No `markdown_scan` references found (good - no direct DuckDB queries in code)
- All SQL goes through DbService.Query() - single point of migration
- 161 total `NewDbService()` calls across all files - but most are in tests
