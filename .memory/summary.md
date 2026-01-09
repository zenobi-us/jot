# OpenNotes - Project Memory

## Project Overview

OpenNotes is a CLI tool for managing markdown-based notes organized in notebooks. It uses DuckDB for SQL-powered search and supports templates.

## Current Status

- **Phase**: Maintenance
- **Last Updated**: 2026-01-09
- **Next Action**: None (awaiting new work)

## Recent Completed Work

### Go Rewrite (archived: 01-migrate-to-golang)

Successfully completed full migration from TypeScript to Go:

- 5 phases completed (Core Infrastructure, Notebook Management, Note Operations, Polish, Testing)
- 131 tests across all packages
- Full feature parity with TypeScript version
- All commands functional with glamour output

## Archive

| Archive                | Description                      | Completed  |
| ---------------------- | -------------------------------- | ---------- |
| `01-migrate-to-golang` | Full Go rewrite of OpenNotes CLI | 2026-01-09 |

## Key Files

```
cmd/                        # CLI commands (Go)
internal/core/              # Validation, string utilities
internal/services/          # Core services (config, db, notebook, note, display, logger)
internal/testutil/          # Test helpers
tests/e2e/                  # End-to-end tests
main.go                     # Entry point
```

## Memory Structure

```
.memory/
  summary.md                # This file - project overview
  archived/
    01-migrate-to-golang/   # Completed Go migration
      spec.md               # Original specification
      tasks/                # Phase task files (01-05)
      research/             # Testing gap analysis
```
