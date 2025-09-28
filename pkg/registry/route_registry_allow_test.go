package registry

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
	"github.com/stretchr/testify/assert"
)

func TestShouldReturnAllowHeaderForStaticPath(t *testing.T) {
	reg := NewRouteRegistry()
	reg.Register("/static", http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})
	reg.Register("/static", http.MethodPost, &routing.RouteOptions{Method: http.MethodPost})

	allow, ok := reg.TryGetAllowHeader("/static")
	assert.True(t, ok)
	assert.Equal(t, "GET, POST", allow)
}

func TestShouldReturnAllowHeaderForParamPath(t *testing.T) {
	reg := NewRouteRegistry()
	reg.Register("/users/{id}", http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})
	reg.Register("/users/{id}", http.MethodDelete, &routing.RouteOptions{Method: http.MethodDelete})

	allow, ok := reg.TryGetAllowHeader("/users/123")
	assert.True(t, ok)
	assert.Equal(t, "GET, DELETE", allow)
}

func TestShouldReturnNoMatchForMissingPath(t *testing.T) {
	reg := NewRouteRegistry()
	reg.Register("/a", http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})
	_, ok := reg.TryGetAllowHeader("/missing")
	assert.False(t, ok)
}
