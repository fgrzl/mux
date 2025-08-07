package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestShouldReturnParamValue(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{"id": "123", "name": "test"}

	// Act
	id, ok := ctx.Param("id")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, "123", id)
}

func TestShouldReturnFalseForMissingParam(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{}

	// Act
	_, ok := ctx.Param("nonexistent")

	// Assert
	assert.False(t, ok)
}

func TestShouldParseValidUUIDParam(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	testUUID := uuid.New()
	ctx.Params = RouteParams{"id": testUUID.String()}

	// Act
	parsedUUID, ok := ctx.ParamUUID("id")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, testUUID, parsedUUID)
}

func TestShouldReturnFalseForInvalidUUIDParam(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{"id": "invalid-uuid"}

	// Act
	_, ok := ctx.ParamUUID("id")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnFalseForMissingUUIDParam(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{}

	// Act
	parsedUUID, ok := ctx.ParamUUID("missing")

	// Assert
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, parsedUUID)
}

func TestShouldParseValidIntParam(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{"count": "42"}

	// Act
	count, ok := ctx.ParamInt("count")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, 42, count)
}

func TestShouldReturnFalseForInvalidIntParam(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{"count": "not-a-number"}

	// Act
	_, ok := ctx.ParamInt("count")

	// Assert
	assert.False(t, ok)
}

func TestShouldParseValidInt16Param(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{"port": "8080"}

	// Act
	port, ok := ctx.ParamInt16("port")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, int16(8080), port)
}

func TestShouldReturnFalseForInvalidInt16Param(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{"port": "99999999"} // Too large for int16

	// Act
	_, ok := ctx.ParamInt16("port")

	// Assert
	assert.False(t, ok)
}

func TestShouldParseValidInt32Param(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{"size": "1048576"}

	// Act
	size, ok := ctx.ParamInt32("size")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, int32(1048576), size)
}

func TestShouldReturnFalseForInvalidInt32Param(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{"size": "not-a-number"}

	// Act
	_, ok := ctx.ParamInt32("size")

	// Assert
	assert.False(t, ok)
}

func TestShouldParseValidInt64Param(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{"timestamp": "1609459200"}

	// Act
	timestamp, ok := ctx.ParamInt64("timestamp")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, int64(1609459200), timestamp)
}

func TestShouldReturnFalseForInvalidInt64Param(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{"timestamp": "invalid"}

	// Act
	_, ok := ctx.ParamInt64("timestamp")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnZeroForMissingIntParams(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	ctx.Params = RouteParams{}

	// Act & Assert
	intVal, ok := ctx.ParamInt("missing")
	assert.False(t, ok)
	assert.Equal(t, 0, intVal)

	int16Val, ok := ctx.ParamInt16("missing")
	assert.False(t, ok)
	assert.Equal(t, int16(0), int16Val)

	int32Val, ok := ctx.ParamInt32("missing")
	assert.False(t, ok)
	assert.Equal(t, int32(0), int32Val)

	int64Val, ok := ctx.ParamInt64("missing")
	assert.False(t, ok)
	assert.Equal(t, int64(0), int64Val)
}
