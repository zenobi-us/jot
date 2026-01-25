---
name: opennote-vault
description: Use when managing markdown-based notes with SQL queries, creating notebooks with frontmatter metadata, searching notes by content or relationships, or executing reusable query views - handles note organization, DuckDB-powered search, and link analysis
---

# OpenNotes Vault Management

## Overview

OpenNotes manages markdown notes in notebooks with DuckDB-powered SQL queries. Create notes with frontmatter metadata, search by content or links, and execute reusable query views.

**Core principle:** Notes are markdown files + metadata searchable via SQL queries.

## When to Use

Use this skill when you need to:
- Create or manage notebooks of markdown notes
- Add notes with frontmatter metadata (tags, author, status, etc.)
- Search notes by content, metadata, or relationships
- Execute SQL queries against note collections
- Find broken links or orphaned notes
- Create reusable query views

Don't use for:
- Simple file-based note taking (just use mkdir/touch)
- Binary attachments or non-markdown content
- Real-time collaborative editing
- Version control (use git separately)

## Quick Reference

### Essential Commands

| Task | Command | Example |
|------|---------|---------|
| **Setup** | `opennotes init` | One-time initialization |
| **Create notebook** | `opennotes notebook create --name <name>` | `opennotes notebook create --name "work"` |
| **Add simple note** | `opennotes notes add "<title>"` | `opennotes notes add "Meeting Notes"` |
| **Add with metadata** | `opennotes notes add "<title>" --data k=v` | `opennotes notes add "Sprint" --data tag=urgent` |
| **List notes** | `opennotes notes list` | Shows all notes in current notebook |
| **Search text** | `opennotes notes search "<query>"` | `opennotes notes search "deadline"` |
| **Boolean query** | `opennotes notes search query <field>=<val>` | `opennotes notes search query tag=meeting` |
| **SQL query** | `opennotes notes search --sql "<query>"` | `opennotes notes search --sql "SELECT * FROM ..."` |
| **Execute view** | `opennotes notes view <name>` | `opennotes notes view today` |
| **List views** | `opennotes notes view` | Shows all available views |

### Version Detection

```bash
# Check which version you have installed
which opennotes

# v0.0.2 = Basic features (list, add, search, remove)
# v0.0.3+ = Advanced features (--data flags, views, boolean queries)

# Test for v0.0.3+ features
opennotes notes add --help | grep -q -- '--data' && echo "v0.0.3+" || echo "v0.0.2"
```

## Notebook Management

### Creating Notebooks

```bash
# Create new notebook
opennotes notebook create --name "project-notes"
# Creates: .opennotes.json and .notes/ directory

# Navigate to notebook
cd project-notes

# View notebook info
opennotes notebook
```

### Notebook Auto-Discovery

OpenNotes finds notebooks automatically:

1. Searches current directory for `.opennotes.json`
2. Walks up parent directories until found
3. Falls back to registered notebooks in global config
4. Override with `--notebook <path>` flag

```bash
# Work from anywhere in notebook tree
cd project-notes/2026/january
opennotes notes list  # Auto-discovers notebook

# Use specific notebook
opennotes notes list --notebook ~/work-notes
```

### Managing Multiple Notebooks

```bash
# List all registered notebooks
opennotes notebook list

# Register existing notebook
opennotes notebook register ~/existing-notes

# Add context path (for auto-discovery)
opennotes notebook add-context /path/to/workspace
```

## Note Creation

### Basic Note Creation

```bash
# Simple note (filename auto-slugified)
opennotes notes add "My First Note"
# Creates: my-first-note.md

# Title with special characters
opennotes notes add "2026-01-24: Daily Log"
# Creates: 2026-01-24-daily-log.md

# Using deprecated --title flag (still works)
opennotes notes add --title "Meeting Notes"
```

### With Metadata (v0.0.3+)

```bash
# Single metadata field
opennotes notes add "Team Standup" --data author="John Doe"

# Multiple fields
opennotes notes add "API Review" \
  --data author="Engineering Team" \
  --data priority=high \
  --data status=draft

# Array values (repeated fields)
opennotes notes add "Sprint Planning" \
  --data tag=meeting \
  --data tag=urgent \
  --data tag=sprint

# Result in frontmatter:
# ---
# author: John Doe
# priority: high
# tag:
#   - meeting
#   - urgent
# ---
```

### Path Options

