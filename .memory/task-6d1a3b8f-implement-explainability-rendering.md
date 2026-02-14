---
id: 6d1a3b8f
title: Implement Explainability Output and Semantic Rendering
created_at: 2026-02-14T23:48:00+10:30
updated_at: 2026-02-15T01:14:00+10:30
status: in-progress
epic_id: 7c9d2e1f
phase_id: 91d3f6a2
story_id: 3d7e9b2a
assigned_to: 2026-02-14-semantic-phase3-execution
---

# Implement Explainability Output and Semantic Rendering

## Objective
Implement `--explain` output contract with match labels and snippets via semantic-specific rendering templates.

## Related Story
- [story-3d7e9b2a-search-explainability.md](story-3d7e9b2a-search-explainability.md)

## Steps
1. Add semantic result view model with explain fields.
2. Implement snippet selection and truncation rules for keyword/semantic hits.
3. Add template for semantic result output with labels and reason snippets.
4. Integrate `--explain` flag behavior into semantic command path.
5. Add tests covering fallback/no-snippet scenarios.

## Expected Outcome
Explain mode renders concise, trustworthy reasons without changing non-explain output.

## Actual Outcome
In progress (checkpointed for next session):
- Added semantic result hit model and explain snippet generation scaffolding in `internal/services/semantic_search.go`.
- Added `SearchSemanticDetailed()` path that returns per-hit match type + explain text metadata.
- Kept existing `SearchSemantic()` backward-compatible by adapting detailed hits back to `[]Note`.
- Build and test pass at checkpoint (`mise run build`, `mise run test`).

Remaining to complete task:
- Add `--explain` flag on semantic CLI command.
- Add semantic-specific output template/rendering with labels and `Why:` snippet lines.
- Add focused tests for explain/no-snippet/truncation output behavior.

## Lessons Learned
Splitting retrieval into detailed-hit and legacy-note paths allows explainability rollout without breaking current command outputs.
