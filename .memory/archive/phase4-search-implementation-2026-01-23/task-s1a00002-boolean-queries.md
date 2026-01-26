---
id: s1a00002
title: Boolean Query Subcommand Implementation
created_at: 2026-01-22T12:55:00+10:30
updated_at: 2026-01-23T07:15:00+10:30
status: completed
epic_id: 3e01c563
phase_id: 4a8b9c0d
assigned_to: unassigned
estimated_hours: 2.5
depends_on: s1a00001
---

# Task: Boolean Query Subcommand Implementation

## Objective

Implement the `opennotes notes search query` subcommand with AND/OR/NOT boolean operators for complex filtering.

## Steps

### 1. Create cmd/notes_search_query.go

New subcommand for boolean queries:

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var notesSearchQueryCmd = &cobra.Command{
    Use:   "query",
    Short: "Search notes with boolean operators",
    Long: `Search notes using AND/OR/NOT boolean operators.

Examples:
  # Single condition
  opennotes notes search query --and data.tag=workflow
  
  # Multiple AND conditions
  opennotes notes search query --and data.tag=workflow --and data.status=active
  
  # OR conditions
  opennotes notes search query --or data.priority=high --or data.priority=critical
  
  # NOT condition
  opennotes notes search query --and data.tag=epic --not data.status=archived`,
    RunE: notesSearchQueryRunE,
}

func init() {
    notesSearchCmd.AddCommand(notesSearchQueryCmd)
    
    notesSearchQueryCmd.Flags().StringArray("and", []string{}, "AND condition (field=value)")
    notesSearchQueryCmd.Flags().StringArray("or", []string{}, "OR condition (field=value)")
    notesSearchQueryCmd.Flags().StringArray("not", []string{}, "NOT condition (field=value)")
}

func notesSearchQueryRunE(cmd *cobra.Command, args []string) error {
    // Parse flags
    andFlags, _ := cmd.Flags().GetStringArray("and")
    orFlags, _ := cmd.Flags().GetStringArray("or")
    notFlags, _ := cmd.Flags().GetStringArray("not")
    
    // Build conditions
    conditions, err := SearchService.ParseConditions(andFlags, orFlags, notFlags)
    if err != nil {
        return err
    }
    
    // Execute query
    notes, err := NoteService.SearchWithConditions(requireNotebook(), conditions)
    if err != nil {
        return err
    }
    
    // Render output
    return TuiRender("note-list", notes)
}
```

### 2. Add query building to SearchService

Extend `internal/services/search.go`:

```go
// QueryCondition represents a single search condition
type QueryCondition struct {
    Type     string // "and", "or", "not"
    Field    string // "data.tag", "data.status", etc.
    Operator string // "=", "!=", "LIKE"
    Value    string // user-provided value
}

// Whitelisted fields for security
var allowedFields = map[string]bool{
    "data.tag":      true,
    "data.tags":     true,
    "data.status":   true,
    "data.priority": true,
    "data.assignee": true,
    "data.author":   true,
    "data.type":     true,
    "data.category": true,
    "data.project":  true,
    "data.sprint":   true,
    "path":          true,
    "title":         true,
    "links-to":      true,
    "linked-by":     true,
}

// ParseConditions parses CLI flags into QueryConditions
func (s *SearchService) ParseConditions(andFlags, orFlags, notFlags []string) ([]QueryCondition, error) {
    var conditions []QueryCondition
    
    for _, flag := range andFlags {
        cond, err := s.parseCondition("and", flag)
        if err != nil {
            return nil, err
        }
        conditions = append(conditions, cond)
    }
    
    for _, flag := range orFlags {
        cond, err := s.parseCondition("or", flag)
        if err != nil {
            return nil, err
        }
        conditions = append(conditions, cond)
    }
    
    for _, flag := range notFlags {
        cond, err := s.parseCondition("not", flag)
        if err != nil {
            return nil, err
        }
        conditions = append(conditions, cond)
    }
    
    return conditions, nil
}

func (s *SearchService) parseCondition(condType, flag string) (QueryCondition, error) {
    parts := strings.SplitN(flag, "=", 2)
    if len(parts) != 2 {
        return QueryCondition{}, fmt.Errorf("invalid condition format: %s (expected field=value)", flag)
    }
    
    field, value := parts[0], parts[1]
    
    // Validate field (security)
    if !allowedFields[field] {
        return QueryCondition{}, fmt.Errorf("invalid field: %s", field)
    }
    
    // Validate value length (security)
    if len(value) > 1000 {
        return QueryCondition{}, fmt.Errorf("value too long (max 1000 chars)")
    }
    
    return QueryCondition{
        Type:     condType,
        Field:    field,
        Operator: "=",
        Value:    value,
    }, nil
}

