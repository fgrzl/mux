package routing

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type smallBody struct {
	Name string `json:"name"`
}

func TestShouldBindJSONWithinDefaultLimit(t *testing.T) {
	// Arrange: Default limit is 1MB; this body should succeed without setting MaxBodyBytes
	body := []byte(`{"name":"ok"}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	var out smallBody
	// Act
	err := c.Bind(&out)
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "ok", out.Name)
}

func TestShouldFailBindJSONOverCustomLimit(t *testing.T) {
	// Arrange: Set a very small limit to trigger MaxBytesReader error; body is larger than 8 bytes
	body := []byte(`{"name":"toolong"}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c := NewRouteContext(rr, req)
	c.SetMaxBodyBytes(8)
	var out smallBody
	// Act
	err := c.Bind(&out)
	// Assert
	assert.Error(t, err)
}
