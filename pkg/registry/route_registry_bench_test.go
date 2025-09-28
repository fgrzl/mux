package registry

import (
	"io"
	"log/slog"
	"strconv"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
)

func init() {
	// silence slog during benchmarks to avoid noisy output
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
}

// helper to register a noop RouteOptions into registry
func registerNoop(r *RouteRegistry, pattern string, method string) {
	opts := &routing.RouteOptions{Method: method, Pattern: pattern, Handler: func(c routing.RouteContext) {}}
	r.Register(pattern, method, opts)
}

func BenchmarkRouteRegistry_Load_ExactMatch(b *testing.B) {
	r := NewRouteRegistry()
	registerNoop(r, "/hello", "GET")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = r.Load("/hello", "GET")
	}
}

func BenchmarkRouteRegistry_Load_ParamMatch(b *testing.B) {
	r := NewRouteRegistry()
	registerNoop(r, "/users/{id}", "GET")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, params, _ := r.Load("/users/12345", "GET")
		_ = params
	}
}

func BenchmarkRouteRegistry_Load_Wildcard(b *testing.B) {
	r := NewRouteRegistry()
	registerNoop(r, "/files/*", "GET")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = r.Load("/files/path/to/file.txt", "GET")
	}
}

func BenchmarkRouteRegistry_Load_CatchAll(b *testing.B) {
	r := NewRouteRegistry()
	registerNoop(r, "/catch/**", "GET")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = r.Load("/catch/any/number/of/segments/here", "GET")
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
		_, _, _ = r.Load("/users/42", "GET")
	}
}

func BenchmarkRouteRegistry_Many_100(b *testing.B)   { benchRegistryMany(b, 100) }
func BenchmarkRouteRegistry_Many_1000(b *testing.B)  { benchRegistryMany(b, 1000) }
func BenchmarkRouteRegistry_Many_10000(b *testing.B) { benchRegistryMany(b, 10000) }

func BenchmarkRouteRegistry_Load_Parallel(b *testing.B) {
	r := NewRouteRegistry()
	populateRegistry(r, 1000)

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, _ = r.Load("/users/42", "GET")
		}
	})
}
