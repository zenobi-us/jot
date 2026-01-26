---
id: d4fca870
title: Views System - Named Reusable Query Presets
created_at: 2026-01-20T23:47:00+10:30
updated_at: 2026-01-22T12:45:00+10:30
status: completed
epic_id: 3e01c563
requires_qa: false
qa_completed: 2026-01-22
---

# Views System Specification

✅ **APPROVED SPECIFICATION - Q&A DISCUSSION COMPLETED 2026-01-22**

**Feature**: Views System for OpenNotes - Named, reusable query presets with parameterization

**Status**: ✅ **APPROVED - READY FOR IMPLEMENTATION**

**Related Epic**: Advanced Note Creation and Search Capabilities (epic-3e01c563)

**Document Purpose**: Complete technical specification for the Views System feature with an Open Questions section for unresolved design decisions that must be addressed via Q&A discussion before implementation begins.

---

## ✅ Q&A DISCUSSION COMPLETED

All 6 design questions have been resolved via Q&A discussion on 2026-01-22.

**Decisions Made**:

| # | Question | Decision |
|---|----------|----------|
| 1 | Command Structure | **A**: `opennotes notes view <name> [--param key=value]` |
| 2 | Output Formatting | **Flag-based**: `--format list|table|json`, views are query-only |
| 3 | View Definition Scope | **A**: Query-only (conditions, order, limit, group by) |
| 4 | Broken Links Detection | **C**: Both frontmatter AND markdown body links |
| 5 | Kanban Parameter Handling | **Hybrid**: param → notebook config → built-in default |
| 6 | Orphans Definition | **Hybrid**: param → config → default (isolated node) |

**This specification is now APPROVED for implementation.**

---

## Table of Contents

