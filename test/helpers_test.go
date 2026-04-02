package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fgrzl/mux/test/testsupport"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	// ensure fresh service state per server instance to avoid cross-test tampering
	testsupport.Service = testsupport.NewFakeService()
	s := httptest.NewServer(mockServerHandler())
	t.Cleanup(func() { s.Close() })
	return s
}

func newTestServerWithHandler(t *testing.T, h http.Handler) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(h)
	t.Cleanup(func() { s.Close() })
	return s
}

var testClient = &http.Client{Timeout: 5 * time.Second}

func mustReadBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return b
}
