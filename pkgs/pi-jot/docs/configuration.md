# Configuration Reference

Complete reference for configuring `@zenobi-us/pi-jot`.

## Table of Contents

- [Pi Settings](#pi-settings)
- [Environment Variables](#environment-variables)
- [Notebook Configuration](#notebook-configuration)
- [Templates](#templates)
- [Views](#views)
- [Advanced Options](#advanced-options)

---

## Pi Settings

Configure in `~/.pi/settings.json`:

```json
{
  "packages": [
    "npm:@zenobi-us/pi-jot"
  ],
  "config": {
    "@zenobi-us/pi-jot": {
      "toolPrefix": "jot_",
      "defaultPageSize": 50,
      "cliPath": "jot",
      "cliTimeout": 30000
    }
  }
}
```

### Options

#### `toolPrefix`

- **Type**: `string`
- **Default**: `"jot_"`
- **Env**: `JOT_TOOL_PREFIX`

Prefix for all tool names.

**Examples**:
```json
"toolPrefix": "jot_"  // jot_search, jot_list
"toolPrefix": "notes_"      // notes_search, notes_list
"toolPrefix": "on_"         // on_search, on_list
```

#### `defaultPageSize`

- **Type**: `number`
- **Default**: `50`
- **Range**: `1` - `1000`
- **Env**: `JOT_PAGE_SIZE`

Default number of results per page when `limit` not specified.

**Examples**:
```json
"defaultPageSize": 20   // Smaller pages for faster responses
"defaultPageSize": 100  // Larger pages for comprehensive results
```

**Considerations**:
- Smaller values = faster responses, more pagination
- Larger values = fewer requests, higher memory usage
- Recommended: 20-100 depending on use case

#### `cliPath`

- **Type**: `string`
- **Default**: `"jot"`
- **Env**: `JOT_CLI_PATH`

Path to Jot CLI binary.

**Examples**:
```json
"cliPath": "jot"                      // In PATH
"cliPath": "/usr/local/bin/jot"       // Absolute path
"cliPath": "/home/user/go/bin/jot"    // User install
"cliPath": "./bin/jot"                // Relative path
```

**When to set**:
- CLI not in PATH
- Multiple Jot versions installed
- Custom installation location

#### `cliTimeout`

- **Type**: `number` (milliseconds)
- **Default**: `30000` (30 seconds)
- **Range**: `1000` - `300000` (1s - 5min)
- **Env**: `JOT_CLI_TIMEOUT`

Maximum time to wait for CLI command completion.

**Examples**:
```json
"cliTimeout": 5000   // 5 seconds - quick queries only
"cliTimeout": 30000  // 30 seconds - default
"cliTimeout": 60000  // 1 minute - large notebooks
"cliTimeout": 120000 // 2 minutes - complex queries
```

**Factors**:
- Notebook size (more notes = longer queries)
- Query complexity (joins, aggregations)
- Filesystem speed (network drives slower)
- System load

---

## Environment Variables

Alternative to pi settings - useful for per-session config.

### Setting Variables

```bash
# Bash/Zsh
export JOT_TOOL_PREFIX="notes_"
export JOT_PAGE_SIZE="100"
export JOT_CLI_PATH="/usr/local/bin/jot"
export JOT_CLI_TIMEOUT="60000"

# Persist
echo 'export JOT_PAGE_SIZE="100"' >> ~/.bashrc
source ~/.bashrc
```

### Precedence

1. **Environment variables** (highest)
2. **Pi settings** (`~/.pi/settings.json`)
3. **Defaults** (lowest)

**Example**:
```bash
# Pi settings: defaultPageSize = 50
# Environment: JOT_PAGE_SIZE=100
# Result: 100 (env overrides settings)
```

### Use Cases

**Development**:
```bash
# Use development CLI build
export JOT_CLI_PATH="./dist/jot-dev"
export JOT_CLI_TIMEOUT="120000"  # Longer timeout for debugging
```

**Testing**:
```bash
# Smaller pages for testing pagination
export JOT_PAGE_SIZE="5"
```

**Different notebooks**:
```bash
# Project-specific settings
cd ~/project-a
export JOT_PAGE_SIZE="20"

cd ~/project-b
export JOT_PAGE_SIZE="100"
```

---

## Notebook Configuration

Each notebook has `.jot.json` at its root.

### Minimal Config

```json
{
  "name": "My Notes"
}
```

### Complete Config

```json
{
  "name": "Work Notes",
  "description": "Work-related notes and tasks",
  "version": "1.0.0",
  "templates": {
    "meeting": {
      "content": "## Attendees\n\n## Agenda\n\n## Notes\n",
      "data": {
        "type": "meeting",
        "tags": ["meeting"]
      }
    }
  },
  "views": {
    "active-tasks": {
      "description": "All active tasks",
      "sql": "SELECT * FROM notes WHERE data->>'status' = 'active'",
      "parameters": {
        "limit": { "type": "number", "default": 20 }
      }
    }
  },
  "indexes": {
    "status": "data->>'status'",
    "priority": "data->>'priority'"
  }
}
```

### Fields

#### `name`

- **Type**: `string`
- **Required**: Yes

Human-readable notebook name.

```json
"name": "Personal Notes"
```

#### `description`

- **Type**: `string`
- **Optional**: Yes

Longer description of notebook purpose.

```json
"description": "Personal journal entries and ideas"
```

#### `version`

- **Type**: `string`
- **Optional**: Yes
- **Format**: Semantic versioning

Notebook schema version.

```json
"version": "1.2.0"
```

Use when:
- Migrating between formats
- Tracking major changes
- Team coordination

---

## Templates

Templates define note structures for quick creation.

### Basic Template

```json
{
  "templates": {
    "simple": {
      "content": "# {{title}}\n\nContent here."
    }
  }
}
```

### Template with Frontmatter

```json
{
  "templates": {
    "task": {
      "content": "## Description\n\n## Progress\n\n- [ ] TODO\n",
      "data": {
        "type": "task",
        "status": "todo",
        "tags": ["task"]
      }
    }
  }
}
```

### Template Fields

#### `content`

- **Type**: `string`
- **Required**: Yes

Markdown content template.

**Variables**:
- `{{title}}` - Note title
- `{{date}}` - Current date (YYYY-MM-DD)
- `{{datetime}}` - Current timestamp (ISO 8601)
- `{{author}}` - User name (from git config)
- Custom variables from `prompts`

#### `data`

- **Type**: `object`
- **Optional**: Yes

Default frontmatter fields.

```json
"data": {
  "type": "project",
  "status": "planning",
  "tags": ["project"],
  "priority": "medium"
}
```

#### `prompts`

- **Type**: `object`
- **Optional**: Yes

Interactive prompts for variables.

```json
"prompts": {
  "overview": {
    "type": "text",
    "label": "Project overview",
    "required": true
  },
  "start_date": {
    "type": "date",
    "label": "Start date",
    "default": "{{date}}"
  }
}
```

### Complete Example

```json
{
  "templates": {
    "project": {
      "content": "# {{title}}\n\n## Overview\n\n{{overview}}\n\n## Goals\n\n{{goals}}\n\n## Timeline\n\n- Start: {{start_date}}\n- Target: {{target_date}}\n",
      "data": {
        "type": "project",
        "status": "planning",
        "tags": ["project"],
        "owner": "{{author}}"
      },
      "prompts": {
        "overview": {
          "type": "text",
          "label": "Brief project description",
          "required": true
        },
        "goals": {
          "type": "textarea",
          "label": "Key objectives"
        },
        "start_date": {
          "type": "date",
          "label": "Project start date",
          "default": "{{date}}"
        },
        "target_date": {
          "type": "date",
          "label": "Target completion"
        }
      }
    }
  }
}
```

---

## Views

Views define reusable SQL queries.

### Basic View

```json
{
  "views": {
    "all-notes": {
      "description": "All notes",
      "sql": "SELECT * FROM notes"
    }
  }
}
```

### View with Parameters

```json
{
  "views": {
    "by-status": {
      "description": "Notes by status",
      "sql": "SELECT * FROM notes WHERE data->>'status' = :status LIMIT :limit",
      "parameters": {
        "status": {
          "type": "string",
          "default": "active",
          "required": true
        },
        "limit": {
          "type": "number",
          "default": 20,
          "min": 1,
          "max": 1000
        }
      }
    }
  }
}
```

### View Fields

#### `description`

- **Type**: `string`
- **Required**: Yes

Human-readable view description.

#### `sql`

- **Type**: `string`
- **Required**: Yes

SQL query with optional `:parameter` placeholders.

**Available tables**:
- `notes` - All notes with metadata

**Schema**:
```sql
CREATE TABLE notes (
  path TEXT,
  title TEXT,
  content TEXT,
  data JSONB,      -- Frontmatter
  created TIMESTAMP,
  modified TIMESTAMP
);
```

**Parameters**:
```sql
-- Use :parameter_name
SELECT * FROM notes 
WHERE data->>'status' = :status 
  AND data->>'priority' = :priority
LIMIT :limit
```

#### `parameters`

- **Type**: `object`
- **Optional**: Yes

Parameter definitions.

**Fields**:
```json
"parameters": {
  "param_name": {
    "type": "string|number|boolean|date",
    "default": <value>,
    "required": true|false,
    "min": <number>,      // For numbers
    "max": <number>,      // For numbers
    "pattern": "<regex>", // For strings
    "enum": [...]         // Allowed values
  }
}
```

### Built-in Views

These are always available:

#### `today`

Notes created or modified today.

```typescript
{ view: "today" }
```

#### `recent`

Recently modified notes.

```typescript
{ view: "recent", params: { days: 7 } }
```

Parameters:
- `days` (number): Days to look back (default: 7)

#### `kanban`

Task board view.

```typescript
{ view: "kanban", params: { status: "todo,in-progress,done" } }
```

Parameters:
- `status` (string): Comma-separated statuses (default: all)

#### `untagged`

Notes without tags.

```typescript
{ view: "untagged" }
```

#### `orphans`

Notes with no links to/from other notes.

```typescript
{ view: "orphans" }
```

#### `broken-links`

Notes with broken links.

```typescript
{ view: "broken-links" }
```

### View Examples

#### Project Dashboard

```json
{
  "views": {
    "project-dashboard": {
      "description": "Active projects with stats",
      "sql": "SELECT path, title, data->>'status' as status, COUNT(DISTINCT data->>'assignee') as team_size FROM notes WHERE data->>'type' = 'project' AND data->>'status' IN ('planning', 'active') GROUP BY path, title, status ORDER BY status, title"
    }
  }
}
```

#### Overdue Tasks

```json
{
  "views": {
    "overdue": {
      "description": "Tasks past due date",
      "sql": "SELECT path, title, data->>'due_date' as due_date FROM notes WHERE data->>'type' = 'task' AND data->>'status' != 'done' AND CAST(data->>'due_date' AS DATE) < CURRENT_DATE ORDER BY due_date",
      "parameters": {
        "limit": { "type": "number", "default": 50 }
      }
    }
  }
}
```

#### Weekly Report

```json
{
  "views": {
    "weekly-report": {
      "description": "Activity for the week",
      "sql": "SELECT DATE(modified) as date, COUNT(*) as notes_modified, SUM(CASE WHEN data->>'status' = 'done' THEN 1 ELSE 0 END) as completed FROM notes WHERE modified >= CURRENT_DATE - INTERVAL 7 DAY GROUP BY date ORDER BY date DESC"
    }
  }
}
```

---

## Advanced Options

### Indexes

**Future feature** - Define indexes for faster queries:

```json
{
  "indexes": {
    "status_idx": "data->>'status'",
    "priority_idx": "data->>'priority'",
    "type_status_idx": "(data->>'type', data->>'status')"
  }
}
```

### Hooks

**Future feature** - Automation on events:

```json
{
  "hooks": {
    "before_create": "./scripts/validate.sh",
    "after_create": "./scripts/index.sh",
    "before_delete": "./scripts/backup.sh"
  }
}
```

### Plugins

**Future feature** - Extend functionality:

```json
{
  "plugins": [
    "@jot/plugin-git",
    "@jot/plugin-sync",
    "custom-plugin"
  ]
}
```

---

## Configuration Tips

### 1. Start Minimal

Begin with basic config:

```json
{
  "name": "Notes",
  "templates": {
    "note": {
      "content": "# {{title}}\n\n"
    }
  }
}
```

Add views and templates as needed.

### 2. Version Control

Track `.jot.json`:

```bash
cd ~/notes
git init
git add .jot.json
git commit -m "feat: initialize notebook config"
```

### 3. Template Hierarchy

Build templates incrementally:

```json
{
  "templates": {
    "base": {
      "data": {
        "created": "{{datetime}}",
        "author": "{{author}}"
      }
    },
    "meeting": {
      "extends": "base",
      "content": "## Attendees\n",
      "data": {
        "type": "meeting"
      }
    }
  }
}
```

### 4. Parameterized Views

Make views flexible:

```json
{
  "views": {
    "by-field": {
      "sql": "SELECT * FROM notes WHERE data->>:field = :value",
      "parameters": {
        "field": { "type": "string", "required": true },
        "value": { "type": "string", "required": true }
      }
    }
  }
}
```

### 5. Validate Config

Always validate JSON:

```bash
cat .jot.json | jq .
```

---

## Migration Guide

### From v0.1 to v0.2

**Changes**:
- Templates now support `prompts`
- Views require `description`
- New `indexes` field

**Migration**:

1. **Add descriptions to views**:
   ```json
   "my-view": {
     "description": "My custom view",  // Add this
     "sql": "SELECT * FROM notes"
   }
   ```

2. **Update template format**:
   ```json
   // Old
   "template": "content"
   
   // New
   "template": {
     "content": "content"
   }
   ```

3. **Test views**:
   ```bash
   jot --notebook ~/notes notebooks info
   # Verify all views listed
   ```

---

## See Also

- [Integration Guide](./integration-guide.md) - Setup instructions
- [Tool Usage Guide](./tool-usage-guide.md) - Tool documentation
- [Troubleshooting](./troubleshooting.md) - Common issues
