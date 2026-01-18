---
id: 03c6064b
title: Create Comprehensive User Guide for JSON SQL Queries
created_at: 2026-01-18T23:30:00+10:30
updated_at: 2026-01-18T23:30:00+10:30
status: todo
epic_id: a2c50b55
phase_id: af19a341
assigned_to: current
---

# Task: Create Comprehensive User Guide for JSON SQL Queries

## Objective

Develop comprehensive documentation that teaches users how to effectively use the new JSON output format for SQL queries, including practical examples, integration patterns, and troubleshooting guidance.

## Steps

### 1. Research User Workflows and Use Cases
- [ ] Identify common SQL query patterns that benefit from JSON output
- [ ] Research automation workflows that will use JSON data
- [ ] Document integration scenarios with external tools (jq, scripts, APIs)
- [ ] Analyze user feedback on current ASCII table limitations

### 2. Create Core Documentation Structure
- [ ] Design user guide structure covering all aspects of JSON SQL usage
- [ ] Plan progression from basic to advanced examples
- [ ] Include troubleshooting section for common issues
- [ ] Organize integration patterns by tool type

### 3. Develop Practical Examples
- [ ] Create examples for basic SQL queries with JSON output
- [ ] Document complex data type handling (maps, arrays, nested structures)
- [ ] Show integration with jq for data processing
- [ ] Provide automation script templates

### 4. Document Integration Patterns
- [ ] Command line piping examples
- [ ] File output and batch processing patterns
- [ ] Script integration for note management automation
- [ ] API integration patterns for external tools

### 5. Create Troubleshooting Guide
- [ ] Document common error scenarios and solutions
- [ ] Provide debugging guidance for JSON parsing issues
- [ ] Include performance optimization tips
- [ ] Address compatibility considerations

## Expected Outcome

**Comprehensive User Guide**: Complete documentation for JSON SQL feature
- Clear progression from basic to advanced usage
- Practical examples that users can immediately apply
- Integration patterns for common automation scenarios
- Troubleshooting guidance for typical issues

**Example Documentation Sections**:
- **Getting Started**: Basic JSON query examples
- **Data Types**: How complex types appear in JSON output
- **Integration**: Using JSON output with external tools
- **Automation**: Script templates for common tasks
- **Troubleshooting**: Error scenarios and solutions

**Quality Standards**:
- All examples tested and verified to work
- Documentation follows OpenNotes style and standards  
- Clear, actionable guidance for users at all levels
- Integration examples work with common tools

## Actual Outcome

*To be filled upon completion*

## Lessons Learned

*To be filled upon completion*

## Documentation Examples

### Basic Usage Section
```markdown
## Basic JSON SQL Queries

Execute SQL queries with JSON output:

```bash
# Simple query
opennotes notes search --sql "SELECT title, path FROM notes LIMIT 5"

# Output:
[
  {"title": "Project Notes", "path": "projects/main.md"},
  {"title": "Meeting Notes", "path": "meetings/2024-01-15.md"}
]
```
```

### Integration Examples Section
```markdown
## Integration with External Tools

### Using jq for data processing
```bash
# Extract specific fields
opennotes notes search --sql "SELECT title, tags, created_date FROM notes" | jq '.[] | select(.tags | contains("urgent"))'

# Format for reporting
opennotes notes search --sql "SELECT title, word_count FROM notes" | jq -r '.[] | "\(.title): \(.word_count) words"'
```

### Automation Scripts
```bash
#!/bin/bash
# Find notes modified in last 7 days
recent_notes=$(opennotes notes search --sql "SELECT path FROM notes WHERE modified_date > date('now', '-7 days')" | jq -r '.[].path')

for note in $recent_notes; do
  echo "Recently modified: $note"
done
```
```

### Troubleshooting Section
```markdown
## Troubleshooting

### JSON Parsing Errors
If you see "invalid character" errors:
1. Check that your SQL query returns valid data
2. Ensure complex data types are properly handled
3. Use `jq` to validate JSON structure: `command | jq .`

### Performance Issues
For large result sets:
1. Use LIMIT clauses to reduce data size
2. Consider pagination with OFFSET for batch processing
3. Monitor memory usage for complex nested structures
```

### Advanced Examples
```markdown
## Advanced Usage Patterns

### Working with Complex Data Types
```bash
# Query with maps and arrays
opennotes notes search --sql "SELECT title, metadata, tags FROM notes WHERE metadata IS NOT NULL"

# Example output:
[
  {
    "title": "Project Plan",
    "metadata": {"priority": "high", "status": "active"},
    "tags": ["work", "planning", "2024"]
  }
]
```

### Batch Processing
```bash
# Export all notes to individual JSON files
opennotes notes search --sql "SELECT title, content, path FROM notes" | jq -c '.[]' | while read note; do
  filename=$(echo "$note" | jq -r '.title | gsub("[^a-zA-Z0-9]"; "_")').json
  echo "$note" > "exports/$filename"
done
```
```

## Quality Assurance

### Documentation Testing
- [ ] All command examples must execute successfully
- [ ] JSON output examples must be valid JSON
- [ ] Integration patterns must work with specified tools
- [ ] Error scenarios must accurately reflect actual behavior

### User Experience Validation
- [ ] Documentation readable by users with basic SQL knowledge
- [ ] Examples progress logically from simple to complex
- [ ] Troubleshooting covers real issues users will encounter
- [ ] Integration patterns address common automation needs