# Note Creation Guide

Quick reference for creating notes with OpenNotes, including metadata support and flexible path handling.

## Table of Contents

- [Quick Start](#quick-start)
- [Basic Usage](#basic-usage)
- [Adding Metadata](#adding-metadata)
- [Path Options](#path-options)
- [Content Sources](#content-sources)
- [Real-World Examples](#real-world-examples)
- [Migration from Old Syntax](#migration-from-old-syntax)

## Quick Start

```bash
# Simple note in notebook root
opennotes notes add "My First Note"

# Note with metadata
opennotes notes add "Meeting Notes" --data tag=meeting --data priority=high

# Note in specific folder
opennotes notes add "Retrospective" meetings/

# Pipe content from stdin
echo "# Quick Note\n\nSome content" | opennotes notes add "Quick Note"
```

## Basic Usage

### Syntax

```bash
opennotes notes add <title> [path] [flags]
```

**Arguments**:

- `<title>` - Note title (required) - will be slugified for filename
- `[path]` - Optional path (file or folder) - defaults to notebook root

**Common Flags**:

- `--data field=value` - Add metadata to frontmatter (repeatable)
- `--template <name>` - Use a specific template
- `--notebook <name>` - Target specific notebook

### Simple Examples

```bash
# Create note in root with default content
opennotes notes add "Daily Log"

# Create with specific title
opennotes notes add "Project Brainstorm 2026-01-24"

# Use a template
opennotes notes add "Weekly Review" --template weekly
```

## Adding Metadata

The `--data` flag adds frontmatter fields to your note. It's **repeatable** - use it multiple times to add multiple fields.

### Single Values

```bash
# Add single metadata fields
opennotes notes add "Sprint Planning" \
  --data author="John Doe" \
  --data status=draft \
  --data priority=high
```

**Result** (`sprint-planning.md`):

```markdown
---
author: John Doe
status: draft
priority: high
---

# Sprint Planning
```

### Array Values (Repeated Fields)

When you repeat the same field name, OpenNotes automatically creates an array:

```bash
# Tags as array
opennotes notes add "Team Standup" \
  --data tag=meeting \
  --data tag=daily \
  --data tag=team
```

**Result** (`team-standup.md`):

```markdown
---
tag:
  - meeting
  - daily
  - team
---

# Team Standup
```

### Mixed Metadata

```bash
opennotes notes add "API Design Review" \
  --data author="Engineering Team" \
  --data status=in-review \
  --data tag=architecture \
  --data tag=api \
  --data tag=design \
  --data reviewers="Alice,Bob,Charlie"
```

**Result**:

```markdown
---
author: Engineering Team
status: in-review
tag:
  - architecture
  - api
  - design
reviewers: Alice,Bob,Charlie
---

# API Design Review
```

## Path Options

The `[path]` argument is flexible - OpenNotes auto-detects whether you mean a file or folder.

### No Path (Default)

```bash
opennotes notes add "Quick Idea"
# Creates: quick-idea.md (in notebook root)
```

### Folder Path

```bash
opennotes notes add "Meeting Notes" meetings/
# Creates: meetings/meeting-notes.md

opennotes notes add "Retrospective" projects/sprint-5/
# Creates: projects/sprint-5/retrospective.md
```

### File Path (with or without .md)

```bash
opennotes notes add "Status Update" reports/weekly.md
# Creates: reports/weekly.md

opennotes notes add "Changelog" CHANGELOG
# Creates: CHANGELOG.md (auto-adds .md extension)
```

### Auto-Detection Rules

1. **Ends with `/`** → Folder path (filename = slugified title)
2. **Ends with `.md`** → Exact file path
3. **No extension** → Adds `.md` extension automatically
4. **No path** → Root of notebook

## Content Sources

OpenNotes prioritizes content from multiple sources:

**Priority Order** (highest to lowest):

1. **Stdin** - Piped or redirected content
2. **Template** - Content from `--template` flag
3. **Default** - Simple `# Title\n\n`

### Stdin Content (Highest Priority)

```bash
# Pipe from echo
echo -e "# Daily Log\n\n- Completed feature X\n- Started feature Y" | \
  opennotes notes add "Daily Log 2026-01-24"

# Redirect from file
opennotes notes add "Import" < existing-note.md

# Pipe from command
curl https://example.com/api/doc | opennotes notes add "API Documentation"
```

### Template Content

```bash
# Use predefined template
opennotes notes add "Weekly Review" --template weekly

# Template with metadata
opennotes notes add "Meeting" --template meeting \
  --data attendees="Alice,Bob" \
  --data date=2026-01-24
```

### Default Content

If no stdin or template is provided, OpenNotes creates a simple note:

```markdown
# Your Title Here
```

## Real-World Examples

### Daily Workflow

```bash
# Morning standup
opennotes notes add "Standup $(date +%Y-%m-%d)" \
  --data tag=standup --data tag=daily \
  standups/

# Quick task capture
echo "- [ ] Review PR #123\n- [ ] Update docs" | \
  opennotes notes add "Tasks $(date +%Y-%m-%d)"

# End of day notes
opennotes notes add "EOD $(date +%Y-%m-%d)" \
  --data completed=5 --data in-progress=2 \
  daily/
```

### Project Management

```bash
# Sprint planning with metadata
opennotes notes add "Sprint 5 Planning" \
  --data sprint=5 \
  --data start-date=2026-01-27 \
  --data end-date=2026-02-07 \
  --data tag=planning --data tag=sprint \
  sprints/

# Feature specification
opennotes notes add "User Authentication" \
  --data status=draft \
  --data priority=high \
  --data tag=feature --data tag=security \
  specs/features/
```

### Meeting Notes

```bash
# Team meeting with attendees
opennotes notes add "Team Sync 2026-01-24" \
  --data type=team-sync \
  --data attendees="Alice,Bob,Charlie" \
  --data tag=meeting \
  meetings/team/

# Client meeting
opennotes notes add "Client Review" \
  --data client="Acme Corp" \
  --data status=completed \
  --data tag=client --data tag=review \
  meetings/clients/
```

### Knowledge Base

```bash
# Technical documentation
opennotes notes add "Database Migration Guide" \
  --data category=infrastructure \
  --data tag=database --data tag=migration --data tag=guide \
  docs/technical/

# How-to guide
opennotes notes add "Deploy to Production" \
  --data type=howto \
  --data difficulty=advanced \
  --data tag=deployment --data tag=production \
  docs/howto/
```

### Automation & Scripting

```bash
# Generate weekly template
for day in {Mon,Tue,Wed,Thu,Fri}; do
  opennotes notes add "Standup - $day Week 5" \
    --data day=$day --data week=5 --data tag=standup \
    standups/2026-01/
done

# Import from external source
curl -s https://api.github.com/repos/user/repo/issues/1 | \
  jq -r '.body' | \
  opennotes notes add "Issue #1 - Bug Report" \
    --data source=github --data type=issue
```

## Migration from Old Syntax

The old `--title` flag is deprecated but still works for backward compatibility.

### Old Syntax (Deprecated)

```bash
# ⚠️ Works but deprecated
opennotes notes add --title "My Note"
```

**Warning shown**:

```
Warning: --title flag is deprecated, use positional argument instead
Example: opennotes notes add "My Note"
```

### New Syntax (Recommended)

```bash
# ✅ New way
opennotes notes add "My Note"

# ✅ With path
opennotes notes add "My Note" path/to/location/

# ✅ With metadata
opennotes notes add "My Note" --data tag=important
```

### Migration Checklist

- [ ] Replace `--title "text"` with positional `"text"`
- [ ] Add `--data` flags for any metadata you need
- [ ] Use path argument for note location (optional)
- [ ] Update any scripts or aliases

**Migration Script Example**:

```bash
# Old command in scripts
opennotes notes add --title "Daily Log"

# New command
opennotes notes add "Daily Log"
```

## Tips & Best Practices

### Consistent Metadata

Use consistent field names across your notes for better querying:

```bash
# Good - consistent tags
opennotes notes add "Note 1" --data tag=meeting --data tag=daily
opennotes notes add "Note 2" --data tag=meeting --data tag=weekly

# Avoid - inconsistent naming
opennotes notes add "Note 1" --data tags=meeting  # plural
opennotes notes add "Note 2" --data tag=meeting   # singular
```

### Title Slugification

Titles are automatically slugified for filenames:

```bash
opennotes notes add "My Great Idea!"
# Creates: my-great-idea.md

opennotes notes add "2026-01-24: Daily Log"
# Creates: 2026-01-24-daily-log.md
```

### Date-Based Organization

```bash
# Organize by year/month/day
opennotes notes add "Standup" \
  $(date +%Y/%m/%d)/

# Creates: 2026/01/24/standup.md
```

### Metadata for Views

Add metadata that works well with the Views System:

```bash
# For kanban view
opennotes notes add "Feature X" \
  --data status=in-progress \
  --data tag=feature

# For filtering
opennotes notes add "Bug Report" \
  --data priority=high \
  --data assignee=alice \
  --data tag=bug
```

## See Also

- **[Views Guide](views-guide.md)** - Query notes with custom views
- **[SQL Guide](sql-guide.md)** - Advanced querying with SQL
- **[Import Workflow](import-workflow-guide.md)** - Import existing notes

## Troubleshooting

### Path Not Created

**Problem**: Folder path doesn't exist

```bash
opennotes notes add "Note" non-existent-folder/
# Error: folder not found
```

**Solution**: Create folder first or use existing path

```bash
mkdir -p non-existent-folder
opennotes notes add "Note" non-existent-folder/
```

### Metadata Not Appearing

**Problem**: Metadata not showing in frontmatter

**Check**:

1. Using `--data field=value` format (no spaces around `=`)
2. Field names are valid YAML keys (no special characters)
3. File was created successfully

### Stdin Content Not Used

**Problem**: Piped content ignored

**Cause**: Template might be overriding stdin

**Solution**: Don't use `--template` when piping content:

```bash
# ❌ Template overrides stdin
echo "content" | opennotes notes add "Note" --template weekly

# ✅ Stdin content used
echo "content" | opennotes notes add "Note"
```

---

**Last Updated**: 2026-01-24  
**Version**: OpenNotes 0.0.4+
