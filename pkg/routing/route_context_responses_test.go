package routing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestShouldReturnServerErrorWithProblemDetails(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.ServerError("Test Error", "Something went wrong")

	// Assert
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, common.MimeProblemJSON, recorder.Header().Get(common.HeaderContentType))
	assert.Contains(t, recorder.Body.String(), "Test Error")
	assert.Contains(t, recorder.Body.String(), "Something went wrong")
	assert.Contains(t, recorder.Body.String(), "\"status\":500")
}

func TestShouldReturnServerErrorWithDefaultTitle(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.ServerError("", "Something went wrong")

	// Assert
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Internal Server Error")
}

func TestShouldReturnBadRequestWithProblemDetails(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.BadRequest("Invalid Input", "The provided data is invalid")

	// Assert
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, common.MimeProblemJSON, recorder.Header().Get(common.HeaderContentType))
	assert.Contains(t, recorder.Body.String(), "Invalid Input")
	assert.Contains(t, recorder.Body.String(), "The provided data is invalid")
	assert.Contains(t, recorder.Body.String(), "\"status\":400")
}

func TestShouldReturnBadRequestWithDefaultTitle(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.BadRequest("", "Invalid data")

	// Assert
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Bad Request")
}

func TestShouldReturnConflictWithProblemDetails(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPut, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.Conflict("Resource Exists", "The resource already exists")

	// Assert
	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.Equal(t, common.MimeProblemJSON, recorder.Header().Get(common.HeaderContentType))
	assert.Contains(t, recorder.Body.String(), "Resource Exists")
	assert.Contains(t, recorder.Body.String(), "The resource already exists")
	assert.Contains(t, recorder.Body.String(), "\"status\":409")
}

func TestShouldReturnUnauthorized(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.Unauthorized()

	// Assert
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Header().Get(common.HeaderContentType), common.MimeTextPlain)
	assert.Contains(t, recorder.Body.String(), "Unauthorized")
}

func TestShouldReturnForbidden(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.Forbidden("Access denied")

	// Assert
	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Contains(t, recorder.Header().Get(common.HeaderContentType), common.MimeTextPlain)
	assert.Contains(t, recorder.Body.String(), "Access denied")
}

func TestShouldReturnNotFound(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.NotFound()

	// Assert: NotFound now returns a Problem Details JSON body
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, common.MimeProblemJSON, recorder.Header().Get(common.HeaderContentType))
	assert.Contains(t, recorder.Body.String(), "Not Found")
	assert.Contains(t, recorder.Body.String(), "\"status\":404")
}

func TestShouldReturnOKWithData(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	data := map[string]string{"message": "success"}

	// Act
	ctx.OK(data)

	// Assert
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, common.MimeJSON, recorder.Header().Get(common.HeaderContentType))
	assert.Contains(t, recorder.Body.String(), "\"message\":\"success\"")
}

func TestShouldReturnCreatedWithData(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	data := map[string]string{"id": "123", "name": "test"}

	// Act
	ctx.Created(data)

	// Assert
	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, common.MimeJSON, recorder.Header().Get(common.HeaderContentType))
	assert.Contains(t, recorder.Body.String(), "\"id\":\"123\"")
	assert.Contains(t, recorder.Body.String(), "\"name\":\"test\"")
}

func TestShouldReturnNoContent(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodDelete, "/test/123", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.NoContent()

	// Assert
	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Empty(t, recorder.Body.String())
}

func TestShouldReturnAcceptWithData(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	data := map[string]string{"status": "processing"}

	// Act
	ctx.Accept(data)

	// Assert
	assert.Equal(t, http.StatusAccepted, recorder.Code)
	assert.Equal(t, common.MimeJSON, recorder.Header().Get(common.HeaderContentType))
	assert.Contains(t, recorder.Body.String(), "\"status\":\"processing\"")
}
