# SQL Functions Reference

Quick reference for DuckDB markdown extension functions available in OpenNotes.

## File Pattern Resolution

**IMPORTANT**: All file patterns are resolved relative to the notebook root directory, providing consistent behavior regardless of current working directory.

### Pattern Security & Resolution
- Patterns always resolve from notebook root
- Path traversal attempts (`../`) are blocked  
- Security violations are logged
- Cross-platform compatibility with forward slashes

### Supported Pattern Examples
```sql
-- Root-level files only
read_markdown('*.md')

-- All files recursively  
read_markdown('**/*.md')

-- Specific subfolder
read_markdown('projects/*.md')

-- Nested subfolder files
read_markdown('docs/guides/*.md')

-- Any subfolder with specific name
read_markdown('**/templates/*.md')
```

## Table Functions

### `read_markdown(glob, ...)`

Reads markdown files matching a glob pattern resolved from notebook root.

**Syntax:**
```sql
read_markdown(glob_pattern, include_filepath := true)
```

**Parameters:**
- `glob_pattern` (string): File pattern resolved from notebook root (e.g., `**/*.md`, `notes/*.md`)
- `include_filepath` (boolean, optional): Include filepath column (default: false)

**Pattern Resolution:**
- All patterns resolve from notebook root directory (not current working directory)
- Security validation prevents access outside notebook boundaries
- Path traversal attempts (`../`) are automatically blocked and logged
- Patterns use forward slashes for cross-platform compatibility

**Returns:**
- `content` (string): Full markdown content including frontmatter
- `file_path` (string): Relative path from notebook root (if include_filepath=true)  
- `metadata` (map): Frontmatter parsed as key-value pairs

**Examples:**
```sql
-- Basic usage - all markdown files in notebook
SELECT * FROM read_markdown('**/*.md')

-- With file paths - see which files are included
SELECT * FROM read_markdown('**/*.md', include_filepath := true)

-- Specific directory from notebook root
SELECT * FROM read_markdown('projects/*.md', include_filepath := true)

-- Root-level files only
SELECT * FROM read_markdown('*.md', include_filepath := true)

-- Nested subfolder pattern
SELECT * FROM read_markdown('docs/guides/*.md')
```

**Security Notes:**
```sql
-- ✅ These patterns work correctly
SELECT * FROM read_markdown('**/*.md')           -- All notebook files
SELECT * FROM read_markdown('projects/*.md')     -- Subfolder files  
SELECT * FROM read_markdown('docs/2024/*.md')    -- Nested subfolder

-- ❌ These patterns are blocked for security
SELECT * FROM read_markdown('../other/*.md')     -- Path traversal blocked
SELECT * FROM read_markdown('/etc/passwd')       -- Absolute path blocked
```

## Scalar Functions

### `md_stats(content)`

Returns statistics about markdown content.

**Syntax:**
```sql
md_stats(content_string)
```

**Parameters:**
- `content_string` (string): Markdown content to analyze

**Returns:** Struct with:
- `word_count` (integer): Number of words
- `character_count` (integer): Number of characters  
- `line_count` (integer): Number of lines

**Examples:**
```sql
-- Get word count
SELECT (md_stats(content)).word_count as words
FROM read_markdown('**/*.md')

-- Get all stats
SELECT 
    (md_stats(content)).word_count as words,
    (md_stats(content)).character_count as chars,
    (md_stats(content)).line_count as lines
FROM read_markdown('**/*.md')

-- Filter by word count
SELECT * FROM read_markdown('**/*.md', include_filepath := true)
WHERE (md_stats(content)).word_count > 500
```

### `md_extract_links(content)`

Extracts all markdown links from content.

**Syntax:**
```sql
md_extract_links(content_string)
```

**Parameters:**
- `content_string` (string): Markdown content to analyze

**Returns:** Array of structs with:
- `text` (string): Link text/title
- `url` (string): Link URL/destination

**Examples:**
```sql
-- Extract all links
SELECT UNNEST(md_extract_links(content)) as link_info
FROM read_markdown('**/*.md')

-- Get link text and URLs separately  
SELECT 
    file_path,
    link.text,
    link.url
FROM read_markdown('**/*.md', include_filepath := true),
     LATERAL UNNEST(md_extract_links(content)) AS link

-- Count links per file
SELECT 
    file_path,
    array_length(md_extract_links(content)) as link_count
FROM read_markdown('**/*.md', include_filepath := true)
```

### `md_extract_code_blocks(content)`

Extracts code blocks from markdown content.

