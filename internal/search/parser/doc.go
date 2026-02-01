// Package parser implements a Gmail-style query parser using Participle.
//
// The parser converts human-readable query strings into structured AST
// that can be used by the search index.
//
// # Syntax
//
// The query language supports:
//
//   - Simple terms: `meeting notes` (searches in all text fields)
//   - Field qualifiers: `tag:work`, `title:meeting`, `path:projects/`
//   - Negation: `-archived`, `-tag:done`
//   - Date comparisons: `created:>2024-01-01`, `modified:<2024-06-30`
//   - Quoted strings: `"exact phrase"`, `title:"project meeting"`
//   - Implicit AND: `tag:work status:todo` (both must match)
//
// # Examples
//
//	parser := parser.New()
//	query, err := parser.Parse("tag:work title:meeting -archived")
//
// # Supported Fields
//
//   - tag: Filter by tag
//   - title: Search in title
//   - body: Search in body only
//   - path: Filter by path prefix
//   - created: Filter by creation date
//   - modified: Filter by modification date
package parser
