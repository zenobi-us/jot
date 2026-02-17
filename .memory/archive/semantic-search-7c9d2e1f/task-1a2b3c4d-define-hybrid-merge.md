---
id: 1a2b3c4d
title: Define Hybrid Merge Strategy for Semantic Search
created_at: 2026-02-03T08:35:00+10:30
updated_at: 2026-02-03T08:45:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: 52a9f0b3
story_id: 8f2c1a4b
assigned_to: 2026-02-03-semantic-epic-planning
---

# Define Hybrid Merge Strategy for Semantic Search

## Objective
Specify how keyword and semantic results are merged (e.g., RRF) and how match labels are surfaced.

## Related Story
- [story-8f2c1a4b-hybrid-semantic-search.md](story-8f2c1a4b-hybrid-semantic-search.md)

## Steps
1. Document candidate merge strategies and select default (RRF recommended).
2. Define match labeling rules (Exact match / Semantic match / Hybrid).
3. Capture fallback behavior when one list is empty.

## Expected Outcome
A clear, documented merge approach with labeling rules for hybrid results.

## Actual Outcome
- Default merge strategy: Reciprocal Rank Fusion (RRF) over top-K from keyword and semantic lists.
- Labeling: "Exact match" when only keyword hit, "Semantic match" when only semantic hit, "Hybrid" when both appear in merged list.
- Fallback: if one list is empty, return the other list unchanged and label accordingly.

## Lessons Learned
- RRF avoids score normalization and keeps merging logic simple.
