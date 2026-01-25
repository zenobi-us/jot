---
id: 3f8e2a91
title: Update Views Documentation with Correct DuckDB Schema
created_at: 2026-01-25T20:46:08+10:30
updated_at: 2026-01-25T20:46:08+10:30
status: todo
epic_id: epic-0fece1be
phase_id: N/A
assigned_to: unassigned
---

# Update Views Documentation with Correct DuckDB Schema

## Objective

Update all views documentation files to reflect the correct DuckDB metadata schema after fixing built-in view definitions. The fix changed from using `data.*` references to `metadata->>'*'` JSON operators.

## Context

**Related Task**: task-b2d67264 (Views Fault Tolerance Investigation)  
**Fix Commit**: 5da5fe9 - fix(views): correct DuckDB metadata schema for all built-in views (#11)

**Changes Made in Fix**:
- `today`: `data.created` → `metadata->>'created_at'`
- `recent`: `updated DESC` → `metadata->>'updated_at' DESC`
- `kanban`: `data.status`, `data.priority` → `metadata->>'status'`, `(metadata->>'priority')::INTEGER`
- `untagged`: `data.tags` → `metadata->>'tags'`
- `orphans`: `created DESC` → `metadata->>'created_at' DESC`
- `broken-links`: `updated DESC` → `metadata->>'updated_at' DESC`

## Steps

1. ✅ Identify documentation files needing updates:
   - `docs/views-guide.md` (18KB) - User guide with examples
   - `docs/views-examples.md` (16KB) - Real-world query examples  
   - `docs/views-api.md` (18KB) - API reference

2. ⏳ Update `docs/views-guide.md`:
   - Search for all `data.` references and replace with `metadata->>`
   - Update field reference examples in "Custom Views" section
   - Update built-in view examples with correct schema
   - Add note about DuckDB JSON operators if not present

3. ⏳ Update `docs/views-examples.md`:
   - Review all example queries for `data.*` references
   - Replace with correct `metadata->>'*'` syntax
   - Verify all examples use consistent schema
   - Add type casting examples: `(metadata->>'priority')::INTEGER`

4. ⏳ Update `docs/views-api.md`:
   - Document correct DuckDB schema in API reference
   - Update field validation section with allowed prefixes:
     - `metadata->>` (primary JSON field extraction)
     - `metadata->` (JSON object access)
     - `file_path`, `content`
     - `stats->`, `stats->>`
   - Document type casting syntax for numeric/date fields
   - Update built-in view specifications

5. ⏳ Verification:
   - Search all docs for remaining `data.` references
   - Verify consistency across all three files
   - Check that examples match actual view implementation

6. ⏳ Testing (optional):
   - Run examples against test notebook
   - Verify all documented queries execute successfully

## Expected Outcome

All views documentation files correctly reference the DuckDB metadata schema:
- No `data.*` references remain
- All examples use `metadata->>'field_name'` syntax
- Type casting documented for numeric/date fields
- Field validation documented with allowed prefixes
- Documentation matches actual implementation

## Actual Outcome

[To be filled after completion]

## Lessons Learned

[To be filled after completion]

## Checklist

- [ ] Update docs/views-guide.md with correct schema
- [ ] Update docs/views-examples.md with correct schema
- [ ] Update docs/views-api.md with correct schema
- [ ] Verify no remaining `data.*` references
- [ ] Test examples against real notebook
- [ ] Commit documentation updates with clear message

## Estimated Time

**30-45 minutes total**:
- views-guide.md: 10-15 minutes
- views-examples.md: 10-15 minutes
- views-api.md: 10-15 minutes

## Priority

**Medium** - Documentation accuracy is important but not blocking functionality. The views system works correctly after the fix; documentation just needs to catch up.

## Dependencies

**Prerequisites**:
- ✅ Views fix commit merged (5da5fe9)
- ✅ All tests passing with new schema

**Blockers**: None

---

**Created**: 2026-01-25 20:46 GMT+10:30  
**Type**: Documentation Task  
**Related Epic**: Storage Abstraction Layer (background research context)  
**Related Task**: task-b2d67264 (Views fault tolerance investigation - COMPLETED)
