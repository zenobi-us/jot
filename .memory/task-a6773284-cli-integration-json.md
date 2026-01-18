---
id: a6773284
title: Integrate JSON Output with CLI Command
created_at: 2026-01-18T23:30:00+10:30
updated_at: 2026-01-18T23:30:00+10:30
status: todo
epic_id: a2c50b55
phase_id: e7394efb
assigned_to: current
---

# Task: Integrate JSON Output with CLI Command

## Objective

Connect the updated JSON output functionality to the CLI command that handles the `--sql` flag, ensuring users can execute SQL queries and receive JSON results through the command line interface.

## Steps

### 1. Identify Current CLI Integration
- [ ] Locate the CLI command that handles `--sql` flag (likely in `cmd/notes_search.go`)
- [ ] Understand current flow from flag parsing to SQL execution
- [ ] Document how `RenderSQLResults` is currently called
- [ ] Review existing error handling in CLI command

### 2. Update CLI Command Integration  
- [ ] Modify CLI command to call updated `RenderSQLResults` function
- [ ] Ensure proper parameter passing to display service
- [ ] Update error handling for JSON serialization failures
- [ ] Verify output goes to correct stream (stdout)

### 3. Test CLI Integration
- [ ] Test basic SQL query execution with JSON output
- [ ] Verify error scenarios work correctly (invalid SQL, JSON failures)
- [ ] Check that output formatting is clean and readable
- [ ] Confirm no regressions in non-SQL command functionality

### 4. Validate Output Format
- [ ] Ensure JSON is properly formatted for command line use
- [ ] Test piping to external tools (jq, file redirection)
- [ ] Verify UTF-8 handling for markdown content
- [ ] Check output consistency across different query types

### 5. Error Message Integration
- [ ] Ensure JSON serialization errors show helpful messages
- [ ] Verify CLI error handling remains consistent
- [ ] Test edge cases (empty results, malformed data)
- [ ] Confirm error codes are appropriate for scripting

## Expected Outcome

**Functional CLI Integration**: `--sql` flag produces JSON output
- Command: `opennotes notes search --sql "SELECT title, path FROM notes LIMIT 5"`
- Output: Valid JSON array with query results
- Errors: Clear, actionable error messages for failures
- Performance: Quick response time matching current table output

**User Experience**: Seamless transition from table to JSON format
- JSON output is immediately usable with common tools
- Error messages help users debug query and data issues
- Output format is consistent and predictable
- Command interface remains intuitive and familiar

**Quality Standards**:
- All existing non-SQL functionality unaffected
- JSON output validates with standard JSON parsers
- Error handling follows OpenNotes CLI patterns
- Performance meets existing standards

## Actual Outcome

*To be filled upon completion*

## Lessons Learned

*To be filled upon completion*

## Technical Notes

### Current CLI Command Analysis
- **Location**: Likely `cmd/notes_search.go` with `--sql` flag handling
- **Current Flow**: Parse flag → Execute SQL → Call RenderSQLResults → Output table
- **Target Flow**: Parse flag → Execute SQL → Call RenderSQLResults → Output JSON

### Integration Points
```go
// Current CLI code (to be verified)
if sqlQuery != "" {
    // Execute SQL query
    // Call display service with results
    err := displayService.RenderSQLResults(/* params */)
    if err != nil {
        return fmt.Errorf("failed to render results: %w", err)
    }
}
```

### Testing Approach
```bash
# Basic functionality test
opennotes notes search --sql "SELECT 'hello' as greeting"

# Expected output:
# [{"greeting": "hello"}]

# Real query test  
opennotes notes search --sql "SELECT title, path FROM notes LIMIT 3"

# Expected output:
# [
#   {"title": "Note 1", "path": "/path/note1.md"},
#   {"title": "Note 2", "path": "/path/note2.md"},
#   {"title": "Note 3", "path": "/path/note3.md"}
# ]
```

### Error Scenarios to Test
- Invalid SQL syntax
- JSON serialization failures
- Empty result sets
- Database connection errors
- Permission errors

### Output Validation
- JSON must parse correctly: `echo "$output" | jq .`
- Piping should work: `command | jq '.[] | .title'`
- File output: `command > results.json`
- Encoding: UTF-8 characters in markdown titles/content