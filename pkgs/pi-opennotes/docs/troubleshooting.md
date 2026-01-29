# Troubleshooting Guide

Common issues and solutions for `@zenobi-us/pi-opennotes`.

## Table of Contents

- [Installation Issues](#installation-issues)
- [CLI Integration](#cli-integration)
- [Notebook Configuration](#notebook-configuration)
- [Query Errors](#query-errors)
- [Performance Issues](#performance-issues)
- [Pi Integration](#pi-integration)

---

## Installation Issues

### Extension Not Loading

**Symptom**: Tools not available in pi

**Check**:
```bash
# Verify extension installed
npm list -g @zenobi-us/pi-opennotes

# Check pi settings
cat ~/.pi/settings.json | grep opennotes
```

**Solutions**:

1. **Not in settings**:
   ```json
   {
     "packages": ["npm:@zenobi-us/pi-opennotes"]
   }
   ```

2. **Restart pi**:
   ```bash
   # If pi is running
   /reload
   ```

3. **Reinstall**:
   ```bash
   npm uninstall -g @zenobi-us/pi-opennotes
   npm install -g @zenobi-us/pi-opennotes
   ```

### Bun vs NPM Conflicts

**Symptom**: Installation fails with dependency errors

**Solution**:

Use consistent package manager:

```bash
# If using bun
bun install -g @zenobi-us/pi-opennotes

# If using npm
npm install -g @zenobi-us/pi-opennotes
```

Don't mix `bun` and `npm` installations.

### Permission Errors on Linux

**Symptom**: `EACCES` during installation

**Solution**:

```bash
# Option 1: Use sudo (not recommended)
sudo npm install -g @zenobi-us/pi-opennotes

# Option 2: Configure npm prefix (recommended)
mkdir ~/.npm-global
npm config set prefix '~/.npm-global'
echo 'export PATH=~/.npm-global/bin:$PATH' >> ~/.bashrc
source ~/.bashrc

# Then install without sudo
npm install -g @zenobi-us/pi-opennotes
```

---

## CLI Integration

### CLI Not Found

**Error**: `OPENNOTES_CLI_NOT_FOUND`

**Check**:
```bash
# Is OpenNotes installed?
which opennotes
opennotes version
```

**Solutions**:

1. **Install OpenNotes**:
   ```bash
   go install github.com/zenobi-us/opennotes@latest
   ```

2. **Check PATH**:
   ```bash
   echo $PATH | grep go/bin
   # Should show ~/go/bin or similar
   ```

3. **Add to PATH** (if missing):
   ```bash
   echo 'export PATH=$PATH:~/go/bin' >> ~/.bashrc
   source ~/.bashrc
   ```

4. **Set explicit path** in pi config:
   ```json
   {
     "config": {
       "@zenobi-us/pi-opennotes": {
         "cliPath": "/home/user/go/bin/opennotes"
       }
     }
   }
   ```

### Wrong CLI Version

**Error**: Tools fail with unexpected output

**Check**:
```bash
opennotes version
# Should be >= 0.0.2
```

**Solution**:
```bash
# Update OpenNotes
go install github.com/zenobi-us/opennotes@latest

# Verify
opennotes version
```

### CLI Timeout

**Error**: `Command timeout after 30000ms`

**Causes**:
- Very large notebooks (1000+ notes)
- Complex SQL queries
- Slow filesystem (network drives)

**Solutions**:

1. **Increase timeout**:
   ```json
   {
     "config": {
       "@zenobi-us/pi-opennotes": {
         "cliTimeout": 60000
       }
     }
   }
   ```

2. **Simplify query**:
   ```typescript
   // Instead of
   { sql: "SELECT * FROM notes" }  // All notes

   // Use
   { sql: "SELECT * FROM notes LIMIT 100" }  // Limited
   ```

3. **Add indexes** (future feature):
   ```sql
   CREATE INDEX idx_status ON notes ((data->>'status'));
   ```

### CLI Crashes

**Symptom**: CLI exits with non-zero code

**Debug**:
```bash
# Run CLI directly to see error
opennotes --notebook ~/notes notes list --output json

# Check for
# - Permission errors
# - Corrupted .opennotes.json
# - Filesystem issues
```

**Solutions**:

1. **Fix permissions**:
   ```bash
   chmod -R u+rw ~/notes
   ```

2. **Validate config**:
   ```bash
   cat ~/notes/.opennotes.json | jq .
   # Should parse without error
   ```

3. **Check disk space**:
   ```bash
   df -h
   ```

---

## Notebook Configuration

### Notebook Not Found

**Error**: `Notebook not found: /path`

**Check**:
```bash
# Path exists?
ls -la /path

# Has .opennotes.json?
ls -la /path/.opennotes.json
```

**Solutions**:

1. **Initialize notebook**:
   ```bash
   cd /path
   opennotes init
   ```

2. **Fix path** in tool call:
   ```typescript
   {
     "query": "test",
     "notebook": "/correct/path"
   }
   ```

3. **Set default notebook**:
   ```bash
   opennotes config set default.notebook ~/notes/main
   ```

### Invalid Config

**Error**: `Failed to parse .opennotes.json`

**Check**:
```bash
cat ~/notes/.opennotes.json | jq .
```

**Common issues**:

1. **Trailing commas**:
   ```json
   {
     "name": "Notes",
     "views": {},  // ❌ Remove trailing comma
   }
   ```

2. **Missing quotes**:
   ```json
   {
     name: "Notes"  // ❌ Should be "name"
   }
   ```

3. **Invalid SQL**:
   ```json
   {
     "views": {
       "my-view": {
         "sql": "SELECT * FROM WHERE"  // ❌ Invalid
       }
     }
   }
   ```

**Solution**:

1. **Validate JSON**:
   ```bash
   cat .opennotes.json | jq .
   ```

2. **Fix syntax** - use proper JSON

3. **Restore from backup** if corrupted:
   ```bash
   cp .opennotes.json.bak .opennotes.json
   ```

### Missing Views

**Error**: `View not found: view-name`

**Check**:
```bash
opennotes --notebook ~/notes notebooks info
# Lists all views
```

**Solutions**:

1. **Add view** to `.opennotes.json`:
   ```json
   {
     "views": {
       "view-name": {
         "description": "My view",
         "sql": "SELECT * FROM notes LIMIT 10"
       }
     }
   }
   ```

2. **Use built-in view**:
   ```typescript
   { view: "today" }  // Built-in
   { view: "recent" }
   { view: "kanban" }
   ```

3. **List available views**:
   ```typescript
   await opennotes_views({})  // Lists all
   ```

---

## Query Errors

### SQL Syntax Error

**Error**: `SQL syntax error near ...`

**Common mistakes**:

1. **Invalid column**:
   ```sql
   SELECT invalid_column FROM notes  -- ❌
   SELECT path, title FROM notes     -- ✅
   ```

2. **Missing quotes**:
   ```sql
   SELECT * FROM notes WHERE data->>'status' = active  -- ❌
   SELECT * FROM notes WHERE data->>'status' = 'active' -- ✅
   ```

3. **JSON access**:
   ```sql
   SELECT data.status FROM notes           -- ❌
   SELECT data->>'status' FROM notes       -- ✅
   ```

**Solution**: Check [SQL reference](./tool-usage-guide.md#sql-query)

### Forbidden SQL

**Error**: `Only SELECT and WITH allowed`

**Why**: Security - prevent destructive operations

**Forbidden**:
```sql
DROP TABLE notes      -- ❌
DELETE FROM notes     -- ❌
UPDATE notes SET ...  -- ❌
INSERT INTO notes ... -- ❌
```

**Allowed**:
```sql
SELECT * FROM notes                          -- ✅
WITH active AS (...) SELECT * FROM active    -- ✅
```

**Solution**: Use SELECT for queries, CLI commands for mutations

### Path Traversal Blocked

**Error**: `Invalid path: contains ..`

**Why**: Security - prevent accessing files outside notebook

**Invalid**:
```typescript
{ path: "../../etc/passwd" }  // ❌
{ path: "/etc/passwd" }        // ❌
```

**Valid**:
```typescript
{ path: "projects/alpha.md" }  // ✅
{ path: "tasks/task-001.md" }  // ✅
```

**Solution**: Use paths relative to notebook root

### Empty Results

**Issue**: Query returns no results unexpectedly

**Debug**:

1. **Check query**:
   ```typescript
   // Too specific?
   { sql: "SELECT * FROM notes WHERE data->>'exact-match' = 'value'" }
   
   // Try broader
   { sql: "SELECT * FROM notes WHERE data->>'field' LIKE '%value%'" }
   ```

2. **Verify data**:
   ```bash
   opennotes --notebook ~/notes notes get path/to/note.md
   # Check frontmatter has expected fields
   ```

3. **Test simpler query**:
   ```typescript
   { sql: "SELECT COUNT(*) FROM notes" }
   // Should return total count
   ```

4. **Check filters**:
   ```typescript
   {
     "filters": {
       "and": ["data.status=active"]  // Status might not exist
     }
   }
   ```

---

## Performance Issues

### Slow Queries

**Symptom**: Queries take >5 seconds

**Causes**:
- Large notebooks (>1000 notes)
- Complex SQL with joins
- Full-text search on entire content
- No LIMIT clause

**Solutions**:

1. **Add LIMIT**:
   ```sql
   SELECT * FROM notes LIMIT 50  -- Don't fetch all
   ```

2. **Specific WHERE**:
   ```sql
   SELECT * FROM notes 
   WHERE data->>'status' = 'active'  -- Filter early
   LIMIT 50
   ```

3. **Metadata only**:
   ```typescript
   { path: "note.md", includeContent: false }
   ```

4. **Paginate**:
   ```typescript
   { query: "search", limit: 50, offset: 0 }
   ```

5. **Use views** (cached):
   ```typescript
   { view: "active-tasks" }  // Faster than ad-hoc SQL
   ```

### High Memory Usage

**Symptom**: Pi process using >500MB

**Causes**:
- Fetching too many notes at once
- Large note contents
- Not using pagination

**Solutions**:

1. **Reduce page size**:
   ```json
   {
     "config": {
       "@zenobi-us/pi-opennotes": {
         "defaultPageSize": 20  // Smaller pages
       }
     }
   }
   ```

2. **Stream results** (future feature)

3. **Clear cache** (if implemented):
   ```bash
   opennotes cache clear
   ```

### Slow Fuzzy Search

**Symptom**: Fuzzy search takes >10 seconds

**Why**: Fuzzy matching is computationally expensive

**Solutions**:

1. **Use exact search** when possible:
   ```typescript
   { query: "exact term" }  // Fast
   // vs
   { query: "aproximate term", fuzzy: true }  // Slow
   ```

2. **Reduce scope**:
   ```typescript
   {
     "query": "term",
     "fuzzy": true,
     "pattern": "tasks/*.md",  // Only search tasks/
     "limit": 20
   }
   ```

3. **Use SQL** if field known:
   ```typescript
   {
     "sql": "SELECT * FROM notes WHERE data->>'title' LIKE '%term%'"
   }
   ```

---

## Pi Integration

### Tools Not Registered

**Symptom**: Pi doesn't recognize `opennotes_*` tools

**Check**:
```
/tools list
# Should show opennotes_search, opennotes_list, etc.
```

**Solutions**:

1. **Restart pi**:
   ```
   /reload
   ```

2. **Check settings**:
   ```bash
   cat ~/.pi/settings.json
   ```

3. **Extension loaded**?:
   ```
   /extensions list
   ```

### Wrong Tool Prefix

**Symptom**: Tools have different names

**Cause**: Custom prefix configured

**Check**:
```json
{
  "config": {
    "@zenobi-us/pi-opennotes": {
      "toolPrefix": "notes_"  // Custom prefix
    }
  }
}
```

**Solution**: Either:

1. Use configured prefix: `notes_search`
2. Or reset to default: `"toolPrefix": "opennotes_"`

### Pi Errors

**Error**: `Tool execution failed`

**Debug**:

1. **Check pi logs**:
   ```bash
   tail -f ~/.pi/logs/error.log
   ```

2. **Test tool directly**:
   ```
   /call opennotes_search {"query": "test"}
   ```

3. **Verify CLI works**:
   ```bash
   opennotes notes list
   ```

### Context Budget Exceeded

**Error**: `Output too large for context`

**Cause**: Too many results returned

**Solutions**:

1. **Reduce page size**:
   ```typescript
   { query: "search", limit: 20 }  // Fewer results
   ```

2. **Metadata only**:
   ```typescript
   { includeContent: false }  // No full content
   ```

3. **Specific fields**:
   ```sql
   SELECT path, title FROM notes  -- Not all fields
   ```

---

## General Debugging

### Enable Debug Logging

1. **Pi debug mode**:
   ```bash
   PI_DEBUG=1 pi
   ```

2. **OpenNotes verbose**:
   ```bash
   opennotes --verbose notes list
   ```

3. **Extension logs**:
   ```bash
   tail -f ~/.pi/extensions/@zenobi-us/pi-opennotes/debug.log
   ```

### Verify Setup

Complete checklist:

```bash
# 1. OpenNotes installed
opennotes version

# 2. Extension installed
npm list -g @zenobi-us/pi-opennotes

# 3. In pi settings
cat ~/.pi/settings.json | grep opennotes

# 4. Notebook initialized
ls ~/.config/opennotes/config.json
opennotes notebooks list

# 5. Tools registered
pi -c "/tools list" | grep opennotes
```

### Reset to Defaults

If all else fails:

```bash
# 1. Uninstall extension
npm uninstall -g @zenobi-us/pi-opennotes

# 2. Remove config
rm -rf ~/.pi/settings.json.opennotes.bak
# (Backup first!)

# 3. Reinstall
npm install -g @zenobi-us/pi-opennotes

# 4. Restart pi
/reload
```

---

## Common Patterns

### Test CLI Directly

Before using tools, verify CLI works:

```bash
# List notes
opennotes --notebook ~/notes notes list

# Search
opennotes --notebook ~/notes notes search --query "test"

# SQL
opennotes --notebook ~/notes notes sql --query "SELECT COUNT(*) FROM notes"

# Get note
opennotes --notebook ~/notes notes get path/to/note.md
```

### Validate JSON

All configs are JSON - validate before use:

```bash
cat .opennotes.json | jq .
cat ~/.pi/settings.json | jq .
```

### Check Permissions

Ensure pi can access notebooks:

```bash
# Read permissions
ls -la ~/notes/.opennotes.json

# Write permissions (for create)
touch ~/notes/test-write.md && rm ~/notes/test-write.md
```

---

## Getting Help

### Where to Report Issues

- **OpenNotes CLI bugs**: [opennotes/issues](https://github.com/zenobi-us/opennotes/issues)
- **Extension bugs**: [pi-opennotes/issues](https://github.com/zenobi-us/pi-opennotes/issues)
- **Pi integration**: [pi/discussions](https://github.com/mariozechner/pi-coding-agent/discussions)

### Include in Bug Reports

```
**Environment**:
- OS: Linux/macOS/Windows
- OpenNotes version: `opennotes version`
- Extension version: `npm list -g @zenobi-us/pi-opennotes`
- Pi version: `pi --version`

**Issue**:
[Description]

**Steps to reproduce**:
1. [Step 1]
2. [Step 2]

**Expected**: [What should happen]
**Actual**: [What actually happens]

**Logs**:
```
[Paste relevant logs]
```

### Before Reporting

1. Check this troubleshooting guide
2. Search existing issues
3. Test CLI directly
4. Verify setup checklist
5. Try with minimal config

---

## See Also

- [Integration Guide](./integration-guide.md) - Setup instructions
- [Tool Usage Guide](./tool-usage-guide.md) - Tool documentation
- [Configuration Reference](./configuration.md) - All config options
