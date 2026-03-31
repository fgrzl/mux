package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/builder"
	openapi "github.com/fgrzl/mux/pkg/openapi"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type routeGroupCloneNested struct {
	Value string `json:"value"`
}

type routeGroupClonePayload struct {
	Name   *string                `json:"name"`
	Nested *routeGroupCloneNested `json:"nested"`
	Tags   []string               `json:"tags"`
}

func TestShouldCreateNestedRouteGroupsWhenCallingNewRouteGroup(t *testing.T) {
	// Arrange: Create a router and nested group structure
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	v1 := api.NewRouteGroup("/v1")
	v1.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	})

	// Act: Make request to the nested route
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	// Assert: Route was registered with correct path and responds
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "users")
}

func TestShouldInheritDefaultsWhenCreatingNestedRouteGroup(t *testing.T) {
	// Arrange: Create parent group with defaults
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.WithTags("API").
		RequireRoles("user").
		WithSummary("API Summary").
		WithDescription("API Description").
		AllowAnonymous().
		Deprecated()

	// Act: Create nested group
	v1 := api.NewRouteGroup("/v1")

	// Assert: All defaults are inherited
	assert.Equal(t, "/api/v1", v1.prefix)
	assert.Equal(t, api.routeRegistry, v1.routeRegistry)
	assert.Equal(t, api.defaultTags, v1.defaultTags)
	assert.Equal(t, api.defaultRoles, v1.defaultRoles)
	assert.Equal(t, api.defaultSummary, v1.defaultSummary)
	assert.Equal(t, api.defaultDescription, v1.defaultDescription)
	assert.Equal(t, api.defaultAllowAnon, v1.defaultAllowAnon)
	assert.Equal(t, api.defaultDeprecated, v1.defaultDeprecated)
}

func TestShouldMaintainIndependentDefaultsWhenModifyingNestedRouteGroup(t *testing.T) {
	// Arrange: Create parent group with defaults
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.WithTags("API")

	// Act: Create nested group and add different defaults
	v1 := api.NewRouteGroup("/v1")
	v1.WithTags("V1")

	// Assert: Parent defaults are preserved and child defaults are added
	assert.Equal(t, []string{"API"}, api.defaultTags)
	assert.Equal(t, []string{"API", "V1"}, v1.defaultTags)
}

func TestShouldKeepParentAndChildMiddlewareDefaultsIndependent(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.defaultMiddleware = make([]Middleware, 0, 4)
	mwAPI := &testMiddleware{invoke: func(c routing.RouteContext, next HandlerFunc) { next(c) }}
	mwV1 := &testMiddleware{invoke: func(c routing.RouteContext, next HandlerFunc) { next(c) }}
	api.Use(mwAPI)

	// Act
	v1 := api.NewRouteGroup("/v1")
	v1.Use(mwV1)

	// Assert
	require.Len(t, api.defaultMiddleware, 1)
	assert.Same(t, mwAPI, api.defaultMiddleware[0])
	assert.Equal(t, []Middleware{mwAPI, mwV1}, v1.defaultMiddleware)
}

func TestShouldCopyGroupMiddlewareIntoRegisteredRoutes(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	mwAPI := &testMiddleware{invoke: func(c routing.RouteContext, next HandlerFunc) { next(c) }}
	mwV1 := &testMiddleware{invoke: func(c routing.RouteContext, next HandlerFunc) { next(c) }}
	api.Use(mwAPI)
	v1 := api.NewRouteGroup("/v1")
	v1.Use(mwV1)

	// Act
	route := v1.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	})

	// Assert
	require.Len(t, route.Options.Middleware, 2)
	assert.Same(t, mwAPI, route.Options.Middleware[0])
	assert.Same(t, mwV1, route.Options.Middleware[1])
}

func TestShouldKeepParentAndChildServiceDefaultsIndependent(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.WithService(routing.ServiceKey("db"), "primary")

	// Act
	v1 := api.NewRouteGroup("/v1")
	v1.WithService(routing.ServiceKey("cache"), "redis")

	// Assert
	assert.Equal(t, map[routing.ServiceKey]any{routing.ServiceKey("db"): "primary"}, api.defaultServices)
	assert.Equal(t, map[routing.ServiceKey]any{
		routing.ServiceKey("db"):    "primary",
		routing.ServiceKey("cache"): "redis",
	}, v1.defaultServices)
}

