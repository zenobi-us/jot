---
id: plan-523-impl
title: Phase 5.2.3 Implementation Plan - SearchWithConditions Migration
created_at: 2026-02-02T07:54:00+10:30
status: ready
phase_id: 02df510c
task_id: 5d8f7e3a
---

# Phase 5.2.3 Implementation Plan

## Overview

Migrate `NoteService.SearchWithConditions()` from DuckDB SQL to Bleve Index.

**Status**: Ready for implementation  
**Assessment**: `.memory/assessment-phase523-migration.md`  
**Estimated Time**: 8-11 hours

---

## Implementation Phases

### Phase 1: Implement BuildQuery() Method

**File**: `internal/services/search.go`  
**Time**: 2-3 hours

#### Step 1.1: Add Method Signature

```go
// BuildQuery converts QueryCondition structs to search.Query AST.
// This mirrors the pattern of BuildWhereClauseWithGlob but outputs
// a search.Query instead of SQL.
//
// Supported fields:
//   - data.* (metadata fields: tag, status, priority, etc.)
//   - path (with glob pattern support)
//   - title
//
// Unsupported fields (return error):
//   - links-to (requires Phase 5.3 link graph index)
//   - linked-by (requires Phase 5.3 link graph index)
//
// Boolean logic:
//   - AND conditions: All must match (ConjunctionQuery)
//   - OR conditions: Any can match (DisjunctionQuery)
//   - NOT conditions: Must not match (BooleanQuery with mustNot)
//
// Example:
//   conditions := []QueryCondition{
//       {Type: "and", Field: "data.tag", Operator: "=", Value: "work"},
//       {Type: "and", Field: "data.status", Operator: "=", Value: "active"},
//   }
//   query, err := searchService.BuildQuery(conditions)
//   // query.Expressions = [FieldExpr{...}, FieldExpr{...}]
func (s *SearchService) BuildQuery(conditions []QueryCondition) (*search.Query, error) {
    // Implementation
}
```

#### Step 1.2: Implement Core Logic

```go
func (s *SearchService) BuildQuery(conditions []QueryCondition) (*search.Query, error) {
    if len(conditions) == 0 {
        return &search.Query{}, nil
    }
    
    var andExprs []search.Expr
    var orExprs []search.Expr
    var notExprs []search.Expr
    
    // Convert each condition to an expression
    for _, cond := range conditions {
        // Check for unsupported link queries
        if cond.Field == "links-to" || cond.Field == "linked-by" {
            return nil, s.buildLinkQueryError(cond.Field)
        }
        
        // Convert to expression
        expr, err := s.conditionToExpr(cond)
        if err != nil {
            return nil, err
        }
        
        // Group by type
        switch cond.Type {
        case "and":
            andExprs = append(andExprs, expr)
        case "or":
            orExprs = append(orExprs, expr)
        case "not":
            notExprs = append(notExprs, search.NotExpr{Expr: expr})
        default:
            return nil, fmt.Errorf("unsupported condition type: %s", cond.Type)
        }
    }
    
    // Build final expression tree
    var allExprs []search.Expr
    
    // Add AND expressions directly
    allExprs = append(allExprs, andExprs...)
    
    // Group OR expressions into nested OrExpr
    if len(orExprs) > 0 {
        if len(orExprs) == 1 {
            allExprs = append(allExprs, orExprs[0])
        } else {
            // Build left-to-right OR tree: (a OR (b OR c))
            orExpr := orExprs[0]
            for i := 1; i < len(orExprs); i++ {
                orExpr = search.OrExpr{Left: orExpr, Right: orExprs[i]}
            }
            allExprs = append(allExprs, orExpr)
        }
    }
    
    // Add NOT expressions
    allExprs = append(allExprs, notExprs...)
    
    s.log.Debug().
        Int("and_count", len(andExprs)).
        Int("or_count", len(orExprs)).
        Int("not_count", len(notExprs)).
        Int("total_exprs", len(allExprs)).
        Msg("built query from conditions")
    
    return &search.Query{Expressions: allExprs}, nil
}
```

#### Step 1.3: Implement conditionToExpr()

