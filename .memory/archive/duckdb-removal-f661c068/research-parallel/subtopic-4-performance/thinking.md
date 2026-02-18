# DuckDB Performance Baseline & Benchmarking - Thinking Notes

## Skills Discovery Process

### Skills Found and Loaded

I discovered and loaded three highly relevant skills for this performance benchmarking research:

1. **golang-pro** (`/home/zenobius/.pi/agent/skills/experts/language-specialists/golang-pro/SKILL.md`)
   - **Why loaded**: Expert Go developer specializing in high-performance systems, concurrent programming
   - **Relevant capabilities**:
     - CPU and memory profiling with pprof
     - Benchmark-driven development
     - Zero-allocation techniques
     - Performance optimization patterns
     - Testing methodology with table-driven benchmarks
   - **Application to task**: Critical for understanding Go benchmarking best practices, profiling methodology, and optimization techniques

2. **performance-engineer** (`/home/zenobius/.pi/agent/skills/experts/quality-security/performance-engineer/SKILL.md`)
   - **Why loaded**: Expert in system optimization, bottleneck identification, scalability engineering
   - **Relevant capabilities**:
     - Performance testing design (load, stress, baseline establishment)
     - Bottleneck analysis (CPU profiling, memory analysis, I/O investigation)
     - Application profiling (code hotspots, method timing, memory allocation)
     - Performance monitoring and metrics collection
   - **Application to task**: Essential for systematic performance analysis methodology and identifying optimization opportunities

3. **database-optimizer** (`/home/zenobius/.pi/agent/skills/experts/data-ai/database-optimizer/SKILL.md`)
   - **Why loaded**: Expert in query optimization, performance tuning across multiple database systems
   - **Relevant capabilities**:
     - Query optimization and execution plan analysis
     - Performance analysis techniques
     - Memory and I/O optimization
     - Database-specific tuning strategies
   - **Application to task**: Provides expertise in database performance characteristics, understanding DuckDB's query execution patterns

### Skills Not Loaded (and Why)

- **deep-researcher**: Not needed - this is technical performance testing, not web research
- **codemapper**: Already familiar with codebase structure from AGENTS.md
- **writing-git-commits**: Not creating commits in this research phase
- **miniproject**: Will use but don't need skill guidance for basic artifact creation
- **playwright-skill**: Not relevant for CLI tool benchmarking
- **brave-search**: Not needed - focusing on existing codebase measurement

## Initial Observations

### Existing Benchmark Infrastructure

OpenNotes already has comprehensive benchmarking infrastructure in place:
- `internal/services/search_bench_test.go` (400+ lines)
- Benchmarks for fuzzy search, text search, boolean queries, link queries
- Multiple dataset sizes: 100, 1k, 10k notes
- Memory allocation tracking with `b.ReportAllocs()`
- Performance targets documented in code comments

### Current Performance Baseline (AMD Ryzen 7 5800X3D, Go 1.24.7)

From initial benchmark run:

**Fuzzy Search (Primary User Interaction)**:
- 100 notes: ~298µs/op, 33KB alloc, 829 allocs
- 1k notes: ~3.1ms/op, 299KB alloc, 8,029 allocs
- 10k notes: ~29.9ms/op, 2.96MB alloc, 80,029 allocs

**Text Search (Fast Path)**:
- 100 notes: ~35µs/op, 32KB alloc, 108 allocs
- 1k notes: ~318µs/op, 288KB alloc, 1,011 allocs
- 10k notes: ~3.2ms/op, 4.31MB alloc, 10,019 allocs

**Boolean Query Building (Query Construction)**:
- Simple: ~128ns/op, 96B alloc, 4 allocs
- Complex (6 conditions): ~1.5µs/op, 1.8KB alloc, 33 allocs
- Many conditions (20): ~4.1µs/op, 6.4KB alloc, 90 allocs

**Observations**:
- Fuzzy search ~10x slower than text search (expected - ranking algorithm)
- Linear scaling with note count (good algorithmic complexity)
- Allocations scale linearly with note count
- Query building is extremely fast (microsecond range)

### DuckDB Infrastructure Costs

**Binary Size Impact**:
- Current binary size: **64MB** (substantial)
- DuckDB dependencies: 186 transitive dependencies in go.mod graph
- Multiple platform-specific bindings required (darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64)

**Build Complexity**:
- Requires platform-specific CGO bindings
- Cross-compilation complexity for different architectures
- Dependency on external C++ DuckDB library

**Runtime Initialization Costs** (from CPU profile):
- DuckDB database opening: ~0.4s (9.5% of benchmark time)
- Markdown extension installation: Additional overhead
- Lazy initialization with sync.Once (75ms overhead in profiles)

### Current Bottlenecks (from pprof analysis)

