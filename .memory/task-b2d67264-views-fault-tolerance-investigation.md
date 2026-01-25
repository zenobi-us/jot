---
id: b2d67264
title: Views Feature Fault Tolerance Investigation
created_at: 2026-01-25T19:44:20+10:30
updated_at: 2026-01-25T19:44:20+10:30
status: completed
epic_id: epic-0fece1be
phase_id: N/A
assigned_to: claude-20260125
---

# Views Feature Fault Tolerance Investigation

## Objective

Investigate the views feature querying approach to determine if fault tolerance issues are caused by:
1. DuckDB markdown extension limitations
2. Our code implementation
3. Configuration issues

## Steps

1. ✅ Set up .memory folder as a test notebook
2. ✅ Create .opennotes.json config in .memory/
3. ✅ Test built-in views with .memory folder
4. ✅ Identify specific error patterns
5. ✅ Trace error origin (DuckDB vs our code)
6. ✅ Document root cause
7. ✅ Propose fix

## Expected Outcome

Clear determination of:
- Whether fault comes from DuckDB markdown extension or our code
- Specific error messages and reproduction steps
- Recommended fix approach

## Actual Outcome

✅ **Investigation Complete** - 2026-01-25T19:46:20+10:30

### Discovery 1: View Command Not in PATH

The `opennotes notes view` command exists but wasn't available because the binary wasn't rebuilt. After running `mise run build`, the command is now functional.

### Discovery 2: Need Notebook Configuration

The .memory folder needs:
- `.opennotes.json` config file to be recognized as a notebook
- Notes directory structure

### Discovery 3: Root Cause Identified ✅

**Verdict**: This is **OUR CODE ISSUE**, not a DuckDB markdown extension limitation.

**The Problem**:
- View definitions reference `data.created`, `data.status`, `data.tags`, etc.
- But `read_markdown()` function does NOT create a `data` table or alias
- Actual schema returned by `read_markdown()`:
  - `content` (TEXT)
  - `file_path` (TEXT)
  - `metadata` (JSON) - contains frontmatter as JSON object
  - `stats` (JSON) - when `include_stats:=true` is used

**Correct Access Pattern**:
```sql
-- WRONG (current views)
WHERE data.created >= '2026-01-25'

-- CORRECT (should be)
WHERE metadata->>'created_at' >= '2026-01-25'
```

**Error Message Received**:
```
Binder Error: Referenced table "data" not found!
Candidate tables: "read_markdown"
```

**Affected Built-in Views**:
1. `today` - uses `data.created`
2. `kanban` - uses `data.status`, `data.priority`
3. `recent` - references `updated` (should be `metadata->>'updated_at'`)
4. `untagged` - uses `data.tags`

## Lessons Learned

### Technical Insights

1. **DuckDB JSON Operators**: Use `->>'field'` to extract JSON fields as text
2. **Schema Validation**: Always verify actual column names before writing queries
3. **Extension Documentation**: The `read_markdown()` function parameters are well-documented in error messages
4. **Metadata Storage**: All frontmatter is stored in a single JSON column, not flattened

### Process Insights

1. **Test with Real Data**: Views were implemented but never tested with actual notebooks
2. **E2E Testing Gaps**: No end-to-end tests for view execution
3. **Documentation Mismatch**: View examples in docs don't match actual schema

## Recommended Fix

### Approach 1: Fix View Definitions (RECOMMENDED)

Update all built-in view definitions in `internal/services/view.go` to use correct column references:

**Changes Required**:
```go
// TODAY VIEW - Fix field references
{
    Field:    "metadata->>'created_at'",  // was: data.created
    Operator: ">=",
    Value:    "{{today}}",
}

// KANBAN VIEW - Fix field references
{
    Field:    "metadata->>'status'",  // was: data.status
    Operator: "IN",
    Value:    "{{status}}",
},
OrderBy: "(metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC",  // was: data.priority DESC, updated DESC

// UNTAGGED VIEW - Fix field references
{
    Field:    "metadata->>'tags'",  // was: data.tags
    Operator: "IS NULL",
    Value:    "",
}

// RECENT VIEW - Fix orderby
OrderBy: "metadata->>'updated_at' DESC",  // was: updated DESC
```

**Field Validation Updates**:
Update `validateField()` function to allow `metadata->>` JSON operator:
```go
allowedPrefixes := []string{
    "metadata->>",  // JSON field extraction
    "path",
    "file_path",
    "content",
    "stats.",
}
```

### Approach 2: Create View Abstraction Layer (COMPLEX)

Create a mapping layer that translates `data.*` references to `metadata->>'*'` automatically. This would maintain backward compatibility but adds complexity.

**NOT RECOMMENDED** because:
- Adds unnecessary abstraction
- Makes debugging harder
- Users need to learn actual schema anyway for custom views

### Testing Requirements

1. **Unit Tests**: Test each built-in view SQL generation
2. **Integration Tests**: Execute views against test notebooks with actual markdown files
3. **E2E Tests**: Full view command execution in test environments

### Documentation Updates

1. Update `docs/views-guide.md` with correct field references
2. Update `docs/views-examples.md` with working examples
3. Update `docs/views-api.md` with DuckDB schema reference