```go
// conditionToExpr converts a single QueryCondition to a search.Expr.
func (s *SearchService) conditionToExpr(cond QueryCondition) (search.Expr, error) {
    switch {
    case strings.HasPrefix(cond.Field, "data."):
        // Metadata field: data.tag -> metadata.tag
        return s.buildMetadataExpr(cond)
        
    case cond.Field == "path":
        // Path field with glob support
        return s.buildPathExpr(cond)
        
    case cond.Field == "title":
        // Title field
        return search.FieldExpr{
            Field: "title",
            Op:    search.OpEquals,
            Value: cond.Value,
        }, nil
        
    default:
        return nil, fmt.Errorf("unsupported field: %s (allowed: data.*, path, title)", cond.Field)
    }
}

// buildMetadataExpr builds expression for metadata fields (data.*).
func (s *SearchService) buildMetadataExpr(cond QueryCondition) (search.Expr, error) {
    // Extract metadata field name: data.tag -> tag
    field := strings.TrimPrefix(cond.Field, "data.")
    
    // Normalize field name
    // data.tags -> data.tag (alias)
    if field == "tags" {
        field = "tag"
    }
    
    // Build field expression
    // Note: Bleve document has Metadata map[string]any
    // We need to search in metadata.field
    return search.FieldExpr{
        Field: "metadata." + field,
        Op:    search.OpEquals,
        Value: cond.Value,
    }, nil
}

// buildPathExpr builds expression for path field with glob support.
func (s *SearchService) buildPathExpr(cond QueryCondition) (search.Expr, error) {
    value := cond.Value
    
    // Detect glob patterns
    hasWildcard := strings.Contains(value, "*") || strings.Contains(value, "?")
    
    if !hasWildcard {
        // Exact path or prefix
        // If ends with /, it's a prefix: projects/ -> projects/*
        if strings.HasSuffix(value, "/") {
            return search.FieldExpr{
                Field: "path",
                Op:    search.OpPrefix,
                Value: value,
            }, nil
        }
        
        // Exact path match
        return search.FieldExpr{
            Field: "path",
            Op:    search.OpEquals,
            Value: value,
        }, nil
    }
    
    // Has wildcards - determine type
    // Simple prefix: projects/* -> prefix query (fast)
    // Complex: **/tasks/*.md -> wildcard query (slower)
    
    if strings.HasSuffix(value, "/*") && !strings.Contains(strings.TrimSuffix(value, "/*"), "*") {
        // Simple prefix pattern: projects/* -> projects/
        prefix := strings.TrimSuffix(value, "*")
        return search.FieldExpr{
            Field: "path",
            Op:    search.OpPrefix,
            Value: prefix,
        }, nil
    }
    
    // Complex wildcard pattern
    wildcardType := detectWildcardType(value)
    return search.WildcardExpr{
        Field:   "path",
        Pattern: value,
        Type:    wildcardType,
    }, nil
}

// detectWildcardType determines wildcard expression type.
func detectWildcardType(pattern string) search.WildcardType {
    hasPrefix := strings.HasPrefix(pattern, "*")
    hasSuffix := strings.HasSuffix(pattern, "*")
    
    if hasPrefix && hasSuffix {
        return search.WildcardBoth
    } else if hasPrefix {
        return search.WildcardSuffix
    } else if hasSuffix {
        return search.WildcardPrefix
    }
    
    // Mid-pattern wildcard: foo*bar
    return search.WildcardBoth
}

// buildLinkQueryError returns a clear error for unsupported link queries.
func (s *SearchService) buildLinkQueryError(field string) error {
    return fmt.Errorf(
        "link queries are not yet supported\n\n" +
        "Field '%s' requires a dedicated link graph index, which is planned for Phase 5.3.\n\n" +
        "Temporary workaround: Use the SQL query interface:\n" +
        "  opennotes notes query \"SELECT * FROM read_markdown('**/*.md') WHERE ...\"\n\n" +
        "Track implementation progress:\n" +
        "  https://github.com/zenobi-us/opennotes/issues/XXX\n\n" +
        "Supported fields:\n" +
        "  - Metadata: data.tag, data.status, data.priority, data.assignee,\n" +
        "              data.author, data.type, data.category, data.project, data.sprint\n" +
        "  - Path: path (with glob support: *, **, ?)\n" +
        "  - Title: title",
        field,
    )
}
```

#### Step 1.4: Add Unit Tests

**File**: `internal/services/search_test.go`

