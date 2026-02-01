# DuckDB Performance Baseline - Verification & Methodology

## Verification Objectives

This document provides:
1. **Reproduction instructions** for all benchmark results
2. **Validation procedures** to confirm findings
3. **Methodology documentation** for benchmark design
4. **Quality assurance** checks for data accuracy

## Environment Setup

### Prerequisites

```bash
# Required tools
go version  # Must be Go 1.21+
pprof --version  # Part of Go toolchain
benchstat --version  # Install: go install golang.org/x/perf/cmd/benchstat@latest

# Optional but recommended
graphviz --version  # For flame graphs (apt install graphviz)
```

### Repository Setup

```bash
# Clone repository
cd /mnt/Store/Projects/Mine/Github/opennotes

# Verify Go version
go version
# Expected: go version go1.24.7 linux/amd64 (or higher)

# Verify dependencies
go mod download
go mod verify

# Build to confirm environment
mise run build
# Expected: Binary at dist/opennotes (~64MB)
```

## Benchmark Reproduction Steps

### 1. Basic Benchmark Run

```bash
# Run all benchmarks with memory reporting
cd /mnt/Store/Projects/Mine/Github/opennotes
go test -bench=. -benchmem -run=^$ ./internal/services/...

# Expected output: ~32 benchmark results
# Total time: ~45-50 seconds
```

**Verification Checklist**:
- [ ] All benchmarks pass (PASS at end)
- [ ] Fuzzy search 10k: ~25-35ms (±20% variance acceptable)
- [ ] Text search 10k: ~2.5-4ms (±20% variance acceptable)
- [ ] No benchmark panics or errors
- [ ] Total time: 40-60 seconds

### 2. CPU Profiling

```bash
# Profile fuzzy search (most expensive operation)
go test -bench=BenchmarkFuzzySearch_10kNotes \
  -benchmem \
  -cpuprofile=/tmp/cpu_fuzzy.prof \
  ./internal/services/...

# Verify profile was created
ls -lh /tmp/cpu_fuzzy.prof
# Expected: ~100-500KB file size

# Analyze top functions
go tool pprof -top -cum /tmp/cpu_fuzzy.prof | head -30

# Expected top consumers:
# - fuzzy.FindFromNoSort (~40-45%)
# - runtime.cgocall (~15-25%)
# - duckdb operations (~8-12%)
```

**Verification Checklist**:
- [ ] Profile file created successfully
- [ ] Fuzzy matching is top consumer (35-50%)
- [ ] CGO overhead visible (15-25%)
- [ ] DuckDB operations present (8-15%)
- [ ] No unexpected functions in top 10

### 3. Memory Profiling

```bash
# Profile memory allocations
go test -bench=BenchmarkFuzzySearch_10kNotes \
  -benchmem \
  -memprofile=/tmp/mem_fuzzy.prof \
  ./internal/services/...

# Analyze top allocators
go tool pprof -top -alloc_space /tmp/mem_fuzzy.prof | head -30

# Expected top allocators:
# - fuzzy.FindFromNoSort (~40-50MB, ~40-45%)
# - fuzzy.Find/FindFrom (~15-20MB, ~10-15%)
# - SearchService.FuzzySearch (~10-15MB, ~8-10%)
```

**Verification Checklist**:
- [ ] Profile file created successfully
- [ ] Fuzzy matching is top allocator (35-50%)
- [ ] Total allocations: ~110-130MB for benchmark
- [ ] Allocation counts align with benchmem output
- [ ] No memory leaks (profile shows expected cleanup)

### 4. Text Search Profiling

```bash
# Profile text search for comparison
go test -bench=BenchmarkTextSearch_10kNotes \
  -benchmem \
  -cpuprofile=/tmp/cpu_text.prof \
  -memprofile=/tmp/mem_text.prof \
  ./internal/services/...

# Compare with fuzzy search
go tool pprof -top /tmp/cpu_text.prof | head -20
go tool pprof -top -alloc_space /tmp/mem_text.prof | head -20
```

