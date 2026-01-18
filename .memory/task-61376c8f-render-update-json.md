---
id: 61376c8f
title: Update RenderSQLResults Function for JSON Output
created_at: 2026-01-18T23:30:00+10:30
updated_at: 2026-01-18T23:30:00+10:30
status: todo
epic_id: a2c50b55
phase_id: e7394efb
assigned_to: current
---

# Task: Update RenderSQLResults Function for JSON Output

## Objective

Modify the existing `RenderSQLResults` function in the display service to output JSON format instead of ASCII tables, while maintaining clean interface design and backward compatibility considerations.

## Steps

### 1. Analyze Current Function Interface
- [ ] Document current `RenderSQLResults` function signature
- [ ] Identify all callers of the function in the codebase
- [ ] Understand input parameters and expected output format
- [ ] Review error handling and return patterns

### 2. Design Format Selection Approach
- [ ] Determine how to specify JSON vs table output format
- [ ] Consider adding format parameter vs. separate function
- [ ] Plan integration with CLI flag parsing
- [ ] Design clean interface that doesn't break existing usage

### 3. Implement JSON Output Path
- [ ] Integrate `renderSQLResultsAsJSON()` from previous task
- [ ] Modify function signature if needed for format selection
- [ ] Update function to call JSON serialization logic
- [ ] Ensure proper error handling and propagation

### 4. Update Function Logic
- [ ] Replace table rendering with JSON output logic
- [ ] Maintain consistent error handling patterns
- [ ] Preserve logging and debugging capabilities
- [ ] Update function documentation and comments

### 5. Verify Integration Points
- [ ] Confirm all callers still work with updated function
- [ ] Test error scenarios and edge cases
- [ ] Validate JSON output format consistency
- [ ] Check memory usage and performance characteristics

## Expected Outcome

**Updated Function**: `RenderSQLResults` produces JSON instead of ASCII tables
- Function interface cleanly supports JSON output
- All existing callers work without modification (if possible)
- JSON serialization integrated into display service flow
- Error handling maintains consistency with service patterns

**Quality Assurance**: 
- Function maintains single responsibility principle
- Error messages are clear and actionable
- JSON output is consistently formatted
- Performance impact is minimal (<2ms additional)

**Integration Ready**:
- Function ready for CLI command integration
- Compatible with existing SQL query execution pipeline
- Proper error propagation through service layer
- Consistent with OpenNotes service architecture patterns

## Actual Outcome

*To be filled upon completion*

## Lessons Learned  

*To be filled upon completion*

## Technical Notes

### Current Function Analysis
```go
// Current function (to be analyzed)
func (d *Display) RenderSQLResults(/* current parameters */) error {
    // Current implementation renders ASCII table
    // Need to replace with JSON output
}
```

### Design Decisions

**Option 1: Modify Existing Function**
```go
func (d *Display) RenderSQLResults(/* params */) error {
    // Direct replacement with JSON output
}
```

**Option 2: Format Parameter Approach**
```go
func (d *Display) RenderSQLResults(/* params */, format string) error {
    switch format {
    case "json":
        return d.renderSQLResultsAsJSON(/* ... */)
    case "table":
        return d.renderSQLResultsAsTable(/* ... */)
    }
}
```

**Recommendation**: Option 1 (direct replacement) for simplicity, since we're moving away from table format entirely.

### Integration Considerations
- CLI command needs to call updated function
- Error handling should be consistent with other service methods
- JSON output should be written directly to output stream
- Function should remain testable with unit tests

### Performance Requirements
- JSON serialization should not significantly impact query performance
- Memory usage should be reasonable for typical result set sizes
- Error handling should not add unnecessary overhead