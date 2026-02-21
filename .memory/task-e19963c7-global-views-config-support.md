---
id: e19963c7
title: Add global views support in user config hierarchy
created_at: 2026-02-20T19:14:00+10:30
updated_at: 2026-02-20T19:14:00+10:30
status: todo
epic_id: f661c068
phase_id: b4e2f7a1-plan
assigned_to: null
---

# Add global views support in user config hierarchy

## Objective

Enable global custom views in `~/.config/jot/config.json` with clear precedence versus notebook-local and built-in views.

## Related Story

- Plan deferred item #5 in [plan-b4e2f7a1-dsl-views-implementation.md](plan-b4e2f7a1-dsl-views-implementation.md)

## Steps

1. Implement/confirm global config loading for views in `ViewService`.
2. Define precedence contract and conflict resolution (notebook > global > built-in).
3. Add tests covering lookup, list output origin, and override behavior.
4. Add migration-safe handling for missing/invalid global config.
5. Document global views configuration and examples.

## Expected Outcome

Users can define once in global config and run the same view in any notebook, with documented precedence behavior.

## Actual Outcome

Not started.

## Lessons Learned

Pending implementation.