**Verification Checklist**:
- [ ] Text search is ~9-10x faster than fuzzy
- [ ] Fewer allocations per operation
- [ ] Different CPU profile (no fuzzy matching overhead)
- [ ] More predictable performance

### 5. Benchmark Stability Testing

```bash
# Run benchmarks 10 times for statistical analysis
go test -bench=BenchmarkFuzzySearch_10kNotes \
  -benchmem \
  -count=10 \
  ./internal/services/... > /tmp/stability.txt

# Analyze variance
cat /tmp/stability.txt | grep BenchmarkFuzzySearch_10kNotes

# Expected variance:
# - Time/op: ±5-15% (should be fairly stable)
# - Memory/op: ±1% (should be very stable)
# - Allocs/op: 0% (should be identical every time)
```

**Verification Checklist**:
- [ ] All 10 runs complete successfully
- [ ] Time variance < 20% (CPU-dependent)
- [ ] Memory variance < 5%
- [ ] Allocation count identical across runs
- [ ] No increasing trend (no memory leaks)

### 6. Binary Size Verification

```bash
# Check binary size
ls -lh dist/opennotes
# Expected: ~60-68MB (64MB typical)

# Analyze binary composition
go tool nm -size dist/opennotes | sort -rn | head -50

# Check DuckDB symbols
go tool nm dist/opennotes | grep -i duckdb | wc -l
# Expected: Many (hundreds to thousands)

# Estimate DuckDB contribution
size dist/opennotes
# Text + Data + BSS should show large segments
```

**Verification Checklist**:
- [ ] Binary size: 60-68MB range
- [ ] DuckDB symbols present (many)
- [ ] Binary is statically linked (no external deps)
- [ ] Executable runs without errors

### 7. Build Time Verification

```bash
# Clean build timing
cd /mnt/Store/Projects/Mine/Github/opennotes
go clean -cache
time mise run build

# Expected: 15-25 seconds (depends on CPU)

# Incremental build timing
touch cmd/root.go
time mise run build

# Expected: 2-5 seconds (cached)
```

**Verification Checklist**:
- [ ] Clean build: 10-30 seconds
- [ ] Incremental build: < 5 seconds
- [ ] Build succeeds without errors
- [ ] CGO compilation visible in output

### 8. Test Suite Performance

```bash
# Run full test suite with timing
time mise run test

# Expected: ~4-6 seconds for all tests
# Expected: 161+ tests passing

# Test breakdown
go test -v ./internal/services/... 2>&1 | grep -E "^(PASS|FAIL)" | wc -l
# Expected: 161+ test results
```

**Verification Checklist**:
- [ ] All tests pass
- [ ] Total time: 3-8 seconds
- [ ] 161+ tests executed
- [ ] No race conditions detected
- [ ] No flaky tests

## Validation Procedures

### Benchmark Accuracy Validation

**Statistical Significance**:
```bash
# Run benchstat comparison against self
go test -bench=. -count=10 ./internal/services/... > baseline1.txt
go test -bench=. -count=10 ./internal/services/... > baseline2.txt
benchstat baseline1.txt baseline2.txt

# Expected: Most benchmarks show "~" (no significant difference)
# If many show +/- changes, system has high variance
```

**Success Criteria**:
- ≥ 80% of benchmarks show `~` (no significant change)
- No benchmark shows > 20% difference between identical runs
- P-values > 0.05 for most benchmarks

### Profiling Accuracy Validation

**CPU Profile Verification**:
```bash
# Run profile twice
go test -bench=BenchmarkFuzzySearch_10kNotes -cpuprofile=/tmp/prof1.prof ./internal/services/...
go test -bench=BenchmarkFuzzySearch_10kNotes -cpuprofile=/tmp/prof2.prof ./internal/services/...

# Compare top functions
go tool pprof -top /tmp/prof1.prof > /tmp/top1.txt
go tool pprof -top /tmp/prof2.prof > /tmp/top2.txt
diff /tmp/top1.txt /tmp/top2.txt

# Expected: Similar top functions (order may vary slightly)
```

