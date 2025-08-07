package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestShouldReturnQueryValue(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?name=john&age=25", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	name, ok := ctx.QueryValue("name")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, "john", name)
}

func TestShouldReturnFalseForMissingQueryValue(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.QueryValue("missing")

	// Assert
	assert.False(t, ok)
}

func TestShouldReturnFirstQueryValueWhenMultiple(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?tag=first&tag=second", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	tag, ok := ctx.QueryValue("tag")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, "first", tag)
}

func TestShouldReturnAllQueryValues(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?tag=first&tag=second&tag=third", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	tags, ok := ctx.QueryValues("tag")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, []string{"first", "second", "third"}, tags)
}

func TestShouldReturnFalseForMissingQueryValues(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.QueryValues("missing")

	// Assert
	assert.False(t, ok)
}

func TestShouldParseValidQueryUUID(t *testing.T) {
	// Arrange
	testUUID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/test?id="+testUUID.String(), nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	parsedUUID, ok := ctx.QueryUUID("id")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, testUUID, parsedUUID)
}

func TestShouldReturnFalseForInvalidQueryUUID(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?id=invalid-uuid", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.QueryUUID("id")

	// Assert
	assert.False(t, ok)
}

func TestShouldParseMultipleQueryUUIDs(t *testing.T) {
	// Arrange
	uuid1 := uuid.New()
	uuid2 := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/test?ids="+uuid1.String()+"&ids="+uuid2.String(), nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	uuids, ok := ctx.QueryUUIDs("ids")

	// Assert
	assert.True(t, ok)
	assert.Len(t, uuids, 2)
	assert.Contains(t, uuids, uuid1)
	assert.Contains(t, uuids, uuid2)
}

func TestShouldReturnFalseForMissingQueryUUIDs(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.QueryUUIDs("missing")

	// Assert
	assert.False(t, ok)
}

func TestShouldParseValidQueryInt(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?page=5&limit=10", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	page, ok := ctx.QueryInt("page")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, 5, page)
}

func TestShouldReturnFalseForInvalidQueryInt(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?page=invalid", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.QueryInt("page")

	// Assert
	assert.False(t, ok)
}

func TestShouldParseMultipleQueryInts(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?numbers=1&numbers=2&numbers=3", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	numbers, ok := ctx.QueryInts("numbers")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, []int{1, 2, 3}, numbers)
}

func TestShouldParseValidQueryBool(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?active=true&deleted=false", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	active, ok := ctx.QueryBool("active")

	// Assert
	assert.True(t, ok)
	assert.True(t, active)
}

func TestShouldReturnFalseForInvalidQueryBool(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?active=maybe", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	_, ok := ctx.QueryBool("active")

	// Assert
	assert.False(t, ok)
}

func TestShouldParseQueryFloat32(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?price=19.99", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	price, ok := ctx.QueryFloat32("price")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, float32(19.99), price)
}

func TestShouldParseQueryFloat64(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?percentage=85.7653", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	percentage, ok := ctx.QueryFloat64("percentage")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, 85.7653, percentage)
}

func TestShouldReturnZeroForMissingQueryParameters(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act & Assert
	intVal, ok := ctx.QueryInt("missing")
	assert.False(t, ok)
	assert.Equal(t, 0, intVal)

	boolVal, ok := ctx.QueryBool("missing")
	assert.False(t, ok)
	assert.False(t, boolVal)

	float32Val, ok := ctx.QueryFloat32("missing")
	assert.False(t, ok)
	assert.Equal(t, float32(0), float32Val)

	float64Val, ok := ctx.QueryFloat64("missing")
	assert.False(t, ok)
	assert.Equal(t, float64(0), float64Val)

	uuidVal, ok := ctx.QueryUUID("missing")
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, uuidVal)
}
