package routing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type adapterRouteContext struct {
	RouteContext
	request *http.Request
}

func (c *adapterRouteContext) Request() *http.Request {
	return c.request
}

func (c *adapterRouteContext) SetRequest(r *http.Request) {
	c.request = r
}

func TestHTTPHandlerShouldBindRouteContextToRequest(t *testing.T) {
	// Arrange
	recorder := httptest.NewRecorder()
	baseRequest := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	baseContext := NewRouteContext(recorder, baseRequest)
	params := &Params{}
	params.Set("id", "42")
	baseContext.SetParamsSlice(params)

	var adapterContext *adapterRouteContext
	handlerCalled := false
	wrapped := HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		routeCtx, ok := RouteContextFromRequest(r)
		require.True(t, ok)
		assert.Same(t, adapterContext, routeCtx)
		assert.Same(t, r, routeCtx.Request())

		id, ok := routeCtx.Param("id")
		require.True(t, ok)
		assert.Equal(t, "42", id)

		w.WriteHeader(http.StatusNoContent)
	}))
	require.NotNil(t, wrapped)

	adapterContext = &adapterRouteContext{
		RouteContext: baseContext,
		request:      httptest.NewRequest(http.MethodGet, "/users/42", nil),
	}

	// Act
	wrapped(adapterContext)

	// Assert
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusNoContent, recorder.Code)
}