**Success Criteria**:
- Same top 5 functions in both profiles
- Percentage variance < 10% for top functions
- No unexpected functions appear/disappear

**Memory Profile Verification**:
```bash
# Run memory profile twice
go test -bench=BenchmarkFuzzySearch_10kNotes -memprofile=/tmp/mem1.prof ./internal/services/...
go test -bench=BenchmarkFuzzySearch_10kNotes -memprofile=/tmp/mem2.prof ./internal/services/...

# Compare allocation totals
go tool pprof -top -alloc_space /tmp/mem1.prof | head -1
go tool pprof -top -alloc_space /tmp/mem2.prof | head -1

# Expected: Within 5% of each other
```

**Success Criteria**:
- Total allocations within 5%
- Same top allocators in both profiles
- Allocation counts match benchmem output

### Cross-Platform Verification

**Linux (Primary Platform)**:
```bash
# Already verified above
GOOS=linux GOARCH=amd64 mise run build
ls -lh dist/opennotes
# Expected: ~64MB
```

**macOS (If Available)**:
```bash
GOOS=darwin GOARCH=arm64 go build -o dist/opennotes-darwin ./cmd/opennotes
ls -lh dist/opennotes-darwin
# Expected: Similar size to Linux (~60-70MB)
```

**Windows (Cross-Compile)**:
```bash
GOOS=windows GOARCH=amd64 go build -o dist/opennotes.exe ./cmd/opennotes
ls -lh dist/opennotes.exe
# Expected: Similar size to Linux (~60-70MB)
```

**Note**: Benchmarks should be run on target platform for accuracy. Cross-compilation is for size verification only.

## Benchmark Methodology

### Design Principles

**1. Realistic Data Generation**

```go
func createBenchmarkNotes(count int) []Note {
    notes := make([]Note, count)
    // Use realistic metadata distributions
    tags := []string{"workflow", "meeting", "project", ...}
    statuses := []string{"active", "done", "archived", ...}
    
    for i := 0; i < count; i++ {
        // Realistic content (variable size)
        content := "This is test content " + strings.Repeat("x", i%100)
        
        // Realistic links (some notes, not all)
        links := []string{}
        if i%5 == 0 { links = append(links, "epics/epic-001.md") }
        
        notes[i] = Note{Content: content, Metadata: map[string]any{...}}
    }
    return notes
}
```

**Why**: Synthetic data should mimic real-world note collections to ensure benchmark relevance.

**2. Multiple Dataset Sizes**

- **100 notes**: Small collections (personal use)
- **1,000 notes**: Medium collections (active projects)
- **10,000 notes**: Large collections (long-term knowledge base)

**Why**: Performance characteristics may change with scale. Linear scaling is not guaranteed.

**3. Separate Setup from Benchmark**

```go
func BenchmarkOperation(b *testing.B) {
    data := setupTestData()  // Outside timer
    
    b.ResetTimer()  // Start timing here
    
    for i := 0; i < b.N; i++ {
        result := operation(data)
        _ = result
    }
}
```

**Why**: Setup cost should not be included in operation timing. Only measure what you're testing.

**4. Memory Allocation Tracking**

```go
func BenchmarkOperation_Memory(b *testing.B) {
    data := setupTestData()
    
    b.ReportAllocs()  // Track allocations
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        result := operation(data)
        _ = result
    }
}
```

**Why**: Memory allocation patterns are as important as CPU time for Go applications.

**5. Sub-Benchmarks for Variants**

```go
func BenchmarkSearch(b *testing.B) {
    sizes := []int{100, 1000, 10000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("%d_notes", size), func(b *testing.B) {
            notes := createBenchmarkNotes(size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = search(notes)
            }
        })
    }
}
```

**Why**: Sub-benchmarks make it easy to run specific tests and see scaling behavior.

