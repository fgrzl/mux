package bench

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkErrorPaths measures performance of error handling paths.
func BenchmarkErrorPaths(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("NotFound", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APIResourceByID, 999999))
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			if resp.StatusCode != http.StatusNotFound {
				b.Fatalf("expected 404, got %d", resp.StatusCode)
			}
			readAndClose(resp)
		}
	})

	b.Run("BadRequest", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Post(server.URL+testsupport.APIBase+"/resources/bulk", common.MimeJSON, bytes.NewReader([]byte("invalid json")))
			if err != nil {
				b.Fatalf("POST failed: %v", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				b.Fatalf("expected 400, got %d", resp.StatusCode)
			}
			readAndClose(resp)
		}
	})

	b.Run("InvalidParam", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + testsupport.APIBase + "/resources/notanumber")
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				b.Fatalf("expected 400, got %d", resp.StatusCode)
			}
			readAndClose(resp)
		}
	})
}
