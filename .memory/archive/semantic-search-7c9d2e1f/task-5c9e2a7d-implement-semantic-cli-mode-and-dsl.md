---
id: 5c9e2a7d
title: Implement Semantic CLI Surface, Mode Controls, and DSL Parity
created_at: 2026-02-14T23:48:00+10:30
updated_at: 2026-02-15T01:10:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: 91d3f6a2
story_id: 2a6d8c4f
assigned_to: 2026-02-14-semantic-phase3-execution
---

# Implement Semantic CLI Surface, Mode Controls, and DSL Parity

## Objective
Add semantic search command path with `--mode` controls and ensure DSL filters apply consistently across semantic/hybrid retrieval.

## Related Story
- [story-2a6d8c4f-search-mode-controls.md](story-2a6d8c4f-search-mode-controls.md)
- [story-9c1b5e7a-exclude-archived-notes.md](story-9c1b5e7a-exclude-archived-notes.md)

## Steps
1. Add semantic subcommand and mode enum validation.
2. Parse and apply `--and/--or/--not` conditions through existing parser/AST path.
3. Enforce unsupported-field behavior consistency.
4. Add mode-specific no-result warnings.
5. Add command/service tests for mode validation and filter parity.

## Expected Outcome
Semantic CLI mode works with predictable validation, warnings, and DSL behavior.

## Actual Outcome
Implemented semantic command surface, mode controls, and DSL parity wiring:
- Added `notes search semantic` subcommand with `--mode`, `--and`, `--or`, `--not`, and `--top-k` flags in `cmd/notes_search_semantic.go`.
- Added retrieval mode parsing/validation and semantic orchestration in `internal/services/semantic_search.go`.
- Reused existing condition parser and `BuildQuery()` pipeline for semantic command DSL parity.
- Added mode-specific no-result guidance and semantic-backend-unavailable fallback messaging.
- Added e2e coverage for invalid mode, keyword mode with DSL filters, hybrid fallback warning, and semantic unavailable behavior in `tests/e2e/search_test.go`.
- Verified with `mise run build` and `mise run test`.

## Lessons Learned
E2E CLI tests depend on rebuilding `dist/opennotes`; running `mise run build` before test execution prevents stale-binary false failures.
