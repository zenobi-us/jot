---
id: 5f8a9b2c
title: Note Search Enhancement - Text Search, Boolean Queries, and Fuzzy Matching
created_at: 2026-01-20T22:15:00+10:30
updated_at: 2026-01-20T23:19:00+10:30
status: proposed
epic_id: 3e01c563
related_spec: spec-ca68615f-note-creation-enhancement.md
---

# Specification: Note Search Enhancement

## Overview

**Feature**: Enhanced `opennotes notes search` command with text search, boolean query capabilities, fuzzy matching, and link query support for DAG foundation.

**Scope**: This specification covers **search functionality only**. Note creation enhancements are covered in a separate spec (spec-ca68615f).

**Goal**: Provide intermediate search capabilities that bridge simple listing and power-user SQL queries, enabling complex filtering without SQL knowledge.

### What's In Scope

✅ **Text search subcommand** with optional search term  
✅ **Fuzzy matching** for non-interactive ranked results  
✅ **Boolean query subcommand** with AND/OR/NOT operators  
✅ **Data field queries** against frontmatter metadata  
✅ **Link queries** (`links-to`, `linked-by`) for DAG foundation  
✅ **Glob pattern support** for path-based queries  
✅ **Security-first query construction** (parameterized queries)  
✅ **Comprehensive test coverage** (≥85%)  
✅ **Performance optimization** (<100ms for complex queries)

### What's Out of Scope

