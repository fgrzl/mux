package router_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/bench"
	"github.com/fgrzl/mux/pkg/middleware/compression"
	"github.com/fgrzl/mux/pkg/middleware/logging"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// BenchmarkPipeline_E2E_Pool measures a full ServeHTTP including logging and compression middleware with context pooling.
func BenchmarkPipelineE2EPool(b *testing.B) {
	r := router.NewRouter(router.WithContextPooling())
	logging.UseLogging(r)
	compression.UseCompression(r)
	rg := r.NewRouteGroup("")
	rg.GET("/files/*", func(c routing.RouteContext) { _, _ = c.Response().Write([]byte(strings.Repeat("x", 512))) })

	rr, req := bench.NewRecorderRequest(http.MethodGet, "/files/some/path/file.txt")
	req.Header.Set("Accept-Encoding", "gzip")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}

// BenchmarkPipeline_E2E_NonPool measures a full ServeHTTP including logging and compression middleware without pooling.
func BenchmarkPipelineE2ENonPool(b *testing.B) {
	r := router.NewRouter()
	logging.UseLogging(r)
	compression.UseCompression(r)
	rg := r.NewRouteGroup("")
	rg.GET("/files/*", func(c routing.RouteContext) { _, _ = c.Response().Write([]byte(strings.Repeat("x", 512))) })

	rr, req := bench.NewRecorderRequest(http.MethodGet, "/files/some/path/file.txt")
	req.Header.Set("Accept-Encoding", "gzip")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rr, req)
	}
}
