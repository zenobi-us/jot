---
id: 5d8f7e3a
title: Phase 5 Task 2.3 - Migrate SearchWithConditions()
created_at: 2026-02-02T07:39:00+10:30
updated_at: 2026-02-02T07:39:00+10:30
status: ready
priority: high
epic_id: f661c068
phase_id: 02df510c
assigned_to: next-session
---

# Phase 5 Task 2.3 - Migrate SearchWithConditions()

## Objective

Replace all DuckDB dependencies in NoteService.SearchWithConditions() with Bleve-based search implementation. This method handles complex queries with multiple conditions (tags, paths, dates, etc.).

## Context - Phase 5.2.2 Complete

**Previous Work**:
- Phase 5.2.1: Added `index search.Index` field to NoteService ✅
- Phase 5.2.2: Migrated getAllNotes() to Index.Find() ✅
  - Created documentToNote() converter function
  - Created testutil.CreateTestIndex() helper
  - 171/172 tests passing

**Status**: Ready to proceed with SearchWithConditions() migration

## Current Implementation

### SearchWithConditions() Method

```go
// Current (DuckDB-based)
func (s *NoteService) SearchWithConditions(ctx context.Context, conditions []QueryCondition) ([]Note, error) {
    // Builds SQL WHERE clause from QueryCondition structs
    // Uses searchService.BuildWhereClauseWithGlob()
    // Executes SQL query via DbService
}
```

### QueryCondition Structure

Located in `internal/search/conditions.go`:
```go
type QueryCondition struct {
    Type     string // "and", "or", "not"
    Field    string // "data.tag", "path", "title", "links-to", "linked-by"
    Operator string // "=" (only equality currently)
    Value    string // user value
}
```

### Current Calling Pattern

- Called from CLI commands: `notes search --tag`, `notes list --path`, etc.
- Builds WHERE clauses like: `WHERE path LIKE 'epics/%' AND data.tag = 'work'`
- Returns sorted Note results

## Implementation Plan

### Step 1: Analyze SearchWithConditions() Usage

**Action Items**:
- [ ] Read `internal/services/note.go` line by line
- [ ] Understand how each QueryCondition type maps to SQL
- [ ] Document all calling patterns from CLI commands
- [ ] Identify edge cases (dates, path globs, tag lists)

**Key Questions**:
- How are tag conditions handled? (single tag or multiple?)
- How are path conditions handled? (exact match or glob?)
- What date formats are supported?
- Are "links-to" and "linked-by" actually implemented or stubs?

### Step 2: Design QueryCondition → search.Query Mapping

**Mapping Strategy**:

| QueryCondition | → | search.Query |
|---|---|---|
| `data.tag=work` | → | Tag query with "work" |
| `path=epics/*` | → | Path prefix query "epics/" |
| `title=meeting` | → | Title field match |
| `and` type | → | BooleanQuery AND |
| `or` type | → | BooleanQuery OR |
| `not` type | → | BooleanQuery NOT |

**Design Decision**: 
- Should SearchService build the query.Query AST?
- Or build it directly in NoteService.SearchWithConditions()?

**Recommendation**:
- Add new method to SearchService: `BuildQuery(conditions []QueryCondition) (*search.Query, error)`
- Keeps separation of concerns
- Reusable if needed elsewhere
- Mirrors current SQL building approach

### Step 3: Implement SearchService.BuildQuery()

**Location**: `internal/services/search.go`

**Method Signature**:
```go
func (s *SearchService) BuildQuery(conditions []QueryCondition) (*search.Query, error)
```

**Implementation Logic**:
1. Start with empty BooleanQuery or single-condition query
2. For each condition:
   - Map field to appropriate query type (tag, path, title, etc.)
   - Apply operator (only "=" currently)
   - Apply condition type (and, or, not)
3. Return final query.Query AST

**Edge Cases to Handle**:
- Empty conditions → match-all query
- Single condition → return as-is (not wrapped in BooleanQuery)
- Multiple conditions → wrap in BooleanQuery
- Unknown fields → return error with helpful message
- Links-to/linked-by → return error (not yet supported)

### Step 4: Update NoteService.SearchWithConditions()

**Changes**:
```go
func (s *NoteService) SearchWithConditions(ctx context.Context, conditions []QueryCondition) ([]Note, error) {
    // Build query AST from conditions
    query, err := s.searchService.BuildQuery(conditions)
    if err != nil {
        return nil, err
    }
    
    // Execute search using Index
    results, err := s.index.Find(ctx, search.FindOpts{
        Query: query,
        // Handle sorting if needed
    })
    if err != nil {
        return nil, err
    }
    
    // Convert results to Notes
    notes := make([]Note, len(results.Items))
    for i, result := range results.Items {
        notes[i] = documentToNote(result.Document)
    }
    
    return notes, nil
}
```

