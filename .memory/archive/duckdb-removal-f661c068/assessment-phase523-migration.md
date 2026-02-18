---
id: assessment-523
title: Phase 5.2.3 Migration Feasibility Assessment
created_at: 2026-02-02T07:54:00+10:30
status: complete
phase_id: 02df510c
task_id: 5d8f7e3a
---

# Phase 5.2.3 - SearchWithConditions() Migration Assessment

## Executive Summary

**VERDICT**: ✅ **PROCEED WITH CAUTION** - Migration is technically feasible but with significant limitations.

**Key Finding**: `links-to` and `linked-by` queries CANNOT be directly migrated to Bleve. These require a separate graph/link index implementation (deferred to future phase).

**Recommendation**: Migrate core functionality now, return clear error for link queries with TODO pointing to Phase 5.3 (Link Graph Implementation).

---

## Current Implementation Analysis

### QueryCondition Structure

```go
type QueryCondition struct {
    Type     string // "and", "or", "not"
    Field    string // "data.tag", "path", "title", "links-to", "linked-by"
    Operator string // "=" (only equality currently)
    Value    string // user value
}
```

### Supported Fields (11 total)

**Metadata Fields** (10):
- `data.tag` / `data.tags` - Note tags
- `data.status` - Status field
- `data.priority` - Priority level
- `data.assignee` - Assigned person
- `data.author` - Note author
- `data.type` - Note type
- `data.category` - Category
- `data.project` - Project name
- `data.sprint` - Sprint identifier

**Path/Title Fields** (2):
- `path` - File path with glob support (`projects/*.md`, `**/*.md`)
- `title` - Note title

**Link Fields** (2) - ⚠️ NOT MIGRATABLE:
- `links-to` - Find notes linking TO target (requires link index)
- `linked-by` - Find notes linked FROM source (requires link index)

### Current SQL Implementation

```sql
-- Example generated SQL
SELECT * FROM read_markdown(?, include_filepath:=true)
WHERE 
  COALESCE(metadata['tag'], '') = ?
  AND file_path LIKE ?
  AND NOT (COALESCE(metadata['status'], '') = ?)
ORDER BY file_path
```

Key features:
- Parameterized queries (secure)
- Glob-to-LIKE conversion for paths
- Metadata map access via `metadata['field']`
- AND/OR/NOT boolean logic
- Link queries via DuckDB subqueries (complex!)

---

## Migration Feasibility Matrix

| QueryCondition Field | Bleve Equivalent | Difficulty | Status |
|---------------------|------------------|------------|--------|
| `data.tag` | Tag field term query | ✅ Easy | Supported |
| `data.status` | Metadata field match | ✅ Easy | Supported |
| `data.priority` | Metadata field match | ✅ Easy | Supported |
| `data.assignee` | Metadata field match | ✅ Easy | Supported |
| `data.author` | Metadata field match | ✅ Easy | Supported |
| `data.type` | Metadata field match | ✅ Easy | Supported |
| `data.category` | Metadata field match | ✅ Easy | Supported |
| `data.project` | Metadata field match | ✅ Easy | Supported |
| `data.sprint` | Metadata field match | ✅ Easy | Supported |
| `path` | Path prefix query | ⚠️ Medium | Glob→Prefix conversion needed |
| `title` | Title match query | ✅ Easy | Supported |
| `links-to` | **NOT POSSIBLE** | ❌ Hard | Requires link index |
| `linked-by` | **NOT POSSIBLE** | ❌ Hard | Requires link index |

### Mapping Details

#### ✅ Metadata Fields (data.*)

**SQL**: `COALESCE(metadata['tag'], '') = ?`  
**Bleve**: `MatchQuery("work").SetField("metadata.tag")`

**Implementation**: Straightforward - Bleve's document model already includes `Metadata map[string]any`.

**Note**: Need to handle metadata field normalization:
- `data.tag` → `metadata.tag`
- `data.status` → `metadata.status`
- etc.

#### ✅ Title Field

**SQL**: `(COALESCE(metadata['title'], '') = ? OR file_path LIKE ?)`  
**Bleve**: `MatchQuery("Meeting").SetField("title")`

**Implementation**: Direct mapping to Bleve's title field.

#### ⚠️ Path Field (Medium Complexity)

