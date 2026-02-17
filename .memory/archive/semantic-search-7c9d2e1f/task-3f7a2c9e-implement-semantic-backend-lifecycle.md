---
id: 3f7a2c9e
title: Implement Semantic Backend and Index Lifecycle
created_at: 2026-02-14T23:48:00+10:30
updated_at: 2026-02-15T00:08:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: 91d3f6a2
story_id: 8f2c1a4b
assigned_to: 2026-02-14-semantic-phase3-execution
---

# Implement Semantic Backend and Index Lifecycle

## Objective
Add semantic index abstraction and concrete backend wiring, including lifecycle operations and graceful fallback behavior.

## Related Story
- [story-8f2c1a4b-hybrid-semantic-search.md](story-8f2c1a4b-hybrid-semantic-search.md)

## Steps
1. Add semantic index interface and result types in service/search layer.
2. Implement backend adapter lifecycle (open, query, close) with error wrapping.
3. Wire semantic backend initialization in notebook/service setup.
4. Implement fallback behavior when semantic backend unavailable.
5. Add unit tests for lifecycle and fallback behavior.

## Expected Outcome
Semantic backend can be initialized and queried safely, with deterministic fallback on failure.

## Actual Outcome
Implemented semantic backend lifecycle scaffolding with safe fallback:
- Added semantic contracts and noop backend in `internal/services/semantic_index.go`.
- Extended `NoteService` with semantic backend wiring (`SetSemanticIndex`, `SemanticAvailable`, `FindSemanticCandidates`).
- Wired semantic backend initialization in `NotebookService.Open()` and `NotebookService.Create()` via `createSemanticIndex()`.
- Added tests for noop behavior, backend fallback, configured backend usage, and notebook open wiring in `internal/services/semantic_index_test.go`.
- Verified with full test run: `mise run test`.

## Lessons Learned
Starting with a noop backend keeps lifecycle and service contracts stable while allowing incremental semantic backend rollout without breaking existing search behavior.
