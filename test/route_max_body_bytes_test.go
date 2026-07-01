package test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fgrzl/mux"
	"github.com/stretchr/testify/require"
)

type maxBodyPayload struct {
	Data string `json:"data"`
}

// maxBodyEcho binds the JSON body and answers 200 with the decoded length, 413
// when the body exceeds the route's limit (http.MaxBytesReader trips during
// Bind and surfaces as *http.MaxBytesError), or 400 for any other bind failure.
// The 413 branch mirrors how a downstream consumer keeps a distinct
// "payload too large" response while using plain c.Bind.
func maxBodyEcho(c mux.RouteContext) {
	var body maxBodyPayload
	if err := c.Bind(&body); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.Problem(&mux.ProblemDetails{
				Type:   "about:blank",
				Title:  "payload_too_large",
				Status: http.StatusRequestEntityTooLarge,
				Detail: "request body exceeds the route limit",
			})
			return
		}
		c.BadRequest("invalid_request", err.Error())
		return
	}
	c.OK(map[string]int{"len": len(body.Data)})
}

// jsonBodyOfSize returns a JSON document {"data":"aaaa..."} whose serialized
// length is at least size bytes.
func jsonBodyOfSize(t *testing.T, size int) string {
	t.Helper()
	encoded, err := json.Marshal(maxBodyPayload{Data: strings.Repeat("a", size)})
	require.NoError(t, err)
	return string(encoded)
}

func postJSON(t *testing.T, router *mux.Router, path, body string) int {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, path, strings.NewReader(body))
	req.Header.Set(mux.HeaderContentType, mux.MimeJSON)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec.Code
}

func TestShouldRaiseBodyLimitAboveRouterDefaultGivenPerRouteOverride(t *testing.T) {
	// Arrange: router-wide default of 1KB; a single route raised to 64KB.
	router := mux.NewRouter(mux.WithMaxBodyBytes(1 << 10))
	router.POST("/big", maxBodyEcho).
		WithJSONBody(maxBodyPayload{}).
		WithMaxBodyBytes(64<<10).
		WithOKResponse(map[string]int{}).
		WithResponse(http.StatusRequestEntityTooLarge, mux.ProblemDetails{}).
		WithBadRequestResponse()

	// Act + Assert: 8KB exceeds the router-wide 1KB but fits the per-route 64KB.
	require.Equal(t, http.StatusOK, postJSON(t, router, "/big", jsonBodyOfSize(t, 8<<10)))
	// 128KB exceeds the per-route limit and is rejected.
	require.Equal(t, http.StatusRequestEntityTooLarge, postJSON(t, router, "/big", jsonBodyOfSize(t, 128<<10)))
}

func TestShouldTightenBodyLimitBelowRouterDefaultGivenPerRouteOverride(t *testing.T) {
	// Arrange: router-wide default of 1KB; a single route tightened to 256 bytes.
	router := mux.NewRouter(mux.WithMaxBodyBytes(1 << 10))
	router.POST("/small", maxBodyEcho).
		WithJSONBody(maxBodyPayload{}).
		WithMaxBodyBytes(256).
		WithOKResponse(map[string]int{}).
		WithResponse(http.StatusRequestEntityTooLarge, mux.ProblemDetails{}).
		WithBadRequestResponse()

	// Act + Assert: a 64-byte body fits the tightened 256-byte limit.
	require.Equal(t, http.StatusOK, postJSON(t, router, "/small", jsonBodyOfSize(t, 64)))
	// A 512-byte body is under the router-wide 1KB but over the per-route 256.
	require.Equal(t, http.StatusRequestEntityTooLarge, postJSON(t, router, "/small", jsonBodyOfSize(t, 512)))
}

func TestShouldInheritRouterDefaultGivenRouteWithoutOverride(t *testing.T) {
	// Arrange: router-wide default of 1KB; a route on the same router with no override.
	router := mux.NewRouter(mux.WithMaxBodyBytes(1 << 10))
	router.POST("/raised", maxBodyEcho).WithMaxBodyBytes(64 << 10)
	router.POST("/default", maxBodyEcho)

	// Act + Assert: the un-overridden route still enforces the router-wide 1KB...
	require.Equal(t, http.StatusRequestEntityTooLarge, postJSON(t, router, "/default", jsonBodyOfSize(t, 8<<10)))
	require.Equal(t, http.StatusOK, postJSON(t, router, "/default", jsonBodyOfSize(t, 64)))
	// ...while the overridden sibling accepts the larger body.
	require.Equal(t, http.StatusOK, postJSON(t, router, "/raised", jsonBodyOfSize(t, 8<<10)))
}

func TestShouldRaiseBodyLimitAboveBuiltInDefaultGivenPerRouteOverride(t *testing.T) {
	// Arrange: no router-wide option, so the built-in 1MB default applies.
	router := mux.NewRouter()
	router.POST("/upload", maxBodyEcho).WithMaxBodyBytes(4 << 20)
	router.POST("/plain", maxBodyEcho)

	// Act + Assert: a ~2MB body exceeds the built-in 1MB but fits the per-route 4MB.
	require.Equal(t, http.StatusOK, postJSON(t, router, "/upload", jsonBodyOfSize(t, 2<<20)))
	// The same body on a route without the override trips the built-in default.
	require.Equal(t, http.StatusRequestEntityTooLarge, postJSON(t, router, "/plain", jsonBodyOfSize(t, 2<<20)))
}