**SQL**: `file_path LIKE '%projects/%'` (glob converted to LIKE)  
**Bleve**: `PrefixQuery("projects/").SetField("path")`

**Challenge**: SQL supports full glob patterns (`*`, `**`, `?`), Bleve only supports prefix queries efficiently.

**Solution Strategy**:
1. **Exact paths**: `path=projects/foo.md` → `TermQuery("projects/foo.md")`
2. **Prefix globs**: `path=projects/*` → `PrefixQuery("projects/")`
3. **Double-star globs**: `path=**/tasks/*.md` → Convert to `WildcardQuery("*/tasks/*.md")`
4. **Mid-path wildcards**: `path=proj*/foo.md` → `WildcardQuery("proj*/foo.md")` (slower)

**Recommendation**: Support prefixes primarily, fallback to wildcard queries with performance warning.

#### ❌ Link Queries (NOT MIGRATABLE)

**Current SQL** (links-to):
```sql
EXISTS (
  SELECT 1 FROM (
    SELECT unnest(COALESCE(TRY_CAST(metadata['links'] AS VARCHAR[]), ARRAY[]::VARCHAR[])) AS link
  ) AS links_table
  WHERE link LIKE ?
)
```

**Current SQL** (linked-by):
```sql
EXISTS (
  SELECT 1 FROM (
    SELECT unnest(COALESCE(TRY_CAST(src.metadata['links'] AS VARCHAR[]), ARRAY[]::VARCHAR[])) AS link
    FROM read_markdown(?, include_filepath:=true) AS src
    WHERE src.file_path LIKE ?
  ) AS source_links
  WHERE file_path LIKE '%' || source_links.link OR file_path LIKE '%/' || source_links.link
)
```

**Why Bleve Can't Handle This**:
1. Bleve indexes individual documents independently
2. Link queries require **joining across documents** (graph traversal)
3. `linked-by` requires reading OTHER documents to check their links
4. Bleve has no JOIN or subquery capability

**Workaround Options**:
1. **Fallback to SQL** (defeats purpose of migration)
2. **Post-filter in memory** (inefficient for large notebooks)
3. **Build separate link index** (correct solution, different phase)

**Chosen Solution**: Return clear error message with reference to Phase 5.3.

---

## Query Condition → Bleve Query Mapping

### Boolean Logic Mapping

| QueryCondition.Type | Bleve Query Type | Notes |
|---------------------|------------------|-------|
| `and` (multiple) | `ConjunctionQuery(queries)` | All must match |
| `or` (multiple) | `DisjunctionQuery(queries)` | Any can match |
| `not` | `BooleanQuery(must, nil, mustNot)` | Negation |
| Mixed AND/OR/NOT | Nested `BooleanQuery` | Complex structure |

### Example Translations

**Simple AND**:
```go
// Input: --and data.tag=workflow --and data.status=active
conditions := []QueryCondition{
    {Type: "and", Field: "data.tag", Operator: "=", Value: "workflow"},
    {Type: "and", Field: "data.status", Operator: "=", Value: "active"},
}

// Bleve Output:
ConjunctionQuery([
    MatchQuery("workflow").SetField("metadata.tag"),
    MatchQuery("active").SetField("metadata.status"),
])
```

**OR Conditions**:
```go
// Input: --or data.priority=high --or data.priority=critical
conditions := []QueryCondition{
    {Type: "or", Field: "data.priority", Operator: "=", Value: "high"},
    {Type: "or", Field: "data.priority", Operator: "=", Value: "critical"},
}

// Bleve Output:
DisjunctionQuery([
    MatchQuery("high").SetField("metadata.priority"),
    MatchQuery("critical").SetField("metadata.priority"),
])
```

**NOT Conditions**:
```go
// Input: --not data.status=archived
conditions := []QueryCondition{
    {Type: "not", Field: "data.status", Operator: "=", Value: "archived"},
}

// Bleve Output:
BooleanQuery(
    must: [MatchAllQuery()],
    should: nil,
    mustNot: [MatchQuery("archived").SetField("metadata.status")],
)
```

