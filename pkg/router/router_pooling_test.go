package router

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
	"github.com/stretchr/testify/assert"
)

func init() {
	// silence slog during tests
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
}

func TestShouldServeExactRouteGivenContextPoolingEnabled(t *testing.T) {
	// Arrange
	r := NewRouter(WithContextPooling())
	rg := r.NewRouteGroup("")
	rg.GET("/hello", func(c routing.RouteContext) {
		c.Response().Header().Set("X-Test", "1")
		c.OK(map[string]string{"msg": "hi"})
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rr := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "1", rr.Header().Get("X-Test"))
	assert.Contains(t, rr.Body.String(), "\"msg\":\"hi\"")
}

func TestShouldServeHeadViaGetWithoutBodyGivenFallbackEnabled(t *testing.T) {
	// Arrange
	r := NewRouter(WithHeadFallbackToGet())
	rg := r.NewRouteGroup("")
	rg.GET("/resource", func(c routing.RouteContext) {
		c.Response().Header().Set("X-From", "GET")
		c.OK("body")
	})

	req := httptest.NewRequest(http.MethodHead, "/resource", nil)
	rr := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "GET", rr.Header().Get("X-From"))
	assert.Equal(t, 0, rr.Body.Len(), "HEAD fallback must suppress body")
}

func TestShouldReturn405WithAllowHeaderGivenHeadWithoutFallback(t *testing.T) {
	// Arrange
	r := NewRouter() // no fallback
	rg := r.NewRouteGroup("")
	rg.GET("/only-get", func(c routing.RouteContext) { c.OK("ok") })

	req := httptest.NewRequest(http.MethodHead, "/only-get", nil)
	rr := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	allow := rr.Header().Get("Allow")
	// Allow should list permitted methods for this path (GET)
	assert.True(t, strings.Contains(allow, http.MethodGet))
}

func TestShouldReturn405WithAllowHeaderGivenMethodNotAllowed(t *testing.T) {
	// Arrange
	r := NewRouter()
	rg := r.NewRouteGroup("")
	rg.GET("/path", func(c routing.RouteContext) { c.OK("ok") })

	req := httptest.NewRequest(http.MethodPost, "/path", nil)
	rr := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	allow := rr.Header().Get("Allow")
	assert.True(t, strings.Contains(allow, http.MethodGet))
}

func TestShouldReturn500GivenHandlerPanic(t *testing.T) {
	// Arrange
	r := NewRouter()
	rg := r.NewRouteGroup("")
	rg.GET("/panic", func(c routing.RouteContext) { panic("boom") })

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestShouldReturn404GivenNoMatchingRoute(t *testing.T) {
	// Arrange
	r := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestShouldReturn404GivenHeadWithNoRouteAndNoFallback(t *testing.T) {
	// Arrange
	r := NewRouter()
	req := httptest.NewRequest(http.MethodHead, "/missing", nil)
	rr := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestShouldReturn405GivenHeadWithFallbackButNoGetRoute(t *testing.T) {
	// Arrange
	r := NewRouter(WithHeadFallbackToGet())
	rg := r.NewRouteGroup("")
	rg.POST("/res", func(c routing.RouteContext) { c.OK("ok") })

	req := httptest.NewRequest(http.MethodHead, "/res", nil)
	rr := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Header().Get("Allow"), http.MethodPost)
}

// middleware to capture order
type orderMW struct {
	id   string
	seen *[]string
}

func (m *orderMW) Invoke(c routing.RouteContext, next HandlerFunc) {
	*m.seen = append(*m.seen, "before:"+m.id)
	next(c)
	*m.seen = append(*m.seen, "after:"+m.id)
}

type stopMW struct {
	id   string
	seen *[]string
}

func (m *stopMW) Invoke(c routing.RouteContext, next HandlerFunc) {
	*m.seen = append(*m.seen, "before:"+m.id)
	// do not call next: short-circuit
	c.OK("stopped")
	*m.seen = append(*m.seen, "after:"+m.id)
}

func TestShouldExecuteMiddlewareInOrderAndStopGivenShortCircuit(t *testing.T) {
	// Arrange
	r := NewRouter()
	var seen []string
	r.Use(&orderMW{id: "A", seen: &seen})
	r.Use(&orderMW{id: "B", seen: &seen})
	rg := r.NewRouteGroup("")
	rg.GET("/x", func(c routing.RouteContext) { c.OK("ok") })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	// Middleware run in registration order (first registered runs first): A then B
	assert.Equal(t, []string{"before:A", "before:B", "after:B", "after:A"}, seen)

	// Arrange - Now add a stopping middleware and assert short-circuit
	seen = nil
	r2 := NewRouter()
	r2.Use(&orderMW{id: "A", seen: &seen})
	r2.Use(&stopMW{id: "S", seen: &seen})
	r2.Use(&orderMW{id: "B", seen: &seen})
	rg2 := r2.NewRouteGroup("")
	rg2.GET("/x", func(c routing.RouteContext) { c.OK("ok") })

	req2 := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr2 := httptest.NewRecorder()

	// Act
	r2.ServeHTTP(rr2, req2)

	// Assert
	assert.Equal(t, http.StatusOK, rr2.Code)
	// Registration order: A, S (stops), B. Execution enters A, then S (stops), so B never runs.
	assert.Equal(t, []string{"before:A", "before:S", "after:S", "after:A"}, seen)
}

func TestShouldExecuteRouterGroupAndRouteMiddlewareInOrder(t *testing.T) {
	// Arrange
	r := NewRouter()
	var seen []string
	r.Use(&orderMW{id: "router", seen: &seen})
	api := r.NewRouteGroup("/api")
	api.Use(&orderMW{id: "group", seen: &seen})
	api.GET("/users", func(c routing.RouteContext) {
		seen = append(seen, "handler")
		c.OK("ok")
	}).Use(&orderMW{id: "route", seen: &seen})

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, []string{
		"before:router",
		"before:group",
		"before:route",
		"handler",
		"after:route",
		"after:group",
		"after:router",
	}, seen)
}

func TestShouldExecuteScopedMiddlewareWithoutRouterMiddlewareOnFastPath(t *testing.T) {
	// Arrange
	r := NewRouter()
	var seen []string
	handlerExecuted := false
	api := r.NewRouteGroup("/api")
	api.Use(&orderMW{id: "group", seen: &seen})
	api.GET("/users", func(c routing.RouteContext) {
		handlerExecuted = true
		c.OK("ok")
	}).Use(&stopMW{id: "route", seen: &seen})

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()

	// Act
	r.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, handlerExecuted)
	assert.Equal(t, []string{"before:group", "before:route", "after:route", "after:group"}, seen)
}