```go
func TestSearchService_BuildQuery_SingleTag(t *testing.T) {
    searchSvc := services.NewSearchService()
    
    conditions := []services.QueryCondition{
        {Type: "and", Field: "data.tag", Operator: "=", Value: "work"},
    }
    
    query, err := searchSvc.BuildQuery(conditions)
    require.NoError(t, err)
    require.NotNil(t, query)
    require.Len(t, query.Expressions, 1)
    
    // Check expression type
    fieldExpr, ok := query.Expressions[0].(search.FieldExpr)
    require.True(t, ok, "expected FieldExpr")
    assert.Equal(t, "metadata.tag", fieldExpr.Field)
    assert.Equal(t, search.OpEquals, fieldExpr.Op)
    assert.Equal(t, "work", fieldExpr.Value)
}

func TestSearchService_BuildQuery_MultipleAnd(t *testing.T) {
    searchSvc := services.NewSearchService()
    
    conditions := []services.QueryCondition{
        {Type: "and", Field: "data.tag", Operator: "=", Value: "work"},
        {Type: "and", Field: "data.status", Operator: "=", Value: "active"},
    }
    
    query, err := searchSvc.BuildQuery(conditions)
    require.NoError(t, err)
    require.Len(t, query.Expressions, 2)
    
    // Both should be FieldExpr
    for _, expr := range query.Expressions {
        _, ok := expr.(search.FieldExpr)
        assert.True(t, ok, "expected FieldExpr for AND conditions")
    }
}

func TestSearchService_BuildQuery_MultipleOr(t *testing.T) {
    searchSvc := services.NewSearchService()
    
    conditions := []services.QueryCondition{
        {Type: "or", Field: "data.priority", Operator: "=", Value: "high"},
        {Type: "or", Field: "data.priority", Operator: "=", Value: "critical"},
    }
    
    query, err := searchSvc.BuildQuery(conditions)
    require.NoError(t, err)
    require.Len(t, query.Expressions, 1)
    
    // Should be nested OrExpr
    orExpr, ok := query.Expressions[0].(search.OrExpr)
    require.True(t, ok, "expected OrExpr for OR conditions")
    
    // Check left and right
    leftField, ok := orExpr.Left.(search.FieldExpr)
    assert.True(t, ok)
    assert.Equal(t, "high", leftField.Value)
    
    rightField, ok := orExpr.Right.(search.FieldExpr)
    assert.True(t, ok)
    assert.Equal(t, "critical", rightField.Value)
}

func TestSearchService_BuildQuery_Not(t *testing.T) {
    searchSvc := services.NewSearchService()
    
    conditions := []services.QueryCondition{
        {Type: "not", Field: "data.status", Operator: "=", Value: "archived"},
    }
    
    query, err := searchSvc.BuildQuery(conditions)
    require.NoError(t, err)
    require.Len(t, query.Expressions, 1)
    
    // Should be NotExpr
    notExpr, ok := query.Expressions[0].(search.NotExpr)
    require.True(t, ok, "expected NotExpr")
    
    // Inner should be FieldExpr
    fieldExpr, ok := notExpr.Expr.(search.FieldExpr)
    require.True(t, ok)
    assert.Equal(t, "metadata.status", fieldExpr.Field)
    assert.Equal(t, "archived", fieldExpr.Value)
}

func TestSearchService_BuildQuery_PathPrefix(t *testing.T) {
    searchSvc := services.NewSearchService()
    
    conditions := []services.QueryCondition{
        {Type: "and", Field: "path", Operator: "=", Value: "projects/*"},
    }
    
    query, err := searchSvc.BuildQuery(conditions)
    require.NoError(t, err)
    
    // Should be FieldExpr with OpPrefix
    fieldExpr, ok := query.Expressions[0].(search.FieldExpr)
    require.True(t, ok)
    assert.Equal(t, "path", fieldExpr.Field)
    assert.Equal(t, search.OpPrefix, fieldExpr.Op)
    assert.Equal(t, "projects/", fieldExpr.Value)
}

func TestSearchService_BuildQuery_PathWildcard(t *testing.T) {
    searchSvc := services.NewSearchService()
    
    conditions := []services.QueryCondition{
        {Type: "and", Field: "path", Operator: "=", Value: "**/tasks/*.md"},
    }
    
    query, err := searchSvc.BuildQuery(conditions)
    require.NoError(t, err)
    
    // Should be WildcardExpr
    wildcardExpr, ok := query.Expressions[0].(search.WildcardExpr)
    require.True(t, ok)
    assert.Equal(t, "path", wildcardExpr.Field)
    assert.Equal(t, "**/tasks/*.md", wildcardExpr.Pattern)
}

func TestSearchService_BuildQuery_EmptyConditions(t *testing.T) {
    searchSvc := services.NewSearchService()
    
    query, err := searchSvc.BuildQuery([]services.QueryCondition{})
    require.NoError(t, err)
    assert.NotNil(t, query)
    assert.Len(t, query.Expressions, 0)
}

func TestSearchService_BuildQuery_LinksToError(t *testing.T) {
    searchSvc := services.NewSearchService()
    
    conditions := []services.QueryCondition{
        {Type: "and", Field: "links-to", Operator: "=", Value: "docs/*.md"},
    }
    
    query, err := searchSvc.BuildQuery(conditions)
    assert.Error(t, err)
    assert.Nil(t, query)
    assert.Contains(t, err.Error(), "link queries are not yet supported")
    assert.Contains(t, err.Error(), "Phase 5.3")
}

func TestSearchService_BuildQuery_LinkedByError(t *testing.T) {
    searchSvc := services.NewSearchService()
    
    conditions := []services.QueryCondition{
        {Type: "and", Field: "linked-by", Operator: "=", Value: "plan.md"},
    }
    
    query, err := searchSvc.BuildQuery(conditions)
    assert.Error(t, err)
    assert.Nil(t, query)
    assert.Contains(t, err.Error(), "link queries are not yet supported")
}

func TestSearchService_BuildQuery_UnknownField(t *testing.T) {
    searchSvc := services.NewSearchService()
    
    conditions := []services.QueryCondition{
        {Type: "and", Field: "unknown-field", Operator: "=", Value: "value"},
    }
    
    query, err := searchSvc.BuildQuery(conditions)
    assert.Error(t, err)
    assert.Nil(t, query)
    assert.Contains(t, err.Error(), "unsupported field")
}

func TestSearchService_BuildQuery_MixedConditions(t *testing.T) {
    searchSvc := services.NewSearchService()
    
    conditions := []services.QueryCondition{
        {Type: "and", Field: "data.tag", Operator: "=", Value: "work"},
        {Type: "or", Field: "data.priority", Operator: "=", Value: "high"},
        {Type: "not", Field: "data.status", Operator: "=", Value: "done"},
    }
    
    query, err := searchSvc.BuildQuery(conditions)
    require.NoError(t, err)
    require.Len(t, query.Expressions, 3)
    
    // Should have FieldExpr (AND), OrExpr (OR), NotExpr (NOT)
    // Note: Single OR becomes FieldExpr directly
    _, hasField := query.Expressions[0].(search.FieldExpr)
    assert.True(t, hasField, "first should be FieldExpr (AND)")
    
    _, hasOr := query.Expressions[1].(search.FieldExpr)
    assert.True(t, hasOr, "second should be FieldExpr (single OR)")
    
    _, hasNot := query.Expressions[2].(search.NotExpr)
    assert.True(t, hasNot, "third should be NotExpr")
}
```

