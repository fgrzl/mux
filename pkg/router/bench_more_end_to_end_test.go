package router_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/bench"
	"github.com/fgrzl/mux/pkg/middleware/compression"
	"github.com/fgrzl/mux/pkg/middleware/logging"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// Benchmark deep SPA path with catch-all fallback: pooled
func BenchmarkServeHTTPSPADeepPathPool(b *testing.B) {
	r := router.NewRouter(router.WithContextPooling())
	rg := r.NewRouteGroup("")
	rg.GET("/app/**", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})
	path := bench.BuildDeepPath(50)
	_, req := bench.NewRecorderRequest(http.MethodGet, path)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

// Benchmark deep SPA path with catch-all fallback: non-pooled
func BenchmarkServeHTTPSPADeepPathNonPool(b *testing.B) {
	r := router.NewRouter()
	rg := r.NewRouteGroup("")
	rg.GET("/app/**", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})
	path := bench.BuildDeepPath(50)
	_, req := bench.NewRecorderRequest(http.MethodGet, path)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

// SPA deep-path with logging+compression middleware: pooled
func BenchmarkServeHTTPSPADeepPathWithMiddlewarePool(b *testing.B) {
	r := router.NewRouter(router.WithContextPooling())
	logging.UseLogging(r)
	compression.UseCompression(r)

	rg := r.NewRouteGroup("")
	rg.GET("/app/**", func(c routing.RouteContext) {
		// write some bytes to exercise compression
		c.Response().Write([]byte(strings.Repeat("x", 512)))
	})

	_, req := bench.NewRecorderRequest(http.MethodGet, bench.BuildDeepPath(50))
	req.Header.Set("Accept-Encoding", "gzip")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

// SPA deep-path with logging+compression middleware: non-pooled
func BenchmarkServeHTTPSPADeepPathWithMiddlewareNonPool(b *testing.B) {
	r := router.NewRouter()
	logging.UseLogging(r)
	compression.UseCompression(r)

	rg := r.NewRouteGroup("")
	rg.GET("/app/**", func(c routing.RouteContext) {
		c.Response().Write([]byte(strings.Repeat("x", 512)))
	})

	_, req := bench.NewRecorderRequest(http.MethodGet, bench.BuildDeepPath(50))
	req.Header.Set("Accept-Encoding", "gzip")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

// Benchmark HEAD->GET fallback cost on match
func BenchmarkServeHTTPHeadFallbackPool(b *testing.B) {
	r := router.NewRouter(router.WithContextPooling(), router.WithHeadFallbackToGet())
	rg := r.NewRouteGroup("")
	rg.GET("/files/*", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})
	_, req := bench.NewRecorderRequest(http.MethodHead, "/files/a/b/c.txt")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkServeHTTPHeadFallbackNonPool(b *testing.B) {
	r := router.NewRouter(router.WithHeadFallbackToGet())
	rg := r.NewRouteGroup("")
	rg.GET("/files/*", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})
	_, req := bench.NewRecorderRequest(http.MethodHead, "/files/a/b/c.txt")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

// Param vs wildcard vs catch-all comparison
func BenchmarkServeHTTPParamVsWildcardVsCatchAll(b *testing.B) {
	cases := []struct {
		name    string
		pattern string
		path    string
	}{
		{"param", "/users/{id}", "/users/12345"},
		{"wildcard", "/files/*", "/files/a/b/c.txt"},
		{"catchall", "/app/**", bench.BuildDeepPath(10)},
	}
	for _, pooled := range []bool{false, true} {
		for _, tc := range cases {
			b.Run(tc.name+func() string {
				if pooled {
					return "_pool"
				}
				return "_nonpool"
			}(), func(b *testing.B) {
				var r *router.Router
				if pooled {
					r = router.NewRouter(router.WithContextPooling())
				} else {
					r = router.NewRouter()
				}
				rg := r.NewRouteGroup("")
				rg.GET(tc.pattern, func(c routing.RouteContext) { c.Response().WriteHeader(http.StatusOK) })
				rr, req := bench.NewRecorderRequest(http.MethodGet, tc.path)

				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					r.ServeHTTP(rr, req)
				}
			})
		}
	}
}
