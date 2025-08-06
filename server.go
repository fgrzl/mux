package mux

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// WebServerOption allows functional options for WebServer.
type WebServerOption func(*WebServer)

// WebServer wraps an http.Server and your custom Router.
type WebServer struct {
	srv      *http.Server
	router   *Router
	certFile string
	keyFile  string
}

// WithTLS enables HTTPS with the given cert and key file paths.
func WithTLS(certFile, keyFile string) WebServerOption {
	return func(ws *WebServer) {
		ws.certFile = certFile
		ws.keyFile = keyFile
	}
}

// NewServer sets up the HTTP server with sane timeouts and a mux Router.
func NewServer(addr string, router *Router, opts ...WebServerOption) *WebServer {
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	ws := &WebServer{srv: srv, router: router}
	for _, opt := range opts {
		opt(ws)
	}
	return ws
}

// Start runs the HTTP or HTTPS server in a goroutine and handles graceful shutdown.
func (ws *WebServer) Start(ctx context.Context) error {
	// Validate address (optional)
	if _, err := net.ResolveTCPAddr("tcp", ws.srv.Addr); err != nil {
		return err
	}

	go func() {
		var err error
		if ws.certFile != "" && ws.keyFile != "" {
			err = ws.srv.ListenAndServeTLS(ws.certFile, ws.keyFile)
		} else {
			err = ws.srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		_ = ws.Stop(context.Background())
	}()

	return nil
}

// Stop shuts down the HTTP server gracefully.
func (ws *WebServer) Stop(ctx context.Context) error {
	return ws.srv.Shutdown(ctx)
}