### Benchmark Naming Convention

**Pattern**: `Benchmark<Component>_<Operation>_<Variant>`

Examples:
- `BenchmarkFuzzySearch_10kNotes` - Fuzzy search with 10k notes
- `BenchmarkTextSearch_Memory` - Memory allocation for text search
- `BenchmarkBooleanQuery_Complex` - Complex boolean query building

**Why**: Clear naming makes benchmark output easy to parse and understand.

### Statistical Rigor

**Run Count**:
```bash
# Minimum: 5 runs for averaging
go test -bench=. -count=5 ./internal/services/...

# Better: 10 runs for statistical analysis
go test -bench=. -count=10 ./internal/services/...

# Use benchstat for comparison
benchstat old.txt new.txt
```

**Why**: Single runs have high variance. Multiple runs enable statistical confidence.

**Variance Analysis**:
- Coefficient of Variation (CV) = StdDev / Mean
- CV < 10%: Low variance (good)
- CV 10-20%: Moderate variance (acceptable)
- CV > 20%: High variance (investigate)

### Profiling Best Practices

**CPU Profiling**:
```bash
# Profile specific benchmark
go test -bench=BenchmarkTarget -cpuprofile=cpu.prof ./...

# Analyze with pprof
go tool pprof cpu.prof
# (pprof) top20      # Top 20 functions by time
# (pprof) top -cum   # Top by cumulative time
# (pprof) list Func  # Source code for function
# (pprof) web        # Visual graph (requires graphviz)
```

**Memory Profiling**:
```bash
# Profile allocations
go test -bench=BenchmarkTarget -memprofile=mem.prof ./...

# Analyze allocations
go tool pprof -alloc_space mem.prof
# (pprof) top20      # Top allocators
# (pprof) list Func  # Allocation sites

# Analyze objects in use (different from alloc_space)
go tool pprof -inuse_space mem.prof
```

**Flame Graphs** (visual profiling):
```bash
# Generate interactive flame graph
go tool pprof -http=:8080 cpu.prof

# Opens browser with:
# - Flame graph view
# - Top functions
# - Source code view
# - Call graph
```

### Common Pitfalls to Avoid

**1. Including Setup in Timer**
```go
// WRONG
func BenchmarkBad(b *testing.B) {
    for i := 0; i < b.N; i++ {
        data := setupData()  // Setup inside loop!
        result := operation(data)
        _ = result
    }
}

// CORRECT
func BenchmarkGood(b *testing.B) {
    data := setupData()  // Setup outside
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result := operation(data)
        _ = result
    }
}
```

**2. Compiler Optimizations**
```go
// WRONG (compiler may optimize away)
func BenchmarkBad(b *testing.B) {
    for i := 0; i < b.N; i++ {
        result := operation()
        // result not used - may be optimized out
    }
}

// CORRECT
func BenchmarkGood(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = operation()
    }
    result = r  // Assign to global
}
```

**3. Shared State Between Iterations**
```go
// WRONG (iterations affect each other)
func BenchmarkBad(b *testing.B) {
    cache := make(map[string]string)
    for i := 0; i < b.N; i++ {
        operation(cache)  // Cache grows each iteration
    }
}

// CORRECT
func BenchmarkGood(b *testing.B) {
    for i := 0; i < b.N; i++ {
        cache := make(map[string]string)
        operation(cache)
    }
}
```

**4. Insufficient Warm-up**
```go
// For operations with caching/lazy init
func BenchmarkWithWarmup(b *testing.B) {
    data := setupData()
    
    // Warm up caches
    for i := 0; i < 100; i++ {
        _ = operation(data)
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = operation(data)
    }
}
```

## Quality Assurance Checklist

### Pre-Benchmark Verification

- [ ] Clean environment (no background processes affecting CPU)
- [ ] Sufficient memory available (no swapping)
- [ ] Go version matches requirements (1.21+)
- [ ] Dependencies up to date (`go mod download`)
- [ ] Code compiles without warnings
- [ ] Tests pass without errors

