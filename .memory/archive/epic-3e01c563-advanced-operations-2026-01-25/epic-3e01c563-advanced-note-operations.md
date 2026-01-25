---
id: 3e01c563
title: Advanced Note Creation and Search Capabilities
created_at: 2026-01-20T20:40:00+10:30
updated_at: 2026-01-23T17:30:00+10:30
status: in-progress
progress: 67%
---

# Epic: Advanced Note Creation and Search Capabilities

## Vision/Goal

**Bridge the gap between simple operations and power-user SQL queries** by providing intermediate note creation and search capabilities that offer flexibility without requiring SQL knowledge.

### Current State
OpenNotes has two extremes:
- **Simple commands**: `opennotes notes add`, `opennotes notes list` - easy but limited
- **Power user SQL**: `opennotes notes search --sql` - powerful but requires SQL expertise

### Desired State
Users can:
1. **Create notes with rich metadata** using intuitive CLI flags for frontmatter fields
2. **Search notes with complex queries** using boolean logic and field filtering without SQL
3. **Use reusable search views** for common query patterns (today, kanban, linking)
4. **Fuzzy find notes** like VSCode's Ctrl+P for quick navigation

### Value Proposition
- **Lower barrier to entry** for advanced features (no SQL required)
- **Faster workflows** for common operations (tags, status, links)
- **Better discoverability** through fuzzy search and views
- **Consistent UX** that scales from beginner to expert

## Success Criteria

### User Experience Goals
1. **Creation Workflow**: Users can create notes with frontmatter in one command (no manual YAML editing)
2. **Search Flexibility**: Users can construct complex boolean queries using intuitive flags
3. **Fuzzy Finding**: Users can quickly locate notes using fuzzy matching like VSCode
4. **View System**: Common search patterns are accessible via simple named views

### Technical Goals
1. **Performance**: Search queries execute in <100ms for notebooks with 10,000+ notes
2. **Backwards Compatibility**: Existing commands and SQL flag continue working unchanged
3. **Test Coverage**: ≥85% coverage for all new functionality
4. **Documentation**: Comprehensive examples for all new features

### Quality Metrics
- **Zero regressions** in existing functionality
- **Zero security vulnerabilities** in new query parsing
- **100% cross-platform** compatibility (Linux, macOS, Windows)
- **Clear error messages** for all failure scenarios

## Feature Implementation Status

### Feature 1: Advanced Note Search (Phase 4) ✅ COMPLETE
**Completed**: 2026-01-23  
**Archive**: `archive/phase4-search-implementation-2026-01-23/`  
**Learning**: `learning-8d0ca8ac-phase4-search-implementation.md`

**Deliverables**:
- ✅ Text search with optional search term
- ✅ Fuzzy matching with `--fuzzy` flag
- ✅ Boolean queries (AND/OR/NOT logic)
- ✅ Link queries (`links-to`, `linked-by`)
- ✅ Glob pattern support with security
- ✅ 87% test coverage (exceeded 85% target)
- ✅ Performance: 3-6x better than targets

### Feature 2: Views System (Phases 1-5) ✅ COMPLETE
**Completed**: 2026-01-23  
**Duration**: ~6-7 hours (5 development sessions including discovery)  
**Archive Location**: `archive/phase5-views-system-2026-01-23/`  
**Completion Timestamp**: 2026-01-23 20:15 GMT+10:30

**Phase 1-4 Core Implementation** (Complete):
- ✅ Core data structures (ViewDefinition, ViewParameter, ViewQuery, ViewCondition)
- ✅ ViewService with 6 built-in views (today, recent, kanban, untagged, orphans, broken-links)
- ✅ Template variable resolution ({{today}}, {{yesterday}}, {{this_week}}, {{this_month}}, {{now}})
- ✅ Parameter validation (string, list, date, bool types)
- ✅ 3-tier view hierarchy (notebook > global > built-in)
- ✅ SQL generation with parameterized queries
- ✅ Configuration integration (ConfigService, NotebookService)
- ✅ CLI command (cmd/notes_view.go)
- ✅ Special view executors (broken-links, orphans detection with graph analysis)
- ✅ 59 unit tests (100% ViewService + SpecialViewExecutor coverage)
- ✅ All security validations (field/operator whitelist, parameterized queries)
- ✅ Performance: <1ms query generation (target: <50ms - **50x better**)

