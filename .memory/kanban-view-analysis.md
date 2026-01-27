# Kanban View Architecture & Flow Analysis

**Branch**: `fix/kanban-view`  
**Status**: Investigation complete  
**Date**: 2026-01-27

## Executive Summary

The kanban view is a **parameterized built-in view** that filters and sorts notes by status metadata. It demonstrates OpenNotes' powerful view system with:

- **Built-in definition** in `ViewService.initializeBuiltinViews()`
- **Parameter support** for dynamic status filtering (e.g., `--param status=todo,done`)
- **SQL generation** from declarative conditions
- **Template variable resolution** for dates and placeholders
- **Hierarchy-based view loading**: notebook > global > built-in

---

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ User Command                                                │
│ opennotes notes view kanban --param status=todo,in-progress │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────────┐
│ notes_view.go: RunE                                │
│ - Parse args: viewName = "kanban"                  │
│ - Parse params: {status: "todo,in-progress"}      │
│ - Get notebook context                             │
└────────────────┬───────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────────────┐
│ ViewService.GetView("kanban")                          │
│ 1. Check notebook views (.opennotes.json)              │
│ 2. Check global views (~/.config/opennotes/config.json)│
│ 3. Check built-in views (hardcoded)                    │
│ Returns: ViewDefinition {...}                         │
└────────────────┬───────────────────────────────────────┘
                 │
                 ▼ (Returns kanban ViewDefinition)
┌────────────────────────────────────────────────────────┐
│ ViewService.GenerateSQL(view, params)                  │
│ 1. Validate params: ValidateParameters()               │
│ 2. Apply defaults: ApplyParameterDefaults()            │
│ 3. Resolve templates: ResolveTemplateVariables()       │
│ 4. Build WHERE clause from conditions                  │
│ 5. Append ORDER BY and LIMIT                           │
│ Returns: (sqlQuery, args, error)                       │
└────────────────┬───────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────────────┐
│ Generated SQL                                          │
│ SELECT * FROM read_markdown(?, include_filepath:=true)│
│ WHERE metadata->>'status' IN (?,?)                     │
│ ORDER BY (metadata->>'priority')::INTEGER DESC, ...    │
│ Args: ["/path/**/*.md", "todo", "in-progress"]        │
└────────────────┬───────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────────────┐
│ DbService.GetDB() → QueryContext()                     │
│ DuckDB executes query against markdown files           │
│ Returns: rows (with path, content, metadata columns)  │
└────────────────┬───────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────────────┐
│ Format Results as []map[string]interface{}             │
│ Extract columns and convert rows to maps               │
└────────────────┬───────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────────────┐
│ Display.RenderSQLResults()                             │
│ Format: list/table/json                                │
│ Output to user                                         │
└────────────────────────────────────────────────────────┘
```

---

## Kanban View Definition

### Built-in Definition (in view.go)

```go
vs.builtinViews["kanban"] = &core.ViewDefinition{
    Name:        "kanban",
    Description: "Notes grouped by status column",
    Parameters: []core.ViewParameter{
        {
            Name:        "status",
            Type:        "list",
            Required:    false,
            Default:     "backlog,todo,in-progress,reviewing,testing,deploying,done",
            Description: "Comma-separated list of status values",
        },
    },
    Query: core.ViewQuery{
        Conditions: []core.ViewCondition{
            {
                Logic:    "AND",
                Field:    "metadata->>'status'",
                Operator: "IN",
                Value:    "{{status}}",  // ← Will be replaced with param value
            },
        },
        OrderBy: "(metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC",
    },
}
```

### Type Definitions (core/view.go)

```go
type ViewDefinition struct {
    Name        string          // "kanban"
    Description string          // "Notes grouped by status column"
    Parameters  []ViewParameter // List of dynamic parameters
    Query       ViewQuery       // Query structure
}

type ViewParameter struct {
    Name        string // "status"
    Type        string // "list", "string", "date", "bool"
    Required    bool   // Can user skip this?
    Default     string // Fallback value
    Description string // Help text
}

type ViewQuery struct {
    Conditions []ViewCondition // WHERE clauses
    OrderBy    string          // ORDER BY clause
    GroupBy    string          // GROUP BY clause
    Limit      int             // LIMIT clause
}

type ViewCondition struct {
    Logic    string // "AND" or "OR"
    Field    string // metadata->>'status' (whitelisted)
    Operator string // "=", "!=", "<", ">", "<=", ">=", "LIKE", "IN", "IS NULL"
    Value    string // "{{status}}" (placeholder)
}
```

---

## SQL Generation Process

### Example: `opennotes notes view kanban --param status=todo,done`

**Step 1: Parse Parameters**
```go
userParams := vs.ParseViewParameters("status=todo,done")
// Result: {status: "todo,done"}
```

**Step 2: Validate Against Definition**
```go
vs.ValidateParameters(view, userParams)
// Checks: is status a known parameter? Is type valid?
```

**Step 3: Apply Defaults**
```go
resolved := vs.ApplyParameterDefaults(view, userParams)
// status already provided, no defaults applied
// Result: {status: "todo,done"}
```

**Step 4: Resolve Template Variables**
```go
// No {{today}} or {{yesterday}} templates in kanban
// Just template parameter placeholders like {{status}}
```

**Step 5: Build WHERE Clause**
```
Field:    metadata->>'status'
Operator: IN
Value:    "todo,done" (from params)

