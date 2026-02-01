---
id: 3639018c
title: Phase 5 Task 2 - Migrate NoteService to Bleve Index
created_at: 2026-02-01T21:35:00+10:30
updated_at: 2026-02-01T21:35:00+10:30
status: in-progress
epic_id: f661c068
phase_id: 02df510c
assigned_to: 2026-02-01-evening
---

# Phase 5 Task 2 - Migrate NoteService to Bleve Index

## Objective

Replace all DuckDB dependencies in NoteService with the new Bleve-based search implementation. This is the core migration that enables removing DuckDB entirely.

## Steps

### Phase 2.1: Design & Interface Changes

- [ ] Update NoteService struct
  - Remove `dbService *DbService` field
  - Add `index search.Index` field
  - Keep `configService`, `searchService`, `notebookPath`, `log`
- [ ] Update constructor signature
  - From: `NewNoteService(cfg *ConfigService, db *DbService, notebookPath string)`
  - To: `NewNoteService(cfg *ConfigService, index search.Index, notebookPath string)`
- [ ] Design Note struct mapping to search.Document
  - Map Note fields to Document fields
  - Handle metadata serialization

### Phase 2.2: Migrate getAllNotes()

**Current**: Uses `read_markdown()` DuckDB function
```go
sqlQuery := `SELECT * FROM read_markdown(?, include_filepath:=true)`
rows, err := db.QueryContext(ctx, sqlQuery, glob)
```

**New Implementation**:
- [ ] Use `index.Find()` with empty query (match all)
- [ ] Or: Walk filesystem + parse markdown directly
- [ ] Return []Note from search.Document results
- [ ] Handle relative path calculation
- [ ] Test with existing test suite

**Decision**: Use Index.Find() - it already has all notes indexed

### Phase 2.3: Migrate Count()

**Current**: `SELECT COUNT(*) FROM read_markdown(?)`

**New Implementation**:
- [ ] Use `index.Count()` method
- [ ] Or: `len(getAllNotes())` if simpler
- [ ] Test accuracy

**Decision**: Use Index.Count() for efficiency

### Phase 2.4: Migrate SearchWithConditions()

**Current**: Builds SQL WHERE clauses from QueryCondition structs

**New Implementation**:
- [ ] Examine QueryCondition struct in search service
- [ ] Map QueryConditions to search.Query AST
- [ ] Use `index.Find()` with constructed query
- [ ] Handle sorting (ORDER BY file_path)
- [ ] Test with existing boolean query tests

**Note**: This already uses searchService.BuildWhereClauseWithGlob() - may need refactor

### Phase 2.5: Remove SQL Methods

**Methods to DELETE**:
- [ ] `ExecuteSQLSafe()` - User SQL queries (Option A: Remove)
- [ ] `Query()` - Direct SQL passthrough (Remove)
- [ ] `ValidateSQL()` - SQL validation (Remove)
- [ ] `rowsToMaps()` - SQL row conversion (Remove)

**Rationale**: Clean break from SQL interface per epic goals

### Phase 2.6: Update SearchNotes()

**Current**: Calls `getAllNotes()` then filters in-memory

**New Implementation**:
- [ ] Review if in-memory filtering is still needed
- [ ] Consider using Index.Find() with query directly
- [ ] Or keep current approach (already pure Go)
- [ ] Test fuzzy search functionality

**Decision**: May already work after getAllNotes() migration

### Phase 2.7: Helper Functions

**Keep These** (no DuckDB dependency):
- [x] `ParseDataFlags()` - Pure string parsing
- [x] `ResolvePath()` - Pure path manipulation
- [x] `Note.DisplayName()` - Pure Note method

## Expected Outcome

### New NoteService Interface

```go
type NoteService struct {
    configService *ConfigService
    index         search.Index
    searchService *SearchService
    notebookPath  string
    log           zerolog.Logger
}

// Constructor
func NewNoteService(cfg *ConfigService, index search.Index, notebookPath string) *NoteService

// Core methods (migrated)
func (s *NoteService) SearchNotes(ctx context.Context, query string, fuzzy bool) ([]Note, error)
func (s *NoteService) getAllNotes(ctx context.Context) ([]Note, error)
func (s *NoteService) Count(ctx context.Context) (int, error)
func (s *NoteService) SearchWithConditions(ctx context.Context, conditions []QueryCondition) ([]Note, error)

// Helper methods (unchanged)
func (s *NoteService) ParseDataFlags(dataFlags []string) (map[string]interface{}, error)
func (s *NoteService) ResolvePath(notebookRoot, inputPath, slugifiedTitle string) string
func (n *Note) DisplayName() string

// REMOVED:
// - ExecuteSQLSafe()
// - Query()
// - ValidateSQL()
// - rowsToMaps()
```

### Note to search.Document Mapping

