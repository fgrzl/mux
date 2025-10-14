package servicelocator

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

func benchServiceLocator(b *testing.B, pooled bool) {
	var r *router.Router
	if pooled {
		r = router.NewRouter(router.WithContextPooling())
	} else {
		r = router.NewRouter()
	}

	// Register some services to be injected into the RouteContext.
	UseServices(r, WithService("svc1", "value1"), WithService("svc2", 123))

	rg := r.NewRouteGroup("")
	rg.GET("/ok", func(c routing.RouteContext) {
		// Access services to ensure the middleware work path is exercised.
		_, _ = c.GetService("svc1")
		_, _ = c.GetService("svc2")
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

func BenchmarkServiceLocator(b *testing.B) {
	b.Run("nonpool", func(b *testing.B) { benchServiceLocator(b, false) })
	b.Run("pool", func(b *testing.B) { benchServiceLocator(b, true) })
}
