# @zenobi-us/pi-opennotes

Pi extension for [OpenNotes](https://github.com/zenobi-us/opennotes) - search and manage markdown notes with AI.

## Overview

This extension integrates OpenNotes into the [pi coding agent](https://github.com/mariozechner/pi-coding-agent), enabling AI assistants to:

- 🔍 **Search notes** using text, fuzzy, SQL, or boolean queries
- 📝 **Create notes** with templates and frontmatter
- 📚 **List notes** with sorting and filtering
- 📖 **Get note content** with full metadata
- 📓 **Manage notebooks** across multiple projects
- 👁️ **Execute views** for predefined queries

## Prerequisites

- [OpenNotes CLI](https://github.com/zenobi-us/opennotes) installed and in PATH
- [pi coding agent](https://github.com/mariozechner/pi-coding-agent) >= 0.50.0

Install OpenNotes:

```bash
go install github.com/zenobi-us/opennotes@latest
```

## Installation

### Via pi packages

Add to your `~/.pi/settings.json`:

```json
{
  "packages": ["npm:@zenobi-us/pi-opennotes"]
}
```

### Manual installation

```bash
npm install -g @zenobi-us/pi-opennotes
```

Then add to your pi extensions in `~/.pi/settings.json`:

```json
{
  "extensions": ["@zenobi-us/pi-opennotes"]
}
```

## Tools

The extension registers 6 tools:

### `opennotes_search`

Search notes using multiple methods:

```typescript
// Text search
{ query: "meeting notes" }

// Fuzzy search (typo-tolerant)
{ query: "meetng", fuzzy: true }

// SQL query
{ sql: "SELECT * FROM read_markdown('**/*.md') WHERE content LIKE '%TODO%' LIMIT 10" }

// Boolean filters
{
  filters: {
    and: ["data.tag=project"],
    or: ["data.status=active", "data.status=pending"],
    not: ["data.archived=true"]
  }
}
```

### `opennotes_list`

List notes with sorting and filtering:

```typescript
{
  sortBy: "modified",     // modified, created, title, path
  sortOrder: "desc",      // asc, desc
  pattern: "tasks/*.md",  // glob pattern
  limit: 50,
  offset: 0
}
```

### `opennotes_get`

Get full note content:

```typescript
{
  path: "projects/alpha.md",
  includeContent: true  // false for metadata only
}
```

### `opennotes_create`

Create a new note:

```typescript
{
  title: "Meeting Notes",
  path: "meetings/",           // directory within notebook
  template: "meeting",         // optional template
  content: "## Attendees\n...", // initial content
  data: {                      // frontmatter fields
    tag: "meeting",
    date: "2026-01-28"
  }
}
```

### `opennotes_notebooks`

List all available notebooks:

```typescript
// No parameters - returns all notebooks
{}
```

### `opennotes_views`

List or execute views:

```typescript
// List available views
{}

// Execute a view
{
  view: "kanban",
  params: { status: "todo,in-progress,done" },
  limit: 50
}
```

Built-in views: `today`, `recent`, `kanban`, `untagged`, `orphans`, `broken-links`

## Configuration

Configure via `~/.pi/settings.json`:

```json
{
  "config": {
    "@zenobi-us/pi-opennotes": {
      "toolPrefix": "opennotes_",    // Tool name prefix
      "defaultPageSize": 50,          // Results per page
      "cliPath": "opennotes",         // Path to CLI binary
      "cliTimeout": 30000             // Command timeout (ms)
    }
  }
}
```

Or via environment variables:

- `OPENNOTES_TOOL_PREFIX` - Tool name prefix
- `OPENNOTES_PAGE_SIZE` - Default page size
- `OPENNOTES_CLI_PATH` - Path to CLI binary
- `OPENNOTES_CLI_TIMEOUT` - Command timeout

## Pagination

All list/search operations support pagination:

```typescript
// First page
{ query: "meeting", limit: 50 }

// Response includes pagination metadata
{
  results: [...],
  pagination: {
    total: 127,
    returned: 50,
    page: 1,
    hasMore: true,
    nextOffset: 50
  }
}

// Next page
{ query: "meeting", limit: 50, offset: 50 }
```

## Error Handling

Errors include helpful hints:

```typescript
{
  error: true,
  message: "OpenNotes CLI not found",
  code: "OPENNOTES_CLI_NOT_FOUND",
  hint: "Install with: go install github.com/zenobi-us/opennotes@latest",
  recoverable: true
}
```

## Commands

The extension also registers a `/opennotes` command for status:

```
/opennotes
```

Shows CLI version and available notebooks.

## Development

```bash
# Install dependencies
bun install

# Run tests
bun test

# Type check
bun run typecheck
```

## License

MIT