func TestShouldCopyGroupServicesIntoRegisteredRoutes(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.WithService(routing.ServiceKey("db"), "primary")
	v1 := api.NewRouteGroup("/v1")
	v1.WithService(routing.ServiceKey("cache"), "redis")

	// Act
	route := v1.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	})

	// Assert
	assert.Equal(t, map[routing.ServiceKey]any{
		routing.ServiceKey("db"):    "primary",
		routing.ServiceKey("cache"): "redis",
	}, route.Options.Services)
}

func TestShouldExposeServiceRegistryOnRouteGroup(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	registry := api.Services()

	// Act
	result := registry.Register(routing.ServiceKey("db"), "primary").Register(routing.ServiceKey("cache"), "redis")
	db, ok := registry.Get(routing.ServiceKey("db"))

	// Assert
	assert.Same(t, registry, result)
	assert.True(t, ok)
	assert.Equal(t, "primary", db)
	assert.Equal(t, map[routing.ServiceKey]any{
		routing.ServiceKey("db"):    "primary",
		routing.ServiceKey("cache"): "redis",
	}, api.defaultServices)
}

func TestShouldKeepParentAndChildDefaultsIndependentWithSharedCapacity(t *testing.T) {
	// Arrange: force extra capacity so append reuse would leak across groups
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.defaultTags = make([]string, 0, 4)
	api.defaultTags = append(api.defaultTags, "API")

	// Act
	v1 := api.NewRouteGroup("/v1")
	v1.WithTags("V1")
	api.WithTags("ADMIN")

	// Assert
	assert.Equal(t, []string{"API", "ADMIN"}, api.defaultTags)
	assert.Equal(t, []string{"API", "V1"}, v1.defaultTags)
}

func TestShouldSupportMultipleLevelsOfNestingWhenCreatingRouteGroups(t *testing.T) {
	// Arrange: Create multiple levels of nesting
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	v1 := api.NewRouteGroup("/v1")
	users := v1.NewRouteGroup("/users")
	users.GET("/profile", func(c routing.RouteContext) {
		c.OK("profile")
	})

	// Act: Make request to deeply nested route
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	// Assert: Route was registered with correct path and responds
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "profile")
}

func TestShouldNormalizePrefixesWhenCreatingNestedRouteGroups(t *testing.T) {
	// Arrange: Define test cases with various prefix formats
	rtr := NewRouter()
	testCases := []struct {
		parentPrefix string
		childPrefix  string
		expectedPath string
	}{
		{"/api", "/v1", "/api/v1"},
		{"/api/", "v1", "/api/v1"},
		{"api", "/v1", "/api/v1"},
		{"api/", "v1/", "/api/v1/"}, // Fixed: trailing slash should be preserved
		{"/api", "/api/v1", "/api/v1"},
		{"/api", "/api/v1/", "/api/v1/"},
	}

	// Act & Assert: Test each case
	for _, tc := range testCases {
		parent := rtr.NewRouteGroup(tc.parentPrefix)
		child := parent.NewRouteGroup(tc.childPrefix)

		assert.Equal(t, tc.expectedPath, child.prefix,
			"Parent: %s, Child: %s should result in: %s",
			tc.parentPrefix, tc.childPrefix, tc.expectedPath)
	}
}

