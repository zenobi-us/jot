---
id: a2c50b55
title: SQL JSON Output Implementation Epic
created_at: 2026-01-18T23:30:00+10:30
updated_at: 2026-01-18T23:30:00+10:30
status: planning
---

# SQL JSON Output Implementation Epic

## Vision/Goal

Transform the OpenNotes CLI SQL query output from ASCII table format to JSON format, providing developers and power users with machine-readable data that can be easily integrated into scripts, pipelines, and automation workflows.

**Current Problem**: The `--sql` flag outputs data in ASCII table format, which is:
- Difficult to parse programmatically
- Limited by column width constraints  
- Problematic with complex data structures from DuckDB
- Shows ugly Go map formatting for complex types
- Creates poor user experience for automation use cases

**Solution Vision**: Replace ASCII table output with clean, structured JSON that preserves data types and enables seamless automation integration.

## Success Criteria

### Primary Success Criteria
- [ ] **JSON Output Functional**: `--sql` flag outputs valid, parseable JSON instead of ASCII tables
- [ ] **Data Preservation**: All data types and structures correctly represented in JSON format
- [ ] **Type Accuracy**: DuckDB data types properly converted to JSON equivalents
- [ ] **Complex Structure Support**: Maps, arrays, and nested data structures rendered correctly
- [ ] **CLI Consistency**: JSON output format aligns with other OpenNotes CLI patterns

### Quality Gates
- [ ] **Zero Regressions**: All existing SQL functionality remains intact
- [ ] **Test Coverage**: ≥90% test coverage for JSON output functionality  
- [ ] **Performance**: JSON serialization adds <5ms overhead to query execution
- [ ] **Error Handling**: Graceful error messages for JSON serialization failures
- [ ] **Documentation**: Complete CLI help and user guide updates

### User Experience Goals
- [ ] **Developer Friendly**: JSON structure optimized for programmatic consumption
- [ ] **Script Integration**: Easy piping to `jq`, file output, and automation workflows
- [ ] **Backward Compatibility**: Clear migration path from table format (if needed)
- [ ] **Error Clarity**: JSON serialization errors clearly explained to users

## Phases

### Phase 1: Core JSON Output Implementation
**Duration**: 2-3 hours  
**Deliverables**:
- JSON serialization logic for SQL result rows
- Updated RenderSQLResults function
- Basic CLI integration
- Core unit tests

**Start Criteria**: Epic approved and task breakdown completed  
**End Criteria**: Basic JSON output working for simple queries

### Phase 2: Complex Data Type Support  
**Duration**: 2-2.5 hours  
**Deliverables**:
- Support for DuckDB maps, arrays, nested structures
- Proper type conversion and formatting
- Edge case handling
- Comprehensive test coverage

**Start Criteria**: Phase 1 completed and tested  
**End Criteria**: All DuckDB data types correctly represented in JSON

### Phase 3: Polish and Documentation
**Duration**: 1-1.5 hours  
**Deliverables**:
- CLI help text updates
- User guide documentation
- Performance validation
- Final testing and validation

**Start Criteria**: Phase 2 completed with full functionality  
**End Criteria**: Production-ready JSON output with complete documentation

## Dependencies

### Technical Dependencies
- **DuckDB Go Integration**: Must preserve existing SQL query execution pipeline
- **JSON Serialization**: Go's encoding/json package for output formatting
- **Display Service**: Integration with existing service architecture
- **Test Infrastructure**: Leverage existing SQL test framework from previous epic

### Knowledge Dependencies  
- **SQL Flag Implementation**: Understanding of current RenderSQLResults function
- **DuckDB Data Types**: Knowledge of type conversion requirements
- **CLI Patterns**: Consistency with existing OpenNotes command output formats

### External Dependencies
- **None**: Self-contained implementation using standard Go libraries

## Success Metrics

### Functional Metrics
- **Query Compatibility**: 100% of existing SQL queries produce valid JSON
- **Data Integrity**: JSON output matches table format data (where comparable)
- **Performance Impact**: <5ms additional latency for JSON serialization
- **Error Rate**: 0% crashes on valid DuckDB result sets

### Quality Metrics  
- **Test Coverage**: ≥90% for new JSON output functionality
- **Code Quality**: Passes all lint checks and follows Go best practices
- **Documentation Coverage**: All new functionality documented in CLI help and guides

### User Experience Metrics
- **JSON Validity**: 100% of outputs parse correctly with standard JSON parsers
- **Integration Ease**: Successful piping to common tools (jq, file output, scripts)
- **Error Clarity**: Clear error messages for all failure scenarios

## Risks and Mitigations

### Technical Risks
| Risk | Impact | Probability | Mitigation |
|------|---------|-------------|------------|
| DuckDB type conversion issues | High | Medium | Comprehensive type mapping research and testing |
| Performance degradation | Medium | Low | Performance benchmarking and optimization |
| JSON serialization failures | High | Low | Robust error handling and fallback patterns |

### Implementation Risks  
| Risk | Impact | Probability | Mitigation |
|------|---------|-------------|------------|
| Breaking existing workflows | High | Low | Maintain ASCII table option if requested |
| Complex data structure representation | Medium | Medium | Research JSON best practices for SQL data |
| Test complexity | Low | Medium | Leverage existing SQL test framework |

## Related Learning Files

### Implementation Guidance
- `learning-2f3c4d5e-sql-flag-epic-complete.md` - Previous SQL implementation patterns
- `learning-5e4c3f2a-codebase-architecture.md` - Service architecture understanding
- `learning-7d9c4e1b-implementation-planning-guidance.md` - Task breakdown best practices

### Technical References
- `learning-8f6a2e3c-architecture-review-sql-flag.md` - Architecture review for SQL features
- Current SQL implementation in `internal/services/display.go` - RenderSQLResults function

## Notes

**Epic Rationale**: The current ASCII table format creates significant barriers for automation and integration use cases. JSON output will unlock programmatic usage of OpenNotes SQL queries, enabling sophisticated data processing workflows and integration with external tools.

**Implementation Strategy**: Build on proven SQL flag infrastructure from previous epic, focusing on output transformation while preserving all existing query capabilities and security measures.

**Quality Focus**: Maintain the high quality standards established in previous epics with comprehensive testing, clear documentation, and zero regression tolerance.