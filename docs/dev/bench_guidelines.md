# Benchmarking Guidelines

This document provides comprehensive benchmarking guidelines for the mux project. All performance-critical code should include benchmarks to track performance over time.

## Table of Contents

- [Benchmarking Philosophy](#benchmarking-philosophy)
- [Naming Conventions](#naming-conventions)
- [Benchmark Structure](#benchmark-structure)
- [Benchmark Helpers](#benchmark-helpers)
- [Middleware Benchmarks](#middleware-benchmarks)
- [Best Practices](#best-practices)
- [Running Benchmarks](#running-benchmarks)
- [Common Patterns](#common-patterns)

## Benchmarking Philosophy

### Why Benchmark?

1. **Track Performance** - Monitor performance changes over time
2. **Identify Regressions** - Catch performance degradations early
3. **Optimize Intelligently** - Know what to optimize and verify improvements
4. **Document Performance** - Show users what to expect

### What to Benchmark

- **Hot Paths** - Code executed frequently (routing, middleware invocation)
- **Core Algorithms** - Pattern matching, parsing, tokenization
- **Public APIs** - Entry points users will call
- **Known Bottlenecks** - Areas where performance matters
- **Trivial Code** - Simple getters/setters unless proven hot
- **Test Code** - Don't benchmark test helpers

## Naming Conventions

### Standard Format

All benchmarks must start with the `Benchmark` prefix:

```go
func BenchmarkRouterExactMatch(b *testing.B) { ... }
```

### Use PascalCase

```go
// Good - PascalCase, descriptive
func BenchmarkRouterExactMatch(b *testing.B)
func BenchmarkRouterParamMatch(b *testing.B)
func BenchmarkRouteRegistryManyRoutes(b *testing.B)
func BenchmarkCompressionGzipSmall(b *testing.B)

// Bad - snake_case, underscores
func Benchmark_router_exact_match(b *testing.B)
func BenchmarkRouter_ExactMatch(b *testing.B)
```

### Descriptive Names

Names should clearly describe **what is being measured**:

```go
// Good - Clear what's being tested
func BenchmarkMiddlewareInvoke(b *testing.B)
func BenchmarkRouterServeHTTP(b *testing.B)
func BenchmarkRegistryLookup(b *testing.B)

// Bad - Vague or unclear
func BenchmarkTest(b *testing.B)
func BenchmarkStuff(b *testing.B)
func BenchmarkPerformance(b *testing.B)
```

### Sub-Benchmarks

For variations (different sizes, modes, etc.), prefer sub-benchmarks instead of encoding details into the function name:

```go
// Good - Sub-benchmarks for variations
func BenchmarkCompression(b *testing.B) {
    b.Run("GzipSmall", func(b *testing.B) { ... })
    b.Run("GzipLarge", func(b *testing.B) { ... })
    b.Run("Brotli", func(b *testing.B) { ... })
}

// Bad - Variations in function name
func BenchmarkCompressionGzipSmall(b *testing.B) { ... }
func BenchmarkCompressionGzipLarge(b *testing.B) { ... }
func BenchmarkCompressionBrotli(b *testing.B) { ... }
```

## Benchmark Structure

### Standard Structure

Every benchmark should follow this pattern:

```go
func BenchmarkSomething(b *testing.B) {
    // Setup (outside the loop)
    data := setupTestData()
    
    // Important: Reset timer after expensive setup
    b.ReportAllocs()
    b.ResetTimer()
    
    // Benchmark loop
    for i := 0; i < b.N; i++ {
        // Code being measured
        _ = DoSomething(data)
    }
}
```

### Key Elements

1. **Setup** - Prepare data outside the loop
2. **b.ReportAllocs()** - Track memory allocations
3. **b.ResetTimer()** - Reset after expensive setup
4. **Loop** - Use `b.N` iterations
5. **Prevent optimization** - Assign result to blank identifier or package variable

### Preventing Compiler Optimization

The compiler might optimize away code that's never used:

```go
// Bad - Compiler might optimize away
func BenchmarkProcess(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Process(data) // Result discarded
    }
}

// Good - Result is used
var result int // Package-level var

func BenchmarkProcess(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = Process(data)
    }
    result = r // Prevent optimization
}

// Also good - Use blank identifier for side effects
func BenchmarkProcess(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = Process(data)
    }
}
```

## Benchmark Helpers

### Middleware Benchmark Helpers

To reduce boilerplate in middleware benchmarks, use the helpers in `internal/middlewarebench`:

```go
import "github.com/fgrzl/mux/internal/middlewarebench"

func BenchmarkMyMiddlewareInvoke(b *testing.B) {
    middleware := &myMiddleware{}
    
    middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, nil)
}
```

### Available Helpers

#### 1. BenchmarkMiddlewareInvoke

Benchmarks middleware's `Invoke` method in isolation:

```go
func BenchmarkMiddlewareInvoke(
    b *testing.B,
    invoke MiddlewareInvokeFunc,
    setupRequest func(*http.Request)
)
```

**Usage:**
```go
func BenchmarkCORSInvoke(b *testing.B) {
    middleware := newCORSMiddleware(options)
    
    b.Run("SimpleRequest", func(b *testing.B) {
        middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, nil)
    })
    
    b.Run("WithOrigin", func(b *testing.B) {
        middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, func(r *http.Request) {
            r.Header.Set("Origin", "https://example.com")
        })
    })
}
```

#### 2. BenchmarkMiddlewareRouterPipeline

Benchmarks middleware through a full router pipeline:

```go
func BenchmarkMiddlewareRouterPipeline(
    b *testing.B,
    rtr *router.Router,
    method, path string,
    setupRequest func(*http.Request)
)
```

**Usage:**
```go
func BenchmarkCORSRouterPipeline(b *testing.B) {
    rtr := router.NewRouter()
    UseCORS(rtr, options)
    rtr.GET("/test", func(c routing.RouteContext) {
        c.Response().WriteHeader(http.StatusOK)
    })
    
    middlewarebench.BenchmarkMiddlewareRouterPipeline(
        b, rtr, http.MethodGet, "/test", nil,
    )
}
```

#### 3. BenchmarkRouterPipelines

Runs multiple router pipeline scenarios (pooled/non-pooled):

```go
func BenchmarkRouterPipelines(
    b *testing.B,
    setupRouter func(*router.Router),
    cases []RouterPipelineCase,
    method, path string
)
```

**Usage:**
```go
func BenchmarkCORSPipeline(b *testing.B) {
    setupRouter := func(rtr *router.Router) {
        UseCORS(rtr, options)
        rtr.GET("/test", func(c routing.RouteContext) {
            c.NoContent()
        })
    }
    
    cases := []middlewarebench.RouterPipelineCase{
        {Name: "Standard", Pooled: false},
        {Name: "Pooled", Pooled: true},
        {Name: "WithHeaders", Pooled: false, SetupRequest: func(r *http.Request) {
            r.Header.Set("Origin", "https://example.com")
        }},
    }
    
    middlewarebench.BenchmarkRouterPipelines(b, setupRouter, cases, http.MethodGet, "/test")
}
```

### Benefits of Helpers

- **60-70% less code** - Reduces benchmark boilerplate
- **Consistent** - Same patterns across all middleware
- **Maintainable** - Changes happen in one place
- **Zero overhead** - Helpers are inline-friendly
- **Readable** - Intent is clear

## Middleware Benchmarks

### Required Benchmarks

Every middleware must include:

1. **Invoke Benchmark** - Middleware overhead in isolation
2. **Router Pipeline Benchmark** - Real-world usage with router
3. **Scenario Variations** - Different configurations/conditions

### Pattern 1: Invoke Benchmarks

Test middleware overhead without router:

```go
func BenchmarkMyMiddlewareInvoke(b *testing.B) {
    middleware := &myMiddleware{config: config}
    
    b.Run("PassThrough", func(b *testing.B) {
        middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, nil)
    })
    
    b.Run("WithProcessing", func(b *testing.B) {
        middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, func(r *http.Request) {
            r.Header.Set("X-Custom", "value")
        })
    })
}
```

### Pattern 2: Router Pipeline Benchmarks

Test middleware in real router:

```go
func BenchmarkMyMiddlewareRouterPipeline(b *testing.B) {
    rtr := router.NewRouter()
    UseMyMiddleware(rtr, options)
    rtr.GET("/test", func(c routing.RouteContext) {
        c.Response().WriteHeader(http.StatusOK)
    })
    
    middlewarebench.BenchmarkMiddlewareRouterPipeline(
        b, rtr, http.MethodGet, "/test", nil,
    )
}
```

### Pattern 3: Multiple Scenarios

Test different configurations:

```go
func BenchmarkMyMiddlewarePipeline(b *testing.B) {
    setupRouter := func(rtr *router.Router) {
        UseMyMiddleware(rtr, DefaultOptions())
        rtr.GET("/test", func(c routing.RouteContext) {
            c.NoContent()
        })
    }
    
    cases := []middlewarebench.RouterPipelineCase{
        {Name: "NonPooled", Pooled: false},
        {Name: "Pooled", Pooled: true},
        {Name: "WithCache", Pooled: false, SetupRequest: func(r *http.Request) {
            r.Header.Set("Cache-Control", "max-age=3600")
        }},
    }
    
    middlewarebench.BenchmarkRouterPipelines(b, setupRouter, cases, http.MethodGet, "/test")
}
```

## Best Practices

### Do's

1. **Use b.ReportAllocs()** - Always track allocations
2. **Reset timer after setup** - `b.ResetTimer()` after expensive setup
3. **Use sub-benchmarks** - `b.Run()` for variations
4. **Test realistic scenarios** - Use real-world data sizes
5. **Prevent optimization** - Store results to prevent compiler optimization
6. **Document expectations** - Add comments about expected performance
7. **Use helpers** - Leverage `internal/middlewarebench` for middleware
8. **Keep it simple** - Benchmark one thing at a time
9. **Run multiple times** - Verify consistency
10. **Commit baseline** - Track performance over time

### Don'ts

1. **Don't include setup in measurement** - Reset timer after setup
2. **Don't benchmark trivial code** - Focus on hot paths
3. **Don't forget b.ReportAllocs()** - Memory matters
4. **Don't use random data** - Makes results non-deterministic
5. **Don't ignore allocations** - Zero allocations in hot paths
6. **Don't benchmark unstable code** - Wait until code is settled
7. **Don't micro-optimize prematurely** - Profile first
8. **Don't compare across machines** - Results vary by hardware
9. **Don't forget edge cases** - Benchmark best/worst cases
10. **Don't duplicate helpers** - Use existing benchmark infrastructure

### Memory Allocations

Pay special attention to allocations in hot paths:

```go
func BenchmarkProcess(b *testing.B) {
    data := prepareData()
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _ = Process(data)
    }
}

// Results show allocations:
// BenchmarkProcess-8    1000000    1234 ns/op    512 B/op    4 allocs/op
// ^^^^^^^^^^^^^ Focus here
```

**Goal:** Zero allocations in hot paths (routing, middleware invocation)

## Running Benchmarks

### Basic Commands

```bash
# Run all benchmarks
go test ./... -bench=.

# Run specific benchmark
go test ./internal/router -bench=BenchmarkRouterExactMatch

# Run with memory stats
go test ./... -bench=. -benchmem

# Run longer for stability
go test ./... -bench=. -benchtime=5s

# Run and save results
go test ./... -bench=. -benchmem > bench.txt

# Compare before/after
go test ./... -bench=. -benchmem > old.txt
# Make changes
go test ./... -bench=. -benchmem > new.txt
benchstat old.txt new.txt
```

### Benchstat Comparison

Use `benchstat` to compare benchmark results:

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Compare results
benchstat old.txt new.txt
```

**Example output:**
```
name                    old time/op    new time/op    delta
RouterExactMatch-8        82.3ns +/- 2%    61.5ns +/- 1%  -25.27%  (p=0.000 n=10+10)

name                    old alloc/op   new alloc/op   delta
RouterExactMatch-8         112B +/- 0%        0B       -100.00%  (p=0.000 n=10+10)

name                    old allocs/op  new allocs/op  delta
RouterExactMatch-8         1.00 +/- 0%      0.00       -100.00%  (p=0.000 n=10+10)
```

### CI Integration

Benchmarks should be run in CI to track regressions:

```bash
# Run benchmarks without tests
go test ./... -bench=. -run=^$ -benchmem
```

## Common Patterns

### Pattern 1: Simple Benchmark

```go
func BenchmarkParseRoute(b *testing.B) {
    route := "/users/{id}/posts/{postId}"
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _ = ParseRoute(route)
    }
}
```

### Pattern 2: Sub-Benchmarks

```go
func BenchmarkRouter(b *testing.B) {
    routes := []string{
        "/exact",
        "/users/{id}",
        "/files/**",
    }
    
    for _, route := range routes {
        b.Run(route, func(b *testing.B) {
            rtr := setupRouter(route)
            req := httptest.NewRequest("GET", route, nil)
            rec := httptest.NewRecorder()
            
            b.ReportAllocs()
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                rtr.ServeHTTP(rec, req)
            }
        })
    }
}
```

### Pattern 3: Table-Driven Benchmarks

```go
func BenchmarkTokenize(b *testing.B) {
    tests := []struct {
        name  string
        input string
    }{
        {"Short", "/users"},
        {"Medium", "/users/{id}/posts"},
        {"Long", "/api/v1/users/{id}/posts/{postId}/comments/{commentId}"},
    }
    
    for _, tt := range tests {
        b.Run(tt.name, func(b *testing.B) {
            b.ReportAllocs()
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                _ = Tokenize(tt.input)
            }
        })
    }
}
```

### Pattern 4: Parallel Benchmarks

For testing concurrent performance:

```go
func BenchmarkRouterParallel(b *testing.B) {
    rtr := setupRouter()
    
    b.RunParallel(func(pb *testing.PB) {
        req := httptest.NewRequest("GET", "/test", nil)
        rec := httptest.NewRecorder()
        
        for pb.Next() {
            rtr.ServeHTTP(rec, req)
        }
    })
}
```

## File Organization

### File Naming

- **Benchmark files**: `*_bench_test.go` (e.g., `router_bench_test.go`)
- **Separate from tests**: Keep benchmarks in separate files for clarity

### Directory Structure

```
internal/router/
|- router.go
|- router_test.go                 # Unit tests
|- router_bench_test.go           # Benchmarks
`- router_comparison_bench_test.go # Comparison benchmarks
```

## Performance Goals

### Target Metrics

**Router:**
- Exact match: < 100ns/op, 0 allocs
- Param match: < 150ns/op, 0 allocs (with pooling)
- Wildcard: < 200ns/op, 0 allocs (with pooling)

**Middleware:**
- Invoke overhead: < 50ns/op, 0 allocs (pass-through)
- Pipeline: < 1us/op with 5 middleware

**Registry:**
- Lookup: < 100ns/op
- Registration: < 1us/op

## Examples

### Complete Middleware Benchmark

```go
package cors

import (
    "net/http"
    "testing"
    
    "github.com/fgrzl/mux/internal/middlewarebench"
    "github.com/fgrzl/mux/internal/router"
    "github.com/fgrzl/mux/internal/routing"
)

func BenchmarkCORSInvoke(b *testing.B) {
    middleware := newCORSMiddleware(CORSOptions{
        AllowedOrigins: []string{"https://example.com"},
    })
    
    b.Run("NoOrigin", func(b *testing.B) {
        middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, nil)
    })
    
    b.Run("WithOrigin", func(b *testing.B) {
        middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, func(r *http.Request) {
            r.Header.Set("Origin", "https://example.com")
        })
    })
}

func BenchmarkCORSRouterPipeline(b *testing.B) {
    rtr := router.NewRouter()
    UseCORS(rtr, CORSOptions{
        AllowedOrigins: []string{"*"},
    })
    rtr.GET("/test", func(c routing.RouteContext) {
        c.Response().WriteHeader(http.StatusOK)
    })
    
    middlewarebench.BenchmarkMiddlewareRouterPipeline(
        b, rtr, http.MethodGet, "/test", nil,
    )
}
```

## Resources

- [Go Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)
- [Benchstat Tool](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [Middleware Benchmark Helpers](../../internal/middlewarebench/helpers.go)

---

**Remember:** Benchmarks are your performance safety net. Write them for critical paths, run them regularly, and use them to guide optimization decisions.

