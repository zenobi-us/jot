# Views System Guide

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Built-in Views](#built-in-views)
- [Creating Custom Views](#creating-custom-views)
- [Parameter System](#parameter-system)
- [Template Variables](#template-variables)
- [Output Formatting](#output-formatting)
- [Configuration Precedence](#configuration-precedence)
- [Troubleshooting](#troubleshooting)

---

## Overview

The Views System provides named, reusable query presets that make it easy to find and organize your notes without writing SQL. Views are perfect for:

- **Daily workflows**: Quick access to today's notes or recent changes
- **Project tracking**: Organize notes by status (kanban-style)
- **Maintenance**: Find untagged, orphaned, or broken-linked notes
- **Custom workflows**: Define your own views for team or personal use

### Key Features

- **6 Built-in Views**: Ready-to-use presets for common tasks
- **Custom Views**: Define your own views in global or notebook configs
- **Parameters**: Make views flexible with runtime parameters
- **Template Variables**: Dynamic date/time values
- **Multiple Outputs**: List, table, or JSON formats
- **Fast Performance**: Sub-millisecond query generation

---

## Quick Start

### List Available Views

```bash
# See all views available in current notebook
opennotes notes view --list

# Get JSON output for programmatic use
opennotes notes view --list --format json
```

### Use a Built-in View

```bash
# View today's notes
opennotes notes view today

# View recent notes (last 20 modified)
opennotes notes view recent

# View notes by status (kanban)
opennotes notes view kanban

# Find untagged notes
opennotes notes view untagged

# Find orphaned notes (no incoming links)
opennotes notes view orphans

# Find broken links
opennotes notes view broken-links
```

### Use Parameters

```bash
# Kanban view with custom statuses
opennotes notes view kanban --param status=todo,in-progress,done

# Orphans view with different definition
opennotes notes view orphans --param definition=isolated
```

### Change Output Format

```bash
# Default: List format (markdown-style)
opennotes notes view today

# Table format (ASCII table)
opennotes notes view today --format table

# JSON format (for automation)
opennotes notes view today --format json
```

---

## Built-in Views

OpenNotes includes 6 pre-configured views for common workflows.

### 1. Today View

**Purpose**: Shows notes modified today

**Query**: All notes where `updated_at` or `created_at` is today

**Use Cases**:

- Daily standup preparation
- End-of-day review
- Quick access to active work

**Example**:

```bash
opennotes notes view today
```

**Output**:

```
### Today's Notes (3)

- [Meeting Notes] meetings/2026-01-24-standup.md
- [Project Update] projects/website-redesign.md
- [Todo List] daily/2026-01-24.md
```

---

### 2. Recent View

**Purpose**: Shows the 20 most recently modified notes

**Query**: All notes ordered by `updated_at DESC LIMIT 20`

**Use Cases**:

- Finding recently edited notes
- Reviewing recent activity
- Quick navigation to active notes

**Example**:

```bash
opennotes notes view recent
```

**Output**:

```
### Recent Notes (20)

- [Meeting Notes] meetings/2026-01-24-standup.md
- [Project Update] projects/website-redesign.md
- [Todo List] daily/2026-01-24.md
...
```

---

### 3. Kanban View

**Purpose**: Organizes notes by status field (task management)

**Parameters**:

- `status` (list, optional): Status values to filter (default: `todo,in-progress,done`)

**Query**: Notes where `metadata->>'status'` IN specified values, grouped by status, ordered by priority

**Use Cases**:

- Project task tracking
- Sprint planning
- Personal todo management

**Examples**:

```bash
# Default statuses
opennotes notes view kanban

# Custom statuses
opennotes notes view kanban --param status=todo,blocked,done

# Only in-progress items
opennotes notes view kanban --param status=in-progress
```

**Output**:

```
### Kanban Board (12)

**Todo** (5)
- [Feature Request] features/add-export.md
- [Bug Fix] bugs/search-crash.md

**In Progress** (4)
- [Documentation] docs/views-guide.md
- [Testing] tests/integration-tests.md

**Done** (3)
- [Release Notes] releases/v1.2.0.md
```

---

### 4. Untagged View

**Purpose**: Finds notes without tags

**Query**: Notes where `metadata->>'tags'` is NULL or empty

**Use Cases**:

- Content organization audit
- Finding notes that need categorization
- Cleanup tasks

**Example**:

```bash
opennotes notes view untagged
```

**Output**:

```
### Untagged Notes (8)

- [Random Thoughts] inbox/2026-01-20-ideas.md
- [Quick Note] scratch/temp.md
```

---

### 5. Orphans View

**Purpose**: Finds isolated or disconnected notes

**Parameters**:

- `definition` (string, optional): Orphan definition (default: `no-incoming`)
  - `no-incoming`: Notes with no incoming links (nothing points to them)
  - `no-links`: Notes with neither incoming nor outgoing links
  - `isolated`: Notes completely disconnected from graph

**Query**: Graph analysis to find notes matching orphan definition

**Use Cases**:

- Knowledge graph maintenance
- Finding disconnected content
- Content integration planning

**Examples**:

```bash
# Default: Notes with no incoming links
opennotes notes view orphans

# Notes with no links at all
opennotes notes view orphans --param definition=no-links

# Completely isolated notes
opennotes notes view orphans --param definition=isolated
```

**Output**:

```
### Orphaned Notes (no-incoming) (15)

- [Draft Post] blog/drafts/unpublished-idea.md
- [Old Notes] archive/2025/meeting-notes.md
```

---

### 6. Broken Links View

**Purpose**: Finds notes with broken references

**Query**: Graph analysis to detect links pointing to non-existent notes

**Use Cases**:

- Link integrity checking
- Cleanup before publishing
- Preventing 404 errors in exported docs

**Example**:

```bash
opennotes notes view broken-links
```

**Output**:

```
### Notes with Broken Links (4)

- [Project Overview] projects/main.md
  Broken: [[old-architecture]] (referenced but doesn't exist)

- [Meeting Notes] meetings/2026-01-15.md
  Broken: [[action-items]] (referenced but doesn't exist)
```

---

## Creating Custom Views

You can define custom views in two locations:

1. **Global Config** (`~/.config/opennotes/config.json`): Available to all notebooks
2. **Notebook Config** (`.opennotes.json`): Available only in that notebook

### Basic Custom View

Add to your config file:

```json
{
  "views": [
    {
      "name": "urgent",
      "description": "Notes marked as urgent",
      "query": {
        "conditions": [
          {
            "field": "metadata->>'priority'",
            "operator": "=",
            "value": "urgent"
          }
        ]
      }
    }
  ]
}
```

Then use it:

```bash
opennotes notes view urgent
```

### View with Parameters

```json
{
  "views": [
    {
      "name": "by-author",
      "description": "Notes by specific author",
      "parameters": [
        {
          "name": "author",
          "type": "string",
          "required": true,
          "description": "Author name to filter"
        }
      ],
      "query": {
        "conditions": [
          {
            "field": "metadata->>'author'",
            "operator": "=",
            "value": "{{author}}"
          }
        ]
      }
    }
  ]
}
```

Usage:

```bash
opennotes notes view by-author --param author="John Doe"
```

### View with Template Variables

```json
{
  "views": [
    {
      "name": "this-week",
      "description": "Notes modified this week",
      "query": {
        "conditions": [
          {
            "field": "updated_at",
            "operator": ">=",
            "value": "{{this_week}}"
          }
        ]
      }
    }
  ]
}
```

Usage:

```bash
opennotes notes view this-week
```

### Complex View with Multiple Conditions

```json
{
  "views": [
    {
      "name": "active-tasks",
      "description": "Urgent or in-progress tasks",
      "query": {
        "conditions": [
          {
            "field": "metadata->>'status'",
            "operator": "IN",
            "value": ["todo", "in-progress"]
          },
          {
            "field": "metadata->>'priority'",
            "operator": "IN",
            "value": ["high", "urgent"]
          }
        ],
        "orderBy": "(metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC"
      }
    }
  ]
}
```

---

## Parameter System

Parameters make views flexible and reusable.

### Parameter Types

| Type     | Description                   | Example Value                   |
| -------- | ----------------------------- | ------------------------------- |
| `string` | Single text value             | `"todo"`                        |
| `list`   | Comma-separated values        | `"todo,done"`                   |
| `date`   | ISO date or template variable | `"2026-01-24"` or `"{{today}}"` |
| `bool`   | Boolean value                 | `true` or `false`               |

### Defining Parameters

```json
{
  "parameters": [
    {
      "name": "status",
      "type": "list",
      "required": false,
      "default": ["todo", "done"],
      "description": "Task statuses to include"
    },
    {
      "name": "author",
      "type": "string",
      "required": true,
      "description": "Note author name"
    },
    {
      "name": "after_date",
      "type": "date",
      "required": false,
      "default": "{{this_week}}",
      "description": "Filter notes after this date"
    }
  ]
}
```

### Using Parameters

```bash
# Single parameter
opennotes notes view my-view --param author="Alice"

# Multiple parameters
opennotes notes view my-view --param status=todo,done --param author="Bob"

# Using default values (omit parameter)
opennotes notes view my-view
```

### Parameter Validation

OpenNotes validates parameters before query execution:

- **Required parameters**: Must be provided or have defaults
- **Type checking**: Values must match declared types
- **Format validation**: Dates must be valid ISO format

**Error Example**:

```bash
$ opennotes notes view by-author
Error: Required parameter 'author' not provided

$ opennotes notes view by-author --param author=
Error: Parameter 'author' cannot be empty

$ opennotes notes view this-week --param date=invalid
Error: Parameter 'date' must be valid ISO date format
```

---

## Template Variables

Template variables provide dynamic values in view queries.

### Available Variables

| Variable         | Description             | Example Value          |
| ---------------- | ----------------------- | ---------------------- |
| `{{today}}`      | Current date (midnight) | `2026-01-24T00:00:00Z` |
| `{{yesterday}}`  | Yesterday's date        | `2026-01-23T00:00:00Z` |
| `{{this_week}}`  | Start of current week   | `2026-01-20T00:00:00Z` |
| `{{this_month}}` | Start of current month  | `2026-01-01T00:00:00Z` |
| `{{now}}`        | Current timestamp       | `2026-01-24T15:30:45Z` |

### Using Template Variables

**In View Definitions**:

```json
{
  "query": {
    "conditions": [
      {
        "field": "updated_at",
        "operator": ">=",
        "value": "{{today}}"
      }
    ]
  }
}
```

**In Parameter Defaults**:

```json
{
  "parameters": [
    {
      "name": "since",
      "type": "date",
      "default": "{{this_week}}"
    }
  ]
}
```

**Combined with Parameters**:

```json
{
  "parameters": [
    {
      "name": "date",
      "type": "date",
      "required": false,
      "default": "{{today}}"
    }
  ],
  "query": {
    "conditions": [
      {
        "field": "created_at",
        "operator": "=",
        "value": "{{date}}"
      }
    ]
  }
}
```

---

## Output Formatting

Views support three output formats.

### List Format (Default)

**Description**: Markdown-style list with note titles and paths

**Use When**: Human-readable output for terminal viewing

**Example**:

```bash
opennotes notes view today
```

**Output**:

```
### Today's Notes (3)

- [Meeting Notes] meetings/2026-01-24-standup.md
- [Project Update] projects/website-redesign.md
- [Todo List] daily/2026-01-24.md
```

---

### Table Format

**Description**: ASCII table with all note fields

**Use When**: Comparing multiple notes or viewing metadata

**Example**:

```bash
opennotes notes view today --format table
```

**Output**:

```
+------------------+--------------------------------+------------+
| Title            | Path                           | Updated At |
+------------------+--------------------------------+------------+
| Meeting Notes    | meetings/2026-01-24-standup.md | 2026-01-24 |
| Project Update   | projects/website-redesign.md   | 2026-01-24 |
| Todo List        | daily/2026-01-24.md            | 2026-01-24 |
+------------------+--------------------------------+------------+
```

---

### JSON Format

**Description**: Valid JSON array of note objects

**Use When**: Automation, scripting, or piping to other tools

**Example**:

```bash
opennotes notes view today --format json
```

**Output**:

```json
[
  {
    "path": "meetings/2026-01-24-standup.md",
    "title": "Meeting Notes",
    "created_at": "2026-01-24T09:00:00Z",
    "updated_at": "2026-01-24T09:30:00Z",
    "data": {
      "tags": ["meeting", "team"],
      "status": "done"
    }
  }
]
```

**Piping to jq**:

```bash
# Extract only paths
opennotes notes view today --format json | jq -r '.[].path'

# Filter by status
opennotes notes view kanban --format json | jq '.[] | select(.metadata.status == "todo")'

# Count results
opennotes notes view untagged --format json | jq '. | length'
```

---

## Configuration Precedence

Views can be defined in three locations with specific precedence:

### 1. Built-in Views (Lowest Priority)

Location: Embedded in OpenNotes binary

Examples: `today`, `recent`, `kanban`, `untagged`, `orphans`, `broken-links`

**Cannot be modified** without recompiling OpenNotes.

---

### 2. Global Views (Medium Priority)

Location: `~/.config/opennotes/config.json`

**Use When**: Views should be available across all notebooks

**Example Config**:

```json
{
  "views": [
    {
      "name": "my-global-view",
      "description": "Available in all notebooks",
      "query": { ... }
    }
  ]
}
```

**Overrides**: Built-in views (if same name)

---

### 3. Notebook Views (Highest Priority)

Location: `.opennotes.json` in notebook root

**Use When**: Views are specific to a project or notebook

**Example Config**:

```json
{
  "notebook": {
    "name": "My Project"
  },
  "views": [
    {
      "name": "my-notebook-view",
      "description": "Only in this notebook",
      "query": { ... }
    }
  ]
}
```

**Overrides**: Global and built-in views (if same name)

---

### Precedence Example

Given:

- Built-in view: `today`
- Global config: `today` (custom definition)
- Notebook config: `today` (custom definition)

**Result**: Notebook config's `today` view is used

**Listing Views**:

```bash
opennotes notes view --list
```

Shows views from all three sources, with notebook views taking precedence.

---

## Troubleshooting

### View Not Found

**Error**: `Error: View 'my-view' not found`

**Causes**:

1. View name misspelled
2. View only defined in different notebook
3. View config has syntax error

**Solutions**:

```bash
# List available views
opennotes notes view --list

# Check view exists in config
cat ~/.config/opennotes/config.json | jq '.views'
cat .opennotes.json | jq '.views'

# Validate JSON syntax
jq empty ~/.config/opennotes/config.json
```

---

### Required Parameter Missing

**Error**: `Error: Required parameter 'author' not provided`

**Cause**: View requires parameter but not supplied

**Solution**:

```bash
# Provide required parameter
opennotes notes view by-author --param author="Alice"

# Check view definition for required parameters
opennotes notes view --list --format json | jq '.[] | select(.name == "by-author") | .parameters'
```

---

### Invalid Parameter Format

**Error**: `Error: Parameter 'status' must be a list (comma-separated values)`

**Cause**: Parameter type mismatch

**Solution**:

```bash
# Correct format for list parameters
opennotes notes view kanban --param status=todo,done

# Not: --param status=todo (single value when list expected)
```

---

### Template Variable Not Resolved

**Error**: `Error: Unknown template variable '{{invalid}}'`

**Cause**: Typo in template variable name

**Solution**:

```bash
# Valid template variables:
# {{today}}, {{yesterday}}, {{this_week}}, {{this_month}}, {{now}}

# Check view definition
cat .opennotes.json | jq '.views[] | select(.name == "my-view") | .query'
```

---

### No Results Returned

**Behavior**: View executes but returns empty results

**Causes**:

1. No notes match criteria
2. Query conditions too restrictive
3. Wrong notebook context

**Debug Steps**:

```bash
# Check notebook context
opennotes notebooks info

# Test with broader criteria
opennotes notes list

# Check if notes have expected metadata
opennotes notes search --sql "SELECT path, data FROM read_markdown('**/*.md') LIMIT 5"

# Verify view query is correct
opennotes notes view --list --format json | jq '.[] | select(.name == "my-view") | .query'
```

---

### Performance Issues

**Behavior**: View takes too long to execute

**Causes**:

1. Large notebook (1000+ notes)
2. Complex graph analysis (orphans, broken-links)
3. Slow disk I/O

**Solutions**:

```bash
# Use simpler views for large notebooks
opennotes notes view recent  # Faster than orphans

# Limit results
# (Modify view to include LIMIT clause)

# Check notebook size
find . -name "*.md" | wc -l

# Profile query
time opennotes notes view my-view
```

---

## Next Steps

- **Examples**: See [views-examples.md](views-examples.md) for real-world use cases
- **API Reference**: See [views-api.md](views-api.md) for complete schema documentation
- **SQL Guide**: See [sql-guide.md](sql-guide.md) for custom SQL queries
- **Automation**: See [automation-recipes.md](automation-recipes.md) for scripting examples

---

**Last Updated**: 2026-01-24  
**Version**: 1.0.0
