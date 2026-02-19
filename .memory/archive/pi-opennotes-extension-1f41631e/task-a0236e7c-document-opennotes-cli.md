---
id: a0236e7c
title: Document OpenNotes CLI Interface
created_at: 2026-01-28T23:30:00+10:30
updated_at: 2026-01-29T09:05:00+10:30
status: done
epic_id: 1f41631e
phase_id: 43842f12
assigned_to: null
---

# Document OpenNotes CLI Interface

## Objective

Document all OpenNotes CLI commands that will be used by the pi extension, including their arguments, output formats, and error conditions.

## Completed Steps

- [x] List all relevant CLI commands
- [x] Document `opennotes notes search` command
- [x] Document `opennotes notes list` command
- [x] Document `opennotes notes add` command
- [x] Document `opennotes notebook list` command
- [x] Document `opennotes notes view` command
- [x] Document output formats (JSON, table, list)
- [x] Document error codes and messages
- [x] Identify commands that support `--format` flag

---

## CLI Interface Reference

### Global Flags

All commands support these flags:

| Flag | Type | Description |
|------|------|-------------|
| `--notebook <path>` | string | Path to notebook (overrides auto-detection) |
| `--help` | bool | Show help |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENNOTES_CONFIG` | `~/.config/opennotes/config.json` | Config file location |
| `DEBUG` | (unset) | Enable debug logging |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `LOG_FORMAT` | `compact` | Log format: compact, console, json, ci |

---

## Command Reference

### 1. `opennotes notes search`

**Purpose**: Search notes using text, fuzzy matching, boolean queries, or SQL.

```
opennotes notes search [query] [flags]
opennotes notes search query [flags]  # Boolean query mode
```

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--fuzzy` | bool | false | Enable fuzzy matching |
| `--sql <query>` | string | - | Execute custom SQL query |
| `--notebook <path>` | string | - | Override notebook path |

#### Subcommand: `search query`

| Flag | Type | Description |
|------|------|-------------|
| `--and <field=value>` | array | AND conditions (all must match) |
| `--or <field=value>` | array | OR conditions (any must match) |
| `--not <field=value>` | array | NOT conditions (exclusions) |

**Supported query fields:**
- `data.tag`, `data.tags`, `data.status`, `data.priority`
- `data.assignee`, `data.author`, `data.type`, `data.category`
- `data.project`, `data.sprint`
- `path`, `title`
- `links-to`, `linked-by`

#### Output Format

**Text search** returns rendered markdown (glamour formatted).

**SQL search** (`--sql`) returns JSON array:

```json
[
  {
    "file_path": "notes/meeting.md",
    "content": "# Meeting Notes\n...",
    "metadata": { "title": "Meeting Notes", "tags": ["meeting"] }
  }
]
```

#### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | No results found / general error |
| 2 | Invalid SQL query |
| 3 | Security violation (path traversal blocked) |

---

### 2. `opennotes notes list`

**Purpose**: List all markdown notes in a notebook.

```
opennotes notes list [flags]
opennotes notes ls [flags]  # Alias
```

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--notebook <path>` | string | - | Override notebook path |

#### Output Format

Returns glamour-rendered list of notes:

```
📄 Notes in "My Notebook" (3 notes)

  • meeting-2026-01-28.md
    Created: 2026-01-28 | Modified: 2026-01-28

  • project-plan.md
    Created: 2026-01-15 | Modified: 2026-01-27
```

**Note**: No native JSON output. For JSON, use SQL:
```bash
opennotes notes search --sql "SELECT file_path, metadata FROM read_markdown('**/*.md')"
```

---

### 3. `opennotes notes add`

**Purpose**: Create a new markdown note.

```
opennotes notes add <title> [path] [flags]
```

#### Arguments

| Position | Required | Description |
|----------|----------|-------------|
| 1 | Yes | Note title |
| 2 | No | Directory path (relative to notebook) |

#### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--template <name>` | string | Template to use |
| `--data <field=value>` | array | Set frontmatter fields |
| `--notebook <path>` | string | Override notebook path |

#### Output Format

Returns created file path:

```
Created: notes/meeting-2026-01-29.md
```

#### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (invalid title, missing notebook, etc.) |

---

### 4. `opennotes notebook list`

**Purpose**: List all registered and discovered notebooks.

