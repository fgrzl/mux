package testsupport

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/fgrzl/mux/pkg/routing"
	"github.com/fgrzl/mux/test/testhelpers"
)

// NewRequestRecorder creates an http.Request and ResponseRecorder for tests.
// This is the preferred test helper - it delegates to testhelpers for the implementation.
func NewRequestRecorder(method, url string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	return testhelpers.NewRequestRecorder(method, url, body)
}

// NewRouteContext creates a routing.RouteContext and returns it along with
// the underlying ResponseRecorder for assertions.
// This is the preferred test helper - it delegates to testhelpers for the implementation.
func NewRouteContext(method, url string, body io.Reader) (routing.RouteContext, *httptest.ResponseRecorder) {
	ctx, rec := testhelpers.NewRouteContext(method, url, body)
	return ctx, rec
}
