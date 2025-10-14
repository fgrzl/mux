package testhelpers

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/fgrzl/mux/pkg/routing"
)

// NewRequestRecorder creates an http.Request and ResponseRecorder for tests.
func NewRequestRecorder(method, url string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, url, body)
	rec := httptest.NewRecorder()
	return req, rec
}

// NewRouteContext creates a routing.RouteContext and returns it along with
// the underlying ResponseRecorder for assertions.
func NewRouteContext(method, url string, body io.Reader) (*routing.DefaultRouteContext, *httptest.ResponseRecorder) {
	req, rec := NewRequestRecorder(method, url, body)
	return routing.NewRouteContext(rec, req), rec
}
