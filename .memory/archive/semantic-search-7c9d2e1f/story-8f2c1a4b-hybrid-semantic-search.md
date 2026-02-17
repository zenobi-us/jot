---
id: 8f2c1a4b
title: Hybrid Semantic Search for Conceptual Queries
created_at: 2026-02-03T08:10:00+10:30
updated_at: 2026-02-14T21:58:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: 52a9f0b3
priority: high
story_points: 5
---

# Hybrid Semantic Search for Conceptual Queries

## User Story
As a casual user, I want search to find notes even if I don’t remember exact words so that I can quickly locate ideas and meeting notes.

## Acceptance Criteria
- [x] Given a paraphrased query, results include semantically related notes without exact keyword matches.
- [x] Results are merged from keyword and semantic search.
- [x] Results indicate whether they matched by keyword or semantic similarity.
- [x] The dedicated semantic subcommand uses the hybrid merge by default.

## Context
Casual users often remember the concept but not the exact phrasing used in a note.

## Out of Scope
- Tuning advanced ranking weights for power users.

## Tasks
- [task-1a2b3c4d-define-hybrid-merge.md](task-1a2b3c4d-define-hybrid-merge.md)

## Notes
- Recommend RRF-based merge for score robustness.
