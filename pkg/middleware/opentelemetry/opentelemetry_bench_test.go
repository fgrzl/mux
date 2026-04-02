package opentelemetry

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"go.opentelemetry.io/otel"
	noop "go.opentelemetry.io/otel/trace/noop"
)

// Ensure OpenTelemetry uses a no-op tracer provider during benchmarks to avoid
// external exporter overhead influencing results.
func init() {
	otel.SetTracerProvider(noop.NewTracerProvider())
}

func benchOtel(b *testing.B, pooled bool) {
	var r *router.Router
	if pooled {
		r = router.NewRouter(router.WithContextPooling())
	} else {
		r = router.NewRouter()
	}
	UseOpenTelemetry(r)
	rg := r.NewRouteGroup("")
	rg.GET("/ok", func(c routing.RouteContext) {
		_, _ = c.Response().Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}
}

func BenchmarkOpenTelemetry(b *testing.B) {
	b.Run("nonpool", func(b *testing.B) { benchOtel(b, false) })
	b.Run("pool", func(b *testing.B) { benchOtel(b, true) })
}
