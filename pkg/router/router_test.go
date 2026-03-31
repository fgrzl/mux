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

func TestShouldServeHTTPAndCallRegisteredStdlibHandlerFunc(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	called := false
	api := rtr.NewRouteGroup("/api")
	api.WithService(routing.ServiceKey("db"), "primary")
	api.Use(&testMiddleware{
		invoke: func(c routing.RouteContext, next HandlerFunc) {
			c.SetContextValue("middleware", "hit")
			next(c)
		},
	})
	api.HandleFunc(http.MethodGet, "/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		called = true
		routeCtx, ok := routing.RouteContextFromRequest(r)
		require.True(t, ok)
		id, ok := routeCtx.Param("id")
		require.True(t, ok)
		assert.Equal(t, "42", id)
		service, ok := routeCtx.GetService(routing.ServiceKey("db"))
		require.True(t, ok)
		assert.Equal(t, "primary", service)
		assert.Equal(t, "hit", r.Context().Value("middleware"))
		w.WriteHeader(http.StatusNoContent)
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/api/users/42", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.True(t, called)
	assert.Equal(t, http.StatusNoContent, rec.Code)
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

func TestShouldInheritRootScopedServicesWhenCreatingRouteGroup(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	rtr.WithService(routing.ServiceKey("db"), "primary")

	// Act
	api := rtr.NewRouteGroup("/api")
	route := api.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	})

	// Assert
	assert.Equal(t, "primary", route.Options.Services[routing.ServiceKey("db")])
}

func TestShouldExposeScopedServicesToRouterMiddlewareAndHandler(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	serviceKey := routing.ServiceKey("db")
	groupKey := routing.ServiceKey("cache")
	var middlewareDB any
	var middlewareCache any
	var handlerDB any
	var handlerCache any
	rtr.Use(&testMiddleware{
		invoke: func(c routing.RouteContext, next HandlerFunc) {
			middlewareDB, _ = c.GetService(serviceKey)
			middlewareCache, _ = c.GetService(groupKey)
			next(c)
		},
	})
	rtr.WithService(serviceKey, "root-db")
	api := rtr.NewRouteGroup("/api")
	api.WithService(groupKey, "redis")
	api.GET("/users", func(c routing.RouteContext) {
		handlerDB, _ = c.GetService(serviceKey)
		handlerCache, _ = c.GetService(groupKey)
		c.OK("users")
	}).WithService(serviceKey, "route-db")

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/api/users", nil)

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "route-db", middlewareDB)
	assert.Equal(t, "redis", middlewareCache)
	assert.Equal(t, "route-db", handlerDB)
	assert.Equal(t, "redis", handlerCache)
}

func TestShouldExposeServiceRegistryOnRouter(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	registry := rtr.Services()

	// Act
	result := registry.Register(routing.ServiceKey("db"), "primary")
	api := rtr.NewRouteGroup("/api")
	route := api.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	})
	retrieved, ok := registry.Get(routing.ServiceKey("db"))

	// Assert
	assert.Same(t, registry, result)
	assert.True(t, ok)
	assert.Equal(t, "primary", retrieved)
	assert.Equal(t, "primary", route.Options.Services[routing.ServiceKey("db")])
}

func TestShouldReturnRouterWhenRegisteringRootScopedService(t *testing.T) {
	// Arrange
	rtr := NewRouter()

	// Act
	result := rtr.WithService(routing.ServiceKey("db"), "primary")
	api := rtr.NewRouteGroup("/api")
	route := api.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	})

	// Assert
	assert.Same(t, rtr, result)
	assert.Equal(t, "primary", route.Options.Services[routing.ServiceKey("db")])
}

func TestShouldAccumulateValidationErrorsOnRouterWithoutPanicking(t *testing.T) {
	// Arrange
	rtr := NewRouter().Safe()

	// Act / Assert
	assert.NotPanics(t, func() {
		rtr.NewRouteGroup("/api").GET("/users", func(c routing.RouteContext) {
			c.OK("users")
		}).WithOperationID("invalid-id")
	})

	// Assert
	require.Len(t, rtr.Errors(), 1)
	assert.ErrorContains(t, rtr.Err(), "invalid OperationID")
}

func TestConfigureShouldReturnValidationErrorsWithoutChangingDefaultPanicBehavior(t *testing.T) {
	// Arrange
	rtr := NewRouter()

	// Act
	err := rtr.Configure(func(router *Router) {
		router.NewRouteGroup("/api").GET("/users", func(c routing.RouteContext) {
			c.OK("users")
		}).WithOperationID("invalid-id")
	})

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid OperationID")
	require.Len(t, rtr.Errors(), 1)
	assert.Panics(t, func() {
		rtr.GET("/panic", func(c routing.RouteContext) {
			c.OK("panic")
		}).WithOperationID("still-invalid")
	})
}

func TestShouldRegisterPatchOptionsAndTraceRoutesViaRouterHelpers(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	rtr.PATCH("/resource", func(c routing.RouteContext) {
		c.Response().Header().Set("X-Method", http.MethodPatch)
		c.NoContent()
	})
	rtr.OPTIONS("/resource", func(c routing.RouteContext) {
		c.Response().Header().Set("X-Method", http.MethodOptions)
		c.NoContent()
	})
	rtr.TRACE("/resource", func(c routing.RouteContext) {
		c.Response().Header().Set("X-Method", http.MethodTrace)
		c.NoContent()
	})

	tests := []struct {
		method string
	}{
		{method: http.MethodPatch},
		{method: http.MethodOptions},
		{method: http.MethodTrace},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req, rec := testhelpers.NewRequestRecorder(tt.method, "/resource", nil)

			// Act
			rtr.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Equal(t, tt.method, rec.Header().Get("X-Method"))
		})
	}
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
