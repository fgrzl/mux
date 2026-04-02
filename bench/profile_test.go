package bench

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/fgrzl/mux/test"
)

// TestAllocationProfile performs detailed allocation tracking for a single GET request.
func TestAllocationProfile(t *testing.T) {
	server := httptest.NewServer(test.MockServerHandler())
	defer server.Close()

	// Warmup
	for i := 0; i < 100; i++ {
		resp, _ := benchClient.Get(server.URL + "/api/v1/resources/1")
		if resp != nil {
			readAndClose(resp)
		}
	}

	// Force GC to get clean state
	runtime.GC()

	// Get initial state
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Single request
	resp, err := benchClient.Get(server.URL + "/api/v1/resources/1")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	readAndClose(resp)

	// Get final state
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	fmt.Printf("\n=== Single GET Request Allocation Profile ===\n")
	fmt.Printf("Alloc delta: %d bytes\n", m2.Alloc-m1.Alloc)
	fmt.Printf("TotalAlloc delta: %d bytes\n", m2.TotalAlloc-m1.TotalAlloc)
	fmt.Printf("Mallocs delta: %d\n", m2.Mallocs-m1.Mallocs)
	fmt.Printf("Frees delta: %d\n", m2.Frees-m1.Frees)
	fmt.Printf("Live objects: %d\n", m2.Mallocs-m2.Frees)
}

// BenchmarkAllocationsDetailed breaks down where allocations come from
func BenchmarkAllocationsDetailed(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("GET/SingleResource/Breakdown", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			resp, _ := benchClient.Get(server.URL + "/api/v1/resources/1")
			if resp != nil {
				readAndClose(resp)
			}
		}
	})

	b.Run("HEAD/Only", func(b *testing.B) {
		// HEAD should be lighter than GET
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("HEAD", server.URL+"/api/v1/resources/1", nil)
			resp, _ := benchClient.Do(req)
			if resp != nil {
				resp.Body.Close()
			}
		}
	})

	b.Run("DELETE/Minimal", func(b *testing.B) {
		// DELETE to check if method matters
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("DELETE", server.URL+"/api/v1/tenants/1", nil)
			resp, _ := benchClient.Do(req)
			if resp != nil {
				readAndClose(resp)
			}
		}
	})

	b.Run("NoContent/204", func(b *testing.B) {
		// Lightest possible response
		server2 := newBenchmarkServer(b)
		// Register a route that returns NoContent
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			resp, _ := benchClient.Get(server2.URL + "/api/v1/resources/999")
			if resp != nil {
				readAndClose(resp)
			}
		}
	})
}
