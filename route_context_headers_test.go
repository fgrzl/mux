package mux

import (
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestShouldReturnHeaderValue(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	auth, ok := ctx.Header("Authorization")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, "Bearer token123", auth)
}

func TestShouldReturnFalseForMissingHeader(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.Header("Missing-Header")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnFalseForEmptyHeader(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Empty-Header", "")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.Header("Empty-Header")

	// Assert
	assert.False(t, ok)
}

func TestShouldParseValidHeaderInt(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Content-Length", "1024")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	length, ok := ctx.HeaderInt("Content-Length")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, 1024, length)
}

func TestShouldReturnFalseForInvalidHeaderInt(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Content-Length", "not-a-number")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.HeaderInt("Content-Length")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnFalseForMissingHeaderInt(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	value, ok := ctx.HeaderInt("Missing-Header")

	// Assert
	assert.False(t, ok)
	assert.Equal(t, 0, value)
}

func TestShouldParseValidHeaderUUID(t *testing.T) {
	// Arrange
	testUUID := uuid.New()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Request-ID", testUUID.String())
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	requestID, ok := ctx.HeaderUUID("Request-ID")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, testUUID, requestID)
}

func TestShouldReturnFalseForInvalidHeaderUUID(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Request-ID", "invalid-uuid")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.HeaderUUID("Request-ID")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnFalseForMissingHeaderUUID(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	value, ok := ctx.HeaderUUID("Missing-Header")

	// Assert
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, value)
}

func TestShouldParseValidHeaderBool(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Debug", "true")
	req.Header.Set("X-Cache", "false")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	debug, ok := ctx.HeaderBool("X-Debug")

	// Assert
	assert.True(t, ok)
	assert.True(t, debug)

	// Act
	cache, ok := ctx.HeaderBool("X-Cache")

	// Assert
	assert.True(t, ok)
	assert.False(t, cache)
}

func TestShouldReturnFalseForInvalidHeaderBool(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Debug", "maybe")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.HeaderBool("X-Debug")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnFalseForMissingHeaderBool(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	value, ok := ctx.HeaderBool("Missing-Header")

	// Assert
	assert.False(t, ok)
	assert.False(t, value)
}

func TestShouldParseValidHeaderFloat64(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Rating", "4.7")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	rating, ok := ctx.HeaderFloat64("X-Rating")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, 4.7, rating)
}

func TestShouldReturnFalseForInvalidHeaderFloat64(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Rating", "not-a-float")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.HeaderFloat64("X-Rating")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnFalseForMissingHeaderFloat64(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	value, ok := ctx.HeaderFloat64("Missing-Header")

	// Assert
	assert.False(t, ok)
	assert.Equal(t, float64(0), value)
}