package bench

import (
	"net/http"
	"testing"
	"time"

	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkConnections measures connection handling overhead.
func BenchmarkConnections(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("KeepAlive", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := benchClientDo(b, http.MethodGet, server.URL+testsupport.APIResources, nil, "")
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
		}
	})

	b.Run("NoKeepAlive", func(b *testing.B) {
		noKeepaliveClient := &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, err := http.NewRequestWithContext(b.Context(), http.MethodGet, server.URL+testsupport.APIResources, nil)
			if err != nil {
				b.Fatalf("NewRequest: %v", err)
			}
			resp, err := noKeepaliveClient.Do(req)
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			_ = readAndClose(resp)
		}
	})

	b.Run("ParallelConnections", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := benchClientDo(b, http.MethodGet, server.URL+testsupport.APIResources, nil, "")
				if err != nil {
					b.Fatalf("GET failed: %v", err)
				}
			}
		})
	})
}