**Syntax:**
```sql
md_extract_code_blocks(content_string)
```

**Parameters:**
- `content_string` (string): Markdown content to analyze

**Returns:** Array of structs with:
- `language` (string): Programming language (if specified)
- `code` (string): Code content

**Examples:**
```sql
-- Extract all code blocks
SELECT UNNEST(md_extract_code_blocks(content)) as code_info  
FROM read_markdown('**/*.md')

-- Get Python code only
SELECT 
    file_path,
    cb.code
FROM read_markdown('**/*.md', include_filepath := true),
     LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
WHERE cb.language = 'python'

-- Count code blocks by language
SELECT 
    cb.language,
    COUNT(*) as count
FROM read_markdown('**/*.md'),
     LATERAL UNNEST(md_extract_code_blocks(content)) AS cb
GROUP BY cb.language
ORDER BY count DESC
```

### `md_extract_headers(content)`

Extracts headers from markdown content.

**Syntax:**
```sql
md_extract_headers(content_string)
```

**Parameters:**
- `content_string` (string): Markdown content to analyze

**Returns:** Array of structs with:
- `level` (integer): Header level (1-6)
- `text` (string): Header text

**Examples:**
```sql
-- Extract all headers
SELECT UNNEST(md_extract_headers(content)) as header_info
FROM read_markdown('**/*.md')

-- Get top-level headers only
SELECT 
    file_path,
    header.text
FROM read_markdown('**/*.md', include_filepath := true),
     LATERAL UNNEST(md_extract_headers(content)) AS header  
WHERE header.level = 1

-- Count headers by level
SELECT 
    header.level,
    COUNT(*) as count
FROM read_markdown('**/*.md'),
     LATERAL UNNEST(md_extract_headers(content)) AS header
GROUP BY header.level
ORDER BY header.level
```

## Standard SQL Functions

These standard SQL functions are particularly useful with markdown content:

### String Functions

#### `LIKE` and `ILIKE`
```sql
-- Case-sensitive pattern matching
SELECT * FROM read_markdown('**/*.md') WHERE content LIKE '%TODO%'

-- Case-insensitive pattern matching  
SELECT * FROM read_markdown('**/*.md') WHERE content ILIKE '%todo%'
```

#### `LOWER()` and `UPPER()`
```sql
-- Convert to lowercase for comparison
SELECT * FROM read_markdown('**/*.md') 
WHERE LOWER(content) LIKE '%meeting%'
```

#### `LENGTH()`
```sql
-- Content length
SELECT file_path, LENGTH(content) as content_length
FROM read_markdown('**/*.md', include_filepath := true)
```

#### `SUBSTRING()`
```sql
-- First 100 characters of content
SELECT file_path, SUBSTRING(content, 1, 100) as preview
FROM read_markdown('**/*.md', include_filepath := true)
```

#### `SPLIT()` and `STRING_SPLIT()`
```sql
-- Split metadata tags
SELECT 
    file_path,
    UNNEST(string_split(metadata['tags'], ',')) as tag
FROM read_markdown('**/*.md', include_filepath := true)
WHERE metadata['tags'] IS NOT NULL
```

### Array Functions

#### `array_length()`
```sql
-- Count links per file
SELECT 
    file_path,
    array_length(md_extract_links(content)) as link_count
FROM read_markdown('**/*.md', include_filepath := true)
```

#### `UNNEST()`
```sql
-- Expand arrays into rows
SELECT 
    file_path,
    UNNEST(md_extract_links(content)) as link
FROM read_markdown('**/*.md', include_filepath := true)
```

### Aggregate Functions

#### `COUNT()`, `SUM()`, `AVG()`
```sql
-- Statistics across all notes
SELECT 
    COUNT(*) as total_notes,
    AVG((md_stats(content)).word_count) as avg_words,
    SUM((md_stats(content)).word_count) as total_words
FROM read_markdown('**/*.md')
```

#### `MAX()`, `MIN()`
```sql
-- Find longest and shortest notes
SELECT 
    MAX((md_stats(content)).word_count) as longest_note,
    MIN((md_stats(content)).word_count) as shortest_note
FROM read_markdown('**/*.md')
```

### Date Functions

#### `CURRENT_DATE`, `NOW()`
```sql
-- Recent notes (if dates in metadata)
SELECT * FROM read_markdown('**/*.md', include_filepath := true)
WHERE metadata['date']::DATE >= CURRENT_DATE - INTERVAL 7 DAY
```

## Working with Metadata

