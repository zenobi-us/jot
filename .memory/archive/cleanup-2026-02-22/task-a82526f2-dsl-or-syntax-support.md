---
id: a82526f2
title: Add OR syntax support to query DSL grammar
created_at: 2026-02-20T19:12:00+10:30
updated_at: 2026-02-21T21:45:00+10:30
status: completed
epic_id: f661c068
phase_id: b4e2f7a1-plan
assigned_to: null
---

# Add OR syntax support to query DSL grammar

## Objective

Implement user-facing OR syntax in the parser grammar so DSL queries can express disjunctions directly, matching parser capabilities noted as deferred.

## Related Story

- Plan deferred item #3 in [plan-b4e2f7a1-dsl-views-implementation.md](plan-b4e2f7a1-dsl-views-implementation.md)
- Context from [research-b4e2f7a1-dsl-based-views-design.md](research-b4e2f7a1-dsl-based-views-design.md)

## Steps

1. Specify CLI/DSL OR syntax (e.g., `term1 OR term2`, or alternative token form) and ambiguity rules.
2. Update parser grammar and conversion code in `internal/search/parser/`.
3. Extend Bleve translation and execution tests to verify OR semantics.
4. Add precedence tests with AND/NOT and quoted terms.
5. Update DSL docs with canonical OR examples.

## Expected Outcome

Queries and views can express OR conditions reliably and consistently across `notes search` and `notes view`.

## Actual Outcome

Not started.

## Lessons Learned

Pending implementation.
## Completion Notes (2026-02-21)

- Extended parser grammar/AST to understand `OR` with proper AND precedence, introduced `search.AndExpr`, and added Bleve translation support.
- Added parser + translator tests validating simple OR, chained OR, and AND precedence.
- Updated CLI docs (`docs/commands/notes-search.md`) to describe OR syntax and examples.
- Targeted tests: `go test ./internal/search/parser -run TestParser_Parse_OrExpressions` and `go test ./internal/search/bleve -run TestTranslateQuery_AndExpr` (run via `mise exec go -- ...`).
