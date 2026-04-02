package bench

import (
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
			resp, err := benchClient.Get(server.URL + testsupport.APITenants)
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("GetTenantByID", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := i%10 + 1
			resp, err := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APITenantByID, id))
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("TenantResources", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tenantID := i % 10
			resp, err := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APITenantResources, tenantID))
			if err != nil {
				b.Fatalf("GET failed: %v", err)
			}
			readAndClose(resp)
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
			resp, err := benchClient.Post(server.URL+testsupport.APITenants, common.MimeJSON, bytes.NewReader(bts))
			if err != nil {
				b.Fatalf("POST failed: %v", err)
			}
			readAndClose(resp)
		}
	})

	b.Run("UpdateTenant", func(b *testing.B) {
		tenant := testsupport.Tenant{TenantID: 1, Name: "updated-tenant", Plan: "platinum"}
		bts, _ := json.Marshal(tenant)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("PUT", server.URL+fmt.Sprintf(testsupport.APITenantByID, 1), bytes.NewReader(bts))
			req.Header.Set(common.HeaderContentType, common.MimeJSON)
			resp, err := benchClient.Do(req)
			if err != nil {
				b.Fatalf("PUT failed: %v", err)
			}
			readAndClose(resp)
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
					resp, _ := benchClient.Post(server.URL+testsupport.APITenants, common.MimeJSON, bytes.NewReader(bts))
					if resp != nil {
						readAndClose(resp)
					}
				case 1: // Read
					resp, _ := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APITenantByID, tenantID))
					if resp != nil {
						readAndClose(resp)
					}
				case 2: // List resources
					resp, _ := benchClient.Get(server.URL + fmt.Sprintf(testsupport.APITenantResources, tenantID))
					if resp != nil {
						readAndClose(resp)
					}
				case 3: // Update
					tenant := testsupport.Tenant{TenantID: int32(tenantID), Name: "updated", Plan: "gold"}
					bts, _ := json.Marshal(tenant)
					req, _ := http.NewRequest("PUT", server.URL+fmt.Sprintf(testsupport.APITenantByID, tenantID), bytes.NewReader(bts))
					req.Header.Set(common.HeaderContentType, common.MimeJSON)
					resp, _ := benchClient.Do(req)
					if resp != nil {
						readAndClose(resp)
					}
				}
			}
		})
	})
}
