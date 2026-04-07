package router

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fgrzl/mux/internal/routing"
)

// setup helper: returns a router with a few routes including wildcard and catch-all
func setupComplexRouter() *Router {
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")
	// add many routes to create a non-trivial tree
	for i := 0; i < 50; i++ {
		rg.GET("/static/route/"+strconv.Itoa(i), func(c routing.RouteContext) {})
		rg.GET("/items/{id}/"+strconv.Itoa(i), func(c routing.RouteContext) {})
	}
	rg.GET("/users/{userId}", func(c routing.RouteContext) {})
	rg.GET("/files/*", func(c routing.RouteContext) {})
	rg.GET("/catch/**", func(c routing.RouteContext) {})
	return r
}

func BenchmarkRouterE2EPool(b *testing.B) {
	r := setupComplexRouter()
	req := httptest.NewRequest(http.MethodGet, "/files/some/path/file.txt", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := routing.AcquireContext(rr, req)
		_ = c.ParamsSlice()
		_ = r.routeRegistry.FindNode(req.URL.Path)
		routing.ReleaseContext(c)
	}
}

func BenchmarkRouterE2ENonPool(b *testing.B) {
	r := NewRouter() // non-pooled
	rg := r.NewRouteGroup("")
	rg.GET("/users/{userId}", func(c routing.RouteContext) {})
	rg.GET("/files/*", func(c routing.RouteContext) {})
	rg.GET("/catch/**", func(c routing.RouteContext) {})

	req := httptest.NewRequest(http.MethodGet, "/files/some/path/file.txt", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := routing.NewRouteContext(rr, req)
		node := r.routeRegistry.FindNode(req.URL.Path)
		if node != nil {
			if _, ok := node.RouteOptions[http.MethodGet]; ok {
				var params routing.Params
				_, _ = r.routeRegistry.LoadIntoSlice(req.URL.Path, http.MethodGet, &params)
			}
		}
		_ = c
	}
}
