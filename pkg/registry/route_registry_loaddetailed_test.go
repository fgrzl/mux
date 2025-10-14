package registry

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
	"github.com/stretchr/testify/assert"
)

const (
	static = "/static"
	users  = "/users/{id}"
)

func TestShouldReturnRouteAndOKStatusGivenStaticRouteWithMatchingMethod(t *testing.T) {
	// Arrange
	reg := NewRouteRegistry()
	reg.Register(static, http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})

	params := make(map[string]string)

	// Act
	opt, res := reg.LoadDetailedInto(static, http.MethodGet, params)

	// Assert
	assert.NotNil(t, opt)
	assert.True(t, res.Found)
	assert.True(t, res.MethodOK)
	assert.Equal(t, "", res.Allow)
	assert.Empty(t, params)
}

func TestShouldReturnAllowHeaderGivenStaticRouteWithWrongMethod(t *testing.T) {
	// Arrange
	reg := NewRouteRegistry()
	reg.Register(static, http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})
	reg.Register(static, http.MethodPost, &routing.RouteOptions{Method: http.MethodPost})

	params := make(map[string]string)

	// Act
	opt, res := reg.LoadDetailedInto(static, http.MethodPut, params)

	// Assert
	assert.Nil(t, opt)
	assert.True(t, res.Found)
	assert.False(t, res.MethodOK)
	assert.Equal(t, "GET, POST", res.Allow)
}

func TestShouldReturnRouteAndExtractParamsGivenParameterizedRouteWithMatchingMethod(t *testing.T) {
	// Arrange
	reg := NewRouteRegistry()
	reg.Register(users, http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})

	params := make(map[string]string)

	// Act
	opt, res := reg.LoadDetailedInto("/users/123", http.MethodGet, params)

	// Assert
	assert.NotNil(t, opt)
	assert.True(t, res.Found)
	assert.True(t, res.MethodOK)
	assert.Equal(t, map[string]string{"id": "123"}, params)
}

func TestShouldReturnAllowHeaderGivenParameterizedRouteWithWrongMethod(t *testing.T) {
	// Arrange
	reg := NewRouteRegistry()
	reg.Register(users, http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})
	reg.Register(users, http.MethodDelete, &routing.RouteOptions{Method: http.MethodDelete})

	params := make(map[string]string)

	// Act
	opt, res := reg.LoadDetailedInto("/users/123", http.MethodPost, params)

	// Assert
	assert.Nil(t, opt)
	assert.True(t, res.Found)
	assert.False(t, res.MethodOK)
	assert.Equal(t, "GET, DELETE", res.Allow)
}

func TestShouldReturnNotFoundGivenNoMatchingRoute(t *testing.T) {
	// Arrange
	reg := NewRouteRegistry()
	reg.Register("/a", http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})

	params := make(map[string]string)

	// Act
	opt, res := reg.LoadDetailedInto("/missing", http.MethodGet, params)

	// Assert
	assert.Nil(t, opt)
	assert.False(t, res.Found)
	assert.False(t, res.MethodOK)
	assert.Equal(t, "", res.Allow)
}
