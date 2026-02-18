# ZK Search Architecture Analysis - Key Insights & Implications

## Strategic Insights

### 1. The Filesystem Abstraction Already Exists

**Finding**: zk has a clean `FileStorage` interface that is **directly compatible** with afero patterns.

**Significance**:
- OpenNotes doesn't need to invent a new filesystem abstraction
- The `FileStorage` interface (9 methods) maps cleanly to afero.Fs
- zk already solved the "how to abstract filesystem access" problem

**Implication**:
- ✅ **Adopt zk's FileStorage interface definition**
- ✅ Implement it with afero as the backing store
- ✅ This enables testing with `afero.MemMapFs` out of the box

**Caveat**: The `paths.Walk()` utility currently uses OS `filepath.Walk` directly, bypassing the interface. This would need refactoring.

---

### 2. SQLite is a Blocker, Not a Feature

**Finding**: zk's entire search architecture is built on SQLite + FTS5, requiring CGO.

**Pattern Observed**:
```go
// This pattern appears throughout the codebase:
db.Query("SELECT ... FROM notes_fts WHERE notes_fts MATCH ?", ftsQuery)
```

**Why This Matters**:
- SQLite's FTS5 extension is **NOT** portable to WASM
- CGO prevents cross-compilation to many platforms
- Pure-Go alternatives exist (Bleve, tantivy-go) but require complete rewrite
- Can't just "swap out" SQLite - it's deeply integrated via triggers, views, CTEs

**Strategic Decision Required**:
- ❌ Cannot adopt zk's implementation directly
- ✅ Can adopt zk's **interface design** (`NoteIndex`)
- ✅ Must implement from scratch with pure-Go search engine

---

### 3. The Query DSL is Gold

**Finding**: zk's Google-like query syntax is well-designed and **implementation-agnostic**.

**Examples of Excellence**:
- `foo bar` → AND search (intuitive)
- `foo | bar` → OR search (familiar operator)
- `-foo` → NOT search (natural prefix)
- `title:foo` → Field-specific search (advanced but discoverable)
- `foo*` → Prefix search (wildcard support)

**Why This is Valuable**:
- Users already know this syntax from Google/DuckDuckGo
- Implementation is separate from interface (`fts5.ConvertQuery()`)
- Can reuse the **syntax** even if backend changes from FTS5 to Bleve

**Recommendation**:
- ✅ **Adopt the query syntax verbatim**
- ✅ Rewrite `ConvertQuery()` to target different backend
- ✅ Keep the same user-facing CLI flags (`--match`, `--match-strategy`)

---

### 4. BM25 Ranking is a Competitive Advantage

**Finding**: zk uses BM25 algorithm with weighted fields (path=1000, title=500, body=1).

**Why BM25 Matters**:
- Industry-standard relevance ranking (used by Lucene, Elasticsearch)
- Better than TF-IDF for short documents (like notes)
- Handles term frequency saturation naturally

**Implementation Challenge**:
- SQLite FTS5 has `bm25()` built-in
- Pure-Go implementations exist in Bleve, tantivy-go
- Could implement manually (well-documented algorithm)

**Insight**:
- ✅ Field weighting strategy (path > title > body) is smart
- ✅ Users expect title matches to rank higher than body matches
- ✅ This is **table stakes** for a modern search tool

**Recommendation**:
- ✅ Implement BM25 or similar (Okapi BM25, BM25F)
- ✅ Use the same field weights as zk (proven effective)

---

### 5. Link Analysis is Underutilized in Note-Taking Tools

**Finding**: zk supports advanced link queries:
- `--linked-by` - Notes that link to X
- `--link-to` - Notes that X links to
- `--related` - Notes 2 hops away in the link graph
- Recursive traversal with max distance limiting

**Why This is Unique**:
- Most note-taking tools treat links as simple references
- zk enables graph-based discovery (Zettelkasten method)
- Transitive closure queries are computationally expensive but powerful

**Implementation in zk**:
```sql
-- Recursive link traversal using transitive_closure
SELECT * FROM transitive_closure WHERE distance = 2
```

**Insight**:
- Graph databases (Neo4j, etc.) excel at this, but are overkill
- SQLite's recursive CTEs (Common Table Expressions) handle it adequately
- Pure-Go graph libraries exist (gonum/graph) for implementing this

