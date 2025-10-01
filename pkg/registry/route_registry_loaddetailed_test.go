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

func TestLoadDetailedStaticFoundMethodOK(t *testing.T) {
	reg := NewRouteRegistry()
	reg.Register(static, http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})

	params := make(map[string]string)
	opt, res := reg.LoadDetailedInto(static, http.MethodGet, params)
	assert.NotNil(t, opt)
	assert.True(t, res.Found)
	assert.True(t, res.MethodOK)
	assert.Equal(t, "", res.Allow)
	assert.Empty(t, params)
}

func TestLoadDetailedStaticMethodNotAllowedAllowReturned(t *testing.T) {
	reg := NewRouteRegistry()
	reg.Register(static, http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})
	reg.Register(static, http.MethodPost, &routing.RouteOptions{Method: http.MethodPost})

	params := make(map[string]string)
	opt, res := reg.LoadDetailedInto(static, http.MethodPut, params)
	assert.Nil(t, opt)
	assert.True(t, res.Found)
	assert.False(t, res.MethodOK)
	assert.Equal(t, "GET, POST", res.Allow)
}

func TestLoadDetailedParamFoundMethodOK(t *testing.T) {
	reg := NewRouteRegistry()
	reg.Register(users, http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})

	params := make(map[string]string)
	opt, res := reg.LoadDetailedInto("/users/123", http.MethodGet, params)
	assert.NotNil(t, opt)
	assert.True(t, res.Found)
	assert.True(t, res.MethodOK)
	assert.Equal(t, map[string]string{"id": "123"}, params)
}

func TestLoadDetailedParamMethodNotAllowedAllowReturned(t *testing.T) {
	reg := NewRouteRegistry()
	reg.Register(users, http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})
	reg.Register(users, http.MethodDelete, &routing.RouteOptions{Method: http.MethodDelete})

	params := make(map[string]string)
	opt, res := reg.LoadDetailedInto("/users/123", http.MethodPost, params)
	assert.Nil(t, opt)
	assert.True(t, res.Found)
	assert.False(t, res.MethodOK)
	assert.Equal(t, "GET, DELETE", res.Allow)
}

func TestLoadDetailedNotFound(t *testing.T) {
	reg := NewRouteRegistry()
	reg.Register("/a", http.MethodGet, &routing.RouteOptions{Method: http.MethodGet})

	params := make(map[string]string)
	opt, res := reg.LoadDetailedInto("/missing", http.MethodGet, params)
	assert.Nil(t, opt)
	assert.False(t, res.Found)
	assert.False(t, res.MethodOK)
	assert.Equal(t, "", res.Allow)
}