**Mixed AND/OR/NOT**:
```go
// Input: --and data.tag=workflow --or data.priority=high --not data.status=done
conditions := []QueryCondition{
    {Type: "and", Field: "data.tag", Operator: "=", Value: "workflow"},
    {Type: "or", Field: "data.priority", Operator: "=", Value: "high"},
    {Type: "not", Field: "data.status", Operator: "=", Value: "done"},
}

// Bleve Output (precedence: AND > OR > NOT):
BooleanQuery(
    must: [
        ConjunctionQuery([MatchQuery("workflow").SetField("metadata.tag")]),
        DisjunctionQuery([MatchQuery("high").SetField("metadata.priority")]),
    ],
    should: nil,
    mustNot: [MatchQuery("done").SetField("metadata.status")],
)
```

---

## Implementation Architecture

### Design Decision: Add BuildQuery() to SearchService

**Location**: `internal/services/search.go`

**Rationale**:
- Mirrors current `BuildWhereClauseWithGlob()` pattern
- Keeps query building logic separate from NoteService
- Testable independently
- Reusable if needed elsewhere

**Method Signature**:
```go
func (s *SearchService) BuildQuery(conditions []QueryCondition) (*search.Query, error)
```

**Returns**:
- `*search.Query` - AST that can be translated by Bleve's `TranslateQuery()`
- `error` - Validation errors, unsupported fields, link queries

### Updated NoteService.SearchWithConditions()

**Before** (DuckDB):
```go
func (s *NoteService) SearchWithConditions(ctx context.Context, conditions []QueryCondition) ([]Note, error) {
    db, err := s.dbService.GetDB(ctx)
    glob := filepath.Join(s.notebookPath, "**", "*.md")
    whereClause, params, err := s.searchService.BuildWhereClauseWithGlob(conditions, glob)
    query := `SELECT * FROM read_markdown(?, include_filepath:=true) WHERE ` + whereClause + ` ORDER BY file_path`
    rows, err := db.QueryContext(timeoutCtx, query, allParams...)
    // ... parse rows into Notes
}
```

**After** (Bleve):
```go
func (s *NoteService) SearchWithConditions(ctx context.Context, conditions []QueryCondition) ([]Note, error) {
    // Build search.Query from conditions
    query, err := s.searchService.BuildQuery(conditions)
    if err != nil {
        return nil, err
    }
    
    // Execute search using Index
    results, err := s.index.Find(ctx, search.FindOpts{
        Query: query,
        Sort: search.SortSpec{
            Field: search.SortByPath,
            Direction: search.SortAsc,
        },
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

**Changes**:
1. Replace SQL building with `BuildQuery()`
2. Replace `db.QueryContext()` with `index.Find()`
3. Replace row parsing with `documentToNote()` (already exists from Phase 5.2.2)
4. Maintain `ORDER BY file_path` via `SortByPath`

---

## Risk Assessment

### 🔴 HIGH RISK: Link Queries Breaking Change

**Issue**: `links-to` and `linked-by` queries will fail after migration.

**Impact**: Users relying on link queries in scripts/workflows will break.

**Mitigation**:
1. Return **clear, actionable error message**:
   ```
   Error: link queries (links-to, linked-by) are not yet supported in the new search system.
   
   These queries require a dedicated link graph index, which is planned for Phase 5.3.
   
   Temporary workaround: Use the SQL query interface:
     opennotes notes query "SELECT * FROM read_markdown('**/*.md') WHERE ..."
   
   Track progress: https://github.com/zenobi-us/opennotes/issues/XXX
   ```

2. **Document breaking change** in CHANGELOG.md
3. **Add deprecation warning** in Phase 5.2.3 release notes
4. **Create GitHub issue** for Phase 5.3 (Link Index Implementation)

**Timeline**:
- Phase 5.2.3: Basic queries only (no links)
- Phase 5.3: Link graph index (separate implementation)
- Phase 5.4: Full feature parity

### ⚠️ MEDIUM RISK: Path Glob Behavior Changes

**Issue**: SQL LIKE patterns vs Bleve wildcard queries have different performance characteristics.

**Current**: Glob patterns converted to LIKE (fast in DuckDB due to B-tree indexes).

**After**: Glob patterns converted to WildcardQuery (slower in Bleve, full scan).

**Example**:
```bash
# Fast in SQL (LIKE '%projects/%')
opennotes notes search query --and path=projects/*

# Slower in Bleve (WildcardQuery requires full index scan)
opennotes notes search query --and path=**/*.md
```

**Mitigation**:
1. **Optimize for prefix patterns**: `path=projects/*` → `PrefixQuery("projects/")` (fast)
2. **Use wildcard for complex patterns**: `path=proj*/foo.md` → `WildcardQuery()` (slower)
3. **Add performance warning** in CLI help for complex globs
4. **Recommend path prefixes** in documentation

### ⚠️ MEDIUM RISK: Operator Limitation

**Issue**: Currently only `Operator: "="` is supported, but SQL implementation could theoretically support `>`, `<`, etc.

**Current State**: Only equality is implemented and documented.

**Risk**: Low - no users depend on this since it's not exposed in CLI.

**Mitigation**: Document that only equality is supported (no change from current behavior).

### 🟡 LOW RISK: Metadata Field Access Pattern

**Issue**: SQL uses `metadata['field']`, Bleve uses `metadata.field`.

**Impact**: None - this is internal implementation detail.

**Mitigation**: Handle normalization in `BuildQuery()`.

### 🟡 LOW RISK: Result Ordering

**Issue**: SQL uses `ORDER BY file_path`, Bleve uses `SortByPath`.

**Impact**: None - both sort by path ascending.

**Mitigation**: Explicitly set `Sort` in `FindOpts`.

---

## Testing Strategy

### Unit Tests for BuildQuery()

**File**: `internal/services/search_test.go`

**Test Cases** (15 tests):
```go
// Basic field mapping
func TestSearchService_BuildQuery_SingleTag
func TestSearchService_BuildQuery_SingleStatus
func TestSearchService_BuildQuery_SinglePath
func TestSearchService_BuildQuery_SingleTitle