**Checklist**:
- [ ] Add BuildQuery() method to SearchService
- [ ] Implement conditionToExpr() helper
- [ ] Implement buildMetadataExpr() helper
- [ ] Implement buildPathExpr() helper
- [ ] Implement detectWildcardType() helper
- [ ] Implement buildLinkQueryError() helper
- [ ] Add 15 unit tests to search_test.go
- [ ] Run tests: `mise run test -- SearchService_BuildQuery`
- [ ] All tests passing

---

### Phase 2: Update SearchWithConditions()

**File**: `internal/services/note.go`  
**Time**: 1 hour

#### Step 2.1: Replace Implementation

```go
// SearchWithConditions executes a boolean query with the given conditions.
// Uses Bleve Index for querying instead of DuckDB SQL.
func (s *NoteService) SearchWithConditions(ctx context.Context, conditions []QueryCondition) ([]Note, error) {
    if s.notebookPath == "" {
        return nil, fmt.Errorf("no notebook selected")
    }
    
    // Build search.Query from conditions
    query, err := s.searchService.BuildQuery(conditions)
    if err != nil {
        return nil, fmt.Errorf("failed to build query: %w", err)
    }
    
    s.log.Info().
        Int("conditionCount", len(conditions)).
        Bool("emptyQuery", query.IsEmpty()).
        Msg("executing boolean query")
    
    // Execute search using Index
    results, err := s.index.Find(ctx, search.FindOpts{
        Query: query,
        Sort: search.SortSpec{
            Field:     search.SortByPath,
            Direction: search.SortAsc,
        },
    })
    if err != nil {
        return nil, fmt.Errorf("search failed: %w", err)
    }
    
    // Convert results to Notes
    notes := make([]Note, len(results.Items))
    for i, result := range results.Items {
        notes[i] = documentToNote(result.Document)
    }
    
    s.log.Debug().Int("count", len(notes)).Msg("boolean query completed")
    return notes, nil
}
```

