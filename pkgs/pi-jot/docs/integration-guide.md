# Integration Guide

Complete setup guide for using `@zenobi-us/pi-jot` with the pi coding agent.

## Prerequisites

### 1. Jot CLI

Install the Jot CLI:

```bash
go install github.com/zenobi-us/jot@latest
```

Verify installation:

```bash
jot version
# Should output: Jot 0.0.2 (or higher)
```

### 2. Pi Coding Agent

Requires pi >= 0.50.0:

```bash
npm install -g @mariozechner/pi-coding-agent
```

Verify:

```bash
pi --version
```

### 3. Initialize Notebook

Create your first notebook:

```bash
mkdir -p ~/notes/work
cd ~/notes/work
jot init
```

Edit `.jot.json` to configure:

```json
{
  "name": "Work Notes",
  "description": "Work-related notes and tasks",
  "templates": {
    "meeting": {
      "content": "## Attendees\n\n## Agenda\n\n## Notes\n\n## Action Items\n\n",
      "data": {
        "type": "meeting",
        "tags": ["meeting"]
      }
    },
    "task": {
      "content": "## Description\n\n## Requirements\n\n## Progress\n\n- [ ] TODO\n",
      "data": {
        "type": "task",
        "status": "todo",
        "tags": ["task"]
      }
    }
  },
  "views": {
    "active-tasks": {
      "description": "All active tasks",
      "sql": "SELECT path, title, data->>'priority' as priority FROM notes WHERE data->>'type' = 'task' AND data->>'status' IN ('todo', 'in-progress') ORDER BY data->>'priority' DESC LIMIT :limit",
      "parameters": {
        "limit": { "type": "number", "default": 20 }
      }
    }
  }
}
```

## Installation

### Option 1: Pi Packages (Recommended)

Add to `~/.pi/settings.json`:

```json
{
  "packages": [
    "npm:@zenobi-us/pi-jot"
  ]
}
```

Restart pi or reload config.

### Option 2: NPM Global

```bash
npm install -g @zenobi-us/pi-jot
```

Then add to `~/.pi/settings.json`:

```json
{
  "extensions": [
    "@zenobi-us/pi-jot"
  ]
}
```

### Option 3: Local Development

For development or testing:

```bash
cd pkgs/pi-jot
bun install
bun run build

# Link locally
npm link
```

Add to pi settings:

```json
{
  "extensions": [
    "@zenobi-us/pi-jot"
  ]
}
```

## Configuration

### Pi Settings

Configure in `~/.pi/settings.json`:

