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

func testClientGET(t *testing.T, url string) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return testClient.Do(req)
}

func testClientPOST(t *testing.T, url, contentType string, body io.Reader) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return testClient.Do(req)
}

func mustReadBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return b
}
