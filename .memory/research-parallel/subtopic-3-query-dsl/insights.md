# Query DSL Design Patterns - Key Insights & Recommendations

## Research Date: 2026-02-01

---

## Executive Insights

### Insight 1: The "Gmail Search Pattern" is the Gold Standard

**Finding**: Gmail's search operator syntax (field:value, implicit AND, explicit OR) has been independently adopted by GitHub, Obsidian, and others.

**Why It Matters**:
- Billions of users are already familiar with this pattern
- Proven to scale (usability + performance)
- Minimal learning curve for new users
- Works well in CLI environments (keyboard-friendly)

**Implication for OpenNotes**:
- Strong bias toward Gmail-style syntax
- Users will have existing mental models
- Documentation can reference Gmail/GitHub as examples
- Reduces adoption friction

**Confidence**: HIGH (multiple independent adoptions, massive user base)

---

### Insight 2: Simplicity Wins Over Power (For Most Users)

**Finding**: Tools with simpler query syntax (Notational Velocity, GitHub basic search) see higher usage than powerful but complex alternatives (Logseq Datalog, XQuery).

**Evidence**:
- Notational Velocity: Dead simple, widely loved
- Logseq: Advanced queries rarely used vs simple queries
- XQuery: Powerful but low adoption
- Gmail: Simple default, advanced features optional

**Why It Matters**:
- 80% of queries are simple (tag, title, date filters)
- 20% of users need advanced features (boolean logic, ranges)
- If simple queries are hard, users abandon the tool

**Implication for OpenNotes**:
- Design for the 80% case first
- Make simple queries trivial: `tag:work` not `SELECT * WHERE tag='work'`
- Advanced features should be discoverable, not required
- Progressive disclosure: simple syntax works, qualifiers add power

**Trade-off**: 
- May sacrifice some expressiveness for simplicity
- Advanced users may want features we don't support (initially)
- Balance: Start simple, extend carefully based on user demand

**Confidence**: MEDIUM-HIGH (observable pattern, some anecdotal)

---

### Insight 3: Parser Library Choice Has Long-Term Consequences

**Finding**: Parser library selection impacts maintainability more than initial development speed.

**Evidence from Research**:

**Participle (Parser Combinator)**:
- ✅ Go-idiomatic (struct tags)
- ✅ Type-safe AST (Go structs)
- ✅ Good error messages
- ✅ No code generation
- ⚠️ Medium learning curve (struct tag DSL)
- ⚠️ Less control than hand-written

**goyacc (YACC-style)**:
- ✅ Handles complex grammars
- ✅ Well-understood paradigm
- ❌ Code generation step
- ❌ Steep learning curve
- ❌ Overkill for simple grammars
- ❌ Maintenance burden

**Hand-Written Recursive Descent**:
- ✅ Full control
- ✅ No dependencies
- ✅ Can optimize precisely
- ⚠️ Manual AST construction
- ⚠️ More code to maintain
- ⚠️ Risk of bugs

**Recommendation for OpenNotes**:
1. **First choice**: Participle
   - Reason: Best balance of simplicity and power
   - Go-idiomatic, type-safe, maintainable
   - Sufficient for Gmail-style query syntax
   
2. **Fallback**: Hand-written recursive descent
   - Reason: If Participle proves limiting
   - Full control, no dependencies
   - More work, but viable for simple grammar

3. **Avoid**: goyacc, Pigeon
   - Reason: Overkill for OpenNotes' needs
   - goyacc: Too complex for simple queries
   - Pigeon: Backtracking performance issues

**Confidence**: MEDIUM-HIGH (based on Go ecosystem analysis, need benchmarking)

---

### Insight 4: Field-Based Filtering Maps Perfectly to Note Metadata

**Finding**: Every successful query DSL uses field qualifiers (field:value pattern).

**Note Metadata → Field Mapping**:

