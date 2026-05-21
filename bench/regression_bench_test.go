package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkRegressionBaseline provides stable baselines for regression detection.
func BenchmarkRegressionBaseline(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("GetSingleResource", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = benchClientDo(b, http.MethodGet, server.URL+fmt.Sprintf(testsupport.APIResourceByID, 1), nil, "")
		}
	})

	b.Run("ListResources", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = benchClientDo(b, http.MethodGet, server.URL+testsupport.APIResources, nil, "")
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
			_, _ = benchClientDo(b, http.MethodPost, server.URL+testsupport.APIBase+"/resources/bulk", bytes.NewReader(bts), common.MimeJSON)
		}
	})

	b.Run("ParallelRead", func(b *testing.B) {
		// Reset service to prevent data accumulation from prior sub-benchmarks
		testsupport.Service.Reset()
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = benchClientDo(b, http.MethodGet, server.URL+testsupport.APIResources, nil, "")
			}
		})
	})
}
