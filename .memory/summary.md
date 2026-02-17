# OpenNotes Project Summary

## Current Focus
- **Active Research**: [research-b4e2f7a1](research-b4e2f7a1-dsl-based-views-design.md) - Researching how to leverage the existing search DSL for the view system, including saved queries for user-created custom views
- **Epic**: Residual cleanup from DuckDB removal (epic-f661c068)

## Research: DSL-Based Views
The `notes view` command is non-functional (hardcoded SQL error). Instead of naively replacing SQL with `getAllNotes()` + Go filtering, we're researching how to leverage the existing Participle-based search DSL (`internal/search/parser/`) to power views. Key questions: expressing builtin views as DSL queries, supporting user-created saved queries, identifying DSL grammar gaps (sorting, limits, grouping), and redesigning `ViewDefinition`/`ViewQuery` types. Graph-based views (orphans, broken-links) remain as special cases.

## Recently Completed
- ✅ **Semantic Search Enhancement** (epic-7c9d2e1f) - Archived Feb 17, 2026
- ✅ **DuckDB Removal** (epic-f661c068) - Core completed, views cleanup remaining

## Project State
- Feature branch: `feat/remove-duckdb-migrate-to-afero-chromedb-with-bleve-search`
- Ready for: Research completion → implementation → PR review and merge to main
