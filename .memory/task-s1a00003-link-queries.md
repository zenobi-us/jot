---
id: s1a00003
title: Link Queries and Glob Pattern Support
created_at: 2026-01-22T12:55:00+10:30
updated_at: 2026-01-22T12:55:00+10:30
status: done
epic_id: 3e01c563
phase_id: 4a8b9c0d
assigned_to: unassigned
estimated_hours: 1.5
depends_on: s1a00002
---

# Task: Link Queries and Glob Pattern Support

## Objective

Implement link query operators (`links-to`, `linked-by`) and glob pattern support for DAG foundation queries.

## Steps

### 1. Add link query handling to SearchService

Extend `internal/services/search.go`:

```go
// globToLike converts glob patterns to SQL LIKE patterns
func globToLike(pattern string) string {
    // Escape SQL special chars first
    pattern = strings.ReplaceAll(pattern, "%", "\\%")
    pattern = strings.ReplaceAll(pattern, "_", "\\_")
    
    // Convert glob patterns
    pattern = strings.ReplaceAll(pattern, "**", "{{DOUBLESTAR}}")
    pattern = strings.ReplaceAll(pattern, "*", "%")
    pattern = strings.ReplaceAll(pattern, "{{DOUBLESTAR}}", "%")
    pattern = strings.ReplaceAll(pattern, "?", "_")
    
    return pattern
}

// buildLinksToCondition creates SQL for finding docs that link TO a target
func (s *SearchService) buildLinksToCondition(pattern string) (string, []interface{}) {
    likePattern := globToLike(pattern)
    
    // DuckDB: unnest array and check if any link matches
    query := `EXISTS (
        SELECT 1 FROM unnest(COALESCE(data.links, [])) AS link
        WHERE link LIKE ?
    )`
    
    return query, []interface{}{likePattern}
}

// buildLinkedByCondition creates SQL for finding docs that a source links TO
func (s *SearchService) buildLinkedByCondition(sourcePath string) (string, []interface{}) {
    // Find all docs that the source document links to
    query := `path IN (
        SELECT unnest(COALESCE(data.links, []))
        FROM notes
        WHERE path = ?
    )`
    
    return query, []interface{}{sourcePath}
}
```

### 2. Update BuildWhereClause to handle link queries

```go
func (s *SearchService) BuildWhereClause(conditions []QueryCondition) (string, []interface{}, error) {
    var andParts, orParts, notParts []string
    var params []interface{}
    
    for _, cond := range conditions {
        var sqlPart string
        var condParams []interface{}
        
        switch cond.Field {
        case "links-to":
            sqlPart, condParams = s.buildLinksToCondition(cond.Value)
        case "linked-by":
            sqlPart, condParams = s.buildLinkedByCondition(cond.Value)
        default:
            // Regular field condition
            sqlPart = fmt.Sprintf("%s = ?", cond.Field)
            condParams = []interface{}{cond.Value}
        }
        
        params = append(params, condParams...)
        
        switch cond.Type {
        case "and":
            andParts = append(andParts, sqlPart)
        case "or":
            orParts = append(orParts, sqlPart)
        case "not":
            notParts = append(notParts, fmt.Sprintf("NOT (%s)", sqlPart))
        }
    }
    
    // ... rest of method unchanged
}
```

### 3. Add glob pattern tests

```go
func TestGlobToLike_SingleStar(t *testing.T) {
    tests := []struct {
        glob     string
        expected string
    }{
        {"*.md", "%.md"},
        {"dir/*", "dir/%"},
        {"prefix-*", "prefix-%"},
    }
    for _, tt := range tests {
        result := globToLike(tt.glob)
        assert.Equal(t, tt.expected, result)
    }
}

func TestGlobToLike_DoubleStar(t *testing.T) {
    tests := []struct {
        glob     string
        expected string
    }{
        {"**/*.md", "%/%.md"},
        {"epics/**", "epics/%"},
        {"**/tasks/*.md", "%/tasks/%.md"},
    }
    for _, tt := range tests {
        result := globToLike(tt.glob)
        assert.Equal(t, tt.expected, result)
    }
}

func TestGlobToLike_QuestionMark(t *testing.T) {
    tests := []struct {
        glob     string
        expected string
    }{
        {"file?.md", "file_.md"},
        {"task-??.md", "task-__.md"},
    }
    for _, tt := range tests {
        result := globToLike(tt.glob)
        assert.Equal(t, tt.expected, result)
    }
}

func TestGlobToLike_Escape(t *testing.T) {
    // Ensure SQL special chars are escaped
    tests := []struct {
        glob     string
        expected string
    }{
        {"100%", "100\\%"},
        {"file_name", "file\\_name"},
    }
    for _, tt := range tests {
        result := globToLike(tt.glob)
        assert.Equal(t, tt.expected, result)
    }
}
```

### 4. Add link query tests

```go
func TestSearchService_LinksTo_ExactPath(t *testing.T) {
    // Setup: Create notes with links
    // noteA links to noteB
    // Query: links-to=noteB.md
    // Expected: Returns noteA
}

func TestSearchService_LinksTo_GlobPattern(t *testing.T) {
    // Setup: Create notes linking to epics/
    // Query: links-to=epics/**/*.md
    // Expected: Returns all notes that link to any epic
}

func TestSearchService_LinkedBy_ExactPath(t *testing.T) {
    // Setup: noteA links to noteB and noteC
    // Query: linked-by=noteA.md
    // Expected: Returns noteB and noteC
}

func TestSearchService_LinkedBy_GlobPattern(t *testing.T) {
    // Setup: Multiple epics linking to tasks
    // Query: linked-by=epics/*.md (not supported - exact path only)
    // Expected: Error or single exact match
}

func TestSearchService_LinkQueries_Combined(t *testing.T) {
    // Query: --and data.tag=epic --and links-to=tasks/**/*.md
    // Expected: Epics that link to tasks
}
```

### 5. Add integration tests

```go
func TestNotesSearchQuery_LinksTo_Integration(t *testing.T) {
    // End-to-end test with real notebook
    notebook := createTestNotebook(t)
    
    // Create linked notes
    createNote(t, notebook, "epic.md", map[string]interface{}{
        "links": []string{"tasks/task1.md", "tasks/task2.md"},
    })
    createNote(t, notebook, "tasks/task1.md", nil)
    createNote(t, notebook, "tasks/task2.md", nil)
    
    // Test links-to
    cmd := exec.Command("opennotes", "notes", "search", "query",
        "--and", "links-to=tasks/task1.md")
    output, err := cmd.Output()
    
    assert.NoError(t, err)
    assert.Contains(t, string(output), "epic.md")
}

func TestNotesSearchQuery_LinkedBy_Integration(t *testing.T) {
    // Similar setup, test linked-by
}
```

## Expected Outcome

- `opennotes notes search query --and links-to=docs/architecture.md` - find docs linking TO architecture
- `opennotes notes search query --and links-to=epics/**/*.md` - find docs linking to any epic
- `opennotes notes search query --and linked-by=planning/q1.md` - find docs that q1 links TO
- Glob patterns work: `**/*.md`, `dir/*`, `prefix-*`, `file?.md`

## Acceptance Criteria

- [x] `links-to` operator finds incoming links
- [x] `linked-by` operator finds outgoing links
- [x] Glob patterns convert correctly to SQL LIKE
- [x] SQL special chars properly escaped
- [x] Empty `data.links` handled gracefully (COALESCE)
- [x] Combined with other conditions (AND/OR/NOT)
- [x] 12+ tests for link queries and globs (31 test functions added)
- [ ] Performance < 50ms for 10k notes with 50k links (not benchmarked)
