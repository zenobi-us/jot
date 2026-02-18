# ZK Search Architecture Analysis - Research Thinking

## Research Session Metadata
- Start Time: 2026-02-01 15:28 GMT+10:30
- Researcher: Claude (pi coding agent)
- Parent Task: Evaluate search implementation strategies for OpenNotes to replace DuckDB
- Storage Path: `/mnt/Store/Projects/Mine/Github/opennotes/.memory/research-parallel/subtopic-1-zk-search`

## Skill Discovery & Selection

### Skills Identified
1. **codemapper** (`/home/zenobius/.pi/agent/skills/devtools/codemapper/SKILL.md`)
   - **Why loaded**: Provides AST-based code analysis using tree-sitter for Go codebases
   - **Relevance**: Essential for mapping zk's search architecture, tracing call paths, and understanding component relationships
   - **Capabilities**: Stats overview, symbol queries, call graph tracing, file inspection with AI-optimized output

2. **golang-pro** (`/home/zenobius/.pi/agent/skills/experts/language-specialists/golang-pro/SKILL.md`)
   - **Why loaded**: Expert Go patterns, idioms, and architecture understanding
   - **Relevance**: Helps interpret Go-specific patterns in zk codebase (interfaces, channels, error handling)
   - **Capabilities**: Idiomatic Go analysis, concurrency patterns, performance optimization insights

3. **architect-reviewer** (`/home/zenobius/.pi/agent/skills/experts/quality-security/architect-reviewer/SKILL.md`)
   - **Why loaded**: System design evaluation and architectural pattern recognition
   - **Relevance**: Assesses zk's search architecture for scalability, maintainability, and integration opportunities
   - **Capabilities**: Pattern identification, scalability assessment, technology evaluation

4. **deep-research** (`/home/zenobius/.pi/agent/skills/research/deep-research/SKILL.md`)
   - **Why loaded**: Structured research methodology with verification standards
   - **Relevance**: Ensures systematic investigation with source tracking and confidence levels
   - **Capabilities**: Multi-source verification, evidence mapping, confidence-level documentation

### Research Methodology

Following deep-research skill's 5-phase approach:

**Phase 1: Topic Scoping**
- Primary questions:
  1. How does zk parse and execute search queries?
  2. What indexing strategy does zk use?
  3. How does zk access the filesystem (afero integration points)?
  4. What query DSL syntax/operators are supported?
  5. What are the key Go packages/types/interfaces?

**Phase 2: Source Collection**
- Primary sources: zk source code, documentation
- Secondary sources: GitHub issues, discussions, commit history
- Tool usage: CodeMapper for code analysis, git clone for source access

**Phase 3: Information Collation**
- Organize by: architecture components, data flow, API boundaries
- Create: State machine diagrams, component diagrams, interface maps

**Phase 4: Verification**
- Cross-reference: Code behavior vs documentation
- Validate: Claims traceable to actual source code locations
- Confidence levels: High (code + tests + docs), Medium (code + docs), Low (code only)

**Phase 5: Output Generation**
- 5 required files: thinking.md, research.md, verification.md, insights.md, summary.md
- Additional artifacts: ASCII diagrams, interface maps, code path traces

## Research Execution Plan

### Step 1: Clone Repository
```bash
git clone https://github.com/zk-org/zk /tmp/zk-analysis
cd /tmp/zk-analysis
```

### Step 2: CodeMapper Overview
```bash
cm stats /tmp/zk-analysis --format ai
cm map /tmp/zk-analysis --level 2 --format ai
```

### Step 3: Identify Search Components
```bash
# Find search-related files
find /tmp/zk-analysis -type f -name "*.go" | grep -i search

# Use CodeMapper to find search-related symbols
cm query "search" /tmp/zk-analysis --format ai
cm query "index" /tmp/zk-analysis --format ai
cm query "query" /tmp/zk-analysis --format ai
```

### Step 4: Trace Key Code Paths
- Query parsing flow
- Index building flow
- Search execution flow
- Filesystem access patterns

### Step 5: Extract Interface Definitions
```bash
# Find interface definitions
grep -r "type.*interface" /tmp/zk-analysis --include="*.go"

# Inspect key interfaces with CodeMapper
cm inspect <search-related-files> --format ai
```

### Step 6: Analyze Query DSL
- Locate query parser
- Extract supported operators
- Document syntax examples
- Map operator implementations

### Step 7: Filesystem Abstraction Analysis
- Identify filesystem access points
- Check for existing abstractions
- Map afero integration opportunities
- Document coupling points