```bash
# Create in subfolder
opennotes notes add "Project Plan" projects/
# Creates: projects/project-plan.md

# Specify exact filename
opennotes notes add "Changelog" CHANGELOG
# Creates: CHANGELOG.md

# Full path with extension
opennotes notes add "README" docs/README.md
# Creates: docs/README.md
```

### Content Priority

Notes are populated in this order:

1. **Stdin** (piped content)
2. **Template** (`--template` flag)
3. **Default** (simple `# Title\n\n`)

```bash
# From stdin
echo "# Meeting Notes\n\n- Point 1" | opennotes notes add "Meeting"

# From template
opennotes notes add "Standup" --template daily-standup

# Default (just title)
opennotes notes add "Quick Note"
```

## Searching Notes

### Text Search (v0.0.2+)

```bash
# Search all fields (content, title, filename)
opennotes notes search "deadline"

# Case-insensitive, partial matches
opennotes notes search "meet"  # Finds "Meeting", "meet", "meetings"

# No results
opennotes notes search "nonexistent"
# Output: No notes found matching 'nonexistent'
```

### Fuzzy Search (v0.0.3+)

```bash
# Typo-tolerant search
opennotes notes search --fuzzy "meetng"  # Finds "meeting"

# Ranked by relevance
opennotes notes search --fuzzy "project"
# Title matches ranked higher than body matches
```

### Boolean Queries (v0.0.3+)

```bash
# Single condition (AND)
opennotes notes search query tag=urgent

# Multiple AND conditions
opennotes notes search query tag=meeting AND status=active

# OR conditions
opennotes notes search query priority=high OR priority=critical

# NOT conditions
opennotes notes search query tag=epic AND NOT status=archived

# Path glob patterns
opennotes notes search query "path=projects/**/*.md"

# Complex queries
opennotes notes search query \
  "tag=workflow AND status=active AND (priority=high OR priority=critical)"
```

**Supported Fields**:
- `tag`, `status`, `priority`, `author` (any metadata field)
- `path` (file path with glob support)
- `title` (note title from frontmatter or filename)
- `links-to` (outgoing links)
- `linked-by` (incoming links)

**Operators**:
- `AND` (all must match)
- `OR` (any must match)
- `NOT` (exclude matches)

### Link Queries (v0.0.3+)

```bash
# Find notes linking to specific note
opennotes notes search query "links-to=README.md"

# With glob patterns
opennotes notes search query "links-to=docs/*.md"

# Find notes linked by specific note
opennotes notes search query "linked-by=index.md"

# Combined with other conditions
opennotes notes search query "links-to=*.md AND tag=reference"
```

### SQL Queries (v0.0.3+)

```bash
# Direct SQL query
opennotes notes search --sql "SELECT * FROM read_markdown('.notes/**/*.md')"

# With WHERE clause
opennotes notes search --sql \
  "SELECT file_path, metadata FROM read_markdown('.notes/**/*.md') 
   WHERE metadata['tag'] = 'urgent'"

# Complex aggregations
opennotes notes search --sql \
  "WITH tagged AS (
     SELECT file_path, metadata['tag'] as tag
     FROM read_markdown('.notes/**/*.md')
   )
   SELECT tag, COUNT(*) as count
   FROM tagged
   GROUP BY tag
   ORDER BY count DESC"
```

**Safety**:
- Only `SELECT` and `WITH` queries allowed
- `INSERT`, `UPDATE`, `DELETE`, `DROP` blocked
- Parameterized queries prevent SQL injection

## Views System (v0.0.3+)

### Built-in Views

```bash
# List all available views
opennotes notes view

# Execute built-in view
opennotes notes view today        # Notes created/updated today
opennotes notes view recent       # Last 30 days
opennotes notes view untagged     # Notes without tags
opennotes notes view orphans      # Notes with no incoming links
opennotes notes view broken-links # Notes with broken links

# Parameterized view
opennotes notes view kanban status=in-progress
```

### Custom Views

Define in notebook's `.opennotes.json` or global `~/.config/opennotes/config.json`:

```json
{
  "views": {
    "active-tasks": {
      "description": "All active tasks",
      "conditions": [
        {"field": "tag", "operator": "=", "value": "task"},
        {"field": "status", "operator": "=", "value": "active"}
      ]
    },
    "recent-meetings": {
      "description": "Meetings in last 7 days",
      "conditions": [
        {"field": "tag", "operator": "=", "value": "meeting"},
        {"field": "created", "operator": ">=", "value": "{{7_days_ago}}"}
      ]
    }
  }
}
```

