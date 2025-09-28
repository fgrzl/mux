package routing

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/stretchr/testify/assert"
)

type address struct {
	City string `json:"city"`
	Tags []int  `json:"tags"`
}

type userExample struct {
	Address address `json:"address"`
}

func TestShouldCoerceShallowDeepObjectBySchema(t *testing.T) {
	// Arrange: user.city=Paris and user[tags]=1,2 (CSV)
	vals := url.Values{}
	vals.Set("user.city", "Paris")
	vals.Set("user[tags]", "1,2")

	req := httptest.NewRequest(http.MethodGet, "/?"+vals.Encode(), nil)
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	userSchema := &openapi.Schema{Type: "object", Properties: map[string]*openapi.Schema{
		"city": {Type: "string"},
		"tags": {Type: "array", Items: &openapi.Schema{Type: "integer"}},
	}}
	po := &openapi.ParameterObject{Name: "user", In: "query", Schema: userSchema}
	opts := &RouteOptions{}
	opts.Operation.Parameters = []*openapi.ParameterObject{po}
	opts.ParamIndex = BuildParamIndex([]*openapi.ParameterObject{po})
	c.SetOptions(opts)

	// Act
	staging := make(map[string]any)
	err := c.collectQueryParams(staging)

	// Assert
	assert.NoError(t, err)
	userMap, ok := staging["user"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "Paris", userMap["city"]) // parsed as string
	tags, ok := userMap["tags"].([]int64)
	assert.True(t, ok)
	assert.EqualValues(t, []int64{1, 2}, tags)
}

func TestShouldLeaveNestedDeepObjectLeavesRaw(t *testing.T) {
	// Arrange: user.address.city=Paris and user[address][tags]=1,2 (nested under first-level object property)
	vals := url.Values{}
	vals.Set("user.address.city", "Paris")
	vals.Set("user[address][tags]", "1,2")

	req := httptest.NewRequest(http.MethodGet, "/?"+vals.Encode(), nil)
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	userSchema := &openapi.Schema{Type: "object", Properties: map[string]*openapi.Schema{
		"address": {Type: "object", Properties: map[string]*openapi.Schema{
			"city": {Type: "string"},
			"tags": {Type: "array", Items: &openapi.Schema{Type: "integer"}},
		}},
	}}
	po := &openapi.ParameterObject{Name: "user", In: "query", Schema: userSchema}
	opts := &RouteOptions{}
	opts.Operation.Parameters = []*openapi.ParameterObject{po}
	opts.ParamIndex = BuildParamIndex([]*openapi.ParameterObject{po})
	c.SetOptions(opts)

	// Act
	staging := make(map[string]any)
	err := c.collectQueryParams(staging)

	// Assert
	assert.NoError(t, err)
	userMap, ok := staging["user"].(map[string]any)
	assert.True(t, ok)
	addr, ok := userMap["address"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "Paris", addr["city"]) // schema for first-level property "address" is object; leaf stays raw
	assert.Equal(t, "1,2", addr["tags"])   // not parsed into []int64 at nested level
}

func TestShouldNotCoerceNestedLeavesWithExampleOnly(t *testing.T) {
	// Arrange: With Example (no schema), nested leaves are detected but example-based parsing may not coerce leaf types
	vals := url.Values{}
	vals.Set("user.address.city", "Berlin")
	vals.Add("user.address.tags", "3")
	vals.Add("user.address.tags", "4")

	req := httptest.NewRequest(http.MethodGet, "/?"+vals.Encode(), nil)
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	ex := &userExample{Address: address{City: "", Tags: []int{}}}
	po := &openapi.ParameterObject{Name: "user", In: "query", Example: ex}
	opts := &RouteOptions{}
	opts.Operation.Parameters = []*openapi.ParameterObject{po}
	opts.ParamIndex = BuildParamIndex([]*openapi.ParameterObject{po})
	c.SetOptions(opts)

	// Act
	staging := make(map[string]any)
	err := c.collectQueryParams(staging)

	// Assert
	assert.NoError(t, err)
	userMap, ok := staging["user"].(map[string]any)
	assert.True(t, ok)
	addr, ok := userMap["address"].(map[string]any)
	assert.True(t, ok)
	// Current behavior: example-based deep parsing doesn't coerce nested leaves
	assert.Nil(t, addr["city"]) // value present but not parsed
	assert.Nil(t, addr["tags"]) // value present but not parsed
}
