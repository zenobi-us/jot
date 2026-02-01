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

## Assessment Complete ✅

**Date**: 2026-02-02T07:54:00+10:30  
**Status**: Ready for implementation  
**Full Assessment**: `.memory/assessment-phase523-migration.md`

### Key Findings

**✅ FEASIBLE** - Migration can proceed with limitations.

**Supported Fields** (11/13):
- ✅ All metadata fields (`data.tag`, `data.status`, etc.) - Direct mapping
- ✅ `path` field - Prefix queries + wildcard fallback
- ✅ `title` field - Direct mapping
- ✅ AND/OR/NOT boolean logic - Full support via BooleanQuery

**⚠️ NOT SUPPORTED** (2/13):
- ❌ `links-to` - Requires link graph index (Phase 5.3)
- ❌ `linked-by` - Requires link graph index (Phase 5.3)

### Migration Strategy

**Approach**: Migrate core functionality, return clear error for link queries.

**Query Building**:
- Add `SearchService.BuildQuery(conditions) -> search.Query`
- Convert QueryCondition structs to search.Expr AST
- Mirrors current `BuildWhereClauseWithGlob()` pattern

**NoteService Changes**:
- Replace SQL building with `BuildQuery()`
- Replace `db.QueryContext()` with `index.Find()`
- Reuse `documentToNote()` converter (from Phase 5.2.2)
- Maintain sort order (ORDER BY file_path)

### Risk Mitigation

**High Risk - Link Queries**:
- Return actionable error message with Phase 5.3 reference
- Document breaking change in CHANGELOG.md
- Provide SQL workaround in error message

**Medium Risk - Path Globs**:
- Optimize prefix patterns (`projects/*` → PrefixQuery)
- Fallback to wildcard for complex patterns (`**/tasks/*.md`)
- Document performance characteristics

### Test Plan

**Unit Tests** (15 new):
- Field mapping (tag, status, path, title)
- Boolean logic (AND, OR, NOT, mixed)
- Path globbing (prefix, wildcard, doublestar)
- Error cases (unknown fields, link queries)

**Integration Tests** (40 existing):
- Update all SearchWithConditions() tests
- Replace DuckDB with testutil.CreateTestIndex()
- Mark link query tests as TODO Phase 5.3

**Manual CLI Testing**:
- Basic queries, OR conditions, NOT conditions
- Path queries, complex combinations
- Link queries (verify error message)

### Implementation Phases

**Phase 1**: Implement BuildQuery() (2-3h)
**Phase 2**: Update SearchWithConditions() (1h)  
**Phase 3**: Update Tests (3-4h)
**Phase 4**: Documentation (1-2h)
**Phase 5**: Integration & Verification (1h)

**Total**: 8-11 hours

### Breaking Changes

**Link Queries**:
```bash
# Will error after migration
opennotes notes search query --and links-to=docs/*.md
opennotes notes search query --and linked-by=plan.md
```

**Error Message**:
```
Error: link queries are not yet supported

Field 'links-to' requires a dedicated link graph index, 
which is planned for Phase 5.3.

Temporary workaround: Use SQL query interface
Track progress: github.com/zenobi-us/opennotes/issues/XXX

Supported fields: data.tag, data.status, path, title, ...
```

### Success Criteria

- ✅ All metadata/path/title queries work
- ✅ AND/OR/NOT logic correct
- ✅ 171/172 tests passing (1 link test expects error)
- ✅ Clear error for link queries
- ✅ Documentation updated
- ✅ Performance maintained

### Next Steps

**Ready to Proceed**: Yes ✅  
**Blockers**: None  
**Dependencies**: Phase 5.2.2 complete ✅

**Recommended Action**: Begin implementation following 5-phase plan.

## Actual Outcome - Phase 1 Complete

**Date Completed**: 2026-02-02  
**Time Taken**: ~1 hour (estimated 2-3 hours)  
**Tests**: 189/190 passing (1 pre-existing failure in TestSpecialViewExecutor_BrokenLinks)  
**Status**: ✅ Phase 1 Complete

### Phase 1 Implementation Summary

**Files Modified**:
- `internal/services/search.go` - Added BuildQuery() method and 5 helper functions
- `internal/services/search_test.go` - Created with 18 unit tests (27 total including subtests)

**New Methods**:
- `BuildQuery(ctx, conditions) -> (*search.Query, error)` - Main conversion method
- `conditionToExpr(cond) -> (search.Expr, error)` - Convert single condition to expression
- `buildMetadataExpr(cond) -> (search.Expr, error)` - Handle metadata fields (data.*)
- `buildPathExpr(cond) -> (search.Expr, error)` - Handle path with glob optimization
- `detectWildcardType(pattern) -> search.WildcardType` - Classify wildcard patterns
- `buildLinkQueryError(field) -> error` - Generate helpful error for link queries

**Test Coverage**:
- ✅ Single tag condition
- ✅ Multiple AND conditions
- ✅ Multiple OR conditions (nested OrExpr)
- ✅ Single OR condition (unwrapped)
- ✅ NOT conditions (NotExpr wrapper)
- ✅ Path prefix optimization (projects/* → OpPrefix)
- ✅ Path with trailing slash (projects/ → OpPrefix)
- ✅ Complex wildcards (**/tasks/*.md → WildcardExpr)
- ✅ Exact path match (OpEquals)
- ✅ Title field queries
- ✅ Empty conditions (returns empty Query)
- ✅ Link queries return clear error (links-to, linked-by)
- ✅ Unknown fields return error
- ✅ Mixed AND/OR/NOT conditions
- ✅ Tags alias (data.tags → metadata.tag)
- ✅ All 9 metadata fields tested
- ✅ Invalid condition type error

### What Went Well

1. **Test-Driven Development**: Wrote all 18 tests first, then implemented BuildQuery()
2. **Clean Implementation**: Code follows existing patterns and Go conventions
3. **Performance Optimization**: Prefix queries optimized for simple globs (projects/*)
4. **Error Messages**: Clear, actionable error for unsupported link queries
5. **No Regressions**: All existing tests still passing (189/190)

### Challenges

None - implementation went smoothly following the detailed plan.

### Deviations from Plan

None - followed the plan exactly as specified.

## Lessons Learned

*To be filled upon completion*

## Notes

- This is critical path for DuckDB removal
- Once complete, only Count() and SQL methods remain
- Performance should match or exceed current (Bleve is fast)
- This is last complex migration before cleanup phases
- **Link queries deferred to Phase 5.3** (separate link graph index needed)
