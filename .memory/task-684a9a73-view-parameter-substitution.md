---
id: 684a9a73
title: Implement named view parameter substitution
created_at: 2026-02-20T19:13:00+10:30
updated_at: 2026-02-20T19:13:00+10:30
status: todo
epic_id: f661c068
phase_id: b4e2f7a1-plan
assigned_to: null
---

# Implement named view parameter substitution

## Objective

Add `{{param_name}}` substitution for saved views so view definitions can be reused with runtime values through `--param`.

## Related Story

- Plan deferred item #4 in [plan-b4e2f7a1-dsl-views-implementation.md](plan-b4e2f7a1-dsl-views-implementation.md)
- API/docs context in `docs/views-api.md` and `docs/views-guide.md`

## Steps

1. Confirm parameter schema contract (types, required/default behavior).
2. Implement substitution pipeline before query parse/execute.
3. Validate provided params against definition with clear errors.
4. Add tests for required params, defaults, type mismatches, and unknown params.
5. Ensure template variable support (`{{today}}`) remains compatible.
6. Document parameterized view examples and troubleshooting.

## Expected Outcome

Users can define parameterized views and execute with:
- `jot notes view by-author --param author="Alice"`

with deterministic validation and rendering.

## Actual Outcome

Not started.

## Lessons Learned

Pending implementation.
