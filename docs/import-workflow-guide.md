# Import Workflow Guide

This guide walks you through importing your existing markdown collection into OpenNotes. Whether you're migrating from another tool or organizing existing files, this guide covers all scenarios and common issues.

## Table of Contents

1. [Why Import Matters](#why-import-matters)
2. [Import Process Overview](#import-process-overview)
3. [Step-by-Step Import](#step-by-step-import)
4. [Collection Organization Patterns](#collection-organization-patterns)
5. [First-Time Setup Workflows](#first-time-setup-workflows)
6. [Preserving Metadata](#preserving-metadata)
7. [Migration from Other Systems](#migration-from-other-systems)
8. [Troubleshooting](#troubleshooting)
9. [Next Steps](#next-steps)

## Why Import Matters

Importing existing markdown gives you **immediate value** from OpenNotes without starting from scratch:

- **Access existing knowledge**: Your notes are already in markdown, no format conversion needed
- **Preserve organization**: Keep your folder structure or reorganize gradually
- **Enable SQL power**: Start querying your collection instantly
- **Gradual adoption**: Use OpenNotes alongside existing workflows

OpenNotes respects your existing filesystem structure—it doesn't lock your notes into a proprietary format. You can edit files directly in your editor, and OpenNotes sees the changes immediately.

---

## Import Process Overview

### The Basic Flow

```
1. Prepare your markdown collection
2. Create a notebook pointing to it
3. Verify the import succeeded
4. Execute first SQL query to unlock power
```

### Key Concepts

**Notebook**: An OpenNotes context pointing to a directory containing markdown files. One notebook = one collection of notes.

**File Pattern**: How OpenNotes discovers files in your collection. Default pattern `**/*.md` recursively finds all markdown files.

**Metadata Extraction**: OpenNotes automatically extracts:
- **Title**: From frontmatter `title` field, first heading (H1), or filename
- **Content**: Full markdown content including all formatting
- **Path**: Relative path from notebook root

---

## Step-by-Step Import

### Step 1: Prepare Your Collection

Before importing, ensure your collection is organized:

```bash
# Navigate to your notes directory
cd ~/my-notes

# Check the structure
find . -name "*.md" -type f | head -20
```

Expected output:
```
./README.md
./projects/project-1.md
./projects/project-2.md
./archive/2023/old-notes.md
./daily/2024-01-15.md
./daily/2024-01-16.md
```

**Good practices before import**:
- ✅ Remove symbolic links (or follow them: see [Symlinks](#symlinks-and-nested-structures))
- ✅ Clean up duplicate files (search across collection)
- ✅ Verify file encoding (UTF-8 recommended)
- ✅ Ensure write permissions on notebook directory

### Step 2: Create a Notebook

Create an OpenNotes notebook pointing to your collection:

```bash
opennotes notebook create "My Notes" --path ~/my-notes
```

Output:
```
Created notebook: My Notes
Location: /home/user/my-notes
Files discovered: 157
```

**What happens**:
- Notebook config created at `~/.config/opennotes/config.json`
- OpenNotes scans your directory for all `*.md` files recursively
- Metadata extraction begins in background
- No files are copied or moved

**Set as current notebook** (optional):
```bash
# List all notebooks
opennotes notebook list

# If needed, set default by path
cd ~/my-notes
opennotes notes list
```

### Step 3: Verify Import Success

Check that your files were discovered:

```bash
# Count total notes
opennotes notes list | wc -l

# Show first 10 notes with titles
opennotes notes list | head -10

# Verify a specific file was found
opennotes notes search "filename"
```

Example output:
```
### Notes (157)

- [Project Alpha Kickoff] projects/project-alpha/kickoff.md
- [Meeting Notes Jan 15] daily/2024-01-15.md
- [TODO List] todo.md
```

**Verify metadata extraction**:
```bash
# Check if titles are extracted from frontmatter
opennotes notes list | grep -i "title"

# List notes with word count (requires SQL)
opennotes notes search --sql "SELECT file_path, (md_stats(content)).word_count as words FROM read_markdown('**/*.md', include_filepath:=true) ORDER BY words DESC LIMIT 10"
```

### Step 4: Execute First SQL Query

Unlock SQL power to verify the import worked completely:

```bash
# Find all notes with "TODO" in content
opennotes notes search --sql \
  "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true) WHERE content ILIKE '%TODO%' LIMIT 5"

# Get statistics about your collection
opennotes notes search --sql \
  "SELECT COUNT(*) as total_notes, SUM((md_stats(content)).word_count) as total_words FROM read_markdown('**/*.md')"
```

**Success indicators**:
- ✅ Correct number of files returned
- ✅ File paths and content are accurate
- ✅ Metadata extraction working (titles, word counts)

---

## Collection Organization Patterns

### Pattern 1: Flat Structure

All notes in a single directory (simplest):

```
my-notes/
├── project-1.md
├── project-2.md
├── meeting-notes.md
├── todo.md
└── README.md
```

**Import command**:
```bash
opennotes notebook create "My Notes" --path ~/my-notes
```

**Query all notes**:
```bash
opennotes notes search --sql "SELECT file_path FROM read_markdown('*.md')"
```

**Best for**: Small collections (<50 notes), single-topic notes, quick reference files

---

### Pattern 2: Hierarchical by Folder

Notes organized in subfolders (recommended):

```
my-notes/
├── projects/
│   ├── project-alpha/
│   │   ├── spec.md
│   │   ├── progress.md
│   │   └── backlog.md
│   └── project-beta/
│       └── overview.md
├── daily/
│   ├── 2024-01-15.md
│   ├── 2024-01-16.md
│   └── 2024-01-17.md
├── archive/
│   └── 2023/
│       └── old-notes.md
└── reference/
    ├── design-patterns.md
    └── api-guide.md
```

**Import command**:
```bash
opennotes notebook create "My Notes" --path ~/my-notes
```

**Query notes by folder**:
```bash
# All project notes
opennotes notes search --sql "SELECT file_path FROM read_markdown('projects/**/*.md', include_filepath:=true)"

# Recent daily notes
opennotes notes search --sql "SELECT file_path FROM read_markdown('daily/2024-*.md', include_filepath:=true) ORDER BY file_path DESC LIMIT 7"

# Specific project
opennotes notes search --sql "SELECT file_path FROM read_markdown('projects/project-alpha/*.md', include_filepath:=true)"
```

**Best for**: Medium to large collections (50-1000+ notes), multiple projects, organized by context

---

### Pattern 3: Multi-Project Environment

Multiple notebooks for different contexts:

```
work-notes/
├── projects/
├── meetings/
└── todos/

personal-notes/
├── learning/
├── ideas/
└── journal/

client-a/
├── documentation/
└── project-files/
```

**Import commands**:
```bash
opennotes notebook create "Work" --path ~/work-notes
opennotes notebook create "Personal" --path ~/personal-notes
opennotes notebook create "Client A" --path ~/client-a
```

**Switch between notebooks by directory**:
```bash
# OpenNotes auto-detects from current directory
cd ~/work-notes && opennotes notes list        # Uses "Work" notebook
cd ~/personal-notes && opennotes notes list    # Uses "Personal" notebook

# Or specify explicitly
opennotes notes list --notebook "Client A"
```

**Best for**: Power users managing multiple contexts, teams with shared repositories, client work

---

## First-Time Setup Workflows

### Workflow A: Solo Developer with Personal Notes

You have Obsidian vault, Bear notes, or scattered markdown files.

**Setup (5 minutes)**:
```bash
# 1. Copy/organize notes into a directory
mkdir -p ~/my-notes
cp ~/Documents/*.md ~/my-notes/
cp ~/Dropbox/Notes/*.md ~/my-notes/

# 2. Create notebook
opennotes notebook create "Personal Notes" --path ~/my-notes

# 3. Verify import
opennotes notes list | head -20

# 4. First SQL query
opennotes notes search --sql "SELECT COUNT(*) as total FROM read_markdown('**/*.md')"
```

**First workflow**: Search and explore
```bash
# Find notes about "Python"
opennotes notes search "Python"

# Get top 10 longest notes
opennotes notes search --sql \
  "SELECT file_path, (md_stats(content)).word_count FROM read_markdown('**/*.md') ORDER BY word_count DESC LIMIT 10"
```

---

### Workflow B: Team with Shared Knowledge Base

Your team uses a Git repository for shared documentation.

**Setup (10 minutes)**:
```bash
# 1. Clone or navigate to repo
cd ~/projects/shared-knowledge

# 2. Create notebook
opennotes notebook create "Team Knowledge" --path ~/projects/shared-knowledge

# 3. Verify structure
opennotes notes list

# 4. Test team search
opennotes notes search "API documentation"
```

**First workflow**: Generate team reports
```bash
# Find all documentation that needs updating
opennotes notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md') WHERE content LIKE '%TODO%' OR content LIKE '%FIXME%' LIMIT 20"

# Export as JSON for processing
opennotes notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md')" | jq -r '.[].file_path'
```

**Git integration**:
```bash
# Notebook points to git repo
# Changes to .md files are tracked by git
# Use OpenNotes for querying, git for collaboration
git add .
git commit -m "Updated documentation"
```

---

### Workflow C: Multi-Project Environment

You manage multiple projects with separate note repositories.

**Setup (15 minutes)**:
```bash
# 1. Create notebooks for each project
opennotes notebook create "Project Alpha" --path ~/projects/alpha/notes
opennotes notebook create "Project Beta" --path ~/projects/beta/notes
opennotes notebook create "Archive" --path ~/projects/archive

# 2. Verify each notebook
opennotes notebook list

# 3. Switch to a project
cd ~/projects/alpha/notes
opennotes notes list

# 4. Cross-project query (requires manual SQL)
# Create a script to search all notebooks (see Advanced below)
```

**First workflow**: Quick project context switching
```bash
# Search in current project
cd ~/projects/alpha/notes && opennotes notes search "feature-x"

# Switch to another project
cd ~/projects/beta/notes && opennotes notes search "bug-report"

# List all projects
opennotes notebook list
```

---

## Preserving Metadata

### Frontmatter Extraction

OpenNotes automatically extracts metadata from YAML frontmatter:

**Example note with frontmatter**:
```markdown
---
title: Project Alpha Kickoff
tags: project, planning
status: active
date: 2024-01-15
---

# Project Details

... content ...
```

**How OpenNotes handles it**:
- **Title source**: Uses `title` field first, falls back to first `# Heading`, then filename
- **Content**: Full markdown including frontmatter preserved as-is
- **Metadata access**: In SQL queries, frontmatter available as structured fields

**Query with metadata**:
```bash
opennotes notes search --sql \
  "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 5"
```

**Note**: DuckDB markdown extension provides access to frontmatter. Check [SQL Functions Reference](sql-functions-reference.md) for advanced metadata queries.

### Automatic Title Detection

If you don't use frontmatter, OpenNotes detects titles from content:

| Priority | Source | Example |
|----------|--------|---------|
| 1 | Frontmatter `title` | `title: "My Note"` |
| 2 | First H1 heading | `# My Note` |
| 3 | Filename | `my-note.md` |

**Examples**:
```markdown
# File: meeting-notes.md
# Frontmatter: (none)
# First heading: # Team Sync January 15
# Result: Title = "Team Sync January 15"
```

```markdown
# File: project-spec.md
# Frontmatter: title: "Alpha Project Specification"
# First heading: # Overview
# Result: Title = "Alpha Project Specification"
```

### Custom Metadata Handling

For structured metadata beyond frontmatter, use SQL queries:

```bash
# Extract frontmatter-like data
opennotes notes search --sql \
  "SELECT file_path, content FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 10"

# Parse content structure (requires md_extract_headings)
opennotes notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%---' LIMIT 10"
```

See [SQL Functions Reference](sql-functions-reference.md) for advanced metadata extraction patterns.

---

## Migration from Other Systems

### Obsidian Vault Import

Obsidian stores notes as markdown with optional frontmatter (like OpenNotes):

**Step 1: Export from Obsidian**
```bash
# Obsidian vaults are already markdown files
# Navigate to your vault directory
cd ~/Obsidian/My-Vault

# Verify structure
find . -name "*.md" -type f | head -10
```

**Step 2: Import into OpenNotes**
```bash
opennotes notebook create "Obsidian Import" --path ~/Obsidian/My-Vault
```

**Step 3: Verify Import**
```bash
# Check note count
opennotes notes list | wc -l

# Test SQL query
opennotes notes search --sql "SELECT COUNT(*) as total FROM read_markdown('**/*.md')"

# Search for specific content
opennotes notes search "important"
```

**Handle Obsidian-specific features**:
- ✅ **Frontmatter**: OpenNotes preserves YAML frontmatter
- ✅ **Wikilinks**: Content preserved as-is (rendered as `[[link]]` text)
- ✅ **Tags**: Stored in content, queryable with SQL `LIKE '%#tag%'`
- ❌ **Obsidian plugins**: Not supported (use plain markdown content)
- ❌ **Vault settings**: Not imported (use OpenNotes config instead)

**Query Obsidian tags**:
```bash
opennotes notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) WHERE content LIKE '%#project%' LIMIT 20"
```

---

### Bear Notes Migration

Bear uses proprietary database format, but supports markdown export:

**Step 1: Export from Bear**
1. Open Bear
2. Select "File" → "Export Notes"
3. Choose "Markdown" format
4. Save to `~/bear-export`

**Step 2: Import into OpenNotes**
```bash
# Bear exports as folder of .md files
opennotes notebook create "Bear Migration" --path ~/bear-export
```

**Step 3: Verify and Clean**
```bash
# Check for attachments (Bear may include images)
find ~/bear-export -type f ! -name "*.md"

# Test import
opennotes notes list

# Verify content
opennotes notes search --sql "SELECT COUNT(*) FROM read_markdown('**/*.md')"
```

**Handle Bear-specific content**:
- ✅ **Note content**: Fully preserved as markdown
- ✅ **Formatting**: Markdown formatting (bold, italic, etc.) preserved
- ⚠️ **Images**: Exported as separate files, links preserved but external
- ⚠️ **Sketches**: Not exported (re-create or use screenshots)
- ❌ **Pinned notes**: Status lost (recreate organization in OpenNotes)

---

### Generic Markdown Folder Import

You have a folder of markdown files from any source:

**Step 1: Organize Files**
```bash
# Collect all markdown files into one directory
mkdir -p ~/my-notes
find ~/Documents -name "*.md" -type f -exec cp {} ~/my-notes/ \;
find ~/Desktop -name "*.md" -type f -exec cp {} ~/my-notes/ \;
```

**Step 2: Clean Up Naming**
```bash
# Remove special characters from filenames
cd ~/my-notes
for file in *; do
  # Replace spaces with underscores (optional)
  # Remove special characters
  newname=$(echo "$file" | tr ' ' '_' | tr -cd '[:alnum:]._-')
  [ "$file" != "$newname" ] && mv "$file" "$newname"
done
```

**Step 3: Import**
```bash
opennotes notebook create "Imported Notes" --path ~/my-notes
```

**Step 4: Verify**
```bash
opennotes notes list
```

---

## Troubleshooting

### Large Collection Import

**Problem**: Import is slow or seems to hang with 1000+ notes

**Solutions**:

1. **Check progress with SQL**:
   ```bash
   # If this completes quickly, database is working
   opennotes notes search --sql "SELECT COUNT(*) FROM read_markdown('**/*.md')"
   ```

2. **Verify file count**:
   ```bash
   # Count files in filesystem
   find ~/my-notes -name "*.md" | wc -l
   
   # Compare with OpenNotes count
   opennotes notes search --sql "SELECT COUNT(*) FROM read_markdown('**/*.md')"
   ```

3. **Import in batches** (if needed):
   ```bash
   # Create multiple notebooks for different folders
   opennotes notebook create "Notes A-M" --path ~/my-notes/a-m
   opennotes notebook create "Notes N-Z" --path ~/my-notes/n-z
   ```

4. **Check system resources**:
   ```bash
   # Monitor memory usage
   top
   
   # Check disk space
   df -h ~/my-notes
   ```

**Typical performance**:
- 100 notes: <100ms
- 1000 notes: <500ms  
- 10000 notes: 2-5 seconds

---

### Special Characters in Filenames

**Problem**: Files with special characters not found or cause errors

**Examples**:
```
project (2024).md
notes-final-v2.md
client[backup].md
```

**Solutions**:

1. **View problematic files**:
   ```bash
   # Find files with special characters
   find ~/my-notes -name "*.md" -type f | grep -E "[\[\](){}]"
   ```

2. **Rename files**:
   ```bash
   # Remove problematic characters
   cd ~/my-notes
   for file in *; do
     newname=$(echo "$file" | sed 's/[()[\]{}]//g' | sed 's/ /-/g')
     [ "$file" != "$newname" ] && mv "$file" "$newname"
   done
   ```

3. **SQL pattern handling**:
   ```bash
   # If you keep special characters, quote in SQL patterns
   opennotes notes search --sql \
     "SELECT * FROM read_markdown('**/*.md') LIMIT 5"
   ```

**Recommended**: Rename files to use only alphanumeric, hyphens, and underscores:
- ✅ `my-note.md`
- ✅ `project_2024.md`
- ❌ `my (note).md`
- ❌ `project[backup].md`

---

### Symlinks and Nested Structures

**Problem**: Symbolic links or deeply nested folders not working as expected

**Symlinks**:
```bash
# Check for symlinks
find ~/my-notes -type l

# OpenNotes follows symlinks by default
# If a symlink points outside notebook, it may be blocked (security)

# Solution: Copy files instead of symlinking
cp ~/other-notes/*.md ~/my-notes/
```

**Deeply nested structures** (e.g., 5+ levels deep):
```
my-notes/
└── level1/
    └── level2/
        └── level3/
            └── level4/
                └── level5/
                    └── note.md
```

**Handle with SQL patterns**:
```bash
# Matches any depth
opennotes notes search --sql \
  "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true)"

# Specific depth
opennotes notes search --sql \
  "SELECT file_path FROM read_markdown('level1/level2/**/*.md', include_filepath:=true)"
```

**Recommendation**: Keep nesting to 3-4 levels for optimal performance and maintainability.

---

### File Encoding Issues

**Problem**: Non-UTF-8 files show garbled content or errors

**Check encoding**:
```bash
# Find non-UTF-8 files
file -i ~/my-notes/*.md

# Example output:
# notes.md: text/plain; charset=utf-8 ✅
# old-notes.md: text/plain; charset=iso-8859-1 ❌
```

**Convert to UTF-8**:
```bash
# For single file
iconv -f ISO-8859-1 -t UTF-8 old-notes.md -o old-notes-utf8.md
mv old-notes-utf8.md old-notes.md

# For all files in directory
cd ~/my-notes
for file in *.md; do
  iconv -f ISO-8859-1 -t UTF-8 "$file" -o "${file}.utf8" 2>/dev/null && \
  mv "${file}.utf8" "$file"
done
```

---

### Permission Denied Errors

**Problem**: Error when importing: "Permission denied" or "Cannot read directory"

**Check permissions**:
```bash
# List permissions
ls -la ~/my-notes

# Check user ownership
whoami
ls -l ~/my-notes | grep -E "^-"
```

**Fix permissions**:
```bash
# Add read permission
chmod +r ~/my-notes/*.md

# Add execute on directories (allows traversal)
chmod +x ~/my-notes
find ~/my-notes -type d -exec chmod +x {} \;

# For current user recursively
chmod -R u+rX ~/my-notes
```

---

### Import Not Discovering Files

**Problem**: Created notebook but `opennotes notes list` shows "0 notes"

**Debugging**:

1. **Verify notebook creation**:
   ```bash
   opennotes notebook list
   ```

2. **Check directory path**:
   ```bash
   # Verify the path exists
   ls -la ~/my-notes
   
   # Count markdown files
   find ~/my-notes -name "*.md" -type f | wc -l
   ```

3. **Manual SQL query**:
   ```bash
   # If this works, database is OK
   opennotes notes search --sql "SELECT COUNT(*) FROM read_markdown('**/*.md')"
   
   # If this returns 0, no .md files found
   ```

4. **Check file extensions**:
   ```bash
   # Files must be .md (lowercase extension)
   find ~/my-notes -type f | grep -E "\.(md|MD|Md)$"
   
   # Rename if needed
   for file in ~/my-notes/*.MD; do
     [ -f "$file" ] && mv "$file" "${file%.MD}.md"
   done
   ```

5. **Verify notebook path in config**:
   ```bash
   cat ~/.config/opennotes/config.json
   ```

---

### Metadata Not Extracting

**Problem**: Titles showing as "untitled" or file paths instead of proper titles

**Check extraction**:
```bash
# List notes with titles
opennotes notes list

# Should show titles like "My Note Title", not "my-note.md"
```

**Solutions**:

1. **Add frontmatter**:
   ```markdown
   ---
   title: "My Proper Title"
   ---
   
   # Content
   ```

2. **Use H1 headings**:
   ```markdown
   # My Proper Title
   
   Content here...
   ```

3. **Verify extraction with SQL**:
   ```bash
   opennotes notes search --sql \
     "SELECT file_path FROM read_markdown('**/*.md', include_filepath:=true) LIMIT 5"
   ```

---

## Next Steps

After successfully importing your notes:

1. **Learn SQL Querying**: See [SQL Quick Reference](sql-quick-reference.md) for practical query examples
2. **Explore Advanced Features**: Read [SQL Query Guide](sql-guide.md) for comprehensive documentation
3. **Integrate with Tools**: Check [JSON SQL Guide](json-sql-guide.md) for automation examples
4. **Multi-Notebook Setup**: See [Notebook Discovery](notebook-discovery.md) for managing multiple notebooks

---

## Related Guides

- 🚀 **[Getting Started for Power Users](getting-started-power-users.md)** - Complete 15-minute onboarding
- 📚 **[SQL Query Guide](sql-guide.md)** - Full SQL syntax and functions
- 📋 **[SQL Quick Reference](sql-quick-reference.md)** - Practical query patterns
- 🔍 **[Notebook Discovery](notebook-discovery.md)** - Multi-notebook management
- 🤖 **[Automation & JSON Integration](json-sql-guide.md)** - Advanced automation patterns

