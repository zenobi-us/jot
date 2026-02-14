---
id: b3d6a9c1
title: Define Query DSL Filter Parity for Semantic Path
created_at: 2026-02-14T22:50:00+10:30
updated_at: 2026-02-14T23:11:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: b2f4c8d1
story_id: 9c1b5e7a
assigned_to: 2026-02-14-semantic-phase2-execution
---

# Define Query DSL Filter Parity for Semantic Path

## Objective
Ensure semantic and hybrid retrieval paths honor existing query DSL filters and `--not` semantics identically to keyword search.

## Related Story
- [story-9c1b5e7a-exclude-archived-notes.md](story-9c1b5e7a-exclude-archived-notes.md)

## Steps
1. Map AST evaluation order for keyword and semantic candidate generation.
2. Define pre-merge filtering contract and unsupported-filter error behavior.
3. Specify parity tests for `data.*`, `path`, `title`, `links-to`, `linked-by`.
4. Capture migration impact for existing query command behavior.

## Expected Outcome
Clear integration contract and test matrix for DSL parity in semantic search.

## Actual Outcome
Completed filter parity contract in [research-a4e1b7c3-dsl-parity-contract-semantic.md](research-a4e1b7c3-dsl-parity-contract-semantic.md):
- Mapped canonical condition flow: `ParseConditions()` → `BuildQuery()` → apply to both retrieval branches.
- Defined pre-merge filtering requirement to enforce identical exclusions in hybrid results.
- Confirmed unsupported `links-to`/`linked-by` remain fail-fast errors in semantic/hybrid until link graph support exists.
- Defined parity test matrix for metadata fields, path/title filters, alias handling, and AND/OR/NOT behavior.

## Lessons Learned
Using one query AST as the semantic/keyword contract eliminates divergence risk between retrieval modes.
