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

// Benchmark route paths
const (
	benchPathAPIV1Health      = "/api/v1/health"
	benchPathUsersID          = "/users/{id}"
	benchPathUsersIDPosts     = "/users/{id}/posts"
	benchPathUsersPostsNested = "/users/{userId}/posts/{postId}"
	benchPathOrgProjects      = "/api/v1/organizations/{orgId}/projects/{projectId}"
	benchPathFiles            = "/files/*"
	benchPathAPIV1Users       = "/api/v1/users"
	benchPathAPIV1UsersID     = "/api/v1/users/{id}"
)

// Benchmark request paths
const (
	benchReqUsers12345         = "/users/12345"
	benchReqUsersPosts         = "/users/12345/posts/67890"
	benchReqOrgProjects        = "/api/v1/organizations/org-123/projects/proj-456"
	benchReqFilesLogoPath      = "/files/images/logo.png"
	benchReqUsersColonID       = "/users/:id"
	benchReqOrgProjectsColonID = "/api/v1/organizations/:orgId/projects/:projectId"
	benchReqUsersIDColon       = "/api/v1/users/:id"
)

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
	rg.GET(benchPathAPIV1Health, noopHandler)

	// Parameter routes
	rg.GET(benchPathUsersID, noopHandler)
	rg.GET(benchPathUsersIDPosts, noopHandler)
	rg.GET(benchPathUsersPostsNested, noopHandler)
	rg.GET(benchPathOrgProjects, noopHandler)

	// Wildcard/catch-all
	rg.GET(benchPathFiles, noopHandler)
	rg.GET("/static/**", noopHandler)

	// Multiple methods
	rg.POST(benchPathAPIV1Users, noopHandler)
	rg.PUT(benchPathAPIV1UsersID, noopHandler)
	rg.DELETE(benchPathAPIV1UsersID, noopHandler)

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
	req := httptest.NewRequest(http.MethodGet, benchReqUsers12345, nil)
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
	req := httptest.NewRequest(http.MethodGet, benchReqUsersPosts, nil)
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
	req := httptest.NewRequest(http.MethodGet, benchReqOrgProjects, nil)
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
	req := httptest.NewRequest(http.MethodGet, benchReqFilesLogoPath, nil)
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
	r.GET(benchPathAPIV1Health, emptyHandler)
	r.GET(benchReqUsersColonID, emptyHandler)
	r.GET("/items/:id/posts", emptyHandler)         // Adjusted to avoid Gin param name conflicts
	r.GET("/content/:userId/:postId", emptyHandler) // Adjusted for Gin's constraints
	r.GET(benchReqOrgProjectsColonID, emptyHandler)
	r.GET("/files/*filepath", emptyHandler)
	r.GET("/static/*filepath", emptyHandler)
	r.POST(benchPathAPIV1Users, emptyHandler)
	r.PUT(benchReqUsersIDColon, emptyHandler)
	r.DELETE(benchReqUsersIDColon, emptyHandler)

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
	req := httptest.NewRequest(http.MethodGet, benchReqUsers12345, nil)
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
	req := httptest.NewRequest(http.MethodGet, benchReqOrgProjects, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGinWildcard(b *testing.B) {
	r := setupGinRouter()
	req := httptest.NewRequest(http.MethodGet, benchReqFilesLogoPath, nil)
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
	r.Get(benchPathAPIV1Health, emptyHandler)
	r.Get(benchPathUsersID, emptyHandler)
	r.Get(benchPathUsersIDPosts, emptyHandler)
	r.Get(benchPathUsersPostsNested, emptyHandler)
	r.Get(benchPathOrgProjects, emptyHandler)
	r.Get(benchPathFiles, emptyHandler)
	r.Get("/static/*", emptyHandler)
	r.Post(benchPathAPIV1Users, emptyHandler)
	r.Put(benchPathAPIV1UsersID, emptyHandler)
	r.Delete(benchPathAPIV1UsersID, emptyHandler)

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
	req := httptest.NewRequest(http.MethodGet, benchReqUsers12345, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonChiMultiParam(b *testing.B) {
	r := setupChiRouter()
	req := httptest.NewRequest(http.MethodGet, benchReqUsersPosts, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonChiDeepPath(b *testing.B) {
	r := setupChiRouter()
	req := httptest.NewRequest(http.MethodGet, benchReqOrgProjects, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonChiWildcard(b *testing.B) {
	r := setupChiRouter()
	req := httptest.NewRequest(http.MethodGet, benchReqFilesLogoPath, nil)
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
	e.GET(benchPathAPIV1Health, emptyHandler)
	e.GET(benchReqUsersColonID, emptyHandler)
	e.GET("/users/:id/posts", emptyHandler)
	e.GET("/users/:userId/posts/:postId", emptyHandler)
	e.GET(benchReqOrgProjectsColonID, emptyHandler)
	e.GET(benchPathFiles, emptyHandler)
	e.GET("/static/*", emptyHandler)
	e.POST(benchPathAPIV1Users, emptyHandler)
	e.PUT(benchReqUsersIDColon, emptyHandler)
	e.DELETE(benchReqUsersIDColon, emptyHandler)

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
	req := httptest.NewRequest(http.MethodGet, benchReqUsers12345, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonEchoMultiParam(b *testing.B) {
	e := setupEchoRouter()
	req := httptest.NewRequest(http.MethodGet, benchReqUsersPosts, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonEchoDeepPath(b *testing.B) {
	e := setupEchoRouter()
	req := httptest.NewRequest(http.MethodGet, benchReqOrgProjects, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonEchoWildcard(b *testing.B) {
	e := setupEchoRouter()
	req := httptest.NewRequest(http.MethodGet, benchReqFilesLogoPath, nil)
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
	r.HandleFunc(benchPathAPIV1Health, emptyHandler).Methods(http.MethodGet)
	r.HandleFunc(benchPathUsersID, emptyHandler).Methods(http.MethodGet)
	r.HandleFunc(benchPathUsersIDPosts, emptyHandler).Methods(http.MethodGet)
	r.HandleFunc(benchPathUsersPostsNested, emptyHandler).Methods(http.MethodGet)
	r.HandleFunc(benchPathOrgProjects, emptyHandler).Methods(http.MethodGet)
	r.PathPrefix("/files/").HandlerFunc(emptyHandler).Methods(http.MethodGet)
	r.PathPrefix("/static/").HandlerFunc(emptyHandler).Methods(http.MethodGet)
	r.HandleFunc(benchPathAPIV1Users, emptyHandler).Methods(http.MethodPost)
	r.HandleFunc(benchPathAPIV1UsersID, emptyHandler).Methods(http.MethodPut)
	r.HandleFunc(benchPathAPIV1UsersID, emptyHandler).Methods(http.MethodDelete)

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
	req := httptest.NewRequest(http.MethodGet, benchReqUsers12345, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGorillaMuxMultiParam(b *testing.B) {
	r := setupGorillaMuxRouter()
	req := httptest.NewRequest(http.MethodGet, benchReqUsersPosts, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGorillaMuxDeepPath(b *testing.B) {
	r := setupGorillaMuxRouter()
	req := httptest.NewRequest(http.MethodGet, benchReqOrgProjects, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonGorillaMuxWildcard(b *testing.B) {
	r := setupGorillaMuxRouter()
	req := httptest.NewRequest(http.MethodGet, benchReqFilesLogoPath, nil)
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
	r.GET(benchPathAPIV1Health, emptyHandler)
	r.GET(benchReqUsersColonID, emptyHandler)
	r.GET("/items/:id/posts", emptyHandler)         // Adjusted to avoid HttpRouter param conflicts
	r.GET("/content/:userId/:postId", emptyHandler) // Adjusted for HttpRouter's constraints
	r.GET(benchReqOrgProjectsColonID, emptyHandler)
	r.GET("/files/*filepath", emptyHandler)
	r.GET("/static/*filepath", emptyHandler)
	r.POST(benchPathAPIV1Users, emptyHandler)
	r.PUT(benchReqUsersIDColon, emptyHandler)
	r.DELETE(benchReqUsersIDColon, emptyHandler)

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
	req := httptest.NewRequest(http.MethodGet, benchReqUsers12345, nil)
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
	req := httptest.NewRequest(http.MethodGet, benchReqOrgProjects, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkComparisonHttpRouterWildcard(b *testing.B) {
	r := setupHttpRouter()
	req := httptest.NewRequest(http.MethodGet, benchReqFilesLogoPath, nil)
	rr := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}