**Recommendation**:
- ✅ **Must-have feature** for serious note-taking tool
- ✅ Implement using graph algorithms (BFS/DFS with distance limiting)
- ⚠️ Performance can degrade with large link graphs (needs testing)

---

### 6. The Migration System Teaches Valuable Lessons

**Finding**: zk uses `PRAGMA user_version` for schema versioning with incremental migrations.

**Pattern**:
```go
migrations := []struct {
    SQL []string
    NeedsReindexing bool
}{
    { SQL: [...], NeedsReindexing: true },  // Migration 1
    { SQL: [...], NeedsReindexing: false }, // Migration 2
}
```

**Why This is Clever**:
- Each migration knows if it requires full reindex
- Version tracking prevents duplicate migrations
- Atomic: either all SQL statements succeed or none

**Implication for OpenNotes**:
- ✅ Need migration system from day one
- ✅ Track schema version in metadata (not just code)
- ✅ Reindexing flag prevents slow startup surprises

**Anti-Pattern Observed**:
- No rollback mechanism (only forward migrations)
- Could cause issues if migration fails mid-flight

---

### 7. Channel-Based Streaming Prevents Memory Bloat

**Finding**: zk uses Go channels for result streaming:

```go
IndexedPaths() (<-chan paths.Metadata, error)
```

**Why This Pattern Works**:
- Listing 10,000 notes doesn't load all into memory
- Consumer controls pace (backpressure)
- Enables lazy evaluation

**Performance Implication**:
- For large notebooks (>10k notes), this is essential
- Single-threaded iteration is fine (disk I/O is bottleneck)

**Recommendation**:
- ✅ **Use channels for all large result sets**
- ✅ Pair with context for cancellation (`context.Context`)
- ✅ Enables streaming JSON output in CLI

---

### 8. Tags are First-Class Citizens

**Finding**: zk stores tags in a normalized `collections` table, not as text fields.

**Schema Design**:
```sql
CREATE TABLE collections (
    id INTEGER PRIMARY KEY,
    kind TEXT NOT NULL,  -- 'tag', 'category', etc.
    name TEXT NOT NULL,
    UNIQUE(kind, name)
)

CREATE TABLE notes_collections (
    note_id INTEGER REFERENCES notes(id),
    collection_id INTEGER REFERENCES collections(id)
)
```

**Why This Matters**:
- Prevents tag duplication ("Project" vs "project")
- Enables tag renaming globally (change one row)
- Supports multiple collection types (tags, categories, authors)

**Insight**:
- Tag queries use GLOB matching: `t.name GLOB 'work*'`
- OR queries: `work OR personal` parsed and executed efficiently
- Negation: `-work` excludes notes with that tag

