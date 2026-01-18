package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkMemoryPressure measures system behavior under memory pressure.
func BenchmarkMemoryPressure(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("HighAllocation", func(b *testing.B) {
		var seq uint64
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resources := make([]testsupport.Resource, 50)
			n := atomic.AddUint64(&seq, 1)
			for j := range resources {
				resources[j] = testsupport.Resource{
					TenantID: int32(j % 10),
					Name:     fmt.Sprintf("resource-%d-%d", n, j),
					Type:     "resource",
				}
			}
			bts, _ := json.Marshal(resources)
			resp, err := benchClient.Post(server.URL+testsupport.APIBase+"/resources/bulk", common.MimeJSON, bytes.NewReader(bts))
			if err != nil {
				b.Fatalf("POST failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("LargeMetadata", func(b *testing.B) {
		metadata := make(map[string]string, 200)
		for i := 0; i < 200; i++ {
			metadata[fmt.Sprintf("key-%03d", i)] = fmt.Sprintf("This is a longer value string for key %d that takes up more memory", i)
		}
		body := map[string]any{"metadata": metadata}
		bts, _ := json.Marshal(body)

		b.ReportAllocs()
		b.SetBytes(int64(len(bts)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("PUT", server.URL+fmt.Sprintf(testsupport.APIResourceMetadata, 1), bytes.NewReader(bts))
			req.Header.Set(common.HeaderContentType, common.MimeJSON)
			resp, err := benchClient.Do(req)
			if err != nil {
				b.Fatalf("PUT failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("RapidGC", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + testsupport.APIResources)
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
			if i%100 == 0 {
				runtime.GC()
			}
		}
	})
}
