---
id: f29cef1b
title: Phase 3 - Query Parser
created_at: 2026-02-01T16:30:00+10:30
updated_at: 2026-02-01T16:45:00+10:30
status: complete
epic_id: f661c068
start_criteria: Phase 2 interfaces complete and reviewed
end_criteria: Parser fully implemented with tests, ready for Bleve integration
---

# Phase 3 - Query Parser

## Overview

Implement a Gmail-style query parser using the Participle library. The parser converts human-readable query strings into the AST defined in Phase 2.

**Key Principle**: User-friendly syntax with excellent error messages.

## Deliverables

### 1. Participle-based Parser
Implementation of the `Parser` interface using `alecthomas/participle/v2`:
- Lexer definition for query tokens
- Grammar rules for query syntax
- AST node generation

### 2. Query Syntax Support

**Phase 3 (MVP)**:
- Simple terms: `meeting notes`
- Field qualifiers: `tag:work`, `title:meeting`, `path:projects/`
- Negation: `-archived`, `-tag:done`
- Date comparisons: `created:>2024-01-01`, `modified:<2024-06-30`
- Implicit AND: `tag:work status:todo` (both must match)

**Deferred to Phase 3b**:
- Date ranges: `created:2024-01..2024-06`
- Relative dates: `created:yesterday`, `modified:this-week`
- OR with parentheses: `(tag:work OR tag:personal)`
- Prefix wildcards: `title:java*`

### 3. Error Messages
User-friendly parse errors with:
- Position indication (line, column)
- Clear description of what went wrong
- Suggestions for fixes when possible

### 4. Help Text
Built-in syntax documentation accessible via `parser.Help()`.

## Tasks

| Task | Description | Status |
|------|-------------|--------|
| 1. Add Participle dependency | `go get github.com/alecthomas/participle/v2` | ✅ |
| 2. Define lexer | Token types for query language | ✅ |
| 3. Define grammar | Participle grammar rules | ✅ |
| 4. Implement Parser | `Parse()`, `Validate()`, `Help()` | ✅ |
| 5. AST conversion | Participle AST → search.Query | ✅ |
| 6. Error handling | User-friendly error messages | ✅ |
| 7. Write tests | Comprehensive parser tests | ✅ |
| 8. Integration test | End-to-end query parsing | ✅ |

## Dependencies

- **Phase 2**: [phase-ed57f7e9-interface-design.md](phase-ed57f7e9-interface-design.md) ✅
- **Research**: [research-parallel/subtopic-3-query-dsl/](research-parallel/subtopic-3-query-dsl/)
- **Library**: `github.com/alecthomas/participle/v2`

## Design Decisions

### Package Structure
```
internal/search/
├── parser.go           # Parser interface (from Phase 2)
├── parser/             # NEW - Parser implementation
│   ├── participle.go   # Participle-based parser
│   ├── grammar.go      # Grammar definitions
│   ├── convert.go      # AST conversion
│   ├── errors.go       # Error formatting
│   └── parser_test.go  # Tests
```

### Grammar Design

```ebnf
Query      = { Expression } .
Expression = NotExpr | FieldExpr | Term .
NotExpr    = "-" ( FieldExpr | Term ) .
FieldExpr  = Field ":" Operator? Value .
Field      = "tag" | "title" | "path" | "created" | "modified" | "body" .
Operator   = ">" | ">=" | "<" | "<=" .
Value      = QuotedString | Word .
Term       = QuotedString | Word .
```

### Example Queries

| Query | Parsed AST |
|-------|------------|
| `meeting` | `TermExpr{Value: "meeting"}` |
| `tag:work` | `FieldExpr{Field: "tag", Value: "work"}` |
| `-archived` | `NotExpr{Expr: TermExpr{Value: "archived"}}` |
| `created:>2024-01-01` | `DateExpr{Field: "created", Op: ">", Value: "2024-01-01"}` |
| `tag:work title:meeting` | `[FieldExpr{...}, FieldExpr{...}]` (implicit AND) |

## Next Steps

After Phase 3 completion:
1. Human review of parser
2. Proceed to Phase 4: Bleve Backend

## Notes

- Participle generates parser from Go struct tags (clean, type-safe)
- Start simple, add complexity incrementally
- Error messages are as important as correct parsing
