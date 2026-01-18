---
id: fba56e5b
title: Update Documentation for SQL Glob Pattern Behavior
created_at: 2026-01-18 21:30:40 GMT+10:30
updated_at: 2026-01-18 21:30:40 GMT+10:30
status: todo
epic_id: TBD
phase_id: TBD
assigned_to: current
priority: MEDIUM
estimated_effort: 45 minutes - 1 hour
---

# Update Documentation for SQL Glob Pattern Behavior

## Objective

Update CLI help text, user documentation, and function reference to clearly explain the new SQL glob pattern preprocessing behavior, ensuring users understand that file patterns are always resolved relative to the notebook root directory regardless of current working directory.

## Problem Context

**Documentation Gap**: Current SQL flag documentation does not explain glob pattern behavior, leaving users unaware of:
- How file patterns are resolved in SQL queries
- Security protections against path traversal
- Consistent behavior across different execution contexts
- Best practices for writing portable SQL queries

**User Experience Impact**: Without clear documentation, users may:
- Write queries assuming current directory resolution
- Be confused by consistent behavior across directories  
- Not understand security restrictions on file access
- Miss opportunities to use powerful glob patterns

## Steps

### 1. Update CLI Help Text

**Location**: `cmd/notes_search.go`
**Target**: `--sql` flag help description

**Current Help Text**:
```
--sql string    Execute custom SQL query with DuckDB
```

**Updated Help Text**:
```
--sql string    Execute custom SQL query with DuckDB. File patterns (*.md, **/*.md) 
                are resolved relative to notebook root directory. Examples:
                  --sql "SELECT * FROM '**/*.md' LIMIT 5"
                  --sql "SELECT title FROM '*.md' WHERE title LIKE '%todo%'"
```

**Extended Help Example**:
Add to command long description:
```
SQL Query Examples:
  Find all notes:           --sql "SELECT * FROM '**/*.md'"
  Search by pattern:        --sql "SELECT * FROM '*.md' WHERE content LIKE '%keyword%'"
  Subdirectory only:        --sql "SELECT * FROM 'projects/*.md'"
  
File Pattern Behavior:
  - All file patterns resolve from notebook root directory
  - Queries work consistently regardless of current directory
  - Security restrictions prevent access to files outside notebook
  - Use forward slashes in patterns (cross-platform compatibility)
```

### 2. Update Function Reference Documentation

**Location**: Create or update in appropriate documentation directory
**File**: `docs/sql-queries.md` (if exists) or inline in code comments

**Content to Add**:

#### SQL Pattern Resolution

File glob patterns in SQL queries are automatically processed to ensure consistent, secure behavior:

**Pattern Types Supported**:
- `*.md` - All markdown files in notebook root
- `**/*.md` - All markdown files recursively in entire notebook
- `subfolder/*.md` - All markdown files in specific subfolder
- `**/subfolder/*.md` - All markdown files in any subfolder named 'subfolder'

**Resolution Behavior**:
- All patterns resolve relative to notebook root directory
- Behavior is identical regardless of current working directory
- Security restrictions prevent access outside notebook boundaries
- Path traversal attempts (../) are blocked and logged

**Examples**:
```sql
-- Find notes with specific title pattern
SELECT file_path, title FROM '*.md' 
WHERE title LIKE '%meeting%' 
ORDER BY file_path;

-- Search all notes for content
SELECT file_path, content FROM '**/*.md' 
WHERE content MATCH 'keyword' 
LIMIT 10;

-- Query specific project folder
SELECT title, tags FROM 'projects/**/*.md'
WHERE tags LIKE '%urgent%';
```

### 3. Add Security Documentation

**Security Section Content**:

#### Security Protections

SQL query processing includes several security measures:

**Path Traversal Protection**:
- Queries cannot access files outside the notebook directory
- Path traversal attempts using `../` are automatically blocked
- Only files within the notebook tree are accessible
- Security violations are logged for monitoring

**Safe Query Practices**:
- Use relative paths from notebook root: `'subfolder/*.md'`
- Avoid absolute paths: `/home/user/notes/*.md` (blocked)
- Stick to forward slashes for cross-platform compatibility
- Test queries with `LIMIT` clauses during development

### 4. Update Error Message Documentation

**Error Reference Section**:

#### Common SQL Errors

**Path Traversal Detected**:
```
Error: path traversal detected: query would access files outside notebook
```
- **Cause**: Query contains `../` or absolute paths
- **Solution**: Use relative paths from notebook root
- **Example**: Change `'../other/*.md'` to `'other/*.md'`

**Pattern Processing Failed**:
```
Error: query preprocessing failed: malformed pattern
```
- **Cause**: Invalid glob pattern syntax
- **Solution**: Check quote matching and pattern format
- **Example**: Ensure patterns like `'*.md'` have matching quotes

### 5. Add Usage Examples to CLI

**Location**: `cmd/notes_search.go`
**Enhancement**: Extended examples in command description

