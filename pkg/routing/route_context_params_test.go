package routing

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
	params := &Params{}
	params.Set("id", "123")
	params.Set("name", "test")
	ctx.paramsSlice = params

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
	params := &Params{}
	ctx.paramsSlice = params

	// Act
	_, ok := ctx.Param("missing")

	// Assert
	assert.False(t, ok)
}

func TestShouldParseValidUUIDParam(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	testUUID := uuid.New()
	params := &Params{}
	params.Set("id", testUUID.String())
	ctx.paramsSlice = params

	// Act
	parsedUUID, ok := ctx.ParamUUID("id")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, testUUID, parsedUUID)
}

func TestShouldReturnFalseForInvalidUUID(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	params := &Params{}
	params.Set("id", "invalid-uuid")
	ctx.paramsSlice = params

	// Act
	_, ok := ctx.ParamUUID("id")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnFalseForMissingUUID(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	params := &Params{}
	ctx.paramsSlice = params

	// Act
	result, ok := ctx.ParamUUID("id")

	// Assert
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, result)
}

func TestShouldReturnIntParam(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	params := &Params{}
	params.Set("count", "42")
	ctx.paramsSlice = params

	// Act
	result, ok := ctx.ParamInt("count")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, 42, result)
}

func TestShouldReturnFalseForInvalidInt(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	params := &Params{}
	params.Set("count", "not-a-number")
	ctx.paramsSlice = params

	// Act
	_, ok := ctx.ParamInt("count")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnInt16Param(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	params := &Params{}
	params.Set("port", "8080")
	ctx.paramsSlice = params

	// Act
	result, ok := ctx.ParamInt16("port")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, int16(8080), result)
}

func TestShouldReturnFalseForInt16Overflow(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	params := &Params{}
	params.Set("port", "99999999") // Too large for int16
	ctx.paramsSlice = params

	// Act
	_, ok := ctx.ParamInt16("port")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnInt32Param(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	params := &Params{}
	params.Set("size", "1048576")
	ctx.paramsSlice = params

	// Act
	result, ok := ctx.ParamInt32("size")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, int32(1048576), result)
}

func TestShouldReturnFalseForInvalidInt32(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	params := &Params{}
	params.Set("size", "not-a-number")
	ctx.paramsSlice = params

	// Act
	_, ok := ctx.ParamInt32("size")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnInt64Param(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	params := &Params{}
	params.Set("timestamp", "1609459200")
	ctx.paramsSlice = params

	// Act
	result, ok := ctx.ParamInt64("timestamp")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, int64(1609459200), result)
}

func TestShouldReturnFalseForInvalidInt64(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)
	params := &Params{}
	params.Set("timestamp", "invalid")
	ctx.paramsSlice = params

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
	params := &Params{}
	ctx.paramsSlice = params

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
