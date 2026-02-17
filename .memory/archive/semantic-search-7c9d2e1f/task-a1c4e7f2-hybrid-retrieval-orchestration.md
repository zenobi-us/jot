---
id: a1c4e7f2
title: Implement Hybrid Retrieval Orchestration and Merge Contract
created_at: 2026-02-14T22:50:00+10:30
updated_at: 2026-02-14T23:06:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: b2f4c8d1
story_id: 8f2c1a4b
assigned_to: 2026-02-14-semantic-phase2-execution
---

# Implement Hybrid Retrieval Orchestration and Merge Contract

## Objective
Design the implementation contract for running keyword and semantic retrieval, then merging with RRF in deterministic order.

## Related Story
- [story-8f2c1a4b-hybrid-semantic-search.md](story-8f2c1a4b-hybrid-semantic-search.md)

## Steps
1. Define semantic pipeline interfaces and candidate set boundaries.
2. Specify RRF merge inputs, tie-breaking, and stable ordering rules.
3. Define result match labels (Exact/Semantic/Hybrid) and fallback behavior.
4. Identify integration points in existing search service flow.

## Expected Outcome
Implementation-ready spec and task checklist for hybrid retrieval orchestration.

## Actual Outcome
Completed architecture contract in [research-f2d6a8c0-hybrid-retrieval-contract.md](research-f2d6a8c0-hybrid-retrieval-contract.md):
- Defined service-layer orchestration entry point and integration with existing `SearchWithConditions` flow.
- Defined candidate set boundaries (`keywordTopK=100`, `semanticTopK=100`) and pre-merge filtering requirement.
- Specified deterministic RRF merge with stable tie-break rules.
- Defined match labels (Exact/Semantic/Hybrid) and one-source-empty fallback behavior.
- Captured proposed interface/types for Phase 3 implementation.

## Lessons Learned
Anchoring hybrid orchestration in the service layer avoids disruptive command rewrites while preserving existing query behavior.