#### Step 2.2: Remove Old SQL Implementation

**Before**:
```go
// OLD CODE - DELETE THIS
db, err := s.dbService.GetDB(ctx)
glob := filepath.Join(s.notebookPath, "**", "*.md")
whereClause, params, err := s.searchService.BuildWhereClauseWithGlob(conditions, glob)
query := `SELECT * FROM read_markdown(?, include_filepath:=true) WHERE ` + whereClause + ` ORDER BY file_path`
rows, err := db.QueryContext(timeoutCtx, query, allParams...)
// ... (150+ lines of SQL parsing code)
```

**After**:
```go
// NEW CODE - Already shown in Step 2.1 above
```

**Checklist**:
- [ ] Replace SearchWithConditions() implementation
- [ ] Remove all SQL-related code from method
- [ ] Maintain same function signature
- [ ] Maintain same sorting behavior (ORDER BY file_path → SortByPath)
- [ ] Keep same error handling patterns
- [ ] Keep same logging patterns

---

### Phase 3: Update Tests

**Files**: `internal/services/note_test.go`, `internal/services/search_test.go`  
**Time**: 3-4 hours

#### Step 3.1: Update SearchWithConditions Tests

**Pattern for All Tests**:

**Before** (DuckDB):
```go
func TestNoteService_SearchWithConditions_SimpleAnd(t *testing.T) {
    ctx := context.Background()
    db := services.NewDbService()
    t.Cleanup(func() {
        if err := db.Close(); err != nil {
            t.Logf("warning: failed to close db: %v", err)
        }
    })
    
    tmpDir := t.TempDir()
    cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
    
    notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
    // ... create test notes ...
    
    noteService := services.NewNoteService(cfg, db, nil, notebookDir)
    
    // ... test execution ...
}
```

**After** (Bleve):
```go
func TestNoteService_SearchWithConditions_SimpleAnd(t *testing.T) {
    ctx := context.Background()
    tmpDir := t.TempDir()
    cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
    
    notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
    // ... create test notes ...
    
    // Create test index
    index := testutil.CreateTestIndex(t, notebookDir)
    
    noteService := services.NewNoteService(cfg, nil, index, notebookDir)
    
    // ... test execution (unchanged) ...
}
```

**Tests to Update** (40 tests):
```
TestNoteService_SearchWithConditions_SimpleAnd
TestNoteService_SearchWithConditions_MultipleAnd
TestNoteService_SearchWithConditions_Or
TestNoteService_SearchWithConditions_Not
TestNoteService_SearchWithConditions_MixedConditions
TestNoteService_SearchWithConditions_PathGlob
TestNoteService_SearchWithConditions_TitleMatch
TestNoteService_SearchWithConditions_EmptyResults
TestNoteService_SearchWithConditions_TagAlias
... (30+ more)
```

**Checklist**:
- [ ] Replace `NewDbService()` with `testutil.CreateTestIndex()`
- [ ] Remove `db.Close()` cleanup
- [ ] Update `NewNoteService(cfg, db, nil, ...)` to `NewNoteService(cfg, nil, index, ...)`
- [ ] Keep test data creation unchanged
- [ ] Keep assertions unchanged
- [ ] Run each test individually: `mise run test -- TestNoteService_SearchWithConditions_SimpleAnd`

#### Step 3.2: Handle Link Query Tests

**Tests Affected**:
```
TestNoteService_SearchWithConditions_LinksTo
TestNoteService_SearchWithConditions_LinkedBy
```

**Update Strategy**:
```go
func TestNoteService_SearchWithConditions_LinksTo(t *testing.T) {
    t.Skip("Link queries deferred to Phase 5.3 - requires link graph index")
    
    // TODO Phase 5.3: Re-enable this test
    // Link queries require separate graph index implementation
    // See: .memory/assessment-phase523-migration.md
}

func TestNoteService_SearchWithConditions_LinksToError(t *testing.T) {
    // NEW TEST: Verify error message for link queries
    ctx := context.Background()
    tmpDir := t.TempDir()
    cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
    
    notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
    index := testutil.CreateTestIndex(t, notebookDir)
    noteService := services.NewNoteService(cfg, nil, index, notebookDir)
    
    conditions := []services.QueryCondition{
        {Type: "and", Field: "links-to", Operator: "=", Value: "docs/*.md"},
    }
    
    _, err := noteService.SearchWithConditions(ctx, conditions)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "link queries are not yet supported")
    assert.Contains(t, err.Error(), "Phase 5.3")
}
```

