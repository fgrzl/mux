package forwardheaders

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// BenchmarkForwardedHeaders_Invoke measures the overhead of the middleware Invoke method
// across common header combinations and trust configurations.
func BenchmarkForwardedHeadersInvoke(b *testing.B) {
	const (
		benchURL   = "http://example.com/test"
		protoHTTPS = "https"

		hdrXForwardedProto = "X-Forwarded-Proto"
		hdrXForwardedFor   = "X-Forwarded-For"
		hdrForwarded       = "Forwarded"
		hdrXForwardedHost  = "X-Forwarded-Host"
		hdrXForwardedPort  = "X-Forwarded-Port"

		trustedCIDR1 = "10.0.0.0/8"
	)
	cases := []struct {
		name  string
		opts  Options
		setup func(r *http.Request)
	}{
		{
			name: "TrustAll_XForwardedProtoAndFor",
			opts: Options{TrustAll: true, RespectForwarded: true},
			setup: func(r *http.Request) {
				r.Header.Set(hdrXForwardedProto, protoHTTPS)
				r.Header.Set(hdrXForwardedFor, "203.0.113.10")
			},
		},
		{
			name: "TrustAll_RFCForwarded",
			opts: Options{TrustAll: true, RespectForwarded: true},
			setup: func(r *http.Request) {
				r.Header.Set(hdrForwarded, "for=203.0.113.20;proto=https;host=example.com")
			},
		},
		{
			name: "TrustedProxy_Applied",
			opts: Options{TrustAll: false, TrustedProxies: []string{trustedCIDR1}, RespectForwarded: true},
			setup: func(r *http.Request) {
				// immediate remote is trusted
				r.RemoteAddr = "10.1.2.3:1234"
				r.Header.Set(hdrXForwardedProto, protoHTTPS)
				r.Header.Set(hdrXForwardedHost, "app.example.com")
				r.Header.Set(hdrXForwardedPort, "443")
				r.Header.Set(hdrXForwardedFor, "203.0.113.30, 10.1.2.3")
			},
		},
		{
			name: "UntrustedProxy_Ignored",
			opts: Options{TrustAll: false, TrustedProxies: []string{trustedCIDR1}, RespectForwarded: true},
			setup: func(r *http.Request) {
				// immediate remote is NOT trusted, so headers are ignored
				r.RemoteAddr = "192.0.2.1:7777"
				r.Header.Set(hdrXForwardedProto, protoHTTPS)
				r.Header.Set(hdrXForwardedFor, "203.0.113.40")
			},
		},
		{
			name: "TrustedProxy_MultiHop",
			opts: Options{TrustAll: false, TrustedProxies: []string{trustedCIDR1, "192.168.0.0/16"}, RespectForwarded: true},
			setup: func(r *http.Request) {
				r.RemoteAddr = "10.9.8.7:8080" // trusted
				// client, private hop (trusted), proxy (trusted)
				r.Header.Set(hdrXForwardedFor, "198.51.100.60, 192.168.1.50, 10.9.8.7")
				r.Header.Set(hdrXForwardedProto, protoHTTPS)
				r.Header.Set(hdrXForwardedHost, "svc.example.com")
				r.Header.Set(hdrXForwardedPort, "443")
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			m := newForwardedHeadersMiddleware(tc.opts)
			// intentionally empty: benchmark focuses on middleware overhead only
			next := func(c routing.RouteContext) {
				// no-op
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
				rec := httptest.NewRecorder()
				if tc.setup != nil {
					tc.setup(req)
				}
				ctx := routing.NewRouteContext(rec, req)
				m.Invoke(ctx, next)
			}
		})
	}
}

// BenchmarkForwardedHeaders_RouterPipeline measures the middleware in a router pipeline.
func BenchmarkForwardedHeadersRouterPipeline(b *testing.B) {
	r := router.NewRouter()
	UseForwardedHeadersWithOptions(r, Options{TrustAll: true, RespectForwarded: true})
	r.GET("/test", func(c routing.RouteContext) { c.NoContent() })

	b.Run("TrustAll_XForwarded", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
			req.Header.Set("X-Forwarded-Proto", "https")
			req.Header.Set("X-Forwarded-Host", "api.example.com")
			req.Header.Set("X-Forwarded-Port", "443")
			req.Header.Set("X-Forwarded-For", "203.0.113.90")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
		}
	})

	b.Run("RFCForwarded", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
			req.Header.Set("Forwarded", "for=203.0.113.100;proto=https;host=gw.example.com")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
		}
	})
}