**Phase 4: View Discovery Features** (Complete):
- ✅ Comprehensive end-to-end testing with real notebooks
- ✅ Performance validation (all targets exceeded)
- ✅ Edge case handling (empty notebooks, special characters, circular references, unicode)
- ✅ Configuration edge cases (missing configs, malformed YAML, fallback behavior)
- ✅ Link extraction completeness (markdown, wiki-style, frontmatter links)
- ✅ Integration with existing features (notebook context, output piping, regression tests)
- ✅ Parameter validation edge cases and error messaging
- ✅ Output format verification (list, table, json integration)
- ✅ Special view performance for orphans and broken-links detection
- ✅ Zero regressions in 300+ existing tests

**Phase 5: Documentation & Release** (Remaining):
- ⏳ User guide and examples (1 hour)
- ⏳ Code documentation and comments (30 min)
- ⏳ Tutorials and advanced examples (30 min)
- ⏳ CHANGELOG update (20 min)
- ⏳ Total: ~2.5 hours

**Files**:
- Spec: `spec-d4fca870-views-system.md`
- Implementation: `internal/services/view.go`, `internal/services/view_special.go`, `cmd/notes_view.go`
- Tests: `internal/services/view*_test.go` (59 tests)
- Implementation Report: `task-views-phase1-3-complete.md`
- Phase 4 Checklist: `phase5-views-integration.md` (completed)
- Next Steps: `phase5-6-next-steps.md`

### Feature 3: Note Creation Enhancement - READY
**Status**: Specification approved, ready for implementation  
**Spec**: `spec-ca68615f-note-creation-enhancement.md`  
**Estimated**: 4-6 hours

**Planned Deliverables**:
- `--data.*` flags for frontmatter on creation
- Path resolution (file vs folder)
- Title extraction and slugification
- Frontmatter generation

## Dependencies

### Internal Dependencies
- **ConfigService**: May need extension for view aliases
- **NoteService**: Core modification for creation and search
- **DbService**: Query generation and execution
- **DisplayService**: Output formatting for search results

### External Dependencies
- **FZF Integration**: Research Go libraries for fuzzy finding (e.g., github.com/junegunn/fzf, github.com/ktr0731/go-fuzzyfinder)
- **Flag Parsing**: Leverage Cobra's flag system for dynamic --data.* parsing
- **DuckDB**: Ensure markdown extension supports all query patterns

### Knowledge Dependencies
- **Existing Learning**: Reference `.memory/learning-2f3c4d5e-sql-flag-epic-complete.md` for SQL integration patterns
- **Architecture**: Reference `.memory/knowledge-codemap.md` and `.memory/knowledge-data-flow.md`

## Risks & Mitigations

### Risk 1: Flag Parsing Complexity
**Risk**: Dynamic --data.* flag parsing may be complex with Cobra
**Impact**: High - Core feature functionality  
**Likelihood**: Medium  
**Mitigation**: Research phase will evaluate multiple approaches (Cobra dynamic flags, custom parsing, viper integration)

### Risk 2: FZF Integration Challenges
**Risk**: FZF integration in Go may have platform-specific issues
**Impact**: Medium - Feature is nice-to-have, not critical  
**Likelihood**: Medium  
**Mitigation**: Fallback to simple substring filtering if FZF proves too complex

