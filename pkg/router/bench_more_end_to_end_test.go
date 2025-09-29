package router_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/middleware/compression"
	"github.com/fgrzl/mux/pkg/middleware/logging"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// buildDeepPath returns a long path under a base like /app/ with N segments.
func buildDeepPath(n int) string {
	segs := make([]string, n)
	for i := range segs {
		segs[i] = "seg"
	}
	return "/app/" + strings.Join(segs, "/")
}

// Benchmark deep SPA path with catch-all fallback: pooled
func BenchmarkServeHTTP_SPA_DeepPath_Pool(b *testing.B) {
	r := router.NewRouter(router.WithContextPooling())
	rg := r.NewRouteGroup("")
	rg.GET("/app/**", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})
	path := buildDeepPath(50)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// Benchmark deep SPA path with catch-all fallback: non-pooled
func BenchmarkServeHTTP_SPA_DeepPath_NonPool(b *testing.B) {
	r := router.NewRouter()
	rg := r.NewRouteGroup("")
	rg.GET("/app/**", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})
	path := buildDeepPath(50)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// SPA deep-path with logging+compression middleware: pooled
func BenchmarkServeHTTP_SPA_DeepPath_WithMiddleware_Pool(b *testing.B) {
	r := router.NewRouter(router.WithContextPooling())
	logging.UseLogging(r)
	compression.UseCompression(r)

	rg := r.NewRouteGroup("")
	rg.GET("/app/**", func(c routing.RouteContext) {
		// write some bytes to exercise compression
		c.Response().Write([]byte(strings.Repeat("x", 512)))
	})

	path := buildDeepPath(50)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

// SPA deep-path with logging+compression middleware: non-pooled
func BenchmarkServeHTTP_SPA_DeepPath_WithMiddleware_NonPool(b *testing.B) {
	r := router.NewRouter()
	logging.UseLogging(r)
	compression.UseCompression(r)

	rg := r.NewRouteGroup("")
	rg.GET("/app/**", func(c routing.RouteContext) {
		c.Response().Write([]byte(strings.Repeat("x", 512)))
	})

	path := buildDeepPath(50)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

// Benchmark HEAD->GET fallback cost on match
func BenchmarkServeHTTP_HeadFallback_Pool(b *testing.B) {
	r := router.NewRouter(router.WithContextPooling(), router.WithHeadFallbackToGet())
	rg := r.NewRouteGroup("")
	rg.GET("/files/*", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodHead, "/files/a/b/c.txt", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkServeHTTP_HeadFallback_NonPool(b *testing.B) {
	r := router.NewRouter(router.WithHeadFallbackToGet())
	rg := r.NewRouteGroup("")
	rg.GET("/files/*", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodHead, "/files/a/b/c.txt", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// Param vs wildcard vs catch-all comparison
func BenchmarkServeHTTP_ParamVsWildcardVsCatchAll(b *testing.B) {
	cases := []struct {
		name    string
		pattern string
		path    string
	}{
		{"param", "/users/{id}", "/users/12345"},
		{"wildcard", "/files/*", "/files/a/b/c.txt"},
		{"catchall", "/app/**", buildDeepPath(10)},
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
				req := httptest.NewRequest(http.MethodGet, tc.path, nil)
				rr := httptest.NewRecorder()

				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					r.ServeHTTP(rr, req)
				}
			})
		}
	}
}