**View Precedence**:
1. Notebook config (`.opennotes.json`)
2. Global config (`~/.config/opennotes/config.json`)
3. Built-in views

### Template Variables

Use in view definitions:

- `{{today}}` - Today's date (YYYY-MM-DD)
- `{{yesterday}}` - Yesterday's date
- `{{this_week}}` - Start of current week
- `{{this_month}}` - Start of current month
- `{{7_days_ago}}` - Date 7 days ago
- `{{30_days_ago}}` - Date 30 days ago

## Common Workflows

### Daily Note Workflow

```bash
# Create today's note
opennotes notes add "$(date +%Y-%m-%d): Daily Log" \
  --data type=daily \
  --data date="$(date +%Y-%m-%d)"

# View today's notes
opennotes notes view today

# Search recent dailies
opennotes notes search query "type=daily AND created>=2026-01-20"
```

### Project Management

```bash
# Create project epic
opennotes notes add "Q1 2026 Goals" projects/ \
  --data tag=epic \
  --data status=planning \
  --data priority=high

# Create tasks
opennotes notes add "Implement Feature X" tasks/ \
  --data tag=task \
  --data epic=q1-2026-goals \
  --data status=todo

# View active tasks
opennotes notes search query "tag=task AND status=todo"

# Kanban board
opennotes notes view kanban status=in-progress
```

### Knowledge Base

```bash
# Create reference note
opennotes notes add "Git Cheatsheet" reference/ \
  --data tag=reference \
  --data tag=git \
  --data category=tools

# Link analysis
opennotes notes view orphans        # Unconnected notes
opennotes notes view broken-links   # Fix broken links

# Find related notes
opennotes notes search query "tag=git"
```

## Common Mistakes

### ❌ Using --title instead of positional argument

```bash
# Don't (deprecated flag)
opennotes notes add --title "My Note"

# Do (positional argument)
opennotes notes add "My Note"
```

### ❌ Forgetting to quote special characters

```bash
# Don't (shell interprets special chars)
opennotes notes search query tag=urgent AND status=active

# Do (quote the entire query)
opennotes notes search query "tag=urgent AND status=active"
```

### ❌ Using --data on old version

```bash
# Check version first
opennotes notes add --help | grep -q -- '--data' || echo "Upgrade needed"

# On v0.0.2: Edit frontmatter manually after creation
opennotes notes add "My Note"
vim .notes/my-note.md  # Add frontmatter manually
```

### ❌ Assuming case-sensitive search

```bash
# Search is case-INsensitive
opennotes notes search "MEETING"  # Finds "meeting", "Meeting", etc.
```

### ❌ Forgetting notebook auto-discovery

```bash
# Don't (unnecessary --notebook flag)
cd ~/my-notes
opennotes notes list --notebook ~/my-notes

# Do (auto-discovery from current directory)
cd ~/my-notes
opennotes notes list
```

## Troubleshooting

### No notebook found

```bash
# Error: No notebook found
# Solution 1: Create notebook
opennotes notebook create --name "my-notes"

# Solution 2: Register existing
opennotes notebook register ~/existing-notes

# Solution 3: Specify explicitly
opennotes notes list --notebook ~/my-notes
```

### Metadata not showing in frontmatter

```bash
# Check version supports --data
opennotes notes add --help | grep -- '--data'

# Verify format: field=value (no spaces around =)
opennotes notes add "Note" --data "tag = value"  # ❌ Wrong
opennotes notes add "Note" --data "tag=value"    # ✅ Correct

# Check created file
cat .notes/note.md
```

### Search returns no results

```bash
# Verify notebook has notes
opennotes notes list

# Check search is case-insensitive
opennotes notes search "KEYWORD"

# Try broader search
opennotes notes search ""  # List all notes

# Use SQL for debugging
opennotes notes search --sql "SELECT * FROM read_markdown('.notes/**/*.md')"
```

## Configuration Files

### Global Config

`~/.config/opennotes/config.json`:

```json
{
  "notebooks": [
    {
      "name": "work",
      "path": "/home/user/work-notes"
    }
  ],
  "views": {
    "my-custom-view": {
      "description": "Custom view",
      "conditions": []
    }
  }
}
```

