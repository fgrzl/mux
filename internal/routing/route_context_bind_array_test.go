package routing

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldBindJSONArrayRootIntoSlice(t *testing.T) {
	// Arrange
	body := []byte(`[{"name":"a"},{"name":"b"}]`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	var out []smallBody
	// Act
	err := c.Bind(&out)
	// Assert
	assert.NoError(t, err)
	assert.Len(t, out, 2)
	assert.Equal(t, "a", out[0].Name)
	assert.Equal(t, "b", out[1].Name)
}

func TestShouldReturnExplicitErrorForJSONPrimitiveRoot(t *testing.T) {
	// Arrange
	body := []byte(`"hello"`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	var out string

	// Act
	err := c.Bind(&out)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported JSON root type")
}

func TestShouldReturnErrorForJSONArrayBodyWithAdditionalInputs(t *testing.T) {
	// Arrange
	body := []byte(`[{"name":"a"}]`)
	req := httptest.NewRequest(http.MethodPost, "/?limit=1", bytes.NewReader(body))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	var out []smallBody

	// Act
	err := c.Bind(&out)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot combine JSON array body")
}
