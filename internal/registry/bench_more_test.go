package registry

import (
	"testing"

	"github.com/fgrzl/mux/internal/routing"
)

// Benchmark the single traversal approach: use FindNodeInto and inspect node
func BenchmarkRouteRegistrySingleTraversalCatchAll(b *testing.B) {
	r := NewRouteRegistry()
	opts := &routing.RouteOptions{Method: "GET", Pattern: "/catch/**", Handler: func(c routing.RouteContext) {}}
	r.Register("/catch/**", "GET", opts)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := r.FindNode("/catch/any/segments/here")
		_ = node
	}
}
