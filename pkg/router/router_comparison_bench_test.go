package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
)

// This file benchmarks mux against other popular Go routers to establish competitive positioning.
// To run these benchmarks, you'll need to install the comparison routers:
//   go get github.com/gin-gonic/gin
//   go get github.com/go-chi/chi/v5
//   go get github.com/labstack/echo/v4
//   go get github.com/gorilla/mux
//   go get github.com/julienschmidt/httprouter
//
// Run with: go test -bench=BenchmarkComparison -benchmem -benchtime=2s

// Test scenarios matching common real-world use cases

// --- MUX Implementation ---

func setupMuxRouter() *Router {
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")

	// Benchmark handlers are intentionally empty to measure only routing performance,
	// not handler execution time. This isolates router overhead for accurate comparison.
	noopHandler := func(c routing.RouteContext) { /* empty by design for benchmarking */ }

	// Static routes
	rg.GET("/", noopHandler)
	rg.GET("/ping", noopHandler)
	rg.GET("/api/v1/health", noopHandler)

	// Parameter routes
	rg.GET("/users/{id}", noopHandler)
	rg.GET("/users/{id}/posts", noopHandler)
	rg.GET("/users/{userId}/posts/{postId}", noopHandler)
	rg.GET("/api/v1/organizations/{orgId}/projects/{projectId}", noopHandler)

	// Wildcard/catch-all
	rg.GET("/files/*", noopHandler)
	rg.GET("/static/**", noopHandler)

	// Multiple methods
	rg.POST("/api/v1/users", noopHandler)
	rg.PUT("/api/v1/users/{id}", noopHandler)
	rg.DELETE("/api/v1/users/{id}", noopHandler)

	return r
}

// --- Benchmark Scenarios ---

// Scenario 1: Static route (most common case)
func BenchmarkComparisonMuxStaticRoute(b *testing.B) {
	r := setupMuxRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// Scenario 2: Single parameter route
func BenchmarkComparisonMuxSingleParam(b *testing.B) {
	r := setupMuxRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// Scenario 3: Multi-parameter route
func BenchmarkComparisonMuxMultiParam(b *testing.B) {
	r := setupMuxRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345/posts/67890", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// Scenario 4: Deep path with parameters
func BenchmarkComparisonMuxDeepPath(b *testing.B) {
	r := setupMuxRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org-123/projects/proj-456", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// Scenario 5: Wildcard/catch-all route
func BenchmarkComparisonMuxWildcard(b *testing.B) {
	r := setupMuxRouter()
	req := httptest.NewRequest(http.MethodGet, "/files/images/logo.png", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// Scenario 6: Many routes (scalability test)
func BenchmarkComparisonMuxManyRoutes(b *testing.B) {
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")

	// Benchmark handler is intentionally empty to measure routing performance only
	noopHandler := func(c routing.RouteContext) { /* empty by design for benchmarking */ }

	// Register 1000 routes
	for i := 0; i < 100; i++ {
		rg.GET("/route"+string(rune(i)), noopHandler)
		rg.GET("/api/resource"+string(rune(i))+"/{id}", noopHandler)
		rg.GET("/v1/items"+string(rune(i))+"/{itemId}/sub/{subId}", noopHandler)
	}

	// Match a route in the middle
	req := httptest.NewRequest(http.MethodGet, "/api/resource50/12345", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// NOTE: Uncomment the sections below after installing the comparison routers
/*
import (
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/labstack/echo/v4"
	"github.com/gorilla/mux" // aliased as gorillamux
	"github.com/julienschmidt/httprouter"
)

// --- Gin Implementation ---

func setupGinRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.GET("/", func(c *gin.Context) {})
	r.GET("/ping", func(c *gin.Context) {})
	r.GET("/api/v1/health", func(c *gin.Context) {})
	r.GET("/users/:id", func(c *gin.Context) {})
	r.GET("/users/:id/posts", func(c *gin.Context) {})
	r.GET("/users/:userId/posts/:postId", func(c *gin.Context) {})
	r.GET("/api/v1/organizations/:orgId/projects/:projectId", func(c *gin.Context) {})
	r.GET("/files/*filepath", func(c *gin.Context) {})
	r.POST("/api/v1/users", func(c *gin.Context) {})
	r.PUT("/api/v1/users/:id", func(c *gin.Context) {})
	r.DELETE("/api/v1/users/:id", func(c *gin.Context) {})

	return r
}

func BenchmarkComparisonGinStaticRoute(b *testing.B) {
	r := setupGinRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGinSingleParam(b *testing.B) {
	r := setupGinRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGinMultiParam(b *testing.B) {
	r := setupGinRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345/posts/67890", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// --- Chi Implementation ---

func setupChiRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/users/{id}/posts", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/users/{userId}/posts/{postId}", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/api/v1/organizations/{orgId}/projects/{projectId}", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/files/*", func(w http.ResponseWriter, r *http.Request) {})
	r.Post("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {})
	r.Put("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {})
	r.Delete("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {})

	return r
}

func BenchmarkComparisonChiStaticRoute(b *testing.B) {
	r := setupChiRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonChiSingleParam(b *testing.B) {
	r := setupChiRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonChiMultiParam(b *testing.B) {
	r := setupChiRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345/posts/67890", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// --- HttpRouter Implementation ---

func setupHttpRouter() *httprouter.Router {
	r := httprouter.New()

	r.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	r.GET("/ping", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	r.GET("/api/v1/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	r.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	r.GET("/users/:id/posts", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	r.GET("/users/:userId/posts/:postId", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	r.GET("/api/v1/organizations/:orgId/projects/:projectId", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	r.GET("/files/*filepath", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	r.POST("/api/v1/users", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	r.PUT("/api/v1/users/:id", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	r.DELETE("/api/v1/users/:id", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})

	return r
}

func BenchmarkComparisonHttpRouterStaticRoute(b *testing.B) {
	r := setupHttpRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonHttpRouterSingleParam(b *testing.B) {
	r := setupHttpRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonHttpRouterMultiParam(b *testing.B) {
	r := setupHttpRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345/posts/67890", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}
*/
