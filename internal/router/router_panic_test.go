package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/internal/routing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldRecoverFromPanicInHandler(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	rtr.GET("/panic", func(c routing.RouteContext) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()

	// Act
	rtr.ServeHTTP(recorder, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, common.MimeProblemJSON, recorder.Header().Get(common.HeaderContentType))

	var problemDetails common.ProblemDetails
	err := json.Unmarshal(recorder.Body.Bytes(), &problemDetails)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, problemDetails.Status)
	assert.Equal(t, "Internal Server Error", problemDetails.Title)
	assert.Equal(t, "An unexpected error occurred", problemDetails.Detail)
}

// Helper middleware that panics for testing
// No unused middleware types kept in this test file.
