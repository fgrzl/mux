package routing

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestShouldCoerceNestedDeepObjectLeavesBySchema(t *testing.T) {
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
	assert.Equal(t, "Paris", addr["city"])
	tags, ok := addr["tags"].([]int64)
	assert.True(t, ok)
	assert.EqualValues(t, []int64{1, 2}, tags)
}

func TestShouldCoerceNestedLeavesWithExampleOnly(t *testing.T) {
	// Arrange
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
	assert.Equal(t, "Berlin", addr["city"])
	tags, ok := addr["tags"].([]int)
	assert.True(t, ok)
	assert.Equal(t, []int{3, 4}, tags)
}

func TestShouldMergeNestedObjectAcrossQueryAndJSONBody(t *testing.T) {
	// Arrange
	vals := url.Values{}
	vals.Set("user.address.city", "Paris")
	req := httptest.NewRequest(http.MethodPost, "/?"+vals.Encode(), strings.NewReader(`{"user":{"address":{"postal":"75001"}}}`))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	ex := struct {
		User struct {
			Address struct {
				City   string `json:"city"`
				Postal string `json:"postal"`
			} `json:"address"`
		} `json:"user"`
	}{}
	opts := Route(http.MethodPost, "/").WithQueryParam("user", ex.User)
	c.SetOptions(opts)

	var out struct {
		User struct {
			Address struct {
				City   string `json:"city"`
				Postal string `json:"postal"`
			} `json:"address"`
		} `json:"user"`
	}

	// Act
	err := c.Bind(&out)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Paris", out.User.Address.City)
	assert.Equal(t, "75001", out.User.Address.Postal)
}
