package registry

import (
	"testing"

	"github.com/fgrzl/mux/internal/routing"
	"github.com/stretchr/testify/assert"
)

func TestShouldReturnMethodsForCatchAllSubpath(t *testing.T) {
	// Arrange
	r := NewRouteRegistry()
	optsGet := &routing.RouteOptions{Method: "GET"}
	optsPost := &routing.RouteOptions{Method: "POST"}
	r.Register("/files/**", "GET", optsGet)
	r.Register("/files/**", "POST", optsPost)

	// Act
	methods, ok := r.TryMatchMethods("/files/a/b/c.txt")

	// Assert
	assert.True(t, ok)
	assert.ElementsMatch(t, []string{"GET", "POST"}, methods)
}

func TestShouldNotReturnMethodsForCatchAllBasePath(t *testing.T) {
	// Arrange
	r := NewRouteRegistry()
	optsGet := &routing.RouteOptions{Method: "GET"}
	r.Register("/files/**", "GET", optsGet)

	// Act
	methods, ok := r.TryMatchMethods("/files/")

	// Assert
	assert.False(t, ok)
	assert.Nil(t, methods)
}