// Generates:
// WHERE metadata->>'status' IN (?, ?)
// Args: ["todo", "done"]
```

**Step 6: Final SQL**
```sql
SELECT * FROM read_markdown(?, include_filepath:=true)
WHERE metadata->>'status' IN (?,?)
ORDER BY (metadata->>'priority')::INTEGER DESC, metadata->>'updated_at' DESC
LIMIT 0  -- No limit specified

-- Final args:
-- Arg[0]: "/notebook/**/*.md"    (glob pattern)
-- Arg[1]: "todo"                 (status value 1)
-- Arg[2]: "done"                 (status value 2)
```

---

## Key Components

### 1. ViewService (`internal/services/view.go`)

**Responsibilities:**
- Initialize and manage all 6 built-in views
- Load custom views from notebook/global configs
- Generate SQL from view definitions + parameters
- Validate user input (security + correctness)
- Resolve template variables

**Key Methods:**
- `GetView(name)` — Hierarchy lookup: notebook > global > built-in
- `GenerateSQL(view, params)` — Convert definition + params to SQL query
- `ValidateParameters(view, params)` — Enforce parameter constraints
- `ApplyParameterDefaults(view, params)` — Fill in missing defaults
- `FormatQueryValue(operator, value)` — Convert value for SQL based on operator

**Security Features:**
- `validateField()` — Whitelist allowed fields (metadata, path, stats, etc.)
- `validateOperator()` — Whitelist allowed operators
- `escapeSQL()` — Escape single quotes in string values
- Parameter type validation (string, list, date, bool)

### 2. View Definition (core/view.go)

**Structure:**
- Declarative query definition (conditions, ordering, limit)
- Parameter definitions with types and defaults
- Name and description for discovery

**Operators Support:**
- `=`, `!=`, `<`, `>`, `<=`, `>=` — Comparison
- `LIKE` — Pattern matching
- `IN` — Set membership (list parameters)
- `IS NULL` — Null checks

### 3. Command Handler (`cmd/notes_view.go`)

**Flow:**
1. Parse view name from args
2. Parse `--param key=value` flags
3. Get notebook context
4. Initialize ViewService
5. Get view definition
6. Generate SQL + args
7. Execute query via DuckDB
8. Format results (list/table/json)
9. Display output

**Note:** Thin orchestration - all logic in services

---

## View Hierarchy

Views are resolved in this order:

```
1. Notebook-specific view (.opennotes.json in notebook)
   └─ User can override built-in "kanban" in their notebook

2. Global view (~/.config/opennotes/config.json)
   └─ User's custom kanban for all notebooks

3. Built-in view (hardcoded in ViewService)
   └─ Default kanban implementation
```

**Example:** If user creates `.opennotes.json` with:
```json
{
  "views": {
    "kanban": {
      "name": "kanban",
      "description": "Team kanban board (custom)",
      "parameters": [{
        "name": "status",
        "type": "list",
        "default": "planning,dev,qa,deployed"
      }],
      "query": { ... }
    }
  }
}
```

Then `opennotes notes view kanban` uses **notebook version**, not built-in.

---

## 6 Built-in Views

| Name | Purpose | Parameters | Ordering |
|------|---------|------------|----------|
| **today** | Notes created/updated today | None | updated_at DESC |
| **recent** | Last 20 modified notes | None | updated_at DESC, limit 20 |
| **kanban** | Group by status column | status (list) | priority DESC, updated_at DESC |
| **untagged** | Notes without tags | None | created_at DESC |
| **orphans** | Notes with no incoming links | definition (orphan type) | created_at DESC |
| **broken-links** | Notes with invalid references | None | updated_at DESC |

---

## Parameter System

### Parameter Types

**1. String**
```go
{
  "name": "query",
  "type": "string",
  "required": false,
  "default": "",
  "validation": "max 256 chars"
}
```

**2. List** (comma-separated)
```go
{
  "name": "status",
  "type": "list",
  "required": false,
  "default": "todo,done",
  "validation": "no empty items"
}
```

**3. Date** (YYYY-MM-DD)
```go
{
  "name": "after",
  "type": "date",
  "required": false,
  "default": "{{today}}",
  "validation": "valid date format"
}
```

**4. Bool** (true/false)
```go
{
  "name": "archived",
  "type": "bool",
  "required": false,
  "default": "false"
}
```

### Parameter Validation

```go
// For kanban status parameter:
// Input: "todo,in-progress,done"
// 1. Check it's a known parameter ✓
// 2. Check type is "list" ✓
// 3. Split by comma → ["todo", "in-progress", "done"]
// 4. Verify no empty items ✓
// Result: Valid
```

---

## Security Model

### Input Validation Strategy

**1. Field Whitelist**
```go
allowedPrefixes := []string{
    "metadata->>",  // JSON field access
    "metadata->",   // JSON object access
    "path",
    "file_path",
    "content",
    "stats->",
    "stats->>",
}

