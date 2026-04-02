package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/test/testsupport"
	"github.com/google/uuid"
)

// BenchmarkHTTPMethods benchmarks all HTTP methods to establish baseline costs.
func BenchmarkHTTPMethods(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("GET/StaticList", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + testsupport.APIResources)
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				b.Fatalf("unexpected status: %d", resp.StatusCode)
			}
			if err := readAndClose(resp); err != nil {
				b.Fatalf("read body failed: %v", err)
			}
		}
	})

	b.Run("GET/SingleResource", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APIResourceByID, 1))
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				b.Fatalf("unexpected status: %d", resp.StatusCode)
			}
			if err := readAndClose(resp); err != nil {
				b.Fatalf("read body failed: %v", err)
			}
		}
	})

	b.Run("HEAD/ResourceExists", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest(http.MethodHead, server.URL+fmt.Sprintf(testsupport.APIResourceByID, 1), nil)
			resp, err := benchClient.Do(req)
			if err != nil {
				b.Fatalf("HEAD failed: %v", err)
			}
			if resp.StatusCode != http.StatusNoContent {
				b.Fatalf("unexpected status: %d", resp.StatusCode)
			}
			resp.Body.Close()
		}
	})

	b.Run("POST/CreateSingle", func(b *testing.B) {
		var seq uint64
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			n := atomic.AddUint64(&seq, 1)
			resource := testsupport.Resource{TenantID: 0, Name: fmt.Sprintf("bench-%d", n), Type: "resource"}
			bts, _ := json.Marshal([]testsupport.Resource{resource})
			resp, err := benchClient.Post(server.URL+testsupport.APIBase+"/resources/bulk", common.MimeJSON, bytes.NewReader(bts))
			if err != nil {
				b.Fatalf("POST failed: %v", err)
			}
			if resp.StatusCode != http.StatusCreated {
				b.Fatalf("unexpected status: %d", resp.StatusCode)
			}
			if err := readAndClose(resp); err != nil {
				b.Fatalf("read body failed: %v", err)
			}
		}
	})

	b.Run("PUT/UpdateMetadata", func(b *testing.B) {
		body := map[string]any{"metadata": map[string]string{"key": "value", "updated": "true"}}
		bts, _ := json.Marshal(body)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest(http.MethodPut, server.URL+fmt.Sprintf(testsupport.APIResourceMetadata, 1), bytes.NewReader(bts))
			req.Header.Set(common.HeaderContentType, common.MimeJSON)
			resp, err := benchClient.Do(req)
			if err != nil {
				b.Fatalf("PUT failed: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				b.Fatalf("unexpected status: %d", resp.StatusCode)
			}
			if err := readAndClose(resp); err != nil {
				b.Fatalf("read body failed: %v", err)
			}
		}
	})

	b.Run("DELETE/Tenant", func(b *testing.B) {
		// Pre-create tenants to delete
		for i := 0; i < b.N+100; i++ {
			testsupport.Service.PutTenant(&testsupport.Tenant{
				TenantID: int32(1000 + i),
				Name:     fmt.Sprintf("bench-tenant-%d", i),
				Plan:     "test",
			})
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest(http.MethodDelete, server.URL+fmt.Sprintf(testsupport.APITenantByID, 1000+i), nil)
			resp, err := benchClient.Do(req)
			if err != nil {
				b.Fatalf("DELETE failed: %v", err)
			}
			if resp.StatusCode != http.StatusNoContent {
				b.Fatalf("unexpected status: %d", resp.StatusCode)
			}
			resp.Body.Close()
		}
	})
}

// BenchmarkRouteMatching measures routing overhead for different path patterns.
func BenchmarkRouteMatching(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("StaticPath", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + testsupport.APIResources)
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("SingleParam", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APIResourceByID, 5))
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("NestedParams", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APITenantResources, 1))
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("UUIDParam", func(b *testing.B) {
		id := uuid.New()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APIItemsUUID, id.String()))
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("QueryParams", func(b *testing.B) {
		resources := testsupport.Service.ListResources(0)
		if len(resources) == 0 {
			b.Fatal("no resources")
		}
		name := url.QueryEscape(resources[0].Name)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + testsupport.APIBase + "/resources/search?name=" + name + "&type=resource")
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("MultipleQueryParams", func(b *testing.B) {
		u1, u2 := uuid.New(), uuid.New()
		query := fmt.Sprintf("?ids=1&ids=2&ids=3&uuids=%s&uuids=%s", u1.String(), u2.String())
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := benchClient.Get(server.URL + testsupport.APIBase + "/filter" + query)
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
	})
}

// BenchmarkHeaders measures header parsing and processing overhead.
func BenchmarkHeaders(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("EchoHeader", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest(http.MethodGet, server.URL+testsupport.APIHeadersEcho, nil)
			req.Header.Set(common.HeaderXEcho, "benchmark-value")
			resp, err := benchClient.Do(req)
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("ManyHeaders", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest(http.MethodGet, server.URL+testsupport.APIResources, nil)
			for j := 0; j < 20; j++ {
				req.Header.Set(fmt.Sprintf("X-Custom-Header-%d", j), fmt.Sprintf("value-%d", j))
			}
			resp, err := benchClient.Do(req)
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
	})
}
