package registry

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/internal/routing"
)

// BenchmarkLoadIntoSlice benchmarks the optimized slice-based parameter extraction
func BenchmarkLoadIntoSlice(b *testing.B) {
	r := NewRouteRegistry()
	opt := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}
	r.Register("/users/{id}/posts/{postId}", http.MethodGet, opt)

	params := routing.AcquireParams()
	defer routing.ReleaseParams(params)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.LoadIntoSlice("/users/123/posts/456", http.MethodGet, params)
	}
}

// BenchmarkLoadIntoMap benchmarks the traditional map-based parameter extraction
func BenchmarkLoadIntoMap(b *testing.B) {
	r := NewRouteRegistry()
	opt := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}
	r.Register("/users/{id}/posts/{postId}", http.MethodGet, opt)

	params := routing.AcquireRouteParams()
	defer routing.ReleaseRouteParams(params)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.LoadInto("/users/123/posts/456", http.MethodGet, params)
	}
}

// BenchmarkLoadIntoSliceSingleParam benchmarks single parameter extraction
func BenchmarkLoadIntoSliceSingleParam(b *testing.B) {
	r := NewRouteRegistry()
	opt := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}
	r.Register("/users/{id}", http.MethodGet, opt)

	params := routing.AcquireParams()
	defer routing.ReleaseParams(params)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.LoadIntoSlice("/users/123", http.MethodGet, params)
	}
}

// BenchmarkLoadIntoMapSingleParam benchmarks single parameter extraction with map
func BenchmarkLoadIntoMapSingleParam(b *testing.B) {
	r := NewRouteRegistry()
	opt := &routing.RouteOptions{Handler: func(c routing.RouteContext) {}}
	r.Register("/users/{id}", http.MethodGet, opt)

	params := routing.AcquireRouteParams()
	defer routing.ReleaseRouteParams(params)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.LoadInto("/users/123", http.MethodGet, params)
	}
}
