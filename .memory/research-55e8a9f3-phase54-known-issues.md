---
id: 55e8a9f3
title: Phase 5.4 Known Issues - Tag Filtering & Fuzzy Search
created_at: 2026-02-02T18:44:00+10:30
updated_at: 2026-02-02T18:44:00+10:30
status: documented
epic_id: f661c068
phase_id: 02df510c
---

# Phase 5.4 Known Issues - Tag Filtering & Fuzzy Search

## Overview

During Phase 5.4 Integration & Testing (manual CLI testing), two issues were identified that need investigation and resolution in future phases.

## Issue 1: Tag Filtering - RESOLVED ✅

### Status: **NOT A BUG** - Tag filtering works correctly

### Original Symptom

Tag-based queries were reported to return no results even when tags are present in note frontmatter.

### Investigation Results (2026-02-02 19:20)

**Test Setup**:
```bash
# Created test notes with tags
note1.md: tags: [work, urgent]
note2.md: tags: [personal, planning]
```

**Test Results**:
```bash
opennotes notes search query --and data.tag=work
# ✅ WORKS: Found 1 note (note1.md)

opennotes notes search query --and data.tag=personal
# ✅ WORKS: Found 1 note (note2.md)
```

### Root Cause

**False alarm** - Tag filtering works correctly. The original issue was likely:
1. **Syntax confusion**: Need to use `notes search query --and` (not just `--and`)
2. **Indexing delay**: Index might not have been created/updated
3. **YAML format**: Tags must be YAML arrays: `tags: [work, personal]`

### Resolution

No code changes needed. Tag filtering is working as designed:
- Bleve indexes tags using `SimpleAnalyzer` (lowercase, no stemming)
- Query translation uses `MatchQuery` for tag fields
- Array fields (tags) are properly handled by Bleve

### Expected Behavior

- Tags in frontmatter should be indexed and searchable
- `data.tag=value` should match notes with that tag
- Multiple tags should be searchable independently

### Investigation Notes

**Hypothesis 1: Array Indexing Issue**
- Tags are stored as YAML arrays in frontmatter: `tags: [work, personal]`
- Bleve may not be indexing array fields correctly
- Check `internal/search/bleve/mapping.go` - field mapping for metadata

**Hypothesis 2: Field Name Mismatch**
- Query uses `data.tag` but Bleve index might use different field name
- Check if frontmatter parser extracts `tags` vs `tag`
- Verify SearchService.BuildQuery() metadata field translation

**Hypothesis 3: Tokenization**
- Tag values might be getting tokenized incorrectly
- Should use KeywordAnalyzer for exact tag matching
- Current mapping might be using StandardAnalyzer

### Test Cases to Add

```go
// Test tag array indexing
func TestBleveIndex_TagArraySupport(t *testing.T) {
    // Create note with tags: [work, personal]
    // Query for data.tag=work
    // Verify note is found
}

// Test multiple tag values
func TestBleveIndex_MultipleTagMatching(t *testing.T) {
    // Create notes with different tag combinations
    // Query each tag independently
    // Verify correct notes returned
}
```

### Resolution Path

1. **Phase 5.6**: Investigate Bleve metadata field mapping
2. Add debug logging for indexed metadata fields
3. Verify frontmatter parser extracts tags correctly
4. Update mapping if needed (KeywordAnalyzer for tags)
5. Add comprehensive tag search tests
6. Document tag query syntax in user docs

### Workaround (Until Fixed)

Use text search for tag-like keywords in note body:
```bash
opennotes notes search "#work"
```

---

## Issue 2: Fuzzy Search - Parser Support Missing

### Status: **FEATURE GAP** - CLI flag works, parser syntax doesn't

### Symptom

The query parser syntax `~term` is not supported. Only the `--fuzzy` flag works.

### Current Behavior

**Works ✅**:
```bash
opennotes notes search "projct" --fuzzy
# Successfully finds "project" with fuzzy matching
```

