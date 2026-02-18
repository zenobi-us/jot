# OpenNotes Project Summary

## Current Focus
- **Active Epic**: [epic-8361d3a2](epic-8361d3a2-rename-to-jot.md) - **Rename Project to "Jot"**
  - Phase 1: Discovery — Identify all rename locations (in-repo and external)
  - Status: Planning
- **Parked Research**: [research-b4e2f7a1](research-b4e2f7a1-dsl-based-views-design.md) - DSL-based views design
  - Phases 1-4 complete, Phases 5-6 pending
  - **Selected Design**: Option C (Hybrid — Pipe Syntax)

## Rename Epic (8361d3a2) Overview
**Goal**: Rebrand from "OpenNotes" → "Jot"
- New binary: `jot`
- New module: `github.com/zenobi-us/jot`
- New repo: `github.com/zenobi-us/jot`
- New config: `~/.config/jot/`, `.jot.json`

## Key Research Findings (DSL Views)
- Participle DSL parser is functional but has zero production callers
- `view.go` has ~600 lines dead SQL code, ~400 lines reusable
- Pipe syntax selected for CLI design
- Architecture maps directly to `FindOpts` with zero interface changes

## Recently Completed
- ✅ **Semantic Search Enhancement** (epic-7c9d2e1f) - Archived Feb 17, 2026
- ✅ **DuckDB Removal** (epic-f661c068) - Core completed, views cleanup remaining

## Project State
- Feature branch: `feat/remove-duckdb-migrate-to-afero-chromedb-with-bleve-search`
- Ready for: Rename epic Phase 1 → discovery and inventory
