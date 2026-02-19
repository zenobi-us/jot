# Note Creation Guide

Quick reference for creating notes with Jot, including metadata support and flexible path handling.

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
jot notes add "My First Note"

# Note with metadata
jot notes add "Meeting Notes" --data tag=meeting --data priority=high

# Note in specific folder
jot notes add "Retrospective" meetings/

# Pipe content from stdin
echo "# Quick Note\n\nSome content" | jot notes add "Quick Note"
```

## Basic Usage

### Syntax

```bash
jot notes add <title> [path] [flags]
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
jot notes add "Daily Log"

# Create with specific title
jot notes add "Project Brainstorm 2026-01-24"

# Use a template
jot notes add "Weekly Review" --template weekly
```

## Adding Metadata

The `--data` flag adds frontmatter fields to your note. It's **repeatable** - use it multiple times to add multiple fields.

### Single Values

```bash
# Add single metadata fields
jot notes add "Sprint Planning" \
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

When you repeat the same field name, Jot automatically creates an array:

```bash
# Tags as array
jot notes add "Team Standup" \
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
jot notes add "API Design Review" \
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

The `[path]` argument is flexible - Jot auto-detects whether you mean a file or folder.

### No Path (Default)

```bash
jot notes add "Quick Idea"
# Creates: quick-idea.md (in notebook root)
```

### Folder Path

```bash
jot notes add "Meeting Notes" meetings/
# Creates: meetings/meeting-notes.md

jot notes add "Retrospective" projects/sprint-5/
# Creates: projects/sprint-5/retrospective.md
```

### File Path (with or without .md)

```bash
jot notes add "Status Update" reports/weekly.md
# Creates: reports/weekly.md

jot notes add "Changelog" CHANGELOG
# Creates: CHANGELOG.md (auto-adds .md extension)
```

### Auto-Detection Rules

1. **Ends with `/`** → Folder path (filename = slugified title)
2. **Ends with `.md`** → Exact file path
3. **No extension** → Adds `.md` extension automatically
4. **No path** → Root of notebook

## Content Sources

Jot prioritizes content from multiple sources:

**Priority Order** (highest to lowest):

1. **Stdin** - Piped or redirected content
2. **Template** - Content from `--template` flag
3. **Default** - Simple `# Title\n\n`

### Stdin Content (Highest Priority)

```bash
# Pipe from echo
echo -e "# Daily Log\n\n- Completed feature X\n- Started feature Y" | \
  jot notes add "Daily Log 2026-01-24"

# Redirect from file
jot notes add "Import" < existing-note.md

# Pipe from command
curl https://example.com/api/doc | jot notes add "API Documentation"
```

### Template Content

```bash
# Use predefined template
jot notes add "Weekly Review" --template weekly

# Template with metadata
jot notes add "Meeting" --template meeting \
  --data attendees="Alice,Bob" \
  --data date=2026-01-24
```

### Default Content

If no stdin or template is provided, Jot creates a simple note:

```markdown
# Your Title Here
```

## Real-World Examples

### Daily Workflow

```bash
# Morning standup
jot notes add "Standup $(date +%Y-%m-%d)" \
  --data tag=standup --data tag=daily \
  standups/

# Quick task capture
echo "- [ ] Review PR #123\n- [ ] Update docs" | \
  jot notes add "Tasks $(date +%Y-%m-%d)"

# End of day notes
jot notes add "EOD $(date +%Y-%m-%d)" \
  --data completed=5 --data in-progress=2 \
  daily/
```

### Project Management

```bash
# Sprint planning with metadata
jot notes add "Sprint 5 Planning" \
  --data sprint=5 \
  --data start-date=2026-01-27 \
  --data end-date=2026-02-07 \
  --data tag=planning --data tag=sprint \
  sprints/

# Feature specification
jot notes add "User Authentication" \
  --data status=draft \
  --data priority=high \
  --data tag=feature --data tag=security \
  specs/features/
```

### Meeting Notes

