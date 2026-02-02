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

## Issue 1: Tag Filtering Returns No Results

### Symptom

Tag-based queries return no results even when tags are present in note frontmatter:

```bash
opennotes notes search --and data.tag=work
# Expected: Notes with tag "work"
# Actual: No results
```

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

## Issue 2: Fuzzy Search Needs Tuning

### Symptom

Fuzzy search results are less accurate than expected. Some obvious matches are missed, while unexpected results appear.

### Expected Behavior

- `~term` should find close matches with spelling variations
- Fuzziness distance should balance precision vs recall
- Common typos should be caught (1-2 character edits)

### Current Configuration

From `internal/search/bleve/query.go`:
```go
// Fuzzy queries use default Bleve fuzziness (likely 1)
case *search.FuzzyQuery:
    return bleve.NewFuzzyQuery(expr.Term)
```

### Investigation Notes

**Parameter to Test**:
- Fuzziness distance (0, 1, 2) - currently uses Bleve default
- Prefix length (how many chars must match exactly)
- Max expansions (limit number of fuzzy matches)

**Comparison Needed**:
- Test with known misspellings
- Compare results at different fuzziness levels
- Benchmark performance impact

### Test Cases to Add

```go
// Test common typos (1-char distance)
func TestFuzzySearch_SingleCharTypo(t *testing.T) {
    // Index: "meeting", "important", "project"
    // Query: ~meetng, ~importnt, ~projct
    // Verify all found
}

// Test 2-char distance
func TestFuzzySearch_TwoCharTypo(t *testing.T) {
    // Index: "documentation"
    // Query: ~documntation (missing 'e')
    // Verify found
}

// Test performance with high fuzziness
func BenchmarkFuzzySearch_Fuzziness2(b *testing.B) {
    // Measure query time with fuzziness=2
}
```

### Resolution Path

1. **Phase 5.6**: Expose fuzziness parameter in FuzzyQuery
2. Test with fuzziness values: 0, 1, 2
3. Benchmark performance at each level
4. Choose optimal default (likely fuzziness=1)
5. Consider exposing as user-configurable option
6. Add fuzzy search examples to docs

### Workaround (Until Tuned)

Use wildcard queries for partial matching:
```bash
opennotes notes search "meet*"
```

---

## Priority

Both issues are **non-blocking** for Phase 5 completion:
- Core search functionality works (text, path, title)
- Tag filtering is a convenience feature
- Fuzzy search is an enhancement over exact matching

**Recommended**: Address in Phase 5.6 (Polish & Optimization) after documentation is complete.

## References

- Manual testing session: 2026-02-02 14:32
- Task: [task-e4f7a1b3-phase54-integration-testing.md](task-e4f7a1b3-phase54-integration-testing.md)
- Code: `internal/search/bleve/mapping.go`, `internal/search/bleve/query.go`
