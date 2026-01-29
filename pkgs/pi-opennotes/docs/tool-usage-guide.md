# Tool Usage Guide

Comprehensive guide to using all `@zenobi-us/pi-opennotes` tools with detailed examples.

## Table of Contents

- [opennotes_search](#opennotes_search)
- [opennotes_list](#opennotes_list)
- [opennotes_get](#opennotes_get)
- [opennotes_create](#opennotes_create)
- [opennotes_notebooks](#opennotes_notebooks)
- [opennotes_views](#opennotes_views)

---

## opennotes_search

Search notes using multiple methods: text, fuzzy, SQL, or boolean filters.

### Text Search

Simple text search across all note content:

```typescript
{
  "query": "meeting notes from January"
}
```

**Returns**: Notes containing the query text, ranked by relevance.

### Fuzzy Search

Typo-tolerant search:

```typescript
{
  "query": "meetng nots",  // Will find "meeting notes"
  "fuzzy": true
}
```

**Use when**: User might have typos or approximate matches.

### SQL Query

Direct SQL access to note database:

```typescript
{
  "sql": "SELECT path, title, data->>'status' as status FROM notes WHERE data->>'priority' = 'high' ORDER BY data->>'created' DESC LIMIT 10"
}
```

**Schema**:
- `path` - File path relative to notebook
- `title` - Note title (from frontmatter or first heading)
- `content` - Full markdown content
- `data` - JSON object with all frontmatter fields
- `created`, `modified` - Timestamps

**Security**: Only SELECT and WITH statements allowed.

### Boolean Filters

Combine multiple conditions:

```typescript
{
  "filters": {
    "and": ["data.status=active", "data.type=task"],
    "or": ["data.priority=high", "data.priority=urgent"],
    "not": ["data.archived=true"]
  }
}
```

**Filter syntax**:
- `data.field=value` - Exact match
- `data.tags~keyword` - Contains match (for arrays)

### Pagination

All search methods support pagination:

```typescript
{
  "query": "project",
  "limit": 50,
  "offset": 0
}
```

**Response includes**:
```typescript
{
  "results": [...],
  "pagination": {
    "total": 127,
    "returned": 50,
    "page": 1,
    "hasMore": true,
    "nextOffset": 50
  }
}
```

### Notebook Selection

Specify which notebook to search:

```typescript
{
  "query": "meeting",
  "notebook": "/path/to/notebook"
}
```

If omitted, uses current directory or configured default.

### Complete Example

```typescript
// Find active tasks for Project Alpha, sorted by priority
{
  "sql": "SELECT path, title, data->>'priority' as priority FROM notes WHERE data->>'project' = 'alpha' AND data->>'status' = 'active' ORDER BY CASE data->>'priority' WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END LIMIT 20",
  "notebook": "~/notes/work"
}
```

---

## opennotes_list

List notes with sorting and filtering.

### Basic Listing

```typescript
{
  "limit": 50,
  "offset": 0
}
```

### Sorting

```typescript
{
  "sortBy": "modified",  // modified, created, title, path
  "sortOrder": "desc"    // asc, desc
}
```

**Sort options**:
- `modified` - Last modification time (default)
- `created` - Creation time
- `title` - Alphabetical by title
- `path` - Alphabetical by file path

### Pattern Filtering

Use glob patterns:

```typescript
{
  "pattern": "projects/**/*.md"  // All notes in projects/ subdirectories
}
```

**Common patterns**:
- `*.md` - All notes in root
- `tasks/*.md` - All notes in tasks/
- `**/meeting-*.md` - All meeting notes recursively
- `2026-*.md` - Notes with dates

### Combined Example

```typescript
{
  "sortBy": "created",
  "sortOrder": "desc",
  "pattern": "tasks/*.md",
  "limit": 20
}
```

**Use case**: List 20 most recent tasks.

---

## opennotes_get

Get full content and metadata for a specific note.

### Basic Usage

```typescript
{
  "path": "projects/alpha.md"
}
```

**Returns**:
```typescript
{
  "path": "projects/alpha.md",
  "title": "Project Alpha",
  "data": {
    "status": "active",
    "tags": ["project", "important"],
    "created": "2026-01-15T10:00:00Z"
  },
  "content": "# Project Alpha\n\n...",
  "links": ["projects/beta.md", "tasks/task-001.md"],
  "backlinks": ["README.md"],
  "created": "2026-01-15T10:00:00Z",
  "modified": "2026-01-28T14:30:00Z"
}
```

### Metadata Only

Skip content for faster response:

```typescript
{
  "path": "projects/alpha.md",
  "includeContent": false
}
```

**Use when**: You only need frontmatter data and metadata.

### Error Handling

```typescript
// Note not found
{
  "error": true,
  "code": "NOTE_NOT_FOUND",
  "message": "Note not found: missing.md",
  "hint": "Check the path is correct and relative to notebook root"
}
```

---

## opennotes_create

Create a new note with optional template.

### Basic Creation

```typescript
{
  "title": "Meeting Notes - 2026-01-28",
  "path": "meetings/",
  "content": "## Attendees\n\n- Alice\n- Bob\n\n## Discussion\n\n..."
}
```

**Note**: `path` is the directory; filename generated from title.

### With Template

```typescript
{
  "title": "Sprint Planning",
  "path": "meetings/",
  "template": "meeting"
}
```

**Requires**: Template defined in notebook's `.opennotes.json`.

### With Frontmatter

```typescript
{
  "title": "Bug Fix: Login Issue",
  "path": "tasks/",
  "data": {
    "type": "task",
    "status": "todo",
    "priority": "high",
    "tags": ["bug", "auth"],
    "assignee": "alice"
  },
  "content": "## Problem\n\nUsers cannot log in...\n\n## Solution\n\n..."
}
```

### Template + Custom Data

```typescript
{
  "title": "Q1 Planning",
  "path": "meetings/",
  "template": "meeting",
  "data": {
    "attendees": ["alice", "bob", "charlie"],
    "date": "2026-01-28"
  }
}
```

**Result**: Template applied with custom frontmatter merged in.

---

## opennotes_notebooks

List all available notebooks.

### Basic Usage

```typescript
{}  // No parameters
```

**Returns**:
```typescript
[
  {
    "name": "Work Notes",
    "path": "/home/user/notes/work",
    "description": "Work-related notes and tasks",
    "viewCount": 5,
    "noteCount": 127
  },
  {
    "name": "Personal",
    "path": "/home/user/notes/personal",
    "noteCount": 43
  }
]
```

### Use Cases

1. **Notebook selection**: Let user choose which notebook to work with
2. **Cross-notebook search**: Search across all notebooks
3. **Notebook discovery**: Find notebooks user has forgotten about

---

## opennotes_views

List available views or execute a predefined query.

### List Views

```typescript
{}  // No parameters
```

**Returns**:
```typescript
[
  {
    "name": "today",
    "description": "Notes created or modified today",
    "builtin": true
  },
  {
    "name": "kanban",
    "description": "Task board view",
    "builtin": true,
    "parameters": {
      "status": { "type": "string", "default": "todo,in-progress,done" }
    }
  },
  {
    "name": "custom-view",
    "description": "Custom project view",
    "builtin": false
  }
]
```

### Execute View

```typescript
{
  "view": "kanban",
  "limit": 50
}
```

### View with Parameters

```typescript
{
  "view": "kanban",
  "params": {
    "status": "todo,in-progress"
  },
  "limit": 30
}
```

### Built-in Views

| View | Description | Parameters |
|------|-------------|------------|
| `today` | Notes from today | none |
| `recent` | Recently modified | `days` (default: 7) |
| `kanban` | Task board | `status` (default: all) |
| `untagged` | Notes without tags | none |
| `orphans` | Notes with no links | none |
| `broken-links` | Notes with broken links | none |

### Custom Views

Defined in notebook's `.opennotes.json`:

```json
{
  "views": {
    "high-priority": {
      "description": "High priority items",
      "sql": "SELECT * FROM notes WHERE data->>'priority' = :priority ORDER BY data->>'created' DESC LIMIT :limit",
      "parameters": {
        "priority": { "type": "string", "default": "high" },
        "limit": { "type": "number", "default": 20 }
      }
    }
  }
}
```

Execute:
```typescript
{
  "view": "high-priority",
  "params": { "priority": "urgent", "limit": 10 }
}
```

---

## Error Handling

All tools return errors in consistent format:

```typescript
{
  "error": true,
  "code": "ERROR_CODE",
  "message": "Human-readable message",
  "hint": "Suggestion to fix the problem",
  "recoverable": true  // or false
}
```

### Common Error Codes

| Code | Meaning | Hint |
|------|---------|------|
| `OPENNOTES_CLI_NOT_FOUND` | CLI not installed | Install OpenNotes CLI |
| `NOTEBOOK_NOT_FOUND` | Notebook doesn't exist | Check path or init notebook |
| `NOTE_NOT_FOUND` | Note file doesn't exist | Verify path is correct |
| `INVALID_SQL` | SQL syntax error | Check query syntax |
| `VIEW_NOT_FOUND` | View doesn't exist | List views to see available |
| `PARAMETER_REQUIRED` | Missing required param | Add missing parameter |
| `TIMEOUT` | Command timed out | Simplify query or increase timeout |

---

## Performance Tips

### 1. Use Pagination

Always set reasonable limits:

```typescript
{
  "query": "project",
  "limit": 50  // Don't fetch thousands at once
}
```

### 2. Metadata-Only When Possible

Skip content for faster queries:

```typescript
{
  "path": "note.md",
  "includeContent": false
}
```

### 3. Specific Queries

Use SQL with specific filters:

```typescript
{
  "sql": "SELECT path, title FROM notes WHERE data->>'status' = 'active' LIMIT 20"
}
// Faster than fetching all and filtering client-side
```

### 4. Index Fields

Frequently queried frontmatter fields should be indexed in DuckDB for speed.

### 5. Pattern Matching

Use specific patterns:

```typescript
{
  "pattern": "2026-01/*.md"  // Specific month
}
// Better than "**/*.md" then filtering
```

---

## Best Practices

### 1. Graceful Degradation

Handle CLI not installed:

```typescript
const result = await opennotes_search({ query: "test" });

if (result.error && result.code === "OPENNOTES_CLI_NOT_FOUND") {
  console.log(result.hint);
  // Fall back to alternative method
}
```

### 2. Pagination Loop

Fetch all results:

```typescript
let offset = 0;
const limit = 50;
let hasMore = true;

while (hasMore) {
  const result = await opennotes_search({
    query: "project",
    limit,
    offset
  });
  
  // Process results
  processResults(result.results);
  
  hasMore = result.pagination.hasMore;
  offset = result.pagination.nextOffset;
}
```

### 3. Timeout Handling

Set appropriate timeouts:

```typescript
// Quick lookup
await opennotes_get({ path: "note.md" });  // Default timeout

// Complex query might need more time
await opennotes_search({
  sql: "SELECT * FROM notes WHERE ...",
  // Complex SQL might take longer
});
```

### 4. Validate User Input

Sanitize before SQL queries:

```typescript
// DON'T: Direct user input in SQL
const userInput = getUserInput();
await opennotes_search({ sql: `SELECT * FROM notes WHERE title = '${userInput}'` });

// DO: Use parameterized views
await opennotes_views({
  view: "search-by-title",
  params: { title: userInput }
});
```

---

## Troubleshooting

See [troubleshooting.md](./troubleshooting.md) for common issues and solutions.
