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
			resp, err := benchClient.Get(server.URL + testsupport.APIResources)
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
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
			resp, err := noKeepaliveClient.Get(server.URL + testsupport.APIResources)
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("ParallelConnections", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				resp, err := benchClient.Get(server.URL + testsupport.APIResources)
				if err != nil {
					b.Fatalf("GET failed: %v", err)
				}
				readAndClose(resp)
			}
		})
	})
}