### Step 5: Update Tests

**Test Files**:
- `internal/services/note_test.go` - NoteService tests
- `internal/services/search_test.go` - SearchService tests (if exists)

**Tests to Add**:
- [ ] BuildQuery() with single tag condition
- [ ] BuildQuery() with path glob condition
- [ ] BuildQuery() with multiple AND conditions
- [ ] BuildQuery() with OR conditions
- [ ] BuildQuery() with NOT conditions
- [ ] BuildQuery() with empty conditions
- [ ] BuildQuery() with unknown field → error
- [ ] SearchWithConditions() integration tests

**Tests to Update**:
- [ ] All existing SearchWithConditions() tests
- [ ] Update mock Index calls
- [ ] Verify result ordering maintained

### Step 6: Integration Testing

**Manual Testing**:
- [ ] Test `notes search "tag:work"`
- [ ] Test `notes list --path epics/`
- [ ] Test combined conditions
- [ ] Test with no matches
- [ ] Test with large result sets

**CLI Commands Using SearchWithConditions**:
- Identify from grep: `grep -r "SearchWithConditions" cmd/`
- Test each affected command

## Expected Outcome

### New Code Structure

```go
// SearchService - new method
func (s *SearchService) BuildQuery(conditions []QueryCondition) (*search.Query, error)

// NoteService - updated method
func (s *NoteService) SearchWithConditions(ctx context.Context, conditions []QueryCondition) ([]Note, error)
```

### Files Modified
- `internal/services/search.go` - Add BuildQuery()
- `internal/services/note.go` - Update SearchWithConditions()
- `internal/services/note_test.go` - Update tests (40+ test cases)
- `internal/services/search_test.go` - Add new tests for BuildQuery()

### Tests
- All existing SearchWithConditions() tests updated and passing
- New BuildQuery() unit tests added (10-15 new tests)

## Testing Strategy

### Unit Tests

**File**: `internal/services/search_test.go`

Tests for BuildQuery():
- [ ] Single tag condition
- [ ] Single path condition
- [ ] Multiple AND conditions
- [ ] Multiple OR conditions
- [ ] NOT conditions
- [ ] Empty conditions
- [ ] Unknown field error
- [ ] Invalid operator error
- [ ] Complex nested conditions

### Integration Tests

**File**: `internal/services/note_test.go`

Tests for SearchWithConditions():
- [ ] Find notes by tag
- [ ] Find notes by path prefix
- [ ] Find notes with multiple conditions
- [ ] Empty result set
- [ ] Large result set

## Risks & Mitigation

### Risk: Complex Condition Logic

**Description**: QueryCondition mapping could have edge cases

**Mitigation**:
- Write unit tests for each condition type
- Add logging/error messages for debugging
- Test with CLI commands before marking complete

### Risk: Links-to/Linked-by Not Implemented

**Description**: These conditions are stubs in current code

**Mitigation**:
- Check if they're actually used in CLI
- If not used: Return clear error message
- If used: Mark as TODO for Phase 5 Task 2.5

### Risk: Test Failures

**Description**: Updating 40+ tests could introduce regressions

**Mitigation**:
- Update tests incrementally
- Run tests after each major change
- Use testutil.CreateTestIndex() for consistency

## Dependencies

**Requires**:
- Phase 5.2.2 complete (getAllNotes() migrated) ✅
- testutil.CreateTestIndex() available ✅
- documentToNote() converter available ✅
- search.Query AST available ✅

**Blocks**:
- Phase 5.2.4 (Migrate Count())
- Phase 5.2.5 (Remove SQL methods)
- Full DuckDB removal

## Design Decisions

### Decision 1: Where Should BuildQuery() Live?

**Option A**: SearchService.BuildQuery()
- Pros: Follows current pattern (BuildWhereClauseWithGlob)
- Pros: Reusable if needed elsewhere
- Cons: SearchService becomes more complex

**Option B**: Inline in NoteService.SearchWithConditions()
- Pros: Simpler, self-contained
- Cons: Not reusable, harder to test

**Chosen**: **Option A** - SearchService.BuildQuery()
- Maintains separation of concerns
- Testable independently
- Mirrors current SQL builder pattern

### Decision 2: Error Handling for Unknown Fields

**Option A**: Return error immediately
- Pros: Clear failure, user sees message
- Cons: May break existing code using unknown fields

**Option B**: Log warning, skip condition
- Pros: Graceful degradation
- Cons: Silent failures are confusing

**Chosen**: **Option A** - Return error
- Helps users catch typos early
- Consistent with strict typing philosophy

## Actual Outcome

*To be filled upon completion*

## Lessons Learned

*To be filled upon completion*

## Notes

- This is critical path for DuckDB removal
- Once complete, only Count() and SQL methods remain
- Performance should match or exceed current (Bleve is fast)
- This is last complex migration before cleanup phases