// Boolean logic
func TestSearchService_BuildQuery_MultipleAnd
func TestSearchService_BuildQuery_MultipleOr
func TestSearchService_BuildQuery_MultipleNot
func TestSearchService_BuildQuery_MixedAndOrNot

// Path globbing
func TestSearchService_BuildQuery_PathPrefix
func TestSearchService_BuildQuery_PathWildcard
func TestSearchService_BuildQuery_PathDoublestar

// Edge cases
func TestSearchService_BuildQuery_EmptyConditions
func TestSearchService_BuildQuery_UnknownFieldError

// Link queries (should error)
func TestSearchService_BuildQuery_LinksToError
func TestSearchService_BuildQuery_LinkedByError
```

### Integration Tests for SearchWithConditions()

**File**: `internal/services/note_test.go`

**Existing Tests to Update** (~40 tests):
```
TestNoteService_SearchWithConditions_SimpleAnd
TestNoteService_SearchWithConditions_MultipleAnd
TestNoteService_SearchWithConditions_Or
TestNoteService_SearchWithConditions_Not
TestNoteService_SearchWithConditions_MixedConditions
TestNoteService_SearchWithConditions_PathGlob
TestNoteService_SearchWithConditions_EmptyResults
... (30+ more tests)
```

**Update Strategy**:
1. Replace DuckDB setup with `testutil.CreateTestIndex()`
2. Keep test data and assertions identical
3. Add new tests for glob pattern edge cases
4. Skip link query tests (mark as TODO for Phase 5.3)

### Manual CLI Testing

**Test Commands**:
```bash
# Basic queries
mise run build
./dist/opennotes notes search query --and data.tag=work
./dist/opennotes notes search query --and data.tag=work --and data.status=active

# OR conditions
./dist/opennotes notes search query --or data.priority=high --or data.priority=critical

# NOT conditions
./dist/opennotes notes search query --and data.tag=work --not data.status=done

# Path queries
./dist/opennotes notes search query --and path=projects/*
./dist/opennotes notes search query --and path=**/*.md

