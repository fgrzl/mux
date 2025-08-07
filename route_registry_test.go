package mux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	options := &RouteOptions{Handler: func(c *RouteContext) {}}

	// Act
	registry.Register("/users", "GET", options)

	// Assert
	loadedOptions, params, found := registry.Load("/users", "GET")
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Empty(t, params)
}

func TestShouldRegisterRouteWithParameters(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &RouteOptions{Handler: func(c *RouteContext) {}}

	// Act
	registry.Register("/users/{id}", "GET", options)

	// Assert
	loadedOptions, params, found := registry.Load("/users/123", "GET")
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Equal(t, "123", params["id"])
}

func TestShouldRegisterRouteWithMultipleParameters(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &RouteOptions{Handler: func(c *RouteContext) {}}

	// Act
	registry.Register("/users/{userId}/posts/{postId}", "GET", options)

	// Assert
	loadedOptions, params, found := registry.Load("/users/123/posts/456", "GET")
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Equal(t, "123", params["userId"])
	assert.Equal(t, "456", params["postId"])
}

func TestShouldRegisterRouteWithWildcard(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &RouteOptions{Handler: func(c *RouteContext) {}}

	// Act
	registry.Register("/static/*", "GET", options)

	// Assert
	loadedOptions, params, found := registry.Load("/static/anything", "GET")
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Empty(t, params)
}

func TestShouldRegisterRouteWithCatchAll(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &RouteOptions{Handler: func(c *RouteContext) {}}

	// Act
	registry.Register("/api/**", "GET", options)

	// Assert
	loadedOptions, params, found := registry.Load("/api/anything/else", "GET")
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Empty(t, params)
}

func TestShouldRegisterMultipleMethodsForSameRoute(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	getOptions := &RouteOptions{Handler: func(c *RouteContext) {}}
	postOptions := &RouteOptions{Handler: func(c *RouteContext) {}}

	// Act
	registry.Register("/users", "GET", getOptions)
	registry.Register("/users", "POST", postOptions)

	// Assert
	loadedGetOptions, _, found := registry.Load("/users", "GET")
	assert.True(t, found)
	assert.Equal(t, getOptions, loadedGetOptions)

	loadedPostOptions, _, found := registry.Load("/users", "POST")
	assert.True(t, found)
	assert.Equal(t, postOptions, loadedPostOptions)
}

func TestShouldReturnFalseForUnregisteredRoute(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()

	// Act
	_, _, found := registry.Load("/nonexistent", "GET")

	// Assert
	assert.False(t, found)
}

func TestShouldReturnFalseForUnregisteredMethod(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &RouteOptions{Handler: func(c *RouteContext) {}}
	registry.Register("/users", "GET", options)

	// Act
	_, _, found := registry.Load("/users", "POST")

	// Assert
	assert.False(t, found)
}

func TestShouldHandleRootRoute(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &RouteOptions{Handler: func(c *RouteContext) {}}

	// Act
	registry.Register("/", "GET", options)

	// Assert
	loadedOptions, params, found := registry.Load("/", "GET")
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Empty(t, params)
}

func TestShouldHandleEmptyRoute(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &RouteOptions{Handler: func(c *RouteContext) {}}

	// Act
	registry.Register("", "GET", options)

	// Assert
	loadedOptions, params, found := registry.Load("", "GET")
	assert.True(t, found)
	assert.NotNil(t, loadedOptions)
	assert.Equal(t, options, loadedOptions)
	assert.Empty(t, params)
}

func TestShouldMatchExactRouteBeforeParameterRoute(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	exactOptions := &RouteOptions{Handler: func(c *RouteContext) { /* exact */ }}
	paramOptions := &RouteOptions{Handler: func(c *RouteContext) { /* param */ }}

	// Act
	registry.Register("/users/{id}", "GET", paramOptions)
	registry.Register("/users/new", "GET", exactOptions)

	// Assert
	// Should match exact route first
	loadedOptions, params, found := registry.Load("/users/new", "GET")
	assert.True(t, found)
	assert.Equal(t, exactOptions, loadedOptions)
	assert.Empty(t, params)

	// Should still match parameter route for other paths
	loadedOptions, params, found = registry.Load("/users/123", "GET")
	assert.True(t, found)
	assert.Equal(t, paramOptions, loadedOptions)
	assert.Equal(t, "123", params["id"])
}

func TestShouldHandleNestedRoutes(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options1 := &RouteOptions{Handler: func(c *RouteContext) {}}
	options2 := &RouteOptions{Handler: func(c *RouteContext) {}}

	// Act
	registry.Register("/api/users", "GET", options1)
	registry.Register("/api/users/profile", "GET", options2)

	// Assert
	loadedOptions1, _, found := registry.Load("/api/users", "GET")
	assert.True(t, found)
	assert.Equal(t, options1, loadedOptions1)

	loadedOptions2, _, found := registry.Load("/api/users/profile", "GET")
	assert.True(t, found)
	assert.Equal(t, options2, loadedOptions2)
}

func TestShouldHandleParameterCleanupOnFailure(t *testing.T) {
	// Arrange
	registry := NewRouteRegistry()
	options := &RouteOptions{Handler: func(c *RouteContext) {}}
	registry.Register("/users/{id}/posts", "GET", options)

	// Act - try to load a path that partially matches
	_, params, found := registry.Load("/users/123/comments", "GET")

	// Assert
	assert.False(t, found)
	// Params should be clean (implementation detail, but good to verify)
	assert.Empty(t, params)
}