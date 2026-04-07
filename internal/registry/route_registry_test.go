package registry

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/internal/routing"
	"github.com/stretchr/testify/assert"
)

func loadRoute(registry *RouteRegistry, path, method string) (*routing.RouteOptions, routing.Params, bool) {
	var params routing.Params
	loadedOptions, found := registry.LoadIntoSlice(path, method, &params)
	return loadedOptions, params, found
}

func TestShouldCreateNewRouteRegistry(t *testing.T) {
	// Arrange & Act
	registry := NewRouteRegistry()

	// Assert
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.root)
	assert.NotNil(t, registry.root.Children)
}

func TestShouldRegisterSimpleRoute(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}

	// Act
	registry.Register("/users", http.MethodGet, options)

	// Assert
	loadedOptions, params, found := loadRoute(registry, "/users", http.MethodGet)
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Empty(t, params)
}

func TestShouldRegisterRouteWithParameters(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}

	// Act
	registry.Register("/users/{id}", http.MethodGet, options)

	// Assert
	loadedOptions, params, found := loadRoute(registry, "/users/123", http.MethodGet)
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Equal(t, "123", params.Get("id"))
}

func TestShouldRegisterRouteWithMultipleParameters(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}

	// Act
	registry.Register("/users/{userId}/posts/{postId}", http.MethodGet, options)

	// Assert
	loadedOptions, params, found := loadRoute(registry, "/users/123/posts/456", http.MethodGet)
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Equal(t, "123", params.Get("userId"))
	assert.Equal(t, "456", params.Get("postId"))
}

func TestShouldRegisterRouteWithWildcard(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}

	// Act
	registry.Register("/static/*", http.MethodGet, options)

	// Assert
	loadedOptions, params, found := loadRoute(registry, "/static/anything", http.MethodGet)
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Empty(t, params)
}

func TestShouldRegisterRouteWithCatchAll(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}

	// Act
	registry.Register("/api/**", http.MethodGet, options)

	// Assert
	loadedOptions, params, found := loadRoute(registry, "/api/anything/else", http.MethodGet)
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Empty(t, params)
}

func TestShouldRegisterMultipleMethodsForSameRoute(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	getOptions := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}
	postOptions := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}

	// Act
	registry.Register("/users", http.MethodGet, getOptions)
	registry.Register("/users", http.MethodPost, postOptions)

	// Assert
	loadedGetOptions, _, found := loadRoute(registry, "/users", http.MethodGet)
	assert.True(t, found)
	assert.Equal(t, getOptions, loadedGetOptions)

	loadedPostOptions, _, found := loadRoute(registry, "/users", http.MethodPost)
	assert.True(t, found)
	assert.Equal(t, postOptions, loadedPostOptions)
}

func TestShouldReturnFalseForUnregisteredRoute(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()

	// Act
	_, _, found := loadRoute(registry, "/nonexistent", http.MethodGet)

	// Assert
	assert.False(t, found)
}

func TestShouldReturnFalseForUnregisteredMethod(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}
	registry.Register("/users", http.MethodGet, options)

	// Act
	_, _, found := loadRoute(registry, "/users", http.MethodPost)

	// Assert
	assert.False(t, found)
}

func TestShouldHandleRootRoute(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}

	// Act
	registry.Register("/", http.MethodGet, options)

	// Assert
	loadedOptions, params, found := loadRoute(registry, "/", http.MethodGet)
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Empty(t, params)
}

func TestShouldHandleEmptyRoute(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}

	// Act
	registry.Register("", http.MethodGet, options)

	// Assert
	loadedOptions, params, found := loadRoute(registry, "", http.MethodGet)
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Empty(t, params)
}

func TestShouldMatchExactRouteBeforeParameterRoute(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	exactOptions := &routing.RouteOptions{Handler: func(c routing.RouteContext) { /* exact */ }}
	paramOptions := &routing.RouteOptions{Handler: func(c routing.RouteContext) { /* param */ }}

	// Act
	registry.Register("/users/{id}", http.MethodGet, paramOptions)
	registry.Register("/users/new", http.MethodGet, exactOptions)

	// Assert
	// Should match exact route first
	loadedOptions, params, found := loadRoute(registry, "/users/new", http.MethodGet)
	assert.True(t, found)
	assert.Equal(t, exactOptions, loadedOptions)
	assert.Empty(t, params)

	// Should still match parameter route for other paths
	loadedOptions, params, found = loadRoute(registry, "/users/123", http.MethodGet)
	assert.True(t, found)
	assert.Equal(t, paramOptions, loadedOptions)
	assert.Equal(t, "123", params.Get("id"))
}

func TestShouldHandleNestedRoutes(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options1 := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}
	options2 := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}

	// Act
	registry.Register("/api/users", http.MethodGet, options1)
	registry.Register("/api/users/profile", http.MethodGet, options2)

	// Assert
	loadedOptions1, _, found := loadRoute(registry, "/api/users", http.MethodGet)
	assert.True(t, found)
	assert.Equal(t, options1, loadedOptions1)

	loadedOptions2, _, found := loadRoute(registry, "/api/users/profile", http.MethodGet)
	assert.True(t, found)
	assert.Equal(t, options2, loadedOptions2)
}

func TestShouldHandleParameterCleanupOnFailure(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}
	registry.Register("/users/{id}/posts", http.MethodGet, options)

	// Act - try to load a path that partially matches
	_, params, found := loadRoute(registry, "/users/123/comments", http.MethodGet)

	// Assert
	assert.False(t, found)
	// Params should be clean (implementation detail, but good to verify)
	assert.Empty(t, params)
}
