package router

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/internal/bench"
	"github.com/fgrzl/mux/internal/routing"
)

// BenchServeHTTP_E2E_Pool measures a full Router.ServeHTTP call for a router with context pooling enabled.
func BenchmarkServeHTTPE2EPool(b *testing.B) {
	patterns := bench.BuildManyRoutePatterns(100)
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")
	for _, p := range patterns {
		pat := p
		rg.GET(pat, func(c routing.RouteContext) {})
	}
	rr, req := bench.NewRecorderRequest(http.MethodGet, "/files/some/path/file.txt")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// BenchServeHTTP_E2E_NonPool measures a full Router.ServeHTTP call for a router without pooling.
func BenchmarkServeHTTPE2ENonPool(b *testing.B) {
	patterns := bench.BuildManyRoutePatterns(100)
	r := NewRouter()
	rg := r.NewRouteGroup("")
	for _, p := range patterns {
		pat := p
		rg.GET(pat, func(c routing.RouteContext) {})
	}
	rr, req := bench.NewRecorderRequest(http.MethodGet, "/files/some/path/file.txt")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}
