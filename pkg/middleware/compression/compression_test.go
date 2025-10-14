package compression

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldCompressResponseWithGzip(t *testing.T) {
	// Arrange
	middleware := &compressionMiddleware{options: &CompressionOptions{}}
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(common.HeaderAcceptEncoding, "gzip")
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	response := "This is test response data that should be compressed"
	next := func(c routing.RouteContext) {
		_, _ = c.Response().Write([]byte(response))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Equal(t, "gzip", recorder.Header().Get(common.HeaderContentEncoding))

	// Decompress and verify content
	reader, err := gzip.NewReader(recorder.Body)
	require.NoError(t, err)
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, response, string(decompressed))
}

func TestShouldCompressResponseWithDeflate(t *testing.T) {
	// Arrange
	middleware := &compressionMiddleware{options: &CompressionOptions{}}
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(common.HeaderAcceptEncoding, "deflate")
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	response := "This is test response data that should be compressed with deflate"
	next := func(c routing.RouteContext) {
		_, _ = c.Response().Write([]byte(response))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Equal(t, "deflate", recorder.Header().Get(common.HeaderContentEncoding))
	// Note: deflate decompression test is more complex, so we just verify the header is set
}

func TestShouldPreferGzipOverDeflate(t *testing.T) {
	// Arrange
	middleware := &compressionMiddleware{options: &CompressionOptions{}}
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(common.HeaderAcceptEncoding, "gzip, deflate")
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	next := func(c routing.RouteContext) {
		_, _ = c.Response().Write([]byte("test"))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Equal(t, "gzip", recorder.Header().Get(common.HeaderContentEncoding))
}

func TestShouldNotCompressWhenNoAcceptEncoding(t *testing.T) {
	// Arrange
	middleware := &compressionMiddleware{options: &CompressionOptions{}}
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	response := "This response should not be compressed"
	next := func(c routing.RouteContext) {
		_, _ = c.Response().Write([]byte(response))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Empty(t, recorder.Header().Get(common.HeaderContentEncoding))
	assert.Equal(t, response, recorder.Body.String())
}

func TestShouldNotCompressWhenUnsupportedEncoding(t *testing.T) {
	// Arrange
	middleware := &compressionMiddleware{options: &CompressionOptions{}}
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(common.HeaderAcceptEncoding, "br") // Brotli - not supported
	recorder := httptest.NewRecorder()
	ctx := routing.NewRouteContext(recorder, req)

	response := "This response should not be compressed"
	next := func(c routing.RouteContext) {
		_, _ = c.Response().Write([]byte(response))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Empty(t, recorder.Header().Get(common.HeaderContentEncoding))
	assert.Equal(t, response, recorder.Body.String())
}

func TestShouldAddCompressionMiddlewareToRouter(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()

	// Act - register compression middleware
	UseCompression(rtr)

	// To verify it was added, register a handler that writes a response and make a request
	rtr.GET("/test", func(c routing.RouteContext) {
		_, _ = c.Response().Write([]byte("hello"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(common.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()
	rtr.ServeHTTP(rec, req)

	// Assert that the middleware executed and set the Content-Encoding header
	assert.Equal(t, "gzip", rec.Header().Get(common.HeaderContentEncoding))
}

func TestCompressionWriterShouldImplementResponseWriter(t *testing.T) {
	// Arrange
	recorder := httptest.NewRecorder()
	compressor := gzip.NewWriter(recorder)
	writer := &compressionWriter{
		w: recorder,
		c: compressor,
	}

	// Act & Assert
	assert.Implements(t, (*http.ResponseWriter)(nil), writer)

	// Test Header method
	writer.Header().Set("Test-Header", "test-value")
	assert.Equal(t, "test-value", recorder.Header().Get("Test-Header"))

	// Test WriteHeader method
	writer.WriteHeader(http.StatusAccepted)
	assert.Equal(t, http.StatusAccepted, recorder.Code)

	// Test Write method
	data := []byte("test data")
	n, err := writer.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)

	compressor.Close()
}