1. [Overview](#overview)
2. [What's In Scope](#whats-in-scope)
3. [What's Out of Scope (Pending Q&A)](#whats-out-of-scope-pending-qa)
4. [Current Working Design (Subject to Change)](#current-working-design-subject-to-change)
5. [Built-in Views (Confirmed)](#built-in-views-confirmed)
6. [Configuration Schema (JSON)](#configuration-schema-json)
7. [Storage Hierarchy & Precedence](#storage-hierarchy--precedence)
8. [Template Variable System](#template-variable-system)
9. [Parameter System](#parameter-system)
10. [Implementation Details](#implementation-details)
11. [Security Considerations](#security-considerations)
12. [Testing Requirements](#testing-requirements)
13. [Performance Targets](#performance-targets)
14. [Documentation Updates Needed](#documentation-updates-needed)
15. [Migration Impact](#migration-impact)
16. [Open Questions (REQUIRES Q&A DISCUSSION)](#open-questions-requires-qa-discussion)
17. [Acceptance Criteria (Pending Q&A)](#acceptance-criteria-pending-qa)
18. [Next Steps](#next-steps)

---

## Overview

### Feature Summary

The **Views System** provides named, reusable search queries with parameterization support. Views act as "saved searches" that users can invoke by name instead of constructing complex queries each time.

### Problem Being Solved

**Current State**:
- Users must write full SQL queries for common patterns (today's notes, kanban boards, broken links)
- No way to save frequently-used queries for reuse
- Boolean search flags still require typing multiple field filters
- Difficult to share query patterns across team members

**Desired State**:
- Users can invoke common queries by name: `opennotes notes view today`
- Built-in views for universal patterns (recent, untagged, orphans, broken-links)
- Custom views can be defined in global config or per-notebook
- Views support parameters for flexibility (e.g., kanban with custom status values)
- Template variables automatically resolve ({{today}}, {{yesterday}}, etc.)

### User Value Proposition

1. **Faster Workflows**: Common queries accessible via simple names
2. **Discoverability**: Built-in views showcase advanced capabilities
3. **Team Collaboration**: Share custom views via notebook config
4. **Progressive Disclosure**: Views bridge gap between simple list and complex SQL
5. **Consistency**: Same query pattern works across all notebooks

---

## What's In Scope

✅ **Built-in views** - 6 predefined views (today, recent, kanban, untagged, orphans, broken-links)  
✅ **Custom user views** - User-defined views in global config (`~/.config/opennotes/config.json`)  
✅ **Notebook-specific views** - Views defined per notebook (`.opennotes.json`)  
✅ **View parameterization** - Dynamic parameters for flexible queries  
✅ **Template variables** - `{{today}}`, `{{yesterday}}`, `{{this_week}}`, `{{this_month}}`, `{{now}}`  
✅ **3-tier hierarchy** - Built-in → Global → Notebook (precedence order)  
✅ **JSON configuration** - Uses existing ConfigService for loading views  
✅ **Security** - Same parameterized query protections as search (defense-in-depth)

---

## Design Decisions (Resolved via Q&A)

✅ **Command structure** - Dedicated command: `opennotes notes view <name> [--param key=value]`  
✅ **Output formatting** - Flag-based: `--format list|table|json` (views are query-only, formatting is orthogonal)  
✅ **View definition scope** - Query-only (conditions, order, limit, group by)  
✅ **Interactive features** - OUT OF SCOPE (views are query presets, not UI components)  
✅ **Broken links detection** - Both frontmatter AND markdown body links (comprehensive)  
✅ **Kanban parameter handling** - Hybrid fallback: param → notebook config → built-in default  
✅ **Orphans definition** - Hybrid fallback: param → config → default (isolated node: no links AND not tagged)

**All questions resolved via Q&A discussion on 2026-01-22.**

---

## Command Structure (Confirmed)

### Command Syntax

```bash
opennotes notes view <name> [--param key=value] [--format list|table|json]
```

**Examples**:
```bash
# Built-in views
opennotes notes view today
opennotes notes view recent
opennotes notes view kanban
opennotes notes view kanban --param status=todo,in-progress,done
opennotes notes view untagged
opennotes notes view orphans
opennotes notes view broken-links

# Output format options
opennotes notes view today --format json
opennotes notes view kanban --format table
opennotes notes view recent --format list

# Custom views (defined in config)
opennotes notes view my-workflow
opennotes notes view sprint-planning --param sprint=2026-Q1-S3
```

**Decision**: Dedicated command under `notes` subcommand, consistent with `notes list` and `notes search`.

---

## Built-in Views (Confirmed)

### 1. Today View

**Name**: `today`

**Description**: Notes created or updated today

**Use Case**: Quick daily review of active work

**Query Logic**:
```sql
SELECT * FROM notes
WHERE (data.created >= '{{today}}' OR data.updated >= '{{today}}')
ORDER BY updated DESC
LIMIT 50
```

**Template Variables**:
- `{{today}}` → Current date in YYYY-MM-DD format (e.g., `2026-01-20`)

**Example Output**:
```
### Today's Notes (3)

- [Daily Standup Notes] journal/2026-01-20-standup.md
- [Project Planning] projects/opennotes-views.md
- [Meeting Notes] meetings/team-sync.md
```

---

### 2. Recent View

**Name**: `recent`

**Description**: Recently modified notes (last 20)

**Use Case**: Quick access to recently edited notes

**Query Logic**:
```sql
SELECT * FROM notes
ORDER BY updated DESC
LIMIT 20
```

**Template Variables**: None

**Example Output**:
```
### Recent Notes (20)

- [Views System Design] specs/views-system.md
- [Search Enhancement] specs/search-enhancement.md
- [Daily Standup] journal/2026-01-20-standup.md
...
```

---

### 3. Kanban View

**Name**: `kanban`

**Description**: Notes grouped by status column

**Use Case**: Project management, issue tracking, workflow visualization

**Parameters**:
- `status` (optional): Comma-separated list of status values
- **Default**: `backlog,todo,in-progress,reviewing,testing,deploying,done`

**Query Logic**:
```sql
SELECT * FROM notes
WHERE data.status IN ({{status}})
ORDER BY data.priority DESC, updated DESC
```

**Template Variables**:
- `{{status}}` → Resolved from `--param status=...` or default value

**Example Usage**:
```bash
# Use default status values
opennotes notes view kanban

# Custom status values
opennotes notes view kanban --param status=todo,in-progress,done
```

**Example Output** (format TBD via Q&A):
```
### Kanban Board (15 notes)

TODO (5):
- [Feature: Views System] tasks/views-system.md
- [Bug: SQL Injection] bugs/sql-injection.md
...

IN-PROGRESS (3):
- [Spec: Search Enhancement] specs/search-enhancement.md
...

DONE (7):
- [Epic: Test Coverage] epics/test-coverage.md
...
```

---

### 4. Untagged View

**Name**: `untagged`

**Description**: Notes without any tags

**Use Case**: Identify notes that need categorization

**Query Logic**:
```sql
SELECT * FROM notes
WHERE (data.tags IS NULL OR data.tags = [] OR data.tag IS NULL)
ORDER BY created DESC
```

**Template Variables**: None

**Example Output**:
```
### Untagged Notes (12)

- [Random Thoughts] journal/2026-01-15-thoughts.md
- [Meeting Notes] meetings/2026-01-10-client-call.md
...
```

---

### 5. Orphans View

**Name**: `orphans`

**Description**: Notes with no incoming links (no other notes reference them)

**Use Case**: Identify isolated notes in knowledge graph, find disconnected content

**Graph Semantics**: Documents with no incoming edges in the link graph

**Query Logic**:
```sql
SELECT * FROM notes n1
WHERE NOT EXISTS (
  SELECT 1 FROM notes n2
  WHERE n1.path IN (SELECT unnest(n2.data.links))
)
ORDER BY created DESC
```

**Template Variables**: None

**Definition** (Confirmed via Q&A):
- **Default**: Isolated nodes (no links AND not tagged/categorized)
- **Configurable**: Hybrid fallback (param → notebook config → default)
- Users can override via `--param definition=no-incoming|no-links|isolated`

**Example Output**:
```
### Orphan Notes (8)

- [Random Note] random/standalone.md
- [Old Archive] archive/2025/old-project.md
...
```

---

### 6. Broken Links View

**Name**: `broken-links`

**Description**: Notes containing links to non-existent files

**Use Case**: Identify and fix broken references in knowledge graph

**Detection Scope** (Confirmed via Q&A):
- ✅ **Both frontmatter AND body links** (comprehensive coverage)
- Frontmatter: `data.links` array
- Body: Parse `[text](path)` and `[[wikilink]]` syntax

**Query Logic** (assuming frontmatter + body links):
```sql
SELECT DISTINCT n.*
FROM notes n
CROSS JOIN unnest(n.data.links) AS link
WHERE NOT EXISTS (
  SELECT 1 FROM notes target
  WHERE target.path = link
)
ORDER BY n.updated DESC
```

**Template Variables**: None

**Example Output**:
```
### Notes with Broken Links (3)

- [Project Documentation] docs/architecture.md
  - Broken: ../specs/missing-spec.md
- [Meeting Notes] meetings/2026-01-15-sync.md
  - Broken: ../tasks/deleted-task.md
...
```

---

## Configuration Schema (JSON)

### Global Config Location

**Path**: `~/.config/opennotes/config.json`

**Purpose**: User-wide custom views available in all notebooks

---

### Notebook Config Location

**Path**: `<notebook>/.opennotes.json`

**Purpose**: Notebook-specific views (team-shared or project-specific)

---

### Schema Structure

```json
{
  "views": {
    "view-name": {
      "description": "Human-readable description",
      "parameters": [
        {
          "name": "param-name",
          "type": "string|list|date|bool",
          "required": true|false,
          "default": "default-value",
          "description": "Parameter description"
        }
      ],
      "query": {
        "conditions": [
          {
            "logic": "AND|OR",
            "field": "data.field-name",
            "operator": "=|!=|<|>|<=|>=|LIKE|IN|IS NULL",
            "value": "value or {{template-var}}"
          }
        ],
        "order_by": "field ASC|DESC",
        "group_by": "field-name",
        "limit": 100
      }
    }
  }
}
```

---

### Example Custom View: My Workflow

```json
{
  "views": {
    "my-workflow": {
      "description": "My active workflow notes",
      "query": {
        "conditions": [
          {
            "logic": "AND",
            "field": "data.tag",
            "operator": "=",
            "value": "workflow"
          },
          {
            "logic": "AND",
            "field": "data.status",
            "operator": "!=",
            "value": "archived"
          }
        ],
        "order_by": "updated DESC",
        "limit": 50
      }
    }
  }
}
```

**Usage**:
```bash
opennotes notes view my-workflow
```

**Output**:
```
### My Workflow Notes (12)

- [Feature: Views System] tasks/views-system.md
- [Bug: SQL Injection] bugs/sql-injection.md
...
```

---

### Example Custom View: Sprint Planning (with Parameter)

```json
{
  "views": {
    "sprint-planning": {
      "description": "Sprint planning notes with parameter",
      "parameters": [
        {
          "name": "sprint",
          "type": "string",
          "required": false,
          "default": "current",
          "description": "Sprint identifier"
        }
      ],
      "query": {
        "conditions": [
          {
            "field": "data.sprint",
            "operator": "=",
            "value": "{{sprint}}"
          }
        ],
        "order_by": "data.priority DESC"
      }
    }
  }
}
```

**Usage**:
```bash
# Use default sprint value
opennotes notes view sprint-planning

# Specify sprint parameter
opennotes notes view sprint-planning --param sprint=2026-Q1-S3
```

---

## Storage Hierarchy & Precedence

### Precedence Order

Views are discovered in the following order (highest to lowest precedence):

1. **Notebook-specific** (`<notebook>/.opennotes.json`) - Highest precedence
2. **Global config** (`~/.config/opennotes/config.json`) - Medium precedence
3. **Built-in views** (hardcoded in Go) - Lowest precedence

**Rationale**: Notebook-specific views override global config, which overrides built-in views. This allows teams to customize built-in views for project-specific needs.

---

### Discovery Algorithm

```go
func (vs *ViewService) GetView(name string) (*ViewDefinition, error) {
    // 1. Check notebook-specific views (if in notebook context)
    if vs.notebookPath != "" {
        if notebookView := vs.loadNotebookView(name); notebookView != nil {
            vs.logger.Debug("Found view in notebook config", "name", name)
            return notebookView, nil
        }
    }
    
    // 2. Check global config views
    if globalView := vs.loadGlobalView(name); globalView != nil {
        vs.logger.Debug("Found view in global config", "name", name)
        return globalView, nil
    }
    
    // 3. Check built-in views
    if builtinView := vs.loadBuiltinView(name); builtinView != nil {
        vs.logger.Debug("Found built-in view", "name", name)
        return builtinView, nil
    }
    
    return nil, fmt.Errorf("view not found: %s", name)
}
```

---

### Example Precedence Scenario

**Built-in View** (`kanban`):
```json
{
  "description": "Notes grouped by status column",
  "parameters": [{"name": "status", "default": "backlog,todo,done"}],
  "query": { "conditions": [...] }
}
```

**Global Config Override** (`~/.config/opennotes/config.json`):
```json
{
  "views": {
    "kanban": {
      "description": "My custom kanban view",
      "parameters": [{"name": "status", "default": "todo,in-progress,done"}],
      "query": { "conditions": [...] }
    }
  }
}
```

**Notebook Config Override** (`<notebook>/.opennotes.json`):
```json
{
  "views": {
    "kanban": {
      "description": "Team kanban board",
      "parameters": [{"name": "status", "default": "planning,dev,qa,deployed"}],
      "query": { "conditions": [...] }
    }
  }
}
```

**Result**: When `opennotes notes view kanban` is invoked in the notebook, the **notebook config version** is used with status values `planning,dev,qa,deployed`.

---

## Template Variable System

### Supported Variables

| Variable | Resolves To | Example | Use Case |
|----------|-------------|---------|----------|
| `{{today}}` | Current date (YYYY-MM-DD) | `2026-01-20` | Daily notes, recent edits |
| `{{yesterday}}` | Yesterday's date (YYYY-MM-DD) | `2026-01-19` | Yesterday's work |
| `{{this_week}}` | Start of current week (Monday) | `2026-01-19` | Weekly reviews |
| `{{this_month}}` | Start of current month (YYYY-MM-DD) | `2026-01-01` | Monthly planning |
| `{{now}}` | Current timestamp (RFC3339) | `2026-01-20T22:30:00+10:30` | Precise timestamps |

---

### Resolution Logic

```go
func resolveTemplateVars(value string) string {
    now := time.Now()
    
    replacements := map[string]string{
        "{{today}}":     now.Format("2006-01-02"),
        "{{yesterday}}": now.AddDate(0, 0, -1).Format("2006-01-02"),
        "{{this_week}}": getStartOfWeek(now).Format("2006-01-02"),
        "{{this_month}}": now.Format("2006-01") + "-01",
        "{{now}}":       now.Format(time.RFC3339),
    }
    
    for placeholder, replacement := range replacements {
        value = strings.ReplaceAll(value, placeholder, replacement)
    }
    
    return value
}

func getStartOfWeek(t time.Time) time.Time {
    // Monday as start of week
    offset := (int(time.Monday) - int(t.Weekday()) - 7) % 7
    return t.AddDate(0, 0, offset)
}
```

---

### Example Template Usage

**View Definition**:
```json
{
  "views": {
    "this-week": {
      "description": "Notes created this week",
      "query": {
        "conditions": [
          {
            "field": "data.created",
            "operator": ">=",
            "value": "{{this_week}}"
          }
        ],
        "order_by": "created DESC"
      }
    }
  }
}
```

**Runtime Resolution** (Monday, 2026-01-20):
- `{{this_week}}` → `2026-01-19`
- Query: `WHERE data.created >= '2026-01-19'`

---

## Parameter System

### Parameter Types

| Type | Description | Example Values | Validation |
|------|-------------|----------------|------------|
| `string` | Single string value | `"workflow"`, `"current"` | Length < 256 chars |
| `list` | Comma-separated values | `"todo,in-progress,done"` | Split on `,`, validate each item |
| `date` | ISO date format | `"2026-01-20"` | Validate YYYY-MM-DD format |
| `bool` | Boolean flag | `"true"`, `"false"` | Validate true/false (case-insensitive) |

---

### Parameter Validation

```go
func validateViewParams(view *ViewDefinition, params map[string]string) error {
    // Check required parameters
    for _, param := range view.Parameters {
        if param.Required {
            if _, ok := params[param.Name]; !ok {
                return fmt.Errorf("missing required parameter: %s", param.Name)
            }
        }
    }
    
    // Validate parameter types
    for name, value := range params {
        param := findParameter(view, name)
        if param == nil {
            return fmt.Errorf("unknown parameter: %s", name)
        }
        
        if err := validateParamType(param, value); err != nil {
            return fmt.Errorf("invalid parameter %s: %w", name, err)
        }
    }
    
    return nil
}

func validateParamType(param *ViewParameter, value string) error {
    switch param.Type {
    case "string":
        if len(value) > 256 {
            return fmt.Errorf("string too long (max 256 chars)")
        }
    case "list":
        items := strings.Split(value, ",")
        for _, item := range items {
            if len(strings.TrimSpace(item)) == 0 {
                return fmt.Errorf("empty list item")
            }
        }
    case "date":
        if _, err := time.Parse("2006-01-02", value); err != nil {
            return fmt.Errorf("invalid date format (expected YYYY-MM-DD)")
        }
    case "bool":
        lower := strings.ToLower(value)
        if lower != "true" && lower != "false" {
            return fmt.Errorf("invalid boolean (expected true/false)")
        }
    default:
        return fmt.Errorf("unsupported parameter type: %s", param.Type)
    }
    return nil
}
```

---

### Example Parameter Usage

**View Definition** (with required parameter):
```json
{
  "views": {
    "by-author": {
      "description": "Notes by specific author",
      "parameters": [
        {
          "name": "author",
          "type": "string",
          "required": true,
          "description": "Author name to filter by"
        }
      ],
      "query": {
        "conditions": [
          {
            "field": "data.author",
            "operator": "=",
            "value": "{{author}}"
          }
        ]
      }
    }
  }
}
```

**Valid Usage**:
```bash
opennotes notes view by-author --param author="John Doe"
```

**Invalid Usage** (missing required parameter):
```bash
opennotes notes view by-author
# Error: missing required parameter: author
```

---

## Implementation Details

### Files to Create/Modify

#### New Files

1. **`cmd/notes_view.go`** (or modify existing command based on Q&A decision)
   - Command handler for view invocation
   - Parameter parsing from CLI flags
   - Output rendering

2. **`internal/services/view.go`**
   - `ViewService` for view management
   - View discovery and loading
   - Parameter validation
   - Template variable resolution
   - Query generation from view definitions

3. **`internal/core/view.go`**
   - View data structures (ViewDefinition, ViewParameter, ViewQuery, ViewCondition)
   - View serialization/deserialization

4. **`internal/services/view_test.go`**
   - Comprehensive tests for ViewService
   - Test coverage ≥85%

#### Modified Files

1. **`internal/services/config.go`**
   - Extend ConfigService to load views from global config
   - Add `GetViews()` method
   - Add `GetView(name string)` method

2. **`internal/services/notebook.go`**
   - Extend NotebookService to load views from notebook config
   - Add `GetViews()` method

---

### Key Data Structures

```go
// ViewDefinition represents a named, reusable query preset
type ViewDefinition struct {
    Name        string           `json:"name"`
    Description string           `json:"description"`
    Parameters  []ViewParameter  `json:"parameters,omitempty"`
    Query       ViewQuery        `json:"query"`
}

// ViewParameter represents a dynamic parameter in a view
type ViewParameter struct {
    Name        string `json:"name"`
    Type        string `json:"type"` // "string", "list", "date", "bool"
    Required    bool   `json:"required"`
    Default     string `json:"default,omitempty"`
    Description string `json:"description,omitempty"`
}

// ViewQuery represents the query logic for a view
type ViewQuery struct {
    Conditions []ViewCondition `json:"conditions,omitempty"`
    OrderBy    string          `json:"order_by,omitempty"`
    GroupBy    string          `json:"group_by,omitempty"`
    Limit      int             `json:"limit,omitempty"`
}

// ViewCondition represents a single query condition
type ViewCondition struct {
    Logic    string `json:"logic,omitempty"` // "AND", "OR"
    Field    string `json:"field"`
    Operator string `json:"operator"` // "=", "!=", "<", ">", "<=", ">=", "LIKE", "IN", "IS NULL"
    Value    string `json:"value"`
}
```

---

### Query Generation Example

```go
func (vs *ViewService) GenerateSQL(view *ViewDefinition, params map[string]string) (string, error) {
    // Validate parameters
    if err := validateViewParams(view, params); err != nil {
        return "", err
    }
    
    // Apply defaults for missing optional parameters
    resolvedParams := applyDefaults(view, params)
    
    // Resolve template variables
    resolvedParams = resolveTemplateVars(resolvedParams)
    
    // Build SQL query
    var conditions []string
    for _, cond := range view.Query.Conditions {
        // Resolve parameter placeholders
        value := cond.Value
        if strings.HasPrefix(value, "{{") && strings.HasSuffix(value, "}}") {
            paramName := strings.Trim(value, "{}")
            if paramValue, ok := resolvedParams[paramName]; ok {
                value = paramValue
            }
        }
        
        // Whitelist field names and operators
        if err := validateField(cond.Field); err != nil {
            return "", err
        }
        if err := validateOperator(cond.Operator); err != nil {
            return "", err
        }
        
        // Build condition SQL
        condSQL := fmt.Sprintf("%s %s ?", cond.Field, cond.Operator)
        conditions = append(conditions, condSQL)
    }
    
    // Combine conditions
    whereClause := strings.Join(conditions, " AND ")
    
    // Build full query
    query := "SELECT * FROM notes"
    if len(whereClause) > 0 {
        query += " WHERE " + whereClause
    }
    if view.Query.OrderBy != "" {
        query += " ORDER BY " + view.Query.OrderBy
    }
    if view.Query.Limit > 0 {
        query += fmt.Sprintf(" LIMIT %d", view.Query.Limit)
    }
    
    return query, nil
}
```

---

## Security Considerations

### Security Model

The Views System uses the **same security model as search queries**:

1. ✅ **Parameterized queries** - ALWAYS use `?` placeholders for user input
2. ✅ **Whitelist field names** - Only allow known fields (data.*, path, created, updated, body)
3. ✅ **Whitelist operators** - Only allow safe operators (=, !=, <, >, <=, >=, LIKE, IN, IS NULL)
4. ✅ **Validate input length** - Prevent DoS attacks (max 256 chars per parameter)
5. ✅ **Audit logging** - Log all view executions for security monitoring

---

### Additional View-Specific Security

1. **Validate view definitions at load time**
   - Check for malicious SQL in view configurations
   - Validate field names and operators before storing

2. **Prevent circular view references**
   - Detect and reject views that reference other views (future feature)

3. **Limit maximum query complexity**
   - Max 10 conditions per view
   - Max 5 parameters per view

4. **Sanitize template variable output**
   - Ensure template variables can't inject SQL
   - Validate date formats before substitution

---

### Security Validation Example

```go
func validateViewDefinition(view *ViewDefinition) error {
    // Validate view name
    if !isValidViewName(view.Name) {
        return fmt.Errorf("invalid view name: %s", view.Name)
    }
    
    // Validate conditions
    if len(view.Query.Conditions) > 10 {
        return fmt.Errorf("too many conditions (max 10)")
    }
    
    for _, cond := range view.Query.Conditions {
        // Whitelist field names
        if err := validateField(cond.Field); err != nil {
            return err
        }
        
        // Whitelist operators
        if err := validateOperator(cond.Operator); err != nil {
            return err
        }
    }
    
    // Validate parameters
    if len(view.Parameters) > 5 {
        return fmt.Errorf("too many parameters (max 5)")
    }
    
    for _, param := range view.Parameters {
        if !isValidParamType(param.Type) {
            return fmt.Errorf("invalid parameter type: %s", param.Type)
        }
    }
    
    return nil
}
```

---

## Testing Requirements

### Test Coverage Target

**≥85% coverage** for all new view functionality

---

### Test Categories

#### 1. Built-in Views Tests

**Coverage**: Each built-in view

**Tests**:
- ✅ `today` view renders correctly
- ✅ `recent` view shows last 20 notes
- ✅ `kanban` view groups by status
- ✅ `untagged` view finds notes without tags
- ✅ `orphans` view finds notes with no incoming links
- ✅ `broken-links` view finds notes with broken references

**Example Test**:
```go
func TestViewService_TodayView(t *testing.T) {
    vs := setupViewService(t)
    
    // Create notes with today's date
    createTestNote(t, "today-1.md", map[string]interface{}{
        "created": time.Now().Format("2006-01-02"),
    })
    
    // Execute today view
    results, err := vs.ExecuteView("today", nil)
    assert.NoError(t, err)
    assert.Len(t, results, 1)
}
```

---

#### 2. Custom View Tests

**Coverage**: User-defined views

**Tests**:
- ✅ Load views from global config
- ✅ Load views from notebook config
- ✅ Precedence hierarchy works correctly (notebook > global > built-in)
- ✅ Invalid view definitions rejected with clear errors

**Example Test**:
```go
func TestViewService_CustomViewPrecedence(t *testing.T) {
    // Setup config with custom "today" view
    globalConfig := `{"views": {"today": {"description": "Global today"}}}`
    notebookConfig := `{"views": {"today": {"description": "Notebook today"}}}`
    
    vs := setupViewServiceWithConfig(t, globalConfig, notebookConfig)
    
    // Get view
    view, err := vs.GetView("today")
    assert.NoError(t, err)
    assert.Equal(t, "Notebook today", view.Description) // Notebook takes precedence
}
```

---

#### 3. Parameter Tests

**Coverage**: Parameter validation and resolution

**Tests**:
- ✅ Required parameters enforced
- ✅ Optional parameters use defaults
- ✅ Parameter type validation (string, list, date, bool)
- ✅ Invalid parameters rejected with clear errors
- ✅ Parameter substitution in queries

**Example Test**:
```go
func TestViewService_RequiredParameter(t *testing.T) {
    view := &ViewDefinition{
        Name: "test",
        Parameters: []ViewParameter{
            {Name: "author", Type: "string", Required: true},
        },
    }
    
    // Missing required parameter should error
    _, err := validateViewParams(view, map[string]string{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "missing required parameter: author")
    
    // Providing required parameter should succeed
    _, err = validateViewParams(view, map[string]string{"author": "John"})
    assert.NoError(t, err)
}
```

---

#### 4. Template Variable Tests

**Coverage**: Template variable resolution

**Tests**:
- ✅ `{{today}}` resolves to current date
- ✅ `{{yesterday}}` resolves to yesterday's date
- ✅ `{{this_week}}` resolves to start of week (Monday)
- ✅ `{{this_month}}` resolves to start of month
- ✅ `{{now}}` resolves to current timestamp
- ✅ Multiple variables in one query
- ✅ Edge cases (timezone, DST transitions)

**Example Test**:
```go
func TestViewService_TemplateVariables(t *testing.T) {
    now := time.Now()
    expected := now.Format("2006-01-02")
    
    result := resolveTemplateVars("{{today}}")
    assert.Equal(t, expected, result)
}
```

---

#### 5. Security Tests

**Coverage**: SQL injection prevention

**Tests**:
- ✅ SQL injection attempts in view queries rejected
- ✅ Field name validation prevents unknown fields
- ✅ Operator validation prevents dangerous operators
- ✅ Malicious view definitions rejected

**Example Test**:
```go
func TestViewService_SQLInjectionPrevention(t *testing.T) {
    view := &ViewDefinition{
        Name: "malicious",
        Query: ViewQuery{
            Conditions: []ViewCondition{
                {Field: "data.tag'; DROP TABLE notes; --", Operator: "=", Value: "test"},
            },
        },
    }
    
    err := validateViewDefinition(view)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid field")
}
```

---

#### 6. Integration Tests

**Coverage**: End-to-end view execution

**Tests**:
- ✅ View execution end-to-end (discovery → validation → execution → output)
- ✅ Config loading and merging (global + notebook)
- ✅ Error handling and messages
- ✅ Performance benchmarks

**Example Test**:
```go
func TestViewService_EndToEnd(t *testing.T) {
    // Setup notebook with test data
    notebook := setupTestNotebook(t)
    createTestNote(t, "note-1.md", map[string]interface{}{"tag": "workflow"})
    
    // Execute view
    cmd := exec.Command("opennotes", "notes", "view", "today")
    cmd.Dir = notebook.Path
    output, err := cmd.CombinedOutput()
    
    assert.NoError(t, err)
    assert.Contains(t, string(output), "note-1.md")
}
```

---

## Performance Targets

| Operation | Target Time | Measurement Method |
|-----------|-------------|--------------------|
| View definition loading | < 5ms | Benchmark with 100 views in config |
| Template variable resolution | < 1ms | Benchmark with all 5 template variables |
| View query execution (simple) | < 20ms | Benchmark with 1,000 notes |
| View query execution (complex) | < 100ms | Benchmark with 10,000 notes + 5 conditions |
| Parameter validation | < 1ms | Benchmark with 5 parameters |

---

### Performance Benchmarks

```go
func BenchmarkViewService_LoadView(b *testing.B) {
    vs := setupViewService(b)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := vs.GetView("today")
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkViewService_ExecuteSimpleView(b *testing.B) {
    vs := setupViewService(b)
    createTestNotes(b, 1000) // 1k notes
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := vs.ExecuteView("recent", nil)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkViewService_ExecuteComplexView(b *testing.B) {
    vs := setupViewService(b)
    createTestNotes(b, 10000) // 10k notes
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := vs.ExecuteView("kanban", map[string]string{
            "status": "todo,in-progress,done",
        })
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

---

## Documentation Updates Needed

### 1. Command Reference

**File**: `docs/commands.md` (or equivalent)

**Content**:
- View command syntax and options
- List of built-in views with examples
- Parameter syntax (`--param key=value`)
- Integration with existing commands

---

### 2. Built-in Views Guide

**File**: `docs/views-builtin.md`

**Content**:
- Description of each built-in view (today, recent, kanban, untagged, orphans, broken-links)
- Use cases and examples
- Output format examples

---

### 3. Custom Views Tutorial

**File**: `docs/views-custom.md`

**Content**:
- How to create custom views
- Global vs notebook-specific views
- Precedence hierarchy
- Step-by-step examples

---

### 4. View Configuration Reference

**File**: `docs/views-config.md`

**Content**:
- Complete JSON schema for view definitions
- Parameter types and validation rules
- Template variable reference
- Configuration locations (global, notebook)

---

### 5. Parameter System Guide

**File**: `docs/views-parameters.md`

**Content**:
- Using and defining parameters
- Parameter types (string, list, date, bool)
- Required vs optional parameters
- Default values

---

### 6. Template Variables Reference

**File**: `docs/views-templates.md`

**Content**:
- All available template variables
- Resolution examples
- Timezone considerations
- Edge cases (DST, leap years)

---

### 7. Examples

**File**: `docs/views-examples.md`

**Content**:
- Common view patterns and use cases
- Team collaboration workflows
- Project management views (kanban, sprint planning)
- Knowledge graph views (orphans, broken-links)

---

## Migration Impact

### Non-Breaking Changes

✅ **Additive feature only** - No changes to existing commands (unless integrated via Q&A decision)  
✅ **No changes to existing data** - No database schema changes required  
✅ **Backward compatible** - Existing workflows continue unchanged

---

### New Capabilities

✅ **Named query presets** - Users can invoke common queries by name  
✅ **Parameterized queries** - Flexible queries with dynamic parameters  
✅ **Custom user views** - User-defined views in global config  
✅ **Notebook-specific views** - Team-shared views via notebook config  
✅ **Template variables** - Automatic date/time resolution

---

### Upgrade Path

**No migration required** - Views are a new feature with no impact on existing data or workflows.

**Onboarding**:
1. Users discover built-in views via `opennotes notes view --help`
2. Users explore custom views by editing `~/.config/opennotes/config.json`
3. Teams share views via `.opennotes.json` in notebook directories

---

## Design Decisions (Q&A Completed 2026-01-22)

All 6 design questions have been resolved. See summary at top of document.

---

### Decision 1: Command Structure ✅

**Selected**: Option A - Dedicated command under notes

**Command**: `opennotes notes view <name> [--param key=value] [--format list|table|json]`

**Rationale**: Clear separation, dedicated help text, consistent with `notes list` and `notes search`.

---

### Decision 2: Output Formatting ✅

**Selected**: Flag-based output format (variant of Option D)

**Implementation**: 
- Views are query-only (no display logic)
- `--format` flag controls output: `list` (default), `table`, `json`
- Kanban "view" is a grouped query; grouping/swimlanes are formatting concerns

**Rationale**: Views are precomposed queries, not UI components. Formatting is orthogonal.

---

### Decision 3: View Definition Scope ✅

**Selected**: Option A - Query-only

**Implementation**: Views define only:
- Conditions (WHERE clauses)
- Order (ORDER BY)
- Limit
- Group by

**Rationale**: Simple, focused, easy to implement. Display logic handled separately.

---

### Decision 4: Broken Links Detection ✅

**Selected**: Option C - Both frontmatter AND body links

**Implementation**:
- Check `data.links` array (frontmatter)
- Parse `[text](path)` syntax (markdown links)
- Parse `[[wikilink]]` syntax (wiki-style links)

**Rationale**: Comprehensive coverage catches all broken references.

---

### Decision 5: Kanban Parameter Handling ✅

**Selected**: Option D - Hybrid fallback chain

**Precedence**:
1. `--param status=...` (if provided)
2. Notebook config `.opennotes.json` `kanban.status` (if defined)
3. Built-in default: `backlog,todo,in-progress,done`

**Rationale**: Maximum flexibility with progressive disclosure.

---

### Decision 6: Orphans Definition ✅

**Selected**: Hybrid fallback with Option C default

**Precedence**:
1. `--param definition=...` (if provided: `no-incoming`, `no-links`, `isolated`)
2. Notebook config `.opennotes.json` `orphans.definition` (if defined)
3. Built-in default: `isolated` (no links AND not tagged/categorized)

**Rationale**: Catches truly forgotten content by default, configurable for other use cases.

---

## Acceptance Criteria (Pending Q&A)

### Must Have (After Q&A Clarification)

- ✅ All 6 built-in views implemented and functional
- ✅ Custom views loadable from global config (`~/.config/opennotes/config.json`)
- ✅ Custom views loadable from notebook config (`.opennotes.json`)
- ✅ 3-tier hierarchy works correctly (notebook > global > built-in)
- ✅ Parameter system functional (validation, defaults, substitution)
- ✅ Template variables resolve correctly ({{today}}, {{yesterday}}, etc.)
- ✅ Security validations in place (field whitelist, operator whitelist, SQL injection prevention)
- ✅ Test coverage ≥85%
- ✅ Documentation complete (7 documents)
- ✅ Performance targets met (see Performance Targets section)

---

### Should Have (After Q&A Clarification)

- ✅ View-specific formatting (if decided in Q&A)
- ✅ Interactive features (if decided in Q&A)
- ✅ Advanced parameter types (if decided in Q&A)

---

### Could Have (Future Enhancements)

- ⏸️ View composition (views that reference other views)
- ⏸️ View templates (reusable query patterns)
- ⏸️ View sharing (export/import views)
- ⏸️ View versioning (track changes to view definitions)

---

### Blocked By

✅ **UNBLOCKED**: All open questions resolved via Q&A discussion on 2026-01-22.

---

## Next Steps

### Step 1: Conduct Q&A Discussion ✅ COMPLETE

1. ✅ Load the `qa-discussion` skill
2. ✅ Prepare questions from Open Questions section
3. ✅ Conduct discussion session with user
4. ✅ Document decisions made

---

### Step 2: Update Specification ✅ COMPLETE

1. ✅ Update this specification with Q&A decisions
2. ✅ Remove ❓ markers and replace with ✅ confirmed decisions
3. ✅ Update command structure examples with actual syntax
4. ✅ Update acceptance criteria based on decisions

---

### Step 3: Review Updated Specification ✅ COMPLETE

1. ✅ Review updated specification with stakeholders
2. ✅ Validate all decisions align with project goals
3. ✅ Confirm no conflicts with existing features

---

### Step 4: Approve for Implementation ✅ COMPLETE

1. ✅ Get human approval of final specification
2. ✅ Mark status as `approved` (change from `draft`)
3. ⏳ Create implementation task breakdown (NEXT)

---

### Step 5: Implementation ⏳ READY TO BEGIN

1. ⏳ Create detailed implementation tasks
2. ⏳ Assign tasks to implementation phase
3. ⏳ Begin implementation work

---

## ✅ SPECIFICATION APPROVED

**Ready for implementation.** All questions resolved via Q&A discussion on 2026-01-22.

---

**Specification Status**: 🟢 **APPROVED - READY FOR IMPLEMENTATION**

**Created**: 2026-01-20T23:47:00+10:30  
**Last Updated**: 2026-01-22T12:45:00+10:30  
**Author**: Claude (AI Assistant)  
**Q&A Completed**: 2026-01-22  
**Reviewer**: ✅ Human approved via Q&A discussion
