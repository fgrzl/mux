package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNestedRouteGroups(t *testing.T) {
	// Create a router
	router := NewRouter()
	
	// Create a nested group structure: /api/v1
	api := router.NewRouteGroup("/api")
	v1 := api.NewRouteGroup("/v1")
	
	// Add a route to the nested group
	v1.GET("/users", func(c *RouteContext) {
		c.OK("users")
	})
	
	// Test that the route was registered with the correct path
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "users")
}

func TestNestedRouteGroupsInheritDefaults(t *testing.T) {
	router := NewRouter()
	
	// Create parent group with defaults
	api := router.NewRouteGroup("/api")
	api.WithTags("API").
		RequireRoles("user").
		WithSummary("API Summary").
		WithDescription("API Description").
		AllowAnonymous().
		Deprecated()
	
	// Create nested group
	v1 := api.NewRouteGroup("/v1")
	
	// Check that defaults are inherited
	assert.Equal(t, "/api/v1", v1.prefix)
	assert.Equal(t, api.authProvider, v1.authProvider)
	assert.Equal(t, api.registry, v1.registry)
	assert.Equal(t, api.defaultTags, v1.defaultTags)
	assert.Equal(t, api.defaultRoles, v1.defaultRoles)
	assert.Equal(t, api.defaultSummary, v1.defaultSummary)
	assert.Equal(t, api.defaultDescription, v1.defaultDescription)
	assert.Equal(t, api.defaultAllowAnon, v1.defaultAllowAnon)
	assert.Equal(t, api.defaultDeprecated, v1.defaultDeprecated)
}

func TestNestedRouteGroupsIndependentDefaults(t *testing.T) {
	router := NewRouter()
	
	// Create parent group with defaults
	api := router.NewRouteGroup("/api")
	api.WithTags("API")
	
	// Create nested group and add different defaults
	v1 := api.NewRouteGroup("/v1")
	v1.WithTags("V1")
	
	// Check that parent defaults are preserved and child defaults are added
	assert.Equal(t, []string{"API"}, api.defaultTags)
	assert.Equal(t, []string{"API", "V1"}, v1.defaultTags)
}

func TestMultipleLevelsOfNesting(t *testing.T) {
	router := NewRouter()
	
	// Create multiple levels: /api/v1/users
	api := router.NewRouteGroup("/api")
	v1 := api.NewRouteGroup("/v1")
	users := v1.NewRouteGroup("/users")
	
	// Add a route to the deeply nested group
	users.GET("/profile", func(c *RouteContext) {
		c.OK("profile")
	})
	
	// Test that the route was registered with the correct path
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "profile")
}

func TestNestedGroupsPrefixNormalization(t *testing.T) {
	router := NewRouter()
	
	// Test various prefix formats
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
	
	for _, tc := range testCases {
		parent := router.NewRouteGroup(tc.parentPrefix)
		child := parent.NewRouteGroup(tc.childPrefix)
		
		assert.Equal(t, tc.expectedPath, child.prefix, 
			"Parent: %s, Child: %s should result in: %s", 
			tc.parentPrefix, tc.childPrefix, tc.expectedPath)
	}
}

func TestNestedGroupsWithParameters(t *testing.T) {
	router := NewRouter()
	
	// Create parent group with parameters
	api := router.NewRouteGroup("/api")
	api.WithPathParam("version", "v1").
		WithQueryParam("limit", 10).
		RequireRoles("admin")
	
	// Create nested group
	v1 := api.NewRouteGroup("/v1")
	
	// Add route to nested group
	v1.GET("/users", func(c *RouteContext) {
		c.OK("users")
	})
	
	// Verify the route inherits parameters and defaults
	options, _, found := router.registry.Load("/api/v1/users", http.MethodGet)
	require.True(t, found, "Route should be registered")
	
	// Check inherited defaults
	assert.Contains(t, options.Roles, "admin")
	assert.Len(t, options.Operation.Parameters, 2) // version and limit params
}

func TestExampleFromIssue(t *testing.T) {
	// Test the exact example from the issue
	router := NewRouter()
	
	api := router.NewRouteGroup("/api")
	v1 := api.NewRouteGroup("/v1")
	v1.GET("/users", func(c *RouteContext) {
		c.OK("users endpoint")
	})
	
	// Test the resulting route
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "users endpoint")
}