// BuildWhereClause constructs parameterized SQL WHERE clause
func (s *SearchService) BuildWhereClause(conditions []QueryCondition) (string, []interface{}, error) {
    var andParts, orParts, notParts []string
    var params []interface{}
    
    for _, cond := range conditions {
        sqlPart := fmt.Sprintf("%s = ?", cond.Field)
        params = append(params, cond.Value)
        
        switch cond.Type {
        case "and":
            andParts = append(andParts, sqlPart)
        case "or":
            orParts = append(orParts, sqlPart)
        case "not":
            notParts = append(notParts, fmt.Sprintf("NOT (%s)", sqlPart))
        }
    }
    
    var whereParts []string
    
    if len(andParts) > 0 {
        whereParts = append(whereParts, strings.Join(andParts, " AND "))
    }
    
    if len(orParts) > 0 {
        whereParts = append(whereParts, fmt.Sprintf("(%s)", strings.Join(orParts, " OR ")))
    }
    
    if len(notParts) > 0 {
        whereParts = append(whereParts, strings.Join(notParts, " AND "))
    }
    
    return strings.Join(whereParts, " AND "), params, nil
}
```

### 3. Add SearchWithConditions to NoteService

```go
func (s *NoteService) SearchWithConditions(notebook *Notebook, conditions []QueryCondition) ([]Note, error) {
    whereClause, params, err := s.searchService.BuildWhereClause(conditions)
    if err != nil {
        return nil, err
    }
    
    query := "SELECT * FROM notes"
    if whereClause != "" {
        query += " WHERE " + whereClause
    }
    query += " ORDER BY updated DESC"
    
    // Log for audit
    s.logger.Info("executing boolean query", "conditions", len(conditions))
    
    return s.db.QueryNotes(query, params...)
}
```

### 4. Write security tests

```go
func TestSearchService_ParseConditions_ValidFields(t *testing.T)
func TestSearchService_ParseConditions_InvalidField(t *testing.T)
func TestSearchService_ParseConditions_ValueTooLong(t *testing.T)
func TestSearchService_ParseConditions_InvalidFormat(t *testing.T)
func TestSearchService_BuildWhereClause_SQLInjection(t *testing.T)
func TestSearchService_BuildWhereClause_Parameterized(t *testing.T)
```

### 5. Write functional tests

```go
func TestNotesSearchQuery_SingleAnd(t *testing.T)
func TestNotesSearchQuery_MultipleAnd(t *testing.T)
func TestNotesSearchQuery_OrConditions(t *testing.T)
func TestNotesSearchQuery_NotCondition(t *testing.T)
func TestNotesSearchQuery_AndOrNot(t *testing.T)
func TestNotesSearchQuery_EmptyResult(t *testing.T)
```

## Expected Outcome

- `opennotes notes search query --and data.tag=workflow` - single filter
- `opennotes notes search query --and data.tag=workflow --and data.status=active` - multiple AND
- `opennotes notes search query --or data.priority=high --or data.priority=critical` - OR
- `opennotes notes search query --and data.tag=epic --not data.status=archived` - AND + NOT
- All queries use parameterized SQL (no injection possible)

## Acceptance Criteria

- [x] `search query` subcommand created
- [x] `--and`, `--or`, `--not` flags work correctly
- [x] Boolean logic combines correctly
- [x] Field whitelist enforced (security)
- [x] Value length validated (security)
- [x] Parameterized queries only (security)
- [x] 15+ tests including security tests (30 tests added)
- [x] Clear error messages for invalid input

## Implementation Notes

### Files Created/Modified

1. **cmd/notes_search_query.go** (new) - The subcommand implementation with:
   - `--and`, `--or`, `--not` StringArray flags
   - Input validation and error handling
   - Comprehensive help text with examples

2. **internal/services/search.go** (extended) - Already contained `QueryCondition`, `ParseConditions()`, `BuildWhereClause()` from Task 1. Minor lint fix applied.

3. **internal/services/search_test.go** (extended) - Added 30 new tests covering:
   - ParseConditions validation (13 tests)
   - BuildWhereClause generation (8 tests)
   - SQL injection prevention (5 tests)
   - Glob pattern conversion (3 tests)
   - Edge cases (1 test)

### Test Coverage

- 30 new tests added for boolean query functionality
- All existing tests continue to pass (161+ total tests)
- Security tests verify parameterized SQL and field whitelist
