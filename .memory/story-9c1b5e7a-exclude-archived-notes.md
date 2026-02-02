---
id: 9c1b5e7a
title: Exclude Archived or Draft Notes
created_at: 2026-02-03T08:10:00+10:30
updated_at: 2026-02-03T08:25:00+10:30
status: proposed
epic_id: 7c9d2e1f
phase_id: 52a9f0b3
priority: medium
story_points: 2
---

# Exclude Notes via Query DSL

## User Story
As a casual user, I want to exclude notes using the existing query DSL so that I can filter results the same way as `opennotes notes search query`.

## Acceptance Criteria
- [ ] Semantic search accepts the same query DSL filters as `opennotes notes search query` (data.*, path, title, links-to, linked-by).
- [ ] Users can combine exclusions with `--not` semantics (e.g., `--and data.tag=epic --not data.status=archived`).
- [ ] Exclusions apply to both keyword and semantic results in hybrid mode.

## Context
The query DSL already supports structured filters; semantic search should honor the same exclusion syntax.

## Out of Scope
- New or alternative DSL formats.

## Tasks
- TBD

## Notes
- Align with fields supported by `opennotes notes search query` (data.*, path, title, links-to, linked-by).