### Notebook Config

`.opennotes.json` in notebook root:

```json
{
  "name": "My Notebook",
  "notesDir": ".notes",
  "templatesDir": ".templates",
  "contexts": [
    "/path/to/workspace"
  ],
  "views": {
    "project-active": {
      "description": "Active project tasks",
      "conditions": [
        {"field": "tag", "operator": "=", "value": "project"},
        {"field": "status", "operator": "=", "value": "active"}
      ]
    }
  }
}
```

## Environment Variables

```bash
# Override config path
export OPENNOTES_CONFIG="~/.config/opennotes/custom.json"

# Enable debug logging
export DEBUG=1
export LOG_LEVEL=debug

# In scripts
OPENNOTES_CONFIG=/tmp/test-config.json opennotes notes list
```

## Integration Examples

### With Git

```bash
# Initialize notebook with git
opennotes notebook create --name "wiki"
cd wiki
git init
echo ".DS_Store" > .gitignore
git add .
git commit -m "Initial commit"

# Commit after note creation
opennotes notes add "New Feature Spec" --data status=draft
git add .notes/new-feature-spec.md
git commit -m "Add feature spec"
```

### With Obsidian

OpenNotes notebooks are compatible with Obsidian:

1. Create OpenNotes notebook: `opennotes notebook create --name "obsidian-vault"`
2. Open in Obsidian: Point to notebook directory
3. Use OpenNotes for CLI operations, Obsidian for GUI editing
4. Frontmatter metadata is shared between both tools

### With Scripts

```bash
#!/bin/bash
# daily-note.sh - Create daily note with template

DATE=$(date +%Y-%m-%d)
TITLE="$DATE: Daily Log"

opennotes notes add "$TITLE" \
  --data type=daily \
  --data date="$DATE" \
  --template daily-standup

# Open in editor
$EDITOR .notes/*${DATE}*.md
```

## Performance Characteristics

### Scalability

- **Small notebooks** (< 100 notes): Instant operations
- **Medium notebooks** (100-1000 notes): ~100ms search latency
- **Large notebooks** (1000+ notes): ~1s search latency (1147 ops/sec tested)

### Memory Usage

- **Per note**: ~4KB memory footprint at scale
- **1000 notes**: ~4MB total memory usage
- **Search**: Memory-efficient streaming queries

### Concurrency

- **Thread-safe**: Safe for concurrent operations
- **Connection pooling**: DuckDB handles locking
- **File operations**: Uses atomic writes where possible

## Real-World Impact

### Use Cases

1. **Personal Knowledge Base**: 500+ reference notes, full-text search, link analysis
2. **Project Management**: Task tracking with metadata, kanban views, status queries
3. **Meeting Notes**: Daily logs, searchable archive, template-based creation
4. **Documentation**: Technical specs, API docs, cross-referenced with link queries

### Advantages Over Alternatives

**vs Plain Markdown + grep**:
- Structured metadata queries (not just text search)
- Link analysis and relationship tracking
- Reusable views with parameters
- SQL query power

**vs Obsidian**:
- CLI automation and scripting
- SSH/remote access friendly
- Version control friendly (plain text)
- SQL query interface

**vs Notion/Roam**:
- Local-first, no cloud dependency
- Plain markdown files (portability)
- Git-friendly for version control
- Open source, extensible

## Version Feature Matrix

| Feature | v0.0.2 | v0.0.3+ |
|---------|--------|---------|
| Create notebooks | ✅ | ✅ |
| Add notes | ✅ | ✅ |
| Text search | ✅ | ✅ |
| List notes | ✅ | ✅ |
| Remove notes | ✅ | ✅ |
| Metadata (`--data`) | ❌ | ✅ |
| Path arguments | ❌ | ✅ |
| Fuzzy search | ❌ | ✅ |
| Boolean queries | ❌ | ✅ |
| Link queries | ❌ | ✅ |
| SQL queries | ❌ | ✅ |
| Views system | ❌ | ✅ |
| Template variables | ❌ | ✅ |

## Further Reading

- **Note Creation Guide**: `docs/note-creation-guide.md` (471 lines)
- **Search Guide**: `docs/commands/notes-search.md` (407 lines)
- **Views Guide**: `docs/views-guide.md` (880 lines)
- **Views API**: `docs/views-api.md` (805 lines)
- **Views Examples**: `docs/views-examples.md` (812 lines)
