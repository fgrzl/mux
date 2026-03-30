package router

import (
	"net/http"
	"testing"

	openapi "github.com/fgrzl/mux/pkg/openapi"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/fgrzl/mux/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type routerCloneNested struct {
	Value string `json:"value"`
}

type routerClonePayload struct {
	Name   *string            `json:"name"`
	Nested *routerCloneNested `json:"nested"`
	Tags   []string           `json:"tags"`
}

func TestShouldCreateNewRouterWithDefaultOptions(t *testing.T) {
	// Arrange & Act
	rtr := NewRouter()

	// Assert
	assert.NotNil(t, rtr)
	assert.NotNil(t, rtr.RouteGroup)
	assert.NotNil(t, rtr.routeRegistry)
	assert.NotNil(t, rtr.options)
	assert.Equal(t, "", rtr.prefix)
}

func TestShouldCreateNewRouterWithOptions(t *testing.T) {
	// Arrange
	title := "Test API"
	version := "1.0.0"

	// Act
	rtr := NewRouter(WithTitle(title), WithVersion(version))

	// Assert
	assert.NotNil(t, rtr)
	assert.NotNil(t, rtr.options.openapi)
	assert.Equal(t, title, rtr.options.openapi.Title)
	assert.Equal(t, version, rtr.options.openapi.Version)
}

func TestInfoObjectShouldReturnIndependentCopy(t *testing.T) {
	// Arrange
	rtr := NewRouter(
		WithTitle("Test API"),
		WithVersion("1.0.0"),
		WithContact("Support", "https://example.com", "support@example.com"),
		WithLicense("MIT", "https://example.com/license"),
	)

	// Act
	info, err := rtr.InfoObject()
	require.NoError(t, err)
	info.Title = "Changed"
	info.Contact.Name = "Changed Contact"
	info.License.Name = "Changed License"
	fresh, err := rtr.InfoObject()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Test API", fresh.Title)
	assert.Equal(t, "Support", fresh.Contact.Name)
	assert.Equal(t, "MIT", fresh.License.Name)
	assert.NotSame(t, info, fresh)
	assert.NotSame(t, info.Contact, fresh.Contact)
	assert.NotSame(t, info.License, fresh.License)
}

func TestShouldCreateNewRouteGroupWithPrefix(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	prefix := "/api/v1"

	// Act
	group := rtr.NewRouteGroup(prefix)

	// Assert
	assert.NotNil(t, group)
	assert.Equal(t, "/api/v1", group.prefix) // Based on normalizeRoute behavior
	assert.Equal(t, rtr.routeRegistry, group.routeRegistry)
}

func TestShouldServeHTTPAndReturnNotFoundForUnknownRoute(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/unknown", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestShouldServeHTTPAndCallRegisteredHandler(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	called := false
	rtr.GET("/test", func(c routing.RouteContext) {
		called = true
		c.OK("success")
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRoutesShouldReturnIndependentOperationCopies(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	name := "stable"
	example := &routerClonePayload{
		Name:   &name,
		Nested: &routerCloneNested{Value: "root"},
		Tags:   []string{"one"},
	}
	security := &openapi.SecurityRequirement{"oauth2": []string{"read"}}
	rtr.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	}).
		WithTags("users").
		WithExternalDocs("https://example.com/docs", "User docs").
		WithSecurity(security).
		WithQueryParam("payload", "pointer-backed payload", example)

	// Act
	routes, err := rtr.Routes()
	require.NoError(t, err)
	require.Len(t, routes, 1)
	returned := routes[0].Options
	returned.Tags[0] = "admins"
	returned.ExternalDocs.Description = "Changed docs"
	returned.Security[0] = &openapi.SecurityRequirement{"oauth2": []string{"write"}}
	returnedExample := returned.Parameters[0].Example.(*routerClonePayload)
	*returnedExample.Name = "changed"
	returnedExample.Nested.Value = "mutated"
	returnedExample.Tags[0] = "two"
	freshRoutes, err := rtr.Routes()

	// Assert
	require.NoError(t, err)
	require.Len(t, freshRoutes, 1)
	fresh := freshRoutes[0].Options
	assert.Equal(t, []string{"users"}, fresh.Tags)
	assert.Equal(t, "User docs", fresh.ExternalDocs.Description)
	assert.Equal(t, []string{"read"}, (*fresh.Security[0])["oauth2"].([]string))
	freshExample := fresh.Parameters[0].Example.(*routerClonePayload)
	assert.Equal(t, "stable", *freshExample.Name)
	assert.Equal(t, "root", freshExample.Nested.Value)
	assert.Equal(t, []string{"one"}, freshExample.Tags)
	assert.NotSame(t, returned, fresh)
	assert.NotSame(t, returned.ExternalDocs, fresh.ExternalDocs)
	assert.NotSame(t, returned.Parameters[0], fresh.Parameters[0])
	assert.NotSame(t, returnedExample, freshExample)
	assert.NotSame(t, returned.Security[0], fresh.Security[0])
}

func TestShouldServeHTTPWithMiddleware(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	middlewareExecuted := false

	// Add middleware
	rtr.middleware = append(rtr.middleware, &testMiddleware{
		invoke: func(c routing.RouteContext, next HandlerFunc) {
			middlewareExecuted = true
			next(c)
		},
	})

	handlerExecuted := false
	rtr.GET("/test", func(c routing.RouteContext) {
		handlerExecuted = true
		c.OK("success")
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.True(t, middlewareExecuted)
	assert.True(t, handlerExecuted)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldSetRouteParamsInContext(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	var receivedID string
	rtr.GET("/users/{id}", func(c routing.RouteContext) {
		receivedID, _ = c.Param("id")
		c.OK("success")
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/users/123", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "123", receivedID)
}

func TestShouldExecuteMiddlewareInCorrectOrder(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	executionOrder := []int{}

	// Add middleware in order
	rtr.middleware = append(rtr.middleware, &testMiddleware{
		invoke: func(c routing.RouteContext, next HandlerFunc) {
			executionOrder = append(executionOrder, 1)
			next(c)
		},
	})
	rtr.middleware = append(rtr.middleware, &testMiddleware{
		invoke: func(c routing.RouteContext, next HandlerFunc) {
			executionOrder = append(executionOrder, 2)
			next(c)
		},
	})

	rtr.GET("/test", func(c routing.RouteContext) {
		executionOrder = append(executionOrder, 3)
		c.OK("success")
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, []int{1, 2, 3}, executionOrder)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestShouldStopMiddlewareChainWhenNotContinuing(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	middlewareExecuted := false
	handlerExecuted := false

	// Add middleware that doesn't call next
	rtr.middleware = append(rtr.middleware, &testMiddleware{
		invoke: func(c routing.RouteContext, next HandlerFunc) {
			middlewareExecuted = true
			c.Unauthorized()
			// Don't call next(c)
		},
	})

	rtr.GET("/test", func(c routing.RouteContext) {
		handlerExecuted = true
		c.OK("success")
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.True(t, middlewareExecuted)
	assert.False(t, handlerExecuted) // Handler should not be executed
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// Test middleware implementation for testing
type testMiddleware struct {
	invoke func(c routing.RouteContext, next HandlerFunc)
}

func (tm *testMiddleware) Invoke(c routing.RouteContext, next HandlerFunc) {
	tm.invoke(c, next)
}
