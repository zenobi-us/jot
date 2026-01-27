# Kanban View - Quick Reference

## Command Usage

```bash
# List all views
opennotes notes view --list

# Execute kanban with defaults (all statuses)
opennotes notes view kanban

# Filter to specific statuses
opennotes notes view kanban --param status=todo,done

# Only "in-progress" tasks
opennotes notes view kanban --param status=in-progress

# Output as JSON
opennotes notes view kanban --param status=todo --format json

# Output as table
opennotes notes view kanban --format table
```

## Parameter Reference

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `status` | list | `backlog,todo,in-progress,reviewing,testing,deploying,done` | Comma-separated status values to include |

## Default Status Values

```
backlog → Not started
todo → Ready to work
in-progress → Currently being worked on
reviewing → Awaiting review
testing → Under test
deploying → Rolling out
done → Complete
```

## Customization

### Global Custom Kanban
File: `~/.config/opennotes/config.json`

```json
{
  "views": {
    "kanban": {
      "name": "kanban",
      "description": "Custom global kanban",
      "parameters": [{
        "name": "status",
        "type": "list",
        "default": "todo,in-progress,done"
      }],
      "query": {
        "conditions": [{
          "field": "metadata->>'status'",
          "operator": "IN",
          "value": "{{status}}"
        }],
        "orderBy": "(metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC"
      }
    }
  }
}
```

### Notebook-Specific Kanban
File: `notebook/.opennotes.json`

```json
{
  "views": {
    "kanban": {
      "name": "kanban",
      "description": "Team kanban board",
      "parameters": [{
        "name": "status",
        "type": "list",
        "default": "planning,dev,qa,shipped"
      }],
      "query": {
        "conditions": [{
          "field": "metadata->>'status'",
          "operator": "IN",
          "value": "{{status}}"
        }],
        "orderBy": "(metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC"
      }
    }
  }
}
```

## How Views Resolve

```
Command: opennotes notes view kanban

↓

ViewService.GetView("kanban")
  1. Check notebook/.opennotes.json ["kanban"] → found? → USE THIS
  2. Check ~/.config/opennotes/config.json ["views"]["kanban"] → found? → USE THIS
  3. Check built-in views ["kanban"] → found? → USE THIS
  4. Not found → Return error

View resolution priority:
  Notebook > Global > Built-in
```

## Note Requirements

Notes must have `status` metadata in YAML front matter:

```markdown
---
title: My Task
status: todo
priority: 1
---

Note content...
```

Available in note:
- `metadata->>'status'` — String value from YAML
- `metadata->>'priority'` — String value (cast to INTEGER in ordering)
- `metadata->>'updated_at'` — Timestamp

## SQL Generated

Input: `opennotes notes view kanban --param status=todo,done`

```sql
SELECT * FROM read_markdown(?, include_filepath:=true)
WHERE metadata->>'status' IN (?,?)
ORDER BY (metadata->>'priority')::INTEGER DESC, 
         metadata->>'updated_at' DESC

Args: [
  "/notebook/**/*.md",
  "todo",
  "done"
]
```

## Results Format

### List Format (default)
```
path: notes/tasks/my-task.md
content: Note content here...
metadata: {"title":"My Task","status":"todo","priority":"1"}
```

### Table Format
```
| Path | Title | Status | Priority | Updated |
|------|-------|--------|----------|---------|
| ... | My Task | todo | 1 | ... |
```

### JSON Format
```json
[
  {
    "path": "notes/tasks/my-task.md",
    "content": "Note content...",
    "metadata": {"title":"My Task","status":"todo"}
  }
]
```

## Troubleshooting

### No results returned?

1. Check note has `status` in YAML:
   ```markdown
   ---
   status: todo
   ---
   ```

2. Check status value matches parameter:
   ```bash
   opennotes notes view kanban --param status=todo
   # Won't find notes with status: "in-progress"
   ```

3. Check default status values:
   ```bash
   opennotes notes view --list
   # Look for kanban view parameters
   ```

### Wrong kanban definition used?

Check which view is being used:
```bash
# List all kanban views (notebook, global, built-in)
opennotes notes view --list --format json | jq '.views[] | select(.name == "kanban")'

# See which one has the description you expect
```

### Can't parse custom kanban?

Validate JSON syntax:
```bash
# Check notebook config
cat notebook/.opennotes.json | jq '.views.kanban'

# Or global config
cat ~/.config/opennotes/config.json | jq '.views.kanban'
```

## Key Files

| File | Purpose |
|------|---------|
| `internal/services/view.go` | ViewService main logic |
| `internal/core/view.go` | Type definitions |
| `cmd/notes_view.go` | CLI command handler |
| `internal/services/view_test.go` | 71 unit tests |
| `tests/e2e/go_smoke_test.go` | Integration tests |

## Advanced: How Status Filtering Works

1. User provides: `--param status=todo,done`
2. Parser splits: `["todo", "done"]`
3. Validator checks: Both are strings, no empty values ✓
4. SQL builder creates: `metadata->>'status' IN (?,?)`
5. DuckDB args: `["todo", "done"]`
6. Query checks: For each note, is metadata.status one of these values?
7. Returns: Only matching notes

## Performance

- View lookup: < 1ms
- SQL generation: < 1ms
- Query (1000 notes): ~70ms
- Format results: < 5ms
- **Total**: ~75ms

Bottleneck: DuckDB markdown parsing (not view system)

## Security

SQL injection attempts are blocked:
```bash
# Won't work:
--param status="todo' OR '1'='1"

# Reason: Value is escaped and parameterized
# SELECT ... WHERE metadata->>'status' IN (?)
# Args: ["todo' OR '1'='1"]
# → Treated as literal string, not SQL
```

## All Built-in Views

| View | Purpose | Example |
|------|---------|---------|
| `today` | Notes from today | `opennotes notes view today` |
| `recent` | Last 20 modified | `opennotes notes view recent` |
| `kanban` | By status | `opennotes notes view kanban` |
| `untagged` | No tags | `opennotes notes view untagged` |
| `orphans` | No incoming links | `opennotes notes view orphans` |
| `broken-links` | Invalid refs | `opennotes notes view broken-links` |