**Checklist**:
- [ ] Skip existing link query tests with Phase 5.3 reference
- [ ] Add new tests to verify error messages
- [ ] Update test count expectations (171/172 → 171/171)

#### Step 3.3: Run Full Test Suite

```bash
# Run all note service tests
mise run test -- NoteService

# Run all search service tests
mise run test -- SearchService

# Run full suite
mise run test
```

**Expected**:
- All tests passing
- No DuckDB-related errors
- Link query tests skipped or error-verified
- Performance similar or better

**Checklist**:
- [ ] All NoteService tests passing
- [ ] All SearchService tests passing
- [ ] Full suite: 171/171 tests passing
- [ ] No regressions in other services

---

### Phase 4: Documentation & Manual Testing

**Time**: 1-2 hours

#### Step 4.1: Update CHANGELOG.md

**Add Breaking Changes Section**:

```markdown
## [Unreleased]

### Breaking Changes

#### Link Queries Temporarily Unavailable

The `notes search query` command no longer supports `links-to` and `linked-by` 
conditions. These queries require a dedicated link graph index, which will be 
implemented in Phase 5.3.

**Affected Commands**:
```bash
# These will return errors
opennotes notes search query --and links-to=docs/*.md
opennotes notes search query --and linked-by=plan.md
```

**Temporary Workaround**:
Use the SQL query interface:
```bash
opennotes notes query "SELECT * FROM read_markdown('**/*.md') WHERE ..."
```

**Timeline**:
- Phase 5.2.3 (current): Basic queries only
- Phase 5.3 (planned): Link graph index with full link query support

**Tracking**: https://github.com/zenobi-us/opennotes/issues/XXX

### Changed

- **Bleve Migration**: `SearchWithConditions()` now uses Bleve Index instead of DuckDB
- **Performance**: Metadata and path queries are faster due to Bleve's inverted index
- **Error Messages**: Clearer error messages for unsupported query types

### Fixed

- Path glob queries now use optimized prefix matching where possible
```

#### Step 4.2: Update Documentation

**File**: `docs/commands/notes-search.md`

**Add Section**:
```markdown
## Link Queries (Coming in Phase 5.3)

Link queries (`links-to`, `linked-by`) are temporarily unavailable while we 
migrate to the new search system. They will be re-implemented in Phase 5.3 
with a dedicated link graph index.

### Affected Queries

- `--and links-to=target.md` - Find notes linking TO target
- `--and linked-by=source.md` - Find notes linked FROM source

### Workaround

Use the SQL query interface for now:

```bash
# Find notes linking to docs/architecture.md
opennotes notes query "
  SELECT * FROM read_markdown('**/*.md') 
  WHERE EXISTS (
    SELECT 1 FROM (
      SELECT unnest(COALESCE(TRY_CAST(metadata['links'] AS VARCHAR[]), ARRAY[]::VARCHAR[])) AS link
    ) WHERE link LIKE '%architecture.md'
  )
"
```

**Update Performance Section**:
```markdown
## Performance Characteristics

### Fast Queries
- Tag queries: `--and data.tag=work` (inverted index)
- Status queries: `--and data.status=active` (inverted index)
- Path prefix: `--and path=projects/*` (prefix index)

### Moderate Queries
- Path wildcards: `--and path=**/tasks/*.md` (wildcard scan)
- Complex combinations: Multiple AND/OR/NOT conditions

### Not Yet Supported
- Link queries: `--and links-to=target.md` (Phase 5.3)
```

#### Step 4.3: Manual CLI Testing

**Test Suite**:

```bash
# Build binary
mise run build

# Setup test notebook
mkdir -p /tmp/test-notebook
cd /tmp/test-notebook
../../dist/opennotes init

# Create test notes
cat > work-active.md << EOF
---
tag: work
status: active
---
# Work Active
EOF

cat > work-done.md << EOF
---
tag: work
status: done
---
# Work Done
EOF

cat > meeting-active.md << EOF
---
tag: meeting
status: active
---
# Meeting Active
EOF

# Test queries
echo "Test 1: Single tag"
../../dist/opennotes notes search query --and data.tag=work
# Expected: 2 notes (work-active, work-done)

echo "Test 2: Multiple AND"
../../dist/opennotes notes search query --and data.tag=work --and data.status=active
# Expected: 1 note (work-active)

