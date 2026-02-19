---
id: 3c4d5e6f
title: Define Semantic Search Performance Targets & Benchmarks
created_at: 2026-02-03T08:35:00+10:30
updated_at: 2026-02-03T09:10:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: 52a9f0b3
story_id: 6a4f2c1d
assigned_to: 2026-02-03-semantic-epic-planning
---

# Define Semantic Search Performance Targets & Benchmarks

## Objective
Establish performance targets and a repeatable benchmark plan for semantic search latency.

## Related Story
- [story-6a4f2c1d-semantic-search-latency.md](story-6a4f2c1d-semantic-search-latency.md)

## Steps
1. Confirm target notebook sizes and hardware assumptions.
2. Define benchmark datasets and query sets.
3. Specify P50/P95 reporting and acceptance thresholds.

## Expected Outcome
Documented performance targets and benchmarking methodology.

## Actual Outcome
- Target sizes: 1k, 10k, 50k notes on typical laptop hardware.
- Query set: 20 mixed queries (exact keyword, paraphrase, short phrase, long phrase) with 5 repeats.
- Metrics: report P50/P95 latency for semantic-only, keyword-only, and hybrid merge.
- Acceptance: P95 ≤ 750 ms for ≤ 50k notes; P50 ≤ 250 ms for ≤ 50k notes.

## Lessons Learned
- Separate metric reporting for each mode prevents masking regressions.
