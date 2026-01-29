# Integration Guide

Complete setup guide for using `@zenobi-us/pi-opennotes` with the pi coding agent.

## Prerequisites

### 1. OpenNotes CLI

Install the OpenNotes CLI:

```bash
go install github.com/zenobi-us/opennotes@latest
```

Verify installation:

```bash
opennotes version
# Should output: OpenNotes 0.0.2 (or higher)
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
opennotes init
```

Edit `.opennotes.json` to configure:

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
    "npm:@zenobi-us/pi-opennotes"
  ]
}
```

Restart pi or reload config.

### Option 2: NPM Global

```bash
npm install -g @zenobi-us/pi-opennotes
```

Then add to `~/.pi/settings.json`:

```json
{
  "extensions": [
    "@zenobi-us/pi-opennotes"
  ]
}
```

### Option 3: Local Development

For development or testing:

```bash
cd pkgs/pi-opennotes
bun install
bun run build

# Link locally
npm link
```

Add to pi settings:

```json
{
  "extensions": [
    "@zenobi-us/pi-opennotes"
  ]
}
```

## Configuration

### Pi Settings

Configure in `~/.pi/settings.json`:

```json
{
  "config": {
    "@zenobi-us/pi-opennotes": {
      "toolPrefix": "opennotes_",
      "defaultPageSize": 50,
      "cliPath": "opennotes",
      "cliTimeout": 30000
    }
  }
}
```

### Environment Variables

Alternatively, use environment variables:

```bash
export OPENNOTES_TOOL_PREFIX="notes_"
export OPENNOTES_PAGE_SIZE="100"
export OPENNOTES_CLI_PATH="/usr/local/bin/opennotes"
export OPENNOTES_CLI_TIMEOUT="60000"
```

Add to `~/.bashrc` or `~/.zshrc` for persistence.

### Configuration Options

| Option | Env Var | Default | Description |
|--------|---------|---------|-------------|
| `toolPrefix` | `OPENNOTES_TOOL_PREFIX` | `opennotes_` | Prefix for tool names |
| `defaultPageSize` | `OPENNOTES_PAGE_SIZE` | `50` | Default results per page |
| `cliPath` | `OPENNOTES_CLI_PATH` | `opennotes` | Path to CLI binary |
| `cliTimeout` | `OPENNOTES_CLI_TIMEOUT` | `30000` | Command timeout (ms) |

## Notebook Setup

### Single Notebook

For one notebook, configure default:

```bash
# In OpenNotes config
opennotes config set default.notebook ~/notes/work
```

Pi will use this when no notebook specified.

### Multiple Notebooks

Register all notebooks:

```bash
opennotes notebooks add work ~/notes/work
opennotes notebooks add personal ~/notes/personal
opennotes notebooks add research ~/notes/research
```

Switch between notebooks:

```typescript
// Use specific notebook
await opennotes_search({
  query: "meeting",
  notebook: "~/notes/work"
});
```

Or use the `/opennotes` command to see all:

```
/opennotes
```

## Usage in Pi

### Interactive Mode

Start pi and use tools:

```
> Search for notes about the meeting yesterday
```

Pi will call:
```typescript
await opennotes_search({
  query: "meeting yesterday"
});
```

### Direct Tool Calls

In pi prompts:

```
Please use opennotes_search to find all tasks with status=active
```

```
Use opennotes_views to show me the kanban board
```

### Command Mode

Quick status check:

```
/opennotes
```

Output:
```
OpenNotes CLI: v0.0.2
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
1. Call `opennotes_create` with meeting template
2. Pre-fill with today's date
3. Return the note path for editing

### 2. Task Management

```
Show me all high-priority tasks that are in-progress
```

Pi will:
1. Call `opennotes_views` with `kanban` view
2. Filter by priority=high and status=in-progress
3. Display formatted task list

### 3. Note Discovery

```
Find all notes related to the authentication system from the last 2 weeks
```

Pi will:
1. Call `opennotes_search` with date filter
2. Search for "authentication system"
3. Show results with snippets

### 4. Cross-Reference

```
Show me what notes link to the architecture document
```

Pi will:
1. Call `opennotes_get` for architecture doc
2. Extract backlinks
3. Fetch and summarize linking notes

## Advanced Integration

### Custom Skills

Create a pi skill using opennotes:

```typescript
// ~/.pi/skills/standup-report.ts
import { Tool } from "@pi/sdk";

export default {
  name: "standup_report",
  description: "Generate daily standup report",
  
  async execute() {
    // Get today's tasks
    const tasks = await opennotes_views({
      view: "active-tasks",
      params: { limit: 50 }
    });
    
    // Get recent meetings
    const meetings = await opennotes_search({
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

pi prompt "Use opennotes_views to show today's completed tasks and create a summary note"
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
Error: OPENNOTES_CLI_NOT_FOUND
```

**Solution**:

1. Verify installation: `opennotes version`
2. Check PATH: `which opennotes`
3. Set explicit path in config:
   ```json
   {
     "config": {
       "@zenobi-us/pi-opennotes": {
         "cliPath": "/home/user/go/bin/opennotes"
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
2. Verify `.opennotes.json` exists
3. Initialize if needed: `opennotes init`

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
       "@zenobi-us/pi-opennotes": {
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
// Define once in .opennotes.json
{
  "views": {
    "my-frequent-query": {
      "sql": "SELECT ..."
    }
  }
}

// Use many times
await opennotes_views({ view: "my-frequent-query" });
```

## Migration from Other Systems

### From Obsidian

OpenNotes is compatible with Obsidian markdown:

1. Copy vault to notebook directory
2. Run `opennotes init`
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
opennotes init

# Index notes
opennotes notes list
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
opennotes init
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
Use opennotes_views with view=orphans to find notes with no links, then review and link or archive them
```

## Next Steps

- [Tool Usage Guide](./tool-usage-guide.md) - Detailed tool documentation
- [Troubleshooting](./troubleshooting.md) - Common issues
- [Configuration Reference](./configuration.md) - All config options

## Support

- [OpenNotes Issues](https://github.com/zenobi-us/opennotes/issues)
- [Pi Extension Issues](https://github.com/zenobi-us/pi-opennotes/issues)
- [Pi Documentation](https://github.com/mariozechner/pi-coding-agent)
