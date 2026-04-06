# Quick Router Comparison

This page summarizes the benchmark snapshots used in the docs. Exact numbers will vary by hardware and workload, but the relative tradeoffs are the important part.

## Performance Summary

### Static Routes

| Router | ns/op | Notes |
|--------|------:|-------|
| HttpRouter | 50 | Fastest in this sample |
| Mux | 72 | 0 allocs |
| Gin | 120 | |
| Echo | 130 | |
| Chi | 180 | |
| Gorilla | 1200 | Legacy baseline |

### Single Parameter Routes

| Router | ns/op | Notes |
|--------|------:|-------|
| HttpRouter | 90 | |
| Mux | 189 | 32 B/op, 2 allocs |
| Gin | 280 | |
| Echo | 320 | |
| Chi | 450 | |
| Gorilla | 2500 | Legacy baseline |

### Multi Parameter Routes

| Router | ns/op | Notes |
|--------|------:|-------|
| HttpRouter | 150 | |
| Mux | 257 | 32 B/op only |
| Gin | 450 | 96 B/op |
| Echo | 500 | 96 B/op |
| Chi | 800 | |

## Feature Comparison

| Feature | Mux | HttpRouter | Gin | Echo | Chi | Gorilla |
|---------|-----|------------|-----|------|-----|---------|
| Performance | Strong | Excellent | Strong | Strong | Moderate | Weak |
| OpenAPI integration | Built in | No | No | Addons | No | No |
| Type-safe binding | Schema-based | No | Basic | Basic | No | No |
| Middleware suite | Comprehensive | Minimal | Basic | Good | Basic | Basic |
| Context pooling | Yes | No | Yes | Yes | No | No |
| Zero-allocation static routes | Yes | Yes | Yes | Yes | No | No |
| Parameter syntax | `{id}` | `:id` | `:id` | `:id` | `{id}` | `{id}` |
| Wildcards | `*`, `**` | `*` | `*` | `*` | `*` | N/A |
| Learning curve | Medium | Low | Low | Low | Low | Low |

## Memory Efficiency

### Parameter Routes

| Router | B/op | Notes |
|--------|-----:|-------|
| Mux | 32 | Best memory profile in this sample |
| HttpRouter | 32 | Matches Mux |
| Gin | 64 | |
| Echo | 64 | |
| Chi | 432 | |
| Gorilla | 1312 | Legacy baseline |

## Mux Sweet Spot

Choose Mux when you need:

- OpenAPI generation, schema validation, or integrated auth/authz support
- Near-top-tier routing performance without stitching together extra libraries
- A production-ready middleware suite and predictable request lifecycle
- Strong memory efficiency on parameter-heavy routes

Consider alternatives when:

- You need the absolute fastest raw router and do not need framework features: HttpRouter
- You are already deeply invested in the Gin or Echo ecosystem
- You want the smallest possible learning curve for a simple API: Gin

## Performance/Features Score

| Router | Score | Summary |
|--------|------:|---------|
| Mux | 9.0/10 | Best balance of speed and features |
| Gin | 8.0/10 | Fast and popular |
| Echo | 8.0/10 | Fast with a solid feature set |
| Chi | 7.0/10 | Flexible but slower |
| HttpRouter | 6.5/10 | Very fast, intentionally minimal |
| Gorilla | 4.0/10 | Slow and legacy |

**Scoring**: Performance (40%) + Features (40%) + DX (20%)

---

**Conclusion**: Mux offers one of the strongest performance-to-features tradeoffs in the Go router ecosystem. It is fast enough for most workloads while keeping OpenAPI, binding, middleware, and production setup in one coherent API.

## See Also

- [Getting Started](getting-started.md) - Start building with Mux
- [Quick Start](quick-start.md) - Get running in 5 minutes
- [Router](router.md) - Routing fundamentals
- [Best Practices](best-practices.md) - Performance optimization patterns
