package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkThroughput measures requests per second under sustained load.
func BenchmarkThroughput(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("Sustained/Get", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		start := time.Now()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + testsupport.APIResources)
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
		elapsed := time.Since(start)
		b.ReportMetric(float64(b.N)/elapsed.Seconds(), "req/s")
	})

	b.Run("Sustained/GetParallel", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		start := time.Now()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				resp, err := benchClient.Get(server.URL + testsupport.APIResources)
				if err != nil {
					b.Fatalf("GET failed: %v", err)
				}
				readAndClose(resp)
			}
		})
		elapsed := time.Since(start)
		b.ReportMetric(float64(b.N)/elapsed.Seconds(), "req/s")
	})

	b.Run("Sustained/MixedLoad", func(b *testing.B) {
		var seq uint64
		b.ReportAllocs()
		b.ResetTimer()

		start := time.Now()
		b.RunParallel(func(pb *testing.PB) {
			local := 0
			for pb.Next() {
				local++
				// Workload mix: 70% reads, 20% searches, 10% writes
				switch local % 10 {
				case 0: // Write
					n := atomic.AddUint64(&seq, 1)
					resources := []testsupport.Resource{{TenantID: 0, Name: fmt.Sprintf("t-%d", n), Type: "resource"}}
					bts, _ := json.Marshal(resources)
					resp, _ := benchClient.Post(server.URL+testsupport.APIBase+"/resources/bulk", common.MimeJSON, bytes.NewReader(bts))
					if resp != nil {
						readAndClose(resp)
					}
				case 1, 2: // Search
					resp, _ := benchClient.Get(server.URL + testsupport.APIBase + "/resources/search?type=resource")
					if resp != nil {
						readAndClose(resp)
					}
				default: // Read
					id := (local % 10) + 1
					resp, _ := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APIResourceByID, id))
					if resp != nil {
						readAndClose(resp)
					}
				}
			}
		})
		elapsed := time.Since(start)
		b.ReportMetric(float64(b.N)/elapsed.Seconds(), "req/s")
	})
}
