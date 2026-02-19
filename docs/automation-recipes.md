# Automation Recipes

Practical shell-friendly workflows using supported Jot commands.

## 1) Daily note count snapshot

```bash
#!/usr/bin/env bash
set -euo pipefail

COUNT=$(jot notes view recent --format json | jq 'length')
echo "$(date -Iseconds) recent_count=$COUNT" >> .jot-metrics.log
```

## 2) Active task report

```bash
#!/usr/bin/env bash
set -euo pipefail

jot notes search query \
  --and data.tag=task \
  --and data.status=active
```

## 3) Weekly semantic discovery report

```bash
#!/usr/bin/env bash
set -euo pipefail

jot notes search semantic "open work items and blockers" --mode hybrid --explain
```

## 4) Kanban snapshot (JSON)

```bash
#!/usr/bin/env bash
set -euo pipefail

jot notes view kanban --format json > kanban-snapshot.json
jq '. | length' kanban-snapshot.json
```

## 5) Notebook health check script

```bash
#!/usr/bin/env bash
set -euo pipefail

jot --version
jot notebook list
jot notes list >/dev/null
jot notes search "test" >/dev/null || true
jot notes view --list

echo "healthcheck ok"
```

## 6) Deterministic CI command pattern

Use explicit notebook path in CI:

```bash
JOT_NOTEBOOK=/abs/path/to/notebook jot notes view recent --format json
```

## 7) Cron-safe daily digest

```bash
#!/usr/bin/env bash
set -euo pipefail

OUT="${HOME}/jot-digest-$(date +%F).txt"
{
  echo "# Jot Daily Digest"
  echo "Generated: $(date -Iseconds)"
  echo
  jot notes view recent
  echo
  jot notes search query --and data.status=active --or data.priority=high
} > "$OUT"
```

## Tips

- Prefer `notes view --format json` when you need stable machine-readable output.
- Use `JOT_NOTEBOOK` explicitly in automation contexts.
- Keep scripts read-only unless you intentionally mutate notebooks.

## Related docs

- [views-guide.md](views-guide.md)
- [commands/notes-search.md](commands/notes-search.md)
- [semantic-search-guide.md](semantic-search-guide.md)
- [getting-started-troubleshooting.md](getting-started-troubleshooting.md)
