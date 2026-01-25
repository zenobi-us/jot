# Notes Search Command

Search notes using text search, fuzzy matching, boolean queries, or SQL.

## Overview

OpenNotes provides multiple search methods to find notes:

1. **Text Search**: Exact substring matching in content and filenames
2. **Fuzzy Search**: Similarity-based matching with ranked results
3. **Boolean Queries**: Structured queries with AND/OR/NOT operators
4. **SQL Queries**: Full DuckDB SQL for power users

## Quick Start

```bash
# Simple text search
opennotes notes search "meeting"

# Fuzzy search (typo-tolerant, ranked results)
opennotes notes search --fuzzy "mtng"

# Boolean query (metadata filtering)
opennotes notes search query --and data.tag=workflow

# SQL query (full power)
opennotes notes search --sql "SELECT * FROM read_markdown('**/*.md') LIMIT 10"
```

---

## Text Search

Basic substring search in note content and filenames.

### Syntax

```bash
opennotes notes search [text]
```

### Examples

```bash
# Search for "meeting" in all notes
opennotes notes search "meeting"

# Search in specific notebook
opennotes notes search "todo" --notebook ~/notes

# List all notes (no search term)
opennotes notes search
```

### Behavior

- Case-insensitive matching
- Searches both file content and filepath
- Returns all matching notes

---

## Fuzzy Search

Similarity-based matching that tolerates typos and partial matches. Results are ranked by relevance.

### Syntax

```bash
opennotes notes search [text] --fuzzy
```

### Examples

```bash
# Fuzzy search for "meeting" (matches "mtng", "meeting", "meetings")
opennotes notes search --fuzzy "mtng"

# Fuzzy search with text
opennotes notes search "project" --fuzzy

# List all notes with fuzzy ranking
opennotes notes search --fuzzy
```

### How Fuzzy Matching Works

