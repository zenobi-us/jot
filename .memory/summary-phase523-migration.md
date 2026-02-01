# Phase 5.2.3 - SearchWithConditions() Migration Summary

## Status: Ready for Implementation ✅

**Task**: Migrate `NoteService.SearchWithConditions()` from DuckDB to Bleve Index  
**Estimated Time**: 8-11 hours  
**Difficulty**: Medium (link queries are hard, rest is straightforward)

---

## Quick Decision Matrix

| Question | Answer |
|----------|--------|
| Can we migrate? | ✅ Yes (with limitations) |
| What works? | All metadata, path, title queries |
| What doesn't? | Link queries (links-to, linked-by) |
| Breaking changes? | Yes - link queries return error |
| Workaround? | Use SQL interface temporarily |
| When will links work? | Phase 5.3 (link graph index) |
| Should we proceed? | ✅ Yes - defer links to Phase 5.3 |

---

## Key Documents

1. **Assessment** (`.memory/assessment-phase523-migration.md`)
   - Full technical feasibility analysis
   - Risk assessment and mitigations
   - Query mapping details (QueryCondition → Bleve)
   - 24KB comprehensive review

2. **Implementation Plan** (`.memory/plan-phase523-implementation.md`)
   - Step-by-step implementation guide
   - Complete code examples
   - All test cases
   - 33KB detailed roadmap

3. **Task File** (`.memory/task-5d8f7e3a-phase523-searchwithconditions.md`)
   - Original task definition
   - Updated with assessment findings
   - Success criteria

---

## Technical Summary

### What Works (11/13 fields)

**Metadata Fields** (9):
```go
data.tag, data.status, data.priority, data.assignee,
data.author, data.type, data.category, data.project, data.sprint
```
→ Direct mapping to `metadata.field` in Bleve

**Path Field** (1):
```go
path=projects/*        → PrefixQuery (fast)
path=**/tasks/*.md     → WildcardQuery (slower)
```
→ Optimized for prefix patterns

**Title Field** (1):
```go
title=Meeting → MatchQuery on title field
```
→ Direct mapping

### What Doesn't Work (2/13 fields)

**Link Queries** (2):
```go
links-to=docs/*.md     → ERROR: Phase 5.3 required
linked-by=plan.md      → ERROR: Phase 5.3 required
```
→ Requires graph index (separate phase)

---

## Migration Strategy

### New Method: SearchService.BuildQuery()

**Purpose**: Convert QueryCondition structs to search.Query AST

**Location**: `internal/services/search.go`

**Signature**:
```go
func (s *SearchService) BuildQuery(conditions []QueryCondition) (*search.Query, error)
```

**Logic**:
1. Group conditions by type (and/or/not)
2. Convert each to search.Expr
3. Build expression tree
4. Return search.Query AST

### Updated Method: NoteService.SearchWithConditions()

**Before**:
```go
db, err := s.dbService.GetDB(ctx)
whereClause, params := s.searchService.BuildWhereClauseWithGlob(...)
rows, err := db.QueryContext(query, params...)
// Parse rows into Notes (150+ lines)
```

**After**:
```go
query, err := s.searchService.BuildQuery(conditions)
results, err := s.index.Find(ctx, search.FindOpts{Query: query})
notes := documentToNote(results.Items) // Reuse from Phase 5.2.2
```

**Reduction**: 200+ lines → 20 lines

---

## Implementation Phases

| Phase | Task | Time | Files |
|-------|------|------|-------|
| 1 | Implement BuildQuery() | 2-3h | search.go |
| 2 | Update SearchWithConditions() | 1h | note.go |
| 3 | Update Tests | 3-4h | *_test.go |
| 4 | Documentation | 1-2h | CHANGELOG, docs |
| 5 | Integration | 1h | All |
| **Total** | | **8-11h** | |

---

## Breaking Changes

### Link Queries Return Error

**Before** (Phase 5.2.2):
```bash
opennotes notes search query --and links-to=docs/*.md
# Returns: Notes linking to docs/*.md
```

**After** (Phase 5.2.3):
```bash
opennotes notes search query --and links-to=docs/*.md
# ERROR: link queries are not yet supported
#
# Field 'links-to' requires a dedicated link graph index,
# which is planned for Phase 5.3.
#
# Temporary workaround: Use SQL query interface
#   opennotes notes query "SELECT * FROM ..."
#
# Track progress: github.com/zenobi-us/opennotes/issues/XXX
```

### Migration Path

1. **Short-term**: Use SQL interface for link queries
2. **Phase 5.3**: Link graph index implemented
3. **Phase 5.4**: Full feature parity restored

