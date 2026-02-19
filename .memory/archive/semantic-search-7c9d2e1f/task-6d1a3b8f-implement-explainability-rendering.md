---
id: 6d1a3b8f
title: Implement Explainability Output and Semantic Rendering
created_at: 2026-02-14T23:48:00+10:30
updated_at: 2026-02-15T01:27:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: 91d3f6a2
story_id: 3d7e9b2a
assigned_to: 2026-02-15-semantic-phase3-task6-task7
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
Completed.
- Added semantic CLI flag `--explain` to `opennotes notes search semantic`.
- Routed semantic command through `SearchSemanticDetailed()` and added an explain-specific render path.
- Added semantic output template `internal/services/templates/note-search-semantic.gotmpl` with match labels and `Why:` snippets.
- Added focused tests:
  - `internal/services/semantic_search_test.go` for explain highlight/no-snippet/truncation behavior.
  - `internal/services/templates_test.go` coverage for semantic explain template rendering.
- Verification passed: `mise run format`, `mise run build`, `mise run test`.

## Lessons Learned
Keeping explain rendering on a dedicated semantic template preserves existing note-list UX while allowing richer per-hit context in explain mode.
