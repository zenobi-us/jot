---
id: s1a00004
title: Testing, Performance, and Documentation
created_at: 2026-01-22T12:55:00+10:30
updated_at: 2026-01-22T12:55:00+10:30
status: todo
epic_id: 3e01c563
phase_id: 4a8b9c0d
assigned_to: unassigned
estimated_hours: 1.5
depends_on: s1a00003
---

# Task: Testing, Performance, and Documentation

## Objective

Complete test coverage, verify performance targets, and update documentation for the Note Search Enhancement feature.

## Steps

### 1. Performance Benchmarks

Create `internal/services/search_bench_test.go`:

```go
func BenchmarkFuzzySearch_100Notes(b *testing.B) {
    notes := createTestNotes(100)
    service := NewSearchService(nil, nil)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.FuzzySearch("meeting", notes)
    }
}

func BenchmarkFuzzySearch_1kNotes(b *testing.B) {
    notes := createTestNotes(1000)
    service := NewSearchService(nil, nil)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.FuzzySearch("meeting", notes)
    }
}

func BenchmarkFuzzySearch_10kNotes(b *testing.B) {
    notes := createTestNotes(10000)
    service := NewSearchService(nil, nil)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.FuzzySearch("meeting", notes)
    }
    // Target: < 50ms
}

func BenchmarkBooleanQuery_Simple(b *testing.B) {
    notebook := setupBenchmarkNotebook(10000)
    conditions := []QueryCondition{
        {Type: "and", Field: "data.tag", Value: "workflow"},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        NoteService.SearchWithConditions(notebook, conditions)
    }
    // Target: < 20ms
}

func BenchmarkBooleanQuery_Complex(b *testing.B) {
    notebook := setupBenchmarkNotebook(10000)
    conditions := []QueryCondition{
        {Type: "and", Field: "data.tag", Value: "workflow"},
        {Type: "and", Field: "links-to", Value: "epics/**/*.md"},
        {Type: "not", Field: "data.status", Value: "archived"},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        NoteService.SearchWithConditions(notebook, conditions)
    }
    // Target: < 100ms
}

func BenchmarkLinkQuery_LinksTo(b *testing.B) {
    notebook := setupBenchmarkNotebook(10000) // with 50k links
    conditions := []QueryCondition{
        {Type: "and", Field: "links-to", Value: "epics/**/*.md"},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        NoteService.SearchWithConditions(notebook, conditions)
    }
    // Target: < 50ms
}
```

### 2. Integration Tests

Create `tests/e2e/search_test.go`:

```go
func TestE2E_TextSearch(t *testing.T) {
    notebook := setupE2ENotebook(t)
    
    // Create test notes
    createNote(t, notebook, "meeting-notes.md", "Team meeting discussion")
    createNote(t, notebook, "project-plan.md", "Project planning document")
    
    // Test text search
    output := runCommand(t, "opennotes", "notes", "search", "meeting")
    assert.Contains(t, output, "meeting-notes.md")
    assert.NotContains(t, output, "project-plan.md")
}

func TestE2E_FuzzySearch(t *testing.T) {
    notebook := setupE2ENotebook(t)
    
    createNote(t, notebook, "meeting-notes.md", "Team meeting")
    createNote(t, notebook, "morning-standup.md", "Morning sync")
    
    // Fuzzy search for "mtng"
    output := runCommand(t, "opennotes", "notes", "search", "--fuzzy", "mtng")
    assert.Contains(t, output, "meeting-notes.md") // Should match
}

func TestE2E_BooleanQuery(t *testing.T) {
    notebook := setupE2ENotebook(t)
    
    createNote(t, notebook, "epic1.md", map[string]interface{}{
        "tag": "epic", "status": "active",
    })
    createNote(t, notebook, "epic2.md", map[string]interface{}{
        "tag": "epic", "status": "archived",
    })
    
    // Test AND + NOT
    output := runCommand(t, "opennotes", "notes", "search", "query",
        "--and", "data.tag=epic",
        "--not", "data.status=archived")
    
    assert.Contains(t, output, "epic1.md")
    assert.NotContains(t, output, "epic2.md")
}

func TestE2E_LinkQuery(t *testing.T) {
    notebook := setupE2ENotebook(t)
    
    createNote(t, notebook, "epic.md", map[string]interface{}{
        "links": []string{"tasks/task1.md"},
    })
    createNote(t, notebook, "tasks/task1.md", nil)
    
    // Test links-to
    output := runCommand(t, "opennotes", "notes", "search", "query",
        "--and", "links-to=tasks/task1.md")
    
    assert.Contains(t, output, "epic.md")
}
```