**Does NOT work ⚠️**:
```bash
opennotes notes search "~projct"
# Returns no results - parser doesn't recognize ~ prefix
```

### Root Cause

The query parser (`internal/search/parser/grammar.go`) does not include fuzzy query syntax:
- No `~` prefix in lexer rules
- No `FuzzyExpr` type in query AST
- No translation from `~term` to Bleve FuzzyQuery

### Expected Behavior

Gmail-style query syntax should support:
- `~term` - Find close matches with spelling variations
- Fuzziness distance should balance precision vs recall
- Common typos should be caught (1-2 character edits)

### Implementation Plan

**Phase 1: Add FuzzyExpr to query AST**
1. Add `FuzzyExpr` type to `internal/search/query.go`:
```go
type FuzzyExpr struct {
    Field string // Optional field qualifier
    Term  string // The term to fuzzy match
    Fuzziness int // 0, 1, or 2 (default: 1)
}
```

**Phase 2: Update parser grammar**
1. Add `~` prefix to lexer in `internal/search/parser/grammar.go`
2. Add fuzzy expression parsing rule
3. Convert grammar AST to FuzzyExpr

**Phase 3: Translate to Bleve**
1. Add FuzzyExpr case to `translateExpr()` in `internal/search/bleve/query.go`
2. Create Bleve FuzzyQuery with configurable fuzziness
3. Default fuzziness=1 for single-char typos

### Test Cases to Add

```go
// Test ~term syntax parsing
func TestParser_FuzzyQuery(t *testing.T) {
    // Parse: "~project"
    // Expected: FuzzyExpr{Term: "project", Fuzziness: 1}
}

// Test fuzzy search execution
func TestBleveIndex_FuzzySearch(t *testing.T) {
    // Index: "meeting", "project"
    // Query: ~meetng, ~projct
    // Verify both found
}
```

### Resolution Path

**Phase 5.6 Work**:
1. Add FuzzyExpr type to query AST
2. Update parser to recognize `~term` syntax
3. Implement Bleve query translation
4. Add comprehensive fuzzy search tests
5. Document fuzzy query syntax in user docs
6. Consider exposing fuzziness level: `~1:term` or `~2:term`

**Estimated Effort**: 3-4 hours
- Parser changes: 1 hour
- Query translation: 30 mins
- Tests: 1 hour
- Documentation: 30 mins
- Testing/verification: 1 hour

### Current Workarounds

**Option 1**: Use `--fuzzy` flag (works well)
```bash
opennotes notes search "projct" --fuzzy
```

**Option 2**: Use wildcard queries for partial matching
```bash
opennotes notes search "proj*"
```

---

## Priority Assessment

**Issue 1: Tag Filtering** - ✅ **RESOLVED** (not a bug)
- Status: Works correctly via `notes search query --and data.tag=value`
- Priority: None (closed)

**Issue 2: Fuzzy Parser Syntax** - ⚠️ **ENHANCEMENT**
- Status: `--fuzzy` flag works, parser syntax `~term` missing
- Impact: Low (workaround available)
- Benefit: Medium (better UX for power users)
- Effort: Medium (3-4 hours estimated)
- Priority: **Optional for Phase 5.6**

### Recommendation

**Option A**: Skip Phase 5.6 entirely
- Tag filtering works (false alarm)
- Fuzzy flag works (parser syntax is nice-to-have)
- Focus on Phase 6 (Semantic Search) instead

**Option B**: Implement fuzzy parser syntax (3-4 hours)
- Complete the query DSL with `~term` support
- Provides feature parity with Gmail-style search
- Minor improvement for power users

**Option C**: Defer to future enhancement
- Document `--fuzzy` flag usage
- Create GitHub issue for `~term` parser support
- Move directly to Phase 6

## References

- Manual testing session: 2026-02-02 14:32
- Task: [task-e4f7a1b3-phase54-integration-testing.md](task-e4f7a1b3-phase54-integration-testing.md)
- Code: `internal/search/bleve/mapping.go`, `internal/search/bleve/query.go`
