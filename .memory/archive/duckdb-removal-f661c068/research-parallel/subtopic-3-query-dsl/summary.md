# Query DSL Design Patterns - Executive Summary

**Research Date**: 2026-02-01  
**Researcher**: Claude (with golang-pro, api-designer, deep-researcher, create-cli skills)  
**Research Duration**: ~2.5 hours  
**Sources Analyzed**: 25+ tools, libraries, and resources  
**Confidence Level**: HIGH (60% findings verified from multiple authoritative sources)

---

## One-Sentence Summary

Gmail-style field qualifier syntax (`tag:work status:todo`) is the proven, universally-adopted pattern for search query DSLs and should be OpenNotes' foundation, implemented with the Participle parser library in Go for maximum maintainability.

---

## Core Recommendation

**Adopt Gmail/GitHub Search Pattern** for OpenNotes query DSL:
- Field qualifiers: `tag:work`, `title:meeting`, `created:2024-01-01`
- Implicit AND: `tag:work status:todo` (space between terms)
- Negation: `-tag:archived` (minus prefix)
- Dates: ISO 8601 format initially (`2024-01-01`), relative dates later (`yesterday`)
- Quotes for phrases: `title:"meeting notes"`

**Why This Pattern**:
- ✅ Familiar to billions of users (Gmail, GitHub, Obsidian use it)
- ✅ Proven at massive scale (performance + usability)
- ✅ Self-documenting (field names explain themselves)
- ✅ Keyboard-friendly (minimal punctuation)
- ✅ Extensible (add new fields without breaking syntax)

**Parser Implementation**: Use **Participle** (Go parser combinator library)
- Go-idiomatic (struct tags define grammar)
- Type-safe AST (Go structs are AST nodes)
- Good error messages (position tracking built-in)
- No code generation (runtime parsing)
- Best balance of simplicity and power for Gmail-style syntax

---

## Key Findings

### Finding 1: Simplicity Wins Over Power
**Evidence**: Notational Velocity (dead simple search) is beloved; Logseq (Datalog queries) has low advanced query adoption.

**Implication**: Design for 80% use case first (simple tag/title/date filters). Add complexity only when users demand it.

### Finding 2: Field-Based Filtering is Universal
**Evidence**: Every successful query DSL surveyed uses field qualifiers (Gmail, GitHub, Obsidian, Lucene, Elasticsearch).

**Implication**: Map note metadata to fields: `tag:`, `title:`, `path:`, `created:`, `modified:`, `links:`, `backlinks:`

### Finding 3: Boolean Precedence is a Usability Trap
**Evidence**: Users consistently misunderstand `A OR B AND C` without parentheses.

**Implication**: Start with implicit AND only (no OR). Add OR with required parentheses later if needed: `(tag:work OR tag:personal) status:todo`

### Finding 4: Error Messages are Part of the DSL
**Evidence**: Rust/Elm compiler success is partly due to excellent error messages.

**Implication**: Invest in error message quality early:
- Show exact position in query
- Explain what went wrong
- Suggest valid syntax
- Provide examples

### Finding 5: Parser Library Choice Has Long-Term Consequences
**Evidence**: Code generation (goyacc) adds build complexity; hand-written parsers are tedious; parser combinators balance both.

**Implication**: Participle is the right choice for OpenNotes (maintainable, Go-idiomatic, sufficient power).

---

## Migration Strategy (SQL → DSL)

**Current State**: OpenNotes uses DuckDB with SQL queries  
**Challenge**: Users may have saved SQL queries

**Translation Examples**:

| SQL | DSL Equivalent |
|-----|----------------|
| `WHERE tag = 'work'` | `tag:work` |
| `WHERE tag = 'work' AND status = 'todo'` | `tag:work status:todo` |
| `WHERE tag != 'archived'` | `-tag:archived` |
| `WHERE created > '2024-01-01'` | `created:>2024-01-01` |
| `WHERE title LIKE '%meeting%'` | `title:meeting` |