### Risk 3: Query Generation SQL Injection
**Risk**: User-constructed queries could allow SQL injection
**Impact**: Critical - Security vulnerability  
**Likelihood**: Low (with proper validation)  
**Mitigation**: Defense-in-depth validation (same as --sql flag), parameterized queries, whitelist approach

### Risk 4: Performance Degradation
**Risk**: Complex boolean queries may be slow on large notebooks
**Impact**: Medium - User experience  
**Likelihood**: Low  
**Mitigation**: DuckDB is optimized for OLAP queries, benchmark with 10k+ notes, add query optimization if needed

### Risk 5: View System Scope Creep
**Risk**: View system could become overly complex
**Impact**: Low - Can be simplified  
**Likelihood**: Medium  
**Mitigation**: Start with hardcoded views, defer custom view configuration to future epic

## Timeline Estimate

**Research & Design Phase**: 7-10 hours total
- Phase 1 (Creation): 2-3 hours
- Phase 2 (Boolean Search): 2-3 hours
- Phase 3 (FZF & Views): 2-3 hours
- Phase 4 (Planning): 1-2 hours

**Implementation Phase** (separate epic after research):
- Estimated 15-20 hours based on research findings
- Will be broken down in Phase 4

## Related Artifacts

### Existing Learning
- `.memory/learning-2f3c4d5e-sql-flag-epic-complete.md` - SQL flag implementation patterns
- `.memory/learning-5e4c3f2a-codebase-architecture.md` - Architecture reference
- `.memory/knowledge-codemap.md` - Codebase structure
- `.memory/knowledge-data-flow.md` - Data flow patterns

### Future Artifacts
- Research documents (4 files, one per research area)
- Phase files (3 files, one per major phase)
- Task files (created during Phase 4 planning)
- Learning file (created after implementation complete)

## Notes

### Design Philosophy Alignment
This epic aligns with OpenNotes' **progressive disclosure** philosophy:
1. **Simple** → `opennotes notes list` (beginner-friendly)
2. **Intermediate** → Boolean search, views, advanced creation (this epic)
3. **Advanced** → `--sql` flag for power users (already implemented)

### User Personas
- **Casual User**: Wants to create notes with tags quickly
- **Organized User**: Needs to filter by multiple criteria (tags, status, dates)
- **Power User**: Appreciates fuzzy find for rapid navigation
- **Developer**: Uses views for kanban, issue tracking workflows

### Inspiration Sources
- **VSCode**: Ctrl+P fuzzy file finder
- **Obsidian**: Tag search, link filtering
- **Notion**: Database filters and views
- **Taskwarrior**: Boolean query syntax

## Next Steps

1. ✅ **Epic Created**: This document
2. ✅ **Research Complete**: Comprehensive research document created (`.memory/research-3e01c563-advanced-operations.md`)
3. 🔄 **Human Review**: Review research findings before Phase 4 planning
4. ⏸️ **Implementation Planning**: Create detailed task breakdown
5. ⏸️ **Human Approval**: Review implementation plan before execution

## Research Completion Summary

**Date Completed**: 2026-01-20T21:30:00+10:30  
**Research Document**: `.memory/research-3e01c563-advanced-operations.md`  
**Executive Summary**: `.memory/research-3e01c563-summary.md`

### Key Findings

1. **Dynamic Flag Parsing**: Use pflag `StringArray` with custom `field=value` parsing ⭐⭐⭐⭐⭐
2. **FZF Integration**: Use `github.com/ktr0731/go-fuzzyfinder` (pure Go) ⭐⭐⭐⭐
3. **Boolean Queries**: Parameterized queries + whitelist validation ⭐⭐⭐⭐⭐
4. **View System**: YAML configuration with built-in and user views ⭐⭐⭐⭐⭐

All recommendations are based on proven patterns from production CLI tools (kubectl, docker, gh) and include:
- Detailed implementation examples
- Security validation strategies
- Performance optimization techniques
- Testing approaches
- Integration patterns with existing OpenNotes architecture