func TestShouldNotDoublePrefixAbsoluteRoutePatternsWithinGroup(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.GET("/api/users", func(c routing.RouteContext) {
		c.OK("users")
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "users")
}

func TestShouldAttachDetachedRouteBuilderWithGroupDefaultsWithoutMutatingSource(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.WithTags("API").RequireRoles("user").WithQueryParam("tenant", "tenant identifier", "tenant-a")

	type createUserRequest struct {
		Name string `json:"name"`
	}
	type createUserResponse struct {
		ID string `json:"id"`
	}

	detached := builder.DetachedRoute(http.MethodPost, "/users").
		WithOperationID("createUser").
		WithSummary("Create user").
		WithJsonBody(createUserRequest{Name: "Jane"}).
		WithCreatedResponse(createUserResponse{ID: "u-1"})

	// Act
	attached := api.HandleRoute(detached, func(c routing.RouteContext) {
		c.NoContent()
	})
	request := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	recorder := httptest.NewRecorder()
	rtr.ServeHTTP(recorder, request)

	// Assert
	require.NotNil(t, attached)
	assert.Equal(t, http.MethodPost, attached.Options.Method)
	assert.Equal(t, "/api/users", attached.Options.Pattern)
	assert.Equal(t, "createUser", attached.Options.OperationID)
	assert.Equal(t, "Create user", attached.Options.Summary)
	assert.Equal(t, []string{"API"}, attached.Options.Tags)
	assert.Equal(t, []string{"user"}, attached.Options.Roles)
	require.Len(t, attached.Options.Parameters, 1)
	assert.Equal(t, "tenant", attached.Options.Parameters[0].Name)
	assert.Equal(t, http.StatusNoContent, recorder.Code)
	require.NotNil(t, attached.Options.RequestBody)
	assert.Contains(t, attached.Options.Responses, "201")

	assert.Equal(t, "/users", detached.Options.Pattern)
	assert.Empty(t, detached.Options.Tags)
	assert.Empty(t, detached.Options.Roles)
	assert.Len(t, detached.Options.Parameters, 0)
}

func TestShouldInheritParametersWhenCreatingNestedRouteGroup(t *testing.T) {
	// Arrange: Create parent group with parameters
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.WithPathParam("version", "", "v1").
		WithQueryParam("limit", "", 10).
		RequireRoles("admin")

	// Act: Create nested group and add route
	v1 := api.NewRouteGroup("/v1")
	v1.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	})

	// Assert: Route inherits parameters and defaults
	options, _, found := rtr.routeRegistry.Load("/api/v1/users", http.MethodGet)
	require.True(t, found, "Route should be registered")
	assert.Contains(t, options.Roles, "admin")
	assert.Len(t, options.Operation.Parameters, 2) // version and limit params
}

func TestShouldWorkAsExpectedWhenUsingExampleFromIssue(t *testing.T) {
	// Arrange: Set up the exact example from the issue
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	v1 := api.NewRouteGroup("/v1")
	v1.GET("/users", func(c routing.RouteContext) {
		c.OK("users endpoint")
	})

	// Act: Make request to the route
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	// Assert: Route works as expected
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "users endpoint")
}

func TestShouldMarkPathParameterRequiredWhenUsingLowLevelWithParam(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.WithParam("id", "path", "user identifier", "123", false)

	// Act
	api.GET("/users/{id}", func(c routing.RouteContext) {
		c.OK("users")
	})
	options, _, found := rtr.routeRegistry.Load("/api/users/{id}", http.MethodGet)

	// Assert
	require.True(t, found, "Route should be registered")
	require.Len(t, options.Operation.Parameters, 1)
	assert.True(t, options.Operation.Parameters[0].Required)
	assert.Equal(t, "123", options.Operation.Parameters[0].Example)
	assert.NotNil(t, options.Operation.Parameters[0].Converter)
	assert.Equal(t, "path", options.Operation.Parameters[0].In)
}

func TestShouldPanicWhenUsingInvalidParameterLocationOnRouteGroup(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")

	// Act / Assert
	assert.PanicsWithValue(t, "invalid parameter 'in': \"matrix\"", func() {
		api.WithParam("id", "matrix", "user identifier", "123", false)
	})
}

