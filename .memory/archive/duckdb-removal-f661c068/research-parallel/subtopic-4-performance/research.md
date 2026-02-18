# DuckDB Performance Baseline & Benchmarking Research

## Executive Summary

This research establishes comprehensive performance baselines for OpenNotes' current DuckDB-based search implementation. Key findings:

- **Current Performance**: Excellent for in-memory operations (10k notes in 3-30ms)
- **Binary Size Impact**: DuckDB adds **64MB** to binary size
- **Architecture Discovery**: Search operations are **in-memory**, DuckDB only used for note loading and SQL queries
- **Bottleneck**: Fuzzy search algorithm (42% CPU time), not DuckDB
- **Opportunity**: DuckDB overhead (19.5% CGO calls, 9.5% initialization) could be eliminated

## Test Environment

- **Hardware**: AMD Ryzen 7 5800X3D 8-Core Processor (16 threads)
- **OS**: Linux (amd64)
- **Go Version**: 1.24.7
- **DuckDB Version**: v2.5.4 (via duckdb-go/v2)
- **Benchmark Tool**: Go's built-in testing.B framework
- **Profiling Tools**: pprof (CPU and memory)

## Performance Baseline Results

### 1. Search Performance (In-Memory Operations)

#### Fuzzy Search Benchmarks

| Dataset Size | Time/Op | Memory/Op | Allocs/Op | Notes |
|--------------|---------|-----------|-----------|-------|
| 100 notes    | 298.7µs | 33.3 KB   | 829       | Baseline |
| 1,000 notes  | 3.15ms  | 299.7 KB  | 8,029     | 10x scale |
| 10,000 notes | 29.9ms  | 2.96 MB   | 80,029    | 100x scale |

**Performance Characteristics**:
- Linear scaling: O(n) with note count
- ~3µs per note for fuzzy matching
- ~300 bytes memory per note
- ~8 allocations per note (high - optimization opportunity)

**Target Comparison**:
- Target: < 50ms for 10k notes
- Actual: ~30ms ✓ **EXCEEDS TARGET by 40%**

#### Text Search Benchmarks

| Dataset Size | Time/Op | Memory/Op | Allocs/Op | Notes |
|--------------|---------|-----------|-----------|-------|
| 100 notes    | 35.4µs  | 32.8 KB   | 108       | Baseline |
| 1,000 notes  | 318.6µs | 288.5 KB  | 1,011     | 10x scale |
| 10,000 notes | 3.24ms  | 4.31 MB   | 10,019    | 100x scale |

**Performance Characteristics**:
- Linear scaling: O(n) with note count
- ~324ns per note for text matching
- ~430 bytes memory per note
- ~1 allocation per note (good)

**Target Comparison**:
- Target: < 10ms for 10k notes
- Actual: ~3.2ms ✓ **EXCEEDS TARGET by 68%**

**Speed Comparison**:
- Text search is **9.2x faster** than fuzzy search
- Fuzzy search has **8x more allocations**
- Trade-off: Speed vs ranking quality

### 2. Query Building Performance

#### Boolean Query Construction

| Query Complexity | Time/Op | Memory/Op | Allocs/Op | Conditions |
|------------------|---------|-----------|-----------|------------|
| Simple (1 cond)  | 128ns   | 96 B      | 4         | AND tag=workflow |
| Complex (6 cond) | 1.55µs  | 1.84 KB   | 33        | AND/OR/NOT mix |
| Many (20 cond)   | 4.13µs  | 6.46 KB   | 90        | Complex query |

**Performance Characteristics**:
- Sub-microsecond for simple queries
- Scales linearly with condition count (~200ns per condition)
- Negligible overhead compared to search execution

**Target Comparison**:
- Simple target: < 20ms, Actual: 0.128ms ✓ **156x faster**
- Complex target: < 100ms, Actual: 1.55µs ✓ **64,516x faster**

#### Link Query Construction

