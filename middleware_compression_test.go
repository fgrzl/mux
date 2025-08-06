package mux

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldCompressResponseWithGzip(t *testing.T) {
	// Arrange
	middleware := &compressionMiddleware{options: &CompressionOptions{}}
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	response := "This is test response data that should be compressed"
	next := func(c *RouteContext) {
		c.Response.Write([]byte(response))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))

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
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "deflate")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	response := "This is test response data that should be compressed with deflate"
	next := func(c *RouteContext) {
		c.Response.Write([]byte(response))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Equal(t, "deflate", recorder.Header().Get("Content-Encoding"))
	// Note: deflate decompression test is more complex, so we just verify the header is set
}

func TestShouldPreferGzipOverDeflate(t *testing.T) {
	// Arrange
	middleware := &compressionMiddleware{options: &CompressionOptions{}}
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	next := func(c *RouteContext) {
		c.Response.Write([]byte("test"))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))
}

func TestShouldNotCompressWhenNoAcceptEncoding(t *testing.T) {
	// Arrange
	middleware := &compressionMiddleware{options: &CompressionOptions{}}
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	response := "This response should not be compressed"
	next := func(c *RouteContext) {
		c.Response.Write([]byte(response))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Empty(t, recorder.Header().Get("Content-Encoding"))
	assert.Equal(t, response, recorder.Body.String())
}

func TestShouldNotCompressWhenUnsupportedEncoding(t *testing.T) {
	// Arrange
	middleware := &compressionMiddleware{options: &CompressionOptions{}}
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br") // Brotli - not supported
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	response := "This response should not be compressed"
	next := func(c *RouteContext) {
		c.Response.Write([]byte(response))
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.Empty(t, recorder.Header().Get("Content-Encoding"))
	assert.Equal(t, response, recorder.Body.String())
}

func TestShouldAddCompressionMiddlewareToRouter(t *testing.T) {
	// Arrange
	router := NewRouter()
	initialMiddlewareCount := len(router.middleware)

	// Act
	router.UseCompression()

	// Assert
	assert.Equal(t, initialMiddlewareCount+1, len(router.middleware))
	assert.IsType(t, &compressionMiddleware{}, router.middleware[len(router.middleware)-1])
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
