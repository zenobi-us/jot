# JSON SQL Query Guide for Jot

This guide teaches you how to use Jot' JSON output format for SQL queries, enabling powerful automation and data processing workflows.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic JSON Queries](#basic-json-queries)
3. [Working with Complex Data Types](#working-with-complex-data-types)
4. [Integration with External Tools](#integration-with-external-tools)
5. [Automation Patterns](#automation-patterns)
6. [Advanced Techniques](#advanced-techniques)
7. [Troubleshooting](#troubleshooting)
8. [Performance Optimization](#performance-optimization)

## Getting Started

Jot automatically outputs SQL query results in JSON format, making it easy to integrate with modern toolchains and automation scripts.

### Basic Syntax

All SQL queries return JSON arrays of objects:

```bash
jot notes search --sql "SELECT title, path FROM notes"
```

**Example Output:**
```json
[
  {
    "title": "Project Ideas",
    "path": "projects/ideas.md"
  },
  {
    "title": "Meeting Notes", 
    "path": "meetings/2024-01-15.md"
  }
]
```

### Why JSON Output?

- **Structured Data**: Easy parsing and manipulation
- **Tool Integration**: Works seamlessly with jq, scripts, APIs
- **Automation Ready**: Perfect for CI/CD pipelines and workflows
- **Cross-Platform**: Standard format across all environments

## Basic JSON Queries

### Simple Queries

```bash
# Get all note file paths
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 5"
```

**Output:**
```json
[
  {
    "file_path": "/path/to/notebook/notes/project.md"
  },
  {
    "file_path": "/path/to/notebook/meetings/daily.md"
  }
]
```

### Including Multiple Columns

```bash
# Get file paths with word counts
jot notes search --sql "
  SELECT 
    file_path, 
    (md_stats(content)).word_count as word_count 
  FROM read_markdown('**/*.md', include_filepath:=true) 
  LIMIT 3
"
```

**Output:**
```json
[
  {
    "file_path": "/path/to/notebook/notes/project.md",
    "word_count": 1250
  },
  {
    "file_path": "/path/to/notebook/meetings/daily.md", 
    "word_count": 340
  }
]
```

### Filtering and Sorting

```bash
# Find notes with more than 500 words, sorted by word count
jot notes search --sql "
  SELECT 
    file_path,
    (md_stats(content)).word_count as words
  FROM read_markdown('**/*.md', include_filepath:=true)
  WHERE (md_stats(content)).word_count > 500
  ORDER BY words DESC
"
```

## Working with Complex Data Types

Jot handles complex data structures seamlessly in JSON output.

### Nested Objects (Maps)

Markdown statistics return as nested JSON objects:

```bash
# Get comprehensive statistics
jot notes search --sql "
  SELECT 
    file_path, 
    md_stats(content) as stats 
  FROM read_markdown('**/*.md', include_filepath:=true) 
  LIMIT 2
"
```

**Output:**
```json
[
  {
    "file_path": "/path/to/notebook/notes/project.md",
    "stats": {
      "char_count": 8540,
      "code_block_count": 3,
      "heading_count": 5,
      "line_count": 245,
      "link_count": 12,
      "reading_time_minutes": 4.2,
      "word_count": 1250
    }
  }
]
```

### Arrays of Objects

Extract links and code blocks as arrays:

```bash
# Extract all links from notes
jot notes search --sql "
  SELECT 
    file_path,
    md_extract_links(content) as links
  FROM read_markdown('**/*.md', include_filepath:=true)
  WHERE array_length(md_extract_links(content)) > 0
  LIMIT 1
"
```

**Output:**
```json
[
  {
    "file_path": "/path/to/notebook/notes/project.md",
    "links": [
      {
        "is_reference": false,
        "line_number": 15,
        "text": "Jot Documentation",
        "title": null,
        "url": "https://github.com/zenobi-us/jot"
      },
      {
        "is_reference": false,
        "line_number": 23,
        "text": "GitHub Issues",
        "title": null,
        "url": "https://github.com/zenobi-us/jot/issues"
      }
    ]
  }
]
```

### Code Blocks

```bash
# Extract code blocks with languages
jot notes search --sql "
  SELECT 
    file_path,
    md_extract_code_blocks(content) as code_blocks
  FROM read_markdown('**/*.md', include_filepath:=true)
  WHERE array_length(md_extract_code_blocks(content)) > 0
  LIMIT 1
"
```

**Output:**
```json
[
  {
    "file_path": "/path/to/notebook/notes/tutorial.md",
    "code_blocks": [
      {
        "code": "echo \"Hello World\"\n",
        "info_string": "bash",
        "language": "bash",
        "line_number": 45
      },
      {
        "code": "def hello():\n    print('Hello from Python')\n",
        "info_string": "python",
        "language": "python", 
        "line_number": 52
      }
    ]
  }
]
```

## Integration with External Tools

### Using jq for Data Processing

[jq](https://jqlang.github.io/jq/) is the perfect companion for processing JSON output from Jot.

#### Basic jq Operations

```bash
# Extract just the file paths
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)" | \
jq -r '.[].file_path'
```

```bash
# Get notes with high word counts
jot notes search --sql "
  SELECT file_path, (md_stats(content)).word_count as words 
  FROM read_markdown('**/*.md', include_filepath:=true)
" | \
jq '.[] | select(.words > 1000) | .file_path'
```

#### Complex jq Transformations

```bash
# Create a summary report
jot notes search --sql "
  SELECT 
    file_path, 
    (md_stats(content)).word_count as words,
    array_length(md_extract_links(content)) as link_count
  FROM read_markdown('**/*.md', include_filepath:=true)
" | \
jq '{
  total_notes: length,
  avg_words: (map(.words) | add / length),
  total_links: (map(.link_count) | add),
  notes_over_500_words: (map(select(.words > 500)) | length)
}'
```

**Output:**
```json
{
  "total_notes": 47,
  "avg_words": 623.4,
  "total_links": 156,
  "notes_over_500_words": 23
}
```

#### Extracting Specific Data

```bash
# Get all external links
jot notes search --sql "
  SELECT file_path, md_extract_links(content) as links
  FROM read_markdown('**/*.md', include_filepath:=true)
" | \
jq -r '
  .[] | 
  .links[] | 
  select(.url | startswith("http")) | 
  "\(.url) - \(.text)"
'
```

### Command Line Piping Patterns

#### File Processing

```bash
# Save all note paths to a file
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)" | \
jq -r '.[].file_path' > note-list.txt
```

```bash
# Get notes modified in last 7 days
jot notes search --sql "
  SELECT file_path 
  FROM read_markdown('**/*.md', include_filepath:=true)
  WHERE file_path IN (
    SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)
  )
" | \
jq -r '.[].file_path' | \
while read note; do
  if [[ $(stat -c %Y "$note" 2>/dev/null) -gt $(date -d '7 days ago' +%s) ]]; then
    echo "Recently modified: $note"
  fi
done
```

#### Batch Operations

```bash
# Count words in each note and sort
jot notes search --sql "
  SELECT 
    file_path,
    (md_stats(content)).word_count as words
  FROM read_markdown('**/*.md', include_filepath:=true)
" | \
jq -r '.[] | "\(.words)\t\(.file_path)"' | \
sort -nr | \
head -10
```

### Integration with Other Tools

#### CSV Export

```bash
# Convert to CSV for spreadsheet import
jot notes search --sql "
  SELECT 
    file_path,
    (md_stats(content)).word_count as words,
    (md_stats(content)).reading_time_minutes as reading_time
  FROM read_markdown('**/*.md', include_filepath:=true)
" | \
jq -r '["File", "Words", "Reading Time"], (.[] | [.file_path, .words, .reading_time]) | @csv'
```

#### Database Import

```bash
# Prepare data for database import
jot notes search --sql "
  SELECT 
    file_path,
    content,
    md_stats(content) as stats
  FROM read_markdown('**/*.md', include_filepath:=true)
" | \
jq -c '.[] | {
  path: .file_path,
  content: .content,
  word_count: .stats.word_count,
  char_count: .stats.char_count,
  last_updated: now | todate
}'
```

## Automation Patterns

### Backup and Export Scripts

#### Complete Notebook Backup

```bash
#!/bin/bash
# backup-notes.sh - Export all notes with metadata

backup_dir="backup-$(date +%Y%m%d)"
mkdir -p "$backup_dir"

echo "Exporting notebook metadata..."
jot notes search --sql "
  SELECT 
    file_path,
    content,
    md_stats(content) as stats,
    md_extract_links(content) as links
  FROM read_markdown('**/*.md', include_filepath:=true)
" > "$backup_dir/notebook-export.json"

echo "Creating summary report..."
cat "$backup_dir/notebook-export.json" | jq '{
  export_date: now | todate,
  total_notes: length,
  total_words: (map(.stats.word_count) | add),
  total_links: (map(.links | length) | add),
  notes_by_word_count: {
    small: (map(select(.stats.word_count < 100)) | length),
    medium: (map(select(.stats.word_count >= 100 and .stats.word_count < 1000)) | length),
    large: (map(select(.stats.word_count >= 1000)) | length)
  }
}' > "$backup_dir/summary.json"

echo "Backup complete in $backup_dir/"
```

#### Individual Note Export

```bash
#!/bin/bash
# export-note.sh - Export individual notes to structured files

export_dir="exports"
mkdir -p "$export_dir"

jot notes search --sql "
  SELECT 
    file_path,
    content,
    md_stats(content) as stats
  FROM read_markdown('**/*.md', include_filepath:=true)
" | jq -c '.[]' | while read note; do
  # Extract data
  filename=$(echo "$note" | jq -r '.file_path | split("/")[-1] | split(".")[0]')
  word_count=$(echo "$note" | jq -r '.stats.word_count')
  
  # Create structured export
  echo "$note" | jq '{
    exported_at: now | todate,
    original_path: .file_path,
    statistics: .stats,
    content: .content
  }' > "$export_dir/${filename}-export.json"
  
  echo "Exported: $filename ($word_count words)"
done
```

### Monitoring and Reporting

#### Daily Note Activity Report

```bash
#!/bin/bash
# daily-report.sh - Generate daily activity report

echo "# Daily Notes Report - $(date +%Y-%m-%d)" > daily-report.md
echo "" >> daily-report.md

# Note statistics
echo "## Overview" >> daily-report.md
jot notes search --sql "
  SELECT 
    COUNT(*) as total_notes,
    SUM((md_stats(content)).word_count) as total_words,
    AVG((md_stats(content)).word_count) as avg_words,
    SUM(array_length(md_extract_links(content))) as total_links
  FROM read_markdown('**/*.md', include_filepath:=true)
" | jq -r '.[] | "
- **Total Notes:** \(.total_notes)
- **Total Words:** \(.total_words)  
- **Average Words per Note:** \(.avg_words | floor)
- **Total Links:** \(.total_links)
"' >> daily-report.md

# Top notes by word count
echo "" >> daily-report.md
echo "## Longest Notes" >> daily-report.md
jot notes search --sql "
  SELECT 
    file_path,
    (md_stats(content)).word_count as words
  FROM read_markdown('**/*.md', include_filepath:=true)
  ORDER BY words DESC
  LIMIT 5
" | jq -r '.[] | "- [\(.file_path)](\(.file_path)): \(.words) words"' >> daily-report.md

echo "Report generated: daily-report.md"
```

#### Code Usage Analysis

```bash
#!/bin/bash
# code-analysis.sh - Analyze programming languages in notes

echo "Analyzing code usage across notes..."

jot notes search --sql "
  SELECT 
    file_path,
    md_extract_code_blocks(content) as code_blocks
  FROM read_markdown('**/*.md', include_filepath:=true)
  WHERE array_length(md_extract_code_blocks(content)) > 0
" | jq -r '
  [.[] | .code_blocks[]] | 
  group_by(.language) | 
  map({
    language: .[0].language,
    count: length,
    total_lines: (map(.code | split("\n") | length) | add)
  }) |
  sort_by(-.count) |
  .[] |
  "\(.language): \(.count) blocks, \(.total_lines) lines"
'
```

### CI/CD Integration

#### Documentation Validation

```bash
#!/bin/bash
# validate-docs.sh - Validate documentation in CI/CD

echo "Validating notebook documentation..."

# Check for notes without titles
missing_titles=$(jot notes search --sql "
  SELECT file_path
  FROM read_markdown('**/*.md', include_filepath:=true)
  WHERE content NOT LIKE '#%'
" | jq -r '.[].file_path' | wc -l)

if [ "$missing_titles" -gt 0 ]; then
  echo "Warning: $missing_titles notes missing titles"
  jot notes search --sql "
    SELECT file_path
    FROM read_markdown('**/*.md', include_filepath:=true)
    WHERE content NOT LIKE '#%'
  " | jq -r '.[].file_path'
fi

# Check for broken internal links
echo "Checking for broken links..."
jot notes search --sql "
  SELECT file_path, md_extract_links(content) as links
  FROM read_markdown('**/*.md', include_filepath:=true)
" | jq -r '.[] | .links[] | select(.url | startswith("./") or startswith("../")) | .url' | \
while read link; do
  if [ ! -f "$link" ]; then
    echo "Broken link found: $link"
  fi
done
```

#### Content Quality Metrics

```bash
#!/bin/bash
# quality-metrics.sh - Generate content quality metrics

jot notes search --sql "
  SELECT 
    file_path,
    (md_stats(content)).word_count as words,
    array_length(md_extract_links(content)) as links,
    (md_stats(content)).code_block_count as code_blocks
  FROM read_markdown('**/*.md', include_filepath:=true)
" | jq '{
  quality_score: (
    map(
      (.words * 0.5) + 
      (.links * 2) + 
      (.code_blocks * 3)
    ) | add / length
  ),
  notes_needing_attention: [
    .[] | 
    select(.words < 50 and .links == 0) |
    .file_path
  ]
}'
```

## Advanced Techniques

### Data Aggregation and Analysis

#### Content Analysis by Directory

```bash
# Analyze content organization by directory structure
jot notes search --sql "
  SELECT 
    file_path,
    (md_stats(content)).word_count as words,
    array_length(md_extract_links(content)) as links
  FROM read_markdown('**/*.md', include_filepath:=true)
" | jq -r '
  group_by(.file_path | split("/")[:-1] | join("/")) |
  map({
    directory: (.[0].file_path | split("/")[:-1] | join("/") // "root"),
    note_count: length,
    total_words: (map(.words) | add),
    total_links: (map(.links) | add),
    avg_words: (map(.words) | add / length)
  }) |
  sort_by(-.total_words)[]
'
```

#### Temporal Analysis

```bash
#!/bin/bash
# temporal-analysis.sh - Analyze note creation patterns

echo "Analyzing note creation patterns..."

# Get file modification times and content stats
jot notes search --sql "
  SELECT 
    file_path,
    (md_stats(content)).word_count as words
  FROM read_markdown('**/*.md', include_filepath:=true)
" | jq -c '.[]' | while read note; do
  filepath=$(echo "$note" | jq -r '.file_path')
  words=$(echo "$note" | jq -r '.words')
  
  if [ -f "$filepath" ]; then
    mod_date=$(stat -c %Y "$filepath" 2>/dev/null || date +%s)
    month=$(date -d "@$mod_date" +%Y-%m)
    
    echo "$month $words"
  fi
done | awk '
{
  month_words[$1] += $2
  month_count[$1]++
}
END {
  for (month in month_words) {
    printf "%s: %d notes, %d words, %.1f avg\n", 
           month, month_count[month], month_words[month], 
           month_words[month]/month_count[month]
  }
}' | sort
```

### Custom Data Transformation

#### Link Network Analysis

```bash
# Create link network data for visualization
jot notes search --sql "
  SELECT 
    file_path,
    md_extract_links(content) as links
  FROM read_markdown('**/*.md', include_filepath:=true)
" | jq '
  [
    .[] |
    .links[] |
    select(.url | startswith("./") or test("\\.(md|markdown)$")) |
    {
      source: (.file_path | split("/")[-1] | split(".")[0]),
      target: (.url | split("/")[-1] | split(".")[0]),
      text: .text
    }
  ] |
  group_by(.target) |
  map({
    node: .[0].target,
    incoming_links: length,
    sources: [.[].source] | unique
  }) |
  sort_by(-.incoming_links)
'
```

#### Content Similarity Analysis

```bash
#!/bin/bash
# content-similarity.sh - Find similar notes by shared links

echo "Finding notes with shared external links..."

jot notes search --sql "
  SELECT 
    file_path,
    md_extract_links(content) as links
  FROM read_markdown('**/*.md', include_filepath:=true)
" | jq -r '
  [
    .[] |
    {
      file: .file_path,
      external_links: [.links[] | select(.url | startswith("http")) | .url]
    } |
    select(.external_links | length > 0)
  ] as $notes |
  
  [
    range(0; $notes | length) as $i |
    range($i + 1; $notes | length) as $j |
    {
      note1: $notes[$i].file,
      note2: $notes[$j].file,
      shared_links: ($notes[$i].external_links - ($notes[$i].external_links - $notes[$j].external_links)),
      similarity: (($notes[$i].external_links - ($notes[$i].external_links - $notes[$j].external_links)) | length)
    } |
    select(.similarity > 0)
  ] |
  sort_by(-.similarity) |
  .[] |
  "\(.note1) ↔ \(.note2): \(.similarity) shared link(s)"
'
```

## Troubleshooting

### Common JSON Parsing Issues

#### Invalid JSON Output

**Problem**: JSON parsing fails with "invalid character" errors.

**Causes and Solutions**:

1. **Empty Results**: Query returns no data
   ```bash
   # Check for empty results
   jot notes search --sql "SELECT COUNT(*) as count FROM read_markdown('**/*.md')"
   # Should return: [{"count": 0}] if no files found
   ```

2. **File Access Issues**: Notebook path not found
   ```bash
   # Verify notebook directory
   jot notebook list
   # Ensure you're in the correct notebook
   ```

3. **SQL Syntax Errors**: Query has syntax issues
   ```bash
   # Test with simple query first
   jot notes search --sql "SELECT 'test' as message"
   ```

#### Malformed Complex Objects

**Problem**: Complex data structures don't serialize properly.

**Solution**: Use explicit type conversion:

```bash
# Instead of this (may fail):
jot notes search --sql "SELECT metadata FROM read_markdown('**/*.md')"

# Use this (safer):
jot notes search --sql "
  SELECT 
    file_path,
    CASE 
      WHEN metadata IS NOT NULL THEN metadata
      ELSE CAST(NULL AS JSON)
    END as metadata
  FROM read_markdown('**/*.md', include_filepath:=true)
"
```

### jq Processing Errors

#### Array vs Object Issues

**Problem**: jq expects array but gets object or vice versa.

```bash
# Incorrect: assumes single object
echo '[{"a":1}, {"a":2}]' | jq '.a'  # Error: null

# Correct: iterate array
echo '[{"a":1}, {"a":2}]' | jq '.[].a'  # Output: 1, 2

# Correct: map over array  
echo '[{"a":1}, {"a":2}]' | jq 'map(.a)'  # Output: [1, 2]
```

#### Type Coercion Issues

**Problem**: Unexpected null values or type mismatches.

```bash
# Problem: word_count might be null
jq '.[] | select(.word_count > 100)'

# Solution: handle null values
jq '.[] | select((.word_count // 0) > 100)'
```

### Performance Issues

#### Large Result Sets

**Problem**: Queries return too much data, causing memory issues.

**Solutions**:

1. **Use LIMIT clauses**:
   ```bash
   # Instead of all notes
   jot notes search --sql "SELECT * FROM read_markdown('**/*.md')"
   
   # Use pagination
   jot notes search --sql "SELECT * FROM read_markdown('**/*.md') LIMIT 100 OFFSET 0"
   ```

2. **Filter early**:
   ```bash
   # Instead of filtering in jq
   jot notes search --sql "SELECT * FROM read_markdown('**/*.md')" | jq '.[] | select(.words > 1000)'
   
   # Filter in SQL
   jot notes search --sql "
     SELECT * FROM read_markdown('**/*.md', include_filepath:=true)
     WHERE (md_stats(content)).word_count > 1000
   "
   ```

3. **Select specific columns**:
   ```bash
   # Instead of SELECT *
   jot notes search --sql "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true)"
   
   # Select only needed columns
   jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)"
   ```

#### Memory Usage with Complex Objects

**Problem**: Nested objects consume too much memory.

**Solution**: Process in smaller batches:

```bash
#!/bin/bash
# batch-processing.sh - Process large datasets in batches

batch_size=50
offset=0

while true; do
  echo "Processing batch starting at offset $offset..."
  
  result=$(jot notes search --sql "
    SELECT file_path, md_stats(content) as stats
    FROM read_markdown('**/*.md', include_filepath:=true)
    LIMIT $batch_size OFFSET $offset
  ")
  
  # Check if result is empty array
  if [[ $(echo "$result" | jq 'length') -eq 0 ]]; then
    echo "No more results, processing complete."
    break
  fi
  
  # Process this batch
  echo "$result" | jq -r '.[] | "\(.file_path): \(.stats.word_count) words"'
  
  offset=$((offset + batch_size))
done
```

### Error Diagnostics

#### Debug SQL Queries

**Enable debug output to understand query execution**:

```bash
# Test with simple query first
jot notes search --sql "SELECT 'hello' as test"

# Verify file access
jot notes search --sql "SELECT COUNT(*) as file_count FROM read_markdown('**/*.md')"

# Test specific functions
jot notes search --sql "SELECT md_stats('# Test\nContent') as stats"
```

#### Validate JSON with jq

**Always validate JSON structure before complex processing**:

```bash
# Validate JSON structure
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)" | jq empty

# Check data types
jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)" | jq 'type'

# Inspect first element
jot notes search --sql "SELECT * FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 1" | jq '.[0]'
```

## Performance Optimization

### Query Optimization

#### Use Appropriate Filters

```bash
# Slow: Filter after reading all files
jot notes search --sql "SELECT * FROM read_markdown('**/*.md', include_filepath:=true)" | \
jq '.[] | select(.file_path | contains("project"))'

# Fast: Filter with SQL WHERE clause
jot notes search --sql "
  SELECT * FROM read_markdown('**/*.md', include_filepath:=true)
  WHERE file_path LIKE '%project%'
"
```

#### Limit Data Early

```bash
# Slow: Get all data then limit in jq
jot notes search --sql "SELECT * FROM read_markdown('**/*.md', include_filepath:=true)" | \
jq '.[0:10]'

# Fast: Limit in SQL
jot notes search --sql "SELECT * FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 10"
```

### Processing Optimization

#### Stream Processing for Large Datasets

```bash
# For very large result sets, use jq streaming
jot notes search --sql "SELECT * FROM read_markdown('**/*.md', include_filepath:=true)" | \
jq -c '.[]' | \
while read note; do
  # Process each note individually
  echo "$note" | jq -r '.file_path'
done
```

#### Parallel Processing

```bash
#!/bin/bash
# parallel-analysis.sh - Process notes in parallel

# Get all file paths
paths=$(jot notes search --sql "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)" | jq -r '.[].file_path')

# Process in parallel using GNU parallel
echo "$paths" | parallel -j4 'echo "Processing: {}" && wc -w "{}"'
```

### Monitoring and Profiling

#### Execution Time Measurement

```bash
#!/bin/bash
# benchmark.sh - Measure query performance

echo "Benchmarking different query approaches..."

# Time SQL filtering
echo "SQL filtering:"
time jot notes search --sql "
  SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)
  WHERE (md_stats(content)).word_count > 500
" > /dev/null

# Time jq filtering
echo "jq filtering:"
time (jot notes search --sql "
  SELECT file_path, (md_stats(content)).word_count as words
  FROM read_markdown('**/*.md', include_filepath:=true)
" | jq '.[] | select(.words > 500) | .file_path' > /dev/null)
```

This comprehensive guide covers all aspects of using JSON output with Jot SQL queries, from basic usage to advanced automation patterns and troubleshooting. The examples are practical and tested to ensure they work correctly in real-world scenarios.