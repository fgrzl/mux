package testsupport

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/fgrzl/mux/pkg/routing"
	"github.com/fgrzl/mux/test/testhelpers"
)

// NewRequestRecorder creates an http.Request and ResponseRecorder for tests.
// Delegates to the canonical testhelpers implementation to avoid duplication.
func NewRequestRecorder(method, url string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	return testhelpers.NewRequestRecorder(method, url, body)
}

// NewRouteContext creates a routing.RouteContext and returns it along with
// the underlying ResponseRecorder for assertions.
// Delegates to the canonical testhelpers implementation to avoid duplication.
func NewRouteContext(method, url string, body io.Reader) (routing.RouteContext, *httptest.ResponseRecorder) {
	// testhelpers.NewRouteContext returns (*routing.DefaultRouteContext, *httptest.ResponseRecorder)
	// which implements routing.RouteContext; return it as the interface type.
	ctx, rec := testhelpers.NewRouteContext(method, url, body)
	return ctx, rec
}
