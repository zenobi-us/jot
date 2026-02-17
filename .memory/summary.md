# OpenNotes Project Summary

## Current Focus
- **Active Research**: [research-b4e2f7a1](research-b4e2f7a1-dsl-based-views-design.md) - DSL-based views design
  - Phases 1-3 complete: DSL pipeline mapped, 3 design options explored, CLI surface designed
  - **Selected Design**: Option C (Hybrid — Pipe Syntax) with full CLI spec
  - Next: Phase 4 (architecture validation), Phase 5 (SQL cleanup plan), Phase 6 (implementation plan)
- **Epic**: Residual cleanup from DuckDB removal (epic-f661c068)

## Key Research Findings
- The Participle DSL parser is fully functional but has **zero production callers** (`FindByQueryString` is unused)
- `view.go` has ~600 lines of dead SQL code and ~400 lines of reusable code (template resolution, view discovery, parameter handling)
- **Pipe syntax selected**: `filter DSL | directives` — single query string with visual separator
- **CLI design complete**: `notes view` supports execute, `--save`, `--delete`, `--list`; pipe syntax also works in `notes search`
- `ViewDefinition.Query` becomes plain `string` (replaces SQL-oriented `ViewQuery` struct)
- CLI flags override query directives; notebook views override global; global overrides built-in
- Three DSL grammar gaps remain: existence checks (`has:`/`missing:`) critical, wildcard values critical, OR syntax medium

## Recently Completed
- ✅ **Semantic Search Enhancement** (epic-7c9d2e1f) - Archived Feb 17, 2026
- ✅ **DuckDB Removal** (epic-f661c068) - Core completed, views cleanup remaining

## Project State
- Feature branch: `feat/remove-duckdb-migrate-to-afero-chromedb-with-bleve-search`
- Ready for: Research phases 4-6 → implementation → PR review and merge to main
