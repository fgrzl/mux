package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkRegressionBaseline provides stable baselines for regression detection.
func BenchmarkRegressionBaseline(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("GetSingleResource", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, _ := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APIResourceByID, 1))
			if resp != nil {
				readAndClose(resp)
			}
		}
	})

	b.Run("ListResources", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, _ := benchClient.Get(server.URL + testsupport.APIResources)
			if resp != nil {
				readAndClose(resp)
			}
		}
	})

	b.Run("CreateResource", func(b *testing.B) {
		var seq uint64
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			n := atomic.AddUint64(&seq, 1)
			resources := []testsupport.Resource{{TenantID: 0, Name: fmt.Sprintf("reg-%d", n), Type: "resource"}}
			bts, _ := json.Marshal(resources)
			resp, _ := benchClient.Post(server.URL+testsupport.APIBase+"/resources/bulk", common.MimeJSON, bytes.NewReader(bts))
			if resp != nil {
				readAndClose(resp)
			}
		}
	})

	b.Run("ParallelRead", func(b *testing.B) {
		// Reset service to prevent data accumulation from prior sub-benchmarks
		testsupport.Service.Reset()
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				resp, _ := benchClient.Get(server.URL + testsupport.APIResources)
				if resp != nil {
					readAndClose(resp)
				}
			}
		})
	})
}
