package registry

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/internal/routing"
	"github.com/stretchr/testify/assert"
)

const (
	staticPath = "/static"
)

func TestShouldReturnAllowHeaderForStaticPath(t *testing.T) {
	// Arrange
	reg := NewRouteRegistry()
	reg.Register(staticPath, http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})
	reg.Register(staticPath, http.MethodPost, &routing.RouteOptions{Method: http.MethodPost})

	// Act
	allow, ok := reg.TryGetAllowHeader(staticPath)

	// Assert
	assert.True(t, ok)
	assert.Equal(t, "GET, POST", allow)
}

func TestShouldReturnAllowHeaderForParamPath(t *testing.T) {
	// Arrange
	reg := NewRouteRegistry()
	reg.Register("/users/{id}", http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})
	reg.Register("/users/{id}", http.MethodDelete, &routing.RouteOptions{Method: http.MethodDelete})

	// Act
	allow, ok := reg.TryGetAllowHeader("/users/123")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, "GET, DELETE", allow)
}

func TestShouldReturnNoMatchForMissingPath(t *testing.T) {
	// Arrange
	reg := NewRouteRegistry()
	reg.Register("/a", http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})

	// Act
	_, ok := reg.TryGetAllowHeader("/missing")

	// Assert
	assert.False(t, ok)
}