```go
// Note struct (current)
type Note struct {
    File struct {
        Filepath string
        Relative string
    }
    Content  string
    Metadata map[string]any
}

// Maps to search.Document (from Phase 2)
type Document struct {
    Path     string
    Title    string
    Content  string
    Tags     []string
    Created  time.Time
    Modified time.Time
    Metadata map[string]any
}

// Mapping logic:
Note.File.Relative → Document.Path
Note.Metadata["title"] → Document.Title
Note.Content → Document.Content
Note.Metadata["tags"] → Document.Tags
Note.Metadata["created"] → Document.Created
Note.Metadata["modified"] → Document.Modified
Note.Metadata → Document.Metadata
```

## Testing Strategy

### Unit Tests to Update

**File**: `internal/services/note_test.go`

Tests needing changes:
- [ ] All tests using `NewDbService()` (61 instances)
  - Replace with mock/in-memory Index
- [ ] Tests for `ExecuteSQLSafe()` - DELETE
- [ ] Tests for `Query()` - DELETE  
- [ ] Tests for `ValidateSQL()` - DELETE
- [ ] Tests for `SearchNotes()` - May work as-is
- [ ] Tests for `getAllNotes()` - Update assertions
- [ ] Tests for `Count()` - Update to use Index
- [ ] Tests for `SearchWithConditions()` - Update to use Index

### New Tests to Add

- [ ] Test Note → Document conversion
- [ ] Test Document → Note conversion
- [ ] Test getAllNotes() with empty index
- [ ] Test getAllNotes() with multiple notes
- [ ] Test Count() accuracy
- [ ] Test SearchWithConditions() with various conditions

## Migration Checklist

### Pre-Migration
- [ ] Review all NoteService tests to understand expected behavior
- [ ] Document any edge cases in current implementation
- [ ] Check if any CLI commands call removed methods directly

### Implementation
- [ ] Update struct and constructor
- [ ] Implement Note ↔ Document conversion functions
- [ ] Migrate getAllNotes()
- [ ] Migrate Count()
- [ ] Migrate SearchWithConditions()
- [ ] Remove SQL methods
- [ ] Update imports (remove database/sql)

### Post-Migration
- [ ] Update all test files
- [ ] Run test suite: `mise run test`
- [ ] Fix any compilation errors
- [ ] Verify all tests pass
- [ ] Check test coverage maintained

## Dependencies

**Requires**:
- Phase 4 complete (Bleve backend implemented)
- search.Index interface available
- search.Document struct available

**Blocks**:
- Phase 5.3 (cmd/root.go initialization)
- Phase 5.4 (CLI command updates)
- Phase 5.5 (NotebookService migration)

## Design Decisions

### Decision 1: How to Get All Notes?

**Option A**: Use `Index.Find()` with match-all query
- Pros: Consistent with search approach
- Cons: Depends on index being populated

**Option B**: Walk filesystem directly
- Pros: Direct, simple
- Cons: Duplicates indexing logic

**Chosen**: **Option A** - Use Index.Find()
- Index already has all notes
- Consistent with architecture
- Leverages existing indexing

### Decision 2: How to Handle SearchWithConditions()?

**Current**: Uses `searchService.BuildWhereClauseWithGlob()` which builds SQL

**QueryCondition Structure**:
```go
type QueryCondition struct {
    Type     string // "and", "or", "not"
    Field    string // "data.tag", "path", "title", "links-to", "linked-by"
    Operator string // "=" (only equality supported)
    Value    string // user value
}
```

**Option A**: Keep searchService for building queries
- Pros: Maintains separation of concerns
- Cons: SearchService builds SQL, not query AST

**Option B**: Refactor searchService to build query AST
- Pros: Clean architecture
- Cons: More refactoring work (but necessary)

**Option C**: Inline query building in NoteService
- Pros: Simple, direct
- Cons: Mixes concerns

**Chosen**: **Option B** - Refactor SearchService
- Add new method: `BuildQuery(conditions []QueryCondition) (*search.Query, error)`
- Map QueryConditions to search.Query terms
- Keep existing SQL methods temporarily (for gradual migration)
- Clean separation: SearchService = query building, Index = execution

**QueryCondition → search.Query Mapping**:
- `data.tag=work` → Tag query with "work"
- `path=epics/*` → Path prefix "epics/"
- `title=meeting` → Title field match
- `links-to=target` → Not supported yet (Phase 6 feature?)
- `linked-by=source` → Not supported yet (Phase 6 feature?)

### Decision 3: SQL Methods?

**Chosen**: **Remove entirely** (Option A from audit)
- ExecuteSQLSafe() - DELETE
- Query() - DELETE
- ValidateSQL() - DELETE
- Users use new query DSL only
- Clean break per epic goals

## Actual Outcome

*To be filled upon completion*

## Lessons Learned

*To be filled upon completion*

## Notes

- Note struct stays the same (maintain backward compatibility)
- Only internal implementation changes
- CLI commands should not need changes (they use NoteService methods)
- This is the critical path - once complete, can remove DbService
- SearchService may need refactoring to build query AST instead of SQL
