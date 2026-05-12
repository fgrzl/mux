package bench

import (
	"context"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/fgrzl/mux/internal/common"
	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkMultiTenant simulates multi-tenant access patterns.
func BenchmarkMultiTenant(b *testing.B) {
	server := newBenchmarkServer(b)

	b.Run("ListAllTenants", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := benchClientDo(b, http.MethodGet, server.URL+testsupport.APITenants, nil, "")
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
		}
	})

	b.Run("GetTenantByID", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := i%10 + 1
			_, err := benchClientDo(b, http.MethodGet, server.URL+fmt.Sprintf(testsupport.APITenantByID, id), nil, "")
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
		}
	})

	b.Run("TenantResources", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tenantID := i % 10
			_, err := benchClientDo(b, http.MethodGet, server.URL+fmt.Sprintf(testsupport.APITenantResources, tenantID), nil, "")
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
		}
	})

	b.Run("CreateTenant", func(b *testing.B) {
		var seq uint64
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			n := atomic.AddUint64(&seq, 1)
			tenant := testsupport.Tenant{TenantID: int32(1000 + n), Name: fmt.Sprintf("tenant-%d", n), Plan: "gold"}
			bts, _ := json.Marshal(tenant)
			_, err := benchClientDo(b, http.MethodPost, server.URL+testsupport.APITenants, bytes.NewReader(bts), common.MimeJSON)
			if err != nil {
				b.Fatalf("POST failed: %v", err)
			}
		}
	})

	b.Run("UpdateTenant", func(b *testing.B) {
		tenant := testsupport.Tenant{TenantID: 1, Name: "updated-tenant", Plan: "platinum"}
		bts, _ := json.Marshal(tenant)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequestWithContext(context.Background(), "PUT", server.URL+fmt.Sprintf(testsupport.APITenantByID, 1), bytes.NewReader(bts))
			req.Header.Set(common.HeaderContentType, common.MimeJSON)
			resp, err := benchClient.Do(req)
			if err != nil {
				b.Fatalf("PUT failed: %v", err)
			}
			_ = readAndClose(resp)
		}
	})

	b.Run("TenantCRUD/Parallel", func(b *testing.B) {
		var seq uint64
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			local := 0
			for pb.Next() {
				local++
				tenantID := local % 10

				switch local % 4 {
				case 0: // Create
					n := atomic.AddUint64(&seq, 1)
					tenant := testsupport.Tenant{TenantID: int32(10000 + n), Name: fmt.Sprintf("t-%d", n), Plan: "test"}
					bts, _ := json.Marshal(tenant)
					_, _ = benchClientDo(b, http.MethodPost, server.URL+testsupport.APITenants, bytes.NewReader(bts), common.MimeJSON)
				case 1: // Read
					_, _ = benchClientDo(b, http.MethodGet, server.URL+fmt.Sprintf(testsupport.APITenantByID, tenantID), nil, "")
				case 2: // List resources
					_, _ = benchClientDo(b, http.MethodGet, server.URL+fmt.Sprintf(testsupport.APITenantResources, tenantID), nil, "")
				case 3: // Update
					tenant := testsupport.Tenant{TenantID: int32(tenantID), Name: "updated", Plan: "gold"}
					bts, _ := json.Marshal(tenant)
					req, _ := http.NewRequestWithContext(context.Background(), "PUT", server.URL+fmt.Sprintf(testsupport.APITenantByID, tenantID), bytes.NewReader(bts))
					req.Header.Set(common.HeaderContentType, common.MimeJSON)
					resp, _ := benchClient.Do(req)
					if resp != nil {
						_ = readAndClose(resp)
					}
				}
			}
		})
	})
}
