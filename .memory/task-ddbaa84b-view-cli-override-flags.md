---
id: ddbaa84b
title: Add CLI override flags for executing saved views
created_at: 2026-02-20T19:11:00+10:30
updated_at: 2026-02-21T21:45:00+10:30
status: completed
epic_id: f661c068
phase_id: b4e2f7a1-plan
assigned_to: null
---

# Add CLI override flags for executing saved views

## Objective

Add runtime overrides (for example `--sort`, `--limit`, `--offset`) to `notes view` so users can tweak execution without editing stored definitions.

## Related Story

- Plan deferred item #2 in [plan-b4e2f7a1-dsl-views-implementation.md](plan-b4e2f7a1-dsl-views-implementation.md)

## Steps

1. Define allowed override flags and precedence rules (CLI > saved view query directives).
2. Add flags and validation in `cmd/notes_view.go`.
3. Extend directive merge logic in view execution path.
4. Ensure grouped views (e.g., kanban) handle overrides correctly.
5. Add tests for each override, invalid values, and precedence behavior.
6. Document examples in views docs.

## Expected Outcome

Users can run:
- `jot notes view recent --limit 5`
- `jot notes view kanban --sort title:asc`

with predictable override behavior.

## Actual Outcome

Not started.

## Lessons Learned

Pending implementation.
## Completion Notes (2026-02-21)

- Added runtime override flags (`--sort`, `--limit`, `--offset`, `--group`) to `notes view`, along with flag validation and helpful errors when misused.
- Introduced directive override plumbing in `ViewService`/`ViewExecutor` so CLI inputs merge cleanly with stored view directives.
- Added unit coverage for CLI validation and override execution paths plus new docs in `docs/views-guide.md`.
- Targeted tests: `go test ./cmd -run TestValidateViewCommandUsage` and `go test ./internal/services -run TestViewService_ExecuteView_WithOverrides` (via `mise exec go -- ...`).