**CPU Profile Top Contributors**:
1. Fuzzy matching algorithm (42% - sahilm/fuzzy.FindFromNoSort)
2. CGO calls to DuckDB (19.5% - runtime.cgocall)
3. DuckDB initialization (9.5% - database opening)
4. DuckDB query execution (8.5% - ExecutePending)

**Key Insight**: Current search is **NOT using DuckDB** for fuzzy/text search!
- FuzzySearch/TextSearch operate on in-memory Note slices
- DuckDB only used for SQL queries and boolean conditions
- This explains why DuckDB overhead is initialization-only

### Performance Targets vs Reality

**Documented Targets** (from code comments):
- Fuzzy search 10k notes: < 50ms (Actual: ~30ms ✓ EXCEEDS)
- Text search 10k notes: < 10ms (Actual: ~3.2ms ✓ EXCEEDS)
- Simple boolean query: < 20ms (Actual: ~0.13ms ✓ EXCEEDS)
- Complex boolean query: < 100ms (Actual: ~1.5ms ✓ EXCEEDS)

**Assessment**: Current performance already excellent for in-memory operations

### DuckDB Limitations Discovered

**Known Issues**:
1. **No afero filesystem support**: DuckDB reads from real filesystem only
   - Blocks testing with virtual filesystems
   - Makes integration testing harder
   - Cannot use memory-backed filesystems for tests

2. **Large binary size**: 64MB for a CLI tool is substantial
   - Impacts distribution size
   - Slower download/install times
   - Memory footprint even when not using DuckDB features

3. **Build complexity**: CGO + platform-specific bindings
   - Complicates cross-compilation
   - Requires C++ toolchain
   - Slower build times

4. **Initialization overhead**: ~0.4-0.5s for DB + extension
   - Impacts CLI startup time for SQL queries
   - Not amortized for short-lived commands

### Memory Usage Characteristics

**Memory Allocation Patterns** (from benchmem):
- Text search 10k: 4.31MB allocations (10,019 allocs)
- Fuzzy search 10k: 2.96MB allocations (80,029 allocs)
- Fuzzy search has 8x more allocations but less total memory
  - Suggests many small allocations in ranking algorithm
  - Opportunity for pooling or pre-allocation

**Scaling Behavior**:
- Both search methods scale linearly with note count
- ~430 bytes per note for text search
- ~296 bytes per note for fuzzy search
- Acceptable memory footprint for 10k notes (~3-4MB)

### Architecture Questions for Next Implementation

1. **Is DuckDB actually needed?**
   - Current search doesn't use it (in-memory only)
   - Only used for SQL queries and boolean conditions
   - Could simpler query language replace it?

2. **What features depend on DuckDB?**
   - `read_markdown()` function for loading notes
   - SQL query interface for advanced users
   - Boolean condition filtering
   - Link analysis queries (links-to, linked-by)

3. **Performance comparison targets**:
   - Need to benchmark DuckDB-based search vs current in-memory
   - Is there value in SQL-powered full-text search?
   - Trade-offs: flexibility vs performance vs binary size

### Research Directions

Based on golang-pro skill guidance:
- [ ] Create CPU and memory profiles for different query types
- [ ] Analyze escape analysis for allocation reduction
- [ ] Benchmark with different note sizes (small vs large content)
- [ ] Profile DuckDB query execution for SQL-based search
- [ ] Compare in-memory vs DuckDB-backed search performance

Based on performance-engineer skill guidance:
- [ ] Establish performance baselines for ALL search methods
- [ ] Create load test scenarios (concurrent queries)
- [ ] Analyze resource usage under sustained load
- [ ] Identify scalability limits (max note count)
- [ ] Document performance regression test suite

Based on database-optimizer skill guidance:
- [ ] Analyze DuckDB query execution plans
- [ ] Profile DuckDB markdown extension overhead
- [ ] Compare DuckDB read_markdown vs direct file I/O
- [ ] Evaluate DuckDB's full-text search capabilities
- [ ] Measure index build time if using DuckDB FTS

## Next Steps

1. **Comprehensive Benchmarking**:
   - Run full benchmark suite with memory profiling
   - Generate CPU and memory flame graphs
   - Benchmark DuckDB-based search implementation
   - Compare in-memory vs DuckDB performance

2. **Bottleneck Analysis**:
   - Profile fuzzy search algorithm (42% of time)
   - Analyze allocation patterns in ranking
   - Measure DuckDB overhead for different query types

3. **Performance Target Setting**:
   - Define targets for replacement implementation
   - Set memory usage limits
   - Establish binary size budget
   - Define startup time requirements

4. **Documentation**:
   - Create benchmark comparison tables
   - Document profiling methodology
   - Provide reproduction instructions
   - Define performance regression tests