### During Benchmark Execution

- [ ] CPU not throttling (check `cpupower frequency-info` on Linux)
- [ ] No other intensive processes running
- [ ] Stable system load (< 1.0 on single core systems)
- [ ] Adequate disk space for profiles
- [ ] No network operations in benchmarks
- [ ] Consistent power mode (not battery saving)

### Post-Benchmark Validation

- [ ] All benchmarks completed successfully
- [ ] No panics or errors in output
- [ ] Results align with documented targets
- [ ] Profiles generated successfully
- [ ] Variance within acceptable range (< 20%)
- [ ] Results reproducible (within 10%)

### Regression Detection

- [ ] Compare against baseline (use benchstat)
- [ ] No >10% regression in critical paths
- [ ] Memory usage not increasing unexpectedly
- [ ] Allocation counts stable or improving
- [ ] No new bottlenecks introduced

## Troubleshooting Guide

### Benchmark Variance Too High (>20%)

**Causes**:
1. System load from other processes
2. CPU frequency scaling/throttling
3. Thermal throttling
4. Disk I/O interference
5. Network activity

**Solutions**:
```bash
# Check system load
top  # Should be mostly idle

# Disable CPU frequency scaling (Linux)
sudo cpupower frequency-set -g performance

# Check thermal status
sensors  # Should not be near thermal limit

# Run with higher priority
sudo nice -n -20 go test -bench=. ./...
```

### Profile Files Not Generated

**Causes**:
1. Insufficient permissions
2. Disk full
3. Invalid file path
4. Benchmark not running to completion

**Solutions**:
```bash
# Check permissions
ls -ld /tmp
# Should be writable

# Check disk space
df -h /tmp

# Use absolute path
go test -bench=. -cpuprofile=$(pwd)/cpu.prof ./...

# Ensure benchmark completes
go test -bench=. -timeout=10m ./...
```

### Inconsistent Results Between Runs

**Causes**:
1. GC timing variance
2. Memory allocation patterns
3. CPU cache effects
4. Context switching

**Solutions**:
```bash
# Increase sample count
go test -bench=. -count=20 ./...

# Use benchstat for statistical analysis
benchstat -alpha=0.1 results.txt

# Run longer benchmarks
go test -bench=. -benchtime=10s ./...
```

### Memory Profile Shows Unexpected Allocations

**Causes**:
1. Test infrastructure allocations
2. Logging or debugging code
3. Benchmark setup not excluded
4. String concatenation

**Solutions**:
- Use `b.ResetTimer()` after setup
- Disable logging in benchmarks
- Use `strings.Builder` instead of `+`
- Pre-allocate slices with `make([]T, 0, capacity)`

## Reproduction Guarantee

To ensure these results are reproducible:

1. **Environment Variables**:
   ```bash
   export GOMAXPROCS=16  # Match CPU thread count
   export GODEBUG=gctrace=0  # Disable GC trace noise
   ```

2. **Consistent Go Version**:
   ```bash
   go version
   # Must be go1.24.7 or compatible
   ```

3. **Clean State**:
   ```bash
   go clean -cache -testcache -modcache
   go mod download
   ```

4. **Run Command**:
   ```bash
   cd /mnt/Store/Projects/Mine/Github/opennotes
   go test -bench=. -benchmem -run=^$ ./internal/services/... | tee benchmark_results.txt
   ```

5. **Expected Variance**:
   - Time/op: ±10-20% (CPU-dependent)
   - Memory/op: ±1-2% (should be very stable)
   - Allocs/op: 0% (exact match expected)

## Conclusion

This verification methodology ensures:
- **Reproducibility**: Results can be independently verified
- **Accuracy**: Profiling data is statistically sound
- **Reliability**: Benchmarks measure what they claim
- **Consistency**: Results stable across runs

All benchmark results in research.md have been validated using these procedures.
