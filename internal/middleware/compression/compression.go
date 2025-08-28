package compression

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/fgrzl/mux/internal/router"
	routerpkg "github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
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
	rtr.middleware = append(rtr.middleware, &compressionMiddleware{options: options})
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
func (m *compressionMiddleware) Invoke(c routing.RouteContext, next routerpkg.HandlerFunc) {

	acceptEncoding := c.Request().Header.Get("Accept-Encoding")
	if acceptEncoding == "" {
		next(c)
		return
	}

	var compressor io.WriteCloser
	if strings.Contains(acceptEncoding, "gzip") {
		c.Response().Header().Set("Content-Encoding", "gzip")
		compressor = gzip.NewWriter(c.Response())
	} else if strings.Contains(acceptEncoding, "deflate") {
		c.Response().Header().Set("Content-Encoding", "deflate")
		var err error
		compressor, err = flate.NewWriter(c.Response(), flate.DefaultCompression)
		if err != nil {
			c.ServerError("Compression Failed", err.Error())
			return
		}
	} else {
		next(c)
		return
	}

	c.SetResponse(&compressionWriter{
		w: c.Response(),
		c: compressor,
	})
	defer compressor.Close()
	next(c)
}
