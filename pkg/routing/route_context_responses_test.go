package routing

import (
	"bytes"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/stretchr/testify/assert"
)

func newRoutingTestLogger(minLevel slog.Level) (*bytes.Buffer, *slog.Logger) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{Level: minLevel}))
	return &logBuffer, logger
}

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

func TestConflictShouldNotEmitRoutineLog(t *testing.T) {
	// Arrange
	logBuffer, logger := newRoutingTestLogger(slog.LevelDebug)
	slog.SetDefault(logger)

	req := httptest.NewRequest(http.MethodPut, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.Conflict("Resource Exists", "The resource already exists")

	// Assert
	assert.Empty(t, logBuffer.String())
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

func TestShouldReturnInternalServerErrorWhenJSONEncodingFails(t *testing.T) {
	// Arrange
	logBuffer, logger := newRoutingTestLogger(slog.LevelError)
	slog.SetDefault(logger)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	data := map[string]float64{"value": math.NaN()}

	// Act
	ctx.OK(data)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.NotEmpty(t, recorder.Body.String())
	assert.Contains(t, recorder.Body.String(), "Internal Server Error")
	assert.NotEqual(t, common.MimeJSON, recorder.Header().Get(common.HeaderContentType))
	assert.Contains(t, logBuffer.String(), "failed to marshal json response")
	assert.Contains(t, logBuffer.String(), "status=200")
	assert.Contains(t, logBuffer.String(), "response_type=json")
	assert.Contains(t, logBuffer.String(), "method=GET")
	assert.Contains(t, logBuffer.String(), "path=/test")
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

func TestShouldReturn301MovedPermanently(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "http://example.com/old", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.MovedPermanently("/new")

	// Assert
	assert.Equal(t, http.StatusMovedPermanently, recorder.Code)
	assert.Equal(t, "http://example.com/new", recorder.Header().Get("Location"))
}

func TestShouldReturn302Found(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "http://example.com/page", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.Found("/redirect")

	// Assert
	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, "http://example.com/redirect", recorder.Header().Get("Location"))
}

func TestShouldReturn303SeeOther(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "http://example.com/form", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.SeeOther("/result")

	// Assert
	assert.Equal(t, http.StatusSeeOther, recorder.Code)
	assert.Equal(t, "http://example.com/result", recorder.Header().Get("Location"))
}

func TestShouldReturn307TemporaryRedirect(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.TemporaryRedirect("/new-api")

	// Assert
	assert.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
	assert.Equal(t, "http://example.com/new-api", recorder.Header().Get("Location"))
}

func TestShouldReturn308PermanentRedirect(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/v1", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.PermanentRedirect("/api/v2")

	// Assert
	assert.Equal(t, http.StatusPermanentRedirect, recorder.Code)
	assert.Equal(t, "http://example.com/api/v2", recorder.Header().Get("Location"))
}

func TestShouldHandleAbsoluteURLInRedirect(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "http://example.com/page", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	// Act
	ctx.Found("https://other.com/page")

	// Assert
	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, "https://other.com/page", recorder.Header().Get("Location"))
}
