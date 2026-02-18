# OpenNotes Project Summary

## Current Focus
- **Future Epic**: [epic-8361d3a2](epic-8361d3a2-rename-to-jot.md) - **Rename Project to "Jot"**
  - Phase 1: Discovery — Identify all rename locations (in-repo and external)
  - Status: Planning
- **Ready for Implementation**: [plan-b4e2f7a1](plan-b4e2f7a1-dsl-views-implementation.md) - DSL-based views
  - Research complete, 10-task implementation plan ready
  - **Selected Design**: Option C (Hybrid — Pipe Syntax)

## Rename Epic (8361d3a2) Overview
**Goal**: Rebrand from "OpenNotes" → "Jot"
- New binary: `jot`
- New module: `github.com/zenobi-us/jot`
- New repo: `github.com/zenobi-us/jot`
- New config: `~/.config/jot/`, `.jot.json`

## Completed Epics

### DuckDB Removal (f661c068) — ✅ Archived
**Completed**: 2026-02-02 | **Duration**: 29 hours | **Archive**: [archive/duckdb-removal-f661c068/](archive/duckdb-removal-f661c068/)

| Metric | Target | Achieved |
|--------|--------|----------|
| Binary | <15MB | 23MB (64% smaller than DuckDB) |
| Startup | <100ms | 17ms ✅ |
| Search | <25ms | 0.754ms ✅ |
| Tests | All pass | 161+ passing ✅ |
| DuckDB refs | 0 | 0 ✅ (except residual converter, deferred) |

**Residual work**: DSL views implementation (plan-b4e2f7a1), DuckDBConverter cleanup

### Semantic Search (7c9d2e1f) — ✅ Archived
**Archive**: [archive/semantic-search-7c9d2e1f/](archive/semantic-search-7c9d2e1f/)

## DSL Views Plan (b4e2f7a1) — Ready
**Status**: Research complete, 10 TDD tasks ready
- Replaces SQL view system with DSL pipe syntax
- Estimated: 4-6 hours
- Use `superpowers:executing-plans` skill

## Project State
- Feature branch: `feat/remove-duckdb-migrate-to-afero-chromedb-with-bleve-search`
- Ready for: DSL views implementation OR Rename epic Phase 1 discovery
