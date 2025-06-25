package mux

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// ---- Functional Options ----

type CompressionOptions struct{}

type CompressionOption func(*CompressionOptions)

func (rtr *Router) UseCompression(opts ...CompressionOption) {
	options := &CompressionOptions{}
	for _, opt := range opts {
		opt(options)
	}
	rtr.middleware = append(rtr.middleware, &compressionMiddleware{options: options})
}

// ---- Middleware ----

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
	acceptEncoding := c.Request.Header.Get("Accept-Encoding")
	if acceptEncoding == "" {
		next(c)
		return
	}

	var compressor io.WriteCloser
	if strings.Contains(acceptEncoding, "gzip") {
		c.Response.Header().Set("Content-Encoding", "gzip")
		compressor = gzip.NewWriter(c.Response)
	} else if strings.Contains(acceptEncoding, "deflate") {
		c.Response.Header().Set("Content-Encoding", "deflate")
		var err error
		compressor, err = flate.NewWriter(c.Response, flate.DefaultCompression)
		if err != nil {
			c.ServerError("Compression Failed", err.Error())
			return
		}
	} else {
		next(c)
		return
	}

	c.Response = &compressionWriter{
		w: c.Response,
		c: compressor,
	}
	defer compressor.Close()
	next(c)
}
