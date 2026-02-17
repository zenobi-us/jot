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

## Execution Plan

### Phase 1: Map the existing DSL pipeline (codemapper)

Use the `codemapper` skill to build a complete picture of how queries flow through the system.

**Actions:**
- `cm trace "cmd" "search"` — trace how CLI commands reach the search service
- `cm query "Parse" --format ai` — find all Parse-related symbols in the parser package
- `cm query "BuildQuery" --format ai` — understand how QueryConditions become search.Query AST
- `cm callers "SearchService.BuildQuery"` — who calls BuildQuery today?
- `cm query "ViewDefinition" --format ai` — map all view type usage
- `cm stats internal/services/view.go` — quantify dead SQL code vs. reusable code
- `cm trace "internal/search" "internal/services"` — map dependencies between search and services

**Deliverable:** A clear map of: DSL string → parser → search.Query AST → Bleve query → results. Identify exactly which connection points the view system needs to plug into.

### Phase 2: Explore design options (brainstorming)

Use the `brainstorming` skill to explore the design space before committing to an approach.

**Key design tensions to explore:**
- **DSL purity vs. view metadata**: Should sort/limit/group be DSL syntax extensions, or separate view-level config alongside a query string?
- **Thin views vs. rich views**: Is a view just a saved query string, or does it need its own execution model?
- **Grammar extension cost**: What's the impact of adding `sort:`, `limit:`, `group:` to the Participle grammar vs. keeping them out?
- **Template resolution timing**: Resolve `{{today}}` before DSL parsing (simple string substitution) or make the parser date-aware?
- **User experience**: `notes view today` vs. `notes search "modified:>=today"` — what's the value-add of named views?

**Deliverable:** 2-3 concrete design options with tradeoffs articulated.

### Phase 3: Design the CLI surface (creating-cli-tools)

Use the `creating-cli-tools` skill to design the user-facing commands.

**Questions to resolve:**
- How does `notes view <name>` execute a saved query?
- How does `notes view --save <name> "<query>"` work?
- What flags does `notes view` need? (`--format`, `--sort`, `--limit`, `--group`?)
- How do users list, edit, delete saved views?
- How do parameters work? (`notes view my-view --param sprint=Q1`)
- Output format: reuse `notes search` output rendering or separate?

**Deliverable:** CLI specification for the view command family.

### Phase 4: Validate the design (architect-reviewer)

Use the `architect-reviewer` skill to validate the chosen approach.

**Review criteria:**
- Does the design compose well with existing search infrastructure?
- Is it backwards-compatible with the working `orphans`/`broken-links` views?
- Does it scale to complex user queries without becoming another SQL?
- Is the ViewDefinition type clean and minimal?
- Are there edge cases that break the model?

**Deliverable:** Go/no-go recommendation with any design adjustments.

### Phase 5: Plan the cleanup (refactoring-specialist)

Use the `refactoring-specialist` skill to plan safe removal of dead SQL code.

**Actions:**
- Identify which code in `view.go` is reusable (template resolution, view discovery, parameter handling)
- Identify which code is dead SQL (GenerateSQL, SQL validation, aggregate functions, escapeSQL, etc.)
- Plan incremental removal strategy (don't break tests that test non-SQL behavior)
- Identify which tests in `view_test.go` are SQL-specific vs. general

**Deliverable:** A file-by-file cleanup plan with safe removal order.

### Phase 6: Write implementation plan (writing-plans)

Use the `writing-plans` skill to produce the final implementation task(s).

**Deliverable:** Detailed implementation plan with exact file paths, code examples, and verification steps — ready for an engineer with zero context to execute.

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
