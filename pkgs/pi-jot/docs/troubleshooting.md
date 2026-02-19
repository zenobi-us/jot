# Troubleshooting Guide

Common issues and solutions for `@zenobi-us/pi-jot`.

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
npm list -g @zenobi-us/pi-jot

# Check pi settings
cat ~/.pi/settings.json | grep jot
```

**Solutions**:

1. **Not in settings**:
   ```json
   {
     "packages": ["npm:@zenobi-us/pi-jot"]
   }
   ```

2. **Restart pi**:
   ```bash
   # If pi is running
   /reload
   ```

3. **Reinstall**:
   ```bash
   npm uninstall -g @zenobi-us/pi-jot
   npm install -g @zenobi-us/pi-jot
   ```

### Bun vs NPM Conflicts

**Symptom**: Installation fails with dependency errors

**Solution**:

Use consistent package manager:

```bash
# If using bun
bun install -g @zenobi-us/pi-jot

# If using npm
npm install -g @zenobi-us/pi-jot
```

Don't mix `bun` and `npm` installations.

### Permission Errors on Linux

**Symptom**: `EACCES` during installation

**Solution**:

```bash
# Option 1: Use sudo (not recommended)
sudo npm install -g @zenobi-us/pi-jot

# Option 2: Configure npm prefix (recommended)
mkdir ~/.npm-global
npm config set prefix '~/.npm-global'
echo 'export PATH=~/.npm-global/bin:$PATH' >> ~/.bashrc
source ~/.bashrc

# Then install without sudo
npm install -g @zenobi-us/pi-jot
```

---

## CLI Integration

### CLI Not Found

**Error**: `JOT_CLI_NOT_FOUND`

**Check**:
```bash
# Is Jot installed?
which jot
jot version
```

**Solutions**:

1. **Install Jot**:
   ```bash
   go install github.com/zenobi-us/jot@latest
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
       "@zenobi-us/pi-jot": {
         "cliPath": "/home/user/go/bin/jot"
       }
     }
   }
   ```

### Wrong CLI Version

**Error**: Tools fail with unexpected output

**Check**:
```bash
jot version
# Should be >= 0.0.2
```

**Solution**:
```bash
# Update Jot
go install github.com/zenobi-us/jot@latest

# Verify
jot version
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
       "@zenobi-us/pi-jot": {
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
jot --notebook ~/notes notes list --output json

# Check for
# - Permission errors
# - Corrupted .jot.json
# - Filesystem issues
```

**Solutions**:

1. **Fix permissions**:
   ```bash
   chmod -R u+rw ~/notes
   ```

2. **Validate config**:
   ```bash
   cat ~/notes/.jot.json | jq .
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

# Has .jot.json?
ls -la /path/.jot.json
```

**Solutions**:

1. **Initialize notebook**:
   ```bash
   cd /path
   jot init
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
   jot config set default.notebook ~/notes/main
   ```

### Invalid Config

**Error**: `Failed to parse .jot.json`

**Check**:
```bash
cat ~/notes/.jot.json | jq .
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
   cat .jot.json | jq .
   ```

2. **Fix syntax** - use proper JSON

3. **Restore from backup** if corrupted:
   ```bash
   cp .jot.json.bak .jot.json
   ```

### Missing Views

**Error**: `View not found: view-name`

**Check**:
```bash
jot --notebook ~/notes notebooks info
# Lists all views
```

**Solutions**:

1. **Add view** to `.jot.json`:
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
   await jot_views({})  // Lists all
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
   jot --notebook ~/notes notes get path/to/note.md
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
       "@zenobi-us/pi-jot": {
         "defaultPageSize": 20  // Smaller pages
       }
     }
   }
   ```

2. **Stream results** (future feature)

3. **Clear cache** (if implemented):
   ```bash
   jot cache clear
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

**Symptom**: Pi doesn't recognize `jot_*` tools

**Check**:
```
/tools list
# Should show jot_search, jot_list, etc.
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
    "@zenobi-us/pi-jot": {
      "toolPrefix": "notes_"  // Custom prefix
    }
  }
}
```

**Solution**: Either:

1. Use configured prefix: `notes_search`
2. Or reset to default: `"toolPrefix": "jot_"`

### Pi Errors

**Error**: `Tool execution failed`

**Debug**:

1. **Check pi logs**:
   ```bash
   tail -f ~/.pi/logs/error.log
   ```

2. **Test tool directly**:
   ```
   /call jot_search {"query": "test"}
   ```

3. **Verify CLI works**:
   ```bash
   jot notes list
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

2. **Jot verbose**:
   ```bash
   jot --verbose notes list
   ```

3. **Extension logs**:
   ```bash
   tail -f ~/.pi/extensions/@zenobi-us/pi-jot/debug.log
   ```

### Verify Setup

Complete checklist:

```bash
# 1. Jot installed
jot version

# 2. Extension installed
npm list -g @zenobi-us/pi-jot

# 3. In pi settings
cat ~/.pi/settings.json | grep jot

# 4. Notebook initialized
ls ~/.config/jot/config.json
jot notebooks list

# 5. Tools registered
pi -c "/tools list" | grep jot
```

### Reset to Defaults

If all else fails:

```bash
# 1. Uninstall extension
npm uninstall -g @zenobi-us/pi-jot

# 2. Remove config
rm -rf ~/.pi/settings.json.jot.bak
# (Backup first!)

# 3. Reinstall
npm install -g @zenobi-us/pi-jot

# 4. Restart pi
/reload
```

---

## Common Patterns

### Test CLI Directly

Before using tools, verify CLI works:

```bash
# List notes
jot --notebook ~/notes notes list

# Search
jot --notebook ~/notes notes search --query "test"

# SQL
jot --notebook ~/notes notes sql --query "SELECT COUNT(*) FROM notes"

# Get note
jot --notebook ~/notes notes get path/to/note.md
```

### Validate JSON

All configs are JSON - validate before use:

```bash
cat .jot.json | jq .
cat ~/.pi/settings.json | jq .
```

### Check Permissions

Ensure pi can access notebooks:

```bash
# Read permissions
ls -la ~/notes/.jot.json

# Write permissions (for create)
touch ~/notes/test-write.md && rm ~/notes/test-write.md
```

---

## Getting Help

### Where to Report Issues

- **Jot CLI bugs**: [jot/issues](https://github.com/zenobi-us/jot/issues)
- **Extension bugs**: [pi-jot/issues](https://github.com/zenobi-us/pi-jot/issues)
- **Pi integration**: [pi/discussions](https://github.com/mariozechner/pi-coding-agent/discussions)

### Include in Bug Reports

```
**Environment**:
- OS: Linux/macOS/Windows
- Jot version: `jot version`
- Extension version: `npm list -g @zenobi-us/pi-jot`
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
