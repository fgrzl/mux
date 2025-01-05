package mux

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type CompressionOptions struct{}

func (rtr *Router) UseCompression(options *CompressionOptions) {
	rtr.middleware = append(rtr.middleware, &compressionMiddleware{options: options})
}

type compressionMiddleware struct {
	options *CompressionOptions
}

type compressionWriter struct {
	w http.ResponseWriter
	c io.WriteCloser
}

func (cw *compressionWriter) Write(p []byte) (int, error) {
	return cw.c.Write(p)
}

func (cw *compressionWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *compressionWriter) WriteHeader(statusCode int) {
	cw.w.WriteHeader(statusCode)
}

func (m *compressionMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	// Check Accept-Encoding header for supported compression algorithms
	acceptEncoding := c.Request.Header.Get("Accept-Encoding")
	if acceptEncoding == "" {
		// If no Accept-Encoding header, just call next handler
		next(c)
		return
	}

	var compressor io.WriteCloser
	// Determine the appropriate compression algorithm (gzip or deflate)
	if strings.Contains(acceptEncoding, "gzip") {
		c.Response.Header().Set("Content-Encoding", "gzip")
		compressor = gzip.NewWriter(c.Response)

	} else if strings.Contains(acceptEncoding, "deflate") {
		c.Response.Header().Set("Content-Encoding", "deflate")
		var err error
		compressor, err = flate.NewWriter(c.Response, flate.DefaultCompression)
		if err != nil {
			c.ServerError("Compression Failed", err.Error()) // Handle error in case of compression creation failure
			return
		}

	} else {
		// If no supported encoding, proceed without compression
		next(c)
		return
	}

	// Wrap the original ResponseWriter with our compression writer
	c.Response = &compressionWriter{
		w: c.Response,
		c: compressor,
	}

	// Ensure compression writer is closed once the response is done
	defer compressor.Close()

	// Call the next middleware/handler in the chain
	next(c)
}
