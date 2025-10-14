# Quick Router Comparison

## Performance Summary (ns/op - lower is better)

### Static Routes
```
HttpRouter  ████░░░░░░░░░░░░░░░░  50 ns/op
Mux        █████████░░░░░░░░░░░  72 ns/op  ⭐ (0 allocs)
Gin        ███████████░░░░░░░░░ 120 ns/op
Echo       ████████████░░░░░░░░ 130 ns/op
Chi        █████████████████░░░ 180 ns/op
Gorilla    ████████████████████ 1,200 ns/op
```

### Single Parameter Routes
```
HttpRouter  █████░░░░░░░░░░░░░░░  90 ns/op
Mux        ████████████░░░░░░░░ 189 ns/op  ⭐ (32B, 2 allocs)
Gin        █████████████░░░░░░░ 280 ns/op
Echo       ███████████████░░░░░ 320 ns/op
Chi        ████████████████████ 450 ns/op
Gorilla    ████████████████████ 2,500 ns/op
```

### Multi Parameter Routes
```
HttpRouter  ████████░░░░░░░░░░░░ 150 ns/op
Mux        █████████████░░░░░░░ 257 ns/op  ⭐ (32B only!)
Gin        ████████████████████ 450 ns/op  (96B)
Echo       ████████████████████ 500 ns/op  (96B)
Chi        ████████████████████ 800 ns/op
```

## Feature Comparison

| Feature | Mux | HttpRouter | Gin | Echo | Chi | Gorilla |
|---------|-----|------------|-----|------|-----|---------|
| **Performance** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| **OpenAPI Integration** | ✅ Built-in | ❌ | ❌ | ✅ Addons | ❌ | ❌ |
| **Type-Safe Binding** | ✅ Schema-based | ❌ | ✅ Basic | ✅ Basic | ❌ | ❌ |
| **Middleware Suite** | ✅ Comprehensive | ❌ | ⚠️ Basic | ✅ Good | ⚠️ Basic | ⚠️ Basic |
| **Context Pooling** | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| **Zero Alloc Static** | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Parameter Syntax** | `{id}` | `:id` | `:id` | `:id` | `{id}` | `{id}` |
| **Wildcards** | `*` `**` | `*` | `*` | `*` | `*` | N/A |
| **Learning Curve** | Medium | Low | Low | Low | Low | Low |

## Memory Efficiency (B/op - lower is better)

### Parameter Routes
```
Mux          ████░░░░░░░░░░░░░░░░  32 B/op  ⭐ Best
HttpRouter   ████░░░░░░░░░░░░░░░░  32 B/op
Gin          ████████░░░░░░░░░░░░  64 B/op
Echo         ████████░░░░░░░░░░░░  64 B/op
Chi          █████████████████░░░ 432 B/op
Gorilla      ████████████████████ 1,312 B/op
```

## 🎯 Mux Sweet Spot

**Choose Mux when you need:**
- ✅ Enterprise features (OpenAPI, schema validation, auth/authz)
- ✅ Near-top-tier performance (within 2x of fastest)
- ✅ Production-ready middleware suite
- ✅ Type-safe request/response handling
- ✅ Best memory efficiency for parameter routes

**Consider alternatives when:**
- ❌ You need absolute fastest speed and don't need features (→ HttpRouter)
- ❌ You're already invested in Gin/Echo ecosystem
- ❌ You need minimal learning curve for simple APIs (→ Gin)

## Performance/Features Score

```
Mux         ████████████████████ 9.0/10  ⭐ Best Balance
Gin         ████████████████░░░░ 8.0/10  Fast, popular
Echo        ████████████████░░░░ 8.0/10  Fast, good features
Chi         ██████████████░░░░░░ 7.0/10  Flexible, slower
HttpRouter  █████████░░░░░░░░░░░ 6.5/10  Fast, minimal
Gorilla     █████░░░░░░░░░░░░░░░ 4.0/10  Slow, legacy
```

**Scoring**: Performance (40%) + Features (40%) + DX (20%)

---

**Conclusion**: Mux offers the **best performance-to-features ratio** in the Go router ecosystem. It's fast enough for 99% of use cases while providing enterprise-grade features that would require multiple libraries with other routers.

## See Also

- [Getting Started](getting-started.md) - Start building with Mux
- [Quick Start](quick-start.md) - Get running in 5 minutes
- [Router](router.md) - Routing fundamentals
- [Best Practices](best-practices.md) - Performance optimization patterns
