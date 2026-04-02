package router

import (
	"testing"

	openapi "github.com/fgrzl/mux/internal/openapi"
	"github.com/fgrzl/mux/internal/routing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type routerSnapshotNested struct {
	Value string `json:"value"`
}

type routerSnapshotPayload struct {
	Name   *string               `json:"name"`
	Nested *routerSnapshotNested `json:"nested"`
	Tags   []string              `json:"tags"`
}

func TestRoutesShouldReturnIndependentCallbackAndServerSnapshots(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	name := "stable"
	payload := &routerSnapshotPayload{
		Name:   &name,
		Nested: &routerSnapshotNested{Value: "root"},
		Tags:   []string{"one"},
	}
	rb := rtr.GET("/users", func(c routing.RouteContext) {
		c.OK("users")
	}).WithOperationID("getUsers")
	rb.Options.Callbacks = map[string]*openapi.PathItem{
		"onData": {
			Post: &openapi.Operation{
				OperationID: "notifyUsers",
				Parameters: []*openapi.ParameterObject{{
					Name:    "payload",
					In:      "query",
					Schema:  &openapi.Schema{Type: "string"},
					Example: payload,
				}},
				Responses: map[string]*openapi.ResponseObject{"200": {Description: "OK"}},
				Extensions: map[string]any{
					"x-callback": map[string]any{"mode": "async"},
				},
			},
		},
	}
	rb.Options.Servers = []*openapi.ServerObject{{
		URL: "https://api.example.com/{version}",
		Variables: map[string]*openapi.ServerVariable{
			"version": {Default: "v1", Enum: []string{"v1"}},
		},
	}}
	rb.Options.Extensions = map[string]any{
		"x-op": map[string]any{"trace": true},
	}

	// Act
	routes, err := rtr.Routes()
	require.NoError(t, err)
	require.Len(t, routes, 1)
	returned := routes[0].Options
	returned.Callbacks["onData"].Post.Extensions["x-callback"].(map[string]any)["mode"] = "sync"
	returnedPayload := returned.Callbacks["onData"].Post.Parameters[0].Example.(*routerSnapshotPayload)
	*returnedPayload.Name = "changed"
	returnedPayload.Nested.Value = "mutated"
	returnedPayload.Tags[0] = "two"
	returned.Servers[0].Variables["version"].Default = "v2"
	returned.Extensions["x-op"].(map[string]any)["trace"] = false
	freshRoutes, err := rtr.Routes()

	// Assert
	require.NoError(t, err)
	require.Len(t, freshRoutes, 1)
	fresh := freshRoutes[0].Options
	assert.Equal(t, "async", rb.Options.Callbacks["onData"].Post.Extensions["x-callback"].(map[string]any)["mode"])
	originalPayload := rb.Options.Callbacks["onData"].Post.Parameters[0].Example.(*routerSnapshotPayload)
	assert.Equal(t, "stable", *originalPayload.Name)
	assert.Equal(t, "root", originalPayload.Nested.Value)
	assert.Equal(t, []string{"one"}, originalPayload.Tags)
	assert.Equal(t, "v1", rb.Options.Servers[0].Variables["version"].Default)
	assert.Equal(t, true, rb.Options.Extensions["x-op"].(map[string]any)["trace"])
	assert.Equal(t, "async", fresh.Callbacks["onData"].Post.Extensions["x-callback"].(map[string]any)["mode"])
	freshPayload := fresh.Callbacks["onData"].Post.Parameters[0].Example.(*routerSnapshotPayload)
	assert.Equal(t, "stable", *freshPayload.Name)
	assert.Equal(t, "root", freshPayload.Nested.Value)
	assert.Equal(t, []string{"one"}, freshPayload.Tags)
	assert.Equal(t, "v1", fresh.Servers[0].Variables["version"].Default)
	assert.Equal(t, true, fresh.Extensions["x-op"].(map[string]any)["trace"])
	assert.NotSame(t, returned, fresh)
	assert.NotSame(t, returned.Callbacks["onData"], fresh.Callbacks["onData"])
	assert.NotSame(t, returned.Servers[0], fresh.Servers[0])
}
