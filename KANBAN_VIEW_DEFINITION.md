# Kanban View Definition

## Overview

The built-in `kanban` view is a pre-configured view that organizes notes by their status field. It demonstrates the full capability of the view system with:
- Status filtering
- Parameter substitution
- Ordering by priority and update time
- GROUP BY support for columnar display

## Current Built-in Definition

**Location**: `internal/services/view.go` (lines 78-104)

```go
vs.builtinViews["kanban"] = &core.ViewDefinition{
    Name:        "kanban",
    Description: "Notes grouped by status column",
    Parameters: []core.ViewParameter{
        {
            Name:        "status",
            Type:        "list",
            Required:    false,
            Default:     "backlog,todo,in-progress,reviewing,testing,deploying,done",
            Description: "Comma-separated list of status values",
        },
    },
    Query: core.ViewQuery{
        Conditions: []core.ViewCondition{
            {
                Logic:    "AND",
                Field:    "metadata->>'status'",
                Operator: "IN",
                Value:    "{{status}}",
            },
        },
        OrderBy: "(metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC",
    },
}
```

## JSON Representation

```json
{
  "name": "kanban",
  "description": "Notes grouped by status column",
  "parameters": [
    {
      "name": "status",
      "type": "list",
      "required": false,
      "default": "backlog,todo,in-progress,reviewing,testing,deploying,done",
      "description": "Comma-separated list of status values"
    }
  ],
  "query": {
    "conditions": [
      {
        "logic": "AND",
        "field": "metadata->>'status'",
        "operator": "IN",
        "value": "{{status}}"
      }
    ],
    "order_by": "(metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC"
  }
}
```

## How It Works

### 1. Parameter Substitution

The `{{status}}` placeholder is replaced with the parameter value provided by the user.

**Default value**: `"backlog,todo,in-progress,reviewing,testing,deploying,done"`

Users can override this when calling the view:
```bash
opennotes notes view kanban --param status="backlog,todo,in-progress"
```

### 2. SQL Generation

After template variable resolution, the query becomes:

```sql
SELECT * FROM notes
WHERE metadata->>'status' IN ('backlog', 'todo', 'in-progress', 'reviewing', 'testing', 'deploying', 'done')
ORDER BY (metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC
```

### 3. Results

Without GROUP BY, returns all matching notes as a flat list ordered by priority (descending) and update time.

**Example output**:
```json
[
  {
    "id": "note-1",
    "title": "High priority task",
    "status": "in-progress",
    "priority": 1,
    "updated_at": "2026-01-28T10:00:00Z"
  },
  {
    "id": "note-2",
    "title": "Medium priority task",
    "status": "in-progress",
    "priority": 2,
    "updated_at": "2026-01-28T09:00:00Z"
  },
  {
    "id": "note-3",
    "title": "Backlog item",
    "status": "backlog",
    "priority": 5,
    "updated_at": "2026-01-27T15:00:00Z"
  }
]
```

## Enhanced Kanban with GROUP BY

To enable true kanban board functionality (columns per status), add `GROUP BY`:

```go
vs.builtinViews["kanban"] = &core.ViewDefinition{
    Name:        "kanban",
    Description: "Notes grouped by status column",
    Parameters: []core.ViewParameter{
        {
            Name:        "status",
            Type:        "list",
            Required:    false,
            Default:     "backlog,todo,in-progress,reviewing,testing,deploying,done",
            Description: "Comma-separated list of status values",
        },
    },
    Query: core.ViewQuery{
        Conditions: []core.ViewCondition{
            {
                Logic:    "AND",
                Field:    "metadata->>'status'",
                Operator: "IN",
                Value:    "{{status}}",
            },
        },
        OrderBy: "(metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC",
        GroupBy: "metadata->>'status'",  // ← ADD THIS LINE
    },
}
```

### SQL with GROUP BY

```sql
SELECT * FROM notes
WHERE metadata->>'status' IN ('backlog', 'todo', 'in-progress', 'reviewing', 'testing', 'deploying', 'done')
GROUP BY metadata->>'status'
ORDER BY (metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC
```

### Results with GROUP BY (Option 2 Format)

Returns `map[string][]map[string]interface{}`:

```json
{
  "backlog": [
    {
      "id": "note-3",
      "title": "Backlog item",
      "status": "backlog",
      "priority": 5,
      "updated_at": "2026-01-27T15:00:00Z"
    },
    {
      "id": "note-5",
      "title": "Another backlog",
      "status": "backlog",
      "priority": 8,
      "updated_at": "2026-01-26T10:00:00Z"
    }
  ],
  "in-progress": [
    {
      "id": "note-1",
      "title": "High priority task",
      "status": "in-progress",
      "priority": 1,
      "updated_at": "2026-01-28T10:00:00Z"
    },
    {
      "id": "note-2",
      "title": "Medium priority task",
      "status": "in-progress",
      "priority": 2,
      "updated_at": "2026-01-28T09:00:00Z"
    }
  ],
  "todo": [
    {
      "id": "note-4",
      "title": "Task for later",
      "status": "todo",
      "priority": 3,
      "updated_at": "2026-01-27T20:00:00Z"
    }
  ],
  "done": [],
  "reviewing": [],
  "testing": [],
  "deploying": []
}
```

## View Structure Reference

### ViewDefinition

