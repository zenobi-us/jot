# Storage Abstraction Layer Epic - CANCELLED

**Status**: CANCELLED  
**Date**: 2026-01-25  
**Reason**: Technical incompatibility with DuckDB markdown extension

## Why This Was Cancelled

The afero virtual filesystem abstraction library cannot work with DuckDB's markdown extension. DuckDB's extension directly accesses the filesystem for reading markdown files, and afero intercepts filesystem calls at the Go level - but DuckDB's C++ markdown extension makes native filesystem syscalls that bypass Go's abstraction layer.

This creates an impossible situation:
- If we use afero: DuckDB extension cannot read files (breaks core functionality)
- If we bypass afero: We're back to direct filesystem access (defeats purpose of abstraction)

## What Was Attempted

1. **Epic Goal**: Abstract filesystem operations behind an interface to support:
   - In-memory testing
   - Virtual file paths
   - Future cloud storage (S3, etc.)

2. **Research Completed**: 
   - `.memory/archive/cancelled-2026-01-25/research-b8c4d5e6-storage-abstraction-architecture.md`
   - Identified afero as promising solution
   - Missed the DuckDB C++ extension incompatibility during research

3. **Files Involved**:
   - `task-c81f27bd-storage-abstraction-overview.md` - Task planning
   - `epic-a9b3f2c1-storage-abstraction-layer.md` - Epic definition
   - `research-b8c4d5e6-storage-abstraction-architecture.md` - Architecture research

## Key Learning

**DuckDB markdown extension is tightly coupled to native filesystem access.**

The extension is implemented in C++ and makes direct OS syscalls:
- `open()`, `read()`, `stat()` etc. at the kernel level
- Not interceptable by Go's `os` package wrappers
- afero only works for Go-level filesystem calls

## Recommended Alternative Approaches (Future)

If filesystem abstraction is needed later:

1. **DuckDB Extension Wrapper**: Create custom DuckDB extension that respects a virtual filesystem layer
2. **File Staging**: Pre-stage files to actual filesystem before DuckDB processes them (defeats purpose)
3. **Switch Databases**: Use database with plugin/extension system that supports abstraction (overkill)
4. **Accept Direct Filesystem**: Keep current direct access, document limitations

## What Still Works

OpenNotes will continue using direct filesystem access:
- ✅ DuckDB markdown extension functions perfectly
- ✅ No performance penalties
- ✅ Simple, maintainable code
- ✅ No architectural debt

This is acceptable for current use cases. Revisit if cloud storage becomes requirement.

## Files Archived

All related files moved to `.memory/archive/cancelled-2026-01-25/`:
- `epic-a9b3f2c1-storage-abstraction-layer.md`
- `research-b8c4d5e6-storage-abstraction-architecture.md`
- `task-c81f27bd-storage-abstraction-overview.md`

These are preserved for reference but not active work.