### 3. Error Scenario Tests

```go
func TestErrorHandling_InvalidField(t *testing.T) {
    output, err := runCommandWithError("opennotes", "notes", "search", "query",
        "--and", "invalid.field=value")
    
    assert.Error(t, err)
    assert.Contains(t, output, "invalid field")
}

func TestErrorHandling_InvalidFormat(t *testing.T) {
    output, err := runCommandWithError("opennotes", "notes", "search", "query",
        "--and", "no-equals-sign")
    
    assert.Error(t, err)
    assert.Contains(t, output, "expected field=value")
}

func TestErrorHandling_ValueTooLong(t *testing.T) {
    longValue := strings.Repeat("a", 2000)
    output, err := runCommandWithError("opennotes", "notes", "search", "query",
        "--and", "data.tag="+longValue)
    
    assert.Error(t, err)
    assert.Contains(t, output, "too long")
}
```

### 4. Documentation Updates

#### Update docs/commands/notes-search.md

```markdown
# Notes Search Command

## Overview

Search notes using text search, fuzzy matching, or boolean queries.

## Text Search

\`\`\`bash
# Search for exact text
opennotes notes search "meeting"

# List all notes
opennotes notes search

# Fuzzy matching (ranks by similarity)
opennotes notes search --fuzzy "mtng"
\`\`\`

## Boolean Queries

\`\`\`bash
# AND conditions (all must match)
opennotes notes search query --and data.tag=workflow --and data.status=active

# OR conditions (any can match)
opennotes notes search query --or data.priority=high --or data.priority=critical

# NOT conditions (exclude matches)
opennotes notes search query --and data.tag=epic --not data.status=archived
\`\`\`

## Link Queries

\`\`\`bash
# Find notes that link TO a document
opennotes notes search query --and links-to=docs/architecture.md

# Find notes that a document links TO
opennotes notes search query --and linked-by=planning/q1.md

# Glob patterns
opennotes notes search query --and links-to=epics/**/*.md
\`\`\`

## Supported Fields

- `data.tag`, `data.tags` - Note tags
- `data.status` - Status field
- `data.priority` - Priority field
- `data.assignee` - Assignee field
- `data.author` - Author field
- `data.type` - Type field
- `data.category` - Category field
- `data.project` - Project field
- `links-to` - Documents linking to target
- `linked-by` - Documents linked from source

## Glob Patterns

| Pattern | Meaning | Example |
|---------|---------|---------|
| `*` | Any characters | `epics/*.md` |
| `**` | Any path depth | `**/*.md` |
| `?` | Single character | `task?.md` |
```

#### Update CLI help text

In `cmd/notes_search.go`:

```go
Short: "Search notes with text, fuzzy matching, or boolean queries",
Long: `Search notes using various methods:

  Text Search:
    opennotes notes search "meeting"        # Exact text search
    opennotes notes search --fuzzy "mtng"   # Fuzzy matching

  Boolean Queries:
    opennotes notes search query --and data.tag=workflow
    opennotes notes search query --or data.priority=high --or data.priority=critical
    opennotes notes search query --and data.tag=epic --not data.status=archived

  Link Queries:
    opennotes notes search query --and links-to=docs/architecture.md
    opennotes notes search query --and linked-by=planning/q1.md`,
```

### 5. Verify Coverage

```bash
# Run tests with coverage
mise run test -- -coverprofile=coverage.out ./internal/services/

# Check coverage percentage
go tool cover -func=coverage.out | grep total

# Target: ≥85% for search.go
```

## Expected Outcome

- All performance benchmarks pass targets
- ≥85% test coverage for search functionality
- Comprehensive E2E tests
- Updated documentation
- Clear error messages

## Acceptance Criteria

- [ ] Performance benchmarks created and passing
  - [ ] Fuzzy search < 50ms for 10k notes
  - [ ] Simple query < 20ms for 10k notes
  - [ ] Complex query < 100ms for 10k notes
  - [ ] Link query < 50ms for 10k notes + 50k links
- [ ] E2E tests cover all command variations
- [ ] Error handling tests verify clear messages
- [ ] Documentation complete:
  - [ ] Command reference updated
  - [ ] CLI help text updated
  - [ ] Examples provided
- [ ] Test coverage ≥85% for search.go
- [ ] All 161+ existing tests still pass (no regressions)