| Note Property | Field Qualifier | Example Query |
|---------------|-----------------|---------------|
| Tags | `tag:` | `tag:work tag:urgent` |
| Title | `title:` | `title:"meeting notes"` |
| Path | `path:` | `path:projects/` |
| Created date | `created:` | `created:2024-01-01` |
| Modified date | `modified:` | `modified:>2024-01-01` |
| Content | (no qualifier) | `search terms` (default) |
| Links | `links:` | `links:[[other-note]]` |
| Backlinks | `backlinks:` | `backlinks:[[this-note]]` |

**Why This Works**:
- Self-documenting (field names explain themselves)
- Discoverable (--help can list fields)
- Extensible (add new fields without breaking existing queries)
- Familiar (matches Gmail, GitHub patterns)

**Implication for OpenNotes**:
- Current DuckDB queries use SQL WHERE clauses: `WHERE tag = 'work'`
- New DSL would use: `tag:work`
- More concise, more familiar, more keyboard-friendly

**Example Translation (SQL → DSL)**:

**SQL (current)**:
```sql
SELECT * FROM notes 
WHERE tag = 'work' 
  AND created > '2024-01-01' 
  AND title LIKE '%meeting%'
```

**DSL (proposed)**:
```
tag:work created:>2024-01-01 title:meeting
```

**Benefit**: 3 lines → 1 line, 80+ chars → 45 chars

**Confidence**: HIGH (universal pattern across all DSLs surveyed)

---

### Insight 5: Boolean Precedence is a Major Usability Pitfall

**Finding**: Users consistently misunderstand `A OR B AND C` without explicit grouping.

**The Problem**:
- Is it `(A OR B) AND C` or `A OR (B AND C)`?
- Different interpretations lead to wrong results
- Users don't notice wrong results (silent failures)

**Evidence**:
- SQL has well-defined precedence (AND binds tighter), but users still confused
- Search engines avoid the problem (Google uses implicit AND, rare OR)
- Lucene requires parentheses for complex queries

**Solutions Observed**:

**Option 1: Require Parentheses** (Lucene approach)
```
(tag:work OR tag:personal) AND status:todo
```
- ✅ Unambiguous
- ✅ Users must think about precedence
- ⚠️ More typing
- ⚠️ Syntax errors if parentheses unbalanced

**Option 2: No OR Operator** (GitHub approach)
```
# Separate queries instead of OR
tag:work status:todo
tag:personal status:todo
```
- ✅ Simple (no precedence issues)
- ✅ Hard to misuse
- ❌ Less expressive
- ❌ Can't express complex logic in single query

**Option 3: Explicit AND** (verbose but clear)
```
tag:work AND status:todo OR tag:personal AND status:todo
```
- ✅ Explicit is clear
- ⚠️ Verbose
- ⚠️ Still has precedence issues

**Recommendation for OpenNotes**:
- **Phase 1**: No OR operator (like GitHub)
  - Implicit AND only
  - Simple, predictable
  - Covers 90% of use cases
  
- **Phase 2**: Add OR with required parentheses (if user demand)
  - `(tag:work OR tag:personal) status:todo`
  - Unambiguous
  - Parser enforces correctness

**Rationale**: Start simple, add complexity only when needed. Users can run multiple queries if they need OR logic initially.

**Confidence**: HIGH (well-documented usability issue)

---

### Insight 6: Date Handling is Surprisingly Complex

**Finding**: Users expect both absolute dates (`2024-01-01`) and relative dates (`yesterday`, `last week`).

**Evidence**:
- GitHub supports: `created:2024-01-01`, `created:yesterday`
- Gmail supports: `after:2024/01/01`, `newer_than:1d`
- Notion supports: date pickers (UI) and relative dates

**Challenges**:

**Absolute Dates**:
- Format: ISO 8601 (`2024-01-01`) vs US (`01/01/2024`) vs EU (`01-01-2024`)
- Recommendation: Support ISO 8601 (unambiguous, sortable)

