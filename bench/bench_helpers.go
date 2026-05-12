package bench

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fgrzl/mux/test"
	"github.com/fgrzl/mux/test/testsupport"
)

// newBenchmarkServer creates a fresh server for benchmarks and resets the
// in-memory test service for isolation across benchmark runs.
func newBenchmarkServer(b *testing.B) *httptest.Server {
	b.Helper()
	testsupport.Service = testsupport.NewFakeService()
	s := httptest.NewServer(test.MockServerHandler())
	b.Cleanup(func() { s.Close() })
	return s
}

// readAndClose reads and discards the response body and closes it.
func readAndClose(resp *http.Response) error {
	defer func() {
		_ = resp.Body.Close()
	}()
	_, err := io.ReadAll(resp.Body)
	return err
}

// benchClientDo issues an HTTP request using the benchmark client's transport
// and the TB's context (supports *testing.B and *testing.T). The response body
// is drained and closed before returning; StatusCode and headers remain valid.
func benchClientDo(tb testing.TB, method, url string, body io.Reader, contentType string) (*http.Response, error) {
	tb.Helper()
	req, err := http.NewRequestWithContext(tb.Context(), method, url, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := benchClient.Do(req)
	if err != nil {
		return nil, err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp, nil
}

// benchClient is an HTTP client optimized for benchmarking with connection pooling.
var benchClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
}

// benchMin returns the smaller of two integers (avoid shadowing the builtin min).
func benchMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
