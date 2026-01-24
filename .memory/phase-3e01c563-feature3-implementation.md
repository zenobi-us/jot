# Quick-Start Implementation Guide

**Epic**: 3e01c563 - Advanced Note Creation and Search Capabilities  
**Based on**: Research Document `.memory/research-3e01c563-advanced-operations.md`

---

## Implementation Order

### ✅ Phase 1: Dynamic Flag Parsing (3-4 hours)

**Priority**: CRITICAL - Foundation for advanced creation

**Task 1.1: Add --data flag parsing** (60 min)
```go
// cmd/notes_add.go
var dataFields []string
addCmd.Flags().StringArrayVar(&dataFields, "data", []string{}, 
    "Set frontmatter field (format: --data field=value, repeatable)")
```

**Task 1.2: Implement parser** (90 min)
```go
func parseDataFlags(flags []string) (map[string]interface{}, error) {
    result := make(map[string]interface{})
    for _, flag := range flags {
        parts := strings.SplitN(flag, "=", 2)
        // Handle multi-value fields as arrays
        // Validate field names
    }
    return result, nil
}
```

**Task 1.3: Integration + Tests** (60 min)
- Integrate with note creation
- Add validation tests
- Add integration tests

**Validation**: `opennotes note add "test" --data tag=one --data tag=two`

---

### ✅ Phase 2: Boolean Query Construction (4-5 hours)

**Priority**: CRITICAL - Advanced search capability

**Task 2.1: Add search condition flags** (60 min)
```go
var andFlags []string
var orFlags []string
var notFlags []string

searchCmd.Flags().StringArrayVar(&andFlags, "and", []string{}, "AND condition")
searchCmd.Flags().StringArrayVar(&orFlags, "or", []string{}, "OR condition")
searchCmd.Flags().StringArrayVar(&notFlags, "not", []string{}, "NOT condition")
```

**Task 2.2: Query builder** (120 min)
```go
func buildSearchQuery(query SearchQuery) (string, []interface{}, error) {
    // Build WHERE clauses with parameterized queries
    // ALWAYS use ? placeholders
    // Whitelist field names
}
```

**Task 2.3: Security validation** (60 min)
```go
func validateFieldName(field string) error {
    // Whitelist: data, body, title, path
}

func validateOperator(op string) error {
    // Whitelist: =, !=, LIKE, GLOB, IN, etc.
}
```

**Task 2.4: Integration + Tests** (90 min)
- NoteService.SearchWithQuery()
- SQL injection prevention tests
- Complex query tests

**Validation**: `opennotes note search --and data.tag workflow --not data.status archived`

---

### ✅ Phase 3: Built-in Views (3-4 hours)

**Priority**: HIGH VALUE - User productivity

**Task 3.1: ViewService implementation** (90 min)
```go
type ViewDefinition struct {
    Name        string
    Description string
    Query       SearchQuery
}

var BuiltInViews = map[string]ViewDefinition{
    "today": { /* ... */ },
    "recent": { /* ... */ },
    "drafts": { /* ... */ },
}
```

**Task 3.2: View resolution** (60 min)
```go
func (vs *ViewService) GetView(name string) (*ViewDefinition, error)
func (vs *ViewService) ListViews() []ViewInfo
```

**Task 3.3: CLI integration** (60 min)
```go
searchCmd.Flags().StringVar(&viewName, "view", "", "Use predefined view")
```

**Task 3.4: Tests** (60 min)
- Built-in view tests
- View resolution tests
- CLI integration tests

**Validation**: `opennotes note search --view today`

---

### ✅ Phase 4: FZF Integration (2-3 hours)

**Priority**: UX ENHANCEMENT - Nice-to-have

**Task 4.1: Add go-fuzzyfinder dependency** (15 min)
```bash
go get github.com/ktr0731/go-fuzzyfinder
```

