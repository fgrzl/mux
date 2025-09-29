package router

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
)

// BenchServeHTTP_E2E_Pool measures a full Router.ServeHTTP call for a router with context pooling enabled.
func BenchmarkServeHTTPE2EPool(b *testing.B) {
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")
	// register some routes to build the tree
	for i := 0; i < 100; i++ {
		rg.GET("/static/route/"+strconv.Itoa(i), func(c routing.RouteContext) {})
		rg.GET("/items/{id}/"+strconv.Itoa(i), func(c routing.RouteContext) {})
	}
	rg.GET("/users/{userId}", func(c routing.RouteContext) {})
	rg.GET("/files/*", func(c routing.RouteContext) {})
	rg.GET("/catch/**", func(c routing.RouteContext) {})

	req := httptest.NewRequest(http.MethodGet, "/files/some/path/file.txt", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// BenchServeHTTP_E2E_NonPool measures a full Router.ServeHTTP call for a router without pooling.
func BenchmarkServeHTTPE2ENonPool(b *testing.B) {
	r := NewRouter()
	rg := r.NewRouteGroup("")
	// register some routes to build the tree
	for i := 0; i < 100; i++ {
		rg.GET("/static/route/"+strconv.Itoa(i), func(c routing.RouteContext) {})
		rg.GET("/items/{id}/"+strconv.Itoa(i), func(c routing.RouteContext) {})
	}
	rg.GET("/users/{userId}", func(c routing.RouteContext) {})
	rg.GET("/files/*", func(c routing.RouteContext) {})
	rg.GET("/catch/**", func(c routing.RouteContext) {})

	req := httptest.NewRequest(http.MethodGet, "/files/some/path/file.txt", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}