**Recommendation**:
- ✅ **Normalize tags** (don't store as JSON arrays in note metadata)
- ✅ Support hierarchical tags (`work/project/client`)
- ✅ Enable tag autocomplete (SELECT DISTINCT from collections)

---

### 9. Snippet Generation is Context-Aware

**Finding**: zk generates snippets differently based on query type:

- **FTS queries**: Use `snippet()` function (highlights matches)
- **Link queries**: Use link snippet from surrounding text
- **Default**: Use `lead` (first paragraph)

**SQL Example**:
```sql
snippetCol := `snippet(fts_match.notes_fts, 2, '<zk:match>', '</zk:match>', '…', 20)`
```

**Parameters**:
- Column 2 = body (not path or title)
- Highlight with `<zk:match>` tags
- Max 20 tokens per snippet

**Insight**:
- Users want to see **why** a note matched
- Highlighting improves scan-ability in CLI output
- Snippet length (20 tokens) balances context vs verbosity

**Recommendation**:
- ✅ Implement context-aware snippet generation
- ✅ Use similar highlight markers (XML-style tags)
- ✅ Make snippet length configurable (flag or config)

---

### 10. Lazy Prepared Statements Optimize Cold Starts

**Finding**: zk uses `LazyStmt` to defer SQL preparation until first use.

**Pattern**:
```go
type LazyStmt struct {
    sql  string
    stmt *sql.Stmt
    tx   Transaction
}

func (s *LazyStmt) Query(args ...any) (*sql.Rows, error) {
    if s.stmt == nil {
        s.stmt, _ = s.tx.Prepare(s.sql)  // Prepare on first use
    }
    return s.stmt.Query(args...)
}
```

**Why This Helps**:
- Commands that don't search (e.g., `zk new`) don't pay for search setup
- Startup time reduced by ~10-20ms (noticeable in CLI)

**Implication**:
- Pure-Go search libraries (Bleve) may not need this pattern
- But concept applies: lazy-load indexes, lazy-open databases

**Recommendation**:
- ✅ Lazy-load search index (don't open on `zk --help`)
- ✅ Profile startup time (sub-100ms is great for CLI UX)

---

## Architectural Patterns Worth Adopting

### 1. Port-Adapter (Hexagonal) Architecture

zk cleanly separates:
- **Ports** (`internal/core/*.go`) - Interfaces, domain logic
- **Adapters** (`internal/adapter/**/*.go`) - Implementations

**Benefit**:
- Can swap SQLite adapter for Bleve adapter without changing core
- Tests can use in-memory implementations
- Clear dependency direction (core → adapter, never reverse)

**Recommendation**: ✅ **Adopt this pattern in OpenNotes**

---

### 2. Functional Options for Queries

zk uses immutable structs with methods that return modified copies:

```go
func (o NoteFindOpts) ExcludingIDs(ids []NoteID) NoteFindOpts {
    o.ExcludeIDs = append(o.ExcludeIDs, ids...)
    return o  // Returns copy
}
```

**Benefit**:
- Thread-safe (no mutation)
- Chainable: `opts.ExcludingIDs(ids).IncludingTags(tags)`
- Easy to test (no side effects)

**Recommendation**: ✅ **Use functional options pattern for query building**

---

### 3. Transaction-Based Batch Operations

zk provides `Commit()` for atomic operations:

```go
Commit(transaction func(idx NoteIndex) error) error
```

**Use Case**:
- Reindexing multiple notes atomically
- Updating links and notes together
- Rollback on partial failure

**Recommendation**: ✅ **Essential for data integrity**

---

## Anti-Patterns to Avoid

### 1. Direct OS Filesystem Access in Utilities

**Problem**: `paths.Walk()` uses `filepath.Walk` directly, not `FileStorage` interface.

**Why It's Bad**:
- Can't test with in-memory filesystem
- Couples utilities to OS filesystem
- Prevents use in WASM context

**Fix**: ✅ Always route through `FileStorage` interface

---

### 2. No Rollback in Migrations

**Problem**: Migrations only go forward, no down migrations.

**Why It's Risky**:
- Failed migration leaves DB in inconsistent state
- Can't downgrade gracefully

**Fix**: ✅ Implement up/down migrations (like Rails, Alembic)

---

### 3. Hardcoded SQL Strings

**Problem**: SQL scattered across `note_dao.go` (936 lines), hard to maintain.

**Better Approach**:
- Use query builder (squirrel, goqu)
- Or centralize SQL in constants
- Or use code generation (sqlc)

**Recommendation**: ⚠️ Consider using Bleve/tantivy-go to avoid SQL entirely

---

## Surprising Findings

### 1. FTS5 Tokenizer Configuration

```sql
tokenize = "porter unicode61 remove_diacritics 1 tokenchars '''&/'"
```

**Insight**:
- Porter stemmer: "running" matches "run"
- Unicode61: Handles international characters
- `remove_diacritics`: "café" matches "cafe"
- `tokenchars '''&/'"`: Keeps apostrophes and slashes in tokens

**Why This Matters**:
- Affects search recall (how many results returned)
- Porter stemming is aggressive (can cause false matches)
- Some users may want exact matching only

**Recommendation**:
- ✅ Make tokenization configurable
- ✅ Offer "simple" and "advanced" modes

---

### 2. Link Snippet Storage

**Finding**: zk stores the surrounding text of each link in the database.

**Schema**:
```sql
snippet TEXT NOT NULL,
snippet_start INTEGER NOT NULL,
snippet_end INTEGER NOT NULL,
```

**Why**:
- Enables "show me where X is mentioned" queries
- Provides context without re-parsing notes
- Increases DB size by ~20% (estimated)

**Trade-off**:
- Storage cost vs query speed
- Could regenerate on demand (slower but smaller DB)

---

### 3. Sortable Path Trick

```go
sortablePath := strings.ReplaceAll(note.Path, "/", "\x01")
```

**Why**: SQLite's text sorting is lexicographic, not filesystem-order.

**Example Problem**:
- `a/z.md` sorts after `a.md` (wrong!)
- With `\x01`: `a\x01z.md` sorts before `a.md` (correct!)

**Insight**: Small hacks like this optimize common operations (listing notes in order)

---

## Emerging Consensus vs Outliers

### Consensus Findings
- ✅ FTS is essential for note search (confirmed across all implementations)
- ✅ BM25 is standard for relevance ranking (industry best practice)
- ✅ Graph-based link queries are valuable (unique to Zettelkasten tools)
- ✅ Tag normalization improves UX (avoid duplicates)

### Outlier Observations
- ⚠️ zk uses SQLite triggers (unusual pattern, most apps manage indexes manually)
- ⚠️ Channel-based streaming (idiomatic Go, but not common in CLI tools)
- ⚠️ Lazy statement preparation (optimization often skipped)

---

## Implications for OpenNotes Decision-Making

### High-Priority Learnings (Must Adopt)

1. **Interface Design**
   - ✅ Use zk's `NoteIndex` interface as-is
   - ✅ Use zk's `FileStorage` interface (afero-compatible)
   - ✅ Keep query options structure (`NoteFindOpts`)

2. **Query Syntax**
   - ✅ Adopt Google-like query DSL
   - ✅ Support field-specific search (`title:foo`)
   - ✅ Implement BM25 ranking

3. **Link Analysis**
   - ✅ Support recursive link queries
   - ✅ Implement max distance limiting
   - ✅ Store bidirectional links

### Medium-Priority (Consider Adopting)

1. **Snippet Generation**
   - ⚠️ Context-aware snippets (nice-to-have)
   - ⚠️ Highlight matching terms (improves UX)

2. **Tag Management**
   - ⚠️ Normalized tag storage (prevents duplicates)
   - ⚠️ GLOB pattern matching (power-user feature)

3. **Streaming Results**
   - ⚠️ Channel-based iteration (prevents memory bloat)
   - ⚠️ Essential for large notebooks (>10k notes)

### Low-Priority (Nice to Have)

1. **Migration System**
   - 🔵 Can defer until schema changes are needed
   - 🔵 But plan for it from day one

2. **Lazy Loading**
   - 🔵 Optimize startup time later
   - 🔵 Profile first, optimize second

---

## Areas Needing Further Research

### Unanswered Questions

1. **Performance at Scale**
   - How does zk perform with 100k notes?
   - What's the index rebuild time?
   - Query latency percentiles (p50, p95, p99)?

2. **Pure-Go FTS Alternatives**
   - Bleve: Feature comparison with FTS5?
   - tantivy-go: Rust bindings, worth the FFI cost?
   - Custom implementation: Feasibility?

3. **Graph Database Integration**
   - Would Neo4j or dgraph improve link queries?
   - Trade-off: complexity vs query power
   - Can pure-Go graph libs (gonum) compete?

### Recommended Next Steps

1. **Benchmark Bleve vs FTS5**
   - Index 10k, 100k notes
   - Measure query latency
   - Compare index size

2. **Prototype Pure-Go Implementation**
   - Use zk's interfaces
   - Implement with Bleve backend
   - Measure performance delta

3. **User Research**
   - Do users actually use recursive link queries?
   - Is BM25 ranking noticeable vs simple scoring?
   - What's the most common query pattern?

---

## Conclusion

**Key Takeaway**: zk's **interface design is gold**, but the **SQLite implementation is a non-starter** for OpenNotes.

**Strategic Recommendation**:
1. ✅ Adopt zk's `NoteIndex` and `FileStorage` interfaces verbatim
2. ✅ Reuse query DSL syntax (Google-like)
3. ✅ Implement BM25 ranking and link analysis
4. ❌ Rewrite search backend with pure-Go FTS (Bleve/tantivy-go)
5. ⚠️ Benchmark before committing to specific implementation

**Risk Assessment**:
- **Low Risk**: Interface adoption (zero regrets likely)
- **Medium Risk**: Bleve performance (needs validation)
- **High Risk**: Custom FTS implementation (engineering months)

**Next Decision Point**: Bleve vs tantivy-go vs custom FTS (requires separate research subtopic)
