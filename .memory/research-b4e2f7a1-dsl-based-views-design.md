---
id: b4e2f7a1
title: "Research: DSL-based views design for opennotes"
created_at: 2026-02-17T18:47:00+10:30
updated_at: 2026-02-17T18:54:00+10:30
status: todo
epic_id: f661c068
phase_id: null
related_task_id: null
assigned_to: unassigned
---

# Research: DSL-based views design for opennotes

## Research Questions

### 1. How can builtin views be expressed as DSL queries?
The search DSL (`internal/search/parser/`) supports Gmail-style queries with fields: `tag`, `title`, `path`, `created`, `modified`, `body`, `status`. Can each builtin view be represented as a single DSL query string?

- **today**: `modified:>=2026-02-17` — straightforward date filter. Does the DSL support `>=` on date fields correctly? Does it resolve relative dates like `{{today}}`?
- **recent**: Needs all notes sorted by modified date DESC, limit 20. The DSL has no `sort` or `limit` concepts — how to handle this?
- **kanban**: Needs notes grouped by `status` field. The DSL has no `group by` concept — is this even expressible, or does it need post-query Go logic?
- **untagged**: Needs notes where tags are empty/nil. The DSL supports `tag:value` but not "tag is absent" (`-tag:*` or similar). What negation patterns exist?

### 2. What gaps exist in the current DSL grammar?
Current DSL supports: field matching, comparison operators (`>`, `>=`, `<`, `<=`), negation (`-term`, `-field:value`), quoted strings, date literals, free text.

Missing for views:
- **Sorting**: No `sort:field` or `order:desc` syntax
- **Limits**: No `limit:N` syntax
- **Grouping**: No `group:field` syntax (needed for kanban)
- **Wildcard negation**: Can you express "field has no value" (e.g., `tag:*` then negate it)?
- **Relative dates**: Template variables like `{{today}}` — should these be in the DSL or resolved before parsing?
- **Existence checks**: "has:tag" or "missing:tag" patterns

Should these be added to the DSL grammar, or handled as view-level metadata outside the query string?

### 3. How should ViewDefinition/ViewQuery types be redesigned?
Current `core.ViewQuery` has SQL-oriented fields: `SelectColumns`, `GroupBy`, `Having`, `AggregateColumns`, `Distinct`. These are dead code.

Proposed direction: A view could be a DSL query string + optional metadata (sort, limit, group). Questions:
- Should `ViewDefinition.Query` just be a `string` (the DSL query)?
- Or a struct with `Query string`, `Sort string`, `Limit int`, `GroupBy string`?
- How do parameters (`{{today}}`, `{{param_name}}`) get resolved — before or after DSL parsing?
- Should `ViewCondition` be removed entirely in favor of raw DSL strings?

### 4. How should user-created custom views (saved queries) work?
Users should be able to save DSL queries as named views. Questions:
- Where are custom views stored? In notebook `.opennotes.json`? In global config? Both (with precedence)?
- What's the format? `{ "name": "my-view", "query": "tag:work status:todo", "sort": "modified:desc", "limit": 50 }`?
- Can users override builtin views?
- How does `opennotes notes view --list` discover and display custom vs builtin views?
- Should there be a `opennotes notes view --save <name> <query>` command?

### 5. What about graph-based views (orphans, broken-links)?
These views require traversing note links — they can't be expressed as search queries. Questions:
- Keep them as special cases in `view_special.go`? (Current approach, already working)
- Should they be flagged differently in `ViewDefinition` (e.g., `type: "special"` vs `type: "query"`)?
- Could they eventually be index-backed (index link relationships in Bleve)?

### 6. How to clean up dead SQL code?
`internal/services/view.go` has ~1000+ lines including `GenerateSQL()`, SQL validation, aggregate function handling. Questions:
- Remove all SQL code in one pass or incrementally?
- What code is still useful? (Template variable resolution `{{today}}`, parameter handling, view discovery/loading)
- What tests in `view_test.go` test SQL behavior vs. general view logic?

## Summary

_To be filled after research is complete._

## Findings

_To be filled during research._

## References

- DSL grammar: `internal/search/parser/grammar.go`
- DSL parser: `internal/search/parser/parser.go`
- Search service (query building): `internal/services/search.go`
- View service: `internal/services/view.go`
- View types: `internal/core/view.go`
- Special views: `internal/services/view_special.go`
- View command: `cmd/notes_view.go`
