package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/internal/routing"
	"github.com/gin-gonic/gin"
)

// This file benchmarks middleware overhead to understand the true cost
// of middleware in different routers

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

// BenchmarkMuxWithMiddleware tests your mux with a simple middleware
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

// BenchmarkGinNoMiddleware tests Gin without middleware
func BenchmarkGinNoMiddleware(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New() // No middleware
	r.GET("/users/:id", func(c *gin.Context) {})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

// BenchmarkGinWithMiddleware tests Gin with middleware
func BenchmarkGinWithMiddleware(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Add a simple middleware
	r.Use(func(c *gin.Context) {
		c.Next() // Just pass through
	})

	r.GET("/users/:id", func(c *gin.Context) {})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

// BenchmarkGinDefault tests Gin with default middleware (Logger + Recovery)
func BenchmarkGinDefault(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	// Disable default logging to avoid console spam
	gin.DefaultWriter = io.Discard
	r := gin.Default() // Includes Logger and Recovery middleware
	r.GET("/users/:id", func(c *gin.Context) {})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}
