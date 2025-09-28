package routing

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldBindJSONArrayRootIntoSlice(t *testing.T) {
	// Arrange
	body := []byte(`[{"name":"a"},{"name":"b"}]`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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