```
opennotes notebook list [flags]
opennotes notebook ls [flags]  # Alias
```

#### Output Format

Returns glamour-rendered notebook list:

```
📚 Notebooks (3 found)

  • Work Notes
    Path: /home/user/work/notes
    Source: registered

  • Personal
    Path: /home/user/notes
    Source: ancestor
```

**Note**: No native JSON output. For programmatic use, the extension will parse this output.

---

### 5. `opennotes notes view`

**Purpose**: Execute named views or list available views.

```
opennotes notes view [name] [flags]
opennotes notes view --list [flags]
```

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--list` | bool | false | List all available views |
| `--format <fmt>` | string | `list` | Output format: list, table, json |
| `--param <key=value>` | string | - | View parameters |
| `--notebook <path>` | string | - | Override notebook path |

#### List Mode Output (`--list --format json`)

```json
{
  "views": [
    {
      "name": "today",
      "origin": "built-in",
      "description": "Notes created or updated today"
    },
    {
      "name": "kanban",
      "origin": "built-in",
      "description": "Notes grouped by status column",
      "parameters": [
        {
          "name": "status",
          "type": "list",
          "required": false,
          "default": "backlog,todo,in-progress,reviewing,testing,deploying,done",
          "description": "Comma-separated list of status values"
        }
      ]
    }
  ]
}
```

#### Execute Mode Output (`--format json`)

Returns JSON array of matching notes (format depends on view SQL).

#### Built-in Views

| Name | Description | Parameters |
|------|-------------|------------|
| `today` | Notes created/updated today | none |
| `recent` | Last 20 modified notes | none |
| `kanban` | Notes grouped by status | `status` (list) |
| `untagged` | Notes without tags | none |
| `orphans` | Notes with no incoming links | `definition` (string) |
| `broken-links` | Notes with broken references | none |

---

### 6. `opennotes notebook` (no subcommand)

**Purpose**: Show current notebook info.

```
opennotes notebook [flags]
```

#### Output Format

```
📓 Notebook: My Notes

  Path: /home/user/notes
  Notes: 42
  Templates: 3
  Created: 2026-01-15
```

---

## Error Scenarios

### Common Errors

| Scenario | Message Pattern | Exit Code |
|----------|-----------------|-----------|
| No notebook found | `Error: no notebook found in current directory or ancestors` | 1 |
| Notebook not found | `Error: notebook not found at path: <path>` | 1 |
| Invalid SQL | `Error: invalid query: <details>` | 2 |
| SQL timeout | `Error: query timeout after 30s` | 2 |
| Path traversal | `Error: path traversal not allowed: <path>` | 3 |
| View not found | `Error: view not found: <name>` | 1 |
| Template not found | `Error: template not found: <name>` | 1 |

### Error Output Format

All errors write to stderr:

```
Error: <message>

For help, run: opennotes <command> --help
```

---

## JSON Output Summary

| Command | Native JSON | Workaround |
|---------|-------------|------------|
| `notes search --sql` | ✅ Yes | - |
| `notes search query` | ❌ No | Use `--sql` |
| `notes list` | ❌ No | Use `notes search --sql` |
| `notes add` | ❌ No | Parse stdout |
| `notebook list` | ❌ No | Parse stdout |
| `notes view --list` | ✅ Yes (`--format json`) | - |
| `notes view <name>` | ✅ Yes (`--format json`) | - |

---

## Extension Strategy

Given the JSON output limitations, the extension will:

1. **Search**: Use `--sql` mode for all searches (guarantees JSON)
2. **List Notes**: Use `notes search --sql "SELECT ... FROM read_markdown('**/*.md')"`
3. **List Notebooks**: Parse text output or use config file directly
4. **Views**: Use `--format json` for both list and execute modes
5. **Add Note**: Parse stdout for created path

---

## Expected Outcome

✅ Comprehensive reference document for CLI interface completed.

## Actual Outcome

All OpenNotes CLI commands documented with:
- Full flag specifications
- Output format examples
- Error scenarios and exit codes
- JSON output availability
- Extension integration strategy

## Lessons Learned

1. Not all commands support JSON output - extension needs workarounds
2. SQL mode provides most flexibility for search operations
3. Views have full JSON support, making them ideal for structured queries
4. Notebook list parsing will require text parsing (acceptable for v1)