**Example Section**:
```go
var searchCmd = &cobra.Command{
    Use:   "search",
    Short: "Search notes using various methods",
    Long: `Search notes using text search, tags, or custom SQL queries.

SQL Query Examples:
  Basic pattern search:
    opennotes notes search --sql "SELECT * FROM '*.md' LIMIT 5"
  
  Content search across all notes:
    opennotes notes search --sql "SELECT file_path FROM '**/*.md' WHERE content LIKE '%todo%'"
  
  Specific folder query:
    opennotes notes search --sql "SELECT title FROM 'projects/*.md' ORDER BY title"
  
  Complex joins and filtering:
    opennotes notes search --sql "SELECT DISTINCT tags FROM '**/*.md' WHERE tags IS NOT NULL"

File Pattern Notes:
  - Patterns resolve from notebook root (not current directory)
  - Use forward slashes for cross-platform compatibility  
  - Security restrictions prevent access outside notebook
  - All DuckDB SQL syntax supported within security constraints`,
}
```

### 6. Create Best Practices Guide

**Location**: Code comments or separate documentation
**Content**: Best practices for SQL query usage

**Best Practices Content**:

#### SQL Query Best Practices

**Pattern Writing**:
- Start patterns from notebook root: `'folder/*.md'` not `'./folder/*.md'`
- Use double asterisk for recursive search: `'**/*.md'`
- Be specific with folders to improve performance: `'docs/*.md'` vs `'**/*.md'`
- Test with LIMIT clauses during development

**Performance Optimization**:
- Use specific patterns instead of `'**/*'` when possible
- Add WHERE clauses to filter results early
- Consider indexing strategies for large notebooks
- Monitor query execution time with complex patterns

**Security Awareness**:
- Never attempt path traversal (automatically blocked)
- Understand that all queries are read-only
- Be aware that query preprocessing is logged
- Report unexpected security errors to maintainers

## Expected Outcome

**Documentation Completeness**:
- ✅ CLI help text clearly explains glob pattern behavior
- ✅ Function reference documents all pattern types
- ✅ Security protections are clearly explained
- ✅ Common errors have clear resolution guidance
- ✅ Best practices guide helps users write effective queries

**User Experience**:
- ✅ Users understand pattern resolution behavior
- ✅ Security restrictions are clearly communicated
- ✅ Examples provide clear usage guidance
- ✅ Error messages lead to quick resolution

**Documentation Artifacts**:
- Updated CLI help text with examples
- Function reference with pattern documentation
- Security guide explaining protections
- Error reference with solutions
- Best practices guide for effective usage

## Acceptance Criteria

### CLI Help Updates
- [ ] `--sql` flag help includes pattern behavior explanation
- [ ] Command description includes pattern resolution notes
- [ ] Examples demonstrate common usage patterns
- [ ] Cross-platform compatibility notes included

### Function Reference
- [ ] All supported pattern types documented with examples
- [ ] Resolution behavior clearly explained
- [ ] Security protections documented
- [ ] Performance considerations included

### Error Documentation
- [ ] Common error scenarios documented
- [ ] Clear resolution steps for each error type
- [ ] Examples of correct vs incorrect patterns
- [ ] Security violation explanations included

### Best Practices
- [ ] Pattern writing guidelines provided
- [ ] Performance optimization tips included
- [ ] Security awareness guidance documented
- [ ] Cross-platform compatibility notes added

### Consistency
- [ ] Documentation style matches existing help text
- [ ] Examples use consistent notebook structure
- [ ] Terminology aligns with existing documentation
- [ ] All pattern examples are tested and verified

## Dependencies

**Implementation Dependencies**:
- Completion of SQL preprocessing implementation
- Test validation of documented examples
- Security testing verification

**Documentation Dependencies**:
- Existing CLI help text patterns
- Current documentation style guide
- Example notebook structure for consistent examples

## Files to Modify

**Primary Files**:
- `cmd/notes_search.go` - CLI help text and examples
- Code comments in `internal/services/db.go` - Function documentation
- Any existing documentation files in `docs/` directory

**Verification Files**:
- Test examples in documentation match test cases
- Verify examples work in actual CLI usage
- Cross-reference with security test cases

## Time Estimate

**Total: 45 minutes - 1 hour**
- CLI help text updates: 15 minutes
- Function reference writing: 15 minutes
- Security documentation: 10 minutes
- Error message documentation: 10 minutes
- Best practices guide: 10 minutes
- Verification and testing: 5 minutes

## Related Tasks

- **Implementation**: [task-847f8a69-implement-sql-preprocessing.md]
- **Testing**: [task-1c5a8eca-sql-glob-preprocessing-tests.md]
- **Research Reference**: [learning-548a8336-sql-glob-rooting-research.md]

---

**Priority**: 🟡 **MEDIUM** - Important for user experience but not blocking  
**Complexity**: ⚡ **LOW** - Straightforward documentation updates  
**Quality Gate**: All documented examples must be tested and verified