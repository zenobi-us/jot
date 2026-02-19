# Notes Search Command

Search notes with text matching, fuzzy ranking, boolean filters, DSL pipe syntax, or semantic retrieval.

## Overview

Jot supports these search workflows:

1. **Text Search** (`jot notes search "..."`)  
   Exact substring matching in content and file paths.
2. **Fuzzy Search** (`--fuzzy`)  
   Typo-tolerant ranking based on character similarity.
3. **Boolean Query Search** (`jot notes search query`)  
   Structured `--and/--or/--not` filtering on metadata, path, title, and links.
4. **DSL Pipe Syntax** (`jot notes search "filter | directives"`)  
   Filter expression + sort/limit/offset directives.
5. **Semantic Search** (`jot notes search semantic`)  
   Keyword / semantic / hybrid retrieval with optional explain output.

---

## Quick Start

```bash
# Text search
jot notes search "meeting"

# Fuzzy search
jot notes search --fuzzy "mtng"

# Boolean query
jot notes search query --and data.tag=workflow --not data.status=archived

# DSL + directives
jot notes search "tag:work | sort:modified:desc limit:10"

# Semantic hybrid search
jot notes search semantic "project planning discussions"
```

---

## Text Search

### Syntax

```bash
jot notes search [query]
```

### Examples

```bash
jot notes search "meeting"
jot notes search "todo" --notebook ~/notes
jot notes search
```

### Behavior

- Case-insensitive matching
- Searches note content and file path
- No query returns all notes

---

## Fuzzy Search

### Syntax

```bash
jot notes search [query] --fuzzy
```

### Examples

```bash
jot notes search --fuzzy "mtng"
jot notes search "project" --fuzzy
jot notes search --fuzzy
```

### How It Ranks

- Character-sequence fuzzy matching (VS Code-like feel)
- Title matches weighted higher than body matches
- Results sorted by best score first
- Body scan optimized (first part of note content)

---

## Boolean Query Search (`query` subcommand)

### Syntax

```bash
jot notes search query [--and field=value] [--or field=value] [--not field=value]
```

### Operators

| Operator | Meaning                   |
| -------- | ------------------------- |
| `--and`  | all conditions must match |
| `--or`   | any condition can match   |
| `--not`  | excludes matching notes   |

### Supported Fields

#### Metadata (`data.*`)

- `data.tag`, `data.tags`
- `data.status`, `data.priority`
- `data.assignee`, `data.author`
- `data.type`, `data.category`
- `data.project`, `data.sprint`

#### File/Title

- `path` (glob-enabled)
- `title`

#### Link Graph

- `links-to`
- `linked-by`

### Examples

```bash
# Basic metadata filtering
jot notes search query --and data.tag=workflow --and data.status=active

# OR logic
jot notes search query --or data.priority=high --or data.priority=critical

# Path filtering
jot notes search query --and path=projects/**/*.md --not path=archive/*

# Link graph filtering
jot notes search query --and links-to=docs/architecture.md
jot notes search query --and linked-by=planning/q1.md
```

---

## DSL Pipe Syntax (`search [query]`)

Use this when you want inline sorting/pagination directives.

### Syntax

```bash
jot notes search "<filter> | <directives>"
```

### Filter Examples

```bash
jot notes search "tag:work | sort:modified:desc"
jot notes search "status:todo | sort:created:asc limit:20"
jot notes search "| sort:title:asc"
```

### Filter Fields (DSL side)

- `tag:<value>`
- `status:<value>`
- `title:<text>`
- `path:<glob-or-prefix>`
- `created:>date`, `created:<date`
- `modified:>date`, `modified:<date`

### Directives

- `sort:<field>:<dir>` where field is `modified|created|title|path`, dir is `asc|desc`
- `limit:<n>`
- `offset:<n>`

---

## Semantic Search (`semantic` subcommand)

### Syntax

```bash
jot notes search semantic [query] [flags]
```

### Key Flags

- `--mode hybrid|keyword|semantic` (default: `hybrid`)
- `--top-k <n>` (default: `100`)
- `--explain`
- `--and`, `--or`, `--not` (same condition format as `query`)

### Examples

```bash
jot notes search semantic "meeting notes"
jot notes search semantic "workflow" --mode keyword --and data.status=active
jot notes search semantic "architecture" --mode hybrid --explain
```

For full semantic behavior and troubleshooting, see:

- [Semantic Search Guide](../semantic-search-guide.md)

---

## Glob Patterns

`path`, `links-to`, and `linked-by` support globs.

| Pattern | Meaning                  | Example     |
| ------- | ------------------------ | ----------- |
| `*`     | any chars (single level) | `docs/*.md` |
| `**`    | recursive path depth     | `**/*.md`   |
| `?`     | single character         | `task?.md`  |

---

## Common Errors

| Error                                | Cause                             | Fix                               |
| ------------------------------------ | --------------------------------- | --------------------------------- |
| `invalid field: X`                   | Unsupported field name            | Use supported fields listed above |
| `expected field=value`               | Missing `=`                       | Use `field=value` format          |
| `value too long`                     | Condition value exceeds limit     | Shorten value                     |
| `at least one condition is required` | `query` called with no conditions | Add `--and`, `--or`, or `--not`   |

---

## Related Docs

- [Semantic Search Guide](../semantic-search-guide.md)
- [Getting Started (Power Users)](../getting-started-power-users.md)
- [Notebook Discovery](../notebook-discovery.md)
