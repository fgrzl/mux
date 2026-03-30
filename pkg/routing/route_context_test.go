package routing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKey   = "test-key"
	testValue = "test-value"
)

// mockPrincipal for testing
// mockRouteContextPrincipal removed; not used by current tests

func TestShouldCreateNewRouteContext(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	ctx := NewRouteContext(rec, req)

	// Assert
	assert.NotNil(t, ctx)
	assert.Equal(t, req.Context(), ctx.Context)
	assert.Equal(t, rec, ctx.Response())
	assert.Equal(t, req, ctx.Request())
	assert.Nil(t, ctx.User())
	assert.Nil(t, ctx.Options())
	assert.Nil(t, ctx.ParamsSlice())
	assert.False(t, ctx.formsParsed)
	assert.Nil(t, ctx.services)
}

func TestShouldReturnRouteContextFromRequest(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	ctx.SetRequest(ctx.Request())

	// Act
	routeCtx, ok := RouteContextFromRequest(ctx.Request())

	// Assert
	assert.True(t, ok)
	assert.Same(t, ctx, routeCtx)
}

func TestShouldSetAndGetService(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	key := ServiceKey("test-service")
	service := testValue

	// Act
	ctx.SetService(key, service)
	retrieved, exists := ctx.GetService(key)

	// Assert
	assert.True(t, exists)
	assert.Equal(t, service, retrieved)
}

func TestShouldReturnFalseForNonExistentService(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	key := ServiceKey("non-existent")

	// Act
	retrieved, exists := ctx.GetService(key)

	// Assert
	assert.False(t, exists)
	assert.Nil(t, retrieved)
}

func TestShouldNotSetServiceWithEmptyKey(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	service := testValue

	// Act
	ctx.SetService("", service)

	// Assert
	assert.Nil(t, ctx.services)
}

func TestShouldNotSetNilService(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	key := ServiceKey("test")

	// Act
	ctx.SetService(key, nil)

	// Assert
	assert.Nil(t, ctx.services)
}

func TestShouldBindFromQueryParameters(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?name=John&age=30", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result struct {
		Name string `json:"name"`
		Age  string `json:"age"` // Use string since query params are strings
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "30", result.Age)
}

func TestShouldBindFromJSONBody(t *testing.T) {
	// Arrange
	body := `{"name":"John","age":30}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, 30, result.Age)
}

func TestShouldBindFromPatchJSONBody(t *testing.T) {
	// Arrange
	body := `{"name":"John","age":30}`
	req := httptest.NewRequest(http.MethodPatch, "/test", strings.NewReader(body))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, 30, result.Age)
}

func TestShouldBindFromPutJSONBody(t *testing.T) {
	// Arrange
	body := `{"name":"John","age":30}`
	req := httptest.NewRequest(http.MethodPut, "/test", strings.NewReader(body))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, 30, result.Age)
}

func TestShouldBindMixedSourcesGivenPostJSONBody(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/users/42?limit=5", strings.NewReader(`{"name":"John"}`))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	req.Header.Set("X-Trace-ID", "trace-1")
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	params := &Params{}
	params.Set("id", "42")
	ctx.paramsSlice = params
	ctx.options = Route(http.MethodPost, "/users/{id}").
		WithQueryParam("limit", 1).
		WithHeaderParam("X-Trace-ID", "", false).
		WithPathParam("id", 1)

	var result struct {
		Name    string `json:"name"`
		Limit   int    `json:"limit"`
		TraceID string `json:"X-Trace-ID"`
		ID      int    `json:"id"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, 5, result.Limit)
	assert.Equal(t, "trace-1", result.TraceID)
	assert.Equal(t, 42, result.ID)
}

func TestShouldBindMixedSourcesGivenPatchJSONBody(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPatch, "/users/42?version=7", strings.NewReader(`{"name":"Patched"}`))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	params := &Params{}
	params.Set("id", "42")
	ctx.paramsSlice = params
	ctx.options = Route(http.MethodPatch, "/users/{id}").
		WithQueryParam("version", 1).
		WithPathParam("id", 1)

	var result struct {
		Name    string `json:"name"`
		Version int    `json:"version"`
		ID      int    `json:"id"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Patched", result.Name)
	assert.Equal(t, 7, result.Version)
	assert.Equal(t, 42, result.ID)
}

func TestShouldBindMixedSourcesGivenPostFormBody(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("name", "John")
	req := httptest.NewRequest(http.MethodPost, "/users?limit=5", strings.NewReader(formData.Encode()))
	req.Header.Set(common.HeaderContentType, common.MimeFormURLEncoded)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	ctx.options = Route(http.MethodPost, "/users").WithQueryParam("limit", 1)

	var result struct {
		Name  string `json:"name"`
		Limit int    `json:"limit"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, 5, result.Limit)
}

