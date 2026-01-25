---
id: b8c4d5e6
title: Storage Abstraction Layer - Cancelled (afero DuckDB Incompatibility)
created_at: 2026-01-25
updated_at: 2026-01-25
status: archived
tags: [cancelled, architecture, duckdb, filesystem, learning]
---

# Storage Abstraction Layer - Cancelled

## Summary

The storage abstraction layer epic was cancelled because afero (Go virtual filesystem library) is fundamentally incompatible with DuckDB's markdown extension. DuckDB's C++ extension makes direct OS syscalls that bypass Go's abstraction layer.

## Why This Failed

**Problem**: DuckDB's markdown extension cannot coexist with afero abstraction.

**Root Cause**: 
- afero intercepts filesystem calls at the Go level (`os.Open`, `os.Stat`, etc.)
- DuckDB's markdown extension is written in C++ and makes native syscalls directly to the kernel
- These syscalls cannot be intercepted by Go's abstraction layer
- Result: DuckDB can't read files when afero is in place

**The Impossible Choice**:
- With afero: DuckDB markdown extension breaks (cannot read files)
- Without afero: Direct filesystem access (defeats entire purpose of abstraction)

## What Was Attempted

1. **Epic Definition**: `epic-a9b3f2c1-storage-abstraction-layer.md`
   - Goal: Abstract filesystem operations for in-memory testing and future cloud storage
   - Planned phases: Design, implementation, testing

2. **Research**: `research-b8c4d5e6-storage-abstraction-architecture.md`
   - Identified afero as promising Go library
   - Designed abstraction pattern
   - **Missed**: DuckDB C++ extension incompatibility

3. **Task Planning**: `task-c81f27bd-storage-abstraction-overview.md`
   - Created detailed task breakdown
   - Ready for implementation (but now cancelled)

## Key Learning

**DuckDB extensions are tightly coupled to native filesystem access.**

When considering new database systems or major architectural changes, evaluate:
1. How does the database/extension make filesystem calls? (kernel vs application level)
2. Is the extension pluggable/overrideable? (if abstraction is requirement)
3. What are the C/C++ dependencies that bypass Go abstractions?

## Recommended Alternatives (If Needed Future)

If filesystem abstraction becomes necessary later:

1. **Custom DuckDB Extension**: Build wrapper that respects abstraction layer
2. **File Staging**: Pre-stage files to filesystem before DuckDB processes
3. **Database Switch**: Use system with plugin architecture supporting abstraction
4. **Accept Direct Access**: Keep current approach (simplest, current project acceptable)

## Current State

✅ **OpenNotes continues with direct filesystem access:**
- DuckDB markdown extension works perfectly
- No performance penalties
- Simple, maintainable code
- No architectural debt

This is appropriate for current use cases. Revisit only if cloud storage becomes requirement.

## Files Archived

All related storage abstraction work moved to `.memory/archive/cancelled-2026-01-25/`:
- `epic-a9b3f2c1-storage-abstraction-layer.md`
- `research-b8c4d5e6-storage-abstraction-architecture.md`
- `task-c81f27bd-storage-abstraction-overview.md`

Preserved for reference; not active work.

## Implications

1. **Architecture**: Do not attempt filesystem abstraction in OpenNotes
2. **Future Design**: Evaluate DuckDB extension compatibility early in architectural decisions
3. **Knowledge**: Document DuckDB's tight coupling to native filesystem in future architecture reviews
4. **Cloud Storage**: If needed, would require different approach (not via filesystem abstraction)