### Step 8: Create Diagrams
- Component architecture diagram (ASCII)
- Query parsing state machine (ASCII)
- Index building state machine (ASCII)
- Search execution state machine (ASCII)

## Constraints & Avoidance Criteria

From parent task requirements:
- ❌ Blog posts older than 2 years
- ❌ Marketing materials
- ❌ C/C++ dependent solutions
- ❌ Solutions incompatible with filesystem abstraction

Additional constraints:
- ✅ Focus on Go code analysis (zk is Go-based)
- ✅ Prioritize primary sources (code, official docs)
- ✅ Verify all claims against actual code
- ✅ Document performance characteristics with evidence

## Progress Tracking

- [x] Repository cloned
- [x] CodeMapper overview completed
- [x] Search components identified
- [x] Query parsing traced
- [x] Indexing strategy documented
- [x] Search execution mapped
- [x] Filesystem access analyzed
- [x] Query DSL documented
- [x] State machines created (3 ASCII diagrams in research.md)
- [x] Interface map completed (component diagram in research.md)
- [x] Performance characteristics gathered (section 7 in research.md)
- [x] Verification completed (verification.md)
- [x] All 5 output files written

## Research Completion Summary

**Status**: ✅ COMPLETE

**Output Files**:
1. ✅ `thinking.md` - Research methodology and progress (this file)
2. ✅ `research.md` - Comprehensive technical analysis (30KB)
3. ✅ `verification.md` - Source traceability and confidence levels (13KB)
4. ✅ `insights.md` - Strategic implications and recommendations (15KB)
5. ✅ `summary.md` - Executive summary with actionable recommendations (12KB)

**Total Research Output**: ~70KB of documentation across 5 files

**Key Deliverables Completed**:
- ✅ Search architecture overview with component diagram
- ✅ Query DSL specification with examples
- ✅ Code path maps (3 state machines: parse, index, execute)
- ✅ Afero integration opportunities assessment
- ✅ Performance characteristics documentation

**Critical Finding**: zk's SQLite dependency prevents direct adoption, but interface design and query DSL are highly reusable.

**Recommendation**: Adopt zk's architecture patterns, reimplement with pure-Go search engine (Bleve).

## Key Findings During Analysis

### Repository Stats (via CodeMapper)
- Languages: Go (122 files), Markdown (159), Python (1)
- Total symbols: 1,427 (functions: 417, classes: 186, methods: 307)
- Codebase size: 641KB across 282 files

### Core Components Identified
1. **NoteFindOpts** (`internal/core/note_find.go`) - Filter/query options structure
2. **NoteIndex** (`internal/core/note_index.go`) - Index interface definition
3. **NoteDAO** (`internal/adapter/sqlite/note_dao.go`) - SQLite implementation (936 lines)
4. **FileStorage** (`internal/core/fs.go`, `internal/adapter/fs/fs.go`) - Filesystem abstraction
5. **FTS5 Query Converter** (`internal/util/fts5/fts5.go`) - Google-like → FTS5 query transformer

### Search Architecture Discovery
- **Database**: SQLite with FTS5 full-text search extension
- **Index Type**: Inverted index via FTS5 virtual table
- **Query DSL**: Google-like syntax (converted to FTS5 queries)
- **Filesystem Abstraction**: Already exists via `core.FileStorage` interface (afero-compatible!)

### Critical Source Files Analyzed
- `/tmp/zk-analysis/internal/core/note_find.go` (193 lines) - Query options
- `/tmp/zk-analysis/internal/core/note_index.go` (218 lines) - Index interface
- `/tmp/zk-analysis/internal/adapter/sqlite/note_dao.go` (936 lines) - Search implementation
- `/tmp/zk-analysis/internal/adapter/sqlite/db.go` (300+ lines) - Database schema
- `/tmp/zk-analysis/internal/util/fts5/fts5.go` (117 lines) - Query transformer
- `/tmp/zk-analysis/internal/core/fs.go` (37 lines) - FS interface
- `/tmp/zk-analysis/internal/adapter/fs/fs.go` (135 lines) - FS implementation

## Next Actions

1. ~~Clone zk repository to /tmp/zk-analysis~~ ✓
2. ~~Run initial CodeMapper analysis~~ ✓
3. ~~Begin systematic component identification~~ ✓
4. Create ASCII state machine diagrams
5. Document performance characteristics
6. Write all 5 output files
