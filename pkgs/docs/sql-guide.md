# SQL Query Guide for OpenNotes

This guide shows you how to use the `--sql` flag to run custom SQL queries against your notebook files using DuckDB's powerful markdown extension.

## Table of Contents

1. [Getting Started](#getting-started)
2. [File Pattern Resolution](#file-pattern-resolution)
3. [Available Functions](#available-functions)
4. [Schema Overview](#schema-overview)
5. [Common Query Patterns](#common-query-patterns)
6. [Troubleshooting](#troubleshooting)
7. [Security Model](#security-model)
8. [Performance Tips](#performance-tips)

## Related Guides

- **[JSON SQL Query Guide](json-sql-guide.md)** - Comprehensive guide for working with JSON output, automation, and external tool integration

## Getting Started

### Basic Syntax

Use the `--sql` flag with the search command:

```bash
opennotes search --sql "SELECT * FROM read_markdown('**/*.md') LIMIT 5"
```

### Your First Query

List all notes in your notebook:

```bash
opennotes search --sql "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true)"
```

**Output format:**

OpenNotes returns SQL query results in **JSON format** by default:

```json
[
  {
    "file_path": "/path/to/notebook/notes/project-ideas.md",
    "content": "# Project Ideas\n\nSome ideas for new..."
  },
  {
    "file_path": "/path/to/notebook/notes/meeting-notes.md", 
    "content": "# Meeting Notes\n\nDiscussed the new..."
  },
  {
    "file_path": "/path/to/notebook/notes/todo.md",
    "content": "# Todo List\n\n- [ ] Finish report..."
  }
]
```

**JSON Benefits:**
- Perfect for automation and scripting
- Easy integration with tools like `jq`
- Structured data for complex processing
- Standard format across all environments

**📖 For comprehensive examples, automation patterns, and tool integration, see the [JSON SQL Query Guide](json-sql-guide.md).**

## File Pattern Resolution

**IMPORTANT**: All file patterns in SQL queries are automatically resolved relative to the notebook root directory, regardless of your current working directory.

### Pattern Types Supported

| Pattern | Description | Example Matches |
|---------|-------------|-----------------|
| `*.md` | All markdown files in notebook root | `README.md`, `notes.md` |
| `**/*.md` | All markdown files recursively in entire notebook | `docs/guide.md`, `projects/todo.md`, `archive/2023/notes.md` |
| `subfolder/*.md` | All markdown files in specific subfolder | `projects/project1.md`, `projects/status.md` |
| `**/subfolder/*.md` | All markdown files in any subfolder named 'subfolder' | `work/projects/todo.md`, `personal/projects/ideas.md` |

### Resolution Behavior

**Consistent Behavior**: Patterns resolve identically regardless of current working directory:

```bash
# These commands produce identical results regardless of current directory
cd ~/notebook && opennotes notes search --sql "SELECT * FROM read_markdown('**/*.md')"
cd ~/notebook/projects && opennotes notes search --sql "SELECT * FROM read_markdown('**/*.md')"
cd /tmp && opennotes notes search --sql "SELECT * FROM read_markdown('**/*.md')" --notebook ~/notebook
```

**Security Protection**: Path traversal attempts are automatically blocked:

```bash
# ❌ These patterns are blocked and logged as security violations
opennotes notes search --sql "SELECT * FROM read_markdown('../secret/*.md')"
opennotes notes search --sql "SELECT * FROM read_markdown('/etc/passwd')"

# ✅ These patterns work correctly  
opennotes notes search --sql "SELECT * FROM read_markdown('**/*.md')"
opennotes notes search --sql "SELECT * FROM read_markdown('projects/*.md')"
```

### Best Practices for Patterns

**Use Forward Slashes**: For cross-platform compatibility, always use forward slashes:

```sql
-- ✅ Cross-platform compatible
SELECT * FROM read_markdown('projects/work/*.md')

-- ❌ Windows-specific (avoid)
SELECT * FROM read_markdown('projects\work\*.md')
```

**Be Specific When Possible**: More specific patterns improve performance:

```sql
-- ✅ Fast: searches only project folder
SELECT * FROM read_markdown('projects/*.md')

-- ⚠️ Slower: searches entire notebook recursively
SELECT * FROM read_markdown('**/*.md')
```

**Test Patterns**: Use LIMIT during development to test patterns quickly:

```sql
-- Test pattern before running full query
SELECT file_path FROM read_markdown('new-pattern/*.md', include_filepath:=true) LIMIT 5
```

## Available Functions

### Table Functions

#### `read_markdown(glob, include_filepath:=boolean)`
Reads markdown files matching the glob pattern.

**Parameters:**
- `glob` (string): File pattern (e.g., `**/*.md`, `notes/*.md`)
- `include_filepath` (boolean, optional): Include filepath column

**Returns:** Table with columns:
- `content` (string): Full markdown content
- `filepath` (string): Full file path (if include_filepath=true)
- `metadata` (map): Frontmatter as key-value pairs

**Examples:**
```sql
-- All notes with file paths
SELECT * FROM read_markdown('**/*.md', include_filepath:=true)

-- Notes in specific directory
SELECT * FROM read_markdown('projects/*.md')

-- Single file
SELECT * FROM read_markdown('README.md')
```

### Scalar Functions

#### `md_stats(content)`
Returns statistics about markdown content.

**Returns:** Struct with:
- `word_count` (integer): Number of words
- `character_count` (integer): Number of characters
- `line_count` (integer): Number of lines

**Example:**
```sql
SELECT 
    file_path,
    (md_stats(content)).word_count as words,
    (md_stats(content)).line_count as lines
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE (md_stats(content)).word_count < 100
```

#### `md_extract_links(content)`
Extracts all markdown links from content.

**Returns:** Array of structs with:
- `text` (string): Link text
- `url` (string): Link URL

**Example:**
```sql
SELECT 
    file_path,
    UNNEST(md_extract_links(content)) as link_info
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE array_length(md_extract_links(content)) > 0
```

#### `md_extract_code_blocks(content)`
Extracts code blocks from content.

**Returns:** Array of structs with:
- `language` (string): Programming language
- `code` (string): Code content

**Example:**
```sql
SELECT 
    file_path,
    cb.language,
    cb.code
FROM read_markdown('**/*.md', include_filepath:=true),
     LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
WHERE cb.language = 'python'
```

## Schema Overview

### `read_markdown()` Columns

| Column | Type | Description |
|--------|------|-------------|
| `content` | string | Full markdown content including frontmatter |
| `file_path` | string | Relative file path from notebook root (if include_filepath=true) |
| `metadata` | map | Frontmatter parsed as key-value pairs |

### Frontmatter Access

Access frontmatter fields using map syntax:

```sql
SELECT 
    file_path,
    metadata['title'] as title,
    metadata['tags'] as tags,
    metadata['date'] as date
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE metadata['title'] IS NOT NULL
```

## Common Query Patterns

### 1. Find Notes by Content

```sql
-- Case-insensitive search
SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)
WHERE LOWER(content) LIKE '%meeting%'

-- Multiple keywords (AND)
SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)
WHERE content LIKE '%project%' AND content LIKE '%deadline%'

-- Multiple keywords (OR)
SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)
WHERE content LIKE '%todo%' OR content LIKE '%task%'
```

### 2. Find Notes by Metadata

```sql
-- Notes with specific tag
SELECT file_path, metadata['title'] as title
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE metadata['tags'] LIKE '%work%'

-- Recent notes
SELECT file_path, metadata['date'] as date
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE metadata['date'] >= '2024-01-01'
ORDER BY metadata['date'] DESC

-- Notes by author
SELECT file_path, metadata['author'] as author
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE metadata['author'] = 'John Doe'
```

### 3. Analyze Content Statistics

```sql
-- Word count analysis
SELECT 
    file_path,
    (md_stats(content)).word_count as words
FROM read_markdown('**/*.md', include_filepath:=true)
ORDER BY words DESC
LIMIT 10

-- Average words per note
SELECT 
    COUNT(*) as note_count,
    AVG((md_stats(content)).word_count) as avg_words,
    MAX((md_stats(content)).word_count) as max_words
FROM read_markdown('**/*.md')

-- Find short notes that need more content
SELECT file_path, (md_stats(content)).word_count as words
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE (md_stats(content)).word_count < 50
ORDER BY words ASC
```

### 4. Code Block Analysis

```sql
-- Find all Python code
SELECT 
    file_path,
    cb.code
FROM read_markdown('**/*.md', include_filepath:=true),
     LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
WHERE cb.language = 'python'

-- Count code blocks by language
SELECT 
    cb.language,
    COUNT(*) as block_count
FROM read_markdown('**/*.md'),
     LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
GROUP BY cb.language
ORDER BY block_count DESC

-- Find notes with most code blocks
SELECT 
    file_path,
    array_length(md_extract_code_blocks(content)) as code_blocks
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE array_length(md_extract_code_blocks(content)) > 0
ORDER BY code_blocks DESC
```

### 5. Link Analysis

```sql
-- Find all external links
SELECT 
    file_path,
    link.text,
    link.url
FROM read_markdown('**/*.md', include_filepath:=true),
     LATERAL UNNEST(md_extract_links(content)) AS link
WHERE link.url LIKE 'http%'

-- Find broken internal links (simple check)
SELECT 
    file_path,
    link.url as potentially_broken
FROM read_markdown('**/*.md', include_filepath:=true),
     LATERAL UNNEST(md_extract_links(content)) AS link
WHERE link.url LIKE './%' OR link.url LIKE '../%'

-- Count links per note
SELECT 
    file_path,
    array_length(md_extract_links(content)) as link_count
FROM read_markdown('**/*.md', include_filepath:=true)
WHERE array_length(md_extract_links(content)) > 5
ORDER BY link_count DESC
```

### 6. Complex Analysis with CTEs

```sql
-- Most active writing days
WITH daily_stats AS (
    SELECT 
        metadata['date'] as date,
        COUNT(*) as notes_written,
        SUM((md_stats(content)).word_count) as words_written
    FROM read_markdown('**/*.md')
    WHERE metadata['date'] IS NOT NULL
    GROUP BY metadata['date']
)
SELECT * FROM daily_stats
WHERE words_written > 1000
ORDER BY date DESC

-- Tag analysis
WITH tag_stats AS (
    SELECT 
        UNNEST(string_split(metadata['tags'], ',')) as tag,
        file_path
    FROM read_markdown('**/*.md', include_filepath:=true)
    WHERE metadata['tags'] IS NOT NULL
)
SELECT 
    TRIM(tag) as tag,
    COUNT(*) as usage_count
FROM tag_stats
GROUP BY TRIM(tag)
ORDER BY usage_count DESC
```

## Troubleshooting

### Common Errors

#### "Only SELECT queries are allowed"
**Cause:** Trying to use INSERT, UPDATE, DELETE, or other write operations.
**Solution:** Use only SELECT or WITH statements.

```bash
# ❌ This fails
opennotes search --sql "DELETE FROM markdown"

# ✅ This works
opennotes search --sql "SELECT * FROM read_markdown('**/*.md')"
```

#### "File or directory does not exist"
**Cause:** No files match your glob pattern.
**Solution:** Check your file pattern and notebook path.

```bash
# ❌ No .md files found
opennotes search --sql "SELECT * FROM read_markdown('*.txt')"

# ✅ Correct pattern
opennotes search --sql "SELECT * FROM read_markdown('**/*.md')"
```

#### "path traversal detected: query would access files outside notebook"
**Cause:** Query contains `../` or attempts to access files outside the notebook directory.
**Solution:** Use relative paths from notebook root, remove `../` patterns.

```bash
# ❌ Path traversal attempt (blocked)
opennotes search --sql "SELECT * FROM read_markdown('../other-folder/*.md')"
opennotes search --sql "SELECT * FROM read_markdown('/home/user/secret/*.md')"

# ✅ Correct patterns from notebook root
opennotes search --sql "SELECT * FROM read_markdown('other-folder/*.md')"
opennotes search --sql "SELECT * FROM read_markdown('**/*.md')"
```

#### "query preprocessing failed: malformed pattern"
**Cause:** Invalid glob pattern syntax or quote mismatch.
**Solution:** Check quote matching and pattern format.

```bash
# ❌ Mismatched quotes
opennotes search --sql "SELECT * FROM read_markdown('**/*.md\")"

# ❌ Unclosed quotes  
opennotes search --sql "SELECT * FROM read_markdown('**/*.md"

# ✅ Proper quotes
opennotes search --sql "SELECT * FROM read_markdown('**/*.md')"
```

#### "keyword 'DROP' is not allowed"
**Cause:** Using blocked dangerous keywords.
**Solution:** Remove dangerous keywords from your query.

#### Query times out after 30 seconds
**Cause:** Query is too complex or dataset too large.
**Solution:** Add LIMIT clauses or simplify the query.

```bash
# ❌ May timeout on large notebooks
opennotes search --sql "SELECT * FROM read_markdown('**/*.md')"

# ✅ Limited results
opennotes search --sql "SELECT * FROM read_markdown('**/*.md') LIMIT 100"
```

### Debug Tips

1. **Start simple:** Begin with `SELECT * FROM read_markdown('**/*.md') LIMIT 5`
2. **Check your pattern:** Use specific glob patterns to limit files
3. **Use LIMIT:** Always limit results during testing
4. **Test patterns:** Use `include_filepath:=true` to see which files are being accessed
5. **Verify notebook path:** Run `opennotes notebook info` to confirm notebook location
6. **Check current directory:** Pattern resolution is independent of current directory

### Pattern Resolution Debugging

**Verify notebook root**: Ensure you understand your notebook root directory:

```bash
# Check notebook configuration
opennotes notebook info

# Test simple pattern first
opennotes search --sql "SELECT file_path FROM read_markdown('*.md', include_filepath:=true) LIMIT 3"

# Then expand to recursive
opennotes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 10"
```

**Test incremental patterns**: Build complexity gradually:

```bash
# 1. Test root files only
opennotes search --sql "SELECT COUNT(*) FROM read_markdown('*.md')"

# 2. Test specific subfolder  
opennotes search --sql "SELECT COUNT(*) FROM read_markdown('projects/*.md')"

# 3. Test recursive pattern
opennotes search --sql "SELECT COUNT(*) FROM read_markdown('**/*.md')"
```

## Security Model

### Read-Only Access
- Only SELECT and WITH (Common Table Expression) queries allowed
- No data modification possible (INSERT, UPDATE, DELETE blocked)
- No schema changes possible (CREATE, ALTER, DROP blocked)
- No system access (PRAGMA, ATTACH blocked)

### Pattern Resolution Security

**Path Traversal Protection**: All file patterns are automatically validated to prevent access outside the notebook directory:

- Path traversal attempts using `../` are automatically blocked
- Absolute paths (starting with `/` or drive letters) are blocked
- Only files within the notebook tree are accessible
- Security violations are logged for monitoring

**Examples of blocked patterns**:
```sql
-- ❌ These are automatically blocked
SELECT * FROM read_markdown('../secrets/*.md')           -- Path traversal
SELECT * FROM read_markdown('/etc/passwd')              -- Absolute path
SELECT * FROM read_markdown('C:\Windows\system32\*')    -- Windows absolute path
SELECT * FROM read_markdown('../../other-notebook/*')   -- Multi-level traversal
```

**Safe patterns that work correctly**:
```sql
-- ✅ These patterns are safe and work as intended
SELECT * FROM read_markdown('**/*.md')                  -- All notebook files
SELECT * FROM read_markdown('projects/*.md')            -- Subfolder files
SELECT * FROM read_markdown('*.md')                     -- Root-level files
SELECT * FROM read_markdown('docs/archive/*.md')        -- Nested subfolder
```

### Validation
Queries are validated before execution to block:
- `INSERT`, `UPDATE`, `DELETE`
- `DROP`, `CREATE`, `ALTER`
- `TRUNCATE`, `REPLACE`
- `ATTACH`, `DETACH`
- `PRAGMA`

### Pattern Processing Security

**Automatic Resolution**: File patterns are processed before query execution to:
- Convert relative patterns to absolute paths from notebook root
- Validate all resolved paths stay within notebook boundaries  
- Log security violations for monitoring and audit trails
- Ensure consistent behavior regardless of current working directory

**Security Logging**: All security violations are logged with details including:
- Original query pattern
- Attempted resolved path
- User context and timestamp
- Specific security rule violated

### Timeout Protection
- All queries have a 30-second timeout
- Prevents runaway queries from blocking the system
- Long-running queries are automatically cancelled

### Isolation
- Separate read-only database connection
- No access to OpenNotes internal tables
- Cannot affect notebook files on disk
- Query processing runs in isolated context

## Performance Tips

### Use Specific Glob Patterns
```sql
-- ❌ Slow: searches all files (may include non-markdown)
SELECT * FROM read_markdown('**/*')

-- ✅ Fast: searches only markdown files
SELECT * FROM read_markdown('**/*.md')

-- ✅ Faster: searches specific directory
SELECT * FROM read_markdown('work-notes/*.md')

-- ✅ Fastest: searches specific file
SELECT * FROM read_markdown('README.md')
```

### Pattern Specificity Impact
More specific patterns significantly improve performance by reducing the number of files processed:

| Pattern | Performance Impact | Use Case |
|---------|-------------------|----------|
| `'file.md'` | Fastest | Single known file |
| `'folder/*.md'` | Fast | Specific folder |
| `'**/*.md'` | Moderate | All notebook files |
| `'**/*'` | Slowest | All files (avoid) |

### Limit Results
```sql
-- ❌ Returns all results (potentially thousands)
SELECT * FROM read_markdown('**/*.md')

-- ✅ Returns manageable number
SELECT * FROM read_markdown('**/*.md') LIMIT 50

-- ✅ Pagination for large datasets
SELECT * FROM read_markdown('**/*.md') LIMIT 50 OFFSET 100
```

### Filter Early
```sql
-- ❌ Processes all notes then filters
SELECT * FROM (
    SELECT *, (md_stats(content)).word_count as words
    FROM read_markdown('**/*.md')
) WHERE words > 1000

-- ✅ Filters during processing
SELECT *, (md_stats(content)).word_count as words
FROM read_markdown('**/*.md')
WHERE (md_stats(content)).word_count > 1000

-- ✅ Even better: filter by pattern first
SELECT *, (md_stats(content)).word_count as words
FROM read_markdown('articles/*.md')  -- Specific folder
WHERE (md_stats(content)).word_count > 1000
```

### Pattern Resolution Performance

**One-time Cost**: Pattern resolution happens once per query before execution:
- Pattern processing adds <1ms overhead per pattern
- Resolved paths are cached for query duration  
- Security validation is lightweight
- Overall impact negligible for typical queries

**Optimization Strategy**: Structure your notebook for efficient patterns:
```
notebook/
├── daily/           # Daily notes
├── projects/        # Project documentation  
├── archive/         # Old content
├── templates/       # Note templates
└── reference/       # Reference materials
```

**Efficient queries for this structure**:
```sql
-- Target specific areas
SELECT * FROM read_markdown('projects/*.md') WHERE content LIKE '%urgent%'
SELECT * FROM read_markdown('daily/2024-*.md') ORDER BY filepath DESC LIMIT 7
SELECT * FROM read_markdown('reference/**/*.md') WHERE metadata['type'] = 'guide'
```

### Use Appropriate Indexes
DuckDB automatically optimizes many queries, but you can help by:
- Filtering on metadata fields early
- Using specific patterns instead of broad searches
- Limiting result sets
- Avoiding complex string operations on large content

### Query Optimization Examples

**Content Search Optimization**:
```sql
-- ❌ Inefficient: searches all content
SELECT * FROM read_markdown('**/*.md') 
WHERE content LIKE '%search_term%'

-- ✅ More efficient: specific folder + limit
SELECT * FROM read_markdown('projects/*.md') 
WHERE content LIKE '%search_term%' 
LIMIT 20

-- ✅ Most efficient: metadata first, then content
SELECT * FROM read_markdown('projects/*.md') 
WHERE metadata['tags'] LIKE '%urgent%'
  AND content LIKE '%search_term%'
LIMIT 10
```

**Statistical Analysis Optimization**:
```sql
-- ❌ Processes all files for simple count
SELECT COUNT(*) FROM read_markdown('**/*.md')

-- ✅ Use filesystem patterns when possible
SELECT COUNT(*) FROM read_markdown('projects/*.md')
```

---

## More Resources

- [DuckDB SQL Reference](https://duckdb.org/docs/sql/introduction)
- [DuckDB Markdown Extension](https://github.com/duckdb/duckdb/blob/main/extension/markdown/README.md)
- [OpenNotes CLI Reference](./cli-reference.md)

Need help? Check the troubleshooting section above or open an issue on GitHub.