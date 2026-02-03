---
id: 6a4f2c1d
title: Semantic Search Performance Targets
created_at: 2026-02-03T08:10:00+10:30
updated_at: 2026-02-03T08:10:00+10:30
status: proposed
epic_id: 7c9d2e1f
phase_id: 52a9f0b3
priority: medium
story_points: 3
---

# Semantic Search Performance Targets

## User Story
As a casual user, I want semantic search to feel instantaneous so that I stay in flow.

## Acceptance Criteria
- [ ] For notebooks with ≤ 50k notes, P95 query latency ≤ 750 ms on typical laptop hardware.
- [ ] Search works offline without requiring a remote service.
- [ ] Performance metrics are documented in the release notes.

## Context
Latency spikes erode confidence in semantic search; target bounds guide implementation choices.

## Out of Scope
- Enterprise-scale (100k+ notes) optimization.

## Tasks
- [task-3c4d5e6f-performance-benchmark-plan.md](task-3c4d5e6f-performance-benchmark-plan.md)

## Notes
- Benchmark against baseline Bleve keyword search.
