package bench

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkStress runs stress tests to find breaking points.
func BenchmarkStress(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("HighConcurrency", func(b *testing.B) {
		numWorkers := runtime.GOMAXPROCS(0) * 4
		var wg sync.WaitGroup
		var ops int64

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		b.ResetTimer()
		start := time.Now()

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
						resp, err := benchClient.Get(server.URL + testsupport.APIResources)
						if err == nil && resp != nil {
							readAndClose(resp)
							atomic.AddInt64(&ops, 1)
						}
					}
				}
			}()
		}

		for atomic.LoadInt64(&ops) < int64(b.N) {
			runtime.Gosched()
		}
		cancel()
		wg.Wait()

		elapsed := time.Since(start)
		b.ReportMetric(float64(ops)/elapsed.Seconds(), "ops/s")
	})

	b.Run("BurstTraffic", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for burst := 0; burst < b.N/100+1; burst++ {
			var wg sync.WaitGroup
			for i := 0; i < min(100, b.N-burst*100); i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					resp, err := benchClient.Get(server.URL + testsupport.APIResources)
					if err == nil && resp != nil {
						readAndClose(resp)
					}
				}()
			}
			wg.Wait()
			time.Sleep(time.Millisecond)
		}
	})
}
