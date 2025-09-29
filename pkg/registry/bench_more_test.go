package registry

import (
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
)

// Benchmark that simulates the "old" behavior: call LoadDetailedInto (to get Allow)
// then call Load to get options when needed. This isolates repeated traversal cost.
func BenchmarkRouteRegistry_OldDoubleTraversal_CatchAll(b *testing.B) {
	r := NewRouteRegistry()
	opts := &routing.RouteOptions{Method: "GET", Pattern: "/catch/**", Handler: func(c routing.RouteContext) {}}
	r.Register("/catch/**", "GET", opts)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// First do detailed to get Allow/method info
		tmp := make(routing.RouteParams, 2)
		_, det := r.LoadDetailedInto("/catch/any/segments/here", "GET", tmp)
		_ = det
		// Then call Load (simulates older code paths)
		_, _, _ = r.Load("/catch/any/segments/here", "GET")
	}
}

// Benchmark the single traversal approach: use FindNodeInto and inspect node
func BenchmarkRouteRegistry_SingleTraversal_CatchAll(b *testing.B) {
	r := NewRouteRegistry()
	opts := &routing.RouteOptions{Method: "GET", Pattern: "/catch/**", Handler: func(c routing.RouteContext) {}}
	r.Register("/catch/**", "GET", opts)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := make(routing.RouteParams, 2)
		node := r.FindNodeInto("/catch/any/segments/here", tmp)
		_ = node
	}
}
