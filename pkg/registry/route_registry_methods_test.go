package registry

import (
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
	"github.com/stretchr/testify/assert"
)

func TestShouldReturnMethodsForStaticPath(t *testing.T) {
	// Arrange
	reg := NewRouteRegistry()
	opts := &routing.RouteOptions{}
	reg.Register("/static", "GET", opts)
	reg.Register("/static", "POST", opts)

	// Act
	methods, ok := reg.TryMatchMethods("/static")

	// Assert
	assert.True(t, ok)
	assert.ElementsMatch(t, []string{"GET", "POST"}, methods)
}

func TestShouldReturnMethodsForParamPath(t *testing.T) {
	// Arrange
	reg := NewRouteRegistry()
	opts := &routing.RouteOptions{}
	reg.Register("/users/{id}", "GET", opts)
	reg.Register("/users/{id}", "PUT", opts)

	// Act
	methods, ok := reg.TryMatchMethods("/users/123")

	// Assert
	assert.True(t, ok)
	assert.ElementsMatch(t, []string{"GET", "PUT"}, methods)
}

func TestShouldReturnFalseWhenNoPathMatches(t *testing.T) {
	// Arrange
	reg := NewRouteRegistry()

	// Act
	methods, ok := reg.TryMatchMethods("/missing")

	// Assert
	assert.False(t, ok)
	assert.Nil(t, methods)
}