# Link queries (should error gracefully)
./dist/opennotes notes search query --and links-to=docs/*.md
./dist/opennotes notes search query --and linked-by=plan.md
```

**Expected Outcomes**:
- Basic queries return correct results
- Link queries return clear error with next steps
- Performance is comparable or better than SQL

---

## Implementation Plan

### Phase 1: Implement BuildQuery() Method

**File**: `internal/services/search.go`

**Steps**:
1. Add `BuildQuery(conditions []QueryCondition) (*search.Query, error)` method
2. Implement field → expression mapping:
   - `data.*` → metadata field expressions
   - `path` → path prefix/wildcard expressions
   - `title` → title match expressions
3. Implement boolean logic:
   - Group by condition type (and/or/not)
   - Build nested expression tree
4. Handle link queries:
   - Detect `links-to` and `linked-by`
   - Return clear error with Phase 5.3 reference
5. Add unit tests (15 tests)

**Time Estimate**: 2-3 hours

### Phase 2: Update SearchWithConditions()

**File**: `internal/services/note.go`

**Steps**:
1. Replace SQL building with `BuildQuery()` call
2. Replace `dbService.GetDB()` with `index.Find()`
3. Replace row parsing with `documentToNote()` converter
4. Add sorting to `FindOpts` (maintain ORDER BY file_path)
5. Remove DuckDB dependencies from method

**Time Estimate**: 1 hour

### Phase 3: Update Tests

**Files**: `internal/services/note_test.go`, `internal/services/search_test.go`

**Steps**:
1. Add BuildQuery() unit tests to `search_test.go`
2. Update SearchWithConditions() tests in `note_test.go`:
   - Replace DuckDB setup with `testutil.CreateTestIndex()`
   - Update to use Bleve Index
   - Skip/mark link query tests as TODO
3. Run full test suite: `mise run test`
4. Fix any failing tests

**Time Estimate**: 3-4 hours

### Phase 4: Manual Testing & Documentation

**Steps**:
1. Build binary: `mise run build`
2. Test CLI commands manually (see testing section above)
3. Update CHANGELOG.md with breaking changes
4. Update docs/commands/notes-search.md:
   - Mark link queries as "coming in Phase 5.3"
   - Add performance notes for path globs
5. Create GitHub issue for Phase 5.3 (Link Index)

**Time Estimate**: 1-2 hours

### Phase 5: Integration & Verification

**Steps**:
1. Run full test suite: `mise run test`
2. Verify 171/172 tests still passing (1 link query test should fail)
3. Update failing link test to expect error
4. Commit with semantic message: `feat(search): migrate SearchWithConditions to Bleve (BREAKING: link queries deferred)`
5. Mark Phase 5.2.3 task as complete

**Time Estimate**: 1 hour

**Total Estimated Time**: 8-11 hours

---

## Breaking Changes & Migration Guide

### Breaking Change: Link Queries Not Supported

**Affected Commands**:
```bash
# These will return errors after Phase 5.2.3
opennotes notes search query --and links-to=docs/*.md
opennotes notes search query --and linked-by=plan.md
```

**Error Message**:
```
Error: link queries are not yet supported

Field 'links-to' requires a dedicated link graph index, which is planned for Phase 5.3.

Temporary workaround: Use the SQL query interface:
  opennotes notes query "SELECT * FROM read_markdown('**/*.md') 
    WHERE EXISTS (
      SELECT 1 FROM (
        SELECT unnest(COALESCE(TRY_CAST(metadata['links'] AS VARCHAR[]), ARRAY[]::VARCHAR[])) AS link
      ) WHERE link LIKE '%docs/%'
    )"

Track implementation progress:
  https://github.com/zenobi-us/opennotes/issues/XXX

Supported fields: data.tag, data.status, data.priority, data.assignee, 
                  data.author, data.type, data.category, data.project, 
                  data.sprint, path, title
