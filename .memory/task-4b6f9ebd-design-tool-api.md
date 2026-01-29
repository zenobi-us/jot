---
id: 4b6f9ebd
title: Design Tool API Specification
created_at: 2026-01-28T23:30:00+10:30
updated_at: 2026-01-29T09:15:00+10:30
status: done
epic_id: 1f41631e
phase_id: 43842f12
assigned_to: null
---

# Design Tool API Specification

## Objective

Create detailed specifications for each tool the extension will register, including TypeBox schemas, descriptions, and return formats with pagination support.

## Completed Steps

- [x] Define `opennotes_search` tool spec
- [x] Define `opennotes_list` tool spec  
- [x] Define `opennotes_get` tool spec
- [x] Define `opennotes_create` tool spec
- [x] Define `opennotes_notebooks` tool spec
- [x] Define `opennotes_views` tool spec
- [x] Document error responses for each tool
- [x] Create TypeBox schemas for all parameters
- [x] Write LLM-friendly descriptions
- [x] Define pagination format with examples

---

## Design Decisions

### Tool Naming Convention

**Prefix**: `opennotes_` (configurable via extension config)

**Rationale**: 
- Avoids collisions with other tools
- Makes tool purpose clear in logs/traces
- Prefix is configurable for white-labeling

### Pagination Strategy

**Approach**: 75% output capacity + metadata

When results exceed limits:
1. Return first N results fitting in ~75% of output budget
2. Include pagination metadata for continuation
3. Provide clear guidance for fetching more

**Output Budget**: 50KB or 2000 lines (pi defaults)
- 75% = ~37.5KB or 1500 lines for content
- 25% = ~12.5KB or 500 lines for metadata + guidance

### Error Handling

All tools return structured errors with:
- Human-readable message
- Installation hints when CLI missing
- Suggested next steps

---

## TypeBox Schema Definitions

### Shared Types

```typescript
import { Type, type Static } from "@sinclair/typebox";
import { StringEnum } from "@mariozechner/pi-ai";

// Pagination metadata returned with large result sets
export const PaginationMeta = Type.Object({
  total: Type.Number({ description: "Total number of results" }),
  returned: Type.Number({ description: "Number returned in this response" }),
  page: Type.Number({ description: "Current page (1-indexed)" }),
  pageSize: Type.Number({ description: "Results per page" }),
  hasMore: Type.Boolean({ description: "Whether more results exist" }),
  nextOffset: Type.Optional(Type.Number({ description: "Offset for next page" })),
});

export type PaginationMeta = Static<typeof PaginationMeta>;

// Note summary for list/search results
export const NoteSummary = Type.Object({
  path: Type.String({ description: "File path relative to notebook" }),
  title: Type.Optional(Type.String({ description: "Note title from frontmatter" })),
  tags: Type.Optional(Type.Array(Type.String())),
  created: Type.Optional(Type.String({ description: "ISO 8601 timestamp" })),
  modified: Type.Optional(Type.String({ description: "ISO 8601 timestamp" })),
});

export type NoteSummary = Static<typeof NoteSummary>;

// Full note content
export const NoteContent = Type.Object({
  path: Type.String(),
  title: Type.Optional(Type.String()),
  content: Type.String({ description: "Full markdown content" }),
  frontmatter: Type.Optional(Type.Record(Type.String(), Type.Unknown())),
  wordCount: Type.Optional(Type.Number()),
});

export type NoteContent = Static<typeof NoteContent>;

// Notebook info
export const NotebookInfo = Type.Object({
  name: Type.String({ description: "Notebook display name" }),
  path: Type.String({ description: "Absolute path to notebook" }),
  source: StringEnum(["registered", "ancestor", "explicit"] as const),
  noteCount: Type.Optional(Type.Number()),
});

export type NotebookInfo = Static<typeof NotebookInfo>;

// View definition
export const ViewDefinition = Type.Object({
  name: Type.String(),
  origin: StringEnum(["built-in", "notebook", "global"] as const),
  description: Type.Optional(Type.String()),
  parameters: Type.Optional(Type.Array(Type.Object({
    name: Type.String(),
    type: Type.String(),
    required: Type.Boolean(),
    default: Type.Optional(Type.String()),
    description: Type.Optional(Type.String()),
  }))),
});

export type ViewDefinition = Static<typeof ViewDefinition>;
```

