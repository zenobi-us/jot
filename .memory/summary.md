# Jot Project Summary

## Current Focus
- **Active Epic**: [epic-8361d3a2](epic-8361d3a2-rename-to-jot.md) - **Rename Project to "Jot"**
  - Phase 1: Discovery — ✅ COMPLETE
  - Phase 2: In-Repo Changes — ✅ COMPLETE (all 10 steps done)
  - Phase 3: GitHub Rename — [NEEDS-HUMAN] manual repo rename
  - Phase 4: External Updates — TODO after Phase 3

## Rename Epic (8361d3a2) Status
**Goal**: Rebrand from "OpenNotes" → "Jot" — ✅ All code changes done

| Completed | Detail |
|-----------|--------|
| Go module | `github.com/zenobi-us/jot` |
| Binary | `dist/jot` |
| Config dir | `~/.config/jot/` |
| Notebook config | `.jot.json` |
| Index dir | `.jot/index/` |
| Env vars | `JOT_CONFIG`, `JOT_NOTEBOOK` |
| Pi extension | `pkgs/pi-jot/` (was `pkgs/pi-opennotes/`) |
| Tests | All passing, 0 lint issues |
| Branch | `feat/rename-to-jot` (2 commits: `bcee714`, `8e332de`) |

**Remaining**: GitHub repo rename (manual), external updates, migration guide

## In-Progress Epics

### Pi Extension (1f41631e) — Tasks & Phases Archived
- All tasks (6) and phases (3) completed and archived to `archive/pi-opennotes-extension-1f41631e/`
- Epic file retained pending final review and learnings distillation
- See [epic-1f41631e](epic-1f41631e-pi-opennotes-extension.md)

## Completed Epics

### DuckDB Removal (f661c068) — ✅ Archived
**Completed**: 2026-02-02 | **Archive**: [archive/duckdb-removal-f661c068/](archive/duckdb-removal-f661c068/)

**Distilled Learnings**:
- [learning-a1b2c3d4](learning-a1b2c3d4-parallel-research-methodology.md) — Parallel research methodology
- [learning-b3c4d5e6](learning-b3c4d5e6-incremental-dependency-replacement.md) — Incremental dependency replacement
- [learning-c5d6e7f8](learning-c5d6e7f8-pure-go-cgo-elimination.md) — Pure Go / CGO elimination
- [learning-d7e8f9a0](learning-d7e8f9a0-interface-first-search-design.md) — Interface-first search design

### Semantic Search (7c9d2e1f) — ✅ Archived
**Archive**: [archive/semantic-search-7c9d2e1f/](archive/semantic-search-7c9d2e1f/)

## Parked Work
- [plan-b4e2f7a1](plan-b4e2f7a1-dsl-views-implementation.md) — DSL views implementation (10 TDD tasks, ready)
- [task-9c4a2f8d](task-9c4a2f8d-github-actions-moonrepo-releases.md) — GitHub Actions CI/CD

## Project State
- Branch: `feat/rename-to-jot`
- Next: Human review → merge → GitHub repo rename → external updates