```

### Workaround for Link Queries

**Option 1**: Use SQL query interface (temporary)
```bash
opennotes notes query "SELECT * FROM read_markdown('**/*.md') WHERE ..."
```

**Option 2**: Wait for Phase 5.3 (recommended)
- Link graph index implementation
- Full feature parity with SQL
- Better performance for link queries

---

## Success Criteria

✅ **Core Functionality**:
- [ ] All metadata field queries work (data.tag, data.status, etc.)
- [ ] Path queries work with prefix matching
- [ ] Title queries work
- [ ] AND/OR/NOT boolean logic works
- [ ] Results sorted by path

✅ **Testing**:
- [ ] 15+ new unit tests for BuildQuery()
- [ ] All existing SearchWithConditions() tests updated
- [ ] 171/172 tests passing (1 link test expects error)
- [ ] Manual CLI testing successful

✅ **Documentation**:
- [ ] CHANGELOG.md updated with breaking changes
- [ ] docs/commands/notes-search.md updated
- [ ] GitHub issue created for Phase 5.3
- [ ] Error messages guide users to workarounds

✅ **Quality**:
- [ ] No performance regressions for non-link queries
- [ ] Clear error messages for unsupported queries
- [ ] Code follows existing patterns (BuildQuery mirrors BuildWhereClause)

---

## Next Steps After Phase 5.2.3

### Phase 5.2.4: Migrate Count()
- Simple migration (same pattern as getAllNotes)
- Use `index.Count(ctx, FindOpts{})`

### Phase 5.2.5: Remove SQL Methods
- Remove `ExecuteSQLSafe()`
- Remove `Query()`
- Remove DuckDB dependency from NoteService

### Phase 5.3: Link Graph Index (NEW)
- Design link index structure
- Implement `links-to` and `linked-by` queries
- Full feature parity achieved

---

## Appendix: Code Examples

### BuildQuery() Implementation Sketch

```go
func (s *SearchService) BuildQuery(conditions []QueryCondition) (*search.Query, error) {
    if len(conditions) == 0 {
        return &search.Query{}, nil
    }
    
    var andExprs []search.Expr
    var orExprs []search.Expr
    var notExprs []search.Expr
    
    for _, cond := range conditions {
        // Check for unsupported fields
        if cond.Field == "links-to" || cond.Field == "linked-by" {
            return nil, fmt.Errorf(
                "link queries are not yet supported\n\n" +
                "Field '%s' requires a dedicated link graph index, which is planned for Phase 5.3.\n\n" +
                "Temporary workaround: Use the SQL query interface:\n" +
                "  opennotes notes query \"SELECT * FROM read_markdown('**/*.md') WHERE ...\"\n\n" +
                "Supported fields: data.tag, data.status, data.priority, data.assignee, " +
                "data.author, data.type, data.category, data.project, data.sprint, path, title",
                cond.Field,
            )
        }
        
        expr, err := s.conditionToExpr(cond)
        if err != nil {
            return nil, err
        }
        
        switch cond.Type {
        case "and":
            andExprs = append(andExprs, expr)
        case "or":
            orExprs = append(orExprs, expr)
        case "not":
            notExprs = append(notExprs, search.NotExpr{Expr: expr})
        }
    }
    
    // Build final expression tree
    var allExprs []search.Expr
    
    // Add AND expressions directly
    allExprs = append(allExprs, andExprs...)
    
    // Group OR expressions
    if len(orExprs) > 0 {
        if len(orExprs) == 1 {
            allExprs = append(allExprs, orExprs[0])
        } else {
            // Build nested OR tree
            orExpr := orExprs[0]
            for i := 1; i < len(orExprs); i++ {
                orExpr = search.OrExpr{Left: orExpr, Right: orExprs[i]}
            }
            allExprs = append(allExprs, orExpr)
        }
    }
    
    // Add NOT expressions
    allExprs = append(allExprs, notExprs...)
    
    return &search.Query{Expressions: allExprs}, nil
}

func (s *SearchService) conditionToExpr(cond QueryCondition) (search.Expr, error) {
    switch {
    case strings.HasPrefix(cond.Field, "data."):
        // Metadata field
        field := strings.TrimPrefix(cond.Field, "data.")
        return search.FieldExpr{
            Field: "metadata." + field,
            Op:    search.OpEquals,
            Value: cond.Value,
        }, nil
        
    case cond.Field == "path":
        // Path field - handle globs
        if strings.Contains(cond.Value, "*") {
            return search.WildcardExpr{
                Field:   "path",
                Pattern: cond.Value,
                Type:    detectWildcardType(cond.Value),
            }, nil
        }
        return search.FieldExpr{
            Field: "path",
            Op:    search.OpPrefix,
            Value: cond.Value,
        }, nil
        
    case cond.Field == "title":
        return search.FieldExpr{
            Field: "title",
            Op:    search.OpEquals,
            Value: cond.Value,
        }, nil
        
    default:
        return nil, fmt.Errorf("unsupported field: %s", cond.Field)
    }
}
```

---

## Conclusion

Migration of `SearchWithConditions()` is **feasible with known limitations**. Core functionality (metadata, path, title queries) can be migrated successfully. Link queries require a separate implementation phase.

**Recommendation**: Proceed with migration, documenting breaking changes clearly and providing migration path for affected users.
