package routing

import (
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/stretchr/testify/require"
)

func TestDetachIndependence(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	// Acquire a pooled context
	c := AcquireContext(rr, req)
	require.NotNil(t, c)

	// set some fields
	c.SetService(common.ServiceKey("svc"), "value")
	params := AcquireRouteParams()
	params["id"] = "123"
	c.SetParams(params)

	// detach a non-pooled clone
	d := Detach(c)
	require.NotNil(t, d)

	// release the original back to pool
	ReleaseContext(c)

	// detached clone should retain copies of params and service
	v, ok := d.GetService(common.ServiceKey("svc"))
	require.True(t, ok)
	require.Equal(t, "value", v)
	pid, ok := d.Param("id")
	require.True(t, ok)
	require.Equal(t, "123", pid)

	// cleanup: detached is not pooled; nothing to release
}

func TestDetachUsableInGoroutine(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	c := AcquireContext(rr, req)
	require.NotNil(t, c)

	c.SetService(common.ServiceKey("svc"), "v2")
	params := AcquireRouteParams()
	params["k"] = "v"
	c.SetParams(params)

	d := Detach(c)
	require.NotNil(t, d)

	// simulate goroutine that outlives the request
	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx *DefaultRouteContext) {
		defer wg.Done()
		// access fields that would otherwise be cleared by ReleaseContext
		v, ok := ctx.GetService(common.ServiceKey("svc"))
		require.True(t, ok)
		require.Equal(t, "v2", v)
		val, ok := ctx.Param("k")
		require.True(t, ok)
		require.Equal(t, "v", val)
	}(d)

	// release original context while goroutine runs
	ReleaseContext(c)
	wg.Wait()
}