**Task 4.2: Interactive search function** (90 min)
```go
func selectNoteFuzzy(notes []Note) (*Note, error) {
    idx, err := fuzzyfinder.Find(
        notes,
        func(i int) string { return notes[i].Title },
        fuzzyfinder.WithPreviewWindow(previewFunc),
    )
    return &notes[idx], nil
}
```

**Task 4.3: CLI flag + fallback** (45 min)
```go
searchCmd.Flags().BoolVar(&useFzf, "fzf", false, "Interactive fuzzy finder")

if useFzf {
    if !isInteractive() {
        // Fallback to normal search
    }
    return runInteractiveSearch()
}
```

**Task 4.4: Tests** (30 min)
- Mock FZF for tests
- Fallback logic tests

**Validation**: `opennotes note search --fzf`

---

## Testing Checklist

### Unit Tests
- ✅ parseDataFlags() - field=value parsing
- ✅ validateFieldName() - whitelist validation
- ✅ buildSearchQuery() - query construction
- ✅ ViewService.GetView() - view resolution

### Integration Tests
- ✅ End-to-end note creation with --data
- ✅ Boolean search with multiple conditions
- ✅ View-based search
- ✅ FZF interactive search

### Security Tests
- ✅ SQL injection prevention
- ✅ Invalid field name rejection
- ✅ Invalid operator rejection
- ✅ Input length validation

---

## Code Locations

### New Files to Create
```
internal/services/view_service.go       # View management
internal/services/query_builder.go      # Boolean query construction
internal/services/view_service_test.go  # View tests
internal/services/query_builder_test.go # Query tests
```

### Files to Modify
```
cmd/notes_add.go           # Add --data flag
cmd/notes_search.go        # Add boolean flags, --view, --fzf
internal/services/note.go  # Add SearchWithQuery(), SearchWithView()
```

---

## Common Pitfalls to Avoid

### ❌ Security Anti-Patterns
```go
// NEVER concatenate user input
query := "SELECT * FROM notes WHERE " + field + " = '" + value + "'"

// NEVER allow arbitrary column names
query := fmt.Sprintf("SELECT * FROM notes WHERE %s = ?", userField)
```

### ✅ Correct Patterns
```go
// ALWAYS use parameterized queries
query := "SELECT * FROM notes WHERE title = ?"
db.Query(query, userValue)

// ALWAYS whitelist field names
if !isValidField(field) {
    return errors.New("invalid field")
}
```

---

## Performance Targets

| Operation | Target |
|-----------|--------|
| Parse --data flags | < 1ms |
| Build boolean query | < 5ms |
| Execute simple search | < 10ms |
| Execute complex search (5 conditions) | < 100ms |
| FZF interactive load | < 50ms |
| View resolution | < 5ms |

---

## Rollout Strategy

### Week 1: Dynamic Flags + Boolean Search
- Implement core flag parsing
- Implement query construction
- Complete security validation
- Comprehensive testing

### Week 2: Views + FZF
- Implement ViewService
- Add built-in views
- FZF integration
- Documentation

### Week 3: Polish + Release
- Performance optimization
- Error message refinement
- User documentation
- Release notes

---

## Documentation to Create

1. **User Guide**: Advanced note creation examples
2. **User Guide**: Boolean search syntax reference
3. **User Guide**: Built-in views list
4. **User Guide**: FZF usage
5. **Developer Guide**: Security validation patterns
6. **Developer Guide**: Query builder architecture

---

## Success Metrics

**Functional**:
- ✅ All 4 features working end-to-end
- ✅ Zero SQL injection vulnerabilities
- ✅ Cross-platform compatibility

**Quality**:
- ✅ ≥85% test coverage
- ✅ All linting passing
- ✅ Performance targets met

**User Experience**:
- ✅ Clear error messages
- ✅ Comprehensive examples
- ✅ Intuitive CLI flags

---

## References

**Full Research**: `.memory/research-3e01c563-advanced-operations.md`  
**Quick Summary**: `.memory/research-3e01c563-summary.md`  
**Epic Document**: `.memory/epic-3e01c563-advanced-note-operations.md`

---

**Ready to Start**: All research complete, implementation path clear ✅
