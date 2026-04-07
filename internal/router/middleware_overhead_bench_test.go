package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/internal/routing"
)

// This file isolates the cost of Mux middleware traversal so we can track
// the framework's own pipeline overhead over time.

// BenchmarkMuxNoMiddleware tests your mux without any middleware
func BenchmarkMuxNoMiddleware(b *testing.B) {
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")
	rg.GET("/users/{id}", func(c routing.RouteContext) {})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

// simpleMiddleware is a test middleware that just passes through
type simpleMiddleware struct{}

func (m *simpleMiddleware) Invoke(c routing.RouteContext, next HandlerFunc) {
	next(c) // Just pass through
}

// BenchmarkMuxWithMiddleware tests Mux with a simple pass-through middleware.
func BenchmarkMuxWithMiddleware(b *testing.B) {
	r := NewRouter(WithContextPooling())

	// Add a simple middleware (similar to what Gin's Recovery does)
	r.Use(&simpleMiddleware{})

	rg := r.NewRouteGroup("")
	rg.GET("/users/{id}", func(c routing.RouteContext) {})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}
