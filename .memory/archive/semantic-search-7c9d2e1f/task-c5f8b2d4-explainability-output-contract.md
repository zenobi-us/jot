---
id: c5f8b2d4
title: Specify Explainability Output Contract
created_at: 2026-02-14T22:50:00+10:30
updated_at: 2026-02-14T23:16:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: b2f4c8d1
story_id: 3d7e9b2a
assigned_to: 2026-02-14-semantic-phase2-execution
---

# Specify Explainability Output Contract

## Objective
Define the CLI output contract for `--explain`, including snippet selection, highlighting, and match-type indicators.

## Related Story
- [story-3d7e9b2a-search-explainability.md](story-3d7e9b2a-search-explainability.md)

## Steps
1. Define exact output structure for explain mode in list and detail contexts.
2. Specify keyword highlight rules and semantic best-sentence fallback.
3. Document truncation, ordering, and no-snippet fallback behavior.
4. Identify rendering constraints in existing display/template services.

## Expected Outcome
Implementation-ready explainability contract with acceptance-test scenarios.

## Actual Outcome
Completed explainability output contract in [research-d2f5a7c9-explainability-output-contract.md](research-d2f5a7c9-explainability-output-contract.md):
- Defined list-output contract for `--explain` with title/path + match label + single-line reason snippet.
- Specified snippet selection hierarchy for keyword and semantic matches with safe fallbacks.
- Defined truncation limits and ordering guarantees.
- Identified rendering constraints in current `displayNoteList()`/`note-list.gotmpl` and proposed semantic-specific view model + template.
- Added acceptance-test scenarios for implementation phase.

## Lessons Learned
Explainability needs a dedicated rendering path; reusing the generic note list template would hide required trust signals.
