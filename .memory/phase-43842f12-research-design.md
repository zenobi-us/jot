---
id: 43842f12
title: Phase 1 - Research & Design
created_at: 2026-01-28T23:25:00+10:30
updated_at: 2026-01-29T09:55:00+10:30
status: complete
epic_id: 1f41631e
---

# Phase 1: Research & Design

## Status: ✅ COMPLETE

**Completed**: 2026-01-29T09:55:00+10:30  
**Duration**: ~1 hour  
**Review Status**: Awaiting human review before Phase 2

## Objective

Research pi extension patterns and design comprehensive specifications for the pi-opennotes extension.

## Completed Tasks

| Task | Status | Artifact |
|------|--------|----------|
| Document OpenNotes CLI Interface | ✅ Done | [task-a0236e7c](task-a0236e7c-document-opennotes-cli.md) |
| Design Tool API Specification | ✅ Done | [task-4b6f9ebd](task-4b6f9ebd-design-tool-api.md) |
| Define Package Structure | ✅ Done | [task-f8bb9c5d](task-f8bb9c5d-define-package-structure.md) |
| Design Service Architecture | ✅ Done | [task-e1x1x1x1](task-e1x1x1x1-design-service-architecture.md) |
| Design Error Handling Strategy | ✅ Done | [task-e2x2x2x2](task-e2x2x2x2-design-error-handling.md) |
| Design Test Strategy | ✅ Done | [task-e3x3x3x3](task-e3x3x3x3-design-test-strategy.md) |

## Key Design Decisions

### Architecture
- **Service-based**: Fat services, thin tools (SOLID principles)
- **Dependency injection**: Services created via `createServices()` factory
- **CLI adapter**: Central abstraction for all CLI interactions

### Tool Naming
- **Prefix**: `opennotes_` (configurable via extension config)
- **Tools**: search, list, get, create, notebooks, views

### Pagination
- **Strategy**: 75% output capacity + metadata
- **Format**: Includes total, returned, page, pageSize, hasMore, nextOffset

### Error Handling
- **Pattern**: OpenNotesError class with code + hint
- **Installation**: Full installation guide for CLI_NOT_FOUND
- **Recoverable flag**: Helps LLM decide when to ask user

### Testing
- **Pyramid**: 70% unit / 25% integration / 5% E2E
- **Mocking**: CliAdapter mocked for unit tests
- **Coverage target**: 85% overall

## Deliverables

1. **CLI Interface Reference** - Complete command documentation
2. **TypeBox Schemas** - All tool parameters and responses
3. **Package Structure** - Directory layout with service layer
4. **Service Interfaces** - All service contracts defined
5. **Error Codes** - 15+ error codes with hints
6. **Test Matrix** - ~98 tests planned across all levels

## Research Findings

- Not all CLI commands support JSON output (workaround: use SQL mode)
- Views have full JSON support via `--format json`
- `StringEnum` required for TypeBox (Google API compatibility)
- pi truncation utilities available (`truncateHead`, `truncateTail`)

## Dependencies Confirmed

- `@mariozechner/pi-coding-agent` >= 0.50.0
- `@sinclair/typebox` >= 0.32.0
- Bun runtime for development
- OpenNotes CLI binary

## Risks Addressed

| Risk | Mitigation |
|------|------------|
| CLI not in PATH | Comprehensive installation hints |
| Large output | 75% budget with pagination metadata |
| No JSON from some commands | Convert to SQL queries |
| API changes | Pin dependency versions |

## Ready for Phase 2

Phase 1 provides complete specifications for implementation:
- All interfaces defined
- All schemas ready
- Error handling designed
- Test strategy established

**⏳ Awaiting human review before proceeding to Phase 2 (Implementation)**
