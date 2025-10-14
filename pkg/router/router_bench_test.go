package router

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
)

func init() {
	// silence slog during benchmarks to avoid noisy output
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
}

// Benchmark route patterns as constants to avoid duplication
const (
	benchmarkHelloPath  = "/hello"
	benchmarkFilesPath  = "/files/*"
	benchmarkCatchPath  = "/catch/**"
	benchmarkUsersParam = "/users/{userId}"
)

// noopHandler is intentionally empty to measure only routing performance,
// not handler execution time. This isolates router overhead for accurate benchmarking.
func noopHandler(c routing.RouteContext) { /* empty by design for benchmarking */ }

// createRouterWithN is a helper to create a router with N routes of the form /item/{id}0..N
func createRouterWithN(n int) *Router {
	r := NewRouter()
	rg := r.NewRouteGroup("")
	for i := 0; i < n; i++ {
		// create a few distinct static routes and a param route
		rg.GET("/static/route/"+strconv.Itoa(i), noopHandler)
		rg.GET("/items/{id}/"+strconv.Itoa(i), noopHandler)
	}
	// also add some parameterized and wildcard routes
	rg.GET(benchmarkUsersParam, noopHandler)
	rg.GET(benchmarkFilesPath, noopHandler)
	rg.GET(benchmarkCatchPath, noopHandler)
	return r
}

func BenchmarkRouterExactMatchSingleRoute(b *testing.B) {
	r := NewRouter()
	rg := r.NewRouteGroup("")
	rg.GET(benchmarkHelloPath, noopHandler)

	req := httptest.NewRequest(http.MethodGet, benchmarkHelloPath, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkRouterExactMatchSingleRoutePool(b *testing.B) {
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")
	rg.GET(benchmarkHelloPath, noopHandler)

	req := httptest.NewRequest(http.MethodGet, benchmarkHelloPath, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkRouterParamMatchSingleRoute(b *testing.B) {
	r := NewRouter()
	rg := r.NewRouteGroup("")
	rg.GET("/users/{id}", noopHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/12345", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkRouterParamMatchSingleRoutePool(b *testing.B) {
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")
	rg.GET("/users/{id}", noopHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/12345", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkRouterWildcardCatchAll(b *testing.B) {
	r := NewRouter()
	rg := r.NewRouteGroup("")
	rg.GET(benchmarkFilesPath, noopHandler)
	rg.GET(benchmarkCatchPath, noopHandler)

	req := httptest.NewRequest(http.MethodGet, "/files/some/path/file.txt", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkRouterWildcardCatchAllPool(b *testing.B) {
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")
	rg.GET(benchmarkFilesPath, noopHandler)
	rg.GET(benchmarkCatchPath, noopHandler)

	req := httptest.NewRequest(http.MethodGet, "/files/some/path/file.txt", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func benchRouterManyRoutes(b *testing.B, routeCount int) {
	r := createRouterWithN(routeCount)

	// pick a path that will exercise param matching and deeper trees
	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkRouterManyRoutes100(b *testing.B)   { benchRouterManyRoutes(b, 100) }
func BenchmarkRouterManyRoutes1000(b *testing.B)  { benchRouterManyRoutes(b, 1000) }
func BenchmarkRouterManyRoutes10000(b *testing.B) { benchRouterManyRoutes(b, 10000) }