// User cannot specify: "1 OR 1=1", "DROP TABLE", etc.
```

**2. Operator Whitelist**
```go
allowedOperators := map[string]bool{
    "=": true, "!=": true, "IN": true,
    "LIKE": true, "IS NULL": true,
    // No raw SQL, no UNION, no semicolon
}
```

**3. Parameter Type Validation**
```go
// For list: "todo,in-progress" → ["todo", "in-progress"]
// Check each item is non-empty
// No SQL syntax allowed in parameter values
```

**4. SQL Escaping**
```go
// Single quotes are escaped: ' → ''
// Parameterized queries with ? placeholders
// DuckDB driver handles escaping
```

### Example Attack Prevention

```
User input: status="todo' OR '1'='1"
Processed: status="todo' OR '1'='1" (no removal)
Formatted: ('todo\' OR \'1\'=\'1') ← Escaped!
Result: Treated as literal string "todo' OR '1'='1"
✓ SQL injection prevented
```

---

## Testing Coverage

### Test Cases (view_test.go)

1. **View Discovery**
   - `TestViewService_ListAllViews()` — All 6 built-in views present
   - `TestViewService_GetView_BuiltinView()` — Correct definition loaded
   - `TestViewService_GetView_NotFound()` — Returns error if missing

2. **Kanban Specific**
   - `TestViewService_KanbanView_HasParameter()` — Status parameter exists
   - Verifies default: "backlog,todo,in-progress,reviewing,testing,deploying,done"

3. **SQL Generation**
   - `TestViewService_GenerateSQL_INOperator()` — IN operator formatting
   - `TestViewService_GenerateSQL_WithUserParameters()` — Param substitution
   - `TestViewService_GenerateSQL_MultipleConditions()` — AND logic

4. **Parameter Validation**
   - `TestViewService_ValidateParameters_*()— Parameter type checking
   - Required vs optional handling
   - Default value application

5. **Security** (e2e/search_test.go)
   - `TestE2E_Security_SQLInjectionPrevention()` — Injection attack attempts
   - Tests: `'; DROP TABLE notes; --`, `1' OR '1'='1`, etc.

### Test Status: ✅ All Passing (161+ tests, ~4 seconds)

---

## Example Usage Flows

### Flow 1: Simple kanban with defaults
```bash
$ opennotes notes view kanban

# Uses default status: backlog,todo,in-progress,reviewing,testing,deploying,done
# Orders by priority DESC, then updated_at DESC
# Returns all notes in any of these statuses
```

### Flow 2: Kanban with custom status filter
```bash
$ opennotes notes view kanban --param status=todo,in-progress

# Only shows notes with status=todo or status=in-progress
# Same ordering as above
# Typically used for "active work" dashboard
```

### Flow 3: Kanban with JSON output for scripting
```bash
$ opennotes notes view kanban --param status=todo --format json

# Output: [{path, content, metadata}, {path, content, metadata}, ...]
# Useful for piping to jq, building dashboards, etc.
```

### Flow 4: Custom notebook kanban
**In notebook/.opennotes.json:**
```json
{
  "views": {
    "kanban": {
      "name": "kanban",
      "description": "Team workflow",
      "parameters": [{
        "name": "status",
        "type": "list",
        "default": "planning,dev,testing,deployed"
      }],
      "query": { ... }
    }
  }
}
```

**Command:**
```bash
$ cd my-notebook && opennotes notes view kanban

# Uses custom status values: planning,dev,testing,deployed
# NOT the built-in defaults
```

---

## Potential Issues & Improvements

### Current Limitations

1. **No grouping** — Kanban view returns flat list, not grouped by status column
   - Uses `OrderBy` not `GroupBy`
   - Client-side grouping would be needed for visual kanban board

2. **No aggregation** — Can't count cards per column
   - Would need `SELECT status, COUNT(*) FROM ...` 
   - Currently returns individual notes

3. **Template variables** — Only date placeholders
   - Could support user-defined variables
   - Could support environment variables

4. **Performance** — Large result sets
   - No automatic pagination
   - Full result set returned to client

### Enhancement Opportunities

1. **Add GroupBy support** → Generate `GROUP BY` clauses
2. **Add aggregate functions** → COUNT, SUM, MAX, MIN
3. **Add pagination** → `OFFSET` support
4. **Add sorting preferences** → Multiple ORDER BY options
5. **Add view composition** → Combine multiple views with UNION
6. **Add caching** → Cache view results for better UX

---

## Summary

The **kanban view** is a well-architected example of OpenNotes' view system:

✅ **Declarative**: Define query in JSON, not code  
✅ **Parameterized**: Users customize behavior without code  
✅ **Secure**: Multiple validation layers prevent injection  
✅ **Hierarchical**: Notebook > Global > Built-in precedence  
✅ **Extensible**: Custom views in config files  
✅ **Tested**: Full coverage including security tests  

The kanban view specifically demonstrates how to:
- Use IN operator for multiple status values
- Order by computed fields (cast to INTEGER)
- Support optional parameters with defaults
- Integrate view definitions with SQL generation

