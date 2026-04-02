package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkConcurrency measures system behavior under parallel load.
func BenchmarkConcurrency(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("ReadOnly/Parallel", func(b *testing.B) {
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

	b.Run("ReadSingle/Parallel", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		var counter uint64
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				id := atomic.AddUint64(&counter, 1)%10 + 1
				resp, err := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APIResourceByID, id))
				if err != nil {
					b.Fatalf("GET failed: %v", err)
				}
				readAndClose(resp)
			}
		})
	})

	b.Run("Search/Parallel", func(b *testing.B) {
		resources := testsupport.Service.ListResources(0)
		if len(resources) == 0 {
			b.Fatal("no resources")
		}
		name := url.QueryEscape(resources[0].Name)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				resp, err := benchClient.Get(server.URL + testsupport.APIBase + "/resources/search?name=" + name + "&type=resource")
				if err != nil {
					b.Fatalf("search failed: %v", err)
				}
				readAndClose(resp)
			}
		})
	})

	b.Run("Write/Parallel", func(b *testing.B) {
		var seq uint64
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				n := atomic.AddUint64(&seq, 1)
				resources := []testsupport.Resource{{TenantID: 0, Name: fmt.Sprintf("bench-p-%d", n), Type: "resource"}}
				bts, _ := json.Marshal(resources)
				resp, err := benchClient.Post(server.URL+testsupport.APIBase+"/resources/bulk", common.MimeJSON, bytes.NewReader(bts))
				if err != nil {
					b.Fatalf("POST failed: %v", err)
				}
				readAndClose(resp)
			}
		})
	})

	b.Run("MixedReadWrite/Parallel", func(b *testing.B) {
		// Reset service to prevent data accumulation from prior sub-benchmarks
		testsupport.Service.Reset()
		var seq uint64
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			local := 0
			for pb.Next() {
				local++
				if local%5 == 0 { // 20% writes
					n := atomic.AddUint64(&seq, 1)
					resources := []testsupport.Resource{{TenantID: 0, Name: fmt.Sprintf("bench-m-%d", n), Type: "resource"}}
					bts, _ := json.Marshal(resources)
					resp, err := benchClient.Post(server.URL+testsupport.APIBase+"/resources/bulk", common.MimeJSON, bytes.NewReader(bts))
					if err != nil {
						b.Fatalf("POST failed: %v", err)
					}
					readAndClose(resp)
				} else { // 80% reads
					resp, err := benchClient.Get(server.URL + testsupport.APIResources)
					if err != nil {
						b.Fatalf("GET failed: %v", err)
					}
					readAndClose(resp)
				}
			}
		})
	})
}
