package routing

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/stretchr/testify/assert"
)

type formModel struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func TestShouldBindFormURLEncoded(t *testing.T) {
	// Arrange
	vals := url.Values{}
	vals.Add("name", "alice")
	vals.Add("tags", "a")
	vals.Add("tags", "b")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	var out formModel
	// Act
	err := c.Bind(&out)
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "alice", out.Name)
	assert.ElementsMatch(t, []string{"a", "b"}, out.Tags)
}

func TestShouldCollectDeclaredHeaderParam(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-Id", "abc123")
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	// Provide RouteOptions with a declared header parameter and ParamIndex
	po := &openapi.ParameterObject{
		Name:   "X-Correlation-Id",
		In:     "header",
		Schema: &openapi.Schema{Type: "string"},
	}
	opts := &RouteOptions{}
	// set parameter on embedded Operation
	opts.Operation.Parameters = []*openapi.ParameterObject{po}
	opts.ParamIndex = BuildParamIndex([]*openapi.ParameterObject{po})
	c.SetOptions(opts)

	// Act
	staging := make(map[string]any)
	err := c.collectHeaderData(staging)

	// Assert
	assert.NoError(t, err)
	v, ok := staging["X-Correlation-Id"].(string)
	assert.True(t, ok)
	assert.Equal(t, "abc123", v)
}