```json
{
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

### Environment Variables

Alternatively, use environment variables:

```bash
export JOT_TOOL_PREFIX="notes_"
export JOT_PAGE_SIZE="100"
export JOT_CLI_PATH="/usr/local/bin/jot"
export JOT_CLI_TIMEOUT="60000"
```

Add to `~/.bashrc` or `~/.zshrc` for persistence.

### Configuration Options

| Option | Env Var | Default | Description |
|--------|---------|---------|-------------|
| `toolPrefix` | `JOT_TOOL_PREFIX` | `jot_` | Prefix for tool names |
| `defaultPageSize` | `JOT_PAGE_SIZE` | `50` | Default results per page |
| `cliPath` | `JOT_CLI_PATH` | `jot` | Path to CLI binary |
| `cliTimeout` | `JOT_CLI_TIMEOUT` | `30000` | Command timeout (ms) |

## Notebook Setup

### Single Notebook

For one notebook, configure default:

```bash
# In Jot config
jot config set default.notebook ~/notes/work
```

Pi will use this when no notebook specified.

### Multiple Notebooks

Register all notebooks:

```bash
jot notebooks add work ~/notes/work
jot notebooks add personal ~/notes/personal
jot notebooks add research ~/notes/research
```

Switch between notebooks:

```typescript
// Use specific notebook
await jot_search({
  query: "meeting",
  notebook: "~/notes/work"
});
```

Or use the `/jot` command to see all:

```
/jot
```

## Usage in Pi

### Interactive Mode

Start pi and use tools:

```
> Search for notes about the meeting yesterday
```

Pi will call:
```typescript
await jot_search({
  query: "meeting yesterday"
});
```

### Direct Tool Calls

In pi prompts:

```
Please use jot_search to find all tasks with status=active
```

```
Use jot_views to show me the kanban board
```

### Command Mode

Quick status check:

```
/jot
```

Output:
```
Jot CLI: v0.0.2
Notebooks: 3
- work (127 notes)
- personal (43 notes)
- research (89 notes)
```

## Workflow Examples

### 1. Daily Standup Notes

```
Create a meeting note for today's standup. Include sections for each team member's updates and action items.
```

Pi will:
1. Call `jot_create` with meeting template
2. Pre-fill with today's date
3. Return the note path for editing

### 2. Task Management

```
Show me all high-priority tasks that are in-progress
```

Pi will:
1. Call `jot_views` with `kanban` view
2. Filter by priority=high and status=in-progress
3. Display formatted task list

### 3. Note Discovery

```
Find all notes related to the authentication system from the last 2 weeks
```

Pi will:
1. Call `jot_search` with date filter
2. Search for "authentication system"
3. Show results with snippets

### 4. Cross-Reference

```
Show me what notes link to the architecture document
```

Pi will:
1. Call `jot_get` for architecture doc
2. Extract backlinks
3. Fetch and summarize linking notes

## Advanced Integration

### Custom Skills

Create a pi skill using jot:

```typescript
// ~/.pi/skills/standup-report.ts
import { Tool } from "@pi/sdk";

export default {
  name: "standup_report",
  description: "Generate daily standup report",
  
  async execute() {
    // Get today's tasks
    const tasks = await jot_views({
      view: "active-tasks",
      params: { limit: 50 }
    });
    
    // Get recent meetings
    const meetings = await jot_search({
      sql: "SELECT * FROM notes WHERE data->>'type' = 'meeting' AND data->>'date' >= CURRENT_DATE - INTERVAL 7 DAY",
      limit: 10
    });
    
    // Format report
    return {
      tasks: tasks.results,
      meetings: meetings.results,
      summary: `${tasks.pagination.total} active tasks, ${meetings.pagination.total} meetings this week`
    };
  }
} satisfies Tool;
```

### Automation

Schedule automated tasks:

```bash
# ~/.pi/automation/daily-summary.sh
#!/bin/bash