Frontmatter metadata is accessible as a map. Common patterns:

### Access Fields
```sql
-- Access specific metadata fields
SELECT 
    file_path,
    metadata['title'] as title,
    metadata['author'] as author,
    metadata['tags'] as tags,
    metadata['date'] as date
FROM read_markdown('**/*.md', include_filepath := true)
```

### Check for Fields
```sql
-- Notes with titles
SELECT * FROM read_markdown('**/*.md', include_filepath := true)
WHERE metadata['title'] IS NOT NULL

-- Notes with tags
SELECT * FROM read_markdown('**/*.md', include_filepath := true)  
WHERE metadata['tags'] IS NOT NULL
```

### Type Conversion
```sql
-- Convert metadata to appropriate types
SELECT 
    file_path,
    metadata['date']::DATE as date,
    metadata['word_goal']::INTEGER as word_goal
FROM read_markdown('**/*.md', include_filepath := true)
WHERE metadata['date'] IS NOT NULL
```

## Common Patterns

### Content Search
```sql
-- Case-insensitive content search
SELECT file_path FROM read_markdown('**/*.md', include_filepath := true)
WHERE LOWER(content) LIKE '%search_term%'

-- Multiple search terms (AND)
SELECT file_path FROM read_markdown('**/*.md', include_filepath := true)
WHERE content LIKE '%term1%' AND content LIKE '%term2%'

-- Multiple search terms (OR)  
SELECT file_path FROM read_markdown('**/*.md', include_filepath := true)
WHERE content LIKE '%term1%' OR content LIKE '%term2%'
```

### Statistics and Analysis
```sql
-- Word count distribution
SELECT 
    CASE 
        WHEN (md_stats(content)).word_count < 100 THEN 'Short'
        WHEN (md_stats(content)).word_count < 500 THEN 'Medium'  
        ELSE 'Long'
    END as length_category,
    COUNT(*) as note_count
FROM read_markdown('**/*.md')
GROUP BY length_category
```

### Complex Queries with CTEs
```sql
-- Using Common Table Expressions for complex analysis
WITH note_stats AS (
    SELECT 
        file_path,
        (md_stats(content)).word_count as words,
        array_length(md_extract_links(content)) as links,
        array_length(md_extract_code_blocks(content)) as code_blocks
    FROM read_markdown('**/*.md', include_filepath := true)
)
SELECT * FROM note_stats 
WHERE words > 1000 AND (links > 5 OR code_blocks > 2)
```

## Function Reference Quick Lookup

| Function | Purpose | Returns |
|----------|---------|---------|
| `read_markdown()` | Read markdown files | Table: content, file_path, metadata |
| `md_stats()` | Content statistics | Struct: word_count, character_count, line_count |
| `md_extract_links()` | Extract links | Array: [{text, url}, ...] |
| `md_extract_code_blocks()` | Extract code | Array: [{language, code}, ...] |
| `md_extract_headers()` | Extract headers | Array: [{level, text}, ...] |

## Error Reference

| Error | Cause | Solution |
|-------|-------|----------|
| "File or directory does not exist" | No files match glob pattern | Check file pattern syntax and verify files exist |
| "Invalid glob pattern" | Malformed pattern syntax | Use proper glob syntax: `*`, `**`, `?` |
| "path traversal detected" | Pattern contains `../` or absolute paths | Use relative paths from notebook root |
| "query preprocessing failed" | Pattern processing error | Check quote matching and pattern format |
| "Function does not exist" | Typo in function name | Check function spelling and availability |
| "Cannot access field" | Invalid metadata field reference | Check available metadata keys with sample query |

### Security Error Details

**Path Traversal Protection**:
```sql
-- ❌ These trigger "path traversal detected" errors
read_markdown('../secret/*.md')
read_markdown('../../other-notebook/*.md')  
read_markdown('/absolute/path/*.md')

-- ✅ These work correctly
read_markdown('subfolder/*.md')
read_markdown('**/*.md')
read_markdown('docs/archive/*.md')
```

**Pattern Validation**:
```sql  
-- ❌ These may trigger "query preprocessing failed"
read_markdown('*.md")          -- Mismatched quotes
read_markdown("*.md')          -- Wrong quote type
read_markdown('unclosed        -- Unclosed quotes

-- ✅ These have correct syntax
read_markdown('*.md')          -- Single quotes
read_markdown("*.md")          -- Double quotes
read_markdown('**/*.md')       -- Complex patterns
```

---

For more detailed examples and usage patterns, see the [SQL Guide](sql-guide.md).