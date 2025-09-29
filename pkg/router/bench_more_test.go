package router

import (
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
)

// Benchmark allocating a fresh context per request via NewRouteContext
func Benchmark_NewRouteContext_Alloc(b *testing.B) {
	req := httptest.NewRequest("GET", "/users/42", nil)
	rr := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := routing.NewRouteContext(rr, req)
		_ = c
	}
}

// Benchmark reusing a context from the pool with AcquireContext/ReleaseContext
func Benchmark_AcquireContext_Pool(b *testing.B) {
	req := httptest.NewRequest("GET", "/users/42", nil)
	rr := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := routing.AcquireContext(rr, req)
		routing.ReleaseContext(c)
	}
}