---

## Tool Specifications

### 1. `opennotes_search`

**Purpose**: Search notes using SQL queries, text search, or boolean conditions.

#### Schema

```typescript
export const SearchParams = Type.Object({
  query: Type.Optional(Type.String({
    description: "Text to search for in note titles and content",
  })),
  sql: Type.Optional(Type.String({
    description: "Raw SQL query (SELECT/WITH only). Use read_markdown('**/*.md') to access notes.",
  })),
  fuzzy: Type.Optional(Type.Boolean({
    description: "Enable fuzzy matching (typo-tolerant, ranked results)",
    default: false,
  })),
  filters: Type.Optional(Type.Object({
    and: Type.Optional(Type.Array(Type.String({
      description: "AND conditions: field=value pairs (all must match)",
    }))),
    or: Type.Optional(Type.Array(Type.String({
      description: "OR conditions: field=value pairs (any must match)",
    }))),
    not: Type.Optional(Type.Array(Type.String({
      description: "NOT conditions: field=value pairs (exclusions)",
    }))),
  })),
  notebook: Type.Optional(Type.String({
    description: "Path to notebook. Omit to use current context.",
  })),
  limit: Type.Optional(Type.Number({
    description: "Maximum results to return (default: 50)",
    default: 50,
    minimum: 1,
    maximum: 1000,
  })),
  offset: Type.Optional(Type.Number({
    description: "Offset for pagination (default: 0)",
    default: 0,
    minimum: 0,
  })),
});

export type SearchParams = Static<typeof SearchParams>;
```

#### LLM Description

```
Search notes in an OpenNotes notebook using multiple methods:

1. **Text Search**: Set 'query' for substring matching
2. **Fuzzy Search**: Set 'query' + 'fuzzy: true' for typo-tolerant search
3. **SQL Query**: Set 'sql' for full DuckDB SQL power
4. **Boolean Filters**: Use 'filters' for structured AND/OR/NOT queries

Common filter fields: data.tag, data.status, data.priority, path, title, links-to, linked-by

Example SQL: SELECT * FROM read_markdown('**/*.md') WHERE content LIKE '%TODO%' LIMIT 10

Returns notes matching criteria with pagination metadata.
```

#### Response Format

```typescript
interface SearchResponse {
  results: NoteSummary[];
  pagination: PaginationMeta;
  query: {
    type: "text" | "fuzzy" | "sql" | "boolean";
    executed: string;  // Actual query/SQL executed
  };
}
```

#### Example Response

```json
{
  "results": [
    {
      "path": "projects/alpha.md",
      "title": "Project Alpha",
      "tags": ["project", "active"],
      "modified": "2026-01-28T14:30:00Z"
    },
    {
      "path": "tasks/task-001.md",
      "title": "Implement feature X",
      "tags": ["task", "alpha"],
      "modified": "2026-01-27T09:15:00Z"
    }
  ],
  "pagination": {
    "total": 47,
    "returned": 2,
    "page": 1,
    "pageSize": 50,
    "hasMore": false
  },
  "query": {
    "type": "boolean",
    "executed": "SELECT ... WHERE data.tag = 'alpha'"
  }
}
```

---

### 2. `opennotes_list`

**Purpose**: List all notes in a notebook with optional filtering and sorting.

#### Schema

```typescript
export const ListParams = Type.Object({
  notebook: Type.Optional(Type.String({
    description: "Path to notebook. Omit to use current context.",
  })),
  sortBy: Type.Optional(StringEnum(["modified", "created", "title", "path"] as const, {
    description: "Sort field (default: modified)",
    default: "modified",
  })),
  sortOrder: Type.Optional(StringEnum(["asc", "desc"] as const, {
    description: "Sort order (default: desc for dates, asc for text)",
    default: "desc",
  })),
  limit: Type.Optional(Type.Number({
    description: "Maximum results to return (default: 50)",
    default: 50,
    minimum: 1,
    maximum: 1000,
  })),
  offset: Type.Optional(Type.Number({
    description: "Offset for pagination (default: 0)",
    default: 0,
    minimum: 0,
  })),
  pattern: Type.Optional(Type.String({
    description: "Glob pattern to filter paths (default: **/*.md)",
    default: "**/*.md",
  })),
});

export type ListParams = Static<typeof ListParams>;
```

