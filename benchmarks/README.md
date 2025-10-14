# Benchmark Results

This directory contains baseline and historical benchmark results for the mux router.

## Latest Baseline

**Date:** October 14, 2025  
**File:** `baseline_20251014.txt`  
**System:** 12th Gen Intel(R) Core(TM) i9-12900HK (20 cores)

### Quick Results

| Scenario | Performance | Allocations |
|----------|-------------|-------------|
| Static Route | 70ns | 0B / 0 allocs |
| Single Param | 190ns | 32B / 2 allocs |
| Multi Param | 262ns | 32B / 2 allocs |
| Deep Path | 344ns | 32B / 2 allocs |
| Wildcard | 73ns | 0B / 0 allocs |

**All tests run with context pooling enabled (`WithContextPooling()`).**

## Running Benchmarks

### Full Comparison Suite
```bash
cd pkg/router
go test -bench=BenchmarkComparison -benchmem -benchtime=5s
```

### Specific Scenario
```bash
go test -bench=BenchmarkComparisonMuxStaticRoute -benchmem -benchtime=5s
```

### With Profiling
```bash
# CPU profile
go test -bench=BenchmarkComparison -cpuprofile=cpu.prof -benchtime=5s
go tool pprof -http=:8080 cpu.prof

# Memory profile
go test -bench=BenchmarkComparison -memprofile=mem.prof -benchtime=5s
go tool pprof -http=:8080 mem.prof
```

### Save Results
```bash
# PowerShell
go test -bench=BenchmarkComparison -benchmem -benchtime=5s | Tee-Object -FilePath ../../benchmarks/result_$(Get-Date -Format 'yyyyMMdd').txt

# Bash
go test -bench=BenchmarkComparison -benchmem -benchtime=5s | tee ../../benchmarks/result_$(date +%Y%m%d).txt
```

## Comparing Results

### Using benchstat
```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Compare two runs
benchstat baseline_20251014.txt result_20251015.txt
```

### Example Output
```
name                         old time/op    new time/op    delta
ComparisonMuxStaticRoute-20    75.0ns ± 2%    50.0ns ± 1%  -33.33%  (p=0.000 n=10+10)

name                         old alloc/op   new alloc/op   delta
ComparisonMuxStaticRoute-20     0.00B          0.00B         ~     (all equal)
```

## Benchmark Guidelines

### Best Practices
1. **Use consistent benchtime:** `-benchtime=5s` minimum
2. **Run multiple times:** `-count=10` for statistical significance
3. **Disable CPU frequency scaling:** For consistent results
4. **Close other applications:** Reduce system noise
5. **Use same hardware:** For meaningful comparisons

### Interpreting Results

**Time (ns/op):**
- Lower is better
- <100ns: Excellent
- 100-500ns: Good
- >500ns: Room for improvement

**Allocations (B/op, allocs/op):**
- 0 allocs: Perfect (with pooling)
- <100B: Good
- >100B: Check for unnecessary allocations

### What to Benchmark

**Core routing:**
- Static routes (most common)
- Single parameter routes
- Multi-parameter routes
- Wildcard/catch-all routes

**Scalability:**
- Many routes (1000+)
- Deep paths (5+ segments)

**Middleware:**
- Pipeline overhead
- Common middleware (logging, auth, etc.)

## Historical Notes

### October 14, 2025 - Baseline Established
- Fixed critical benchmark bug in wildcard route
  - Changed `/files/*` to `/files/**` to match test path
  - Result: 1473ns → 73ns (20x improvement!)
- Confirmed zero-allocation performance with context pooling
- Established competitive positioning vs. Gin/Echo/HttpRouter

### Key Findings
1. ✅ Wildcard performance: Competitive with best routers (73ns, 0 allocs)
2. ✅ Static routes: Production-ready (70ns, 0 allocs)
3. ✅ Context pooling: Critical for best performance (3x throughput boost)
4. 🎯 Optimization opportunity: Static route fast path (20-25ns gain expected)

## Benchmark Scenarios Explained

### Static Routes
```go
// Pattern: /ping
// Request: /ping
// Tests: Exact match, no parameters
```
Most common case in APIs (health checks, root endpoints).

### Single Parameter
```go
// Pattern: /users/{id}
// Request: /users/12345
// Tests: One path parameter extraction
```
Very common in REST APIs (resource by ID).

### Multi Parameter
```go
// Pattern: /users/{userId}/posts/{postId}
// Request: /users/12345/posts/67890
// Tests: Multiple parameters, deeper path
```
Common in nested resources.

### Deep Path
```go
// Pattern: /api/v1/organizations/{orgId}/projects/{projectId}
// Request: /api/v1/organizations/org-123/projects/proj-456
// Tests: Many segments, multiple params
```
Real-world complex APIs.

### Wildcard (Catch-All)
```go
// Pattern: /files/**
// Request: /files/images/logo.png
// Tests: Catch-all for remaining path
```
Used for static file serving, SPA fallback.

### Many Routes
```go
// 1000+ routes registered
// Tests: Scalability, trie depth
```
Ensures performance doesn't degrade with route count.

## See Also

- `PERFORMANCE_ANALYSIS.md` - Detailed analysis and findings
- `PERFORMANCE_SUMMARY.md` - Executive summary and recommendations
- `PERFORMANCE_CHART.md` - Visual comparisons
- `OPTIMIZATION_PLAN.md` - Optimization roadmap
- `docs/dev/bench_guidelines.md` - Full benchmarking guidelines