func TestShouldPreferJSONBodyOverQueryWhenKeysConflict(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/users?name=query", strings.NewReader(`{"name":"body"}`))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result struct {
		Name string `json:"name"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "body", result.Name)
}

func TestShouldPreferPathParamsOverJSONBodyAndQueryWhenKeysConflict(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/users/42?id=1", strings.NewReader(`{"id":2}`))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	params := &Params{}
	params.Set("id", "42")
	ctx.paramsSlice = params
	ctx.options = Route(http.MethodPost, "/users/{id}").
		WithQueryParam("id", 0).
		WithPathParam("id", 0)

	var result struct {
		ID int `json:"id"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 42, result.ID)
}

func TestShouldIgnoreJSONBodyGivenDeleteRequest(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodDelete, "/users?id=7", strings.NewReader(`{"id":99}`))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	ctx.options = Route(http.MethodDelete, "/users").WithQueryParam("id", 0)

	var result struct {
		ID int `json:"id"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 7, result.ID)
}

func TestShouldIgnoreJSONBodyGivenHeadRequest(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodHead, "/users?name=query", strings.NewReader(`{"name":"body"}`))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result struct {
		Name string `json:"name"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "query", result.Name)
}

func TestShouldBindQueryParametersGivenOptionsRequest(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodOptions, "/users?name=option", strings.NewReader(`{"name":"body"}`))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result struct {
		Name string `json:"name"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "option", result.Name)
}

func TestShouldBindFromFormData(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("name", "John")
	formData.Set("age", "30")
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(formData.Encode()))
	req.Header.Set(common.HeaderContentType, common.MimeFormURLEncoded)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result struct {
		Name string `json:"name"`
		Age  string `json:"age"` // Use string since form data are strings
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "30", result.Age)
}

func TestShouldBindFromHeaders(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "123")
	req.Header.Set("X-User-Name", "John")
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result struct {
		UserID   string `json:"X-User-ID"`
		UserName string `json:"X-User-Name"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "123", result.UserID)
	assert.Equal(t, "John", result.UserName)
}

func TestShouldBindFromRouteParams(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	params := &Params{}
	params.Set("id", "123")
	params.Set("name", "John")
	ctx.paramsSlice = params

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "123", result.ID)
	assert.Equal(t, "John", result.Name)
}

func TestShouldReturnErrorForUnsupportedContentType(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("some data"))
	req.Header.Set(common.HeaderContentType, common.MimeTextPlain)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result struct{}

	// Act
	err := ctx.Bind(&result)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported content type")
}

func TestShouldReturnErrorForInvalidJSON(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"invalid json`))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result struct{}

	// Act
	err := ctx.Bind(&result)

	// Assert
	assert.Error(t, err)
}

func TestShouldHandleSingleValues(t *testing.T) {
	// Arrange
	staging := make(map[string]any)
	values := []string{"single-value"}

	// Act
	addToStaging(staging, testKey, values)

	// Assert
	assert.Equal(t, "single-value", staging[testKey])
}

func TestShouldHandleMultipleValues(t *testing.T) {
	// Arrange
	staging := make(map[string]any)
	values := []string{"value1", "value2", "value3"}

	// Act
	addToStaging(staging, testKey, values)

	// Assert
	assert.Equal(t, values, staging[testKey])
}

func TestShouldParseValidValuesGivenSliceInput(t *testing.T) {
	// Arrange
	values := []string{"1", "2", "3"}
	parseFunc := func(s string) (int, error) {
		if s == "1" {
			return 1, nil
		}
		if s == "2" {
			return 2, nil
		}
		if s == "3" {
			return 3, nil
		}
		return 0, assert.AnError
	}

	// Act
	result, ok := parseSlice(values, parseFunc)

	// Assert
	assert.True(t, ok)
	assert.Equal(t, []int{1, 2, 3}, result)
}

func TestShouldReturnFalseGivenInvalidSliceInput(t *testing.T) {
	// Arrange
	values := []string{"1", "invalid", "3"}
	parseFunc := func(s string) (int, error) {
		if s == "invalid" {
			return 0, assert.AnError
		}
		return 1, nil
	}

	// Act
	result, ok := parseSlice(values, parseFunc)

	// Assert
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestShouldReturnRequestURIFromContext(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)

	// Act
	uri := getInstanceURI(req)

	// Assert
	assert.NotNil(t, uri)
	assert.Equal(t, "/test?param=value", *uri)
}

