---
id: 3d7e9b2a
title: Search Explainability for Trust
created_at: 2026-02-03T08:10:00+10:30
updated_at: 2026-02-03T08:10:00+10:30
status: proposed
epic_id: 7c9d2e1f
phase_id: 52a9f0b3
priority: medium
story_points: 3
---

# Search Explainability for Trust

## User Story
As a casual user, I want to understand why a note was returned so that I trust the results.

## Acceptance Criteria
- [ ] A flag (e.g., --explain) shows a short snippet from the best-matching part of the note.
- [ ] Keyword hits highlight matched terms in the snippet.
- [ ] Semantic hits show a best-effort closest sentence or summary snippet.

## Context
Hybrid results can feel opaque; showing “why this matched” improves confidence.

## Out of Scope
- Full semantic attribution or model introspection.

## Tasks
- [task-2b3c4d5e-explainability-spec.md](task-2b3c4d5e-explainability-spec.md)

## Notes
- Consider a compact label in the result list (Exact match / Semantic match).
