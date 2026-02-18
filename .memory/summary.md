# OpenNotes Project Summary

## Current Focus
- **Active Epic**: [epic-8361d3a2](epic-8361d3a2-rename-to-jot.md) - **Rename Project to "Jot"**
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

## DSL Views Research Complete (b4e2f7a1)
**Status**: All 6 phases complete ✅

**Key Findings**:
- Participle DSL parser is functional but has zero production callers
- `view.go`: ~500 lines removable SQL, ~550 lines reusable, ~150 lines to replace
- Pipe syntax selected: `"filter DSL | directives"`
- Architecture maps directly to `FindOpts` with zero interface changes

**Implementation Plan**: [plan-b4e2f7a1-dsl-views-implementation.md](plan-b4e2f7a1-dsl-views-implementation.md)
- 10 TDD tasks ready for execution
- Estimated: 4-6 hours
- Use `superpowers:executing-plans` skill

## Recently Completed
- ✅ DSL Views Research Phase 6 (Feb 18, 2026) — Implementation plan created
- ✅ **Semantic Search Enhancement** (epic-7c9d2e1f) - Archived Feb 17, 2026
- ✅ **DuckDB Removal** (epic-f661c068) - Core completed, views cleanup planned

## Project State
- Feature branch: `feat/remove-duckdb-migrate-to-afero-chromedb-with-bleve-search`
- Ready for: DSL views implementation OR Rename epic Phase 1 discovery
