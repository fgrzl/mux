package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fgrzl/mux/pkg/middleware/authorization"
	"github.com/fgrzl/mux/pkg/router"
)

func main() {
	// Build router similar to benchmark
	r := router.NewRouter(router.WithContextPooling())
	authorization.UseAuthorization(r, authorization.WithRoles("admin"))
	r.GET("/test", func(c router.RouteContext) { c.NoContent() })

	// Warmup
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	for i := 0; i < 1000; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}

	// Run workload using a reusable nop writer to avoid measuring httptest allocations
	N := 200000
	start := time.Now()
	w := newNopWriter()
	for i := 0; i < N; i++ {
		// reset headers/status
		for k := range w.header {
			delete(w.header, k)
		}
		w.status = 0
		r.ServeHTTP(w, req)
	}
	duration := time.Since(start)
	println("done", N, "requests in", duration.String())

	// Force GC and write heap profile
	runtime.GC()
	f, err := os.Create("heap.prof")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := pprof.WriteHeapProfile(f); err != nil {
		panic(err)
	}
	println("wrote heap.prof")
}

type nopWriter struct {
	header http.Header
	status int
}

func newNopWriter() *nopWriter {
	return &nopWriter{header: make(http.Header)}
}

func (w *nopWriter) Header() http.Header       { return w.header }
func (w *nopWriter) Write([]byte) (int, error) { return 0, nil }
func (w *nopWriter) WriteHeader(status int)    { w.status = status }
