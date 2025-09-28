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

// noop handler used for registration
func noopHandler(c routing.RouteContext) {}

// helper to create a router with N routes of the form /item/{id}0..N
func createRouterWithN(n int) *Router {
	r := NewRouter()
	rg := r.NewRouteGroup("")
	for i := 0; i < n; i++ {
		// create a few distinct static routes and a param route
		rg.GET("/static/route/"+strconv.Itoa(i), noopHandler)
		rg.GET("/items/{id}/"+strconv.Itoa(i), noopHandler)
	}
	// also add some parameterized and wildcard routes
	rg.GET("/users/{userId}", noopHandler)
	rg.GET("/files/*", noopHandler)
	rg.GET("/catch/**", noopHandler)
	return r
}

func BenchmarkRouter_ExactMatch_SingleRoute(b *testing.B) {
	r := NewRouter()
	rg := r.NewRouteGroup("")
	rg.GET("/hello", noopHandler)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkRouter_ExactMatch_SingleRoute_Pool(b *testing.B) {
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")
	rg.GET("/hello", noopHandler)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkRouter_ParamMatch_SingleRoute(b *testing.B) {
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

func BenchmarkRouter_ParamMatch_SingleRoute_Pool(b *testing.B) {
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

func BenchmarkRouter_Wildcard_CatchAll(b *testing.B) {
	r := NewRouter()
	rg := r.NewRouteGroup("")
	rg.GET("/files/*", noopHandler)
	rg.GET("/catch/**", noopHandler)

	req := httptest.NewRequest(http.MethodGet, "/files/some/path/file.txt", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkRouter_Wildcard_CatchAll_Pool(b *testing.B) {
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")
	rg.GET("/files/*", noopHandler)
	rg.GET("/catch/**", noopHandler)

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

func BenchmarkRouter_ManyRoutes_100(b *testing.B)   { benchRouterManyRoutes(b, 100) }
func BenchmarkRouter_ManyRoutes_1000(b *testing.B)  { benchRouterManyRoutes(b, 1000) }
func BenchmarkRouter_ManyRoutes_10000(b *testing.B) { benchRouterManyRoutes(b, 10000) }
