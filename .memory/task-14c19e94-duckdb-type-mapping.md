---
id: 14c19e94
title: Research and Implement DuckDB to JSON Type Mapping
created_at: 2026-01-18T23:30:00+10:30
updated_at: 2026-01-18T23:30:00+10:30
status: todo
epic_id: a2c50b55
phase_id: 686d28b6
assigned_to: current
---

# Task: Research and Implement DuckDB to JSON Type Mapping

## Objective

Develop a comprehensive mapping system that converts DuckDB data types to appropriate JSON representations, focusing on the complex types that currently cause problems with ASCII table output (maps, arrays, nested structures).

## Steps

### 1. Research DuckDB Type System
- [ ] Analyze DuckDB documentation for all supported data types
- [ ] Review DuckDB Go driver type handling
- [ ] Identify types currently supported by OpenNotes SQL queries
- [ ] Document problematic types from ASCII table output

### 2. Study Current Type Conversion
- [ ] Examine how DuckDB results are currently processed in `RenderSQLResults`
- [ ] Identify where Go map formatting occurs (the ugly output we're fixing)
- [ ] Review DuckDB markdown extension specific types
- [ ] Document current handling of NULL values and edge cases

### 3. Design JSON Type Mapping
- [ ] Create comprehensive mapping from DuckDB types to JSON representations
- [ ] Design handling for DuckDB MAP types → JSON objects
- [ ] Plan DuckDB ARRAY types → JSON arrays
- [ ] Address nested structures and complex compositions

### 4. Implement Type Conversion Functions
- [ ] Create type detection and conversion logic
- [ ] Implement MAP type conversion to JSON objects
- [ ] Implement ARRAY type conversion to JSON arrays
- [ ] Handle nested and composite data structures

### 5. Address Special Cases
- [ ] Handle NULL values in complex structures
- [ ] Address empty maps and arrays
- [ ] Handle deeply nested structures
- [ ] Manage type coercion edge cases

## Expected Outcome

**Complete Type Mapping System**: All DuckDB types properly converted to JSON
- DuckDB MAP types → Clean JSON objects (not Go map formatting)
- DuckDB ARRAY types → JSON arrays with proper element typing  
- Primitive types → Appropriate JSON primitives
- NULL handling → Consistent JSON null representation

**Conversion Functions**: Robust type conversion implementation
- `convertDuckDBValueToJSON()` function handling all types
- Error handling for unsupported or malformed data
- Performance optimized for typical OpenNotes data sizes
- Comprehensive logging for debugging conversion issues

**Documentation**: Clear mapping reference
- Complete table of DuckDB types to JSON representations
- Examples of complex type conversions
- Edge case handling documentation
- Performance characteristics for each type

## Actual Outcome

*To be filled upon completion*

## Lessons Learned

*To be filled upon completion*

## Technical Notes

### Current Problems to Solve

**DuckDB Map Output** (current ASCII):
```
map[key1:value1 key2:complex_value]  // Ugly Go formatting
```

**Target JSON Output**:
```json
{"key1": "value1", "key2": "complex_value"}
```

**DuckDB Array Output** (current ASCII):
```
[item1 item2 item3]  // Basic array formatting
```

**Target JSON Output**:
```json
["item1", "item2", "item3"]
```

### Research Areas

1. **DuckDB Type System**:
   - Primitive types: INTEGER, VARCHAR, DOUBLE, BOOLEAN, DATE, TIMESTAMP
   - Complex types: MAP, ARRAY, STRUCT
   - Special types: UUID, JSON (native), BLOB

2. **Go Driver Integration**:
   - How types are represented in Go interface{}
   - Type assertions and conversions needed
   - Performance considerations for type detection

3. **JSON Representation Standards**:
   - Best practices for SQL-to-JSON conversion
   - Handling of NULL vs undefined vs empty
   - Date/timestamp formatting conventions

### Implementation Strategy

```go
func convertDuckDBValueToJSON(value interface{}) (interface{}, error) {
    switch v := value.(type) {
    case map[string]interface{}:
        // Convert DuckDB MAP to JSON object
        return convertMapType(v)
    case []interface{}:
        // Convert DuckDB ARRAY to JSON array  
        return convertArrayType(v)
    case time.Time:
        // Convert timestamps to ISO 8601
        return v.Format(time.RFC3339)
    case nil:
        // Handle NULL values
        return nil, nil
    default:
        // Handle primitive types
        return v, nil
    }
}
```

### Performance Targets
- Type conversion should add <1ms per 1000 rows
- Memory usage should remain reasonable for typical result sets
- Complex nested structure conversion should not cause stack overflow