**Migration Plan**:
1. Dual support for 2-3 releases (SQL + DSL)
2. Show deprecation warnings for SQL queries
3. Provide migration guide with examples
4. Consider auto-translation for common patterns

---

## Implementation Roadmap

### Phase 1: MVP (Core Syntax)
**Timeline**: 2-3 weeks

**Features**:
- Field qualifiers: `tag:`, `title:`, `path:`, `created:`, `modified:`
- Implicit AND between terms
- Negation with `-` prefix
- Date operators: `>`, `>=`, `<`, `<=`, `..` (range)
- Quoted values for phrases: `"exact match"`
- Comprehensive error messages

**Parser**: Participle (struct tag grammar)

**Testing**: Table-driven tests for all syntax patterns

**Documentation**: Examples and field reference

### Phase 2: Enhanced (Based on Feedback)
**Timeline**: 1-2 weeks (if user demand)

**Features**:
- Relative dates: `yesterday`, `this-week`, `last-month`
- OR operator with required parentheses: `(tag:A OR tag:B)`
- Prefix wildcards: `title:java*`
- New fields: `links:`, `backlinks:`, `wordcount:`, `size:`

### Phase 3: Advanced (Future)
**Timeline**: TBD

**Features**:
- Regex support: `title:/pattern/`
- Fuzzy search: `title:~meeting` (typo tolerance)
- Proximity search: `"word1 word2"~5` (within 5 words)
- Query composition: Combine saved queries
- Interactive query builder (UI)

---

## Syntax Quick Reference

### Basic Queries
```bash
# Full-text search
meeting

# Tag filter
tag:work

# Multiple filters (AND)
tag:work status:todo

# Title search
title:meeting

# Path prefix
path:projects/

# Date filters
created:2024-01-01
modified:>2024-01-01
created:2024-01-01..2024-12-31

# Negation
-tag:archived
tag:work -status:done

# Quoted values (spaces)
tag:"project management"
title:"meeting notes"
```

### Advanced Queries (Phase 2)
```bash
# OR with parentheses
(tag:work OR tag:personal) status:todo

# Relative dates
created:yesterday
modified:this-week

# Prefix wildcards
title:java*
tag:project-*
```

---

## Technical Specifications

### Parser Grammar (EBNF-style)
```
Query      ::= Clause+
Clause     ::= Field | Negation | Content
Field      ::= FieldName ":" Operator? Value
Negation   ::= "-" Clause
FieldName  ::= "tag" | "title" | "path" | "created" | "modified" | ...
Operator   ::= ">" | ">=" | "<" | "<=" | ".."
Value      ::= QuotedString | UnquotedString | Date
QuotedString ::= '"' .* '"'
UnquotedString ::= [^\s:]+
Date       ::= ISO8601 | RelativeDate
Content    ::= UnquotedString  (no field prefix = full-text search)
```

### Field Types
- **tag**: Exact match (case-insensitive)
- **title**: Substring match (case-insensitive)
- **path**: Prefix match
- **created**: Date comparison
- **modified**: Date comparison
- (content): Full-text search when no field specified

### Date Format
- **Absolute**: ISO 8601 (`2024-01-01`, `2024-01-01T15:30:00Z`)
- **Relative** (Phase 2): `today`, `yesterday`, `this-week`, `last-month`

### Operators
- `:` - Equals or contains (default)
- `:>` - Greater than
- `:>=` - Greater than or equal
- `:<` - Less than
- `:<=` - Less than or equal
- `:..` - Range (inclusive)
- `-` - NOT (negation)

---

## Performance Targets

- **Parse time**: < 1ms for typical query (10-20 tokens)
- **Search time**: 10-100ms for 10k notes
- **Memory**: < 1MB for parser + AST

**Rationale**: Parsing should be negligible compared to search execution.

---

## Risks & Mitigations

