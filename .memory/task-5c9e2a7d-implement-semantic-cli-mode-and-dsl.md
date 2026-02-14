---
id: 5c9e2a7d
title: Implement Semantic CLI Surface, Mode Controls, and DSL Parity
created_at: 2026-02-14T23:48:00+10:30
updated_at: 2026-02-14T23:48:00+10:30
status: todo
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
Pending.

## Lessons Learned
TBD.
