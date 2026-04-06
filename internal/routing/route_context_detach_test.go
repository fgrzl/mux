package routing

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/fgrzl/mux/internal/common"
	"github.com/stretchr/testify/require"
)

func TestDetachIndependence(t *testing.T) {
	// Arrange
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	c := AcquireContext(rr, req)
	require.NotNil(t, c)

	c.SetService(common.ServiceKey("svc"), "value")
	params := AcquireParams()
	params.Set("id", "123")
	c.SetParamsSlice(params)

	// Act
	d := Detach(c)
	require.NotNil(t, d)

	ReleaseContext(c)

	// Assert
	v, ok := d.GetService(common.ServiceKey("svc"))
	require.True(t, ok)
	require.Equal(t, "value", v)
	pid, ok := d.Param("id")
	require.True(t, ok)
	require.Equal(t, "123", pid)
}

func TestDetachUsableInGoroutine(t *testing.T) {
	// Arrange
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	c := AcquireContext(rr, req)
	require.NotNil(t, c)

	c.SetService(common.ServiceKey("svc"), "v2")
	params := AcquireParams()
	params.Set("k", "v")
	c.SetParamsSlice(params)

	d := Detach(c)
	require.NotNil(t, d)

	var wg sync.WaitGroup
	wg.Add(1)

	// Act
	go func(ctx *DefaultRouteContext) {
		defer wg.Done()
		// Assert (within goroutine)
		v, ok := ctx.GetService(common.ServiceKey("svc"))
		require.True(t, ok)
		require.Equal(t, "v2", v)
		val, ok := ctx.Param("k")
		require.True(t, ok)
		require.Equal(t, "v", val)
	}(d)

	ReleaseContext(c)
	wg.Wait()
}

func TestDetachShouldNotWriteToOriginalResponse(t *testing.T) {
	// Arrange
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	c := AcquireContext(rr, req)
	require.NotNil(t, c)

	d := Detach(c)
	require.NotNil(t, d)

	// Act
	d.OK(map[string]string{"status": "background"})
	ReleaseContext(c)

	// Assert
	require.Empty(t, rr.Body.String())
	require.Empty(t, rr.Header().Get(common.HeaderContentType))
}

func TestDetachShouldUseBackgroundContext(t *testing.T) {
	// Arrange
	baseCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil).WithContext(baseCtx)
	c := AcquireContext(rr, req)
	require.NotNil(t, c)

	d := Detach(c)
	require.NotNil(t, d)

	// Act
	cancel()
	ReleaseContext(c)

	// Assert
	require.NoError(t, d.Err())
	require.NoError(t, d.Request().Context().Err())
}