echo "Test 3: OR conditions"
../../dist/opennotes notes search query --or data.tag=work --or data.tag=meeting
# Expected: 3 notes (all)

echo "Test 4: NOT condition"
../../dist/opennotes notes search query --and data.tag=work --not data.status=done
# Expected: 1 note (work-active)

echo "Test 5: Path query"
../../dist/opennotes notes search query --and path=*.md
# Expected: 3 notes (all)

echo "Test 6: Link query (should error)"
../../dist/opennotes notes search query --and links-to=docs/*.md
# Expected: Clear error message with Phase 5.3 reference
```

**Checklist**:
- [ ] Build binary successfully
- [ ] All basic queries return correct results
- [ ] Link queries return clear error
- [ ] Performance is acceptable
- [ ] No crashes or panics

---

### Phase 5: Integration & Verification

**Time**: 1 hour

#### Step 5.1: Final Test Run

```bash
# Full test suite
mise run test

# Expected output:
# PASS: 171/171 tests
# COVERAGE: ~XX%
# TIME: ~4-5 seconds
```

#### Step 5.2: Code Review Checklist

**Code Quality**:
- [ ] BuildQuery() follows existing patterns
- [ ] Error messages are clear and actionable
- [ ] Logging is appropriate (Debug/Info/Error)
- [ ] No TODO comments without issue references
- [ ] All new code has tests

**Migration Completeness**:
- [ ] SearchWithConditions() no longer uses DuckDB
- [ ] All SQL code removed from method
- [ ] documentToNote() reused from Phase 5.2.2
- [ ] Sorting behavior maintained

**Testing**:
- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] Link query tests appropriately handled
- [ ] No test flakiness

**Documentation**:
- [ ] CHANGELOG.md updated
- [ ] Breaking changes documented
- [ ] Workarounds provided
- [ ] Phase 5.3 referenced

#### Step 5.3: Commit & Tag

**Commit Message**:
```
feat(search)!: migrate SearchWithConditions to Bleve

BREAKING CHANGE: Link queries (links-to, linked-by) temporarily unavailable

Migrates NoteService.SearchWithConditions() from DuckDB SQL to Bleve Index.
This is part of the broader DuckDB removal effort (Phase 5.2).

Changes:
- Add SearchService.BuildQuery() to convert QueryCondition to search.Query
- Update SearchWithConditions() to use Index.Find() instead of SQL
- Maintain sorting behavior (ORDER BY file_path → SortByPath)
- Optimize path queries (prefix matching for simple globs)

Breaking Changes:
- Link queries (links-to, linked-by) now return clear error
- These require dedicated link graph index (planned for Phase 5.3)
- Temporary workaround: Use SQL query interface

Migration Guide:
- See .memory/assessment-phase523-migration.md for full details
- See docs/commands/notes-search.md for updated documentation
- Track Phase 5.3: https://github.com/zenobi-us/opennotes/issues/XXX

Tests: 171/171 passing
Performance: Maintained or improved
Coverage: Maintained

Related:
- Phase 5.2.1: Add Index field to NoteService
- Phase 5.2.2: Migrate getAllNotes() to Bleve
- Phase 5.2.3: Migrate SearchWithConditions() (this PR)
- Phase 5.2.4: Migrate Count() (next)
```

**Git Commands**:
```bash
git add internal/services/search.go
git add internal/services/note.go
git add internal/services/search_test.go
git add internal/services/note_test.go
git add CHANGELOG.md
git add docs/commands/notes-search.md
git add .memory/assessment-phase523-migration.md
git add .memory/plan-phase523-implementation.md

git commit -F commit-message.txt

git tag phase-5.2.3
```

**Checklist**:
- [ ] Clean commit message following Conventional Commits
- [ ] All files staged
- [ ] Memory files included
- [ ] Tag applied

---

## Success Criteria Review

Before marking task complete, verify:

✅ **Functionality**:
- [ ] All metadata field queries work
- [ ] Path queries work with prefix optimization
- [ ] Title queries work
- [ ] AND/OR/NOT logic correct
- [ ] Results sorted by path

✅ **Testing**:
- [ ] 15+ BuildQuery() unit tests passing
- [ ] 171/171 total tests passing
- [ ] Link query tests appropriately handled
- [ ] Manual CLI testing successful

✅ **Documentation**:
- [ ] CHANGELOG.md updated
- [ ] docs/commands/notes-search.md updated
- [ ] Error messages clear and actionable
- [ ] Phase 5.3 referenced

✅ **Quality**:
- [ ] No performance regressions
- [ ] Code follows existing patterns
- [ ] All dependencies removed (DuckDB from SearchWithConditions)
- [ ] Memory files complete

---

## Post-Implementation

### Update Task Status

**File**: `.memory/task-5d8f7e3a-phase523-searchwithconditions.md`

**Update Sections**:
```markdown
## Actual Outcome

**Date Completed**: 2026-02-02  
**Time Taken**: X hours (estimated 8-11 hours)  
**Tests**: 171/171 passing  
**Status**: ✅ Complete

### What Went Well
- (Fill in after implementation)

### Challenges
- (Fill in after implementation)

### Deviations from Plan
- (Fill in after implementation)

## Lessons Learned

### Technical Insights
- (Fill in after implementation)

### Process Improvements
- (Fill in after implementation)

### Future Considerations
- (Fill in after implementation)
```

### Create Phase 5.3 Issue

**GitHub Issue Template**:
```markdown
# Phase 5.3: Implement Link Graph Index

## Context

Link queries (links-to, linked-by) were deferred from Phase 5.2.3 because they
require a dedicated graph index. This phase implements that index and restores
full link query functionality.

## Current State

After Phase 5.2.3:
- ❌ Link queries return error
- ✅ Workaround: SQL query interface
- ✅ All other queries working via Bleve

## Requirements

### Link Query Support

1. **links-to**: Find notes linking TO target
   ```bash
   opennotes notes search query --and links-to=docs/*.md
   ```

2. **linked-by**: Find notes linked FROM source
   ```bash
   opennotes notes search query --and linked-by=plan.md
   ```

### Architecture

**Option A**: Bleve document with links array
- Store links in Document.Metadata['links']
- Query with array contains
- Simple but inefficient for graph traversal

**Option B**: Separate graph data structure
- Build adjacency list on index
- Maintain forward/backward link maps
- Efficient but complex

**Option C**: Hybrid approach
- Store links in documents
- Build in-memory graph cache
- Balance simplicity and performance

### Implementation Tasks

- [ ] Design link graph data structure
- [ ] Implement link extraction from markdown
- [ ] Add links to Document model
- [ ] Implement links-to query
- [ ] Implement linked-by query
- [ ] Add tests (30+ tests)
- [ ] Update documentation
- [ ] Re-enable skipped tests from Phase 5.2.3

### Success Criteria

- Link queries work as before
- Performance acceptable (< 100ms for small notebooks)
- Graph queries scale to 10k+ notes
- Full test coverage

### Resources

- Phase 5.2.3 Assessment: .memory/assessment-phase523-migration.md
- DuckDB SQL Implementation (reference): internal/services/search.go (buildConditionSQL)

## Estimated Time

**Research**: 2-3 hours  
**Implementation**: 8-10 hours  
**Testing**: 4-6 hours  
**Total**: 14-19 hours
```

---

## Appendix: Quick Reference

### Key Files

**Implementation**:
- `internal/services/search.go` - Add BuildQuery()
- `internal/services/note.go` - Update SearchWithConditions()

**Tests**:
- `internal/services/search_test.go` - BuildQuery() tests
- `internal/services/note_test.go` - SearchWithConditions() tests

**Documentation**:
- `CHANGELOG.md` - Breaking changes
- `docs/commands/notes-search.md` - Updated docs
- `.memory/assessment-phase523-migration.md` - Full assessment
- `.memory/plan-phase523-implementation.md` - This plan

### Commands

**Build & Test**:
```bash
mise run build                          # Build binary
mise run test                           # Full test suite
mise run test -- SearchService          # Search service tests
mise run test -- NoteService            # Note service tests
mise run lint                           # Check code quality
```

**Manual Testing**:
```bash
./dist/opennotes init                   # Initialize notebook
./dist/opennotes notes search query ... # Test queries
```

### Time Tracking

| Phase | Estimated | Actual | Notes |
|-------|-----------|--------|-------|
| 1: BuildQuery() | 2-3h | | |
| 2: Update Method | 1h | | |
| 3: Tests | 3-4h | | |
| 4: Docs | 1-2h | | |
| 5: Integration | 1h | | |
| **Total** | **8-11h** | | |

---

## Ready to Implement ✅

This plan provides:
- ✅ Step-by-step implementation guide
- ✅ Complete code examples
- ✅ Comprehensive test cases
- ✅ Documentation updates
- ✅ Success criteria
- ✅ Time estimates

**Next Action**: Begin Phase 1 - Implement BuildQuery()
