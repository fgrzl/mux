package compression

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/fgrzl/mux/pkg/bench"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

func benchCompression(b *testing.B, encoding string, pooled bool, sizes []int) {
	for _, size := range sizes {
		// sub-bench names like: gzip_pool_size_1024
		b.Run(encoding+func() string {
			if pooled {
				return "_pool_size_"
			}
			return "_nonpool_size_"
		}()+strconv.Itoa(size), func(b *testing.B) {
			var r *router.Router
			if pooled {
				r = router.NewRouter(router.WithContextPooling())
			} else {
				r = router.NewRouter()
			}
			UseCompression(r)
			rg := r.NewRouteGroup("")
			payload := strings.Repeat("x", size)
			rg.GET("/data", func(c routing.RouteContext) {
				c.Response().Write([]byte(payload))
			})
			_, req := bench.NewRecorderRequest(http.MethodGet, "/data")
			req.Header.Set("Accept-Encoding", encoding)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rr := httptest.NewRecorder()
				r.ServeHTTP(rr, req)
			}
		})
	}
}

func BenchmarkCompressionGzip(b *testing.B) {
	sizes := []int{0, 128, 1024, 8192, 32768}
	benchCompression(b, "gzip", false, sizes)
	benchCompression(b, "gzip", true, sizes)
}

func BenchmarkCompressionDeflate(b *testing.B) {
	sizes := []int{0, 128, 1024, 8192, 32768}
	benchCompression(b, "deflate", false, sizes)
	benchCompression(b, "deflate", true, sizes)
}
