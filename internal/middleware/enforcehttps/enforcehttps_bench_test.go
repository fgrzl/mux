package enforcehttps

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/internal/middlewarebench"
	"github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
)

// BenchmarkEnforceHTTPSInvoke measures the middleware overhead in isolation.
func BenchmarkEnforceHTTPSInvoke(b *testing.B) {
	middleware := &enforceHTTPSMiddleware{}

	b.Run("HTTP_Redirect", func(b *testing.B) {
		middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, func(r *http.Request) {
			r.URL.Scheme = "http"
			r.Host = "example.com"
		})
	})

	b.Run("HTTPS_PassThrough", func(b *testing.B) {
		middlewarebench.BenchmarkMiddlewareInvoke(b, middleware.Invoke, func(r *http.Request) {
			r.URL.Scheme = "https"
			r.Host = "example.com"
		})
	})
}

// BenchmarkEnforceHTTPSRouterPipeline measures the middleware in a real router pipeline.
func BenchmarkEnforceHTTPSRouterPipeline(b *testing.B) {
	rtr := router.NewRouter()
	rtr.Use(&enforceHTTPSMiddleware{})
	rtr.GET("/test", func(c routing.RouteContext) {
		c.Response().WriteHeader(http.StatusOK)
	})

	b.Run("HTTP_Redirect", func(b *testing.B) {
		middlewarebench.BenchmarkMiddlewareRouterPipeline(b, rtr, http.MethodGet, "http://example.com/test", func(r *http.Request) {
			r.URL.Scheme = "http"
		})
	})

	b.Run("HTTPS_PassThrough", func(b *testing.B) {
		middlewarebench.BenchmarkMiddlewareRouterPipeline(b, rtr, http.MethodGet, "https://example.com/test", func(r *http.Request) {
			r.URL.Scheme = "https"
		})
	})
}
