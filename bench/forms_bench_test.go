package bench

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkFormProcessing measures form data handling performance.
func BenchmarkFormProcessing(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("SimpleForm", func(b *testing.B) {
		formData := url.Values{"field": {"test-value"}}
		body := formData.Encode()
		b.ReportAllocs()
		b.SetBytes(int64(len(body)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("POST", server.URL+testsupport.APIFormSubmit, bytes.NewReader([]byte(body)))
			req.Header.Set(common.HeaderContentType, "application/x-www-form-urlencoded")
			resp, err := benchClient.Do(req)
			if err != nil {
				b.Fatalf("POST failed: %v", err)
			}
			readAndClose(resp)
		}
	})
}
