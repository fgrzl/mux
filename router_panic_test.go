package mux

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldRecoverFromPanicInHandler(t *testing.T) {
	// Arrange
	router := NewRouter()
	router.GET("/panic", func(c RouteContext) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()

	// Act
	router.ServeHTTP(recorder, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))

	var problemDetails ProblemDetails
	err := json.Unmarshal(recorder.Body.Bytes(), &problemDetails)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, problemDetails.Status)
	assert.Equal(t, "Internal Server Error", problemDetails.Title)
	assert.Equal(t, "An unexpected error occurred", problemDetails.Detail)
}

// Helper middleware that panics for testing
type testPanicMiddleware struct{}

func (m *testPanicMiddleware) Invoke(ctx *DefaultRouteContext, next HandlerFunc) {
	panic("middleware panic")
}