---

## Testing Plan

### Unit Tests (15 new)

**File**: `internal/services/search_test.go`

```go
TestSearchService_BuildQuery_SingleTag           // ✅ Basic
TestSearchService_BuildQuery_MultipleAnd         // ✅ AND logic
TestSearchService_BuildQuery_MultipleOr          // ✅ OR logic
TestSearchService_BuildQuery_Not                 // ✅ NOT logic
TestSearchService_BuildQuery_PathPrefix          // ✅ Path prefix
TestSearchService_BuildQuery_PathWildcard        // ✅ Path wildcard
TestSearchService_BuildQuery_EmptyConditions     // ✅ Edge case
TestSearchService_BuildQuery_LinksToError        // ❌ Error expected
TestSearchService_BuildQuery_LinkedByError       // ❌ Error expected
TestSearchService_BuildQuery_UnknownField        // ❌ Error expected
TestSearchService_BuildQuery_MixedConditions     // ✅ Complex
... (5 more)
```

### Integration Tests (40 updated)

**File**: `internal/services/note_test.go`

**Pattern**:
```go
// Before: DuckDB setup
db := services.NewDbService()
noteService := services.NewNoteService(cfg, db, nil, notebookDir)

// After: Bleve setup
index := testutil.CreateTestIndex(t, notebookDir)
noteService := services.NewNoteService(cfg, nil, index, notebookDir)
```

**Expected**: 171/171 tests passing

---

## Risk Assessment

### 🔴 High Risk: Link Query Breaking Change

**Impact**: Users relying on link queries will break

**Mitigation**:
- Clear error message with workaround
- Document in CHANGELOG.md
- Create Phase 5.3 issue
- Provide SQL fallback example

### ⚠️ Medium Risk: Path Glob Performance

**Impact**: Complex globs may be slower

**Mitigation**:
- Optimize simple patterns (`projects/*` → prefix)
- Document performance characteristics
- Recommend prefix patterns in docs

### 🟡 Low Risk: Test Migration

**Impact**: 40 tests need updating

**Mitigation**:
- Update incrementally
- Run after each change
- Use testutil.CreateTestIndex() consistently

---

## Success Criteria

✅ **Functionality**:
- All metadata queries work
- Path queries work (prefix optimized)
- Title queries work
- AND/OR/NOT logic correct
- Link queries error gracefully

✅ **Testing**:
- 15+ BuildQuery() unit tests
- 171/171 total tests passing
- Link tests appropriately handled
- Manual CLI testing successful

✅ **Documentation**:
- CHANGELOG.md updated
- docs/ updated
- Error messages clear
- Phase 5.3 referenced

✅ **Quality**:
- No performance regressions
- Code follows patterns
- Memory files complete

---

## Next Steps

### Immediate (Phase 5.2.3)

1. ✅ Read assessment and plan documents
2. ⏳ Implement BuildQuery() method
3. ⏳ Update SearchWithConditions()
4. ⏳ Update tests
5. ⏳ Update documentation
6. ⏳ Verify and commit

### Future Phases

**Phase 5.2.4**: Migrate Count()
- Simple migration (similar to getAllNotes)
- Use `index.Count(ctx, FindOpts{})`

**Phase 5.2.5**: Remove SQL Methods
- Remove ExecuteSQLSafe()
- Remove Query()
- Remove DuckDB from NoteService

**Phase 5.3**: Link Graph Index (NEW)
- Design link graph structure
- Implement links-to and linked-by
- Full feature parity

---

## Quick Reference

### Commands

```bash
# Build & test
mise run build
mise run test
mise run test -- SearchService
mise run test -- NoteService

# Manual testing
./dist/opennotes init
./dist/opennotes notes search query --and data.tag=work
./dist/opennotes notes search query --and links-to=docs/*.md  # Should error
```

### Key Files

**Implementation**:
- `internal/services/search.go` - BuildQuery()
- `internal/services/note.go` - SearchWithConditions()

**Tests**:
- `internal/services/search_test.go`
- `internal/services/note_test.go`

**Documentation**:
- `CHANGELOG.md`
- `docs/commands/notes-search.md`
- `.memory/assessment-phase523-migration.md`
- `.memory/plan-phase523-implementation.md`

---

## Recommendation

**PROCEED** ✅

Migration is technically sound with clear path forward:
1. Core functionality migrates cleanly
2. Link queries deferred to dedicated phase
3. Breaking changes documented
4. Workarounds provided
5. No blockers

**Start with**: Phase 1 - Implement BuildQuery() method

See `.memory/plan-phase523-implementation.md` for detailed step-by-step guide.