#### LLM Description

```
List all notes in an OpenNotes notebook with metadata.

Use 'sortBy' to order by: modified (default), created, title, or path.
Use 'pattern' to filter: e.g., 'tasks/*.md' for only task notes.

Returns note summaries with pagination. For full content, use opennotes_get.
```

#### Response Format

```typescript
interface ListResponse {
  notes: NoteSummary[];
  notebook: {
    name: string;
    path: string;
  };
  pagination: PaginationMeta;
}
```

---

### 3. `opennotes_get`

**Purpose**: Get the full content and metadata of a specific note.

#### Schema

```typescript
export const GetParams = Type.Object({
  path: Type.String({
    description: "Path to the note file (relative to notebook root)",
  }),
  notebook: Type.Optional(Type.String({
    description: "Path to notebook. Omit to use current context.",
  })),
  includeContent: Type.Optional(Type.Boolean({
    description: "Whether to include full markdown content (default: true)",
    default: true,
  })),
});

export type GetParams = Static<typeof GetParams>;
```

#### LLM Description

```
Get the full content and metadata of a specific note by path.

The path should be relative to the notebook root (e.g., 'tasks/task-001.md').

Set 'includeContent: false' to get only metadata without the full body (faster for large notes).
```

#### Response Format

```typescript
interface GetResponse {
  note: NoteContent;
  notebook: {
    name: string;
    path: string;
  };
}
```

---

### 4. `opennotes_create`

**Purpose**: Create a new note with optional template and metadata.

#### Schema

```typescript
export const CreateParams = Type.Object({
  title: Type.String({
    description: "Note title (required)",
    minLength: 1,
  }),
  path: Type.Optional(Type.String({
    description: "Directory path within notebook (default: root)",
  })),
  template: Type.Optional(Type.String({
    description: "Template name to use (must exist in notebook)",
  })),
  content: Type.Optional(Type.String({
    description: "Initial markdown content (after frontmatter)",
  })),
  data: Type.Optional(Type.Record(Type.String(), Type.Union([
    Type.String(),
    Type.Number(),
    Type.Boolean(),
    Type.Array(Type.String()),
  ]), {
    description: "Frontmatter fields as key-value pairs",
  })),
  notebook: Type.Optional(Type.String({
    description: "Path to notebook. Omit to use current context.",
  })),
});

export type CreateParams = Static<typeof CreateParams>;
```

#### LLM Description

```
Create a new note in an OpenNotes notebook.

Required: 'title' - becomes the note's title in frontmatter

Optional:
- 'path': Directory within notebook (e.g., 'tasks/' creates in tasks folder)
- 'template': Use a predefined template (e.g., 'meeting', 'task')
- 'content': Initial markdown body
- 'data': Additional frontmatter fields (e.g., {"tag": "meeting", "priority": "high"})

Returns the created note path.
```

#### Response Format

```typescript
interface CreateResponse {
  created: {
    path: string;
    absolutePath: string;
    title: string;
  };
  notebook: {
    name: string;
    path: string;
  };
}
```

---

### 5. `opennotes_notebooks`

**Purpose**: List all available notebooks.

#### Schema

```typescript
export const NotebooksParams = Type.Object({
  // No parameters - lists all notebooks
});

export type NotebooksParams = Static<typeof NotebooksParams>;
```

#### LLM Description

```
List all available OpenNotes notebooks.

Returns notebooks from:
- Global config (registered notebooks)
- Ancestor directories (discovered notebooks)

Each notebook includes name, path, and source.
```

#### Response Format

```typescript
interface NotebooksResponse {
  notebooks: NotebookInfo[];
  current: NotebookInfo | null;  // Active notebook from context, if any
}
```

---

### 6. `opennotes_views`

**Purpose**: List available views or execute a named view.

#### Schema

