// Package search provides interfaces and types for the OpenNotes search system.
//
// This package defines the contracts for indexing and querying notes without
// depending on any specific search backend. Implementations (like Bleve) are
// provided in subpackages.
//
// Key interfaces:
//   - Index: Core search index operations (add, remove, find)
//   - Parser: Query string parsing to AST
//   - Storage: Filesystem abstraction (afero-compatible)
//
// Design principles:
//   - Small, focused interfaces (single responsibility)
//   - Functional options for query building (immutable, chainable)
//   - Context support for cancellation
//   - Pure Go, no CGO dependencies
package search
