package router

import (
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
)

// BenchmarkNewRouteContextAlloc benchmarks allocating a fresh context per request via NewRouteContext.
func BenchmarkNewRouteContextAlloc(b *testing.B) {
	req := httptest.NewRequest("GET", "/users/42", nil)
	rr := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := routing.NewRouteContext(rr, req)
		_ = c
	}
}

// BenchmarkAcquireContextPool benchmarks reusing a context from the pool with AcquireContext/ReleaseContext.
func BenchmarkAcquireContextPool(b *testing.B) {
	req := httptest.NewRequest("GET", "/users/42", nil)
	rr := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := routing.AcquireContext(rr, req)
		routing.ReleaseContext(c)
	}
}