```typescript
export const ViewsParams = Type.Object({
  view: Type.Optional(Type.String({
    description: "View name to execute. Omit to list all views.",
  })),
  params: Type.Optional(Type.Record(Type.String(), Type.String(), {
    description: "Parameters for the view (e.g., {'status': 'todo,done'})",
  })),
  notebook: Type.Optional(Type.String({
    description: "Path to notebook. Omit to use current context.",
  })),
  limit: Type.Optional(Type.Number({
    description: "Maximum results when executing view (default: 50)",
    default: 50,
  })),
  offset: Type.Optional(Type.Number({
    description: "Offset for pagination when executing view",
    default: 0,
  })),
});

export type ViewsParams = Static<typeof ViewsParams>;
```

#### LLM Description

```
List available views or execute a named view.

**List Mode** (no 'view' parameter):
Returns all available views with descriptions and parameters.

**Execute Mode** (with 'view' parameter):
Executes the named view and returns results.

Built-in views: today, recent, kanban, untagged, orphans, broken-links

Example: { view: 'kanban', params: { status: 'todo,in-progress,done' } }
```

#### Response Format (List Mode)

```typescript
interface ViewsListResponse {
  views: ViewDefinition[];
  notebook: {
    name: string;
    path: string;
  };
}
```

#### Response Format (Execute Mode)

```typescript
interface ViewsExecuteResponse {
  view: {
    name: string;
    description: string;
  };
  results: Record<string, unknown>[];  // View-specific result shape
  pagination: PaginationMeta;
  notebook: {
    name: string;
    path: string;
  };
}
```

---

## Error Response Format

All tools use consistent error format:

```typescript
interface ErrorResponse {
  error: true;
  message: string;
  code: string;
  hint?: string;
  details?: Record<string, unknown>;
}
```

### Error Codes

| Code | Message | Hint |
|------|---------|------|
| `CLI_NOT_FOUND` | OpenNotes CLI not found | Install: `go install github.com/zenobi-us/opennotes@latest` |
| `NOTEBOOK_NOT_FOUND` | No notebook found in context | Specify --notebook or cd to a notebook directory |
| `NOTE_NOT_FOUND` | Note not found: {path} | Check path is relative to notebook root |
| `VIEW_NOT_FOUND` | View not found: {name} | Use opennotes_views to list available views |
| `TEMPLATE_NOT_FOUND` | Template not found: {name} | Check notebook .opennotes.json for templates |
| `INVALID_SQL` | Invalid SQL query: {error} | Only SELECT/WITH allowed, see docs |
| `QUERY_TIMEOUT` | Query timed out after 30s | Simplify query or add LIMIT |
| `PATH_TRAVERSAL` | Path traversal blocked: {path} | Use paths relative to notebook |

### Example Error

```json
{
  "error": true,
  "message": "OpenNotes CLI not found in PATH",
  "code": "CLI_NOT_FOUND",
  "hint": "Install OpenNotes: go install github.com/zenobi-us/opennotes@latest\nOr ensure the 'opennotes' binary is in your PATH.",
  "details": {
    "searchedPaths": ["/usr/local/bin", "/usr/bin", "/home/user/.local/bin"]
  }
}
```

---

## Pagination Example

**Request**: Search for notes tagged "meeting" (assume 127 results)

```typescript
{ query: "meeting", limit: 50 }
```

**Response** (truncated to fit 75% budget):

```json
{
  "results": [
    // ... 50 notes ...
  ],
  "pagination": {
    "total": 127,
    "returned": 50,
    "page": 1,
    "pageSize": 50,
    "hasMore": true,
    "nextOffset": 50
  },
  "query": {
    "type": "text",
    "executed": "SELECT ... WHERE content LIKE '%meeting%' LIMIT 50 OFFSET 0"
  },
  "_guidance": "More results available. To fetch next page: { offset: 50 }"
}
```

**Follow-up Request**:

```typescript
{ query: "meeting", limit: 50, offset: 50 }
```

---

## Expected Outcome

✅ Complete API specification with TypeBox schemas ready for implementation.

## Actual Outcome

Comprehensive tool specifications created with:
- Full TypeBox schemas for all 6 tools
- LLM-friendly descriptions
- Response format definitions
- Pagination strategy (75% content + metadata)
- Consistent error handling format
- Example responses for all tools

## Lessons Learned

1. `StringEnum` required for Google API compatibility (not `Type.Union`)
2. Views tool needs dual-mode (list vs execute) based on presence of `view` param
3. Pagination metadata should include `nextOffset` for easy continuation
4. Error responses need both code (for programmatic handling) and hint (for user guidance)
