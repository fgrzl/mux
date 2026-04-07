package middlewarebench

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
)

// MiddlewareInvokeFunc is the signature for middleware Invoke methods.
type MiddlewareInvokeFunc func(routing.RouteContext, router.HandlerFunc)

// BenchmarkMiddlewareInvoke benchmarks a middleware's Invoke method in isolation.
// It measures the overhead of the middleware without router pipeline costs.
func BenchmarkMiddlewareInvoke(b *testing.B, invoke MiddlewareInvokeFunc, setupRequest func(*http.Request)) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	if setupRequest != nil {
		setupRequest(req)
	}
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	next := func(c routing.RouteContext) {
		// noop - measuring middleware overhead only
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		invoke(ctx, next)
	}
}

// BenchmarkMiddlewareRouterPipeline benchmarks a middleware through a full router pipeline.
// This measures real-world performance including routing overhead.
func BenchmarkMiddlewareRouterPipeline(b *testing.B, rtr *router.Router, method, path string, setupRequest func(*http.Request)) {
	req := httptest.NewRequest(method, path, nil)
	if setupRequest != nil {
		setupRequest(req)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		rtr.ServeHTTP(rec, req)
	}
}

// RouterPipelineCase defines a test case for router pipeline benchmarks.
type RouterPipelineCase struct {
	Name         string
	Pooled       bool
	SetupRequest func(*http.Request)
}

// BenchmarkRouterPipelines runs multiple router pipeline scenarios (pooled/non-pooled).
func BenchmarkRouterPipelines(b *testing.B, setupRouter func(*router.Router), cases []RouterPipelineCase, method, path string) {
	for _, tc := range cases {
		b.Run(tc.Name, func(b *testing.B) {
			var rtr *router.Router
			if tc.Pooled {
				rtr = router.NewRouter(router.WithContextPooling())
			} else {
				rtr = router.NewRouter()
			}
			setupRouter(rtr)
			BenchmarkMiddlewareRouterPipeline(b, rtr, method, path, tc.SetupRequest)
		})
	}
}
