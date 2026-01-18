---
id: e7394efb
title: Core JSON Output Implementation
created_at: 2026-01-18T23:30:00+10:30
updated_at: 2026-01-18T23:30:00+10:30
status: planning
epic_id: a2c50b55
start_criteria: Epic approved and task breakdown completed
end_criteria: Basic JSON output working for simple queries
---

# Phase 1: Core JSON Output Implementation

## Overview

Implement the foundational JSON output capability by replacing the ASCII table format in the SQL query results. This phase focuses on getting basic JSON output working for simple DuckDB query results while maintaining all existing functionality.

## Deliverables

- [ ] **JSON Serialization Logic**: Core function to convert SQL result rows to JSON format
- [ ] **Updated RenderSQLResults**: Modified display service function with JSON output
- [ ] **CLI Integration**: JSON output working with `--sql` flag
- [ ] **Basic Unit Tests**: Core functionality validated with comprehensive tests
- [ ] **Error Handling**: Graceful handling of JSON serialization failures

## Tasks

### Core Implementation Tasks
1. **[task-core-json]** Implement JSON serialization for SQL result sets
2. **[task-render-update]** Update RenderSQLResults function for JSON output  
3. **[task-cli-integration]** Integrate JSON output with CLI command
4. **[task-error-handling]** Implement robust error handling for JSON failures

### Testing Tasks
5. **[task-unit-tests]** Create comprehensive unit tests for JSON functionality
6. **[task-integration-tests]** Validate end-to-end JSON output with sample queries

## Dependencies

### Technical Dependencies
- Existing `RenderSQLResults` function in `internal/services/display.go`
- Current SQL query execution pipeline from previous epic
- Go's `encoding/json` package for serialization
- DuckDB result handling patterns established in DbService

### Knowledge Dependencies
- Understanding of current SQL result processing flow
- DuckDB data type mapping requirements  
- JSON output format standards and best practices

## Next Steps

After completion of this phase:
1. **Validation**: Verify JSON output works for basic SELECT queries
2. **Testing**: Confirm all existing SQL tests still pass
3. **Transition**: Move to Phase 2 for complex data type support
4. **Review**: Optional human review checkpoint before complex implementation

## Expected Outcome

At phase completion:
- `opennotes notes search --sql "SELECT title, path FROM notes LIMIT 5"` outputs valid JSON
- All existing SQL functionality preserved without regressions
- Foundation established for complex data type support in Phase 2
- Comprehensive test coverage for basic JSON serialization

## Quality Gates

- [ ] JSON output validates with standard JSON parsers
- [ ] Zero regressions in existing SQL query functionality  
- [ ] All new functionality covered by unit tests
- [ ] Performance impact <2ms for basic queries
- [ ] Error messages clear and actionable for common failure scenarios