| Query Type | Time/Op | Memory/Op | Allocs/Op |
|------------|---------|-----------|-----------|
| links-to   | 291ns   | 128 B     | 6         |
| linked-by  | 216ns   | 112 B     | 5         |
| Combined   | 1.39µs  | 1.88 KB   | 21        |

**Observations**:
- Link queries are extremely fast (< 300ns)
- Combined queries still sub-microsecond
- Query building is NOT a bottleneck

### 3. Glob Pattern Conversion

| Pattern Type | Time/Op | Memory/Op | Allocs/Op |
|--------------|---------|-----------|-----------|
| Simple (*.md, dir/*) | 206ns | 16 B | 2 |
| Complex (**/*.md, dir/**/sub/*.md) | 811ns | 280 B | 13 |

**Observations**:
- Pattern conversion is negligible overhead
- Even complex patterns < 1µs
- Not a performance concern

### 4. Condition Parsing Performance

| Parse Complexity | Time/Op | Memory/Op | Allocs/Op |
|------------------|---------|-----------|-----------|
| Simple (1 flag)  | 98.5ns  | 96 B      | 2         |
| Complex (3+3+1)  | 658ns   | 1.15 KB   | 10        |

**Observations**:
- CLI flag parsing is extremely fast
- Even complex conditions parse in < 1µs
- CLI overhead is negligible

### 5. DuckDB-Specific Performance

#### Database Operations

| Operation | Time/Op | Overhead % | Notes |
|-----------|---------|------------|-------|
| DB Initialization | ~400ms | 9.5% | One-time cost |
| Extension Loading | ~100ms | - | Markdown extension |
| CGO Calls | - | 19.5% | Runtime overhead |
| Query Execution | - | 8.5% | For SQL queries |

**Key Findings**:
- **Total DuckDB overhead: ~29% of execution time**
- Initialization is one-time but significant for CLI
- CGO boundary crossings add 19.5% overhead
- Current search does NOT use DuckDB (in-memory only)

#### SQL Preprocessing

| Query Type | Time/Op | Memory/Op | Allocs/Op |
|------------|---------|-----------|-----------|
| No patterns | 1.07µs | 226 B | 6 |
| Single pattern | 1.41µs | 381 B | 10 |
| Multiple patterns | 1.56µs | 472 B | 9 |
| Complex query | 3.71µs | 787 B | 10 |
| Very large query | 22.4µs | 8.85 KB | 9 |

**Observations**:
- SQL preprocessing adds < 25µs for most queries
- Scales with query complexity
- Security validation (path traversal) is negligible

### 6. Display/Rendering Performance

| Dataset Size | JSON Rendering | Memory | Notes |
|--------------|----------------|--------|-------|
| Small (10 rows) | 13.2µs | 8.3 KB | Fast |
| Medium (100 rows) | 742µs | 435 KB | Acceptable |
| Large (1000 rows) | 11.4ms | 6.17 MB | Noticeable |

**Observations**:
- JSON rendering scales super-linearly
- Large result sets (>100 rows) show rendering overhead
- Not a primary concern for typical use cases

## CPU Profiling Analysis

### Top CPU Consumers (from pprof -top -cum)

```
flat   flat%   cum   cum%    Function
1.77s  42.0%   1.77s 42.0%   fuzzy.FindFromNoSort (fuzzy matching)
0.82s  19.5%   0.82s 19.5%   runtime.cgocall (DuckDB calls)
0.40s   9.5%   0.40s  9.5%   duckdb.OpenExt (DB initialization)
0.37s   8.5%   0.37s  8.5%   duckdb.ExecutePending (query exec)
```

### Performance Breakdown

| Component | CPU Time | % of Total | Optimization Potential |
|-----------|----------|------------|------------------------|
| Fuzzy matching algorithm | 1.77s | 42.0% | HIGH - algorithmic |
| CGO calls (DuckDB) | 0.82s | 19.5% | HIGH - remove DuckDB |
| DB initialization | 0.40s | 9.5% | HIGH - remove DuckDB |
| Query execution | 0.37s | 8.5% | MEDIUM - needed for SQL |
| Other (stdlib, etc.) | 0.85s | 20.5% | LOW |

**Key Insight**: DuckDB overhead (29% total) could be eliminated for search operations, as search is already in-memory.

## Memory Profiling Analysis

### Memory Allocation by Component (from pprof memory profile)

| Component | Allocation | % of Total | Notes |
|-----------|------------|------------|-------|
| Fuzzy matching | 67.5 MB | 43.2% | Largest allocator |
| Fuzzy ranking setup | 18 MB | 11.5% | FindFrom |
| Benchmark setup | 13.3 MB | 8.5% | Test data |
| Search service | 12.5 MB | 8.0% | FuzzySearch |
| JSON marshaling | 8.3 MB | 5.3% | Display |
| Markdown rendering | 5.0 MB | 3.2% | Glamour |

**Memory Hotspots**:
1. **fuzzy.FindFromNoSort**: 67.5MB (43.2%) - largest single allocator
   - Opportunities: Object pooling, pre-allocation
2. **fuzzy.Find/FindFrom**: 18MB (11.5%) - ranking setup
   - Many small allocations (80k for 10k notes)
3. **Search wrapper**: 12.5MB (8%) - result collection

### Allocation Patterns

**Fuzzy Search (10k notes)**:
- Total allocations: 2.96 MB
- Allocation count: 80,029 (8 per note)
- Pattern: Many small allocations
- Opportunity: Pre-allocate result slices, use sync.Pool

**Text Search (10k notes)**:
- Total allocations: 4.31 MB (larger despite faster)
- Allocation count: 10,019 (1 per note)
- Pattern: Fewer, larger allocations
- More memory-efficient allocation strategy

## Binary Size Analysis

### Current Binary Characteristics

| Metric | Value | Impact |
|--------|-------|--------|
| Total binary size | **64 MB** | Large for CLI tool |
| DuckDB contribution | ~50-55 MB | 78-86% of total |
| Go runtime + stdlib | ~5-7 MB | Normal |
| Application code | ~2-3 MB | Minimal |

### Dependency Analysis

**DuckDB Dependencies**:
- Direct: github.com/duckdb/duckdb-go/v2 v2.5.4
- Platform bindings (6 platforms):
  - linux-amd64, linux-arm64
  - darwin-amd64, darwin-arm64
  - windows-amd64
  - Bindings version: v0.1.24
- Transitive dependencies: **186** (from go.mod graph)

**Binary Size Breakdown** (estimated):
```
DuckDB shared library:     ~45 MB (CGO C++ library)
Platform-specific bindings: ~5 MB
DuckDB Go wrapper:         ~2 MB
Other dependencies:        ~10 MB
Application code:          ~2 MB
--------------------------------
Total:                     ~64 MB
```

### Build Complexity Impact

**Cross-Compilation Challenges**:
- Requires platform-specific CGO bindings
- Cannot trivially cross-compile (C++ dependency)
- Build matrix explosion: 3 OS × 2 arch = 6 binaries
- CI/CD complexity: Need platform-specific builders

**Development Overhead**:
- Slower builds (CGO + large dependency tree)
- Requires C++ toolchain for development
- Cannot use pure Go toolchain

## Build Performance

### Build Time Analysis

| Build Type | Time | Notes |
|------------|------|-------|
| Clean build | ~15-20s | With CGO compilation |
| Incremental | ~2-3s | Cached CGO objects |
| Test suite | ~4-5s | 161+ tests |

**Observations**:
- CGO adds 5-10s to clean builds
- Majority of time is DuckDB C++ compilation
- Incremental builds fast (good caching)

### Test Performance

**Full Test Suite**:
- Total tests: 161+
- Total time: ~4 seconds
- Benchmarks: ~48 seconds (includes profiling overhead)

**Test Breakdown**:
- Unit tests: ~2s (fast)
- Integration tests: ~1s (DuckDB initialization)
- Benchmarks: ~45s (includes warm-up, profiling)

## Scalability Analysis

### Note Count Scaling

| Metric | 100 notes | 1k notes | 10k notes | Scaling |
|--------|-----------|----------|-----------|---------|
| Fuzzy search | 298µs | 3.15ms | 29.9ms | Linear O(n) |
| Text search | 35µs | 318µs | 3.24ms | Linear O(n) |
| Memory (fuzzy) | 33 KB | 300 KB | 2.96 MB | Linear |
| Memory (text) | 33 KB | 289 KB | 4.31 MB | Linear |

**Extrapolated Performance (100k notes)**:
- Fuzzy search: ~300ms (acceptable)
- Text search: ~32ms (excellent)
- Memory usage: ~30-43 MB (acceptable)

**Conclusion**: Current implementation scales well to 100k+ notes

### Concurrent Query Performance

**Benchmark: Concurrent SQL Preprocessing**:
- Time per operation: 337ns
- Memory per operation: 439 B
- 10 allocations per operation

**Observations**:
- Concurrent queries show minimal contention
- Mutex overhead is negligible
- Good multi-threaded scalability

## Performance Targets for Replacement Implementation

Based on current baseline performance:

### Must-Meet Targets (Critical)

| Metric | Target | Current | Rationale |
|--------|--------|---------|-----------|
| Fuzzy search 10k | < 30ms | 29.9ms | Match current |
| Text search 10k | < 5ms | 3.24ms | Match current |
| Memory 10k notes | < 5 MB | 2.96 MB | Match current |
| Binary size | < 15 MB | 64 MB | **78% reduction** |
| Startup time | < 50ms | ~500ms | **10x faster** |

### Should-Meet Targets (Important)

| Metric | Target | Current | Rationale |
|--------|--------|---------|-----------|
| Fuzzy search 100k | < 500ms | ~300ms (est) | Maintain scaling |
| Allocations 10k | < 40k | 80k | **50% reduction** |
| CGO overhead | 0% | 19.5% | Eliminate CGO |
| Build time | < 5s | ~15-20s | **3-4x faster** |

### Could-Meet Targets (Nice-to-Have)

| Metric | Target | Current | Rationale |
|--------|--------|---------|-----------|
| Fuzzy search 10k | < 15ms | 29.9ms | 2x faster |
| Text search 10k | < 2ms | 3.24ms | 1.5x faster |
| Memory efficiency | +20% | - | Better pooling |

## Benchmark Comparison Framework

### Test Suite Structure

For comparing implementations, use this benchmark structure:

```go
// Core search benchmarks (must have)
BenchmarkFuzzySearch_100Notes
BenchmarkFuzzySearch_1kNotes
BenchmarkFuzzySearch_10kNotes
BenchmarkTextSearch_100Notes
BenchmarkTextSearch_1kNotes
BenchmarkTextSearch_10kNotes

// Memory profiling (must have)
BenchmarkFuzzySearch_Memory
BenchmarkTextSearch_Memory

// Query building (nice to have)
BenchmarkBooleanQuery_Simple
BenchmarkBooleanQuery_Complex

// Concurrent performance (nice to have)
BenchmarkConcurrentSearch
```

### Comparison Metrics

**Primary Metrics** (must compare):
1. Time/operation (ns/op or µs/op)
2. Memory allocations (bytes/op)
3. Allocation count (allocs/op)
4. Throughput (queries/second)

**Secondary Metrics** (should compare):
1. P50, P95, P99 latencies
2. Memory peak usage
3. CPU utilization
4. Concurrent scalability

**Tertiary Metrics** (nice to compare):
1. Binary size
2. Build time
3. Startup time
4. Cross-compilation support

### Regression Prevention

**Automated Performance Tests**:
```bash
# Run benchmarks with baseline comparison
go test -bench=. -benchmem -count=5 ./internal/services/... | tee current.txt
benchstat baseline.txt current.txt
```

**CI/CD Integration**:
- Run benchmarks on every PR
- Compare against main branch baseline
- Fail if >10% regression in critical metrics
- Report performance improvements

## DuckDB Limitations Documented

### 1. No Afero Virtual Filesystem Support

**Impact**:
- Cannot use memory-backed filesystems in tests
- Integration tests require real filesystem
- Harder to test edge cases (permissions, missing files)
- Slower test execution (disk I/O)

**Workaround**:
- Use temporary directories for tests
- Clean up after each test
- Accept slower test execution

**Replacement Opportunity**:
- Pure Go implementation can use afero
- Memory-backed testing
- Faster, more isolated tests

### 2. Large Binary Size (64 MB)

**Impact**:
- Distribution overhead (download time)
- Storage footprint
- Memory overhead (even when idle)
- Deployment complexity

**Cause**:
- C++ DuckDB library (~45 MB)
- Platform-specific bindings (~5 MB)
- Multiple architecture support

**Replacement Opportunity**:
- Pure Go: 10-15 MB binary
- **78-85% size reduction**

### 3. CGO Build Complexity

**Impact**:
- Requires C++ toolchain
- Cross-compilation challenges
- Slower build times
- Platform-specific issues

**Cause**:
- DuckDB is C++ library
- CGO bridge overhead
- Platform-specific bindings

**Replacement Opportunity**:
- Pure Go: trivial cross-compilation
- Fast builds (no CGO)
- Single toolchain

### 4. Initialization Overhead (~500ms)

**Impact**:
- Slow CLI startup for quick queries
- User experience degradation
- Not amortized for short commands

**Cause**:
- Database initialization (~400ms)
- Extension loading (~100ms)
- CGO setup overhead

**Replacement Opportunity**:
- Pure Go: < 50ms startup
- **10x faster** initialization

### 5. CGO Runtime Overhead (19.5%)

**Impact**:
- Slower query execution
- Memory allocation overhead
- GC pressure from CGO

**Cause**:
- Go ↔ C boundary crossings
- Type conversions
- Memory management

**Replacement Opportunity**:
- Pure Go: zero CGO overhead
- **19.5% performance improvement** potential

## Profiling Methodology

### CPU Profiling

**Command**:
```bash
go test -bench=BenchmarkFuzzySearch_10kNotes \
  -benchmem \
  -cpuprofile=/tmp/cpu.prof \
  -memprofile=/tmp/mem.prof \
  ./internal/services/...
```

**Analysis**:
```bash
# Top functions by cumulative time
go tool pprof -top -cum /tmp/cpu.prof

# Interactive exploration
go tool pprof /tmp/cpu.prof
# (pprof) top20
# (pprof) list FuzzySearch
# (pprof) web

# Flame graph
go tool pprof -http=:8080 /tmp/cpu.prof
```

### Memory Profiling

**Command**: Same as CPU profiling (generates both)

**Analysis**:
```bash
# Top allocators
go tool pprof -top -alloc_space /tmp/mem.prof

# Allocation sites
go tool pprof -list=FuzzySearch -alloc_space /tmp/mem.prof

# Interactive
go tool pprof /tmp/mem.prof
# (pprof) top20
# (pprof) list fuzzy
```

### Benchmark Comparison

**Using benchstat**:
```bash
# Run baseline
go test -bench=. -count=10 ./internal/services/... > old.txt

# Make changes
# ...

# Run new version
go test -bench=. -count=10 ./internal/services/... > new.txt

# Compare
benchstat old.txt new.txt
```

**Interpreting Results**:
- `~` means no statistically significant change
- `+X%` means X% slower (regression)
- `-X%` means X% faster (improvement)
- P-values < 0.05 indicate statistical significance

## Go Benchmarking Best Practices

### 1. Benchmark Structure

```go
func BenchmarkOperation(b *testing.B) {
    // Setup (outside timer)
    data := setupTestData()
    
    // Reset timer (exclude setup)
    b.ResetTimer()
    
    // Run benchmark
    for i := 0; i < b.N; i++ {
        result := operation(data)
        _ = result // Prevent compiler optimization
    }
}
```

### 2. Memory Benchmarks

```go
func BenchmarkOperation_Memory(b *testing.B) {
    data := setupTestData()
    
    b.ReportAllocs() // Report allocation stats
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        result := operation(data)
        _ = result
    }
}
```

### 3. Sub-Benchmarks

```go
func BenchmarkOperation(b *testing.B) {
    sizes := []int{100, 1000, 10000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("%d_items", size), func(b *testing.B) {
            data := setupTestData(size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                result := operation(data)
                _ = result
            }
        })
    }
}
```

### 4. Preventing Compiler Optimizations

```go
var result int // Global to prevent optimization

func BenchmarkOperation(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = operation()
    }
    result = r // Assign to global
}
```

### 5. Benchmarking Tips

**DO**:
- Use `b.ResetTimer()` after setup
- Use `b.ReportAllocs()` for memory benchmarks
- Use `b.Run()` for sub-benchmarks
- Run with `-benchmem` flag
- Run multiple times (`-count=10`) for statistics
- Use `benchstat` for comparison

**DON'T**:
- Include setup in timer
- Forget to prevent compiler optimizations
- Run single iteration (use `-count`)
- Mix CPU and memory work
- Benchmark with insufficient iterations

## Recommendations for Replacement Implementation

Based on baseline analysis:

### 1. Eliminate DuckDB for Search

**Rationale**:
- Search is already in-memory (not using DuckDB)
- DuckDB overhead: 29% (19.5% CGO + 9.5% init)
- Binary size: 64 MB → 10-15 MB (78% reduction)
- Startup: 500ms → 50ms (10x improvement)

**Keep DuckDB for**:
- SQL query interface (optional feature)
- Advanced users who need SQL flexibility
- Or: Implement simpler query DSL without DuckDB

### 2. Optimize Fuzzy Search Algorithm

**Current Bottleneck**: fuzzy.FindFromNoSort (42% CPU time)

**Optimization Strategies**:
- Pre-allocate result slices (reduce 80k allocations)
- Use sync.Pool for fuzzy match structs
- Limit search to first 500 chars (already done)
- Consider faster ranking algorithm
- Index-based pre-filtering before fuzzy matching

**Expected Improvement**: 20-40% faster

### 3. Reduce Memory Allocations

**Current**: 80,029 allocations for 10k notes (8 per note)

**Strategies**:
- Pre-size slices based on input
- Reuse fuzzyMatch structs with sync.Pool
- Batch allocations where possible
- Use string builders instead of concatenation

**Expected Improvement**: 50% fewer allocations

### 4. Build Performance Improvements

**Remove CGO**:
- Pure Go implementation
- Trivial cross-compilation
- Faster builds (5s vs 15-20s)
- Simpler CI/CD

### 5. Testing Improvements

**With Afero Support**:
- Memory-backed filesystem tests
- No disk I/O overhead
- Better isolation
- Faster test execution

## Conclusion

Current OpenNotes performance is excellent for in-memory operations, exceeding all documented targets. However, DuckDB introduces:

1. **Binary bloat**: 64 MB (78% removable)
2. **Startup overhead**: 500ms (90% reducible)
3. **Runtime overhead**: 29% CPU time (eliminable)
4. **Build complexity**: CGO + platform bindings (simplifiable)

**Primary optimization opportunity**: Remove DuckDB for search operations while optionally retaining SQL query capability.

**Performance targets for replacement**:
- Match current search speed (< 30ms for 10k notes)
- Reduce binary size by 78% (64 MB → 10-15 MB)
- Reduce startup time by 90% (500ms → 50ms)
- Eliminate CGO overhead (29% → 0%)
- Improve memory efficiency (50% fewer allocations)

All targets are achievable with pure Go implementation using existing fuzzy search library or improved alternative.