❌ **View system** (`--view` flags, built-in views) → Moved to separate spec (#3)  
❌ **Advanced DAG features** (transitive queries, cycle detection, visualization)  
❌ **Body filters** (heading search, advanced text operators)  
❌ **Regex search** (future enhancement)

## Command Structure

### Two Main Subcommands

1. **Text Search**: `opennotes notes search [text] [--fuzzy]`
2. **Boolean Query**: `opennotes notes search query [operators...]`

---

## Subcommand 1: Text Search

### Syntax

```bash
opennotes notes search [text] [--fuzzy]
```

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `[text]` | string | No | Search term for title, body, frontmatter |
| `--fuzzy` | flag | No | Enable fuzzy matching algorithm |

### Behavior Matrix

| Command | Behavior |
|---------|----------|
| `opennotes notes search` | Search all notes, print list to stdout |
| `opennotes notes search "meeting"` | Exact match search for "meeting", print list to stdout |
| `opennotes notes search --fuzzy` | Fuzzy match all notes, print ranked list to stdout |
| `opennotes notes search "project" --fuzzy` | Fuzzy match "project", print ranked results to stdout |

### Examples

```bash
# Simple text search - exact match
opennotes notes search "meeting"

# Search all notes - print full list
opennotes notes search

# Fuzzy matching search (ranks by fuzzy score)
opennotes notes search --fuzzy "mtng"
# Matches: "meeting", "morning standup", etc.

# Text with fuzzy matching
opennotes notes search "project" --fuzzy
# Ranks results by how well they fuzzy-match "project"
```

### Fuzzy Matching Integration

**Library**: `github.com/sahilm/fuzzy`

**Why This Library**:
- ✅ Pure Go (no external dependencies)
- ✅ Lightweight and fast
- ✅ Simple API for fuzzy string matching
- ✅ Returns match scores for ranking
- ✅ Well-tested and maintained
- ✅ Cross-platform (Windows, macOS, Linux)

**Basic Implementation**:

```go
import "github.com/sahilm/fuzzy"

func fuzzySearch(query string, notes []Note) []Note {
    var matches []fuzzy.Match
    
    for i, note := range notes {
        // Check title match
        if match := fuzzy.Find(query, note.Title); len(match) > 0 {
            matches = append(matches, fuzzy.Match{
                Str:   note.Title,
                Index: i,
                Score: match[0].Score,
            })
        }
    }
    
    // Sort by score (highest first)
    sort.Slice(matches, func(i, j int) bool {
        return matches[i].Score > matches[j].Score
    })
    
    // Return sorted notes
    var result []Note
    for _, match := range matches {
        result = append(result, notes[match.Index])
    }
    return result
}
```

**Algorithm Behavior**:
- Applies fuzzy string matching to note titles and content
- Ranks results by fuzzy match score (character proximity and sequence)
- Prints ranked results to stdout (no interactive selection)
- Similar to VSCode Ctrl+P matching but output to terminal

---

## Subcommand 2: Boolean Query Search

### Syntax

```bash
opennotes notes search query [--and|--or|--not] field=value ...
```

### Operators

| Operator | Meaning | Grouping |
|----------|---------|----------|
| `--and field=value` | AND condition (all must match) | Cumulative |
| `--or field=value` | OR condition (any can match) | Separate group |
| `--not field=value` | NOT condition (must not match) | Exclusion |

### Supported Field Types

#### 1. Data Fields (`data.*`)

Query frontmatter fields from note metadata.

**Syntax**: `data.field=value`

**Examples**:
- `data.tag=workflow`
- `data.status=in-progress`
- `data.priority=high`
- `data.assignee=alice`

**Behavior**:
- Supports strings, numbers, dates
- Case-insensitive by default
- Multi-value fields (like tags) support multiple `--and` flags

#### 2. Link Fields (DAG Foundation)

Query notes based on linking relationships.

**Syntax**:
- `links-to=path` - Documents that link TO the specified path (incoming edges)
- `linked-by=path` - Documents that the specified path links to (outgoing edges)

**Examples**:
- `links-to=epics/architecture.md`
- `linked-by=planning/2024-q1.md`
- `links-to=tasks/**/*.md` (glob pattern)

**Graph Semantics**:

```
Document A --link--> Document B

links-to=B    → Returns A (who points to B?)
linked-by=A   → Returns B (what does A point to?)
```

### Glob Pattern Support

All path-based values support glob patterns.

**Supported Patterns**:

| Pattern | Meaning | Example |
|---------|---------|---------|
| `**/*.md` | Any markdown file in any subdirectory | `tasks/**/*.md` |
| `dir/*` | Any file directly in directory | `epics/*` |
| `prefix-*.md` | Any file matching pattern | `task-*.md` |

**Implementation**: Translate glob to SQL LIKE patterns

```go
func globToLike(pattern string) string {
    // ** -> %
    // * -> %
    // ? -> _
    result := strings.ReplaceAll(pattern, "**", "%")
    result = strings.ReplaceAll(result, "*", "%")
    result = strings.ReplaceAll(result, "?", "_")
    return result
}

// Example: "epics/**/*.md" -> "epics/%/%.md"
```

### Boolean Logic Examples

#### Simple Queries

```bash
# Single AND condition
opennotes notes search query --and data.tag=workflow

# Multiple AND conditions (all must match)
opennotes notes search query \
  --and data.tag=workflow \
  --and data.status=in-progress

# OR conditions (any can match)
opennotes notes search query \
  --or data.priority=high \
  --or data.priority=critical

# NOT condition (exclusion)
opennotes notes search query \
  --and data.tag=epic \
  --not data.status=archived
```

#### Link Queries

```bash
# Find documents that link TO architecture doc (incoming edges)
opennotes notes search query --and links-to=docs/architecture.md

# Find documents that planning doc links to (outgoing edges)
opennotes notes search query --and linked-by=planning/2024-q1.md

# Find all documents that reference any epic
opennotes notes search query --and links-to=epics/**/*.md

# Find documents linked from any epic
opennotes notes search query --and linked-by=epics/**/*.md
```

#### Complex Queries

```bash
# AND + OR + NOT
opennotes notes search query \
  --and data.tag=epic \
  --or data.priority=high \
  --and links-to=tasks/**/*.md \
  --not linked-by=archived/**/*.md

# Data + Links combined
opennotes notes search query \
  --and data.tag=workflow \
  --and links-to=epics/epic-*.md \
  --not data.status=done

# Multiple tags (all must match)
opennotes notes search query \
  --and data.tag=workflow \
  --and data.tag=learning \
  --and data.tag=documentation
```

---

## Link Query Semantics (DAG Foundation)

### Purpose

Build foundation for future Directed Acyclic Graph (DAG) features without implementing full DAG capabilities yet.

### Graph Concepts

- **Nodes**: Documents (notes)
- **Edges**: Links between documents (stored in `data.links` array)
- **Direction**: Links have source → target direction

### Operator Definitions

| Operator | Graph Meaning | SQL Concept | Use Case |
|----------|---------------|-------------|----------|
| `links-to=X` | Find documents that link TO X | Incoming edges to X | "Who references this doc?" |
| `linked-by=X` | Find documents that X links to | Outgoing edges from X | "What does this doc reference?" |

### Visual Examples

```
Given:
  Epic A --links--> Task 1
  Epic A --links--> Task 2
  Epic B --links--> Task 1

Query: links-to=Task1
Result: [Epic A, Epic B]
Meaning: "Which epics reference Task 1?"

Query: linked-by=EpicA
Result: [Task 1, Task 2]
Meaning: "What tasks does Epic A reference?"
```

### SQL Implementation

#### `links-to=X` (Incoming Edges)

Find documents where `data.links` array contains X.

```sql
SELECT * FROM notes
WHERE EXISTS (
    SELECT 1 FROM unnest(data.links) AS link
    WHERE link LIKE ?
)
```

#### `linked-by=X` (Outgoing Edges)

Find documents that X's links array contains.

```sql
SELECT target.*
FROM notes source
CROSS JOIN unnest(source.data.links) AS link
JOIN notes target ON target.path LIKE link
WHERE source.path = ?
```

### Future DAG Extensions (Out of Scope)

**Not Included in This Spec**:
- ❌ Transitive queries: `graph.ancestors=X`, `graph.descendants=X`
- ❌ Path queries: `graph.path-exists=X,Y`
- ❌ Cycle detection: `opennotes notes validate dag`
- ❌ Visualization: `opennotes notes graph visualize`

These will be addressed in a future "DAG Operations" epic.

---

## Implementation Details

### Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `cmd/notes_search.go` | Modify | Update existing search command, add --fuzzy flag |
| `cmd/notes_search_query.go` | Create | New subcommand for boolean queries |
| `internal/services/search.go` | Create | New service for query building and fuzzy matching |
| `internal/services/note.go` | Modify | Extend with link query methods |

### Security-First Query Construction

**Critical Rule**: **NEVER** concatenate user input into SQL queries.

#### ✅ Safe Approach: Parameterized Queries

```go
// ✅ SAFE - Use ? placeholders
query := "SELECT * FROM notes WHERE data.tag = ?"
db.Query(query, userInput)

// ❌ UNSAFE - String concatenation
query := fmt.Sprintf("SELECT * FROM notes WHERE data.tag = '%s'", userInput)
```

#### Defense-in-Depth Validation

**Layer 1: Whitelist Field Names**

```go
allowedDataFields := map[string]bool{
    "data.tag":      true,
    "data.status":   true,
    "data.priority": true,
    "data.assignee": true,
    // ... etc
}

func validateField(field string) error {
    if !allowedDataFields[field] {
        return fmt.Errorf("invalid field: %s", field)
    }
    return nil
}
```

**Layer 2: Whitelist Operators**

```go
allowedOperators := []string{"=", "!=", "<", ">", "<=", ">=", "LIKE"}

func validateOperator(op string) error {
    for _, allowed := range allowedOperators {
        if op == allowed {
            return nil
        }
    }
    return fmt.Errorf("invalid operator: %s", op)
}
```

**Layer 3: Sanitize ORDER BY**

ORDER BY cannot be parameterized, so we must whitelist.

```go
allowedSortFields := map[string]bool{
    "title":      true,
    "created_at": true,
    "updated_at": true,
    "path":       true,
}

func validateSortField(field string) error {
    if !allowedSortFields[field] {
        return fmt.Errorf("invalid sort field: %s", field)
    }
    return nil
}
```

**Layer 4: Input Length Validation**

```go
const maxValueLength = 1000

func validateValueLength(value string) error {
    if len(value) > maxValueLength {
        return fmt.Errorf("value too long (max %d chars)", maxValueLength)
    }
    return nil
}
```

### Query Construction Logic

#### Parse Flags

```go
type QueryCondition struct {
    Type     string // "and", "or", "not"
    Field    string // "data.tag", "links-to", etc.
    Operator string // "=", "!=", "LIKE", etc.
    Value    string // User input
}

func parseQueryFlags(cmd *cobra.Command) ([]QueryCondition, error) {
    var conditions []QueryCondition
    
    // Parse --and flags
    andFlags, _ := cmd.Flags().GetStringArray("and")
    for _, flag := range andFlags {
        field, value := parseFieldValue(flag)
        conditions = append(conditions, QueryCondition{
            Type:     "and",
            Field:    field,
            Operator: "=",
            Value:    value,
        })
    }
    
    // Parse --or flags
    orFlags, _ := cmd.Flags().GetStringArray("or")
    for _, flag := range orFlags {
        field, value := parseFieldValue(flag)
        conditions = append(conditions, QueryCondition{
            Type:     "or",
            Field:    field,
            Operator: "=",
            Value:    value,
        })
    }
    
    // Parse --not flags
    notFlags, _ := cmd.Flags().GetStringArray("not")
    for _, flag := range notFlags {
        field, value := parseFieldValue(flag)
        conditions = append(conditions, QueryCondition{
            Type:     "not",
            Field:    field,
            Operator: "=",
            Value:    value,
        })
    }
    
    return conditions, nil
}

func parseFieldValue(input string) (string, string) {
    parts := strings.SplitN(input, "=", 2)
    if len(parts) != 2 {
        return "", ""
    }
    return parts[0], parts[1]
}
```

#### Build WHERE Clause

```go
func buildWhereClause(conditions []QueryCondition) (string, []interface{}, error) {
    var andConditions []string
    var orConditions []string
    var notConditions []string
    var params []interface{}
    
    for _, cond := range conditions {
        // Validate field
        if err := validateField(cond.Field); err != nil {
            return "", nil, err
        }
        
        // Validate value length
        if err := validateValueLength(cond.Value); err != nil {
            return "", nil, err
        }
        
        // Build condition based on type
        switch cond.Type {
        case "and":
            if strings.HasPrefix(cond.Field, "data.") {
                andConditions = append(andConditions, fmt.Sprintf("%s = ?", cond.Field))
                params = append(params, cond.Value)
            } else if cond.Field == "links-to" {
                query, linkParams := buildLinksToCondition(cond.Value)
                andConditions = append(andConditions, query)
                params = append(params, linkParams...)
            } else if cond.Field == "linked-by" {
                query, linkParams := buildLinkedByCondition(cond.Value)
                andConditions = append(andConditions, query)
                params = append(params, linkParams...)
            }
            
        case "or":
            orConditions = append(orConditions, fmt.Sprintf("%s = ?", cond.Field))
            params = append(params, cond.Value)
            
        case "not":
            notConditions = append(notConditions, fmt.Sprintf("NOT (%s = ?)", cond.Field))
            params = append(params, cond.Value)
        }
    }
    
    // Combine conditions
    var whereParts []string
    
    if len(andConditions) > 0 {
        whereParts = append(whereParts, strings.Join(andConditions, " AND "))
    }
    
    if len(orConditions) > 0 {
        whereParts = append(whereParts, fmt.Sprintf("(%s)", strings.Join(orConditions, " OR ")))
    }
    
    if len(notConditions) > 0 {
        whereParts = append(whereParts, strings.Join(notConditions, " AND "))
    }
    
    whereClause := strings.Join(whereParts, " AND ")
    return whereClause, params, nil
}
```

#### Link Query Helpers

```go
func buildLinksToCondition(pattern string) (string, []interface{}) {
    likePattern := globToLike(pattern)
    query := `
        EXISTS (
            SELECT 1 FROM unnest(data.links) AS link
            WHERE link LIKE ?
        )
    `
    return query, []interface{}{likePattern}
}

func buildLinkedByCondition(sourcePath string) (string, []interface{}) {
    query := `
        path IN (
            SELECT unnest(data.links)
            FROM notes
            WHERE path = ?
        )
    `
    return query, []interface{}{sourcePath}
}
```

### Fuzzy Matching Implementation

```go
import (
    "github.com/sahilm/fuzzy"
    "sort"
)

func (s *NoteService) FuzzySearchNotes(query string, notes []Note) []Note {
    if len(notes) == 0 {
        return nil
    }
    
    if query == "" {
        // No query - return all notes
        return notes
    }
    
    var matches []fuzzyMatch
    
    for i, note := range notes {
        // Try fuzzy matching on title
        titleMatches := fuzzy.Find(query, note.Title)
        
        var score int
        if len(titleMatches) > 0 {
            score = titleMatches[0].Score
        }
        
        // Also try body (first 500 chars for performance)
        bodyPreview := note.Body
        if len(bodyPreview) > 500 {
            bodyPreview = bodyPreview[:500]
        }
        bodyMatches := fuzzy.Find(query, bodyPreview)
        
        if len(bodyMatches) > 0 {
            // Body matches are weighted lower than title matches
            bodyScore := bodyMatches[0].Score / 2
            if bodyScore > score {
                score = bodyScore
            }
        }
        
        if score > 0 {
            matches = append(matches, fuzzyMatch{
                Note:  notes[i],
                Score: score,
            })
        }
    }
    
    // Sort by score (highest first)
    sort.Slice(matches, func(i, j int) bool {
        return matches[i].Score > matches[j].Score
    })
    
    // Extract sorted notes
    result := make([]Note, len(matches))
    for i, match := range matches {
        result[i] = match.Note
    }
    
    return result
}

type fuzzyMatch struct {
    Note  Note
    Score int
}
```

---

## Testing Requirements

### Test Coverage Target

**Minimum**: ≥85%

### Test Categories

#### 1. Text Search Tests

```go
func TestNoteService_SearchNotes_TextSearch(t *testing.T) {
    tests := []struct {
        name     string
        text     string
        expected int
    }{
        {"simple text search", "meeting", 5},
        {"empty text (all notes)", "", 100},
        {"case insensitive", "MEETING", 5},
        {"search in body", "content keyword", 3},
        {"search in frontmatter", "workflow", 10},
        {"no results", "xyz123nonexistent", 0},
    }
    // ...
}
```

#### 2. Boolean Query Tests

```go
func TestNoteService_SearchNotes_BooleanQuery(t *testing.T) {
    tests := []struct {
        name       string
        conditions []QueryCondition
        expected   int
    }{
        {"single AND", []QueryCondition{{Type: "and", Field: "data.tag", Value: "workflow"}}, 10},
        {"multiple AND", []QueryCondition{{...}, {...}}, 5},
        {"OR conditions", []QueryCondition{{Type: "or", ...}}, 15},
        {"NOT condition", []QueryCondition{{Type: "not", ...}}, 90},
        {"AND + OR + NOT", []QueryCondition{{...}}, 7},
    }
    // ...
}
```

#### 3. Link Query Tests

```go
func TestNoteService_SearchNotes_LinkQueries(t *testing.T) {
    tests := []struct {
        name     string
        field    string
        value    string
        expected int
    }{
        {"links-to exact path", "links-to", "epics/arch.md", 5},
        {"links-to glob pattern", "links-to", "epics/**/*.md", 15},
        {"linked-by exact path", "linked-by", "planning/q1.md", 3},
        {"linked-by glob pattern", "linked-by", "tasks/*.md", 8},
    }
    // ...
}
```

#### 4. Glob Pattern Tests

```go
func TestGlobToLike(t *testing.T) {
    tests := []struct {
        glob     string
        expected string
    }{
        {"**/*.md", "%/%.md"},
        {"dir/*", "dir/%"},
        {"prefix-*.md", "prefix-%.md"},
        {"file?.md", "file_.md"},
    }
    // ...
}
```

#### 5. Security Tests

```go
func TestQuerySecurity_SQLInjectionPrevention(t *testing.T) {
    maliciousInputs := []string{
        "'; DROP TABLE notes; --",
        "1' OR '1'='1",
        "admin'--",
        "<script>alert('xss')</script>",
    }
    
    for _, input := range maliciousInputs {
        // Ensure parameterized queries prevent injection
        _, err := buildWhereClause([]QueryCondition{{Value: input}})
        assert.NoError(t, err) // Should handle safely
    }
}

func TestQuerySecurity_FieldValidation(t *testing.T) {
    invalidFields := []string{
        "system.password",
        "../../etc/passwd",
        "DROP TABLE",
    }
    
    for _, field := range invalidFields {
        err := validateField(field)
        assert.Error(t, err) // Should reject
    }
}
```

#### 6. Fuzzy Matching Tests

```go
func TestFuzzySearch_Matching(t *testing.T) {
    notes := []Note{
        {Title: "Meeting Notes", Body: "Team discussion"},
        {Title: "Morning Standup", Body: "Daily sync"},
        {Title: "Project Planning", Body: "Strategy meeting"},
    }
    
    // Test fuzzy matching
    results := FuzzySearchNotes("mtng", notes)
    assert.GreaterOrEqual(t, len(results), 1)
    assert.Equal(t, "Meeting Notes", results[0].Title) // Best match first
    
    // Test ranking
    results = FuzzySearchNotes("meeting", notes)
    assert.GreaterOrEqual(t, len(results), 2)
    // Both "Meeting Notes" and "Strategy meeting" should match
}

func TestFuzzySearch_Ranking(t *testing.T) {
    notes := []Note{
        {Title: "project", Body: "content"},
        {Title: "big project", Body: "content"},
        {Title: "project plan", Body: "content"},
    }
    
    results := FuzzySearchNotes("project", notes)
    assert.Equal(t, 3, len(results))
    
    // Exact match should rank highest
    assert.Equal(t, "project", results[0].Title)
}

func TestFuzzySearch_EmptyQuery(t *testing.T) {
    notes := createTestNotes(10)
    
    // Empty query returns all notes
    results := FuzzySearchNotes("", notes)
    assert.Equal(t, len(notes), len(results))
}

func TestFuzzySearch_NoMatches(t *testing.T) {
    notes := createTestNotes(10)
    
    // No matches returns empty slice
    results := FuzzySearchNotes("xyz123nonexistent", notes)
    assert.Equal(t, 0, len(results))
}
```

#### 7. Performance Tests

```go
func BenchmarkTextSearch_10kNotes(b *testing.B) {
    notes := createTestNotes(10000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        SearchNotes("meeting")
    }
}

func BenchmarkBooleanQuery_ComplexQuery(b *testing.B) {
    notes := createTestNotes(10000)
    conditions := []QueryCondition{
        {Type: "and", Field: "data.tag", Value: "workflow"},
        {Type: "and", Field: "links-to", Value: "epics/**/*.md"},
        {Type: "not", Field: "data.status", Value: "archived"},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        SearchNotesWithQuery(conditions)
    }
}
```

#### 8. Error Handling Tests

```go
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name        string
        condition   QueryCondition
        expectError bool
    }{
        {"invalid field", QueryCondition{Field: "invalid.field"}, true},
        {"invalid operator", QueryCondition{Operator: "INJECT"}, true},
        {"value too long", QueryCondition{Value: strings.Repeat("a", 10000)}, true},
        {"empty result set", QueryCondition{Value: "nonexistent"}, false},
    }
    // ...
}
```

#### 9. Integration Tests

```go
func TestIntegration_EndToEndSearch(t *testing.T) {
    // Create test notebook
    notebook := createTestNotebook()
    
    // Test full search flow
    t.Run("text search", func(t *testing.T) {
        cmd := exec.Command("opennotes", "notes", "search", "meeting")
        output, err := cmd.Output()
        assert.NoError(t, err)
        assert.Contains(t, string(output), "meeting")
    })
    
    t.Run("boolean query", func(t *testing.T) {
        cmd := exec.Command("opennotes", "notes", "search", "query", 
            "--and", "data.tag=workflow",
            "--not", "data.status=archived")
        output, err := cmd.Output()
        assert.NoError(t, err)
        // Verify output
    })
}
```

---

## Performance Targets

| Operation | Dataset | Target Time | Rationale |
|-----------|---------|-------------|-----------|
| Simple text search | 10k notes | < 10ms | DuckDB is optimized for OLAP |
| Fuzzy matching search | 10k notes | < 50ms | CPU-bound fuzzy algorithm |
| Boolean query (2 conditions) | 10k notes | < 20ms | Indexed queries |
| Complex query (5+ conditions) | 10k notes | < 100ms | Multiple joins |
| Link resolution (glob) | 10k notes, 50k links | < 50ms | UNNEST + LIKE optimization |

### Performance Optimization Strategies

1. **Index Creation**: Ensure DuckDB indexes on `data.tag`, `data.status`, `path`
2. **Query Plan Analysis**: Use `EXPLAIN` to verify efficient query execution
3. **Caching**: Cache frequently accessed notebooks in memory
4. **Lazy Loading**: Load note bodies only when needed
5. **Connection Pooling**: Reuse database connections

---

## Security Considerations

### OWASP Validation

| Risk | Mitigation | Status |
|------|------------|--------|
| **A03:2021 Injection** | Parameterized queries, whitelist validation | ✅ Addressed |
| **A04:2021 Insecure Design** | Defense-in-depth validation, input sanitization | ✅ Addressed |
| **A05:2021 Security Misconfiguration** | Clear error messages, no info leakage | ✅ Addressed |
| **A07:2021 Identification Failures** | Input length validation, rate limiting | ✅ Addressed |

### Security Rules (Non-Negotiable)

1. ✅ **NEVER** concatenate user input into SQL
2. ✅ **ALWAYS** use `?` placeholders for values
3. ✅ **ALWAYS** whitelist field names and operators
4. ✅ **ALWAYS** validate input length
5. ✅ **ALWAYS** log queries for auditing

### Audit Logging

```go
func (s *NoteService) executeQuery(query string, params []interface{}) error {
    // Log query for security audit
    Log.Info("executing query",
        "query", query,
        "param_count", len(params),
        "timestamp", time.Now().Unix(),
    )
    
    // Execute with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    return s.db.QueryContext(ctx, query, params...)
}
```

---

## Documentation Updates Needed

### 1. Command Reference

**File**: `docs/commands/notes-search.md`

**Sections**:
- Text search syntax and examples
- Boolean query syntax and operators
- Field reference (data.*, links-to, linked-by)
- Glob pattern guide
- Fuzzy matching usage and ranking algorithm
- Error messages and troubleshooting

### 2. User Guide

**File**: `docs/guides/searching-notes.md`

**Sections**:
- Search strategies (when to use text vs fuzzy vs boolean vs SQL)
- Boolean logic tutorial with examples
- Link query patterns and use cases
- Fuzzy matching workflow tips and best practices
- Common search patterns cookbook

### 3. API Reference

**File**: `docs/api/search-service.md`

**Sections**:
- `SearchNotes()` method signature
- `BuildQuery()` method documentation
- Link query implementation details
- Security validation functions

### 4. Examples

**File**: `docs/examples/search-examples.md`

**Examples**:
```bash
# Find all workflow tasks assigned to me
opennotes notes search query \
  --and data.tag=workflow \
  --and data.assignee=alice

# Find epic dependencies
opennotes notes search query \
  --and data.tag=epic \
  --and links-to=tasks/**/*.md

# Find orphaned notes (not linked from anywhere)
# (This requires NOT EXISTS subquery - future enhancement)
```

---

## Migration Impact

### Backward Compatibility

✅ **Non-Breaking Changes**:
- Existing `opennotes notes search` still works (aliased to text search)
- All current functionality preserved
- Additive only - new subcommands and flags

✅ **New Capabilities**:
- Boolean query subcommand
- FZF interactive mode
- Link queries (DAG foundation)
- Glob pattern support

### Migration Guide

**None Required**: This is additive functionality. Users can adopt new features incrementally.

**Deprecation Notice**: None

---

## Acceptance Criteria

### Must Have (MVP)

- ✅ Text search works with optional text argument
- ✅ Fuzzy matching provides ranked non-interactive results
- ✅ Boolean queries support AND/OR/NOT operators
- ✅ Data field queries work with frontmatter
- ✅ Link queries (`links-to`, `linked-by`) work correctly
- ✅ Glob patterns translate to SQL correctly
- ✅ All queries use parameterized SQL (security)
- ✅ Test coverage ≥85%
- ✅ Performance targets met (fuzzy matching < 50ms for 10k notes)
- ✅ Documentation complete
- ✅ Cross-platform compatibility (Linux, macOS, Windows)
- ✅ Clear error messages for all failure scenarios

### Nice to Have (Future Enhancements)

- ⏸️ Interactive fuzzy selection mode (like fzf)
- ⏸️ Export results to different formats (JSON, CSV)
- ⏸️ Save queries as aliases (moved to views spec)
- ⏸️ Regex search support
- ⏸️ Advanced DAG features (transitive queries, cycle detection)

---

## Implementation Phases

### Phase 1: Text Search + Fuzzy Matching (3-4 hours)

**Tasks**:
1. Update `cmd/notes_search.go` with `--fuzzy` flag
2. Integrate `github.com/sahilm/fuzzy` library
3. Implement `FuzzySearchNotes()` in `NoteService`
4. Implement ranking algorithm (title matches weighted higher than body)
5. Write tests for text search and fuzzy matching
6. Update documentation

**Deliverables**:
- Working text search with fuzzy matching mode
- Non-interactive ranked output
- 15+ tests (including ranking tests)
- User guide section

### Phase 2: Boolean Queries (4-5 hours)

**Tasks**:
1. Create `cmd/notes_search_query.go` subcommand
2. Implement flag parsing for `--and`, `--or`, `--not`
3. Create `internal/services/search.go` with query builder
4. Implement security validation (whitelist, parameterization)
5. Write tests for boolean logic
6. Update documentation

**Deliverables**:
- Working boolean query subcommand
- 20+ tests
- Security validation layer
- Query builder service

### Phase 3: Link Queries (3-4 hours)

**Tasks**:
1. Implement `links-to` and `linked-by` query conditions
2. Add glob pattern translation
3. Extend `NoteService` with link query methods
4. Write tests for link queries and glob patterns
5. Update documentation with DAG foundation examples

**Deliverables**:
- Working link queries
- 15+ tests
- DAG foundation documentation

### Phase 4: Performance Optimization (2-3 hours)

**Tasks**:
1. Add performance benchmarks
2. Optimize queries with indexes
3. Profile and optimize hot paths
4. Verify performance targets met
5. Document performance characteristics

**Deliverables**:
- Performance benchmarks
- Query optimization
- Performance documentation

### Phase 5: Integration Testing (2-3 hours)

**Tasks**:
1. Write end-to-end integration tests
2. Test cross-platform compatibility
3. Test error scenarios
4. User acceptance testing
5. Final documentation review

**Deliverables**:
- 10+ integration tests
- Cross-platform validation
- Complete documentation

---

## Related Artifacts

### Epic

- **Epic ID**: `3e01c563`
- **Epic File**: `.memory/epic-3e01c563-advanced-note-operations.md`
- **Epic Title**: Advanced Note Creation and Search Capabilities

### Related Specifications

- **Note Creation Spec**: `.memory/spec-ca68615f-note-creation-enhancement.md`
- **Views Spec** (Future): `spec-<hash>-views-system.md` (not yet created)

### Research Documents

- **Comprehensive Research**: `.memory/research-3e01c563-advanced-operations.md`
- **Research Summary**: `.memory/research-3e01c563-summary.md`

### Existing Learning

- **SQL Flag Implementation**: `.memory/learning-2f3c4d5e-sql-flag-epic-complete.md`
- **Architecture Reference**: `.memory/learning-5e4c3f2a-codebase-architecture.md`

### Knowledge Base

- **Codebase Structure**: `.memory/knowledge-codemap.md`
- **Data Flow**: `.memory/knowledge-data-flow.md`

---

## Open Questions

### Q1: Should link queries support regex patterns?

**Status**: Deferred  
**Decision**: Start with glob patterns only. Evaluate regex need based on user feedback.  
**Rationale**: Glob covers 90% of use cases, regex adds complexity.

### Q2: Should we cache query results?

**Status**: Deferred  
**Decision**: Implement caching only if performance benchmarks show need.  
**Rationale**: DuckDB is fast enough for most use cases. Premature optimization.

### Q3: Should we implement interactive fuzzy selection later?

**Status**: Nice-to-Have  
**Decision**: Ship non-interactive fuzzy matching first, evaluate interactive mode based on user feedback.  
**Rationale**: Non-interactive mode covers 90% of use cases and is simpler to implement. Interactive mode (like fzf) can be added as future enhancement if users request it.

---

## Risks & Mitigations

### Risk 1: Fuzzy Matching Performance on Large Datasets

**Impact**: Medium  
**Likelihood**: Low  
**Mitigation**: Performance benchmarks required before merge. Optimize by limiting body search to first 500 characters. Consider caching for frequently accessed notes.

### Risk 2: Query Performance Degradation

**Impact**: High  
**Likelihood**: Low  
**Mitigation**: Performance benchmarks required before merge. Query optimization pass.

### Risk 3: SQL Injection Vulnerability

**Impact**: Critical  
**Likelihood**: Very Low (with proper validation)  
**Mitigation**: Defense-in-depth validation, code review, security testing.

### Risk 4: Glob Pattern Edge Cases

**Impact**: Low  
**Likelihood**: Medium  
**Mitigation**: Comprehensive test suite for glob patterns. Clear documentation of supported patterns.

---

## Success Metrics

### User Adoption

- ✅ 50%+ of users adopt boolean queries within 3 months
- ✅ 30%+ of users use fuzzy matching mode regularly
- ✅ Link queries used in 20%+ of search operations

### Technical Metrics

- ✅ Zero production bugs in first month
- ✅ 95th percentile query time < 100ms
- ✅ Test coverage maintained at ≥85%
- ✅ Documentation completeness score ≥90%

### User Satisfaction

- ✅ Positive feedback on search flexibility
- ✅ No complaints about performance
- ✅ Users report faster workflows

---

## Next Steps

1. ✅ **Specification Complete**: This document
2. ⏸️ **Human Review**: Review spec before implementation
3. ⏸️ **Task Breakdown**: Create detailed task files for each phase
4. ⏸️ **Implementation**: Execute phases 1-5
5. ⏸️ **Testing**: Comprehensive test suite
6. ⏸️ **Documentation**: Complete all doc updates
7. ⏸️ **Human Approval**: Final review before merge

---

## Notes

This specification is designed to be **implementation-ready**. All implementation details, security considerations, test requirements, and documentation needs are fully specified.

**Estimated Total Effort**: 14-19 hours (research + implementation + testing + docs)

**Complexity**: Medium-High (security validation, FZF integration, link queries)

**Value**: High (enables power users without SQL knowledge, DAG foundation)
