---
name: migrate-impl-review-loop
description: Implements one migration-plan step, then immediately reviews it before moving to next step.
---

## worker

Implement this step from task-8281af6b: {task}

Constraints:
- Work only in /mnt/Store/Projects/Mine/Github/opennotes
- Follow project conventions (thin cmd/fat services, mise tasks)
- Make concrete code/doc changes needed for this step
- Run relevant verification commands and report evidence
- Return concise summary of files changed and why.

## reviewer

Review the implementation for this step: {task}

Use the implementer report below as context:
{previous}

Perform code review for correctness, safety, architecture fit, and tests. If issues are found, suggest exact fixes and severity. If acceptable, state APPROVED for this step.