1. **Algorithm**: Uses character sequence matching (similar to VS Code's Ctrl+P)
2. **Ranking**: Results sorted by match score (best matches first)
3. **Weighting**: Title matches weighted 2x higher than body matches
4. **Performance**: Searches first 500 characters of body for efficiency

### When to Use Fuzzy Search

| Scenario                     | Use Fuzzy?         |
| ---------------------------- | ------------------ |
| Exact phrase lookup          | ❌ Use text search |
| Don't remember exact wording | ✅ Yes             |
| Searching for abbreviations  | ✅ Yes             |
| Finding similar note titles  | ✅ Yes             |

---

## Boolean Query Search

Structured queries with AND/OR/NOT operators for filtering by metadata fields.

### Syntax

```bash
opennotes notes search query [--and field=value] [--or field=value] [--not field=value]
```

### Boolean Operators

| Operator | Description               | Example                      |
| -------- | ------------------------- | ---------------------------- |
| `--and`  | All conditions must match | `--and data.tag=workflow`    |
| `--or`   | Any condition can match   | `--or data.priority=high`    |
| `--not`  | Exclude matching notes    | `--not data.status=archived` |

### Operator Precedence

1. AND conditions evaluated first (intersection)
2. OR conditions combined (union)
3. NOT conditions applied as exclusions

### Examples

```bash
# Single condition - find workflow notes
opennotes notes search query --and data.tag=workflow

# Multiple AND - all must match
opennotes notes search query --and data.tag=workflow --and data.status=active

# OR conditions - any can match
opennotes notes search query --or data.priority=high --or data.priority=critical

# Combined - find active epics excluding archived
opennotes notes search query --and data.tag=epic --not data.status=archived

# Complex query
opennotes notes search query \
  --and data.tag=workflow \
  --and data.status=active \
  --or data.priority=high \
  --not data.assignee=bob
```

---

## Supported Fields

Fields available for boolean queries:

### Metadata Fields (`data.*`)

| Field           | Description       | Example               |
| --------------- | ----------------- | --------------------- |
| `data.tag`      | Note tags         | `data.tag=workflow`   |
| `data.tags`     | Note tags (alias) | `data.tags=meeting`   |
| `data.status`   | Note status       | `data.status=active`  |
| `data.priority` | Priority level    | `data.priority=high`  |
| `data.assignee` | Assigned person   | `data.assignee=alice` |
| `data.author`   | Note author       | `data.author=bob`     |
| `data.type`     | Note type         | `data.type=epic`      |
| `data.category` | Category          | `data.category=docs`  |
| `data.project`  | Project name      | `data.project=alpha`  |
| `data.sprint`   | Sprint identifier | `data.sprint=s23`     |

### Path and Title Fields

| Field   | Description                | Example           |
| ------- | -------------------------- | ----------------- |
| `path`  | File path (supports globs) | `path=projects/*` |
| `title` | Note title                 | `title=Meeting`   |

### Link Fields (DAG Queries)

| Field       | Description              | Example               |
| ----------- | ------------------------ | --------------------- |
| `links-to`  | Notes linking TO target  | `links-to=epics/*.md` |
| `linked-by` | Notes linked FROM source | `linked-by=plan.md`   |

---

## Link Queries

Query notes based on their linking relationships (DAG foundation).

### Concepts

```
Document A --link--> Document B

links-to=B    → Returns A (who points to B?)
linked-by=A   → Returns B (what does A point to?)
```

### Examples

```bash
# Find notes that link to architecture.md
opennotes notes search query --and links-to=docs/architecture.md

# Find notes that planning.md links to
opennotes notes search query --and linked-by=planning/q1.md

# Find notes linking to any epic
opennotes notes search query --and links-to=epics/**/*.md

# Find notes linking to any task, but not archived
opennotes notes search query \
  --and links-to=tasks/**/*.md \
  --not data.status=archived
```

### Link Query Use Cases

| Use Case                      | Query                                              |
| ----------------------------- | -------------------------------------------------- |
| What references this doc?     | `--and links-to=target.md`                         |
| What does this doc reference? | `--and linked-by=source.md`                        |
| Find all epic dependencies    | `--and data.tag=epic --and links-to=tasks/**/*.md` |
| Find orphaned tasks           | `--not linked-by=epics/**/*.md`                    |

### Known Limitations

> **Note**: Link queries require the `links` field in frontmatter to be properly
> parsed as a YAML array by DuckDB's markdown extension. Currently, DuckDB's
> markdown extension has limited support for YAML arrays - complex nested
> structures may not be parsed correctly. If link queries return no results,
> verify the links field is being parsed as an array using SQL:
>
> ```bash
> opennotes notes search --sql "SELECT file_path, metadata['links'] FROM read_markdown('**/*.md', include_filepath:=true)"
> ```

---

## Glob Patterns

Both `path` and link fields support glob patterns.

### Pattern Reference

| Pattern | Meaning                       | Matches                       |
| ------- | ----------------------------- | ----------------------------- |
| `*`     | Any characters (single level) | `docs/*.md` → `docs/guide.md` |
| `**`    | Any path depth                | `**/*.md` → `a/b/c/file.md`   |
| `?`     | Single character              | `task?.md` → `task1.md`       |

### Glob Examples

```bash
# All markdown files in docs/
opennotes notes search query --and path=docs/*.md

# All markdown files in any subdirectory
opennotes notes search query --and path=**/*.md

# Files matching pattern
opennotes notes search query --and path=task-???.md

# Epics linking to any task
opennotes notes search query --and path=epics/* --and links-to=tasks/**/*.md
```

---

## SQL Queries

Full DuckDB SQL for advanced queries.

### Syntax

```bash
opennotes notes search --sql "SELECT ..."
```

### Basic Examples

```bash
# List first 10 notes
opennotes notes search --sql "SELECT * FROM read_markdown('**/*.md') LIMIT 10"

# Search content
opennotes notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%todo%'"

# Word count statistics
opennotes notes search --sql "SELECT file_path, (md_stats(content)).word_count as words FROM read_markdown('**/*.md', include_filepath:=true) WHERE (md_stats(content)).word_count > 1000"
```

### SQL Security

- Only `SELECT` and `WITH` queries allowed
- Read-only access enforced
- 30-second timeout per query
- Path traversal (`../`) blocked
- All file access restricted to notebook directory

### Resources

- [SQL Guide](../sql-guide.md)
- [SQL Functions Reference](../sql-functions-reference.md)
- [JSON/SQL Guide](../json-sql-guide.md)

---

## Performance

### Performance Targets

| Operation              | Dataset       | Target  | Actual |
| ---------------------- | ------------- | ------- | ------ |
| Text search            | 10k notes     | < 10ms  | ~1.4ms |
| Fuzzy search           | 10k notes     | < 50ms  | ~18ms  |
| Boolean query building | Any           | < 20ms  | < 1ms  |
| Complex query building | 5+ conditions | < 100ms | < 1ms  |
| Link query building    | Any           | < 50ms  | < 1ms  |

### Performance Tips

1. **Use specific paths**: `path=projects/*.md` faster than `path=**/*.md`
2. **Limit fuzzy body search**: Only first 500 chars searched
3. **Use boolean queries over SQL**: Optimized for common patterns
4. **Combine conditions**: More specific queries are faster

---

## Error Messages

### Common Errors

| Error                                 | Cause                    | Solution                            |
| ------------------------------------- | ------------------------ | ----------------------------------- |
| `invalid field: X`                    | Unknown field name       | Use supported field from list above |
| `expected field=value`                | Missing `=` in condition | Use `--and field=value` format      |
| `value too long`                      | Value exceeds 1000 chars | Shorten the search value            |
| `value cannot be empty`               | Empty value after `=`    | Provide a value: `field=value`      |
| `linked-by requires notebook context` | Missing notebook         | Specify `--notebook` flag           |

### Security Validation

All queries are validated for security:

1. **Field whitelist**: Only allowed fields can be queried
2. **Value length limit**: Max 1000 characters
3. **Parameterized SQL**: No SQL injection possible
4. **Path restrictions**: Can't access outside notebook

---

## Examples Cookbook

### Find Active Work

```bash
# All active tasks
opennotes notes search query --and data.tag=task --and data.status=active

# High priority items
opennotes notes search query --or data.priority=high --or data.priority=critical

# My assignments
opennotes notes search query --and data.assignee=myname --not data.status=done
```

### Explore Relationships

```bash
# What depends on architecture doc?
opennotes notes search query --and links-to=docs/architecture.md

# What does the Q1 plan reference?
opennotes notes search query --and linked-by=planning/q1.md

# Epics with task dependencies
opennotes notes search query --and data.tag=epic --and links-to=tasks/**/*.md
```

### Project Organization

```bash
# All notes in projects folder
opennotes notes search query --and path=projects/**/*.md

# Project Alpha notes
opennotes notes search query --and data.project=alpha

# Specs not yet implemented
opennotes notes search query --and data.type=spec --not data.status=implemented
```

### Content Discovery

````bash
# Fuzzy find meeting notes
opennotes notes search --fuzzy "mtng"

# Find code-related notes
opennotes notes search "```python"

# SQL: Notes with images
opennotes notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%![%](%)%'"
````
