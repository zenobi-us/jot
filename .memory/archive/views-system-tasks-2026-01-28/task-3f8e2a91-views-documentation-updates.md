---
id: 3f8e2a91
title: Update Views Documentation with Correct DuckDB Schema
created_at: 2026-01-25T20:46:08+10:30
updated_at: 2026-01-25T21:15:00+10:30
status: completed
epic_id: epic-0fece1be
phase_id: N/A
assigned_to: unassigned
completed_at: 2026-01-25T21:15:00+10:30
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

**Successfully completed all documentation updates**:

1. ✅ Updated `docs/views-guide.md`:
   - Replaced all `data.*` references with `metadata->>'*'` syntax
   - Updated 7 field references in custom view examples
   - Added type casting example in complex view pattern
   - Fixed jq JSON access to use `.metadata.field` syntax

2. ✅ Updated `docs/views-examples.md`:
   - Replaced all `data.*` references with `metadata->>'*'` syntax
   - Updated 30+ field references across daily workflow, project management, and team collaboration examples
   - Fixed all jq examples to use correct `.metadata.field` JSON syntax (not SQL syntax)
   - Updated custom view definitions with correct schema

3. ✅ Updated `docs/views-api.md`:
   - Replaced all `data.*` references with `metadata->>'*'` syntax
   - Updated field documentation table to show `metadata->>'*'` as the correct pattern
   - Updated 20+ field references in condition examples
   - Ensured consistency across all API reference examples

**Commit**: 49bbfe8 - docs(views): update schema references from data.* to metadata->>'*'

**Verification**:
- All three files verified clean (no remaining `data.*` references)
- git diff --check passed (no whitespace issues)
- prettier formatted all files successfully
- Changes are backward-compatible (documentation only)

## Lessons Learned

1. **Different syntaxes for different contexts**: 
   - SQL queries use `metadata->>'field'` (DuckDB JSON operator)
   - jq JSON processing uses `.metadata.field` (JSON object access)
   - Must distinguish between SQL syntax and JSON tool syntax in docs

2. **Search vs Views are separate systems**:
   - Search command (`notes search`) uses `data.*` prefix intentionally (user-facing abstraction)
   - Views system uses raw DuckDB SQL with `metadata->>'*'` (direct schema access)
   - Not all `data.*` references should be changed - context matters

3. **Type casting required for numeric comparisons**:
   - Metadata fields are JSON strings by default
   - Numeric ordering requires explicit casting: `(metadata->>'priority')::INTEGER`
   - Should be documented in examples for clarity

4. **sed is efficient for bulk replacements**:
   - Used sed for systematic field name replacements across large files
   - Regex word boundaries (`\b`) prevent partial matches
   - Faster than manual Edit tool for repetitive changes

5. **Always verify jq syntax separately**:
   - Initially replaced all syntax uniformly
   - Had to revert jq examples to use JSON syntax not SQL syntax
   - Future: review jq examples separately from SQL examples

## Checklist

- [x] Update docs/views-guide.md with correct schema
- [x] Update docs/views-examples.md with correct schema
- [x] Update docs/views-api.md with correct schema
- [x] Verify no remaining `data.*` references (in views docs)
- [ ] Test examples against real notebook (optional - deferred)
- [x] Commit documentation updates with clear message

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