func TestShouldDeepCopyParametersWhenCreatingNestedRouteGroups(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.WithQueryParam("versions", "supported versions", []string{"v1"})

	// Act
	v1 := api.NewRouteGroup("/v1")
	childParam := v1.defaultParams[0]
	childParam.Schema.Items.Type = "integer"
	childExample := childParam.Example.([]string)
	childExample[0] = "v2"

	// Assert
	require.Len(t, api.defaultParams, 1)
	require.Len(t, v1.defaultParams, 1)
	assert.NotSame(t, api.defaultParams[0], childParam)
	assert.NotSame(t, api.defaultParams[0].Schema, childParam.Schema)
	assert.NotSame(t, api.defaultParams[0].Schema.Items, childParam.Schema.Items)
	assert.Equal(t, []string{"v1"}, api.defaultParams[0].Example.([]string))
	assert.Equal(t, "string", api.defaultParams[0].Schema.Items.Type)
	assert.Equal(t, []string{"v2"}, childParam.Example.([]string))
	assert.Equal(t, "integer", childParam.Schema.Items.Type)
}

func TestShouldKeepRegisteredRouteParametersIndependentFromGroupDefaults(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.WithQueryParam("versions", "supported versions", []string{"v1"})

	// Act
	usersRoute := api.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	})
	usersParam := usersRoute.Options.Operation.Parameters[0]
	usersParam.Schema.Items.Type = "integer"
	usersExample := usersParam.Example.([]string)
	usersExample[0] = "v2"
	adminsRoute := api.GET("/admins", func(c routing.RouteContext) {
		c.OK("admins")
	})

	// Assert
	require.Len(t, api.defaultParams, 1)
	require.Len(t, usersRoute.Options.Operation.Parameters, 1)
	require.Len(t, adminsRoute.Options.Operation.Parameters, 1)
	assert.NotSame(t, api.defaultParams[0], usersParam)
	assert.Equal(t, []string{"v1"}, api.defaultParams[0].Example.([]string))
	assert.Equal(t, "string", api.defaultParams[0].Schema.Items.Type)
	assert.Equal(t, []string{"v1"}, adminsRoute.Options.Operation.Parameters[0].Example.([]string))
	assert.Equal(t, "string", adminsRoute.Options.Operation.Parameters[0].Schema.Items.Type)
}

func TestShouldOwnPointerBackedExamplesWhenAddingGroupDefaults(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	name := "stable"
	example := &routeGroupClonePayload{
		Name:   &name,
		Nested: &routeGroupCloneNested{Value: "root"},
		Tags:   []string{"one"},
	}

	// Act
	api.WithQueryParam("payload", "pointer-backed payload", example)
	stored := api.defaultParams[0].Example.(*routeGroupClonePayload)
	*example.Name = "changed"
	example.Nested.Value = "mutated"
	example.Tags[0] = "two"

	// Assert
	require.Len(t, api.defaultParams, 1)
	assert.NotSame(t, example, stored)
	assert.NotSame(t, example.Name, stored.Name)
	assert.NotSame(t, example.Nested, stored.Nested)
	assert.Equal(t, "stable", *stored.Name)
	assert.Equal(t, "root", stored.Nested.Value)
	assert.Equal(t, []string{"one"}, stored.Tags)
}

func TestShouldOwnSecurityRequirementsWhenAddingGroupDefaults(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	security := &openapi.SecurityRequirement{"oauth2": []string{"read"}}

	// Act
	api.WithSecurity(security)
	stored := api.defaultSecurity[0]
	(*security)["oauth2"].([]string)[0] = "write"

	// Assert
	require.Len(t, api.defaultSecurity, 1)
	assert.NotSame(t, security, stored)
	assert.Equal(t, []string{"read"}, (*stored)["oauth2"].([]string))
}

func TestShouldDeepCopySecurityRequirementsWhenCreatingNestedRouteGroups(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.WithSecurity(&openapi.SecurityRequirement{"oauth2": []string{"read"}})

	// Act
	v1 := api.NewRouteGroup("/v1")
	childSecurity := v1.defaultSecurity[0]
	childScopes := (*childSecurity)["oauth2"].([]string)
	childScopes[0] = "write"

	// Assert
	require.Len(t, api.defaultSecurity, 1)
	require.Len(t, v1.defaultSecurity, 1)
	assert.NotSame(t, api.defaultSecurity[0], childSecurity)
	assert.Equal(t, []string{"read"}, (*api.defaultSecurity[0])["oauth2"].([]string))
	assert.Equal(t, []string{"write"}, (*childSecurity)["oauth2"].([]string))
}