pi prompt "Use jot_views to show today's completed tasks and create a summary note"
```

Run via cron:

```cron
0 17 * * * ~/.pi/automation/daily-summary.sh
```

### Template System

Advanced templates with variables:

```json
{
  "templates": {
    "project": {
      "content": "# {{title}}\n\n## Overview\n\n{{overview}}\n\n## Goals\n\n{{goals}}\n\n## Timeline\n\n- Start: {{start_date}}\n- Target: {{target_date}}\n\n## Resources\n\n{{resources}}\n",
      "data": {
        "type": "project",
        "status": "planning",
        "tags": ["project"]
      },
      "prompts": {
        "overview": "Brief project description",
        "goals": "Key objectives",
        "start_date": "Project start date",
        "target_date": "Target completion",
        "resources": "Team members and tools"
      }
    }
  }
}
```

## Troubleshooting

### CLI Not Found

```
Error: JOT_CLI_NOT_FOUND
```

**Solution**:

1. Verify installation: `jot version`
2. Check PATH: `which jot`
3. Set explicit path in config:
   ```json
   {
     "config": {
       "@zenobi-us/pi-jot": {
         "cliPath": "/home/user/go/bin/jot"
       }
     }
   }
   ```

### Notebook Not Found

```
Error: Notebook not found: /path/to/notebook
```

**Solution**:

1. Check path exists: `ls /path/to/notebook`
2. Verify `.jot.json` exists
3. Initialize if needed: `jot init`

### Timeout Errors

```
Error: Command timeout after 30000ms
```

**Solution**:

1. Simplify query - reduce scope
2. Increase timeout in config:
   ```json
   {
     "config": {
       "@zenobi-us/pi-jot": {
         "cliTimeout": 60000
       }
     }
   }
   ```

### Permission Errors

```
Error: EACCES: permission denied
```

**Solution**:

1. Check file permissions
2. Ensure pi can read notebook directory
3. Fix with: `chmod -R u+rw ~/notes`

## Performance Tuning

### 1. Pagination

Always paginate large result sets:

```typescript
{
  "query": "project",
  "limit": 50,  // Reasonable page size
  "offset": 0
}
```

### 2. Specific Queries

Use SQL for targeted searches:

```typescript
{
  "sql": "SELECT path, title FROM notes WHERE data->>'status' = 'active' LIMIT 20"
}
```

Faster than:
```typescript
{
  "query": "status:active"  // Full-text search
}
```

### 3. Metadata Only

Skip content when not needed:

```typescript
{
  "path": "note.md",
  "includeContent": false
}
```

### 4. View Caching

Pi may cache view results - use views for repeated queries:

```typescript
// Define once in .jot.json
{
  "views": {
    "my-frequent-query": {
      "sql": "SELECT ..."
    }
  }
}

// Use many times
await jot_views({ view: "my-frequent-query" });
```

## Migration from Other Systems

### From Obsidian

Jot is compatible with Obsidian markdown:

1. Copy vault to notebook directory
2. Run `jot init`
3. Frontmatter and links work as-is

**Differences**:
- Wikilinks: Both `[[note]]` and `[note](note.md)` supported
- Tags: Both `#tag` and `tags: [tag]` in frontmatter
- Embeds: Use standard markdown images

### From Notion

Export Notion to markdown, then:

```bash
# Clean exported files
find . -name "*.md" -exec sed -i 's/Notion-specific-syntax/markdown-syntax/g' {} \;

# Initialize notebook
jot init

# Index notes
jot notes list
```

### From Evernote

Use Evernote export to HTML, convert to markdown:

```bash
# Using pandoc
find exports -name "*.html" -exec pandoc -f html -t markdown -o {}.md {} \;

# Move to notebook
mv exports/*.md ~/notes/imported/

# Initialize
cd ~/notes
jot init
```

## Best Practices

### 1. Consistent Frontmatter

Use consistent field names:

```yaml
---
title: Note Title
type: task|meeting|project|note
status: todo|in-progress|done|archived
tags: [tag1, tag2]
created: 2026-01-28T10:00:00Z
---
```

### 2. View Organization

Group related views:

```json
{
  "views": {
    "tasks-todo": { ... },
    "tasks-in-progress": { ... },
    "tasks-done": { ... },
    "meetings-recent": { ... },
    "meetings-upcoming": { ... }
  }
}
```

### 3. Template Hierarchy

Base templates for variations:

```json
{
  "templates": {
    "meeting-base": { ... },
    "meeting-standup": { "extends": "meeting-base", ... },
    "meeting-planning": { "extends": "meeting-base", ... }
  }
}
```

### 4. Regular Maintenance

Clean up orphaned notes and broken links:

```
Use jot_views with view=orphans to find notes with no links, then review and link or archive them
```

## Next Steps

- [Tool Usage Guide](./tool-usage-guide.md) - Detailed tool documentation
- [Troubleshooting](./troubleshooting.md) - Common issues
- [Configuration Reference](./configuration.md) - All config options

## Support

- [Jot Issues](https://github.com/zenobi-us/jot/issues)
- [Pi Extension Issues](https://github.com/zenobi-us/pi-jot/issues)
- [Pi Documentation](https://github.com/mariozechner/pi-coding-agent)