| Field | Type | Required | Purpose |
|-------|------|----------|---------|
| `name` | string | Yes | Unique identifier (lowercase, alphanumeric, hyphens) |
| `description` | string | No | Human-readable description |
| `parameters` | []ViewParameter | No | Dynamic parameters |
| `query` | ViewQuery | Yes | SQL query logic |

### ViewParameter

| Field | Type | Required | Purpose |
|-------|------|----------|---------|
| `name` | string | Yes | Parameter identifier (used in {{name}} template) |
| `type` | string | Yes | Type: "string", "list", "date", "bool" |
| `required` | bool | No | Whether parameter must be provided |
| `default` | string | No | Default value if not provided |
| `description` | string | No | Help text |

### ViewQuery

| Field | Type | Required | Purpose |
|-------|------|----------|---------|
| `conditions` | []ViewCondition | No | WHERE clause conditions |
| `distinct` | bool | No | Enable SELECT DISTINCT |
| `order_by` | string | No | ORDER BY clause |
| `group_by` | string | No | GROUP BY field (returns grouped map) |
| `having` | []ViewCondition | No | HAVING clause (post-GROUP BY filter) |
| `select_columns` | []string | No | Explicit columns (instead of SELECT *) |
| `aggregate_columns` | map[string]string | No | Aggregate functions {columnName: "COUNT"} |
| `limit` | int | No | LIMIT rows |
| `offset` | int | No | OFFSET rows |

### ViewCondition

| Field | Type | Required | Purpose |
|-------|------|----------|---------|
| `logic` | string | No | "AND" or "OR" (default: "AND") |
| `field` | string | Yes | Field name (supports DuckDB expressions) |
| `operator` | string | Yes | Comparison operator |
| `value` | string | Yes | Value (supports {{template}} substitution) |

**Supported Operators**:
- `=`, `!=`
- `<`, `>`, `<=`, `>=`
- `LIKE` (pattern matching)
- `IN` (list membership)
- `IS NULL`

## Template Variables

Supported template variables for substitution:

### Time & Date
- `{{today}}` - Today's date (YYYY-MM-DD)
- `{{today-N}}` - N days ago
- `{{today+N}}` - N days from now
- `{{this_week-N}}` - N weeks ago start date
- `{{this_month-N}}` - N months ago start date
- `{{next_week}}`, `{{last_week}}`
- `{{next_month}}`, `{{last_month}}`
- `{{start_of_month}}`, `{{end_of_month}}`
- `{{start_of_quarter}}`, `{{end_of_quarter}}`
- `{{quarter}}` - Current quarter (1-4)
- `{{year}}` - Current year

### Environment Variables
- `{{env:VAR_NAME}}` - Environment variable value
- `{{env:DEFAULT_VALUE:VAR_NAME}}` - With default fallback

### Parameters
- `{{parameter_name}}` - Substitutes parameter value

## Example: Multi-Status Kanban with Aggregates

For a dashboard showing task counts per status:

```json
{
  "name": "kanban-dashboard",
  "description": "Task counts by status",
  "query": {
    "conditions": [
      {
        "field": "metadata->>'status'",
        "operator": "IN",
        "value": "backlog,todo,in-progress,done"
      }
    ],
    "group_by": "metadata->>'status'",
    "select_columns": ["metadata->>'status'"],
    "aggregate_columns": {
      "task_count": "COUNT"
    },
    "order_by": "metadata->>'status' ASC"
  }
}
```

**Returns**:
```json
{
  "backlog": [
    {
      "status": "backlog",
      "task_count": 12
    }
  ],
  "todo": [
    {
      "status": "todo",
      "task_count": 8
    }
  ],
  "in-progress": [
    {
      "status": "in-progress",
      "task_count": 3
    }
  ],
  "done": [
    {
      "status": "done",
      "task_count": 45
    }
  ]
}
```

## Usage

### Command Line
```bash
# Use default view (all statuses)
opennotes notes view kanban

# Override with specific statuses
opennotes notes view kanban --param status="todo,in-progress,done"

# Get JSON output
opennotes notes view kanban --format json
```

### Output Format

The view command returns the result of `GroupResults()`:

**Without GROUP BY**: `[]map[string]interface{}`
```bash
$ opennotes notes view kanban --format json
[
  {"id": "...", "title": "...", "status": "todo", ...},
  {"id": "...", "title": "...", "status": "in-progress", ...}
]
```

**With GROUP BY**: `map[string][]map[string]interface{}`
```bash
$ opennotes notes view kanban --format json
{
  "todo": [
    {"id": "...", "title": "...", "status": "todo", ...}
  ],
  "in-progress": [
    {"id": "...", "title": "...", "status": "in-progress", ...}
  ]
}
```

## Related Files

- **View Service**: `internal/services/view.go`
  - Built-in views initialization
  - SQL generation logic
  - Template variable resolution
  - GroupResults() implementation

- **View Types**: `internal/core/view.go`
  - ViewDefinition struct
  - ViewQuery struct
  - ViewCondition struct
  - ViewParameter struct

- **Tests**: `internal/services/view_test.go`
  - 80+ test functions covering all features
  - Test cases for GROUP BY, HAVING, aggregates, templates

## Next Steps

To enhance the kanban view for UI rendering:

1. **Add GROUP BY** to enable columnar display
2. **Client implementation** reads the grouped map
3. **Render columns** side-by-side (one per status)
4. **Drag/drop** to move notes between columns
5. **Filtering** via parameters or conditions

The data structure is ready (Option 2: pure grouped map). Any frontend can now consume it and render as needed.