```bash
# Team meeting with attendees
jot notes add "Team Sync 2026-01-24" \
  --data type=team-sync \
  --data attendees="Alice,Bob,Charlie" \
  --data tag=meeting \
  meetings/team/

# Client meeting
jot notes add "Client Review" \
  --data client="Acme Corp" \
  --data status=completed \
  --data tag=client --data tag=review \
  meetings/clients/
```

### Knowledge Base

```bash
# Technical documentation
jot notes add "Database Migration Guide" \
  --data category=infrastructure \
  --data tag=database --data tag=migration --data tag=guide \
  docs/technical/

# How-to guide
jot notes add "Deploy to Production" \
  --data type=howto \
  --data difficulty=advanced \
  --data tag=deployment --data tag=production \
  docs/howto/
```

### Automation & Scripting

```bash
# Generate weekly template
for day in {Mon,Tue,Wed,Thu,Fri}; do
  jot notes add "Standup - $day Week 5" \
    --data day=$day --data week=5 --data tag=standup \
    standups/2026-01/
done

# Import from external source
curl -s https://api.github.com/repos/user/repo/issues/1 | \
  jq -r '.body' | \
  jot notes add "Issue #1 - Bug Report" \
    --data source=github --data type=issue
```

## Migration from Old Syntax

The old `--title` flag is deprecated but still works for backward compatibility.

### Old Syntax (Deprecated)

```bash
# ⚠️ Works but deprecated
jot notes add --title "My Note"
```

**Warning shown**:

```
Warning: --title flag is deprecated, use positional argument instead
Example: jot notes add "My Note"
```

### New Syntax (Recommended)

```bash
# ✅ New way
jot notes add "My Note"

# ✅ With path
jot notes add "My Note" path/to/location/

# ✅ With metadata
jot notes add "My Note" --data tag=important
```

### Migration Checklist

- [ ] Replace `--title "text"` with positional `"text"`
- [ ] Add `--data` flags for any metadata you need
- [ ] Use path argument for note location (optional)
- [ ] Update any scripts or aliases

**Migration Script Example**:

```bash
# Old command in scripts
jot notes add --title "Daily Log"

# New command
jot notes add "Daily Log"
```

## Tips & Best Practices

### Consistent Metadata

Use consistent field names across your notes for better querying:

```bash
# Good - consistent tags
jot notes add "Note 1" --data tag=meeting --data tag=daily
jot notes add "Note 2" --data tag=meeting --data tag=weekly

# Avoid - inconsistent naming
jot notes add "Note 1" --data tags=meeting  # plural
jot notes add "Note 2" --data tag=meeting   # singular
```

### Title Slugification

Titles are automatically slugified for filenames:

```bash
jot notes add "My Great Idea!"
# Creates: my-great-idea.md

jot notes add "2026-01-24: Daily Log"
# Creates: 2026-01-24-daily-log.md
```

### Date-Based Organization

```bash
# Organize by year/month/day
jot notes add "Standup" \
  $(date +%Y/%m/%d)/

# Creates: 2026/01/24/standup.md
```

### Metadata for Views

Add metadata that works well with the Views System:

```bash
# For kanban view
jot notes add "Feature X" \
  --data status=in-progress \
  --data tag=feature

# For filtering
jot notes add "Bug Report" \
  --data priority=high \
  --data assignee=alice \
  --data tag=bug
```

## See Also

- **[Views Guide](views-guide.md)** - Query notes with reusable views
- **[Search Command Reference](commands/notes-search.md)** - Advanced search workflows
- **[Import Workflow](import-workflow-guide.md)** - Import existing notes

## Troubleshooting

### Path Not Created

**Problem**: Folder path doesn't exist

```bash
jot notes add "Note" non-existent-folder/
# Error: folder not found
```

**Solution**: Create folder first or use existing path

```bash
mkdir -p non-existent-folder
jot notes add "Note" non-existent-folder/
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
echo "content" | jot notes add "Note" --template weekly

# ✅ Stdin content used
echo "content" | jot notes add "Note"
```

---

**Last Updated**: 2026-01-24  
**Version**: Jot 0.0.4+
