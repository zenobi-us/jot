---
id: d7a1c3e5
title: Define Retrieval Mode Controls and Validation
created_at: 2026-02-14T22:50:00+10:30
updated_at: 2026-02-14T23:21:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: b2f4c8d1
story_id: 2a6d8c4f
assigned_to: 2026-02-14-semantic-phase2-execution
---

# Define Retrieval Mode Controls and Validation

## Objective
Specify CLI mode control behavior (`hybrid`, `keyword`, `semantic`) including validation, precedence, and no-result warnings.

## Related Story
- [story-2a6d8c4f-search-mode-controls.md](story-2a6d8c4f-search-mode-controls.md)

## Steps
1. Confirm command surface and flag interactions with existing search command options.
2. Define validation errors for invalid modes and conflicting options.
3. Define warning text for empty results in narrowed modes.
4. Document behavior across semantic subcommand and shared search pathways.

## Expected Outcome
Consistent CLI behavior contract for retrieval mode controls.

## Actual Outcome
Completed mode-control contract in [research-e1c3a5d7-retrieval-mode-controls-contract.md](research-e1c3a5d7-retrieval-mode-controls-contract.md):
- Defined dedicated semantic subcommand surface with `--mode {hybrid|keyword|semantic}` and default `hybrid`.
- Defined validation behavior for invalid modes and incompatible flag usage.
- Defined mode-specific no-result advisory warnings with suggested retry modes.
- Defined service metadata contract to support accurate warning/reporting behavior.
- Confirmed backward compatibility for existing non-semantic search commands.

## Lessons Learned
Keeping mode controls isolated to semantic command path minimizes risk to established search UX while enabling power-user diagnostics.
