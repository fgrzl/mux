package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	gorillamux "github.com/gorilla/mux"
	"github.com/julienschmidt/httprouter"
	"github.com/labstack/echo/v4"
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

// --- Gin Implementation ---
// Note: Gin has stricter parameter naming - can't have :id and :userId in same tree

func setupGinRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Benchmark handlers are intentionally empty to measure only routing performance
	emptyHandler := func(c *gin.Context) { /* empty by design for benchmarking */ }

	r.GET("/", emptyHandler)
	r.GET("/ping", emptyHandler)
	r.GET("/api/v1/health", emptyHandler)
	r.GET("/users/:id", emptyHandler)
	r.GET("/items/:id/posts", emptyHandler)         // Adjusted to avoid Gin param name conflicts
	r.GET("/content/:userId/:postId", emptyHandler) // Adjusted for Gin's constraints
	r.GET("/api/v1/organizations/:orgId/projects/:projectId", emptyHandler)
	r.GET("/files/*filepath", emptyHandler)
	r.GET("/static/*filepath", emptyHandler)
	r.POST("/api/v1/users", emptyHandler)
	r.PUT("/api/v1/users/:id", emptyHandler)
	r.DELETE("/api/v1/users/:id", emptyHandler)

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
	req := httptest.NewRequest(http.MethodGet, "/content/12345/67890", nil) // Adjusted for Gin route
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGinDeepPath(b *testing.B) {
	r := setupGinRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org-123/projects/proj-456", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGinWildcard(b *testing.B) {
	r := setupGinRouter()
	req := httptest.NewRequest(http.MethodGet, "/files/images/logo.png", nil)
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

	// Benchmark handlers are intentionally empty to measure only routing performance
	emptyHandler := func(w http.ResponseWriter, r *http.Request) { /* empty by design for benchmarking */ }

	r.Get("/", emptyHandler)
	r.Get("/ping", emptyHandler)
	r.Get("/api/v1/health", emptyHandler)
	r.Get("/users/{id}", emptyHandler)
	r.Get("/users/{id}/posts", emptyHandler)
	r.Get("/users/{userId}/posts/{postId}", emptyHandler)
	r.Get("/api/v1/organizations/{orgId}/projects/{projectId}", emptyHandler)
	r.Get("/files/*", emptyHandler)
	r.Get("/static/*", emptyHandler)
	r.Post("/api/v1/users", emptyHandler)
	r.Put("/api/v1/users/{id}", emptyHandler)
	r.Delete("/api/v1/users/{id}", emptyHandler)

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

func BenchmarkComparisonChiDeepPath(b *testing.B) {
	r := setupChiRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org-123/projects/proj-456", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonChiWildcard(b *testing.B) {
	r := setupChiRouter()
	req := httptest.NewRequest(http.MethodGet, "/files/images/logo.png", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// --- Echo Implementation ---

func setupEchoRouter() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Benchmark handlers are intentionally empty to measure only routing performance
	emptyHandler := func(c echo.Context) error { return nil /* empty by design for benchmarking */ }

	e.GET("/", emptyHandler)
	e.GET("/ping", emptyHandler)
	e.GET("/api/v1/health", emptyHandler)
	e.GET("/users/:id", emptyHandler)
	e.GET("/users/:id/posts", emptyHandler)
	e.GET("/users/:userId/posts/:postId", emptyHandler)
	e.GET("/api/v1/organizations/:orgId/projects/:projectId", emptyHandler)
	e.GET("/files/*", emptyHandler)
	e.GET("/static/*", emptyHandler)
	e.POST("/api/v1/users", emptyHandler)
	e.PUT("/api/v1/users/:id", emptyHandler)
	e.DELETE("/api/v1/users/:id", emptyHandler)

	return e
}

func BenchmarkComparisonEchoStaticRoute(b *testing.B) {
	e := setupEchoRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonEchoSingleParam(b *testing.B) {
	e := setupEchoRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonEchoMultiParam(b *testing.B) {
	e := setupEchoRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345/posts/67890", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonEchoDeepPath(b *testing.B) {
	e := setupEchoRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org-123/projects/proj-456", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonEchoWildcard(b *testing.B) {
	e := setupEchoRouter()
	req := httptest.NewRequest(http.MethodGet, "/files/images/logo.png", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rr, req)
	}
}

// --- Gorilla Mux Implementation ---

func setupGorillaMuxRouter() *gorillamux.Router {
	r := gorillamux.NewRouter()

	// Benchmark handlers are intentionally empty to measure only routing performance
	emptyHandler := func(w http.ResponseWriter, r *http.Request) { /* empty by design for benchmarking */ }

	r.HandleFunc("/", emptyHandler).Methods(http.MethodGet)
	r.HandleFunc("/ping", emptyHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/health", emptyHandler).Methods(http.MethodGet)
	r.HandleFunc("/users/{id}", emptyHandler).Methods(http.MethodGet)
	r.HandleFunc("/users/{id}/posts", emptyHandler).Methods(http.MethodGet)
	r.HandleFunc("/users/{userId}/posts/{postId}", emptyHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/organizations/{orgId}/projects/{projectId}", emptyHandler).Methods(http.MethodGet)
	r.PathPrefix("/files/").HandlerFunc(emptyHandler).Methods(http.MethodGet)
	r.PathPrefix("/static/").HandlerFunc(emptyHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/users", emptyHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/users/{id}", emptyHandler).Methods(http.MethodPut)
	r.HandleFunc("/api/v1/users/{id}", emptyHandler).Methods(http.MethodDelete)

	return r
}

func BenchmarkComparisonGorillaMuxStaticRoute(b *testing.B) {
	r := setupGorillaMuxRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGorillaMuxSingleParam(b *testing.B) {
	r := setupGorillaMuxRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGorillaMuxMultiParam(b *testing.B) {
	r := setupGorillaMuxRouter()
	req := httptest.NewRequest(http.MethodGet, "/users/12345/posts/67890", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGorillaMuxDeepPath(b *testing.B) {
	r := setupGorillaMuxRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org-123/projects/proj-456", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGorillaMuxWildcard(b *testing.B) {
	r := setupGorillaMuxRouter()
	req := httptest.NewRequest(http.MethodGet, "/files/images/logo.png", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// --- HttpRouter Implementation ---
// Note: HttpRouter also has strict parameter naming like Gin

func setupHttpRouter() *httprouter.Router {
	r := httprouter.New()

	// Benchmark handlers are intentionally empty to measure only routing performance
	emptyHandler := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) { /* empty by design for benchmarking */
	}

	r.GET("/", emptyHandler)
	r.GET("/ping", emptyHandler)
	r.GET("/api/v1/health", emptyHandler)
	r.GET("/users/:id", emptyHandler)
	r.GET("/items/:id/posts", emptyHandler)         // Adjusted to avoid HttpRouter param conflicts
	r.GET("/content/:userId/:postId", emptyHandler) // Adjusted for HttpRouter's constraints
	r.GET("/api/v1/organizations/:orgId/projects/:projectId", emptyHandler)
	r.GET("/files/*filepath", emptyHandler)
	r.GET("/static/*filepath", emptyHandler)
	r.POST("/api/v1/users", emptyHandler)
	r.PUT("/api/v1/users/:id", emptyHandler)
	r.DELETE("/api/v1/users/:id", emptyHandler)

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
	req := httptest.NewRequest(http.MethodGet, "/content/12345/67890", nil) // Adjusted for HttpRouter route
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonHttpRouterDeepPath(b *testing.B) {
	r := setupHttpRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org-123/projects/proj-456", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonHttpRouterWildcard(b *testing.B) {
	r := setupHttpRouter()
	req := httptest.NewRequest(http.MethodGet, "/files/images/logo.png", nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}
