---
id: 3d477ab8
title: Implement Missing View System Features (GROUP BY, DISTINCT, OFFSET, HAVING, Aggregations)
created_at: 2026-01-27T22:46:00+10:30
updated_at: 2026-01-27T22:46:00+10:30
status: planning
epic_id: null
phase_id: null
assigned_to: null
priority: high
---

# Missing View System Features Implementation

## Objective

Analyze and implement missing SQL functionality in the OpenNotes view system. The view system has a solid foundation but lacks critical features for aggregations, pagination, and analytics. This task breaks down the investigation findings into an actionable implementation roadmap with 3 phases.

**Key Finding**: `ViewQuery.GroupBy` field is **defined but not implemented** in `GenerateSQL()` - a quick 4-line fix with huge impact.

## Context

### Investigation Complete ✅

Investigation branch: `fix/kanban-view`
Investigation completed: 2026-01-27
Deliverables: 4 detailed analysis documents with implementation guidance

- **kanban-groupby-statemachine.txt** (26 KB): 10-state execution flow diagram
- **missing-functionality.md** (14 KB): Comprehensive gap analysis
- **MISSING_FUNCTIONALITY_SUMMARY.txt** (22 KB): Executive summary
- **FINAL_ANALYSIS.txt** (23 KB): Complete reference guide

### Current Status

- ✅ All 161+ tests passing
- ✅ View hierarchy working (notebook > global > built-in)
- ✅ Parameters system complete (4 types)
- ✅ Basic WHERE, ORDER BY, LIMIT working
- ❌ GROUP BY missing (type defined, SQL implementation missing)
- ❌ HAVING, Aggregations, OFFSET, DISTINCT missing
- 🟠 JOIN, CTE, UNION deferred (too complex)

## Steps

### Phase 1: SQL Completeness (2 hours) - IMMEDIATE

**Task 1: GROUP BY Implementation (30 mins)** 🔴 CRITICAL
- File: `internal/services/view.go`
- Location: `GenerateSQL()` method, line 614
- Action: Add 4 lines before ORDER BY section
- Code:
  ```go
  if view.Query.GroupBy != "" {
      if err := validateField(view.Query.GroupBy); err != nil {
          return "", nil, fmt.Errorf("invalid group by field: %w", err)
      }
      query += " GROUP BY " + view.Query.GroupBy
  }
  ```
- Validation: Reuse existing `validateField()` (line 402)
- Tests: 3 new test cases
  - GROUP BY with valid field
  - GROUP BY with invalid field (injection attempt)
  - GROUP BY with ORDER BY combination
- Impact: Unlock aggregations, kanban dashboards, summaries
- Risk: 🟢 LOW (reuses existing patterns)

**Task 2: DISTINCT Support (20 mins)**
- File 1: `internal/core/view.go`
  - Add `Distinct bool` field to `ViewQuery` (line ~30)
- File 2: `internal/services/view.go`
  - Location: `GenerateSQL()` method, line 602
  - Add 1 line: Check for `Distinct` flag
- Tests: 2 new test cases
  - DISTINCT query
  - DISTINCT with WHERE
- Impact: Enable uniqueness queries
- Risk: 🟢 MINIMAL

**Task 3: OFFSET Support (1 hour)**
- File 1: `internal/core/view.go`
  - Add `Offset int` field to `ViewQuery` (line ~30)
- File 2: `internal/services/view.go`
  - Location: `GenerateSQL()` method, line 627
  - Add SQL: `LIMIT x OFFSET y` support
  - Add parameter handling for pagination
- Tests: 3 new test cases
  - OFFSET with LIMIT
  - OFFSET alone
  - Pagination calculation (page to offset)
- Impact: Enable pagination, large result sets
- Risk: 🟢 LOW

**Phase 1 Summary**:
- Total Effort: ~2 hours
- New Tests: 8 test cases
- Code Changes: 3 files, ~20 lines total
- Breaking Changes: None (all new optional fields)
- Impact: Huge (dashboards, pagination, uniqueness queries)

### Phase 2: Aggregation Support (4 hours) - NEXT SPRINT

**Task 1: HAVING Clause (1.5 hours)**
- File 1: `internal/core/view.go`
  - Add `Having []ViewCondition` to `ViewQuery`
- File 2: `internal/services/view.go`
  - Add validation: `validateHavingCondition()`
  - Add SQL generation in `GenerateSQL()`
- Tests: 4 new test cases
- Impact: Filter aggregated results
- Complexity: Medium

**Task 2: Aggregate Functions (2.5 hours)**
- File 1: `internal/core/view.go`
  - Add `SelectColumns []string` to `ViewQuery`
- File 2: `internal/services/view.go`
  - Redesign SELECT clause (not always `SELECT *`)
  - Support: COUNT, SUM, AVG, MAX, MIN
  - Implement safe column selection with validation
- Tests: 5 new test cases
- Impact: Enable analytics (counts, sums, averages)
- Complexity: High

**Phase 2 Summary**:
- Total Effort: ~4 hours
- New Tests: 9 test cases
- Code Changes: 2 files, ~50 lines
- Breaking Changes: Minor (new optional fields)
- Impact: Full analytics capabilities

### Phase 3: Enhanced Templates (2 hours) - OPTIONAL

**Task 1: Time Arithmetic (1 hour)**
- File: `internal/services/view.go`
- Enhance: `ResolveTemplateVariables()`
- Add support for:
  - `{{today-N}}` (N days ago)
  - `{{today+N}}` (N days forward)
  - `{{this_week-N}}, {{this_month-N}}`
- Tests: 3 new test cases

**Task 2: Environment Variables (45 mins)**
- File: `internal/services/view.go`
- Add parsing: `{{env:VAR}}` syntax
- Call: `os.Getenv()`
- Error handling for missing vars
- Tests: 2 new test cases

**Task 3: Period Shortcuts (30 mins)**
- Add: `{{next_week}}, {{next_month}}`
- Add: `{{end_of_month}}, {{start_of_month}}`
- Add: `{{quarter}}, {{year}}`
- Tests: 4 new test cases

**Phase 3 Summary**:
- Total Effort: ~2 hours
- New Tests: 9 test cases
- Breaking Changes: None

### Phase 4: Advanced (DEFER) - If Requirements Emerge

- ❌ JOIN Support (16+ hours, complex)
- ❌ CTE Support (12+ hours, can use views instead)
- ❌ UNION Support (10+ hours, use view composition)

## Expected Outcome

### After Phase 1 (2 hours)
✅ GROUP BY implementation working
✅ Kanban board dashboards possible (count per status)
✅ Pagination working (LIMIT + OFFSET)
✅ Unique value queries (DISTINCT)
✅ 8 new passing test cases
✅ All 169+ tests passing

### After Phase 2 (4 hours additional)
✅ HAVING clause for aggregate filtering
✅ Aggregate functions (COUNT, SUM, AVG, MAX, MIN)
✅ Full analytics capabilities
✅ 9 new passing test cases
✅ All 178+ tests passing

### After Phase 3 (2 hours additional)
✅ Time arithmetic in templates
✅ Environment variable substitution
✅ Period shortcuts
✅ Dynamic scheduling support
✅ 9 new passing test cases
✅ All 187+ tests passing

## Actual Outcome

*To be filled after execution*

## Lessons Learned

*To be filled after execution*

## Notes

- All implementation follows established patterns (reuses validation, security model)
- No breaking changes - all features are optional
- Full backward compatibility maintained
- Investigation documents available in `/tmp/` for reference
