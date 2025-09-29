package compression

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// ---- Functional Options ----

// CompressionOptions configures the compression middleware behavior.
type CompressionOptions struct{}

// CompressionOption is a function type for configuring compression options.
type CompressionOption func(*CompressionOptions)

// UseCompression adds response compression middleware that supports gzip and deflate encoding.
func UseCompression(rtr *router.Router, opts ...CompressionOption) {
	options := &CompressionOptions{}
	for _, opt := range opts {
		opt(options)
	}
	// Use the exported API to register middleware so we don't rely on
	// unexported router internals.
	rtr.Use(&compressionMiddleware{options: options})
}

// ---- Middleware ----

// compressionMiddleware handles response compression using gzip or deflate.
type compressionMiddleware struct {
	options *CompressionOptions
}

// compressionWriter wraps an http.ResponseWriter to provide compression.
type compressionWriter struct {
	w http.ResponseWriter
	c io.WriteCloser
}

// Write implements io.Writer, writing compressed data to the underlying writer.
func (cw *compressionWriter) Write(p []byte) (int, error) {
	return cw.c.Write(p)
}

// Header returns the header map of the underlying ResponseWriter.
func (cw *compressionWriter) Header() http.Header {
	return cw.w.Header()
}

// WriteHeader sends an HTTP response header with the provided status code.
func (cw *compressionWriter) WriteHeader(statusCode int) {
	cw.w.WriteHeader(statusCode)
}

// Invoke implements the Middleware interface, applying compression based on Accept-Encoding headers.
func (m *compressionMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	// Never attempt to compress upgraded WebSocket connections; compression wrappers
	// can break hijacking and raw upgrade semantics.
	if strings.ToLower(c.Request().Header.Get("Upgrade")) == "websocket" {
		next(c)
		return
	}

	acceptEncoding := c.Request().Header.Get(common.HeaderAcceptEncoding)
	if acceptEncoding == "" {
		next(c)
		return
	}

	var compressor io.WriteCloser
	if strings.Contains(acceptEncoding, "gzip") {
		c.Response().Header().Set(common.HeaderContentEncoding, "gzip")
		c.Response().Header().Add(common.HeaderVary, common.HeaderAcceptEncoding)
		c.Response().Header().Del(common.HeaderContentLength)
		gw := gzipPool.Get().(*gzip.Writer)
		gw.Reset(c.Response())
		compressor = gw
	} else if strings.Contains(acceptEncoding, "deflate") {
		c.Response().Header().Set(common.HeaderContentEncoding, "deflate")
		c.Response().Header().Add(common.HeaderVary, common.HeaderAcceptEncoding)
		c.Response().Header().Del(common.HeaderContentLength)
		dw := deflatePool.Get().(*flate.Writer)
		dw.Reset(c.Response())
		compressor = dw
	} else {
		next(c)
		return
	}

	c.SetResponse(&compressionWriter{
		w: c.Response(),
		c: compressor,
	})
	defer func() {
		// Always close to flush
		_ = compressor.Close()
		switch z := compressor.(type) {
		case *gzip.Writer:
			gzipPool.Put(z)
		case *flate.Writer:
			deflatePool.Put(z)
		}
	}()
	next(c)
}

// Pools for gzip and deflate writers to reduce per-request allocations.
var gzipPool = sync.Pool{New: func() any { return gzip.NewWriter(io.Discard) }}
var deflatePool = sync.Pool{New: func() any {
	w, _ := flate.NewWriter(io.Discard, flate.DefaultCompression)
	return w
}}