### Risk 1: DSL Too Simple
- **Risk**: Users demand features we don't support (complex boolean, regex, etc.)
- **Likelihood**: MEDIUM
- **Mitigation**: Design for extensibility, add features based on user feedback

### Risk 2: SQL Migration Painful
- **Risk**: Users can't translate SQL queries
- **Likelihood**: LOW (most queries are simple)
- **Mitigation**: Dual support period, migration guide, auto-translation tool

### Risk 3: Parser Library Limiting
- **Risk**: Participle can't express complex grammars
- **Likelihood**: LOW (Gmail syntax is simple enough)
- **Mitigation**: Fallback to hand-written recursive descent if needed

---

## Success Metrics

### User Adoption
- **Target**: 80% of users switch from SQL to DSL within 3 months
- **Measure**: Query format usage telemetry (opt-in)

### Usability
- **Target**: New users can write first query in < 30 seconds
- **Measure**: User testing, time-to-first-query metric

### Performance
- **Target**: Parse time < 1ms, search time < 100ms for 10k notes
- **Measure**: Benchmarks, performance tests

### Error Rate
- **Target**: < 5% of queries result in parse errors
- **Measure**: Error rate telemetry (opt-in)

---

## Open Questions (To Resolve)

1. **CLI Integration**: Flag vs stdin vs positional argument?
   - `opennotes notes search "tag:work"`
   - `opennotes notes search --query="tag:work"`
   - `echo "tag:work" | opennotes notes search`

2. **Saved Queries**: Format and storage?
   - JSON with version field?
   - Plain text with metadata?
   - How to handle DSL changes?

3. **Full-Text Search Backend**: DuckDB FTS capabilities?
   - What FTS features does DuckDB support?
   - Can we use DuckDB's FTS or need separate index?

4. **Interactive Query Builder**: Worth building?
   - UI-based filter builder (like Notion)
   - Generates DSL from selections
   - Good for discoverability vs added complexity

---

## Next Actions

### Immediate (Before Implementation)
1. ✅ Research complete (this document)
2. → Write formal syntax specification (EBNF grammar)
3. → Create Participle prototype (validate parser approach)
4. → Benchmark parser performance (< 1ms target)
5. → Design error message system (fixtures for common errors)

### Implementation (Phase 1)
1. Implement parser with Participle
2. Build query AST types (Go structs)
3. Write query executor (AST → search results)
4. Create comprehensive test suite (100+ test cases)
5. Implement error messages with position tracking
6. Document syntax with examples

### Migration (Transition Period)
1. Dual support: SQL + DSL
2. Deprecation warnings for SQL
3. Migration guide documentation
4. Auto-translation for common SQL patterns

---

## Conclusion

The research provides a clear, high-confidence path forward for OpenNotes query DSL design:

**Core Decision**: Adopt Gmail-style field qualifier syntax with Participle parser  
**Why**: Proven at scale, familiar to users, maintainable in Go, extensible for future needs  
**Risk**: Low (well-validated pattern, proven parser library)  
**Timeline**: 2-3 weeks for MVP, iterative enhancement based on user feedback

**Key Success Factor**: Start simple, invest in error messages, extend based on actual user demand (not speculative features).

The research validates that a simple, familiar, well-executed query DSL will serve OpenNotes users better than a complex, powerful but hard-to-use alternative.

---

## Research Artifacts

This research produced 5 comprehensive documents:

1. **thinking.md** - Research methodology, skill discovery, process decisions
2. **research.md** - Raw findings organized by category (25+ sources analyzed)
3. **verification.md** - Source credibility, confidence levels, evidence trails
4. **insights.md** - Synthesized patterns, recommendations, surprising findings
5. **summary.md** - This executive summary

**Total Research Output**: ~90KB of documentation  
**Source Verification**: 60% HIGH confidence, 23% MEDIUM-HIGH confidence  
**Contradictions Found**: 0 (strong consensus on DSL design patterns)

---

**Research Complete**: 2026-02-01  
**Ready for**: Implementation planning and prototype development
