package router

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
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

// Old pooled behavior: acquire context (pooled) then perform LoadDetailedInto then Load
func BenchmarkRouterE2EOldPool(b *testing.B) {
	r := setupComplexRouter()
	req := httptest.NewRequest(http.MethodGet, "/files/some/path/file.txt", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := routing.AcquireContext(rr, req)
		// Old map-based params path removed - using slice-based params now
		_ = c.ParamsSlice()
		_, _, _ = r.routeRegistry.Load(req.URL.Path, req.Method)
		routing.ReleaseContext(c)
	}
}

// New pooled behavior: acquire context then perform single FindNodeInto
func BenchmarkRouterE2ENewPool(b *testing.B) {
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

// Old non-pooled behavior: NewRouteContext, Load (fast) then LoadDetailedInto when needed
func BenchmarkRouterE2EOldNonPool(b *testing.B) {
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
		// non-pooled: emulate previous flow where Load then LoadDetailedInto may be used
		_, pmap, ok := r.routeRegistry.Load(req.URL.Path, req.Method)
		if ok {
			_ = pmap
		} else {
			tmp := make(routing.RouteParams, 2)
			_, _ = r.routeRegistry.LoadDetailedInto(req.URL.Path, req.Method, tmp)
		}
		_ = c
	}
}

// New non-pooled behavior: use FindNode (non-allocating) and only allocate when attaching params
func BenchmarkRouterE2ENewNonPool(b *testing.B) {
	r := NewRouter()
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
				// allocate only when needed
				tmp := make(routing.RouteParams, 2)
				_, _ = r.routeRegistry.LoadInto(req.URL.Path, http.MethodGet, tmp)
			}
		}
		_ = c
	}
}
