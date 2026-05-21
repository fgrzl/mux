package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkLatencyDistribution captures latency percentiles for different operations.
func BenchmarkLatencyDistribution(b *testing.B) {
	server := newBenchmarkServer(b)

	measureLatencies := func(b *testing.B, name string, doRequest func() error) {
		// Warmup
		for i := 0; i < 100; i++ {
			_ = doRequest()
		}

		// Collect samples
		const samples = 10000
		latencies := make([]time.Duration, samples)

		b.ResetTimer()
		for i := 0; i < samples; i++ {
			start := time.Now()
			if err := doRequest(); err != nil {
				b.Fatalf("request failed: %v", err)
			}
			latencies[i] = time.Since(start)
		}
		b.StopTimer()

		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

		p50 := latencies[int(float64(samples)*0.50)]
		p90 := latencies[int(float64(samples)*0.90)]
		p95 := latencies[int(float64(samples)*0.95)]
		p99 := latencies[int(float64(samples)*0.99)]
		pMax := latencies[samples-1]

		b.ReportMetric(float64(p50.Microseconds()), fmt.Sprintf("%s_p50_us", name))
		b.ReportMetric(float64(p90.Microseconds()), fmt.Sprintf("%s_p90_us", name))
		b.ReportMetric(float64(p95.Microseconds()), fmt.Sprintf("%s_p95_us", name))
		b.ReportMetric(float64(p99.Microseconds()), fmt.Sprintf("%s_p99_us", name))
		b.ReportMetric(float64(pMax.Microseconds()), fmt.Sprintf("%s_max_us", name))
	}

	b.Run("GET/Resource", func(b *testing.B) {
		measureLatencies(b, "GET", func() error {
			_, err := benchClientDo(b, http.MethodGet, server.URL+fmt.Sprintf(testsupport.APIResourceByID, 1), nil, "")
			return err
		})
	})

	b.Run("GET/List", func(b *testing.B) {
		measureLatencies(b, "List", func() error {
			_, err := benchClientDo(b, http.MethodGet, server.URL+testsupport.APIResources, nil, "")
			return err
		})
	})

	b.Run("POST/Create", func(b *testing.B) {
		var seq uint64
		measureLatencies(b, "Create", func() error {
			n := atomic.AddUint64(&seq, 1)
			resources := []testsupport.Resource{{TenantID: 0, Name: fmt.Sprintf("lat-%d", n), Type: "resource"}}
			bts, _ := json.Marshal(resources)
			_, err := benchClientDo(b, http.MethodPost, server.URL+testsupport.APIBase+"/resources/bulk", bytes.NewReader(bts), common.MimeJSON)
			return err
		})
	})
}
