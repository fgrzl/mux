package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/test/testsupport"
)

// BenchmarkPayloadSizes measures performance impact of different payload sizes.
func BenchmarkPayloadSizes(b *testing.B) {
	server := newBenchmarkServer(b)

	sizes := []struct {
		name  string
		count int
	}{
		{"Small/1", 1},
		{"Medium/10", 10},
		{"Large/100", 100},
		{"XLarge/500", 500},
	}

	for _, size := range sizes {
		b.Run("Create/"+size.name, func(b *testing.B) {
			var seq uint64
			resources := make([]testsupport.Resource, size.count)
			for i := range resources {
				resources[i] = testsupport.Resource{TenantID: 0, Name: fmt.Sprintf("r-%d", i), Type: "resource"}
			}

			b.ReportAllocs()
			bts, _ := json.Marshal(resources)
			b.SetBytes(int64(len(bts)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				n := atomic.AddUint64(&seq, 1)
				for j := range resources {
					resources[j].Name = fmt.Sprintf("r-%d-%d", n, j)
				}
				payload, _ := json.Marshal(resources)
				resp, err := benchClient.Post(server.URL+testsupport.APIBase+"/resources/bulk", common.MimeJSON, bytes.NewReader(payload))
				if err != nil {
					b.Fatalf("POST failed: %v", err)
				}
				if resp.StatusCode != 201 {
					b.Fatalf("unexpected status: %d", resp.StatusCode)
				}
				readAndClose(resp)
			}
		})
	}

	metadataSizes := []struct {
		name  string
		pairs int
	}{
		{"Tiny/1", 1},
		{"Small/5", 5},
		{"Medium/20", 20},
		{"Large/100", 100},
	}

	for _, size := range metadataSizes {
		b.Run("Metadata/"+size.name, func(b *testing.B) {
			metadata := make(map[string]string, size.pairs)
			for i := 0; i < size.pairs; i++ {
				metadata[fmt.Sprintf("key-%d", i)] = fmt.Sprintf("value-%d with some additional content", i)
			}
			body := map[string]any{"metadata": metadata}
			bts, _ := json.Marshal(body)

			b.ReportAllocs()
			b.SetBytes(int64(len(bts)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req, _ := http.NewRequest("PUT", server.URL+fmt.Sprintf(testsupport.APIResourceMetadata, 1), bytes.NewReader(bts))
				req.Header.Set(common.HeaderContentType, common.MimeJSON)
				resp, err := benchClient.Do(req)
				if err != nil {
					b.Fatalf("PUT failed: %v", err)
				}
				readAndClose(resp)
			}
		})
	}
}
