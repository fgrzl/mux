package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldCreateNestedRouteGroupsWhenCallingNewRouteGroup(t *testing.T) {
	// Arrange: Create a router and nested group structure
	router := NewRouter()
	api := router.NewRouteGroup("/api")
	v1 := api.NewRouteGroup("/v1")
	v1.GET("/users", func(c RouteContext) {
		c.OK("users")
	})

	// Act: Make request to the nested route
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Route was registered with correct path and responds
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "users")
}

func TestShouldInheritDefaultsWhenCreatingNestedRouteGroup(t *testing.T) {
	// Arrange: Create parent group with defaults
	router := NewRouter()
	api := router.NewRouteGroup("/api")
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
	assert.Equal(t, api.registry, v1.registry)
	assert.Equal(t, api.defaultTags, v1.defaultTags)
	assert.Equal(t, api.defaultRoles, v1.defaultRoles)
	assert.Equal(t, api.defaultSummary, v1.defaultSummary)
	assert.Equal(t, api.defaultDescription, v1.defaultDescription)
	assert.Equal(t, api.defaultAllowAnon, v1.defaultAllowAnon)
	assert.Equal(t, api.defaultDeprecated, v1.defaultDeprecated)
}

func TestShouldMaintainIndependentDefaultsWhenModifyingNestedRouteGroup(t *testing.T) {
	// Arrange: Create parent group with defaults
	router := NewRouter()
	api := router.NewRouteGroup("/api")
	api.WithTags("API")

	// Act: Create nested group and add different defaults
	v1 := api.NewRouteGroup("/v1")
	v1.WithTags("V1")

	// Assert: Parent defaults are preserved and child defaults are added
	assert.Equal(t, []string{"API"}, api.defaultTags)
	assert.Equal(t, []string{"API", "V1"}, v1.defaultTags)
}

func TestShouldSupportMultipleLevelsOfNestingWhenCreatingRouteGroups(t *testing.T) {
	// Arrange: Create multiple levels of nesting
	router := NewRouter()
	api := router.NewRouteGroup("/api")
	v1 := api.NewRouteGroup("/v1")
	users := v1.NewRouteGroup("/users")
	users.GET("/profile", func(c RouteContext) {
		c.OK("profile")
	})

	// Act: Make request to deeply nested route
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Route was registered with correct path and responds
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "profile")
}

func TestShouldNormalizePrefixesWhenCreatingNestedRouteGroups(t *testing.T) {
	// Arrange: Define test cases with various prefix formats
	router := NewRouter()
	testCases := []struct {
		parentPrefix string
		childPrefix  string
		expectedPath string
	}{
		{"/api", "/v1", "/api/v1"},
		{"/api/", "v1", "/api/v1"},
		{"api", "/v1", "/api/v1"},
		{"api/", "v1/", "/api/v1/"}, // Fixed: trailing slash should be preserved
	}

	// Act & Assert: Test each case
	for _, tc := range testCases {
		parent := router.NewRouteGroup(tc.parentPrefix)
		child := parent.NewRouteGroup(tc.childPrefix)

		assert.Equal(t, tc.expectedPath, child.prefix,
			"Parent: %s, Child: %s should result in: %s",
			tc.parentPrefix, tc.childPrefix, tc.expectedPath)
	}
}

func TestShouldInheritParametersWhenCreatingNestedRouteGroup(t *testing.T) {
	// Arrange: Create parent group with parameters
	router := NewRouter()
	api := router.NewRouteGroup("/api")
	api.WithPathParam("version", "v1").
		WithQueryParam("limit", 10).
		RequireRoles("admin")

	// Act: Create nested group and add route
	v1 := api.NewRouteGroup("/v1")
	v1.GET("/users", func(c RouteContext) {
		c.OK("users")
	})

	// Assert: Route inherits parameters and defaults
	options, _, found := router.registry.Load("/api/v1/users", http.MethodGet)
	require.True(t, found, "Route should be registered")
	assert.Contains(t, options.Roles, "admin")
	assert.Len(t, options.Operation.Parameters, 2) // version and limit params
}

func TestShouldWorkAsExpectedWhenUsingExampleFromIssue(t *testing.T) {
	// Arrange: Set up the exact example from the issue
	router := NewRouter()
	api := router.NewRouteGroup("/api")
	v1 := api.NewRouteGroup("/v1")
	v1.GET("/users", func(c RouteContext) {
		c.OK("users endpoint")
	})

	// Act: Make request to the route
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Route works as expected
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "users endpoint")
}
