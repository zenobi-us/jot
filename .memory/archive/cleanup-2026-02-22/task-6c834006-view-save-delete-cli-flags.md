---
id: 6c834006
title: Add view save/delete CLI flags for custom DSL views
created_at: 2026-02-20T19:10:00+10:30
updated_at: 2026-02-20T22:58:00+10:30
status: completed
epic_id: f661c068
phase_id: 4adb81db
assigned_to: deferred-views-2026-02-20
---

# Add view save/delete CLI flags for custom DSL views

## Objective

Implement `notes view --save` and `notes view --delete` so users can create and remove saved views directly from the CLI without manual config editing.

## Related Story

- Research: [research-b4e2f7a1-dsl-based-views-design.md](research-b4e2f7a1-dsl-based-views-design.md)
- Plan: [plan-b4e2f7a1-dsl-views-implementation.md](plan-b4e2f7a1-dsl-views-implementation.md)

## Steps

1. [x] Define UX contract for `--save` and `--delete` (required args, conflicts, error text).
2. [x] Extend `cmd/notes_view.go` flags and command routing.
3. [x] Add/extend service methods to persist/remove notebook view definitions in `.jot.json` safely.
4. [x] Validate saved view names and query syntax before write.
5. [x] Address review blockers from Task 1B:
   - Enforced `--description` only when `--save` is set.
   - Improved conflict/error messaging for invalid flag combinations.
   - Added tests for `--description` misuse and delete-missing-path behavior.
6. [x] Re-ran task verification and closed review findings.
7. [x] Confirmed docs status for this fix cycle (no additional docs delta required).

## Expected Outcome

Users can run commands like:
- `jot notes view --save work-inbox "tag:work status:todo | sort:created:desc"`
- `jot notes view --delete work-inbox`

and immediately see changes in `jot notes view --list`.

## Actual Outcome

Task 1 fix cycle completed successfully.

- Initial implementation: commit `c9a9542`.
- Review-fix implementation: commit `39ea50cbc6b472e1f4ae9bed0ed4db2c24056c1e`.
- Review verdict after fix cycle: **PASS**.
- Blocking findings: **none**.

Task status is now complete for phase `4adb81db`.

## Lessons Learned

- Enforce flag dependencies as command-level invariants (`--description` requires `--save`) instead of relying on implied usage.
- Treat review feedback as contract-hardening: targeted edge-case tests prevented silent CLI misuse.
- Clear conflict messages reduce user retries and improve discoverability of valid command combinations.
