# Views System API Reference

Complete technical reference for the Jot Views System.

## Table of Contents

- [Overview](#overview)
- [ViewDefinition Schema](#viewdefinition-schema)
- [Parameter Types](#parameter-types)
- [Query Schema](#query-schema)
- [Template Variables](#template-variables)
- [Built-in Views Specifications](#built-in-views-specifications)
- [Configuration Files](#configuration-files)
- [Error Codes](#error-codes)

---

## Overview

The Views System uses JSON configuration to define reusable query presets. This document provides the complete API schema for creating custom views.

### Key Concepts

- **ViewDefinition**: Complete view specification (name, parameters, query)
- **Parameter**: Runtime input that makes views flexible
- **Query**: Query-generation instructions (conditions, ordering)
- **Template Variable**: Dynamic values resolved at runtime

---

## ViewDefinition Schema

Complete specification for a view.

### JSON Structure

```json
{
  "name": "string",
  "description": "string",
  "parameters": [
    {
      "name": "string",
      "type": "string|list|date|bool",
      "required": true|false,
      "default": "any",
      "description": "string"
    }
  ],
  "query": {
    "conditions": [
      {
        "field": "string",
        "operator": "string",
        "value": "any"
      }
    ],
    "orderBy": "string",
    "limit": 0
  }
}
```

### Field Descriptions

| Field         | Type     | Required | Description                                                       |
| ------------- | -------- | -------- | ----------------------------------------------------------------- |
| `name`        | `string` | ✅ Yes   | Unique view identifier (lowercase, alphanumeric + hyphens)        |
| `description` | `string` | ❌ No    | Human-readable description                                        |
| `parameters`  | `array`  | ❌ No    | Runtime parameters (see [Parameter Schema](#parameter-schema))    |
| `query`       | `object` | ✅ Yes   | Query-generation instructions (see [Query Schema](#query-schema)) |

### Constraints

- **`name`**: Must match `^[a-z0-9-]+$` (lowercase letters, numbers, hyphens)
- **`name`**: Must be unique within configuration scope
- **`description`**: Max 200 characters
- **`parameters`**: Max 10 parameters per view
- **`query`**: Must have at least one condition

### Example

```json
{
  "name": "urgent-tasks",
  "description": "High-priority incomplete tasks",
  "parameters": [
    {
      "name": "priority",
      "type": "list",
      "required": false,
      "default": ["high", "urgent"],
      "description": "Priority levels to include"
    }
  ],
  "query": {
    "conditions": [
      {
        "field": "metadata->>'priority'",
        "operator": "IN",
        "value": "{{priority}}"
      },
      {
        "field": "metadata->>'status'",
        "operator": "!=",
        "value": "done"
      }
    ],
    "orderBy": "metadata->>'priority' ASC, updated_at DESC",
    "limit": 50
  }
}
```

---

## Parameter Schema

Parameters allow runtime customization of views.

### JSON Structure

```json
{
  "name": "string",
  "type": "string|list|date|bool",
  "required": true|false,
  "default": "any",
  "description": "string"
}
```

### Field Descriptions

| Field         | Type      | Required | Description                                                  |
| ------------- | --------- | -------- | ------------------------------------------------------------ |
| `name`        | `string`  | ✅ Yes   | Parameter identifier (lowercase, alphanumeric + underscores) |
| `type`        | `string`  | ✅ Yes   | Data type: `string`, `list`, `date`, `bool`                  |
| `required`    | `boolean` | ❌ No    | Whether parameter must be provided (default: `false`)        |
| `default`     | varies    | ❌ No    | Default value if not provided                                |
| `description` | `string`  | ❌ No    | Human-readable description                                   |

### Constraints

- **`name`**: Must match `^[a-z0-9_]+$` (lowercase letters, numbers, underscores)
- **`name`**: Must be unique within view
- **`type`**: Must be one of: `string`, `list`, `date`, `bool`
- **`default`**: Type must match parameter type
- **`required`**: If `true`, `default` cannot be set

### Type Details

#### String Type

**Description**: Single text value

**Example**:

```json
{
  "name": "author",
  "type": "string",
  "required": true,
  "description": "Note author name"
}
```

**CLI Usage**:

```bash
jot notes view my-view --param author="Alice"
```

**Validation**:

- Cannot be empty
- Max length: 1000 characters

---

#### List Type

**Description**: Comma-separated values

**Example**:

```json
{
  "name": "status",
  "type": "list",
  "required": false,
  "default": ["todo", "done"],
  "description": "Task statuses"
}
```

**CLI Usage**:

```bash
jot notes view my-view --param status=todo,in-progress,done
```

**Validation**:

- At least one value required
- Each value: max 100 characters
- Max values: 50

---

#### Date Type

**Description**: ISO 8601 date or template variable

**Example**:

```json
{
  "name": "after_date",
  "type": "date",
  "required": false,
  "default": "{{this_week}}",
  "description": "Filter notes after this date"
}
```

**CLI Usage**:

```bash
jot notes view my-view --param after_date=2026-01-24
jot notes view my-view --param after_date="{{today}}"
```

**Validation**:

- Must be valid ISO 8601 format: `YYYY-MM-DD` or `YYYY-MM-DDTHH:MM:SSZ`
- Or valid template variable (see [Template Variables](#template-variables))

---

#### Bool Type

**Description**: Boolean value

**Example**:

```json
{
  "name": "include_archived",
  "type": "bool",
  "required": false,
  "default": false,
  "description": "Include archived notes"
}
```

**CLI Usage**:

```bash
jot notes view my-view --param include_archived=true
jot notes view my-view --param include_archived=false
```

**Validation**:

- Must be `true` or `false` (case-insensitive)

---

## Query Schema

Defines query-generation instructions.

### JSON Structure

```json
{
  "conditions": [
    {
      "field": "string",
      "operator": "string",
      "value": "any"
    }
  ],
  "orderBy": "string",
  "limit": 0
}
```

### Field Descriptions

| Field        | Type      | Required | Description                                                  |
| ------------ | --------- | -------- | ------------------------------------------------------------ |
| `conditions` | `array`   | ✅ Yes   | Query conditions (see [Condition Schema](#condition-schema)) |
| `orderBy`    | `string`  | ❌ No    | Sort instruction (field and direction)                       |
| `limit`      | `integer` | ❌ No    | Max results to return                                        |

### Constraints

- **`conditions`**: At least 1 condition required, max 20
- **`orderBy`**: Must use supported sort syntax (field:direction)
- **`limit`**: Must be positive integer (1-10000)

### Condition Schema

Individual query condition.

#### JSON Structure

```json
{
  "field": "string",
  "operator": "string",
  "value": "any"
}
```

#### Field Descriptions

| Field      | Type     | Required | Description                                                      |
| ---------- | -------- | -------- | ---------------------------------------------------------------- |
| `field`    | `string` | ✅ Yes   | Note field to filter (see [Queryable Fields](#queryable-fields)) |
| `operator` | `string` | ✅ Yes   | Comparison operator (see [Operators](#operators))                |
| `value`    | varies   | ✅ Yes   | Value to compare (string, array, template variable)              |

#### Queryable Fields

| Field            | Type       | Description                                                              |
| ---------------- | ---------- | ------------------------------------------------------------------------ |
| `path`           | `string`   | Note file path (relative to notebook root)                               |
| `title`          | `string`   | Note title (from frontmatter or filename)                                |
| `created_at`     | `datetime` | Note creation timestamp                                                  |
| `updated_at`     | `datetime` | Note last modified timestamp                                             |
| `metadata->>'*'` | varies     | Any frontmatter field (e.g., `metadata->>'status'`, `metadata->>'tags'`) |

#### Operators

| Operator   | Description       | Value Type                 | Example                                                                         |
| ---------- | ----------------- | -------------------------- | ------------------------------------------------------------------------------- |
| `=`        | Exact match       | `string`, `number`, `bool` | `{"field": "metadata->>'status'", "operator": "=", "value": "done"}`            |
| `!=`       | Not equal         | `string`, `number`, `bool` | `{"field": "metadata->>'status'", "operator": "!=", "value": "done"}`           |
| `>`        | Greater than      | `number`, `datetime`       | `{"field": "updated_at", "operator": ">", "value": "{{today}}"}`                |
| `>=`       | Greater or equal  | `number`, `datetime`       | `{"field": "updated_at", "operator": ">=", "value": "{{this_week}}"}`           |
| `<`        | Less than         | `number`, `datetime`       | `{"field": "created_at", "operator": "<", "value": "2026-01-01"}`               |
| `<=`       | Less or equal     | `number`, `datetime`       | `{"field": "created_at", "operator": "<=", "value": "{{yesterday}}"}`           |
| `IN`       | Value in list     | `array`                    | `{"field": "metadata->>'status'", "operator": "IN", "value": ["todo", "done"]}` |
| `NOT IN`   | Value not in list | `array`                    | `{"field": "metadata->>'priority'", "operator": "NOT IN", "value": ["low"]}`    |
| `LIKE`     | Pattern match     | `string`                   | `{"field": "path", "operator": "LIKE", "value": "projects/%"}`                  |
| `NOT LIKE` | Pattern not match | `string`                   | `{"field": "metadata->>'tags'", "operator": "NOT LIKE", "value": "%archive%"}`  |

#### Value Types

**String**:

```json
{ "field": "metadata->>'status'", "operator": "=", "value": "done" }
```

**Number**:

```json
{ "field": "metadata->>'priority'", "operator": ">", "value": 5 }
```

**Boolean**:

```json
{ "field": "metadata->>'archived'", "operator": "=", "value": true }
```

**Array** (for IN/NOT IN):

```json
{ "field": "metadata->>'status'", "operator": "IN", "value": ["todo", "done"] }
```

**Template Variable**:

```json
{ "field": "updated_at", "operator": ">=", "value": "{{today}}" }
```

**Parameter Reference**:

```json
{ "field": "metadata->>'author'", "operator": "=", "value": "{{author}}" }
```

---

## Template Variables

Dynamic values resolved at query execution time.

### Available Variables

| Variable         | Type       | Description                    | Example Value          |
| ---------------- | ---------- | ------------------------------ | ---------------------- |
| `{{today}}`      | `datetime` | Current date at midnight (UTC) | `2026-01-24T00:00:00Z` |
| `{{yesterday}}`  | `datetime` | Yesterday at midnight (UTC)    | `2026-01-23T00:00:00Z` |
| `{{this_week}}`  | `datetime` | Start of current week (Monday) | `2026-01-20T00:00:00Z` |
| `{{this_month}}` | `datetime` | Start of current month         | `2026-01-01T00:00:00Z` |
| `{{now}}`        | `datetime` | Current timestamp              | `2026-01-24T15:30:45Z` |

### Usage

**In Query Conditions**:

```json
{
  "field": "updated_at",
  "operator": ">=",
  "value": "{{today}}"
}
```

**In Parameter Defaults**:

```json
{
  "name": "since",
  "type": "date",
  "default": "{{this_week}}"
}
```

**Combined with Parameters**:

```json
{
  "parameters": [
    {
      "name": "date",
      "type": "date",
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

### Resolution Order

1. User-provided parameters (via `--param`)
2. Parameter defaults
3. Template variables
4. Literal values

---

## Built-in Views Specifications

Complete specifications for all 6 built-in views.

### 1. Today View

```json
{
  "name": "today",
  "description": "Notes modified or created today",
  "query": {
    "conditions": [
      {
        "field": "updated_at",
        "operator": ">=",
        "value": "{{today}}"
      }
    ],
    "orderBy": "updated_at DESC"
  }
}
```

---

### 2. Recent View

```json
{
  "name": "recent",
  "description": "20 most recently modified notes",
  "query": {
    "conditions": [],
    "orderBy": "updated_at DESC",
    "limit": 20
  }
}
```

---

### 3. Kanban View

```json
{
  "name": "kanban",
  "description": "Notes organized by status",
  "parameters": [
    {
      "name": "status",
      "type": "list",
      "required": false,
      "default": ["todo", "in-progress", "done"],
      "description": "Status values to include"
    }
  ],
  "query": {
    "conditions": [
      {
        "field": "metadata->>'status'",
        "operator": "IN",
        "value": "{{status}}"
      }
    ],
    "orderBy": "metadata->>'status' ASC, updated_at DESC"
  }
}
```

---

### 4. Untagged View

```json
{
  "name": "untagged",
  "description": "Notes without tags",
  "query": {
    "conditions": [
      {
        "field": "metadata->>'tags'",
        "operator": "IS",
        "value": "NULL"
      }
    ],
    "orderBy": "updated_at DESC"
  }
}
```

---

### 5. Orphans View

```json
{
  "name": "orphans",
  "description": "Notes with no incoming links",
  "parameters": [
    {
      "name": "definition",
      "type": "string",
      "required": false,
      "default": "no-incoming",
      "description": "Orphan definition: no-incoming, no-links, isolated"
    }
  ],
  "special_executor": "orphans"
}
```

**Note**: This view uses a special executor for graph analysis, not standard query generation.

---

### 6. Broken Links View

```json
{
  "name": "broken-links",
  "description": "Notes with broken references",
  "special_executor": "broken-links"
}
```

**Note**: This view uses a special executor for link validation, not standard query generation.

---

## Configuration Files

Views can be defined in two configuration files.

### Global Config

**Location**: `~/.config/jot/config.json`

**Scope**: All notebooks

**Structure**:

```json
{
  "views": [
    {
      "name": "global-view",
      "description": "Available everywhere",
      "query": { ... }
    }
  ]
}
```

**Example**:

```json
{
  "notebooks": {
    "root": "~/notes"
  },
  "views": [
    {
      "name": "urgent",
      "description": "Urgent tasks",
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

---

### Notebook Config

**Location**: `.jot.json` (in notebook root)

**Scope**: Current notebook only

**Structure**:

```json
{
  "notebook": {
    "name": "My Project"
  },
  "views": [
    {
      "name": "project-view",
      "description": "Project-specific view",
      "query": { ... }
    }
  ]
}
```

**Example**:

```json
{
  "notebook": {
    "name": "Engineering Notebook",
    "root": "."
  },
  "views": [
    {
      "name": "sprint",
      "description": "Current sprint tasks",
      "parameters": [
        {
          "name": "sprint_number",
          "type": "string",
          "required": true
        }
      ],
      "query": {
        "conditions": [
          {
            "field": "metadata->>'sprint'",
            "operator": "=",
            "value": "{{sprint_number}}"
          }
        ]
      }
    }
  ]
}
```

---

## Error Codes

Views System error codes and descriptions.

### View Errors

| Code                  | Message                        | Cause                             | Solution                                            |
| --------------------- | ------------------------------ | --------------------------------- | --------------------------------------------------- |
| `VIEW_NOT_FOUND`      | `View '{name}' not found`      | View doesn't exist                | Check available views with `--list`                 |
| `VIEW_INVALID_NAME`   | `Invalid view name '{name}'`   | Name doesn't match `^[a-z0-9-]+$` | Use only lowercase letters, numbers, hyphens        |
| `VIEW_DUPLICATE_NAME` | `View '{name}' already exists` | Name conflict in config           | Use unique names or override with higher precedence |

### Parameter Errors

| Code                   | Message                                    | Cause                      | Solution                                       |
| ---------------------- | ------------------------------------------ | -------------------------- | ---------------------------------------------- |
| `PARAM_REQUIRED`       | `Required parameter '{name}' not provided` | Missing required parameter | Provide via `--param` flag                     |
| `PARAM_INVALID_TYPE`   | `Parameter '{name}' must be {type}`        | Type mismatch              | Check parameter type in view definition        |
| `PARAM_INVALID_FORMAT` | `Parameter '{name}' has invalid format`    | Parsing failed             | Check format (e.g., comma-separated for lists) |
| `PARAM_UNKNOWN`        | `Unknown parameter '{name}'`               | Parameter not defined      | Check view parameters with `--list`            |

### Query Errors

| Code                     | Message                                  | Cause              | Solution                                                     |
| ------------------------ | ---------------------------------------- | ------------------ | ------------------------------------------------------------ |
| `QUERY_INVALID_FIELD`    | `Field '{field}' is not queryable`       | Invalid field name | Use valid fields (see [Queryable Fields](#queryable-fields)) |
| `QUERY_INVALID_OPERATOR` | `Operator '{operator}' is not supported` | Invalid operator   | Use valid operators (see [Operators](#operators))            |
| `QUERY_BUILD_ERROR`      | `Query generation failed: {details}`     | Query build error  | Check query syntax and field names                           |

### Template Errors

| Code                         | Message                                   | Cause                 | Solution                                                            |
| ---------------------------- | ----------------------------------------- | --------------------- | ------------------------------------------------------------------- |
| `TEMPLATE_UNKNOWN_VAR`       | `Unknown template variable '{{var}}'`     | Invalid variable name | Use valid variables (see [Template Variables](#template-variables)) |
| `TEMPLATE_RESOLUTION_FAILED` | `Failed to resolve template '{template}'` | Resolution error      | Check template syntax and parameter names                           |

---

## JSON Schema

Complete JSON Schema for view definitions (for validation).

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Jot View Definition",
  "type": "object",
  "required": ["name", "query"],
  "properties": {
    "name": {
      "type": "string",
      "pattern": "^[a-z0-9-]+$",
      "minLength": 1,
      "maxLength": 100
    },
    "description": {
      "type": "string",
      "maxLength": 200
    },
    "parameters": {
      "type": "array",
      "maxItems": 10,
      "items": {
        "type": "object",
        "required": ["name", "type"],
        "properties": {
          "name": {
            "type": "string",
            "pattern": "^[a-z0-9_]+$"
          },
          "type": {
            "enum": ["string", "list", "date", "bool"]
          },
          "required": {
            "type": "boolean"
          },
          "default": {},
          "description": {
            "type": "string"
          }
        }
      }
    },
    "query": {
      "type": "object",
      "required": ["conditions"],
      "properties": {
        "conditions": {
          "type": "array",
          "minItems": 1,
          "maxItems": 20,
          "items": {
            "type": "object",
            "required": ["field", "operator", "value"],
            "properties": {
              "field": {
                "type": "string"
              },
              "operator": {
                "enum": [
                  "=",
                  "!=",
                  ">",
                  ">=",
                  "<",
                  "<=",
                  "IN",
                  "NOT IN",
                  "LIKE",
                  "NOT LIKE",
                  "IS",
                  "IS NOT"
                ]
              },
              "value": {}
            }
          }
        },
        "orderBy": {
          "type": "string"
        },
        "limit": {
          "type": "integer",
          "minimum": 1,
          "maximum": 10000
        }
      }
    },
    "special_executor": {
      "enum": ["orphans", "broken-links"]
    }
  }
}
```

---

## Next Steps

- **User Guide**: See [views-guide.md](views-guide.md) for usage documentation
- **Examples**: See [views-examples.md](views-examples.md) for real-world use cases
- **Search Reference**: See [commands/notes-search.md](commands/notes-search.md) for query syntax

---

**Last Updated**: 2026-01-24  
**Version**: 1.0.0