func TestShouldKeepRegisteredRouteSecurityIndependentFromGroupDefaults(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	api.WithSecurity(&openapi.SecurityRequirement{"oauth2": []string{"read"}})

	// Act
	usersRoute := api.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	})
	usersSecurity := usersRoute.Options.Operation.Security[0]
	usersScopes := (*usersSecurity)["oauth2"].([]string)
	usersScopes[0] = "write"
	adminsRoute := api.GET("/admins", func(c routing.RouteContext) {
		c.OK("admins")
	})

	// Assert
	require.Len(t, api.defaultSecurity, 1)
	require.Len(t, usersRoute.Options.Operation.Security, 1)
	require.Len(t, adminsRoute.Options.Operation.Security, 1)
	assert.NotSame(t, api.defaultSecurity[0], usersSecurity)
	assert.Equal(t, []string{"read"}, (*api.defaultSecurity[0])["oauth2"].([]string))
	assert.Equal(t, []string{"read"}, (*adminsRoute.Options.Operation.Security[0])["oauth2"].([]string))
}

func TestShouldDeepCopyPointerBackedExamplesWhenCreatingNestedRouteGroups(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	name := "stable"
	example := &routeGroupClonePayload{
		Name:   &name,
		Nested: &routeGroupCloneNested{Value: "root"},
		Tags:   []string{"one"},
	}
	api.WithQueryParam("payload", "pointer-backed payload", example)

	// Act
	v1 := api.NewRouteGroup("/v1")
	clonedExample := v1.defaultParams[0].Example.(*routeGroupClonePayload)
	*clonedExample.Name = "beta"
	clonedExample.Nested.Value = "child"
	clonedExample.Tags[0] = "two"

	// Assert
	require.Len(t, api.defaultParams, 1)
	assert.NotSame(t, example, clonedExample)
	assert.NotSame(t, example.Name, clonedExample.Name)
	assert.NotSame(t, example.Nested, clonedExample.Nested)
	assert.Equal(t, "stable", *example.Name)
	assert.Equal(t, "root", example.Nested.Value)
	assert.Equal(t, []string{"one"}, example.Tags)
	assert.Equal(t, "stable", *api.defaultParams[0].Example.(*routeGroupClonePayload).Name)
}

func TestShouldKeepRegisteredRoutePointerExamplesIndependentFromGroupDefaults(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	api := rtr.NewRouteGroup("/api")
	name := "stable"
	example := &routeGroupClonePayload{
		Name:   &name,
		Nested: &routeGroupCloneNested{Value: "root"},
		Tags:   []string{"one"},
	}
	api.WithQueryParam("payload", "pointer-backed payload", example)

	// Act
	usersRoute := api.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	})
	usersExample := usersRoute.Options.Operation.Parameters[0].Example.(*routeGroupClonePayload)
	*usersExample.Name = "beta"
	usersExample.Nested.Value = "child"
	usersExample.Tags[0] = "two"
	adminsRoute := api.GET("/admins", func(c routing.RouteContext) {
		c.OK("admins")
	})
	adminsExample := adminsRoute.Options.Operation.Parameters[0].Example.(*routeGroupClonePayload)

	// Assert
	require.Len(t, usersRoute.Options.Operation.Parameters, 1)
	require.Len(t, adminsRoute.Options.Operation.Parameters, 1)
	assert.NotSame(t, example, usersExample)
	assert.NotSame(t, example.Name, usersExample.Name)
	assert.NotSame(t, example.Nested, usersExample.Nested)
	assert.Equal(t, "stable", *example.Name)
	assert.Equal(t, "root", example.Nested.Value)
	assert.Equal(t, []string{"one"}, example.Tags)
	assert.Equal(t, "stable", *adminsExample.Name)
	assert.Equal(t, "root", adminsExample.Nested.Value)
	assert.Equal(t, []string{"one"}, adminsExample.Tags)
}