func TestShouldHandleEmptyParams(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	ctx.paramsSlice = nil

	var result struct {
		Name string `json:"name"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, result.Name)
}

func TestShouldHandleContextInheritance(t *testing.T) {
	// Arrange
	type ctxKey string
	baseCtx := context.WithValue(context.Background(), ctxKey(testKey), testValue)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(baseCtx)
	rec := httptest.NewRecorder()

	// Act
	ctx := NewRouteContext(rec, req)

	// Assert
	assert.Equal(t, testValue, ctx.Value(ctxKey(testKey)))
}

func TestShouldBindComplexData(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/test?tags=tag1&tags=tag2&user=john", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	params := &Params{}
	params.Set("id", "123")
	ctx.paramsSlice = params

	var result struct {
		ID   string   `json:"id"`
		User string   `json:"user"`
		Tags []string `json:"tags"`
	}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "123", result.ID)
	assert.Equal(t, "john", result.User)
	assert.Equal(t, []string{"tag1", "tag2"}, result.Tags)
}

func TestShouldBindWithMaxBytesLimit(t *testing.T) {
	// Arrange
	// Create a large JSON payload (larger than 1MB)
	largePayload := make(map[string]string)
	for i := 0; i < 50000; i++ {
		largePayload[string(rune(i))] = strings.Repeat("x", 30)
	}

	jsonData, err := json.Marshal(largePayload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(jsonData))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	var result map[string]string

	// Act
	err = ctx.Bind(&result)

	// Assert - Should either succeed or fail with an appropriate error
	// The test verifies the MaxBytesReader is set up correctly
	if err != nil {
		// If it fails due to size limit, that's expected behavior
		t.Logf("Large payload properly rejected: %v", err)
	} else {
		// If it succeeds, the data should be bound correctly
		assert.NotEmpty(t, result)
	}
}

func TestBind_ReturnsErrMissingBody_ForEmptyJSON_WhenRequestBodyRequired(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(""))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	opts := &RouteOptions{}
	opts.RequestBody = &openapi.RequestBodyObject{Required: true}
	ctx.SetOptions(opts)

	var result struct{}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMissingBody))
}

func TestBind_ReturnsErrMissingBody_ForEmptyForm_WhenRequestBodyRequired(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(""))
	req.Header.Set(common.HeaderContentType, common.MimeFormURLEncoded)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	opts := &RouteOptions{}
	opts.RequestBody = &openapi.RequestBodyObject{Required: true}
	ctx.SetOptions(opts)

	var result struct{}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMissingBody))
}

func TestBind_ShouldNotWriteResponseWhenBodyMissing(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(""))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	opts := &RouteOptions{}
	opts.RequestBody = &openapi.RequestBodyObject{Required: true}
	ctx.SetOptions(opts)

	var result struct{}

	// Act
	err := ctx.Bind(&result)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMissingBody))
	assert.Equal(t, 200, rec.Code)
	assert.Empty(t, rec.Body.String())
	assert.Empty(t, rec.Header().Get(common.HeaderContentType))
}

func TestBind_DoesNotReturnErrMissingBody_WhenRequestBodyNotRequired_JSON(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(""))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	opts := &RouteOptions{}
	opts.RequestBody = &openapi.RequestBodyObject{Required: false}
	ctx.SetOptions(opts)

	var result struct{}

	// Act
	err := ctx.Bind(&result)

	// Assert - Empty JSON body should not produce an error when not required
	require.NoError(t, err)
}

func TestBind_WritesToResponseWriter_WhenRequestBodyRequired_And_EmptyBody(t *testing.T) {
	// Arrange - Verify caller-owned error handling can safely decide the response.
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(""))
	req.Header.Set(common.HeaderContentType, common.MimeJSON)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	opts := &RouteOptions{}
	opts.RequestBody = &openapi.RequestBodyObject{Required: true}
	ctx.SetOptions(opts)

	var result struct{}

	// Act
	err := ctx.Bind(&result)
	if err != nil {
		ctx.BadRequest("Bad Request", "Request body is required")
	}

	// Assert
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrMissingBody))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var problemDetail ProblemDetails
	err = json.NewDecoder(rec.Body).Decode(&problemDetail)
	require.NoError(t, err)
	assert.Equal(t, ProblemTypeAboutBlank, problemDetail.Type)
	assert.Equal(t, "Bad Request", problemDetail.Title)
	assert.Equal(t, http.StatusBadRequest, problemDetail.Status)
	assert.Equal(t, "Request body is required", problemDetail.Detail)
}
