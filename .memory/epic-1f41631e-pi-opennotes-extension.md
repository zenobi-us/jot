---
id: 1f41631e
title: Pi-OpenNotes Extension for pi-mono
created_at: 2026-01-28T23:25:00+10:30
updated_at: 2026-01-29T10:00:00+10:30
status: in-progress
---

# Pi-OpenNotes Extension for pi-mono

## Vision/Goal

Create a **pi-mono extension** that integrates OpenNotes into the pi coding agent, enabling AI assistants to:
- Search and query notes using DuckDB SQL
- Create and manage notes in notebooks
- Execute reusable views
- Access note metadata and relationships

This extension will be written in **Bun/TypeScript** and stored in `pkgs/pi-opennotes/`, following pi package conventions for distribution via npm.

## Success Criteria

### Functional Requirements
- [ ] Extension registers custom tools callable by the LLM
- [ ] Tools support: search notes, list notes, get note, create note, list notebooks, execute view
- [ ] Integration with OpenNotes CLI binary (calls `opennotes` commands)
- [ ] Proper error handling and user-friendly messages
- [ ] Truncation of large outputs to respect context limits

### Quality Requirements
- [ ] TypeScript strict mode enabled
- [ ] Comprehensive test coverage (unit + integration)
- [ ] Documentation with examples
- [ ] Follows pi extension best practices

### Distribution Requirements
- [ ] Package published to npm as `@zenobi-us/pi-opennotes`
- [ ] Works as both global and project-local installation
- [ ] Versioned releases following semver

### Integration Requirements
- [ ] Works with existing OpenNotes binary (no internal API exposure)
- [ ] Respects notebook context (detects current notebook from cwd)
- [ ] Handles multiple notebooks gracefully

## Phases

| Phase | Title | Status | File |
|-------|-------|--------|------|
| 1 | Research & Design | ✅ `complete` | [phase-43842f12-research-design.md](phase-43842f12-research-design.md) |
| 2 | Implementation | ⏳ `pending-review` | [phase-5e1ddedc-implementation.md](phase-5e1ddedc-implementation.md) |
| 3 | Testing & Distribution | 🔜 `proposed` | [phase-16d937de-testing-distribution.md](phase-16d937de-testing-distribution.md) |

## Dependencies

### Technical Dependencies
- **pi coding agent** (`@mariozechner/pi-coding-agent`) - Extension API
- **TypeBox** (`@sinclair/typebox`) - Schema definitions for tool parameters
- **OpenNotes binary** - CLI tool must be installed and accessible in PATH
- **Bun runtime** - For development and testing

### Knowledge Dependencies
- Pi extension API documentation (reviewed ✓)
- Pi packages documentation (reviewed ✓)
- OpenNotes CLI interface (known from development)

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                     pi-opennotes Extension                   │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Tool:       │  │ Tool:       │  │ Tool:               │  │
│  │ search_notes│  │ list_notes  │  │ get_note            │  │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘  │
│         │                │                     │             │
│         └────────────────┼─────────────────────┘             │
│                          │                                   │
│                          ▼                                   │
│                 ┌─────────────────┐                          │
│                 │ OpenNotes CLI   │                          │
│                 │ Adapter Layer   │                          │
│                 └────────┬────────┘                          │
│                          │                                   │
└──────────────────────────┼───────────────────────────────────┘
                           │
                           ▼
                   ┌───────────────┐
                   │  opennotes    │
                   │  CLI binary   │
                   └───────────────┘
                           │
                           ▼
                   ┌───────────────┐
                   │   DuckDB      │
                   │  + Markdown   │
                   └───────────────┘
```

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| OpenNotes CLI not in PATH | High | Medium | Provide clear installation instructions; check on load |
| Large output overwhelms context | Medium | High | Use pi truncation utilities; write to temp files |
| Performance (shell out to CLI) | Low | Medium | Cache notebook config; batch operations |
| API changes in pi | Medium | Low | Pin dependency versions; follow changelogs |

## Notes

- Extension will NOT expose internal Go APIs; it will use the public CLI interface
- This ensures loose coupling and allows independent versioning
- Consider adding MCP server support in future versions