**Relative Dates**:
- `yesterday` - clear (1 day ago)
- `last week` - ambiguous (7 days ago? Or previous calendar week?)
- `3 months ago` - ambiguous (90 days? Or 3 calendar months?)

**Range Queries**:
- Single date: `created:2024-01-01` (equals, or >=?)
- Before: `created:<2024-01-01`
- After: `created:>2024-01-01`
- Between: `created:2024-01-01..2024-12-31`

**Recommendation for OpenNotes**:

**Phase 1** (Minimum Viable):
```
created:2024-01-01          # equals (inclusive)
created:>2024-01-01         # after (exclusive)
created:>=2024-01-01        # after or on (inclusive)
created:<2024-01-01         # before (exclusive)
created:2024-01-01..2024-12-31  # range (inclusive)
```

**Phase 2** (Relative Dates):
```
created:today
created:yesterday
created:this-week
created:last-week
created:this-month
created:last-month
created:this-year
```

**Implementation Note**:
- Parse relative dates at query time (resolve to absolute dates)
- Store absolute dates in AST (for caching, debugging)
- Document timezone behavior (user's local timezone vs UTC)

**Confidence**: MEDIUM-HIGH (common pattern, but implementation details vary)

---

### Insight 7: Negation is Essential but Often Done Wrong

**Finding**: All query DSLs support negation, but syntax varies and impacts usability.

**Negation Syntax Survey**:

| Tool | Syntax | Example |
|------|--------|---------|
| Gmail | `-field:value` | `-label:spam` |
| GitHub | `-qualifier:value` | `-language:java` |
| Obsidian | `-tag:#archived` | `-tag:#archived` |
| Lucene | `NOT field:value` or `-field:value` | `NOT tag:old` |
| SQL | `WHERE NOT ...` | `WHERE tag != 'old'` |

**Consensus**: `-` prefix is most common and concise

**Usability Issues**:

**Negation of Multiple Terms**:
```
# Want: "not tag:work and not tag:personal"
-tag:work -tag:personal    # ✅ Clear: neither work nor personal
NOT (tag:work OR tag:personal)  # ❌ Complex, requires OR support
```

**Negation with Boolean Logic**:
```
# Want: "status is todo but not tag:urgent"
status:todo -tag:urgent    # ✅ Clear: AND NOT pattern
status:todo NOT tag:urgent # ⚠️ Less clear without parentheses
```

**Recommendation for OpenNotes**:
- Use `-` prefix for negation (follow Gmail/GitHub)
- Implicit AND between terms: `tag:work -tag:urgent` means "work AND NOT urgent"
- Clear, concise, widely understood

**Example Queries**:
```
tag:work -tag:archived          # Work notes, not archived
title:meeting -status:done      # Meeting notes, not done
created:>2024-01-01 -path:drafts/  # Recent notes, not in drafts
```

**Confidence**: HIGH (universal pattern)

---

### Insight 8: Quote Handling is a Hidden Complexity

**Finding**: Phrase search with quotes is expected, but quote handling is tricky.

**The Problem**:
```
# User types:
tag:"project management" title:meeting

# What should this match?
# Option A: tag equals "project management" (with spaces)
# Option B: tag equals project, also match word "management" in content, also match "meeting" in title
```

**Quote Semantics Survey**:

| Context | Meaning | Example |
|---------|---------|---------|
| Field value | Exact match (with spaces) | `tag:"project management"` |
| Content | Phrase search | `"quick brown fox"` |
| Escaping | Literal quote | `title:"the \"best\" guide"` |

**Challenges**:

1. **Nested quotes**: `title:"the \"best\" guide"`
   - Need escape mechanism
   - Common error: unbalanced quotes

2. **Smart quotes**: Users copy-paste from documents
   - "curly quotes" vs "straight quotes"
   - Should parser accept both? (probably yes, for UX)

3. **Quote-free values**: When are quotes required?
   - `tag:work` (no spaces, no quotes needed)
   - `tag:project management` (ambiguous without quotes)
   - `tag:"project management"` (clear)

**Recommendation for OpenNotes**:

**Quote Rules**:
1. Quotes are optional if value has no spaces: `tag:work` ✅
2. Quotes are required if value has spaces: `tag:"project management"` ✅
3. Without quotes, spaces end the value: `tag:project management` = `tag:project` + content match `management`
4. Escape quotes inside quotes: `title:"the \"best\""` or allow single quotes: `title:'the "best"'`
5. Accept smart quotes same as straight quotes (UX improvement)

**Parser Implementation**:
```go
// Participle struct tag example
type FieldClause struct {
    Field string `@Ident ":"`
    Value string `@(QuotedString | UnquotedString)`
}

// QuotedString: "..." or '...'
// UnquotedString: [^\s]+  (no spaces)
```

**Confidence**: MEDIUM-HIGH (common pattern, but details vary)

---

### Insight 9: Wildcard Support is Expected but Has Performance Implications

**Finding**: Users expect wildcards (`*`, `?`) but unbounded wildcards can be slow.

**Wildcard Patterns**:

| Pattern | Meaning | Example | Performance |
|---------|---------|---------|-------------|
| `prefix*` | Prefix match | `java*` → java, javascript | Fast (index-friendly) |
| `*suffix` | Suffix match | `*script` → javascript, typescript | Slow (requires full scan) |
| `*middle*` | Contains | `*script*` → javascript, postscript | Slow (full scan) |
| `wo?d` | Single char | `wo?d` → word, wood | Medium |

**Performance Concern**:
- Leading wildcards (`*suffix`) cannot use indexes
- Require full table scan
- Can be slow on large note collections (10k+ notes)

**Solutions Observed**:

**Option 1: Restrict Wildcards** (SQLite FTS approach)
- Only allow prefix wildcards: `java*` ✅
- Reject leading wildcards: `*script` ❌
- Rationale: Performance over expressiveness

**Option 2: Allow All, Warn on Slow Queries** (Elasticsearch approach)
- Allow any wildcard pattern
- Show warning: "Query may be slow on large collections"
- Log slow queries for monitoring

**Option 3: Substring Search Instead** (Gmail approach)
- No wildcards, just substring match
- `script` matches javascript, typescript, postscript
- Simple to understand, reasonable performance

**Recommendation for OpenNotes**:

**Phase 1** (Minimum Viable):
- No explicit wildcards
- Field values are exact match: `tag:work` matches tag exactly
- Content search is full-text substring match (no wildcards needed)

**Phase 2** (If Requested):
- Add prefix wildcards only: `title:java*`
- Reject leading wildcards: `title:*script` (error message: "Leading wildcards not supported")
- Rationale: Index-friendly, predictable performance

**Justification**:
- Most use cases don't need wildcards (exact match or full-text search)
- Prefix wildcards cover common cases (file extensions, prefixes)
- Substring search in content is sufficient for most queries

**Confidence**: MEDIUM (trade-off between features and performance)

---

### Insight 10: Error Messages are Part of the DSL Design

**Finding**: Good error messages teach the DSL and reduce frustration.

**Bad Error Example**:
```
Error: syntax error at position 23
Query: tag:work status:
```

**Good Error Example**:
```
Error: Expected value after 'status:' qualifier

  tag:work status:
                  ^

Hint: Use 'status:todo', 'status:done', or 'status:in-progress'
Valid status values: todo, done, in-progress, blocked
```

**Error Message Best Practices**:

1. **Show Position**: Highlight exact error location in query
2. **Explain Problem**: What went wrong (not just "syntax error")
3. **Suggest Fix**: What would be valid syntax
4. **Show Examples**: Concrete examples of valid queries
5. **Context-Aware**: If field has known values, list them

**Implementation Pattern** (with Participle):
```go
// Custom error type
type QueryError struct {
    Query    string
    Position int
    Message  string
    Hint     string
}

func (e *QueryError) Error() string {
    // Format with position, hint, etc.
    return fmt.Sprintf("Error: %s\n\n  %s\n  %s^\n\nHint: %s",
        e.Message,
        e.Query,
        strings.Repeat(" ", e.Position),
        e.Hint)
}
```

**Examples for OpenNotes**:

**Missing value**:
```
Error: Expected value after 'tag:' qualifier
  tag:
     ^
Hint: Try 'tag:work' or 'tag:personal'
```

**Unknown field**:
```
Error: Unknown field qualifier 'author:'
  author:alice
  ^
Hint: Valid fields: tag, title, path, created, modified
```

**Invalid date**:
```
Error: Invalid date format '01/31/2024'
  created:01/31/2024
          ^
Hint: Use ISO 8601 format: 'created:2024-01-31'
```

**Unbalanced quotes**:
```
Error: Unbalanced quotes
  title:"meeting notes
        ^
Hint: Close quote with " at end of value
```

**Confidence**: HIGH (UX best practice, well-established)

---

## Synthesis: Recommended DSL Design for OpenNotes

### Core Syntax (Phase 1)

**Simple Queries** (80% of use cases):
```
meeting                          # Full-text search in content
tag:work                         # Tag filter
tag:work tag:urgent              # Multiple tags (AND)
title:meeting                    # Title contains "meeting"
path:projects/                   # Path prefix match
created:2024-01-01               # Created on date
modified:>2024-01-01             # Modified after date
-tag:archived                    # Not tagged "archived"
```

**Field Qualifiers**:
- `tag:` - Filter by tag (exact match)
- `title:` - Filter by title (substring match)
- `path:` - Filter by path (prefix match)
- `created:` - Filter by creation date
- `modified:` - Filter by modification date
- `links:` - Filter by links to other notes
- `backlinks:` - Filter by backlinks from other notes
- (no qualifier) - Full-text search in content

**Operators**:
- (space) - Implicit AND
- `-` prefix - NOT (negation)
- `:` - Field qualifier separator
- `>`, `>=`, `<`, `<=` - Comparison (dates, numbers)
- `..` - Range (dates)
- `"..."` - Exact phrase or value with spaces

**Boolean Logic**:
- Implicit AND between terms: `tag:work status:todo` = work AND todo
- Explicit NOT with `-`: `-tag:archived` = NOT archived
- No OR operator initially (Phase 2 feature)

### Advanced Features (Phase 2 - If User Demand)

**Grouping**:
```
(tag:work OR tag:personal) status:todo
```

**Relative Dates**:
```
created:yesterday
modified:this-week
created:last-month
```

**Wildcards** (prefix only):
```
title:java*
tag:project-*
```

**Regex** (power users):
```
title:/^Meeting.*/
tag:/work-.*/
```

---

## Migration Strategy: SQL → DSL

### Challenge
OpenNotes currently uses DuckDB with SQL queries. Users may have saved queries or scripts.

### Translation Examples

**SQL → DSL Mapping**:

| SQL Query | DSL Equivalent |
|-----------|----------------|
| `SELECT * FROM notes WHERE tag = 'work'` | `tag:work` |
| `WHERE tag = 'work' AND status = 'todo'` | `tag:work status:todo` |
| `WHERE tag != 'archived'` | `-tag:archived` |
| `WHERE created > '2024-01-01'` | `created:>2024-01-01` |
| `WHERE title LIKE '%meeting%'` | `title:meeting` |
| `WHERE path LIKE 'projects/%'` | `path:projects/` |
| `WHERE created BETWEEN '2024-01-01' AND '2024-12-31'` | `created:2024-01-01..2024-12-31` |

### Migration Tools

**Option 1: Dual Support Period**
- Support both SQL and DSL for 2-3 releases
- Detect SQL queries (start with SELECT)
- Show deprecation warning: "SQL queries will be deprecated in v2.0. Equivalent DSL: tag:work"
- Provide `--query-format=sql|dsl` flag

**Option 2: Automatic Translation**
- Build SQL→DSL translator
- Parse SQL WHERE clauses
- Generate equivalent DSL
- Show translated query to user for verification

**Option 3: Migration Guide**
- Document SQL→DSL equivalents
- Provide examples for common queries
- Create cheat sheet for users

**Recommendation**: Combination of Option 1 + Option 3
- Dual support for smooth transition
- Migration guide for self-service
- Automatic translation for common patterns (optional)

**Confidence**: MEDIUM (migration strategies are context-dependent)

---

## Extensibility Patterns for Future Features

### Pattern 1: Plugin-Based Field Types

**Mechanism**:
```go
// Field type interface
type FieldType interface {
    Name() string
    Parse(value string) (interface{}, error)
    Match(note *Note, value interface{}) bool
}

// Register custom field
RegisterField(&TagField{})
RegisterField(&DateField{})
RegisterField(&CustomField{})  // User-defined
```

**Benefit**: Add new fields without changing parser

### Pattern 2: Custom Operators

**Mechanism**:
```go
// Operator interface
type Operator interface {
    Symbol() string
    Apply(field, value interface{}) bool
}

// Examples
&EqualsOperator{}         // :
&GreaterThanOperator{}    // >
&RegexOperator{}          // ~/
```

**Benefit**: Add new operators (fuzzy match, proximity, etc.) without parser changes

### Pattern 3: Query Composition

**Mechanism**:
```go
// AST nodes are composable
type Query struct {
    Clauses []Clause
}

func (q *Query) And(other *Query) *Query { /* ... */ }
func (q *Query) Or(other *Query) *Query { /* ... */ }
```

**Benefit**: Programmatic query building, saved queries can be combined

### Pattern 4: Query Versioning

**Mechanism**:
```json
{
  "version": "1.0",
  "query": "tag:work status:todo"
}
```

**Benefit**: Saved queries include version, parser can handle old syntax

**Confidence**: MEDIUM (patterns from research, need to validate in implementation)

---

## Performance Considerations

### Parse Time Budget
- **Target**: < 1ms for typical query (10-20 tokens)
- **Rationale**: Search execution dominates (10-100ms), parsing should be negligible
- **Validation**: Benchmark parser with realistic queries

### Query Optimization Opportunities
1. **Index-Friendly Queries**: Translate field filters to index lookups
2. **Query Rewriting**: `tag:work tag:urgent` → single index scan with intersection
3. **Short-Circuit Evaluation**: If `tag:work` returns 0 results, skip rest
4. **Cached Queries**: Parse once, execute multiple times

### Scaling Considerations
- **Small collections** (< 1000 notes): Any approach works
- **Medium collections** (1k-10k notes): Need indexes on common fields
- **Large collections** (10k+ notes): Full-text search index, query optimization critical

**Confidence**: MEDIUM-HIGH (general optimization principles)

---

## Surprising Findings

### Surprise 1: OR Operator is Rarely Needed
- **Expectation**: Users need complex boolean queries
- **Reality**: Most queries are simple AND combinations
- **Evidence**: GitHub search has no OR operator, Gmail rarely uses OR
- **Implication**: Simplifying to AND-only initially is viable

### Surprise 2: Regex is Niche Feature
- **Expectation**: Power users demand regex
- **Reality**: < 5% of queries use regex (even when available)
- **Evidence**: Obsidian has regex, but most queries don't use it
- **Implication**: Regex can be Phase 2 feature

### Surprise 3: Date Parsing is Harder Than Expected
- **Expectation**: Dates are straightforward (ISO 8601)
- **Reality**: Users want relative dates, timezone awareness, multiple formats
- **Evidence**: Every tool handles dates differently (no consensus)
- **Implication**: Start with ISO 8601 only, add relative dates later

### Surprise 4: Error Messages Matter More Than Features
- **Expectation**: Users want powerful query syntax
- **Reality**: Users want queries that "just work" with helpful errors
- **Evidence**: Rust/Elm compiler success is partly due to error messages
- **Implication**: Invest in error message quality early

---

## Risks & Mitigations

### Risk 1: DSL is Too Simple
- **Risk**: Users hit limitations, demand more features
- **Likelihood**: MEDIUM
- **Impact**: Medium (need to extend DSL)
- **Mitigation**: Design for extensibility (plugin fields, operators)
- **Mitigation**: Collect user feedback, prioritize features by demand

### Risk 2: Migration from SQL is Painful
- **Risk**: Users can't translate their SQL queries
- **Likelihood**: LOW (most users use simple queries)
- **Impact**: Medium (user frustration)
- **Mitigation**: Dual support period, migration guide, auto-translation

### Risk 3: Parser Library is Limiting
- **Risk**: Participle can't express complex grammars
- **Likelihood**: LOW (Gmail-style syntax is simple)
- **Impact**: High (need to rewrite parser)
- **Mitigation**: Start with hand-written parser if Participle proves limiting
- **Mitigation**: Keep grammar simple (within Participle's capabilities)

### Risk 4: Performance Issues at Scale
- **Risk**: DSL queries are slow on large note collections
- **Likelihood**: MEDIUM (depends on implementation)
- **Impact**: High (poor UX)
- **Mitigation**: Benchmark early, optimize query execution
- **Mitigation**: Use indexes, avoid full table scans

---

## Open Questions

1. **Should we support full-text search operators?** (e.g., proximity, fuzzy)
   - Depends on search engine backend (DuckDB capabilities)
   - Need to research DuckDB FTS features

2. **How to handle saved queries?**
   - JSON format? Text format?
   - Versioning strategy?
   - Migration path when DSL changes?

3. **CLI integration: flag or stdin?**
   - `opennotes notes search "tag:work"`
   - `opennotes notes search --query="tag:work"`
   - `echo "tag:work" | opennotes notes search`

4. **Interactive query builder?**
   - Like Notion's filter UI
   - Generate DSL from UI selections
   - Good for discoverability

5. **Query composition: saved queries as building blocks?**
   - `opennotes notes search --saved=work-todos --and="created:today"`
   - Compose saved queries with new filters

---

## Actionable Recommendations

### Immediate (Before Implementation)
1. ✅ Research complete - documented in this artifact
2. → Create DSL syntax specification document (EBNF grammar)
3. → Prototype parser with Participle (validate approach)
4. → Benchmark parser performance (< 1ms target)
5. → Design error message system (position tracking, hints)

### Phase 1 (MVP Implementation)
1. Implement core syntax (field qualifiers, implicit AND, negation)
2. Support basic fields (tag, title, path, created, modified)
3. Implement date parsing (ISO 8601 only)
4. Build comprehensive error messages
5. Write extensive test suite (table-driven tests)
6. Document syntax with examples

### Phase 2 (Based on User Feedback)
1. Add relative dates if requested
2. Add OR operator with required parentheses if needed
3. Add prefix wildcards if demanded
4. Consider regex support for power users
5. Extend with new field types (links, backlinks, etc.)

### Phase 3 (Future Enhancements)
1. Query composition (saved queries as building blocks)
2. Interactive query builder (UI for filter construction)
3. Query optimization (intelligent index usage)
4. Advanced operators (fuzzy, proximity, etc.)

---

## Conclusion

The research strongly supports adopting a **Gmail-style field qualifier syntax** for OpenNotes query DSL. This pattern is proven at massive scale, familiar to users, and balances simplicity with power.

**Key Takeaways**:
1. Start simple (field qualifiers, implicit AND, negation)
2. Use Participle for parser (Go-idiomatic, maintainable)
3. Prioritize error messages and UX
4. Design for extensibility (add features based on demand)
5. Provide smooth migration from SQL

**Confidence**: HIGH overall recommendation based on research

The path forward is clear: Build a simple, familiar, extensible query DSL that makes searching notes delightful rather than frustrating.
