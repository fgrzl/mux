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
	defer resp.Body.Close()
	_, err := io.ReadAll(resp.Body)
	return err
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

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
