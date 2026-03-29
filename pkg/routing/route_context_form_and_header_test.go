package routing

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
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
	req.Header.Set(common.HeaderContentType, common.MimeFormURLEncoded)
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

func TestShouldBindFormURLEncodedWithCharset(t *testing.T) {
	// Arrange
	vals := url.Values{}
	vals.Add("name", "alice")
	vals.Add("tags", "a")
	vals.Add("tags", "b")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(vals.Encode()))
	req.Header.Set(common.HeaderContentType, common.MimeFormURLEncoded+"; charset=utf-8")
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

func TestShouldBindMultipartFormData(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	assert.NoError(t, writer.WriteField("name", "alice"))
	assert.NoError(t, writer.WriteField("tags", "a"))
	assert.NoError(t, writer.WriteField("tags", "b"))
	assert.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set(common.HeaderContentType, writer.FormDataContentType())
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
	req.Header.Set(common.HeaderXCorrelationID, "abc123")
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	// Provide RouteOptions with a declared header parameter and ParamIndex
	po := &openapi.ParameterObject{
		Name:   common.HeaderXCorrelationID,
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
	v, ok := staging[common.HeaderXCorrelationID].(string)
	assert.True(t, ok)
	assert.Equal(t, "abc123", v)
}
