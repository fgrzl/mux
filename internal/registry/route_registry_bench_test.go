package registry

import (
	"io"
	"log/slog"
	"strconv"
	"testing"

	"github.com/fgrzl/mux/internal/routing"
)

func init() {
	// silence slog during benchmarks to avoid noisy output
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
}

// registerNoop is a helper to register a route with a no-op handler into the registry.
// The handler is intentionally empty because benchmarks measure only routing performance,
// not handler execution time.
func registerNoop(r *RouteRegistry, pattern string, method string) {
	opts := &routing.RouteOptions{
		Method:  method,
		Pattern: pattern,
		Handler: func(c routing.RouteContext) { /* empty by design for benchmarking */ },
	}
	r.Register(pattern, method, opts)
}

func BenchmarkRouteRegistryLoadExactMatch(b *testing.B) {
	r := NewRouteRegistry()
	registerNoop(r, "/hello", "GET")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.LoadIntoSlice("/hello", "GET", nil)
	}
}

func BenchmarkRouteRegistryLoadParamMatch(b *testing.B) {
	r := NewRouteRegistry()
	registerNoop(r, "/users/{id}", "GET")
	var params routing.Params

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.LoadIntoSlice("/users/12345", "GET", &params)
	}
}

func BenchmarkRouteRegistryLoadWildcard(b *testing.B) {
	r := NewRouteRegistry()
	registerNoop(r, "/files/*", "GET")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.LoadIntoSlice("/files/path/to/file.txt", "GET", nil)
	}
}

func BenchmarkRouteRegistryLoadCatchAll(b *testing.B) {
	r := NewRouteRegistry()
	registerNoop(r, "/catch/**", "GET")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.LoadIntoSlice("/catch/any/number/of/segments/here", "GET", nil)
	}
}

// helper to populate registry with n routes
func populateRegistry(r *RouteRegistry, n int) {
	for i := 0; i < n; i++ {
		registerNoop(r, "/static/route/"+strconv.Itoa(i), "GET")
		registerNoop(r, "/items/{id}/"+strconv.Itoa(i), "GET")
	}
	registerNoop(r, "/users/{userId}", "GET")
	registerNoop(r, "/files/*", "GET")
	registerNoop(r, "/catch/**", "GET")
}

func benchRegistryMany(b *testing.B, n int) {
	r := NewRouteRegistry()
	populateRegistry(r, n)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.LoadIntoSlice("/users/42", "GET", nil)
	}
}

func BenchmarkRouteRegistryMany100(b *testing.B)   { benchRegistryMany(b, 100) }
func BenchmarkRouteRegistryMany1000(b *testing.B)  { benchRegistryMany(b, 1000) }
func BenchmarkRouteRegistryMany10000(b *testing.B) { benchRegistryMany(b, 10000) }

func BenchmarkRouteRegistryLoadParallel(b *testing.B) {
	r := NewRouteRegistry()
	populateRegistry(r, 1000)

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var params routing.Params
		for pb.Next() {
			_, _ = r.LoadIntoSlice("/users/42", "GET", &params)
		}
	})
}
