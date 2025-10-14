package exportcontrol

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/middlewarebench"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

const (
	benchURL         = "http://example.com/test"
	benchRemoteAddr  = "192.168.1.1:12345"
	benchForwardedIP = "203.0.113.1"
	benchXRealIP     = "203.0.113.2"
	headerXFF        = common.HeaderXForwardedFor
	headerXRealIP    = common.HeaderXRealIP
)

// BenchmarkExportControlInvoke measures the middleware overhead in isolation.
func BenchmarkExportControlInvoke(b *testing.B) {
	middleware := &exportControlMiddleware{
		options: &ExportControlOptions{DB: nil},
	}

	b.Run("NoDB_PassThrough", func(b *testing.B) {
		middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, func(r *http.Request) {
			r.RemoteAddr = benchRemoteAddr
		})
	})

	b.Run("NoDB_WithHeaders", func(b *testing.B) {
		middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, func(r *http.Request) {
			r.RemoteAddr = benchRemoteAddr
			r.Header.Set(headerXFF, benchForwardedIP)
		})
	})

	b.Run("InvalidIP", func(b *testing.B) {
		middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, func(r *http.Request) {
			r.RemoteAddr = "invalid-ip"
		})
	})
}

// BenchmarkExportControlRouterPipeline measures the middleware in a real router pipeline.
func BenchmarkExportControlRouterPipeline(b *testing.B) {
	setupRouter := func(rtr *router.Router) {
		UseExportControl(rtr, WithGeoIPDatabase(nil))
		rtr.GET("/test", func(c routing.RouteContext) {
			_, _ = c.Response().Write([]byte("ok"))
		})
	}

	cases := []middlewarebench.RouterPipelineCase{
		{Name: "Pipeline_NoDB", Pooled: false, SetupRequest: func(r *http.Request) {
			r.RemoteAddr = benchRemoteAddr
		}},
		{Name: "Pipeline_Pooled", Pooled: true, SetupRequest: func(r *http.Request) {
			r.RemoteAddr = benchRemoteAddr
		}},
		{Name: "Pipeline_WithForwardedFor", Pooled: false, SetupRequest: func(r *http.Request) {
			r.RemoteAddr = benchRemoteAddr
			r.Header.Set(headerXFF, "203.0.113.1, 192.168.1.1")
		}},
	}

	middlewarebench.BenchmarkRouterPipelines(b, setupRouter, cases, http.MethodGet, benchURL)
}

// BenchmarkGetRealIP measures the IP extraction helper performance.
func BenchmarkGetRealIP(b *testing.B) {
	b.Run("RemoteAddr", func(b *testing.B) {
		req := httptest.NewRequest(http.MethodGet, benchURL, nil)
		req.RemoteAddr = benchRemoteAddr

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = getRealIP(req)
		}
	})

	b.Run("XForwardedFor", func(b *testing.B) {
		req := httptest.NewRequest(http.MethodGet, benchURL, nil)
		req.RemoteAddr = benchRemoteAddr
		req.Header.Set(headerXFF, "203.0.113.1, 192.168.1.1")

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = getRealIP(req)
		}
	})

	b.Run("XRealIP", func(b *testing.B) {
		req := httptest.NewRequest(http.MethodGet, benchURL, nil)
		req.RemoteAddr = benchRemoteAddr
		req.Header.Set(headerXRealIP, benchXRealIP)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = getRealIP(req)
		}
	})
}
