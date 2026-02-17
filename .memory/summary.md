# OpenNotes Project Summary

## Current Focus
- **Active Research**: [research-b4e2f7a1](research-b4e2f7a1-dsl-based-views-design.md) - DSL-based views design
  - Phases 1-2 complete: DSL pipeline mapped, 3 design options explored
  - **Recommendation**: Option A (Thin Views — DSL query string + metadata sidecar)
  - Next: Phase 3 (CLI surface design), Phase 4 (architecture validation)
- **Epic**: Residual cleanup from DuckDB removal (epic-f661c068)

## Key Research Findings
- The Participle DSL parser is fully functional but has **zero production callers** (`FindByQueryString` is unused)
- `view.go` has ~600 lines of dead SQL code and ~400 lines of reusable code (template resolution, view discovery, parameter handling)
- Views should be DSL query strings + metadata sidecar (sort/limit/group) — no grammar changes needed except `has:`/`missing:` existence checks
- Three DSL grammar gaps identified: existence checks (critical), wildcard values (critical), OR syntax (medium)

## Recently Completed
- ✅ **Semantic Search Enhancement** (epic-7c9d2e1f) - Archived Feb 17, 2026
- ✅ **DuckDB Removal** (epic-f661c068) - Core completed, views cleanup remaining

## Project State
- Feature branch: `feat/remove-duckdb-migrate-to-afero-chromedb-with-bleve-search`
- Ready for: Research phases 3-6 → implementation → PR review and merge to main
